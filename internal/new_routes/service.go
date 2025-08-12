package new_routes

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	db "geolocation/db/sqlc"
	"geolocation/internal/get_token"
	"geolocation/internal/route_enterprise"
	"geolocation/internal/routes"
	"geolocation/internal/zonas_risco"
	cache "geolocation/pkg"
	"geolocation/validation"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	neturl "net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"googlemaps.github.io/maps"
)

type InterfaceService interface {
	CalculateRoutes(ctx context.Context, frontInfo FrontInfo, idPublicToken int64, idSimp int64) (FinalOutput, error)
	CalculateRoutesWithCEP(ctx context.Context, frontInfo FrontInfoCEP, idPublicToken int64, idSimp int64, payloadSimp get_token.PayloadDTO) (FinalOutput, error)
	CalculateDistancesBetweenPoints(ctx context.Context, data FrontInfoCEPRequest) (Response, error)
	CalculateDistancesFromOrigin(ctx context.Context, data FrontInfoCEPRequest) ([]DetailedRoute, error)
	CalculateRoutesWithCoordinate(ctx context.Context, frontInfo FrontInfoCoordinate, idPublicToken int64, idSimp int64) (FinalOutput, error)
	GetFavoriteRouteService(ctx context.Context, id int64) ([]FavoriteRouteResponse, error)
	RemoveFavoriteRouteService(ctx context.Context, id, idUser int64) error
	GetSimpleRoute(data SimpleRouteRequest) (SimpleRouteResponse, error)
	CalculateDistancesBetweenPointsWithRiskAvoidance(ctx context.Context, data FrontInfoCEPRequest) (Response, error)
}

type Service struct {
	InterfaceService         routes.InterfaceRepository
	InterfaceRouteEnterprise route_enterprise.InterfaceRepository
	GoogleMapsAPIKey         string
	RiskZonesRepository      zonas_risco.InterfaceService
}

func NewRoutesNewService(interfaceService routes.InterfaceRepository, interfaceRouteEnterprise route_enterprise.InterfaceRepository, googleMapsAPIKey string, RiskZonesRepository zonas_risco.InterfaceService) *Service {
	return &Service{
		InterfaceService:         interfaceService,
		InterfaceRouteEnterprise: interfaceRouteEnterprise,
		GoogleMapsAPIKey:         googleMapsAPIKey,
		RiskZonesRepository:      RiskZonesRepository,
	}
}

func (s *Service) CalculateRoutes(ctx context.Context, frontInfo FrontInfo, idPublicToken int64, idSimp int64) (FinalOutput, error) {
	if strings.ToLower(frontInfo.PublicOrPrivate) == "public" {
		if err := s.updateNumberOfRequest(ctx, idPublicToken); err != nil {
			return FinalOutput{}, err
		}
	}

	cacheKey := fmt.Sprintf("route:%s:%s:%s:axles:%d:type:%s",
		strings.ToLower(frontInfo.Origin),
		strings.ToLower(frontInfo.Destination),
		strings.ToLower(strings.Join(frontInfo.Waypoints, ",")),
		frontInfo.Axles,
		strings.ToLower(frontInfo.Type),
	)
	cached, err := cache.Rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		var cachedOutput FinalOutput

		if json.Unmarshal([]byte(cached), &cachedOutput) != nil {
			log.Printf("Erro ao deserializar o cache: %v", err)
			return FinalOutput{}, err
		}

		routeOptionsChanged := cachedOutput.Summary.RouteOptions.IncludeFuelStations != frontInfo.RouteOptions.IncludeFuelStations ||
			cachedOutput.Summary.RouteOptions.IncludeRouteMap != frontInfo.RouteOptions.IncludeRouteMap ||
			cachedOutput.Summary.RouteOptions.IncludeTollCosts != frontInfo.RouteOptions.IncludeTollCosts ||
			cachedOutput.Summary.RouteOptions.IncludeWeighStations != frontInfo.RouteOptions.IncludeWeighStations ||
			cachedOutput.Summary.RouteOptions.IncludeFreightCalc != frontInfo.RouteOptions.IncludeFreightCalc ||
			cachedOutput.Summary.RouteOptions.IncludePolyline != frontInfo.RouteOptions.IncludePolyline

		if !routeOptionsChanged {
			waypointsStr := strings.ToLower(strings.Join(frontInfo.Waypoints, ","))
			responseJSON, _ := json.Marshal(cachedOutput)
			requestJSON, _ := json.Marshal(frontInfo)

			routeHistID, errSavedRoutes := s.savedRoutes(ctx, frontInfo.PublicOrPrivate,
				cachedOutput.Summary.LocationOrigin.Address,
				cachedOutput.Summary.LocationDestination.Address,
				waypointsStr, idPublicToken, idSimp, responseJSON, requestJSON, frontInfo.Favorite)
			if errSavedRoutes != nil {
				log.Printf("Erro ao salvar rota/favorita (cache): %v", errSavedRoutes)
			}

			cachedOutput.Summary.RouteHistID = routeHistID
			return cachedOutput, nil
		}
	} else if !errors.Is(err, redis.Nil) {
		log.Printf("Erro ao recuperar cache do Redis (CalculateRoutes): %v", err)
	}

	origin, err := s.getGeocodeAddress(ctx, frontInfo.Origin)
	if err != nil {
		return FinalOutput{}, fmt.Errorf("erro ao geocodificar origem: %w", err)
	}
	destination, err := s.getGeocodeAddress(ctx, frontInfo.Destination)
	if err != nil {
		return FinalOutput{}, fmt.Errorf("erro ao geocodificar destino: %w", err)
	}

	var waypointResults []GeocodeResult
	for _, wp := range frontInfo.Waypoints {
		wp = strings.TrimSpace(wp)
		if wp != "" {
			res, err := s.getGeocodeAddress(ctx, wp)
			if err != nil {
				return FinalOutput{}, fmt.Errorf("erro ao geocodificar waypoint (%s): %w", wp, err)
			}
			waypointResults = append(waypointResults, res)
		}
	}

	fuelPrice := FuelPrice{
		Price:    frontInfo.Price,
		Currency: "BRL",
		Units:    "km",
		FuelUnit: "liter",
	}
	fuelEfficiency := FuelEfficiency{
		City:     frontInfo.ConsumptionCity,
		Hwy:      frontInfo.ConsumptionHwy,
		Units:    "km",
		FuelUnit: "liter",
	}

	coords := fmt.Sprintf("%f,%f", origin.Location.Longitude, origin.Location.Latitude)
	for _, wp := range waypointResults {
		coords += fmt.Sprintf(";%f,%f", wp.Location.Longitude, wp.Location.Latitude)
	}
	coords += fmt.Sprintf(";%f,%f", destination.Location.Longitude, destination.Location.Latitude)
	baseOSRMURL := "http://34.207.174.233:5001/route/v1/driving/" + url.PathEscape(coords)
	client := http.Client{Timeout: 120 * time.Second}

	osrmURLFast := baseOSRMURL + "?" + url.Values{
		"alternatives":      {"3"},
		"steps":             {"true"},
		"overview":          {"full"},
		"continue_straight": {"false"},
	}.Encode()

	osrmURLNoTolls := baseOSRMURL + "?" + url.Values{
		"alternatives": {"3"},
		"steps":        {"true"},
		"overview":     {"full"},
		"exclude":      {"toll"},
	}.Encode()

	osrmURLEfficient := baseOSRMURL + "?" + url.Values{
		"alternatives": {"3"},
		"steps":        {"true"},
		"overview":     {"full"},
		"exclude":      {"motorway"},
	}.Encode()

	type osrmResult struct {
		resp     OSRMResponse
		err      error
		category string
	}
	resultsCh := make(chan osrmResult, 3)

	makeOSRMRequest := func(url, category, errMsg string) {
		resp, err := client.Get(url)
		if err != nil {
			resultsCh <- osrmResult{err: fmt.Errorf("%s: %w", errMsg, err), category: category}
			return
		}
		defer resp.Body.Close()
		var osrmResp OSRMResponse
		if err := json.NewDecoder(resp.Body).Decode(&osrmResp); err != nil {
			resultsCh <- osrmResult{err: fmt.Errorf("erro ao decodificar resposta OSRM (%s): %w", category, err), category: category}
			return
		}
		if osrmResp.Code != "Ok" || len(osrmResp.Routes) == 0 {
			resultsCh <- osrmResult{err: fmt.Errorf("OSRM (%s) retornou erro ou nenhuma rota encontrada", category), category: category}
			return
		}
		resultsCh <- osrmResult{resp: osrmResp, category: category}
	}

	go makeOSRMRequest(osrmURLFast, "fatest", "erro na requisição OSRM (rota rápida)")
	go makeOSRMRequest(osrmURLNoTolls, "cheapest", "erro na requisição OSRM (rota com menos pedágio)")
	go makeOSRMRequest(osrmURLEfficient, "efficient", "erro na requisição OSRM (rota eficiente)")

	var osrmRespFast, osrmRespNoTolls, osrmRespEfficient OSRMResponse
	for i := 0; i < 3; i++ {
		res := <-resultsCh
		if res.err != nil {
			return FinalOutput{}, res.err
		}
		switch res.category {
		case "fatest":
			osrmRespFast = res.resp
		case "cheapest":
			osrmRespNoTolls = res.resp
		case "efficient":
			osrmRespEfficient = res.resp
		}
	}

	dbCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	googleURL := fmt.Sprintf("https://www.google.com/maps/dir/?api=1&origin=%s&destination=%s",
		neturl.QueryEscape(origin.FormattedAddress),
		neturl.QueryEscape(destination.FormattedAddress))
	if len(frontInfo.Waypoints) > 0 {
		googleURL += "&waypoints=" + neturl.QueryEscape(strings.Join(frontInfo.Waypoints, "|"))
	}
	currentTimeMillis := (time.Now().UnixNano() + int64(osrmRespFast.Routes[0].Duration*float64(time.Second))) / int64(time.Millisecond)
	wazeURL := ""
	if origin.PlaceID != "" && destination.PlaceID != "" {
		wazeURL = fmt.Sprintf("https://www.waze.com/pt-BR/live-map/directions/br?to=place.%s&from=place.%s&time=%d&reverse=yes",
			neturl.QueryEscape(destination.PlaceID),
			neturl.QueryEscape(origin.PlaceID),
			currentTimeMillis,
		)
		if len(frontInfo.Waypoints) > 0 {
			wazeURL += "&via=place." + neturl.QueryEscape(frontInfo.Waypoints[0])
		}
	}

	processRoutes := func(osrmResp OSRMResponse, routeCategory string) []RouteOutput {
		var output []RouteOutput
		for _, route := range osrmResp.Routes {
			routeGasStations, err := s.findGasStations(dbCtx, route.Geometry)
			if err != nil {
				log.Printf("Erro ao consultar postos de gasolina: %v", err)
				routeGasStations = nil
			}

			balancas, err := s.InterfaceService.GetBalanca(ctx)
			if err != nil {
				log.Printf("Erro ao obter balanças: %v", err)
				balancas = nil
			}

			distText, distVal := formatDistance(route.Distance)
			durText, durVal := formatDuration(route.Duration)

			var finalInstructions []Instruction
			if len(route.Legs) > 0 {
				for _, step := range route.Legs[0].Steps {
					text := translateInstruction(step)
					instructionLower := strings.ToLower(text)
					var valueImg string
					switch {
					case strings.Contains(instructionLower, "direita") && (strings.Contains(instructionLower, "curva") || strings.Contains(instructionLower, "mantenha-se")):
						valueImg = "https://plates-routes.s3.us-east-1.amazonaws.com/curva-direita.png"
					case strings.Contains(instructionLower, "esquerda") && (strings.Contains(instructionLower, "curva") || strings.Contains(instructionLower, "mantenha-se")):
						valueImg = "https://plates-routes.s3.us-east-1.amazonaws.com/curva-esquerda.png"
					case strings.Contains(instructionLower, "esquerda") && !strings.Contains(instructionLower, "curva"):
						valueImg = "https://plates-routes.s3.us-east-1.amazonaws.com/esquerda.png"
					case strings.Contains(instructionLower, "direita") && !strings.Contains(instructionLower, "curva"):
						valueImg = "https://plates-routes.s3.us-east-1.amazonaws.com/direita.png"
					case strings.Contains(instructionLower, "continue"), strings.Contains(instructionLower, "siga"), strings.Contains(instructionLower, "pegue"), strings.Contains(instructionLower, "fusão"), strings.Contains(instructionLower, "inicie"):
						valueImg = "https://plates-routes.s3.us-east-1.amazonaws.com/reto.png"
					case strings.Contains(instructionLower, "rotatória"), strings.Contains(instructionLower, "rotatoria"), strings.Contains(instructionLower, "retorno"):
						valueImg = "https://plates-routes.s3.us-east-1.amazonaws.com/rotatoria.png"
					case strings.Contains(instructionLower, "voltar"), strings.Contains(instructionLower, "volta"):
						valueImg = "https://plates-routes.s3.us-east-1.amazonaws.com/voltar.png"
					case strings.Contains(instructionLower, "vire"):
						valueImg = "https://plates-routes.s3.us-east-1.amazonaws.com/direita.png"
					default:
						valueImg = ""
					}

					finalInstructions = append(finalInstructions, Instruction{
						Text: text,
						Img:  valueImg,
					})
				}
			}

			rawTolls, err := s.findTollsOnRoute(dbCtx, route.Geometry, frontInfo.Type, float64(frontInfo.Axles))
			if err != nil {
				log.Printf("Erro ao filtrar pedágios: %v", err)
				rawTolls = nil
			}
			var routeTolls []Toll
			for _, t := range rawTolls {
				routeTolls = append(routeTolls, t)
			}

			routeBalancas, err := s.findBalancaOnRoute(route.Geometry, balancas)
			if err != nil {
				log.Printf("Erro ao filtrar balanças: %v", err)
				routeBalancas = nil
			}

			kmValue := route.Distance / 1000.0
			freight, err := s.getAllFreight(dbCtx, frontInfo.Axles, kmValue)
			if err != nil {
				log.Printf("Erro ao calcular freight: %v", err)
				freight = nil
			}

			routeType := routeCategory
			var totalTollCost float64
			for _, toll := range rawTolls {
				totalTollCost += toll.CashCost
			}

			fuelCostCity := math.Round((frontInfo.Price / frontInfo.ConsumptionCity) * (float64(distVal) / 1000))
			fuelCostHwy := math.Round((frontInfo.Price / frontInfo.ConsumptionHwy) * (float64(distVal) / 1000))

			avgConsumption := (frontInfo.ConsumptionCity + frontInfo.ConsumptionHwy) / 2
			totalKm := float64(distVal) / 1000
			totalFuelCost := math.Round((frontInfo.Price / avgConsumption) * totalKm)

			output = append(output, RouteOutput{
				Summary: RouteSummary{
					RouteType: routeType,
					HasTolls:  len(routeTolls) > 0,
					Distance: Distance{
						Text:  distText,
						Value: distVal,
					},
					Duration: Duration{
						Text:  durText,
						Value: durVal,
					},
					URL:           googleURL,
					URLWaze:       wazeURL,
					TotalFuelCost: totalFuelCost,
				},
				Costs: func() *Costs {
					if frontInfo.RouteOptions.IncludeTollCosts {
						return &Costs{
							TagAndCash:      totalTollCost,
							FuelInTheCity:   fuelCostCity,
							FuelInTheHwy:    fuelCostHwy,
							Tag:             (totalTollCost - (totalTollCost * 0.05)) * float64(frontInfo.Axles),
							Cash:            totalTollCost * float64(frontInfo.Axles),
							PrepaidCard:     totalTollCost * float64(frontInfo.Axles),
							MaximumTollCost: totalTollCost * float64(frontInfo.Axles),
							MinimumTollCost: totalTollCost * float64(frontInfo.Axles),
							Axles:           int(frontInfo.Axles),
						}
					}
					return nil
				}(),

				Tolls: func() []Toll {
					if frontInfo.RouteOptions.IncludeTollCosts {
						return routeTolls
					}
					return nil
				}(),

				Balances: func() interface{} {
					if frontInfo.RouteOptions.IncludeWeighStations {
						return routeBalancas
					}
					return nil
				}(),

				GasStations: func() []GasStation {
					if frontInfo.RouteOptions.IncludeFuelStations {
						return routeGasStations
					}
					return nil
				}(),

				Instructions: func() []Instruction {
					if frontInfo.RouteOptions.IncludeRouteMap {
						return finalInstructions
					}
					return nil
				}(),

				FreightLoad: func() map[string]interface{} {
					if frontInfo.RouteOptions.IncludeFreightCalc {
						return freight
					}
					return nil
				}(),

				Polyline: func() string {
					if frontInfo.RouteOptions.IncludePolyline {
						return route.Geometry
					}
					return ""
				}(),
			})
		}

		sort.Slice(output, func(i, j int) bool {
			return len(output[i].Tolls) < len(output[j].Tolls)
		})
		return output
	}

	if isAllRouteOptionsDisabled(frontInfo.RouteOptions) {
		var osrmRoute OSRMResponse
		if len(osrmRespEfficient.Routes) > 0 {
			osrmRoute = osrmRespEfficient
		} else if len(osrmRespFast.Routes) > 0 {
			osrmRoute = osrmRespFast
		} else if len(osrmRespNoTolls.Routes) > 0 {
			osrmRoute = osrmRespNoTolls
		} else {
			return FinalOutput{}, fmt.Errorf("nenhuma rota disponível para retorno mínimo")
		}

		route := osrmRoute.Routes[0]
		distText, distVal := formatDistance(route.Distance)
		durText, durVal := formatDuration(route.Duration)

		avgConsumption := (frontInfo.ConsumptionCity + frontInfo.ConsumptionHwy) / 2
		totalKm := float64(distVal) / 1000
		totalFuelCost := math.Round((frontInfo.Price / avgConsumption) * totalKm)

		minimalRoute := RouteOutput{
			Summary: RouteSummary{
				RouteType:     "efficient",
				HasTolls:      false,
				Distance:      Distance{Text: distText, Value: distVal},
				Duration:      Duration{Text: durText, Value: durVal},
				URL:           googleURL,
				URLWaze:       wazeURL,
				TotalFuelCost: totalFuelCost,
			},
		}

		finalOutput := FinalOutput{
			Summary: Summary{
				LocationOrigin: AddressInfo{
					Location: Location{
						Latitude:  origin.Location.Latitude,
						Longitude: origin.Location.Longitude,
					},
					Address: origin.FormattedAddress,
				},
				LocationDestination: AddressInfo{
					Location: Location{
						Latitude:  destination.Location.Latitude,
						Longitude: destination.Location.Longitude,
					},
					Address: destination.FormattedAddress,
				},
				AllStoppingPoints: func() []interface{} {
					var stops []interface{}
					for _, wp := range waypointResults {
						stops = append(stops, wp)
					}
					return stops
				}(),
				FuelPrice:      fuelPrice,
				FuelEfficiency: fuelEfficiency,
				RouteOptions:   frontInfo.RouteOptions,
			},
			Routes: []RouteOutput{minimalRoute},
		}

		responseJSON, _ := json.Marshal(finalOutput)
		requestJSON, _ := json.Marshal(frontInfo)
		waypointsStr := strings.ToLower(strings.Join(frontInfo.Waypoints, ","))

		result, errSavedRoutes := s.savedRoutes(ctx, frontInfo.PublicOrPrivate,
			origin.FormattedAddress, destination.FormattedAddress,
			waypointsStr, idPublicToken, idSimp, responseJSON, requestJSON, frontInfo.Favorite)
		if errSavedRoutes != nil {
			return FinalOutput{}, errSavedRoutes
		}
		finalOutput.Summary.RouteHistID = result

		return finalOutput, nil
	}

	routesFast := processRoutes(osrmRespFast, "fatest")
	routesNoTolls := processRoutes(osrmRespNoTolls, "cheapest")
	routesEfficient := processRoutes(osrmRespEfficient, "efficient")

	var combinedRoutes []RouteOutput
	switch strings.ToLower(frontInfo.TypeRoute) {
	case "efficient", "eficiente":
		if len(routesEfficient) > 0 {
			combinedRoutes = []RouteOutput{routesEfficient[0]}
		}
	case "fatest", "fast", "rapida":
		if len(routesFast) > 0 {
			combinedRoutes = []RouteOutput{routesFast[0]}
		}
	case "cheapest", "cheap", "barata":
		if len(routesNoTolls) > 0 {
			combinedRoutes = []RouteOutput{routesNoTolls[0]}
		}
	default:
		combinedRoutes = append(append(routesFast, routesNoTolls...), routesEfficient...)
	}

	finalOutput := FinalOutput{
		Summary: Summary{
			LocationOrigin: AddressInfo{
				Location: Location{
					Latitude:  origin.Location.Latitude,
					Longitude: origin.Location.Longitude,
				},
				Address: origin.FormattedAddress,
			},
			LocationDestination: AddressInfo{
				Location: Location{
					Latitude:  destination.Location.Latitude,
					Longitude: destination.Location.Longitude,
				},
				Address: destination.FormattedAddress,
			},
			AllStoppingPoints: func() []interface{} {
				var stops []interface{}
				for _, wp := range waypointResults {
					stops = append(stops, wp)
				}
				return stops
			}(),
			FuelPrice:      fuelPrice,
			FuelEfficiency: fuelEfficiency,
		},
		Routes: combinedRoutes,
	}

	if data, err := json.Marshal(finalOutput); err == nil {
		if err := cache.Rdb.Set(ctx, cacheKey, data, 10000*24*time.Hour).Err(); err != nil {
			log.Printf("Erro ao salvar cache do Redis (CalculateRoutes): %v", err)
		}
	}

	waypointsStr := strings.ToLower(strings.Join(frontInfo.Waypoints, ","))
	responseJSON, _ := json.Marshal(finalOutput)
	requestJSON, _ := json.Marshal(frontInfo)

	result, errSavedRoutes := s.savedRoutes(ctx, frontInfo.PublicOrPrivate,
		origin.FormattedAddress, destination.FormattedAddress,
		waypointsStr, idPublicToken, idSimp, responseJSON, requestJSON, frontInfo.Favorite)
	if errSavedRoutes != nil {
		return FinalOutput{}, errSavedRoutes
	}
	finalOutput.Summary.RouteHistID = result

	return finalOutput, nil
}

func (s *Service) CalculateRoutesWithCEP(ctx context.Context, frontInfo FrontInfoCEP, idPublicToken int64, idSimp int64, payloadSimp get_token.PayloadDTO) (FinalOutput, error) {
	if strings.ToLower(frontInfo.PublicOrPrivate) == "public" {
		if err := s.updateNumberOfRequest(ctx, idPublicToken); err != nil {
			return FinalOutput{}, err
		}
	}

	cepOrigin := frontInfo.OriginCEP
	if frontInfo.Enterprise {
		resultOrg, errOrg := s.InterfaceRouteEnterprise.GetOrganizationByTenant(ctx, db.GetOrganizationByTenantParams{
			AccessID: sql.NullInt64{
				Int64: payloadSimp.AccessID,
				Valid: true,
			},
			TenantID: uuid.NullUUID{
				UUID:  payloadSimp.TenantID,
				Valid: true,
			},
			Cnpj: payloadSimp.Document,
		})
		if errOrg != nil {
			return FinalOutput{}, errOrg
		}

		cepOrigin = resultOrg.String
	}

	waypointsStr := strings.ToLower(strings.Join(frontInfo.WaypointsCEP, ","))
	cacheKey := fmt.Sprintf("route:%s:%s:%s:%s:waypoints:%s:axles:%d:type:%s",
		strings.ToLower(cepOrigin),
		strings.ToLower(frontInfo.DestinationCEP),
		waypointsStr,
		frontInfo.Axles,
		strings.ToLower(frontInfo.Type),
	)
	cached, err := cache.Rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		var cachedOutput FinalOutput
		if json.Unmarshal([]byte(cached), &cachedOutput) == nil {
			routeOptionsChanged := cachedOutput.Summary.RouteOptions.IncludeFuelStations != frontInfo.RouteOptions.IncludeFuelStations ||
				cachedOutput.Summary.RouteOptions.IncludeRouteMap != frontInfo.RouteOptions.IncludeRouteMap ||
				cachedOutput.Summary.RouteOptions.IncludeTollCosts != frontInfo.RouteOptions.IncludeTollCosts ||
				cachedOutput.Summary.RouteOptions.IncludeWeighStations != frontInfo.RouteOptions.IncludeWeighStations ||
				cachedOutput.Summary.RouteOptions.IncludeFreightCalc != frontInfo.RouteOptions.IncludeFreightCalc ||
				cachedOutput.Summary.RouteOptions.IncludePolyline != frontInfo.RouteOptions.IncludePolyline

			if !routeOptionsChanged {
				responseJSON, _ := json.Marshal(cachedOutput)
				requestJSON, _ := json.Marshal(frontInfo)

				routeHistID, errSavedRoutes := s.savedRoutes(ctx, frontInfo.PublicOrPrivate,
					cachedOutput.Summary.LocationOrigin.Address,
					cachedOutput.Summary.LocationDestination.Address,
					waypointsStr, idPublicToken, idSimp, responseJSON, requestJSON, frontInfo.Favorite)

				if errSavedRoutes != nil {
					log.Printf("Erro ao salvar rota/favorita (cache): %v", errSavedRoutes)
				}

				cachedOutput.Summary.RouteHistID = routeHistID
				return cachedOutput, nil
			}
		}
	} else if !errors.Is(err, redis.Nil) {
		log.Printf("Erro ao recuperar cache do Redis (CalculateRoutes): %v", err)
	}

	originLat, originLon, err := s.getCoordByCEP(ctx, cepOrigin)
	if err != nil {
		return FinalOutput{}, err
	}
	destLat, destLon, err := s.getCoordByCEP(ctx, frontInfo.DestinationCEP)
	if err != nil {
		return FinalOutput{}, err
	}
	originAddress, err := s.reverseGeocode(originLat, originLon)
	if err != nil {
		return FinalOutput{}, fmt.Errorf("erro ao obter endereço reverso da origem: %w", err)
	}
	destinationAddress, err := s.reverseGeocode(destLat, destLon)
	if err != nil {
		return FinalOutput{}, fmt.Errorf("erro Lat obter endereço reverso do destino: %w", err)
	}

	originGeocode, err := s.getGeocodeAddress(ctx, originAddress)
	if err != nil {
		return FinalOutput{}, fmt.Errorf("erro ao geocodificar a origem: %w", err)
	}
	origin := originGeocode
	origin.Location = Location{Latitude: originLat, Longitude: originLon}

	destinationGeocode, err := s.getGeocodeAddress(ctx, destinationAddress)
	if err != nil {
		return FinalOutput{}, fmt.Errorf("erro ao geocodificar o destino: %w", err)
	}
	destination := destinationGeocode
	destination.Location = Location{Latitude: destLat, Longitude: destLon}

	var waypointResults []GeocodeResult
	for _, wp := range frontInfo.WaypointsCEP {
		wpCordLat, wpCordLon, err := s.getCoordByCEP(ctx, wp)
		if err != nil {
			return FinalOutput{}, err
		}

		address, err := s.reverseGeocode(wpCordLat, wpCordLon)
		if err != nil {
			log.Printf("Erro ao buscar endereço reverso do waypoint (%f, %f): %v", wpCordLat, wpCordLon, err)
			address = fmt.Sprintf("%.6f, %.6f", wpCordLat, wpCordLon)
		}

		placeId, err := s.getGeocodeAddress(ctx, address)
		if err != nil {
			return FinalOutput{}, fmt.Errorf("erro ao geocodificar a origem: %w", err)
		}
		waypointResults = append(waypointResults, GeocodeResult{
			Location:         Location{Latitude: wpCordLat, Longitude: wpCordLon},
			FormattedAddress: address,
			PlaceID:          placeId.PlaceID,
		})
	}

	fuelPrice := FuelPrice{
		Price:    frontInfo.Price,
		Currency: "BRL",
		Units:    "km",
		FuelUnit: "liter",
	}
	fuelEfficiency := FuelEfficiency{
		City:     frontInfo.ConsumptionCity,
		Hwy:      frontInfo.ConsumptionHwy,
		Units:    "km",
		FuelUnit: "liter",
	}

	coords := fmt.Sprintf("%f,%f", origin.Location.Longitude, origin.Location.Latitude)
	for _, wp := range waypointResults {
		coords += fmt.Sprintf(";%f,%f", wp.Location.Longitude, wp.Location.Latitude)
	}
	coords += fmt.Sprintf(";%f,%f", destination.Location.Longitude, destination.Location.Latitude)
	baseOSRMURL := "http://34.207.174.233:5001/route/v1/driving/" + url.PathEscape(coords)
	client := http.Client{Timeout: 120 * time.Second}

	osrmURLFast := baseOSRMURL + "?" + url.Values{
		"alternatives":      {"3"},
		"steps":             {"true"},
		"overview":          {"full"},
		"continue_straight": {"false"},
	}.Encode()

	osrmURLNoTolls := baseOSRMURL + "?" + url.Values{
		"alternatives": {"3"},
		"steps":        {"true"},
		"overview":     {"full"},
		"exclude":      {"toll"},
	}.Encode()

	osrmURLEfficient := baseOSRMURL + "?" + url.Values{
		"alternatives": {"3"},
		"steps":        {"true"},
		"overview":     {"full"},
		"exclude":      {"motorway"},
	}.Encode()

	type osrmResult struct {
		resp     OSRMResponse
		err      error
		category string
	}
	resultsCh := make(chan osrmResult, 3)

	fmt.Println(osrmURLEfficient)
	fmt.Println(osrmURLFast)
	fmt.Println(osrmURLNoTolls)
	makeOSRMRequest := func(url, category, errMsg string) {
		resp, err := client.Get(url)
		if err != nil {
			resultsCh <- osrmResult{err: fmt.Errorf("%s: %w", errMsg, err), category: category}
			return
		}
		defer resp.Body.Close()
		var osrmResp OSRMResponse
		if err := json.NewDecoder(resp.Body).Decode(&osrmResp); err != nil {
			resultsCh <- osrmResult{err: fmt.Errorf("erro ao decodificar resposta OSRM (%s): %w", category, err), category: category}
			return
		}
		if osrmResp.Code != "Ok" {
			resultsCh <- osrmResult{err: fmt.Errorf("OSRM (%s) retornou erro ou nenhuma rota encontrada", category), category: category}
			return
		}
		resultsCh <- osrmResult{resp: osrmResp, category: category}
	}

	go makeOSRMRequest(osrmURLFast, "fatest", "erro na requisição OSRM (rota rápida)")
	go makeOSRMRequest(osrmURLNoTolls, "cheapest", "erro na requisição OSRM (rota com menos pedágio)")
	go makeOSRMRequest(osrmURLEfficient, "efficient", "erro na requisição OSRM (rota eficiente)")

	var osrmRespFast, osrmRespNoTolls, osrmRespEfficient OSRMResponse
	for i := 0; i < 3; i++ {
		res := <-resultsCh
		if res.err != nil {
			return FinalOutput{}, res.err
		}
		switch res.category {
		case "fatest":
			osrmRespFast = res.resp
		case "cheapest":
			osrmRespNoTolls = res.resp
		case "efficient":
			osrmRespEfficient = res.resp
		}
	}

	dbCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	googleURL := fmt.Sprintf("https://www.google.com/maps/dir/?api=1&origin=%s&destination=%s",
		neturl.QueryEscape(origin.FormattedAddress),
		neturl.QueryEscape(destination.FormattedAddress))

	if len(frontInfo.WaypointsCEP) > 0 {
		var googleWp []string
		for _, wp := range frontInfo.WaypointsCEP {
			wpCordLat, wpCordLon, err := s.getCoordByCEP(ctx, wp)
			if err != nil {
				return FinalOutput{}, err
			}

			latStr := strconv.FormatFloat(wpCordLat, 'f', -1, 64)
			lngStr := strconv.FormatFloat(wpCordLon, 'f', -1, 64)
			googleWp = append(googleWp, latStr+","+lngStr)
		}
		googleURL += "&waypoints=" + neturl.QueryEscape(strings.Join(googleWp, "|"))
	}

	currentTimeMillis := (time.Now().UnixNano() + int64(osrmRespFast.Routes[0].Duration*float64(time.Second))) / int64(time.Millisecond)

	wazeURL := ""
	if origin.PlaceID != "" && destination.PlaceID != "" {
		wazeURL = fmt.Sprintf("https://www.waze.com/pt-BR/live-map/directions/br?to=place.%s&from=place.%s&time=%d&reverse=yes",
			neturl.QueryEscape(destination.PlaceID),
			neturl.QueryEscape(origin.PlaceID),
			currentTimeMillis,
		)
		if len(waypointResults) > 0 && waypointResults[0].PlaceID != "" {
			wazeURL += "&via=place." + neturl.QueryEscape(waypointResults[0].PlaceID)
		}
	}

	processRoutes := func(osrmResp OSRMResponse, routeCategory string) []RouteOutput {
		var output []RouteOutput
		for _, route := range osrmResp.Routes {
			balancas, err := s.InterfaceService.GetBalanca(ctx)
			if err != nil {
				log.Printf("Erro ao obter balanças: %v", err)
				balancas = nil
			}

			routeGasStations, err := s.findGasStations(dbCtx, osrmRespFast.Routes[0].Geometry)
			if err != nil {
				log.Printf("Erro ao consultar postos de gasolina: %v", err)
				routeGasStations = nil
			}

			distText, distVal := formatDistance(route.Distance)
			durText, durVal := formatDuration(route.Duration)

			var finalInstructions []Instruction
			if len(route.Legs) > 0 {
				for _, step := range route.Legs[0].Steps {
					text := translateInstruction(step)
					instructionLower := strings.ToLower(text)
					var valueImg string
					switch {
					case strings.Contains(instructionLower, "direita") && (strings.Contains(instructionLower, "curva") || strings.Contains(instructionLower, "mantenha-se")):
						valueImg = "https://plates-routes.s3.us-east-1.amazonaws.com/curva-direita.png"
					case strings.Contains(instructionLower, "esquerda") && (strings.Contains(instructionLower, "curva") || strings.Contains(instructionLower, "mantenha-se")):
						valueImg = "https://plates-routes.s3.us-east-1.amazonaws.com/curva-esquerda.png"
					case strings.Contains(instructionLower, "esquerda") && !strings.Contains(instructionLower, "curva"):
						valueImg = "https://plates-routes.s3.us-east-1.amazonaws.com/esquerda.png"
					case strings.Contains(instructionLower, "direita") && !strings.Contains(instructionLower, "curva"):
						valueImg = "https://plates-routes.s3.us-east-1.amazonaws.com/direita.png"
					case strings.Contains(instructionLower, "continue"), strings.Contains(instructionLower, "siga"), strings.Contains(instructionLower, "pegue"), strings.Contains(instructionLower, "fusão"), strings.Contains(instructionLower, "inicie"):
						valueImg = "https://plates-routes.s3.us-east-1.amazonaws.com/reto.png"
					case strings.Contains(instructionLower, "rotatória"), strings.Contains(instructionLower, "rotatoria"), strings.Contains(instructionLower, "retorno"):
						valueImg = "https://plates-routes.s3.us-east-1.amazonaws.com/rotatoria.png"
					case strings.Contains(instructionLower, "voltar"), strings.Contains(instructionLower, "volta"):
						valueImg = "https://plates-routes.s3.us-east-1.amazonaws.com/voltar.png"
					case strings.Contains(instructionLower, "vire"):
						valueImg = "https://plates-routes.s3.us-east-1.amazonaws.com/direita.png"
					default:
						valueImg = ""
					}

					finalInstructions = append(finalInstructions, Instruction{
						Text: text,
						Img:  valueImg,
					})
				}
			}

			rawTolls, err := s.findTollsOnRoute(dbCtx, route.Geometry, frontInfo.Type, float64(frontInfo.Axles))
			if err != nil {
				log.Printf("Erro ao filtrar pedágios: %v", err)
				rawTolls = nil
			}
			var routeTolls []Toll
			for _, t := range rawTolls {
				routeTolls = append(routeTolls, t)
			}

			routeBalancas, err := s.findBalancaOnRoute(route.Geometry, balancas)
			if err != nil {
				log.Printf("Erro ao filtrar balanças: %v", err)
				routeBalancas = nil
			}

			kmValue := route.Distance / 1000.0
			freight, err := s.getAllFreight(dbCtx, frontInfo.Axles, kmValue)
			if err != nil {
				log.Printf("Erro ao calcular freight: %v", err)
				freight = nil
			}

			routeType := routeCategory
			var totalTollCost float64
			for _, toll := range rawTolls {
				totalTollCost += toll.CashCost
			}

			fuelCostCity := math.Round((frontInfo.Price / frontInfo.ConsumptionCity) * (float64(distVal) / 1000))
			fuelCostHwy := math.Round((frontInfo.Price / frontInfo.ConsumptionHwy) * (float64(distVal) / 1000))

			avgConsumption := (frontInfo.ConsumptionCity + frontInfo.ConsumptionHwy) / 2
			totalKm := float64(distVal) / 1000
			totalFuelCost := math.Round((frontInfo.Price / avgConsumption) * totalKm)

			output = append(output, RouteOutput{
				Summary: RouteSummary{
					RouteType: routeType,
					HasTolls:  len(routeTolls) > 0,
					Distance: Distance{
						Text:  distText,
						Value: distVal,
					},
					Duration: Duration{
						Text:  durText,
						Value: durVal,
					},
					URL:           googleURL,
					URLWaze:       wazeURL,
					TotalFuelCost: totalFuelCost,
				},
				Costs: func() *Costs {
					if frontInfo.RouteOptions.IncludeTollCosts {
						return &Costs{
							TagAndCash:      totalTollCost,
							FuelInTheCity:   fuelCostCity,
							FuelInTheHwy:    fuelCostHwy,
							Tag:             (totalTollCost - (totalTollCost * 0.05)) * float64(frontInfo.Axles),
							Cash:            totalTollCost * float64(frontInfo.Axles),
							PrepaidCard:     totalTollCost * float64(frontInfo.Axles),
							MaximumTollCost: totalTollCost * float64(frontInfo.Axles),
							MinimumTollCost: totalTollCost * float64(frontInfo.Axles),
							Axles:           int(frontInfo.Axles),
						}
					}
					return nil
				}(),

				Tolls: func() []Toll {
					if frontInfo.RouteOptions.IncludeTollCosts {
						return routeTolls
					}
					return nil
				}(),

				Balances: func() interface{} {
					if frontInfo.RouteOptions.IncludeWeighStations {
						return routeBalancas
					}
					return nil
				}(),

				GasStations: func() []GasStation {
					if frontInfo.RouteOptions.IncludeFuelStations {
						return routeGasStations
					}
					return nil
				}(),

				Instructions: func() []Instruction {
					if frontInfo.RouteOptions.IncludeRouteMap {
						return finalInstructions
					}
					return nil
				}(),

				FreightLoad: func() map[string]interface{} {
					if frontInfo.RouteOptions.IncludeFreightCalc {
						return freight
					}
					return nil
				}(),

				Polyline: func() string {
					if frontInfo.RouteOptions.IncludePolyline {
						return route.Geometry
					}
					return ""
				}(),
			})
		}

		sort.Slice(output, func(i, j int) bool {
			return len(output[i].Tolls) < len(output[j].Tolls)
		})
		return output
	}

	if isAllRouteOptionsDisabled(frontInfo.RouteOptions) {
		var osrmRoute OSRMResponse
		if len(osrmRespEfficient.Routes) > 0 {
			osrmRoute = osrmRespEfficient
		} else if len(osrmRespFast.Routes) > 0 {
			osrmRoute = osrmRespFast
		} else if len(osrmRespNoTolls.Routes) > 0 {
			osrmRoute = osrmRespNoTolls
		} else {
			return FinalOutput{}, fmt.Errorf("nenhuma rota disponível para retorno mínimo")
		}

		route := osrmRoute.Routes[0]
		distText, distVal := formatDistance(route.Distance)
		durText, durVal := formatDuration(route.Duration)

		avgConsumption := (frontInfo.ConsumptionCity + frontInfo.ConsumptionHwy) / 2
		totalKm := float64(distVal) / 1000
		totalFuelCost := math.Round((frontInfo.Price / avgConsumption) * totalKm)

		minimalRoute := RouteOutput{
			Summary: RouteSummary{
				RouteType:     "efficient",
				HasTolls:      false,
				Distance:      Distance{Text: distText, Value: distVal},
				Duration:      Duration{Text: durText, Value: durVal},
				URL:           googleURL,
				URLWaze:       wazeURL,
				TotalFuelCost: totalFuelCost,
			},
		}

		finalOutput := FinalOutput{
			Summary: Summary{
				LocationOrigin: AddressInfo{
					Location: Location{
						Latitude:  origin.Location.Latitude,
						Longitude: origin.Location.Longitude,
					},
					Address: origin.FormattedAddress,
				},
				LocationDestination: AddressInfo{
					Location: Location{
						Latitude:  destination.Location.Latitude,
						Longitude: destination.Location.Longitude,
					},
					Address: destination.FormattedAddress,
				},
				AllStoppingPoints: func() []interface{} {
					var stops []interface{}
					for _, wp := range waypointResults {
						stops = append(stops, wp)
					}
					return stops
				}(),
				FuelPrice:      fuelPrice,
				FuelEfficiency: fuelEfficiency,
			},
			Routes: []RouteOutput{minimalRoute},
		}

		responseJSON, _ := json.Marshal(finalOutput)
		requestJSON, _ := json.Marshal(frontInfo)
		var wpStrings []string
		for _, wp := range frontInfo.WaypointsCEP {
			wpCordLat, wpCordLon, err := s.getCoordByCEP(ctx, wp)
			if err != nil {
				return FinalOutput{}, err
			}

			latStr := strconv.FormatFloat(wpCordLat, 'f', -1, 64)
			lngStr := strconv.FormatFloat(wpCordLon, 'f', -1, 64)

			wpStrings = append(wpStrings, fmt.Sprintf("%s,%s", latStr, lngStr))
		}
		waypointsStr := strings.ToLower(strings.Join(wpStrings, ","))

		result, errSavedRoutes := s.savedRoutes(ctx, frontInfo.PublicOrPrivate,
			origin.FormattedAddress, destination.FormattedAddress,
			waypointsStr, idPublicToken, idSimp, responseJSON, requestJSON, frontInfo.Favorite)
		if errSavedRoutes != nil {
			return FinalOutput{}, errSavedRoutes
		}
		finalOutput.Summary.RouteHistID = result

		return finalOutput, nil
	}

	routesFast := processRoutes(osrmRespFast, "fatest")
	routesNoTolls := processRoutes(osrmRespNoTolls, "cheapest")
	routesEfficient := processRoutes(osrmRespEfficient, "efficient")

	var combinedRoutes []RouteOutput
	switch strings.ToLower(frontInfo.TypeRoute) {
	case "efficient", "eficiente":
		if len(routesEfficient) > 0 {
			combinedRoutes = []RouteOutput{routesEfficient[0]}
		}
	case "fatest", "fast", "rapida":
		if len(routesFast) > 0 {
			combinedRoutes = []RouteOutput{routesFast[0]}
		}
	case "cheapest", "cheap", "barata":
		if len(routesNoTolls) > 0 {
			combinedRoutes = []RouteOutput{routesNoTolls[0]}
		}
	default:
		combinedRoutes = append(append(routesFast, routesNoTolls...), routesEfficient...)
	}

	finalOutput := FinalOutput{
		Summary: Summary{
			LocationOrigin: AddressInfo{
				Location: Location{
					Latitude:  origin.Location.Latitude,
					Longitude: origin.Location.Longitude,
				},
				Address: origin.FormattedAddress,
			},
			LocationDestination: AddressInfo{
				Location: Location{
					Latitude:  destination.Location.Latitude,
					Longitude: destination.Location.Longitude,
				},
				Address: destination.FormattedAddress,
			},
			AllStoppingPoints: func() []interface{} {
				var stops []interface{}
				for _, wp := range waypointResults {
					stops = append(stops, wp)
				}
				return stops
			}(),
			FuelPrice:      fuelPrice,
			FuelEfficiency: fuelEfficiency,
			RouteOptions:   frontInfo.RouteOptions,
		},
		Routes: combinedRoutes,
	}

	if data, err := json.Marshal(finalOutput); err == nil {
		if err := cache.Rdb.Set(ctx, cacheKey, data, 10000*24*time.Hour).Err(); err != nil {
			log.Printf("Erro ao salvar cache do Redis (CalculateRoutes): %v", err)
		}
	}

	var wpStringsResponse []string
	for _, wp := range frontInfo.WaypointsCEP {
		wpCordLat, wpCordLon, err := s.getCoordByCEP(ctx, wp)
		if err != nil {
			return FinalOutput{}, err
		}

		latStr := strconv.FormatFloat(wpCordLat, 'f', -1, 64)
		lngStr := strconv.FormatFloat(wpCordLon, 'f', -1, 64)

		wpStringsResponse = append(wpStringsResponse, fmt.Sprintf("%s,%s", latStr, lngStr))
	}
	waypointsStrResponse := strings.ToLower(strings.Join(wpStringsResponse, ","))

	responseJSON, _ := json.Marshal(finalOutput)
	requestJSON, _ := json.Marshal(frontInfo)

	result, errSavedRoutes := s.savedRoutes(ctx, frontInfo.PublicOrPrivate,
		origin.FormattedAddress, destination.FormattedAddress,
		waypointsStrResponse, idPublicToken, idSimp, responseJSON, requestJSON, frontInfo.Favorite)
	if errSavedRoutes != nil {
		return FinalOutput{}, errSavedRoutes
	}
	finalOutput.Summary.RouteHistID = result

	if frontInfo.Enterprise {
		resultE, errE := s.InterfaceRouteEnterprise.CreateRouteEnterprise(ctx, db.CreateRouteEnterpriseParams{
			Origin:      origin.FormattedAddress,
			Destination: destination.FormattedAddress,
			Waypoints: sql.NullString{
				String: waypointsStrResponse,
				Valid:  true,
			},
			Response:   responseJSON,
			CreatedWho: payloadSimp.UserNickname,
			TenantID:   payloadSimp.TenantID,
			AccessID:   payloadSimp.AccessID,
		})
		if errE != nil {
			return FinalOutput{}, errSavedRoutes
		}
		finalOutput.RouteEnterpriseId = resultE.ID
	}

	return finalOutput, nil
}

func (s *Service) CalculateDistancesBetweenPoints(ctx context.Context, data FrontInfoCEPRequest) (Response, error) {
	if len(data.CEPs) < 2 {
		return Response{}, fmt.Errorf("é necessário pelo menos dois pontos para calcular distâncias")
	}

	client := http.Client{Timeout: 60 * time.Second}
	var resultRoutes []DetailedRoute
	var totalDistance float64
	var totalDuration float64

	for i := 0; i < len(data.CEPs)-1; i++ {
		originCEP := data.CEPs[i]
		destCEP := data.CEPs[i+1]

		originLat, originLon, err := s.getCoordByCEP(ctx, originCEP)
		if err != nil {
			return Response{}, fmt.Errorf("erro ao buscar coordenadas da origem %s: %w", originCEP, err)
		}
		destLat, destLon, err := s.getCoordByCEP(ctx, destCEP)
		if err != nil {
			return Response{}, fmt.Errorf("erro ao buscar coordenadas do destino %s: %w", destCEP, err)
		}
		originAddress, _ := s.reverseGeocode(originLat, originLon)
		destAddress, _ := s.reverseGeocode(destLat, destLon)

		originGeocode, _ := s.getGeocodeAddress(ctx, originAddress)
		destGeocode, _ := s.getGeocodeAddress(ctx, destAddress)

		coords := fmt.Sprintf("%f,%f;%f,%f",
			originLon, originLat,
			destLon, destLat,
		)
		baseURL := "http://34.207.174.233:5001/route/v1/driving/" + url.PathEscape(coords)

		type osrmResult struct {
			resp     OSRMResponse
			category string
			err      error
		}
		resultsCh := make(chan osrmResult, 3)

		makeRequest := func(params url.Values, category string) {
			fullURL := baseURL + "?" + params.Encode()
			resp, err := client.Get(fullURL)
			if err != nil {
				resultsCh <- osrmResult{err: err, category: category}
				return
			}
			defer resp.Body.Close()
			var osrmResp OSRMResponse
			if err := json.NewDecoder(resp.Body).Decode(&osrmResp); err != nil {
				resultsCh <- osrmResult{err: err, category: category}
				return
			}
			if osrmResp.Code != "Ok" {
				resultsCh <- osrmResult{err: fmt.Errorf("erro OSRM %s", category), category: category}
				return
			}
			resultsCh <- osrmResult{resp: osrmResp, category: category}
		}

		var routeTypes []string
		if strings.TrimSpace(strings.ToLower(data.TypeRoute)) == "" {
			go makeRequest(url.Values{
				"alternatives":      {"3"},
				"steps":             {"true"},
				"overview":          {"full"},
				"continue_straight": {"false"},
			}, "fastest")
			go makeRequest(url.Values{
				"alternatives": {"3"},
				"steps":        {"true"},
				"overview":     {"full"},
				"exclude":      {"toll"},
			}, "cheapest")
			go makeRequest(url.Values{
				"alternatives": {"3"},
				"steps":        {"true"},
				"overview":     {"full"},
				"exclude":      {"motorway"},
			}, "efficient")
			routeTypes = []string{"fastest", "cheapest", "efficient"}
		} else {
			routeTypes = []string{strings.ToLower(data.TypeRoute)}
			switch routeTypes[0] {
			case "rapida", "fastest":
				makeRequest(url.Values{
					"alternatives":      {"3"},
					"steps":             {"true"},
					"overview":          {"full"},
					"continue_straight": {"false"},
				}, "fastest")
			case "barata", "cheapest":
				makeRequest(url.Values{
					"alternatives": {"3"},
					"steps":        {"true"},
					"overview":     {"full"},
					"exclude":      {"toll"},
				}, "cheapest")
			case "eficiente", "efficient":
				makeRequest(url.Values{
					"alternatives": {"3"},
					"steps":        {"true"},
					"overview":     {"full"},
					"exclude":      {"motorway"},
				}, "efficient")
			}
		}

		var summaries []RouteSummary

		for range routeTypes {
			res := <-resultsCh
			if res.err != nil || len(res.resp.Routes) == 0 {
				continue
			}

			route := res.resp.Routes[0]
			distText, distVal := formatDistance(route.Distance)
			durText, durVal := formatDuration(route.Duration)
			avgConsumption := (data.ConsumptionCity + data.ConsumptionHwy) / 2
			totalKm := route.Distance / 1000
			totalFuelCost := math.Round((data.Price / avgConsumption) * totalKm)

			googleURL := fmt.Sprintf("https://www.google.com/maps/dir/?api=1&origin=%s&destination=%s",
				neturl.QueryEscape(originGeocode.FormattedAddress),
				neturl.QueryEscape(destGeocode.FormattedAddress),
			)
			currentTimeMillis := (time.Now().UnixNano() + int64(route.Duration*float64(time.Second))) / int64(time.Millisecond)
			wazeURL := fmt.Sprintf("https://www.waze.com/pt-BR/live-map/directions/br?to=place.%s&from=place.%s&time=%d&reverse=yes",
				neturl.QueryEscape(destGeocode.PlaceID),
				neturl.QueryEscape(originGeocode.PlaceID),
				currentTimeMillis,
			)

			rawTolls, err := s.findTollsOnRoute(ctx, res.resp.Routes[0].Geometry, data.Type, float64(data.Axles))
			if err != nil {
				log.Printf("Erro ao filtrar pedágios: %v", err)
				rawTolls = nil
			}
			var routeTolls []Toll
			for _, t := range rawTolls {
				routeTolls = append(routeTolls, t)
			}

			var totalTollCost float64
			for _, toll := range routeTolls {
				totalTollCost += toll.CashCost
			}

			summaries = append(summaries, RouteSummary{
				RouteType:     res.category,
				HasTolls:      len(routeTolls) > 0,
				Distance:      Distance{Text: distText, Value: distVal},
				Duration:      Duration{Text: durText, Value: durVal},
				URL:           googleURL,
				URLWaze:       wazeURL,
				TotalFuelCost: totalFuelCost,
				Tolls:         routeTolls,
				TotalTolls:    math.Round(totalTollCost*100) / 100,
				Polyline:      res.resp.Routes[0].Geometry,
			})

			totalDistance += route.Distance
			totalDuration += route.Duration
		}

		resultRoutes = append(resultRoutes, DetailedRoute{
			LocationOrigin: AddressInfo{
				Location: Location{Latitude: originLat, Longitude: originLon},
				Address:  originGeocode.FormattedAddress,
			},
			LocationDestination: AddressInfo{
				Location: Location{Latitude: destLat, Longitude: destLon},
				Address:  destGeocode.FormattedAddress,
			},
			Summaries: summaries,
		})
	}

	var totalRoute TotalSummary
	var allCoords []string
	var waypoints []string
	var originLocation, destinationLocation Location
	for idx, cep := range data.CEPs {
		coordLat, coordLon, err := s.getCoordByCEP(ctx, cep)
		if err != nil {
			return Response{}, fmt.Errorf("erro ao buscar coordenadas para total_route no CEP %s: %w", cep, err)
		}
		allCoords = append(allCoords, fmt.Sprintf("%f,%f", coordLon, coordLat))

		reverse, _ := s.reverseGeocode(coordLat, coordLon)
		geocode, _ := s.getGeocodeAddress(ctx, reverse)
		waypoints = append(waypoints, geocode.FormattedAddress)

		if idx == 0 {
			originLocation = Location{Latitude: coordLat, Longitude: coordLon}
		}
		if idx == len(data.CEPs)-1 {
			destinationLocation = Location{Latitude: coordLat, Longitude: coordLon}
		}
	}

	coordsStr := strings.Join(allCoords, ";")
	urlTotal := fmt.Sprintf("http://34.207.174.233:5001/route/v1/driving/%s?alternatives=0&steps=true&overview=full&continue_straight=false", url.PathEscape(coordsStr))

	resp, err := client.Get(urlTotal)
	if err == nil {
		defer resp.Body.Close()
		var osrmResp OSRMResponse
		if err := json.NewDecoder(resp.Body).Decode(&osrmResp); err == nil && len(osrmResp.Routes) > 0 {
			route := osrmResp.Routes[0]

			distText, distVal := formatDistance(totalDistance)
			durText, durVal := formatDuration(totalDuration)

			avgConsumption := (data.ConsumptionCity + data.ConsumptionHwy) / 2
			totalKm := route.Distance / 1000
			totalFuelCost := math.Round((data.Price / avgConsumption) * totalKm)

			tolls, _ := s.findTollsOnRoute(ctx, route.Geometry, data.Type, float64(data.Axles))
			var totalTollCost float64
			for _, toll := range tolls {
				totalTollCost += toll.CashCost
			}

			originAddress := waypoints[0]
			destAddress := waypoints[len(waypoints)-1]
			waypointStr := ""
			if len(waypoints) > 2 {
				waypointStr = "&waypoints=" + neturl.QueryEscape(strings.Join(waypoints[1:len(waypoints)-1], "|"))
			}

			googleURL := fmt.Sprintf("https://www.google.com/maps/dir/?api=1&origin=%s&destination=%s%s&travelmode=driving",
				neturl.QueryEscape(originAddress),
				neturl.QueryEscape(destAddress),
				waypointStr,
			)

			currentTimeMillis := (time.Now().UnixNano() + int64(route.Duration*float64(time.Second))) / int64(time.Millisecond)
			wazeURL := fmt.Sprintf("https://www.waze.com/pt-BR/live-map/directions/br?to=%s&from=%s&time=%d&reverse=yes",
				neturl.QueryEscape(destAddress),
				neturl.QueryEscape(originAddress),
				currentTimeMillis,
			)

			totalRoute = TotalSummary{
				LocationOrigin: AddressInfo{
					Location: originLocation,
					Address:  originAddress,
				},
				LocationDestination: AddressInfo{
					Location: destinationLocation,
					Address:  destAddress,
				},
				TotalDistance: Distance{Text: distText, Value: distVal},
				TotalDuration: Duration{Text: durText, Value: durVal},
				URL:           googleURL,
				URLWaze:       wazeURL,
				Tolls:         tolls,
				TotalTolls:    math.Round(totalTollCost*100) / 100,
				Polyline:      route.Geometry,
				TotalFuelCost: totalFuelCost,
			}
		}
	}

	return Response{
		Routes:     resultRoutes,
		TotalRoute: totalRoute,
	}, nil
}

func (s *Service) CalculateDistancesFromOrigin(ctx context.Context, data FrontInfoCEPRequest) ([]DetailedRoute, error) {
	if len(data.CEPs) < 2 {
		return nil, fmt.Errorf("é necessário pelo menos dois pontos para calcular distâncias")
	}

	client := http.Client{Timeout: 60 * time.Second}
	originCEP := data.CEPs[0]

	originLat, originLon, err := s.getCoordByCEP(ctx, originCEP)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar coordenadas da origem %s: %w", originCEP, err)
	}
	originAddressRaw, _ := s.reverseGeocode(originLat, originLon)
	originGeocode, _ := s.getGeocodeAddress(ctx, originAddressRaw)

	var results []DetailedRoute

	for _, destCEP := range data.CEPs[1:] {
		destLat, destLon, err := s.getCoordByCEP(ctx, destCEP)
		if err != nil {
			continue
		}
		destAddressRaw, _ := s.reverseGeocode(destLat, destLon)
		destGeocode, _ := s.getGeocodeAddress(ctx, destAddressRaw)

		coords := fmt.Sprintf("%f,%f;%f,%f", originLon, originLat, destLon, destLat)
		baseURL := fmt.Sprintf("http://34.207.174.233:5001/route/v1/driving/%s?alternatives=0&steps=true&overview=full&continue_straight=false", url.PathEscape(coords))

		resp, err := client.Get(baseURL)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		var osrmResp OSRMResponse
		if err := json.NewDecoder(resp.Body).Decode(&osrmResp); err != nil || len(osrmResp.Routes) == 0 {
			continue
		}

		route := osrmResp.Routes[0]
		distText, distVal := formatDistance(route.Distance)
		durText, durVal := formatDuration(route.Duration)

		tolls, _ := s.findTollsOnRoute(ctx, route.Geometry, data.Type, float64(data.Axles))

		var totalTollCost float64
		for _, toll := range tolls {
			totalTollCost += toll.CashCost
		}

		googleURL := fmt.Sprintf("https://www.google.com/maps/dir/?api=1&origin=%s&destination=%s&travelmode=driving",
			neturl.QueryEscape(originGeocode.FormattedAddress),
			neturl.QueryEscape(destGeocode.FormattedAddress),
		)

		currentTimeMillis := (time.Now().UnixNano() + int64(route.Duration*float64(time.Second))) / int64(time.Millisecond)
		wazeURL := fmt.Sprintf("https://www.waze.com/pt-BR/live-map/directions/br?to=%s&from=%s&time=%d&reverse=yes",
			neturl.QueryEscape(originGeocode.FormattedAddress),
			neturl.QueryEscape(destGeocode.FormattedAddress),
			currentTimeMillis,
		)

		results = append(results, DetailedRoute{
			LocationOrigin: AddressInfo{
				Location: Location{Latitude: originLat, Longitude: originLon},
				Address:  originGeocode.FormattedAddress,
			},
			LocationDestination: AddressInfo{
				Location: Location{Latitude: destLat, Longitude: destLon},
				Address:  destGeocode.FormattedAddress,
			},
			Summaries: []RouteSummary{
				{
					RouteType:  "efficient",
					HasTolls:   len(tolls) > 0,
					Distance:   Distance{Text: distText, Value: distVal},
					Duration:   Duration{Text: durText, Value: durVal},
					URL:        googleURL,
					URLWaze:    wazeURL,
					Tolls:      tolls,
					TotalTolls: math.Round(totalTollCost*100) / 100,
					Polyline:   route.Geometry,
				},
			},
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Summaries[0].Distance.Value < results[j].Summaries[0].Distance.Value
	})

	return results, nil
}

func (s *Service) CalculateRoutesWithCoordinate(ctx context.Context, frontInfo FrontInfoCoordinate, idPublicToken int64, idSimp int64) (FinalOutput, error) {
	if strings.ToLower(frontInfo.PublicOrPrivate) == "public" {
		if err := s.updateNumberOfRequest(ctx, idPublicToken); err != nil {
			return FinalOutput{}, err
		}
	}

	var wpStrings []string
	for _, wp := range frontInfo.Waypoints {
		wpStrings = append(wpStrings, fmt.Sprintf("%s,%s", wp.Lat, wp.Lng))
	}
	waypointsStr := strings.ToLower(strings.Join(wpStrings, ","))

	cacheKey := fmt.Sprintf("route:%s:%s:%s:%s:waypoints:%s:axles:%d:type:%s",
		strings.ToLower(frontInfo.OriginLat),
		strings.ToLower(frontInfo.OriginLng),
		strings.ToLower(frontInfo.DestinationLat),
		strings.ToLower(frontInfo.DestinationLng),
		waypointsStr,
		frontInfo.Axles,
		strings.ToLower(frontInfo.Type),
	)
	cached, err := cache.Rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		var cachedOutput FinalOutput
		if json.Unmarshal([]byte(cached), &cachedOutput) == nil {
			routeOptionsChanged := cachedOutput.Summary.RouteOptions.IncludeFuelStations != frontInfo.RouteOptions.IncludeFuelStations ||
				cachedOutput.Summary.RouteOptions.IncludeRouteMap != frontInfo.RouteOptions.IncludeRouteMap ||
				cachedOutput.Summary.RouteOptions.IncludeTollCosts != frontInfo.RouteOptions.IncludeTollCosts ||
				cachedOutput.Summary.RouteOptions.IncludeWeighStations != frontInfo.RouteOptions.IncludeWeighStations ||
				cachedOutput.Summary.RouteOptions.IncludeFreightCalc != frontInfo.RouteOptions.IncludeFreightCalc ||
				cachedOutput.Summary.RouteOptions.IncludePolyline != frontInfo.RouteOptions.IncludePolyline

			if !routeOptionsChanged {
				responseJSON, _ := json.Marshal(cachedOutput)
				requestJSON, _ := json.Marshal(frontInfo)

				routeHistID, errSavedRoutes := s.savedRoutes(ctx, frontInfo.PublicOrPrivate,
					cachedOutput.Summary.LocationOrigin.Address,
					cachedOutput.Summary.LocationDestination.Address,
					waypointsStr, idPublicToken, idSimp, responseJSON, requestJSON, frontInfo.Favorite)

				if errSavedRoutes != nil {
					log.Printf("Erro ao salvar rota/favorita (cache): %v", errSavedRoutes)
				}

				cachedOutput.Summary.RouteHistID = routeHistID
				return cachedOutput, nil
			}
		}
	} else if !errors.Is(err, redis.Nil) {
		log.Printf("Erro ao recuperar cache do Redis (CalculateRoutes): %v", err)
	}

	originLat, _ := validation.ParseStringToFloat(frontInfo.OriginLat)
	originLng, _ := validation.ParseStringToFloat(frontInfo.OriginLng)
	destinationLat, _ := validation.ParseStringToFloat(frontInfo.DestinationLat)
	destinationLng, _ := validation.ParseStringToFloat(frontInfo.DestinationLng)

	originAddress, err := s.reverseGeocode(originLat, originLng)
	if err != nil {
		return FinalOutput{}, fmt.Errorf("erro ao obter endereço reverso da origem: %w", err)
	}
	destinationAddress, err := s.reverseGeocode(destinationLat, destinationLng)
	if err != nil {
		return FinalOutput{}, fmt.Errorf("erro ao obter endereço reverso do destino: %w", err)
	}

	originGeocode, err := s.getGeocodeAddress(ctx, originAddress)
	if err != nil {
		return FinalOutput{}, fmt.Errorf("erro ao geocodificar a origem: %w", err)
	}
	origin := originGeocode
	origin.Location = Location{Latitude: originLat, Longitude: originLng}

	destinationGeocode, err := s.getGeocodeAddress(ctx, destinationAddress)
	if err != nil {
		return FinalOutput{}, fmt.Errorf("erro ao geocodificar o destino: %w", err)
	}
	destination := destinationGeocode
	destination.Location = Location{Latitude: destinationLat, Longitude: destinationLng}

	var waypointResults []GeocodeResult
	for _, wp := range frontInfo.Waypoints {
		lat, err1 := strconv.ParseFloat(strings.TrimSpace(wp.Lat), 64)
		lng, err2 := strconv.ParseFloat(strings.TrimSpace(wp.Lng), 64)
		if err1 == nil && err2 == nil {
			address, err := s.reverseGeocode(lat, lng)
			if err != nil {
				log.Printf("Erro ao buscar endereço reverso do waypoint (%f, %f): %v", lat, lng, err)
				address = fmt.Sprintf("%.6f, %.6f", lat, lng)
			}

			placeId, err := s.getGeocodeAddress(ctx, address)
			if err != nil {
				return FinalOutput{}, fmt.Errorf("erro ao geocodificar a origem: %w", err)
			}
			waypointResults = append(waypointResults, GeocodeResult{
				Location:         Location{Latitude: lat, Longitude: lng},
				FormattedAddress: address,
				PlaceID:          placeId.PlaceID,
			})
		}
	}

	fuelPrice := FuelPrice{
		Price:    frontInfo.Price,
		Currency: "BRL",
		Units:    "km",
		FuelUnit: "liter",
	}
	fuelEfficiency := FuelEfficiency{
		City:     frontInfo.ConsumptionCity,
		Hwy:      frontInfo.ConsumptionHwy,
		Units:    "km",
		FuelUnit: "liter",
	}

	coords := fmt.Sprintf("%f,%f", origin.Location.Longitude, origin.Location.Latitude)
	for _, wp := range waypointResults {
		coords += fmt.Sprintf(";%f,%f", wp.Location.Longitude, wp.Location.Latitude)
	}
	coords += fmt.Sprintf(";%f,%f", destination.Location.Longitude, destination.Location.Latitude)
	baseOSRMURL := "http://34.207.174.233:5001/route/v1/driving/" + url.PathEscape(coords)
	client := http.Client{Timeout: 120 * time.Second}

	osrmURLFast := baseOSRMURL + "?" + url.Values{
		"alternatives":      {"3"},
		"steps":             {"true"},
		"overview":          {"full"},
		"continue_straight": {"false"},
	}.Encode()

	osrmURLNoTolls := baseOSRMURL + "?" + url.Values{
		"alternatives": {"3"},
		"steps":        {"true"},
		"overview":     {"full"},
		"exclude":      {"toll"},
	}.Encode()

	osrmURLEfficient := baseOSRMURL + "?" + url.Values{
		"alternatives": {"3"},
		"steps":        {"true"},
		"overview":     {"full"},
		"exclude":      {"motorway"},
	}.Encode()

	type osrmResult struct {
		resp     OSRMResponse
		err      error
		category string
	}
	resultsCh := make(chan osrmResult, 3)

	makeOSRMRequest := func(url, category, errMsg string) {
		resp, err := client.Get(url)
		if err != nil {
			resultsCh <- osrmResult{err: fmt.Errorf("%s: %w", errMsg, err), category: category}
			return
		}
		defer resp.Body.Close()
		var osrmResp OSRMResponse
		if err := json.NewDecoder(resp.Body).Decode(&osrmResp); err != nil {
			resultsCh <- osrmResult{err: fmt.Errorf("erro ao decodificar resposta OSRM (%s): %w", category, err), category: category}
			return
		}
		if osrmResp.Code != "Ok" || len(osrmResp.Routes) == 0 {
			resultsCh <- osrmResult{err: fmt.Errorf("OSRM (%s) retornou erro ou nenhuma rota encontrada", category), category: category}
			return
		}
		resultsCh <- osrmResult{resp: osrmResp, category: category}
	}

	go makeOSRMRequest(osrmURLFast, "fatest", "erro na requisição OSRM (rota rápida)")
	go makeOSRMRequest(osrmURLNoTolls, "cheapest", "erro na requisição OSRM (rota com menos pedágio)")
	go makeOSRMRequest(osrmURLEfficient, "efficient", "erro na requisição OSRM (rota eficiente)")

	var osrmRespFast, osrmRespNoTolls, osrmRespEfficient OSRMResponse
	for i := 0; i < 3; i++ {
		res := <-resultsCh
		if res.err != nil {
			return FinalOutput{}, res.err
		}
		switch res.category {
		case "fatest":
			osrmRespFast = res.resp
		case "cheapest":
			osrmRespNoTolls = res.resp
		case "efficient":
			osrmRespEfficient = res.resp
		}
	}

	dbCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	googleURL := fmt.Sprintf("https://www.google.com/maps/dir/?api=1&origin=%s&destination=%s",
		neturl.QueryEscape(origin.FormattedAddress),
		neturl.QueryEscape(destination.FormattedAddress))

	if len(frontInfo.Waypoints) > 0 {
		var googleWp []string
		for _, wp := range frontInfo.Waypoints {
			googleWp = append(googleWp, fmt.Sprintf("%s,%s", wp.Lat, wp.Lng))
		}
		googleURL += "&waypoints=" + neturl.QueryEscape(strings.Join(googleWp, "|"))
	}

	currentTimeMillis := (time.Now().UnixNano() + int64(osrmRespFast.Routes[0].Duration*float64(time.Second))) / int64(time.Millisecond)

	wazeURL := ""
	if origin.PlaceID != "" && destination.PlaceID != "" {
		wazeURL = fmt.Sprintf("https://www.waze.com/pt-BR/live-map/directions/br?to=place.%s&from=place.%s&time=%d&reverse=yes",
			neturl.QueryEscape(destination.PlaceID),
			neturl.QueryEscape(origin.PlaceID),
			currentTimeMillis,
		)
		if len(waypointResults) > 0 && waypointResults[0].PlaceID != "" {
			wazeURL += "&via=place." + neturl.QueryEscape(waypointResults[0].PlaceID)
		}
	}

	processRoutes := func(osrmResp OSRMResponse, routeCategory string) []RouteOutput {
		var output []RouteOutput
		for _, route := range osrmResp.Routes {
			balancas, err := s.InterfaceService.GetBalanca(ctx)
			if err != nil {
				log.Printf("Erro ao obter balanças: %v", err)
				balancas = nil
			}

			routeGasStations, err := s.findGasStations(dbCtx, osrmRespFast.Routes[0].Geometry)
			if err != nil {
				log.Printf("Erro ao consultar postos de gasolina: %v", err)
				routeGasStations = nil
			}

			distText, distVal := formatDistance(route.Distance)
			durText, durVal := formatDuration(route.Duration)

			var finalInstructions []Instruction
			if len(route.Legs) > 0 {
				for _, step := range route.Legs[0].Steps {
					text := translateInstruction(step)
					instructionLower := strings.ToLower(text)
					var valueImg string
					switch {
					case strings.Contains(instructionLower, "direita") && (strings.Contains(instructionLower, "curva") || strings.Contains(instructionLower, "mantenha-se")):
						valueImg = "https://plates-routes.s3.us-east-1.amazonaws.com/curva-direita.png"
					case strings.Contains(instructionLower, "esquerda") && (strings.Contains(instructionLower, "curva") || strings.Contains(instructionLower, "mantenha-se")):
						valueImg = "https://plates-routes.s3.us-east-1.amazonaws.com/curva-esquerda.png"
					case strings.Contains(instructionLower, "esquerda") && !strings.Contains(instructionLower, "curva"):
						valueImg = "https://plates-routes.s3.us-east-1.amazonaws.com/esquerda.png"
					case strings.Contains(instructionLower, "direita") && !strings.Contains(instructionLower, "curva"):
						valueImg = "https://plates-routes.s3.us-east-1.amazonaws.com/direita.png"
					case strings.Contains(instructionLower, "continue"), strings.Contains(instructionLower, "siga"), strings.Contains(instructionLower, "pegue"), strings.Contains(instructionLower, "fusão"), strings.Contains(instructionLower, "inicie"):
						valueImg = "https://plates-routes.s3.us-east-1.amazonaws.com/reto.png"
					case strings.Contains(instructionLower, "rotatória"), strings.Contains(instructionLower, "rotatoria"), strings.Contains(instructionLower, "retorno"):
						valueImg = "https://plates-routes.s3.us-east-1.amazonaws.com/rotatoria.png"
					case strings.Contains(instructionLower, "voltar"), strings.Contains(instructionLower, "volta"):
						valueImg = "https://plates-routes.s3.us-east-1.amazonaws.com/voltar.png"
					case strings.Contains(instructionLower, "vire"):
						valueImg = "https://plates-routes.s3.us-east-1.amazonaws.com/direita.png"
					default:
						valueImg = ""
					}

					finalInstructions = append(finalInstructions, Instruction{
						Text: text,
						Img:  valueImg,
					})
				}
			}

			rawTolls, err := s.findTollsOnRoute(dbCtx, route.Geometry, frontInfo.Type, float64(frontInfo.Axles))
			if err != nil {
				log.Printf("Erro ao filtrar pedágios: %v", err)
				rawTolls = nil
			}
			var routeTolls []Toll
			for _, t := range rawTolls {
				routeTolls = append(routeTolls, t)
			}

			routeBalancas, err := s.findBalancaOnRoute(route.Geometry, balancas)
			if err != nil {
				log.Printf("Erro ao filtrar balanças: %v", err)
				routeBalancas = nil
			}

			kmValue := route.Distance / 1000.0
			freight, err := s.getAllFreight(dbCtx, frontInfo.Axles, kmValue)
			if err != nil {
				log.Printf("Erro ao calcular freight: %v", err)
				freight = nil
			}

			routeType := routeCategory
			var totalTollCost float64
			for _, toll := range rawTolls {
				totalTollCost += toll.CashCost
			}

			fuelCostCity := math.Round((frontInfo.Price / frontInfo.ConsumptionCity) * (float64(distVal) / 1000))
			fuelCostHwy := math.Round((frontInfo.Price / frontInfo.ConsumptionHwy) * (float64(distVal) / 1000))

			avgConsumption := (frontInfo.ConsumptionCity + frontInfo.ConsumptionHwy) / 2
			totalKm := float64(distVal) / 1000
			totalFuelCost := math.Round((frontInfo.Price / avgConsumption) * totalKm)

			output = append(output, RouteOutput{
				Summary: RouteSummary{
					RouteType: routeType,
					HasTolls:  len(routeTolls) > 0,
					Distance: Distance{
						Text:  distText,
						Value: distVal,
					},
					Duration: Duration{
						Text:  durText,
						Value: durVal,
					},
					URL:           googleURL,
					URLWaze:       wazeURL,
					TotalFuelCost: totalFuelCost,
				},
				Costs: func() *Costs {
					if frontInfo.RouteOptions.IncludeTollCosts {
						return &Costs{
							TagAndCash:      totalTollCost,
							FuelInTheCity:   fuelCostCity,
							FuelInTheHwy:    fuelCostHwy,
							Tag:             (totalTollCost - (totalTollCost * 0.05)) * float64(frontInfo.Axles),
							Cash:            totalTollCost * float64(frontInfo.Axles),
							PrepaidCard:     totalTollCost * float64(frontInfo.Axles),
							MaximumTollCost: totalTollCost * float64(frontInfo.Axles),
							MinimumTollCost: totalTollCost * float64(frontInfo.Axles),
							Axles:           int(frontInfo.Axles),
						}
					}
					return nil
				}(),

				Tolls: func() []Toll {
					if frontInfo.RouteOptions.IncludeTollCosts {
						return routeTolls
					}
					return nil
				}(),

				Balances: func() interface{} {
					if frontInfo.RouteOptions.IncludeWeighStations {
						return routeBalancas
					}
					return nil
				}(),

				GasStations: func() []GasStation {
					if frontInfo.RouteOptions.IncludeFuelStations {
						return routeGasStations
					}
					return nil
				}(),

				Instructions: func() []Instruction {
					if frontInfo.RouteOptions.IncludeRouteMap {
						return finalInstructions
					}
					return nil
				}(),

				FreightLoad: func() map[string]interface{} {
					if frontInfo.RouteOptions.IncludeFreightCalc {
						return freight
					}
					return nil
				}(),

				Polyline: func() string {
					if frontInfo.RouteOptions.IncludePolyline {
						return route.Geometry
					}
					return ""
				}(),
			})
		}

		sort.Slice(output, func(i, j int) bool {
			return len(output[i].Tolls) < len(output[j].Tolls)
		})
		return output
	}

	if isAllRouteOptionsDisabled(frontInfo.RouteOptions) {
		var osrmRoute OSRMResponse
		if len(osrmRespEfficient.Routes) > 0 {
			osrmRoute = osrmRespEfficient
		} else if len(osrmRespFast.Routes) > 0 {
			osrmRoute = osrmRespFast
		} else if len(osrmRespNoTolls.Routes) > 0 {
			osrmRoute = osrmRespNoTolls
		} else {
			return FinalOutput{}, fmt.Errorf("nenhuma rota disponível para retorno mínimo")
		}

		route := osrmRoute.Routes[0]
		distText, distVal := formatDistance(route.Distance)
		durText, durVal := formatDuration(route.Duration)

		avgConsumption := (frontInfo.ConsumptionCity + frontInfo.ConsumptionHwy) / 2
		totalKm := float64(distVal) / 1000
		totalFuelCost := math.Round((frontInfo.Price / avgConsumption) * totalKm)

		minimalRoute := RouteOutput{
			Summary: RouteSummary{
				RouteType:     "efficient",
				HasTolls:      false,
				Distance:      Distance{Text: distText, Value: distVal},
				Duration:      Duration{Text: durText, Value: durVal},
				URL:           googleURL,
				URLWaze:       wazeURL,
				TotalFuelCost: totalFuelCost,
			},
		}

		finalOutput := FinalOutput{
			Summary: Summary{
				LocationOrigin: AddressInfo{
					Location: Location{
						Latitude:  origin.Location.Latitude,
						Longitude: origin.Location.Longitude,
					},
					Address: origin.FormattedAddress,
				},
				LocationDestination: AddressInfo{
					Location: Location{
						Latitude:  destination.Location.Latitude,
						Longitude: destination.Location.Longitude,
					},
					Address: destination.FormattedAddress,
				},
				AllStoppingPoints: func() []interface{} {
					var stops []interface{}
					for _, wp := range waypointResults {
						stops = append(stops, wp)
					}
					return stops
				}(),
				FuelPrice:      fuelPrice,
				FuelEfficiency: fuelEfficiency,
			},
			Routes: []RouteOutput{minimalRoute},
		}

		responseJSON, _ := json.Marshal(finalOutput)
		requestJSON, _ := json.Marshal(frontInfo)
		var wpStrings []string
		for _, wp := range frontInfo.Waypoints {
			wpStrings = append(wpStrings, fmt.Sprintf("%s,%s", wp.Lat, wp.Lng))
		}
		waypointsStr := strings.ToLower(strings.Join(wpStrings, ","))

		result, errSavedRoutes := s.savedRoutes(ctx, frontInfo.PublicOrPrivate,
			origin.FormattedAddress, destination.FormattedAddress,
			waypointsStr, idPublicToken, idSimp, responseJSON, requestJSON, frontInfo.Favorite)
		if errSavedRoutes != nil {
			return FinalOutput{}, errSavedRoutes
		}
		finalOutput.Summary.RouteHistID = result

		return finalOutput, nil
	}

	routesFast := processRoutes(osrmRespFast, "fatest")
	routesNoTolls := processRoutes(osrmRespNoTolls, "cheapest")
	routesEfficient := processRoutes(osrmRespEfficient, "efficient")

	var combinedRoutes []RouteOutput
	switch strings.ToLower(frontInfo.TypeRoute) {
	case "efficient", "eficiente":
		if len(routesEfficient) > 0 {
			combinedRoutes = []RouteOutput{routesEfficient[0]}
		}
	case "fatest", "fast", "rapida":
		if len(routesFast) > 0 {
			combinedRoutes = []RouteOutput{routesFast[0]}
		}
	case "cheapest", "cheap", "barata":
		if len(routesNoTolls) > 0 {
			combinedRoutes = []RouteOutput{routesNoTolls[0]}
		}
	default:
		combinedRoutes = append(append(routesFast, routesNoTolls...), routesEfficient...)
	}

	finalOutput := FinalOutput{
		Summary: Summary{
			LocationOrigin: AddressInfo{
				Location: Location{
					Latitude:  origin.Location.Latitude,
					Longitude: origin.Location.Longitude,
				},
				Address: origin.FormattedAddress,
			},
			LocationDestination: AddressInfo{
				Location: Location{
					Latitude:  destination.Location.Latitude,
					Longitude: destination.Location.Longitude,
				},
				Address: destination.FormattedAddress,
			},
			AllStoppingPoints: func() []interface{} {
				var stops []interface{}
				for _, wp := range waypointResults {
					stops = append(stops, wp)
				}
				return stops
			}(),
			FuelPrice:      fuelPrice,
			FuelEfficiency: fuelEfficiency,
			RouteOptions:   frontInfo.RouteOptions,
		},
		Routes: combinedRoutes,
	}

	if data, err := json.Marshal(finalOutput); err == nil {
		if err := cache.Rdb.Set(ctx, cacheKey, data, 10000*24*time.Hour).Err(); err != nil {
			log.Printf("Erro ao salvar cache do Redis (CalculateRoutes): %v", err)
		}
	}

	var wpStringsResponse []string
	for _, wp := range frontInfo.Waypoints {
		wpStringsResponse = append(wpStringsResponse, fmt.Sprintf("%s,%s", wp.Lat, wp.Lng))
	}
	waypointsStrResponse := strings.ToLower(strings.Join(wpStringsResponse, ","))

	responseJSON, _ := json.Marshal(finalOutput)
	requestJSON, _ := json.Marshal(frontInfo)

	result, errSavedRoutes := s.savedRoutes(ctx, frontInfo.PublicOrPrivate,
		origin.FormattedAddress, destination.FormattedAddress,
		waypointsStrResponse, idPublicToken, idSimp, responseJSON, requestJSON, frontInfo.Favorite)
	if errSavedRoutes != nil {
		return FinalOutput{}, errSavedRoutes
	}
	finalOutput.Summary.RouteHistID = result

	return finalOutput, nil
}

func (s *Service) GetFavoriteRouteService(ctx context.Context, id int64) ([]FavoriteRouteResponse, error) {
	result, err := s.InterfaceService.GetFavoriteByUserId(ctx, id)
	if err != nil {
		return []FavoriteRouteResponse{}, err
	}

	var getAllFavoriteRoute []FavoriteRouteResponse
	for _, trailer := range result {
		getFavoriteRouteResponse := FavoriteRouteResponse{}
		getFavoriteRouteResponse.ParseFromFavoriteRouteObject(trailer)
		getAllFavoriteRoute = append(getAllFavoriteRoute, getFavoriteRouteResponse)
	}

	return getAllFavoriteRoute, nil
}

func (s *Service) RemoveFavoriteRouteService(ctx context.Context, id, idUser int64) error {
	err := s.InterfaceService.RemoveFavorite(ctx, db.RemoveFavoriteParams{
		ID:     id,
		IDUser: idUser,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) GetSimpleRoute(data SimpleRouteRequest) (SimpleRouteResponse, error) {
	coords := fmt.Sprintf("%f,%f;%f,%f", data.OriginLng, data.OriginLat, data.DestLng, data.DestLat)
	baseOSRMURL := "http://34.207.174.233:5001/route/v1/driving/" + url.PathEscape(coords)
	client := http.Client{Timeout: 120 * time.Second}

	osrmURL := baseOSRMURL + "?" + url.Values{
		"alternatives": {"false"},
		"steps":        {"false"},
		"overview":     {"full"},
	}.Encode()

	resp, err := client.Get(osrmURL)
	if err != nil {
		return SimpleRouteResponse{}, fmt.Errorf("erro na requisição OSRM: %w", err)
	}
	defer resp.Body.Close()

	var osrmResp OSRMResponse
	if err := json.NewDecoder(resp.Body).Decode(&osrmResp); err != nil {
		return SimpleRouteResponse{}, fmt.Errorf("erro ao decodificar resposta OSRM: %w", err)
	}

	if osrmResp.Code != "Ok" || len(osrmResp.Routes) == 0 {
		return SimpleRouteResponse{}, fmt.Errorf("OSRM retornou erro ou nenhuma rota encontrada")
	}

	distanceText, distanceValue := formatDistance(osrmResp.Routes[0].Distance)
	durationText, durationValue := formatDuration(osrmResp.Routes[0].Duration)

	originAddress, err := s.reverseGeocode(data.OriginLat, data.OriginLng)
	if err != nil {
		return SimpleRouteResponse{}, fmt.Errorf("erro ao obter endereço da origem: %w", err)
	}
	destinationAddress, err := s.reverseGeocode(data.DestLat, data.DestLng)
	if err != nil {
		return SimpleRouteResponse{}, fmt.Errorf("erro ao obter endereço do destino: %w", err)
	}

	output := SimpleRouteResponse{
		Summary: SimpleSummary{
			LocationOrigin: AddressInfo{
				Location: Location{
					Latitude:  data.OriginLat,
					Longitude: data.OriginLng,
				},
				Address: originAddress,
			},
			LocationDestination: AddressInfo{
				Location: Location{
					Latitude:  data.DestLat,
					Longitude: data.DestLng,
				},
				Address: destinationAddress,
			},
			SimpleRoute: SimpleRouteSummary{
				Distance: Distance{
					Text:  distanceText,
					Value: distanceValue,
				},
				Duration: Duration{
					Text:  durationText,
					Value: durationValue,
				},
			},
		},
	}

	return output, nil
}

func (s *Service) reverseGeocode(lat, lng float64) (string, error) {
	geocodeURL := fmt.Sprintf("https://nominatim.openstreetmap.org/reverse?format=json&lat=%f&lon=%f", lat, lng)
	client := http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get(geocodeURL)
	if err != nil {
		return "", fmt.Errorf("erro na requisição de geocodificação reversa: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		DisplayName string `json:"display_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("erro ao decodificar resposta de geocodificação reversa: %w", err)
	}

	return result.DisplayName, nil
}

func (s *Service) savedRoutes(ctx context.Context, PublicOrPrivate, origin, destination, waypoints string, idPublicToken, IdUser int64, responseJSON, requestJSON json.RawMessage, favorite bool) (int64, error) {
	var idTokenHist int64
	if strings.ToLower(PublicOrPrivate) == "public" {
		idTokenHist = idPublicToken
	} else {
		idTokenHist = IdUser
	}

	isPublic := strings.ToLower(PublicOrPrivate) == "public"

	var routeHistID int64
	existingRoute, err := s.InterfaceService.GetRouteHistByUnique(ctx, db.GetRouteHistByUniqueParams{
		IDUser:      idTokenHist,
		Origin:      origin,
		Destination: destination,
		Waypoints: sql.NullString{
			String: waypoints,
			Valid:  true,
		},
		IsPublic: isPublic,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			newRouteHist, err := s.InterfaceService.CreateRouteHist(ctx, db.CreateRouteHistParams{
				IDUser:      idTokenHist,
				Origin:      origin,
				Destination: destination,
				Waypoints: sql.NullString{
					String: waypoints,
					Valid:  true,
				},
				Response:      responseJSON,
				IsPublic:      isPublic,
				NumberRequest: 1,
			})
			if err != nil {
				return 0, err
			}
			routeHistID = newRouteHist.ID
		} else {
			return 0, err
		}
	} else {
		newCount := existingRoute.NumberRequest + 1
		err = s.InterfaceService.UpdateNumberOfRequestRequest(ctx, db.UpdateNumberOfRequestParams{
			ID:            existingRoute.ID,
			NumberRequest: newCount,
		})
		if err != nil {
			return 0, err
		}
		routeHistID = existingRoute.ID
	}

	_, err = s.InterfaceService.CreateSavedRoutes(ctx, db.CreateSavedRoutesParams{
		Origin:      origin,
		Destination: destination,
		Waypoints: sql.NullString{
			String: waypoints,
			Valid:  true,
		},
		Request:   requestJSON,
		Response:  responseJSON,
		ExpiredAt: time.Now().Add(30 * 24 * time.Hour),
	})
	if err != nil {
		if !strings.Contains(err.Error(), `duplicate key value violates unique constraint "idx_saved_routes_unique"`) {
			return 0, err
		}
	}

	if favorite {
		_, err := s.InterfaceService.CreateFavoriteRoute(ctx, db.CreateFavoriteRouteParams{
			IDUser:      idTokenHist,
			Origin:      origin,
			Destination: destination,
			Waypoints: sql.NullString{
				String: waypoints,
				Valid:  true,
			},
			Response: responseJSON,
		})
		if err != nil {
			log.Printf("Erro ao salvar rota favorita: %v", err)
		}
	}

	return routeHistID, nil
}

func (s *Service) getAllFreight(ctx context.Context, axles int64, kmValue float64) (map[string]interface{}, error) {
	results, err := s.InterfaceService.GetFreightLoadAll(ctx)
	if err != nil {
		return nil, err
	}

	grouped := make(map[string][]FreightLoad)
	for _, result := range results {
		var fl FreightLoad
		fl.ParseFromFreightObject(result)
		grouped[fl.Name] = append(grouped[fl.Name], fl)
	}

	finalResult := make(map[string]interface{})
	for tableName, loads := range grouped {
		simplifiedLoads := make([]map[string]interface{}, 0)
		for _, fl := range loads {
			var rateStr string
			switch axles {
			case 2:
				rateStr = fl.TwoAxes
			case 3:
				rateStr = fl.ThreeAxes
			case 4:
				rateStr = fl.FourAxes
			case 5:
				rateStr = fl.FiveAxes
			case 6:
				rateStr = fl.SixAxes
			case 7:
				rateStr = fl.SevenAxes
			case 8:
				rateStr = fl.SevenAxes
			case 9:
				rateStr = fl.NineAxes
			default:
				rateStr = fl.TwoAxes
			}

			rateStr = strings.Replace(rateStr, ",", ".", -1)
			rate, err := strconv.ParseFloat(rateStr, 64)
			if err != nil {
				rate = 0
			}

			totalValue := kmValue * rate

			simplifiedLoad := map[string]interface{}{
				"description":  fl.Description,
				"type_of_load": fl.TypeOfLoad,
				"qtd_axle":     axles,
				"total_value":  totalValue,
			}
			simplifiedLoads = append(simplifiedLoads, simplifiedLoad)
		}
		finalResult[tableName] = simplifiedLoads
	}
	return finalResult, nil
}

func (s *Service) findTollsOnRoute(ctx context.Context, routeGeometry string, vehicle string, axes float64) ([]Toll, error) {
	tollsDB, err := s.InterfaceService.GetTollsByLonAndLat(ctx)
	if err != nil {
		return nil, err
	}

	resultTags, err := s.InterfaceService.GetTollTags(ctx)
	if err != nil {
		return nil, err
	}

	polyPoints, err := decodePolyline(routeGeometry)
	if err != nil {
		return nil, err
	}
	if len(polyPoints) < 2 {
		return nil, nil
	}

	routeIsCrescente := polyPoints[0].Lat < polyPoints[len(polyPoints)-1].Lat
	uniqueTolls := make(map[int64]bool)
	var candidateTolls []Toll

	for _, toll := range tollsDB {
		latitude, _ := validation.ParseNullStringToFloat(toll.Latitude)
		longitude, _ := validation.ParseNullStringToFloat(toll.Longitude)
		tollPos := LatLng{Lat: latitude, Lng: longitude}
		minDistance := math.MaxFloat64

		for i := 0; i < len(polyPoints)-1; i++ {
			d := distancePointToSegment(tollPos, polyPoints[i], polyPoints[i+1])
			if d < minDistance {
				minDistance = d
			}
		}

		if minDistance > 50 {
			continue
		}

		if toll.Sentido.String != "" {
			if toll.Sentido.String == "Crescente" && !routeIsCrescente {
				continue
			}
			if toll.Sentido.String == "Decrescente" && routeIsCrescente {
				continue
			}
		}

		imgConcession := getConcessionImage(toll.Concessionaria.String)
		if !uniqueTolls[toll.ID] {
			uniqueTolls[toll.ID] = true

			candidateTolls = append(candidateTolls, Toll{
				ID:            int(toll.ID),
				Latitude:      latitude,
				Longitude:     longitude,
				Name:          validation.GetStringFromNull(toll.PracaDePedagio),
				Concession:    toll.Concessionaria.String,
				ConcessionImg: imgConcession,
				Road:          validation.GetStringFromNull(toll.Rodovia),
				State:         validation.GetStringFromNull(toll.Uf),
				Country:       "Brasil",
				Type:          "Pedágio",
				FreeFlow:      toll.FreeFlow.Bool,
				PayFreeFlow:   toll.PayFreeFlow.String,
			})
		}
	}

	for i := range candidateTolls {
		var correspondingToll db.Toll
		found := false
		for _, t := range tollsDB {
			if int(t.ID) == candidateTolls[i].ID {
				correspondingToll = t
				found = true
				break
			}
		}
		if !found {
			continue
		}

		tarifaFloat := 0.0
		if correspondingToll.Tarifa.Valid {
			tf, _ := strconv.ParseFloat(correspondingToll.Tarifa.String, 64)
			tarifaFloat = tf
		}
		totalToll, errPrice := PriceTollsFromVehicle(strings.ToLower(vehicle), tarifaFloat, axes)
		if errPrice != nil {
			continue
		}

		var tags []string
		concession := validation.GetStringFromNull(correspondingToll.Concessionaria)
		for _, tagRecord := range resultTags {
			acceptedList := strings.Split(tagRecord.DealershipAccepts, ",")
			for _, accepted := range acceptedList {
				if strings.TrimSpace(accepted) == concession {
					tags = append(tags, tagRecord.Name)
					break
				}
			}
		}

		var imgTags []string
		for _, tag := range tags {
			var imgTag string
			switch tag {
			case "veloe":
				imgTag = "https://tags-tolls.s3.us-east-1.amazonaws.com/veloe.png"
			case "semParar":
				imgTag = "https://tags-tolls.s3.us-east-1.amazonaws.com/semparar.png"
			case "moveMais":
				imgTag = "https://tags-tolls.s3.us-east-1.amazonaws.com/moveMais.png"
			case "greenPass":
				imgTag = "https://tags-tolls.s3.us-east-1.amazonaws.com/greenpass.png"
			case "ecotaggy":
				imgTag = "https://tags-tolls.s3.us-east-1.amazonaws.com/ecotaggy.png"
			case "autoExpresso":
				imgTag = "https://tags-tolls.s3.us-east-1.amazonaws.com/auto-expresso.png"
			case "c6Taggy":
				imgTag = "https://tags-tolls.s3.us-east-1.amazonaws.com/c6-tag.png"
			case "dBTrans":
				imgTag = "https://tags-tolls.s3.us-east-1.amazonaws.com/dbTrans.png"
			case "taggy":
				imgTag = "https://tags-tolls.s3.us-east-1.amazonaws.com/taggy.png"
			case "conectCar":
				imgTag = "https://tags-tolls.s3.us-east-1.amazonaws.com/conectcar.png"
			default:
				imgTag = ""
			}
			imgTags = append(imgTags, imgTag)
		}

		candidateTolls[i].TagCost = math.Round(totalToll - (totalToll * 0.05))
		candidateTolls[i].CashCost = totalToll
		candidateTolls[i].Currency = "BRL"
		candidateTolls[i].PrepaidCardCost = math.Round(totalToll - (totalToll * 0.05))
		candidateTolls[i].TagPrimary = tags
		candidateTolls[i].TagImg = imgTags
	}

	originStr := fmt.Sprintf("%.6f,%.6f", polyPoints[0].Lat, polyPoints[0].Lng)
	arrivalMap, _ := s.calculateTollsArrivalTimes(originStr, candidateTolls)
	for i := range candidateTolls {
		if arr, ok := arrivalMap[int64(candidateTolls[i].ID)]; ok {
			formattedTime := arr.Time.Round(time.Second).String()
			candidateTolls[i].ArrivalResponse = ArrivalResponse{
				Distance: arr.Distance,
				Time:     formattedTime,
			}
		}
	}

	sort.Slice(candidateTolls, func(i, j int) bool {
		di := haversineDistanceTolls(candidateTolls[i].Latitude, candidateTolls[i].Longitude, polyPoints[0].Lat, polyPoints[0].Lng)
		dj := haversineDistanceTolls(candidateTolls[j].Latitude, candidateTolls[j].Longitude, polyPoints[0].Lat, polyPoints[0].Lng)
		return di < dj
	})

	return candidateTolls, nil
}

func (s *Service) findBalancaOnRoute(routeGeometry string, balancas []db.Balanca) ([]db.Balanca, error) {
	polyPoints, err := decodePolyline(routeGeometry)
	if err != nil {
		return nil, err
	}
	if len(polyPoints) < 2 {
		return nil, nil
	}

	routeIsCrescente := polyPoints[0].Lat < polyPoints[len(polyPoints)-1].Lat
	var foundBalancas []db.Balanca

	for _, b := range balancas {
		latitude, _ := validation.ParseStringToFloat(b.Lat)
		longitude, _ := validation.ParseStringToFloat(b.Lng)
		pos := LatLng{Lat: latitude, Lng: longitude}
		minDistance := math.MaxFloat64

		for i := 0; i < len(polyPoints)-1; i++ {
			d := distancePointToSegment(pos, polyPoints[i], polyPoints[i+1])
			if d < minDistance {
				minDistance = d
			}
		}

		if minDistance > 50 {
			continue
		}

		if b.Sentido != "" {
			if b.Sentido == "Crescente" && !routeIsCrescente {
				continue
			}
			if b.Sentido == "Decrescente" && routeIsCrescente {
				continue
			}
		}

		foundBalancas = append(foundBalancas, b)
	}

	return foundBalancas, nil
}

func (s *Service) findGasStations(ctx context.Context, routeGeometry string) ([]GasStation, error) {
	points, err := decodePolyline(routeGeometry)
	if err != nil {
		return nil, err
	}
	if len(points) == 0 {
		return nil, nil
	}

	minLat, minLng := points[0].Lat, points[0].Lng
	maxLat, maxLng := points[0].Lat, points[0].Lng
	for _, p := range points[1:] {
		if p.Lat < minLat {
			minLat = p.Lat
		}
		if p.Lat > maxLat {
			maxLat = p.Lat
		}
		if p.Lng < minLng {
			minLng = p.Lng
		}
		if p.Lng > maxLng {
			maxLng = p.Lng
		}
	}
	padding := 0.05
	minLat -= padding
	minLng -= padding
	maxLat += padding
	maxLng += padding

	resultStations, err := s.InterfaceService.GetGasStationsByBoundingBox(ctx, db.GetGasStationsByBoundingBoxParams{
		Column1: minLat,
		Column2: maxLat,
		Column3: minLng,
		Column4: maxLng,
	})
	if err != nil {
		return nil, fmt.Errorf("erro ao consultar postos no banco: %w", err)
	}

	tolerance := 150.0
	var stations []GasStation
	for _, stationRow := range resultStations {
		gs := convertGasStation(db.GetGasStationRow(stationRow))
		stationPos := LatLng{
			Lat: gs.Location.Latitude,
			Lng: gs.Location.Longitude,
		}
		minDistance := math.MaxFloat64
		for i := 0; i < len(points)-1; i++ {
			d := distancePointToSegment(stationPos, points[i], points[i+1])
			if d < minDistance {
				minDistance = d
			}
		}
		if minDistance <= tolerance {
			stations = append(stations, gs)
		}
	}

	return stations, nil
}

func (s *Service) calculateTollsArrivalTimes(origin string, tolls []Toll) (map[int64]Arrival, error) {
	parts := strings.Split(origin, ",")
	if len(parts) != 2 {
		return nil, fmt.Errorf("origem inválida: %s", origin)
	}
	originLat, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter latitude da origem: %w", err)
	}
	originLng, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter longitude da origem: %w", err)
	}

	avgSpeed := 60.0

	arrivalMap := make(map[int64]Arrival)
	for _, toll := range tolls {
		distanceMeters := haversineDistanceTolls(originLat, originLng, toll.Latitude, toll.Longitude)
		distanceKm := distanceMeters / 1000.0
		estimatedTimeHours := distanceKm / avgSpeed
		estimatedDuration := time.Duration(estimatedTimeHours * float64(time.Hour))

		arrivalMap[int64(toll.ID)] = Arrival{
			Distance: fmt.Sprintf("%.2f km", distanceKm),
			Time:     estimatedDuration,
		}
	}

	return arrivalMap, nil
}

func (s *Service) getGeocodeAddress(ctx context.Context, address string) (GeocodeResult, error) {
	//address = StateToCapital(strings.ToLower(address))
	//cacheKey := fmt.Sprintf("geocode:%s", address)
	//cached, err := cache.Rdb.Get(cache.Ctx, cacheKey).Result()
	//if err == nil {
	//	var result GeocodeResult
	//	if json.Unmarshal([]byte(cached), &result) == nil {
	//		return result, nil
	//	}
	//} else if !errors.Is(err, redis.Nil) {
	//	fmt.Printf("Erro ao recuperar cache do Redis (geocode): %v\n", err)
	//}

	client, err := maps.NewClient(maps.WithAPIKey(s.GoogleMapsAPIKey))
	if err != nil {
		return GeocodeResult{}, fmt.Errorf("erro ao criar cliente Google Maps: %v", err)
	}

	autoCompleteReq := &maps.PlaceAutocompleteRequest{
		Input:    address,
		Location: &maps.LatLng{Lat: -14.2350, Lng: -51.9253},
		Radius:   1000000,
		Language: "pt-BR",
		Types:    "geocode",
	}
	autoCompleteResp, autoCompleteErr := client.PlaceAutocomplete(ctx, autoCompleteReq)
	if autoCompleteErr == nil && len(autoCompleteResp.Predictions) > 0 {
		address = autoCompleteResp.Predictions[0].Description
	} else if autoCompleteErr != nil {
		fmt.Printf("Erro no Autocomplete: %v\n", autoCompleteErr)
	}

	req := &maps.GeocodingRequest{
		Address: address,
		Region:  "br",
	}
	results, err := client.Geocode(ctx, req)
	if err != nil || len(results) == 0 {
		return GeocodeResult{}, fmt.Errorf("endereço não encontrado para: %s. Verifique se a pesquisa está escrita corretamente ou seja mais específico(Ex: %s, São Paulo)", address, address)
	}

	result := GeocodeResult{
		FormattedAddress: results[0].FormattedAddress,
		PlaceID:          results[0].PlaceID,
		Location: Location{
			Latitude:  results[0].Geometry.Location.Lat,
			Longitude: results[0].Geometry.Location.Lng,
		},
	}

	//data, err := json.Marshal(result)
	//if err == nil {
	//	if err := cache.Rdb.Set(cache.Ctx, cacheKey, data, 30*24*time.Hour).Err(); err != nil {
	//		fmt.Printf("Erro ao salvar cache do Redis (geocode): %v\n", err)
	//	}
	//}
	return result, nil
}

func (s *Service) updateNumberOfRequest(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("ID inválido")
	}

	result, err := s.InterfaceService.GetTokenHist(ctx, id)
	if err != nil {
		return fmt.Errorf("falha ao obter token do histórico: %w", err)
	}
	number := result.NumberRequest + 1
	if number > 2 {
		return errors.New("você atingiu o limite de requisições por dia")
	}
	err = s.InterfaceService.UpdateNumberOfRequest(ctx, db.UpdateNumberOfRequestParams{
		NumberRequest: number,
		ID:            id,
	})
	if err != nil {
		return fmt.Errorf("falha ao atualizar número de requisições: %w", err)
	}
	return nil
}

func convertGasStation(row db.GetGasStationRow) GasStation {
	latitude, _ := validation.ParseStringToFloat(row.Latitude)
	longitude, _ := validation.ParseStringToFloat(row.Longitude)
	return GasStation{
		Name:    row.Name,
		Address: row.AddressName,
		Location: Location{
			Latitude:  latitude,
			Longitude: longitude,
		},
	}
}

func (s *Service) getCoordByCEP(ctx context.Context, cep string) (lat float64, lon float64, error error) {
	cepData, err := s.InterfaceService.FindAddressByCEPNew(ctx, cep)
	if err != nil {
		return 0, 0, fmt.Errorf("erro na busca local: %w", err)
	}

	if cepData.ID.Valid && cepData.ID.Int64 > 0 {
		return cepData.Lat.Float64, cepData.Lon.Float64, nil
	}

	log.Printf("cep nao encontrado na nossa base, buscando por api %w", cep)
	url := "https://gateway.apibrasil.io/api/v2/cep/cep"
	bodyData := map[string]string{"cep": cep}
	bodyJSON, _ := json.Marshal(bodyData)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyJSON))
	if err != nil {
		return 0, 0, err
	}

	bearer := os.Getenv("BEARER_TOKEN")
	device := os.Getenv("DEVICE_TOKEN_CEP")
	req.Header.Set("Authorization", "Bearer "+bearer)
	req.Header.Set("DeviceToken", device)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, err
	}

	if resp.StatusCode != http.StatusOK {
		return 0, 0, fmt.Errorf("erro na APIBrasil: %s", string(body))
	}

	var apiResp APIBrasilResponse
	err = json.Unmarshal(body, &apiResp)
	if err != nil {
		return 0, 0, err
	}

	if apiResp.Error {
		return 0, 0, errors.New(apiResp.Message)
	}

	lat, _ = strconv.ParseFloat(apiResp.Response.CEP.Latitude, 64)
	long, _ := strconv.ParseFloat(apiResp.Response.CEP.Longitude, 64)

	return lat, long, nil
}

func getConcessionImage(concession string) string {
	switch concession {
	case "VIAPAULISTA":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/viapaulista.png"
	case "ROTA 116":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/rota_116.png"
	case "EPR VIAS DO CAFÉ":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/epr_vias_do_cafe.png"
	case "VIARONDON":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/viarondon.png"
	case "ROTA DO OESTE":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/rota_do_oeste.png"
	case "VIA ARAUCÁRIA":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/via_araucaria.png"
	case "VIA BRASIL MT-163":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/via_brasil_mt_163.png"
	case "MUNICIPAL":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/municipal.png"
	case "ROTA DE SANTA MARIA":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/rota_de_santa_maria.png"
	case "RODOANEL OESTE":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/ccr_rodoanel.png"
	case "CSG":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/csg.png"
	case "ROTA DAS BANDEIRAS":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/rota_das_bandeiras.png"
	case "CONCEF":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/concef.png"
	case "TRIUNFO":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/triunfo.png"
	case "ECO 050":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/eco50.png"
	case "AB NASCENTES DAS GERAIS":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/ab_nascentes_das_gerais.png"
	case "FLUMINENSE":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/fluminense.png"
	case "Associação Gleba Barreiro":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/associacao_gleba_barreiro.png"
	case "RODOVIA DO AÇO":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/rodovia_do_aco.png"
	case "ECO RIOMINAS":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/eco_riominas.png"
	case "CSG - Free Flow":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/csg.png"
	case "RODOVIAS DO TIETÊ":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/rodovias_do_tietÃª.png"
	case "ECO RODOVIAS":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/eco_rodovias.png"
	case "EPR TRIANGULO":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/epr-triangulo.png"
	case "VIA RIO":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/via_rio.png"
	case "WAY - 306":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/way_306.png"
	case "EPR SUL DE MINAS":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/epr_sul_de_minas.png"
	case "ECO101":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/eco101.png"
	case "ECO SUL":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/eco_sul.png"
	case "ROTA DO ATLÂNTICO":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/rota_do_atlÃ¢ntico.png"
	case "VIA BRASIL - MT-100":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/via_brasil___mt_100.png"
	case "ROTA DOS GRÃOS":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/rota_dos_graos.png"
	case "TRANSBRASILIANA":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/transbrasiliana.png"
	case "APASI":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/apasi.png"
	case "RODOVIA DA MUDANÇA":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/rodovia_da_mudanca.png"
	case "ENTREVIAS":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/entrevias.png"
	case "AB COLINAS":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/ab_colinas.png"
	case "CCR ViaLagos":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/ccr_vialagos.png"
	case "ROTA DOS COQUEIROS":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/rota_dos_coqueiros.png"
	case "CRP CONCESSIONARIA":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/crp_concessionaria.png"
	case "WAY - 112":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/way_112.png"
	case "EPR LITORAL PIONEIRO":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/epr_litoral_pioneiro.png"
	case "PLANALTO SUL":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/planalto_sul.png"
	case "CCR VIA COSTEIRA":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/ccr_via_costeira.png"
	case "LITORAL SUL":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/litoral_sul.png"
	case "SPVIAS":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/spvias.png"
	case "AUTOBAN":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/autoban.png"
	case "ECOVIAS DO CERRADO":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/ecovias_do_cerrado.png"
	case "EPR VIA MINEIRA":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/epr-via-mineira.png"
	case "SPMAR":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/spmar.png"
	case "JOTEC":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/jotec.png"
	case "VIA NORTE SUL":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/via_norte_sul.png"
	case "CONCER":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/concer.png"
	case "ECONOROESTE":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/econoroeste.png"
	case "ECOPONTE":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/ecoponte.png"
	case "ECO 135":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/eco_135.png"
	case "VIA BRASIL MT-246":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/via_brasil_mt_246.png"
	case "ECOVIAS DO ARAGUAIA":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/ecovias_do_araguaia.png"
	case "VIABAHIA":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/viabahia.png"
	case "GUARUJÁ":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/guaruja.png"
	case "CONCEBRA":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/truinfo_concebra.png"
	case "DER":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/der.png"
	case "EGR":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/egr.png"
	case "PREFEITURA DE ITIRAPINA":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/prefeitura_de_itirapina.png"
	case "VIA PAULISTA":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/via_paulista.png"
	case "CCR VIASUL":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/ccr_viasul.png"
	case "INTERVIAS":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/intervias.png"
	case "CCR MSVia":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/ccr_msvia.png"
	case "EIXO SP":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/eixo_sp.png"
	case "RÉGIS BITTENCOURT":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/regis_bittencourt.png"
	case "FERNÃO DIAS":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/fernao_dias.png"
	case "CART":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/cart.png"
	case "CCR RioSP":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/ccr-riosp.png"
	case "VIAOESTE":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/viaoeste.png"
	case "MORRO DA MESA":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/morro_da_mesa.png"
	case "TOMOIOS":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/tomoios.png"
	case "EPG Sul de Minas":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/epg_sul_de_minas.png"
	case "ECOPISTAS":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/ecopistas.png"
	case "LAMSA":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/lamsa.png"
	case "TEBE":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/tebe.png"
	case "BAHIA NORTE":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/bahia_norte.png"
	default:
		return ""
	}
}


//---- NEW FUNCTIONS 
func (s *Service) CalculateDistancesBetweenPointsWithRiskAvoidance(ctx context.Context, data FrontInfoCEPRequest) (Response, error) {
	if len(data.CEPs) < 2 {
		return Response{}, fmt.Errorf("é necessário pelo menos dois pontos para calcular distâncias")
	}

	client := http.Client{Timeout: 60 * time.Second}
	var resultRoutes []DetailedRoute
	var totalDistance float64
	var totalDuration float64

	// ====== NOVO: carregar zonas de risco e mapear p/ struct simples ======
	type RiskZone struct {
		Lat    float64
		Lon    float64
		Radius float64 // metros
	}

	dbZones, _ := s.RiskZonesRepository.GetAllZonasRiscoService(ctx) // ajuste conforme sua assinatura
	var zones []RiskZone
	for _, z := range dbZones {
		// TODO: ajuste nomes de campos conforme seu modelo real
		zlat := z.Lat       // ex.: z.Latitude
		zlon := z.Lng       // ex.: z.Longitude
		zrad := z.Radius   // ex.: z.RadiusMeters ou z.Raio
		zones = append(zones, RiskZone{Lat: zlat, Lon: zlon, Radius: float64(zrad)})
	}
	// ======================================================================

	// ====== Helpers geométricos locais ======
	const earthR = 6371000.0
	deg2rad := func(d float64) float64 { return d * math.Pi / 180 }
	rad2deg := func(r float64) float64 { return r * 180 / math.Pi }

	bearing := func(from, to Point) float64 {
		φ1, φ2 := deg2rad(from.Lat), deg2rad(to.Lat)
		λ1, λ2 := deg2rad(from.Lon), deg2rad(to.Lon)
		y := math.Sin(λ2-λ1) * math.Cos(φ2)
		x := math.Cos(φ1)*math.Sin(φ2) - math.Sin(φ1)*math.Cos(φ2)*math.Cos(λ2-λ1)
		θ := math.Atan2(y, x)
		return math.Mod(rad2deg(θ)+360.0, 360.0)
	}
	offsetPoint := func(p Point, bearingDeg, dist float64) Point {
		br := deg2rad(bearingDeg)
		lat1 := deg2rad(p.Lat)
		lon1 := deg2rad(p.Lon)
		lat2 := math.Asin(math.Sin(lat1)*math.Cos(dist/earthR) + math.Cos(lat1)*math.Sin(dist/earthR)*math.Cos(br))
		lon2 := lon1 + math.Atan2(math.Sin(br)*math.Sin(dist/earthR)*math.Cos(lat1),
			math.Cos(dist/earthR)-math.Sin(lat1)*math.Sin(lat2))
		return Point{Lat: rad2deg(lat2), Lon: rad2deg(lon2)}
	}
	segmentCircleHits := func(a, b Point, center Point, radiusM float64) bool {
		lat0 := deg2rad(center.Lat)
		proj := func(p Point) (x, y float64) {
			x = earthR * deg2rad(p.Lon) * math.Cos(lat0)
			y = earthR * deg2rad(p.Lat)
			return
		}
		ax, ay := proj(a)
		bx, by := proj(b)
		cx, cy := proj(center)

		abx, aby := bx-ax, by-ay
		apx, apy := cx-ax, cy-ay
		ab2 := abx*abx + aby*aby
		t := 0.0
		if ab2 > 0 {
			t = (apx*abx + apy*aby) / ab2
		}
		if t < 0 {
			t = 0
		} else if t > 1 {
			t = 1
		}
		nx := ax + t*abx
		ny := ay + t*aby
		dx, dy := nx-cx, ny-cy
		dist := math.Hypot(dx, dy)
		return dist <= radiusM
	}
	routeIntersectsZones := func(poly string, zs []RiskZone) (hit *RiskZone, err error) {
		pts, err := decodePolyline5(poly)
		if err != nil {
			return nil, err
		}
		for i := 0; i < len(pts)-1; i++ {
			for _, z := range zs {
				if segmentCircleHits(pts[i], pts[i+1], Point{Lat: z.Lat, Lon: z.Lon}, z.Radius) {
					return &z, nil
				}
			}
		}
		return nil, nil
	}
	tangentCandidates := func(z RiskZone, around Point, padding float64) (Point, Point) {
		dir := bearing(Point{Lat: z.Lat, Lon: z.Lon}, around)
		r := z.Radius + padding
		return offsetPoint(Point{Lat: z.Lat, Lon: z.Lon}, dir+90, r),
			offsetPoint(Point{Lat: z.Lat, Lon: z.Lon}, dir-90, r)
	}
	// ========================================

	for i := 0; i < len(data.CEPs)-1; i++ {
		originCEP := data.CEPs[i]
		destCEP := data.CEPs[i+1]

		originLat, originLon, err := s.getCoordByCEP(ctx, originCEP)
		if err != nil {
			return Response{}, fmt.Errorf("erro ao buscar coordenadas da origem %s: %w", originCEP, err)
		}
		destLat, destLon, err := s.getCoordByCEP(ctx, destCEP)
		if err != nil {
			return Response{}, fmt.Errorf("erro ao buscar coordenadas do destino %s: %w", destCEP, err)
		}
		originAddress, _ := s.reverseGeocode(originLat, originLon)
		destAddress, _ := s.reverseGeocode(destLat, destLon)

		originGeocode, _ := s.getGeocodeAddress(ctx, originAddress)
		destGeocode, _ := s.getGeocodeAddress(ctx, destAddress)

		// ====== alteramos a forma de montar a URL p/ permitir inserir vias ======
		osrmRoot := "http://34.207.174.233:5001/route/v1/driving/"
		origin := Point{Lat: originLat, Lon: originLon}
		dest := Point{Lat: destLat, Lon: destLon}
		// =======================================================================

		type osrmResult struct {
			resp     OSRMResponse
			category string
			err      error
		}
		resultsCh := make(chan osrmResult, 3)

		// helper para chamar OSRM com coords dinâmicos e params
		callOSRM := func(params url.Values, coords []Point) (OSRMResponse, error) {
			parts := make([]string, 0, len(coords))
			for _, p := range coords {
				parts = append(parts, fmt.Sprintf("%f,%f", p.Lon, p.Lat))
			}
			// força geometries=polyline para bater com decodePolyline5
			if params.Get("geometries") == "" {
				params.Set("geometries", "polyline")
			}
			fullURL := osrmRoot + url.PathEscape(strings.Join(parts, ";")) + "?" + params.Encode()

			resp, err := client.Get(fullURL)
			if err != nil {
				return OSRMResponse{}, err
			}
			defer resp.Body.Close()

			var osrmResp OSRMResponse
			if err := json.NewDecoder(resp.Body).Decode(&osrmResp); err != nil {
				return OSRMResponse{}, err
			}
			if osrmResp.Code != "Ok" || len(osrmResp.Routes) == 0 {
				return OSRMResponse{}, fmt.Errorf("OSRM retornou %s", osrmResp.Code)
			}
			return osrmResp, nil
		}

		// ====== NOVO: faz a requisição e tenta desviar das zonas se necessário ======
		makeSafeRequest := func(params url.Values, category string) {
			// 1) rota direta
			routeResp, err := callOSRM(params, []Point{origin, dest})
			if err == nil {
				if hz, _ := routeIntersectsZones(routeResp.Routes[0].Geometry, zones); hz == nil {
					resultsCh <- osrmResult{resp: routeResp, category: category, err: nil}
					return
				}
			}

			// 2) tentar vias tangentes
			padding := 200.0 // metros
			const maxTries = 3
			var final OSRMResponse
			var finalErr error = fmt.Errorf("não foi possível evitar zonas")

			for attempt := 0; attempt < maxTries; attempt++ {
				// usa a última zona detectada na rota direta para gerar candidatos;
				// se a rota direta falhou na chamada, vamos usar o ponto médio entre origin e dest
				var z RiskZone
				if err == nil {
					if hz, _ := routeIntersectsZones(routeResp.Routes[0].Geometry, zones); hz != nil {
						z = *hz
					}
				} else if len(zones) > 0 {
					z = zones[0] // fallback simples
				} else {
					break
				}

				v1, v2 := tangentCandidates(z, origin, padding)

				// origin -> v1 -> dest
				if r1, e1 := callOSRM(params, []Point{origin, v1, dest}); e1 == nil {
					if hit, _ := routeIntersectsZones(r1.Routes[0].Geometry, zones); hit == nil {
						final = r1
						finalErr = nil
					}
				}

				// origin -> v2 -> dest (se ainda não achou final ou se v2 for melhor)
				if r2, e2 := callOSRM(params, []Point{origin, v2, dest}); e2 == nil {
					if hit, _ := routeIntersectsZones(r2.Routes[0].Geometry, zones); hit == nil {
						if finalErr != nil || r2.Routes[0].Distance < final.Routes[0].Distance {
							final = r2
							finalErr = nil
						}
					}
				}

				if finalErr == nil {
					break
				}
				padding *= 1.6 // abre mais a curva
			}

			if finalErr != nil {
				resultsCh <- osrmResult{err: finalErr, category: category}
				return
			}
			resultsCh <- osrmResult{resp: final, category: category}
		}
		// ==========================================================================

		var routeTypes []string
		if strings.TrimSpace(strings.ToLower(data.TypeRoute)) == "" {
			go makeSafeRequest(url.Values{
				"alternatives":      {"3"},
				"steps":             {"true"},
				"overview":          {"full"},
				"continue_straight": {"false"},
			}, "fastest")
			go makeSafeRequest(url.Values{
				"alternatives": {"3"},
				"steps":        {"true"},
				"overview":     {"full"},
				"exclude":      {"toll"},
			}, "cheapest")
			go makeSafeRequest(url.Values{
				"alternatives": {"3"},
				"steps":        {"true"},
				"overview":     {"full"},
				"exclude":      {"motorway"},
			}, "efficient")
			routeTypes = []string{"fastest", "cheapest", "efficient"}
		} else {
			routeTypes = []string{strings.ToLower(data.TypeRoute)}
			switch routeTypes[0] {
			case "rapida", "fastest":
				makeSafeRequest(url.Values{
					"alternatives":      {"3"},
					"steps":             {"true"},
					"overview":          {"full"},
					"continue_straight": {"false"},
				}, "fastest")
			case "barata", "cheapest":
				makeSafeRequest(url.Values{
					"alternatives": {"3"},
					"steps":        {"true"},
					"overview":     {"full"},
					"exclude":      {"toll"},
				}, "cheapest")
			case "eficiente", "efficient":
				makeSafeRequest(url.Values{
					"alternatives": {"3"},
					"steps":        {"true"},
					"overview":     {"full"},
					"exclude":      {"motorway"},
				}, "efficient")
			}
		}

		var summaries []RouteSummary

		for range routeTypes {
			res := <-resultsCh
			if res.err != nil || len(res.resp.Routes) == 0 {
				continue
			}

			route := res.resp.Routes[0]
			distText, distVal := formatDistance(route.Distance)
			durText, durVal := formatDuration(route.Duration)
			avgConsumption := (data.ConsumptionCity + data.ConsumptionHwy) / 2
			totalKm := route.Distance / 1000
			totalFuelCost := math.Round((data.Price / avgConsumption) * totalKm)

			googleURL := fmt.Sprintf("https://www.google.com/maps/dir/?api=1&origin=%s&destination=%s",
				neturl.QueryEscape(originGeocode.FormattedAddress),
				neturl.QueryEscape(destGeocode.FormattedAddress),
			)
			currentTimeMillis := (time.Now().UnixNano() + int64(route.Duration*float64(time.Second))) / int64(time.Millisecond)
			wazeURL := fmt.Sprintf("https://www.waze.com/pt-BR/live-map/directions/br?to=place.%s&from=place.%s&time=%d&reverse=yes",
				neturl.QueryEscape(destGeocode.PlaceID),
				neturl.QueryEscape(originGeocode.PlaceID),
				currentTimeMillis,
			)

			rawTolls, err := s.findTollsOnRoute(ctx, route.Geometry, data.Type, float64(data.Axles))
			if err != nil {
				log.Printf("Erro ao filtrar pedágios: %v", err)
			}

			var routeTolls []Toll
			for _, t := range rawTolls {
				routeTolls = append(routeTolls, t)
			}

			var totalTollCost float64
			for _, toll := range routeTolls {
				totalTollCost += toll.CashCost
			}

			summaries = append(summaries, RouteSummary{
				RouteType:     res.category,
				HasTolls:      len(routeTolls) > 0,
				Distance:      Distance{Text: distText, Value: distVal},
				Duration:      Duration{Text: durText, Value: durVal},
				URL:           googleURL,
				URLWaze:       wazeURL,
				TotalFuelCost: totalFuelCost,
				Tolls:         routeTolls,
				TotalTolls:    math.Round(totalTollCost*100) / 100,
				Polyline:      route.Geometry, // já é a rota "segura"
			})

			totalDistance += route.Distance
			totalDuration += route.Duration
		}

		resultRoutes = append(resultRoutes, DetailedRoute{
			LocationOrigin: AddressInfo{
				Location: Location{Latitude: originLat, Longitude: originLon},
				Address:  originGeocode.FormattedAddress,
			},
			LocationDestination: AddressInfo{
				Location: Location{Latitude: destLat, Longitude: destLon},
				Address:  destGeocode.FormattedAddress,
			},
			Summaries: summaries,
		})
	}

	// (mantive sua lógica do TotalRoute; se quiser, dá para montar um "total safe polyline"
	// concatenando as polylines dos legs após decodificar e re-encodar, mas deixei como estava.)
	var totalRoute TotalSummary
	var allCoords []string
	var waypoints []string
	var originLocation, destinationLocation Location
	for idx, cep := range data.CEPs {
		coordLat, coordLon, err := s.getCoordByCEP(ctx, cep)
		if err != nil {
			return Response{}, fmt.Errorf("erro ao buscar coordenadas para total_route no CEP %s: %w", cep, err)
		}
		allCoords = append(allCoords, fmt.Sprintf("%f,%f", coordLon, coordLat))

		reverse, _ := s.reverseGeocode(coordLat, coordLon)
		geocode, _ := s.getGeocodeAddress(ctx, reverse)
		waypoints = append(waypoints, geocode.FormattedAddress)

		if idx == 0 {
			originLocation = Location{Latitude: coordLat, Longitude: coordLon}
		}
		if idx == len(data.CEPs)-1 {
			destinationLocation = Location{Latitude: coordLat, Longitude: coordLon}
		}
	}

	coordsStr := strings.Join(allCoords, ";")
	urlTotal := fmt.Sprintf("http://34.207.174.233:5001/route/v1/driving/%s?alternatives=0&steps=true&overview=full&continue_straight=false", url.PathEscape(coordsStr))

	resp, err := client.Get(urlTotal)
	if err == nil {
		defer resp.Body.Close()
		var osrmResp OSRMResponse
		if err := json.NewDecoder(resp.Body).Decode(&osrmResp); err == nil && len(osrmResp.Routes) > 0 {
			route := osrmResp.Routes[0]

			distText, distVal := formatDistance(totalDistance)
			durText, durVal := formatDuration(totalDuration)

			avgConsumption := (data.ConsumptionCity + data.ConsumptionHwy) / 2
			totalKm := route.Distance / 1000
			totalFuelCost := math.Round((data.Price / avgConsumption) * totalKm)

			tolls, _ := s.findTollsOnRoute(ctx, route.Geometry, data.Type, float64(data.Axles))
			var totalTollCost float64
			for _, toll := range tolls {
				totalTollCost += toll.CashCost
			}

			originAddress := waypoints[0]
			destAddress := waypoints[len(waypoints)-1]
			waypointStr := ""
			if len(waypoints) > 2 {
				waypointStr = "&waypoints=" + neturl.QueryEscape(strings.Join(waypoints[1:len(waypoints)-1], "|"))
			}

			googleURL := fmt.Sprintf("https://www.google.com/maps/dir/?api=1&origin=%s&destination=%s%s&travelmode=driving",
				neturl.QueryEscape(originAddress),
				neturl.QueryEscape(destAddress),
				waypointStr,
			)

			currentTimeMillis := (time.Now().UnixNano() + int64(route.Duration*float64(time.Second))) / int64(time.Millisecond)
			wazeURL := fmt.Sprintf("https://www.waze.com/pt-BR/live-map/directions/br?to=%s&from=%s&time=%d&reverse=yes",
				neturl.QueryEscape(destAddress),
				neturl.QueryEscape(originAddress),
				currentTimeMillis,
			)

			totalRoute = TotalSummary{
				LocationOrigin: AddressInfo{
					Location: originLocation,
					Address:  originAddress,
				},
				LocationDestination: AddressInfo{
					Location: destinationLocation,
					Address:  destAddress,
				},
				TotalDistance: Distance{Text: distText, Value: distVal},
				TotalDuration: Duration{Text: durText, Value: durVal},
				URL:           googleURL,
				URLWaze:       wazeURL,
				Tolls:         tolls,
				TotalTolls:    math.Round(totalTollCost*100) / 100,
				Polyline:      route.Geometry,
				TotalFuelCost: totalFuelCost,
			}
		}
	}

	return Response{
		Routes:     resultRoutes,
		TotalRoute: totalRoute,
	}, nil
}

const R = 6371000.0 

func deg2rad(d float64) float64 { return d * math.Pi / 180 }
func rad2deg(r float64) float64 { return r * 180 / math.Pi }

func offsetPoint(p Point, bearingDeg, dist float64) Point {
	br := deg2rad(bearingDeg)
	lat1 := deg2rad(p.Lat)
	lon1 := deg2rad(p.Lon)
	lat2 := math.Asin(math.Sin(lat1)*math.Cos(dist/R) + math.Cos(lat1)*math.Sin(dist/R)*math.Cos(br))
	lon2 := lon1 + math.Atan2(math.Sin(br)*math.Sin(dist/R)*math.Cos(lat1),
		math.Cos(dist/R)-math.Sin(lat1)*math.Sin(lat2))
	return Point{Lat: rad2deg(lat2), Lon: rad2deg(lon2)}
}

func bearing(from, to Point) float64 {
	φ1, φ2 := deg2rad(from.Lat), deg2rad(to.Lat)
	λ1, λ2 := deg2rad(from.Lon), deg2rad(to.Lon)
	y := math.Sin(λ2-λ1) * math.Cos(φ2)
	x := math.Cos(φ1)*math.Sin(φ2) - math.Sin(φ1)*math.Cos(φ2)*math.Cos(λ2-λ1)
	θ := math.Atan2(y, x)
	b := math.Mod(rad2deg(θ)+360.0, 360.0)
	return b
}

func segmentCircleHits(a, b, center Point, radiusM float64) bool {
	lat0 := deg2rad(center.Lat)
	// projeta p/ metros
	proj := func(p Point) (x, y float64) {
		x = R * deg2rad(p.Lon) * math.Cos(lat0)
		y = R * deg2rad(p.Lat)
		return
	}
	ax, ay := proj(a)
	bx, by := proj(b)
	cx, cy := proj(center)

	abx, aby := bx-ax, by-ay
	apx, apy := cx-ax, cy-ay
	ab2 := abx*abx + aby*aby
	t := 0.0
	if ab2 > 0 {
		t = (apx*abx + apy*aby) / ab2
	}
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	// ponto mais próximo
	nx := ax + t*abx
	ny := ay + t*aby
	dx, dy := nx-cx, ny-cy
	dist := math.Hypot(dx, dy)
	return dist <= radiusM
}

func routeIntersectsZones(polyline string, zones []RiskZone) (*RiskZone, error) {
	pts, err := decodePolyline5(polyline)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(pts)-1; i++ {
		segA := pts[i]
		segB := pts[i+1]
		for _, z := range zones {
			if segmentCircleHits(segA, segB, Point{z.Lat, z.Lon}, z.Radius) {
				return &z, nil
			}
		}
	}
	return nil, nil
}

func tangentCandidates(z RiskZone, around Point, padding float64) (Point, Point) {
	dir := bearing(Point{z.Lat, z.Lon}, around)
	r := z.Radius + padding
	return offsetPoint(Point{z.Lat, z.Lon}, dir+90, r),
		offsetPoint(Point{z.Lat, z.Lon}, dir-90, r)
}

func osrmRoute(baseURL string, coords []Point) (OSRMResponse, error) {
	parts := make([]string, 0, len(coords))
	for _, p := range coords {
		parts = append(parts, fmt.Sprintf("%f,%f", p.Lon, p.Lat))
	}
	u := fmt.Sprintf("%s/route/v1/driving/%s?alternatives=3&steps=true&overview=full",
		baseURL, url.PathEscape(strings.Join(parts, ";")))
	c := http.Client{Timeout: 30 * time.Second}
	resp, err := c.Get(u)
	if err != nil {
		return OSRMResponse{}, err
	}
	defer resp.Body.Close()
	var out OSRMResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return OSRMResponse{}, err
	}
	if out.Code != "Ok" || len(out.Routes) == 0 {
		return OSRMResponse{}, fmt.Errorf("OSRM código %s", out.Code)
	}
	return out, nil
}

func EnsureSafeLeg(baseURL string, origin, dest Point, zones []RiskZone) (OSRMResponse, error) {
	const maxTries = 3
	padding := 200.0 

	route, err := osrmRoute(baseURL, []Point{origin, dest})
	if err != nil {
		return OSRMResponse{}, err
	}
	for i := 0; i < maxTries; i++ {
		if z, _ := routeIntersectsZones(route.Routes[0].Geometry, zones); z == nil {
			return route, nil
		} else {
			v1, v2 := tangentCandidates(*z, origin, padding)
			r1, e1 := osrmRoute(baseURL, []Point{origin, v1, dest})
			r2, e2 := osrmRoute(baseURL, []Point{origin, v2, dest})
			best := OSRMResponse{}
			ok1, ok2 := false, false
			if e1 == nil {
				if hit, _ := routeIntersectsZones(r1.Routes[0].Geometry, zones); hit == nil {
					best = r1
					ok1 = true
				}
			}
			if e2 == nil {
				if hit, _ := routeIntersectsZones(r2.Routes[0].Geometry, zones); hit == nil {
					if !ok1 || r2.Routes[0].Distance < best.Routes[0].Distance {
						best = r2
					}
					ok2 = true
				}
			}
			if ok1 || ok2 {
				return best, nil
			}
			// amplia o padding e tenta de novo
			padding *= 1.6
		}
	}
	return route, fmt.Errorf("não consegui evitar zonas após %d tentativas", maxTries)
}

func decodePolyline5(encoded string) ([]Point, error) {
	var points []Point
	var index, lat, lng int

	for index < len(encoded) {
		// decodifica latitude
		var result, shift int
		for {
			if index >= len(encoded) {
				return nil, fmt.Errorf("polyline truncada ao ler latitude")
			}
			b := int(encoded[index] - 63)
			index++
			result |= (b & 0x1f) << shift
			shift += 5
			if b < 0x20 {
				break
			}
		}
		dlat := ((result >> 1) ^ (-(result & 1)))
		lat += dlat

		// decodifica longitude
		result, shift = 0, 0
		for {
			if index >= len(encoded) {
				return nil, fmt.Errorf("polyline truncada ao ler longitude")
			}
			b := int(encoded[index] - 63)
			index++
			result |= (b & 0x1f) << shift
			shift += 5
			if b < 0x20 {
				break
			}
		}
		dlng := ((result >> 1) ^ (-(result & 1)))
		lng += dlng

		points = append(points, Point{
			Lat: float64(lat) / 1e5,
			Lon: float64(lng) / 1e5,
		})
	}
	return points, nil
}