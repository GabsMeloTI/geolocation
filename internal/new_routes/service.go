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
	CalculateDistancesBetweenPointsWithRiskAvoidance(ctx context.Context, data FrontInfoCEPRequest) (Response, error)
	CalculateDistancesFromOrigin(ctx context.Context, data FrontInfoCEPRequest) ([]DetailedRoute, error)
	CalculateRoutesWithCoordinate(ctx context.Context, frontInfo FrontInfoCoordinate, idPublicToken int64, idSimp int64) (FinalOutput, error)
	GetFavoriteRouteService(ctx context.Context, id int64) ([]FavoriteRouteResponse, error)
	RemoveFavoriteRouteService(ctx context.Context, id, idUser int64) error
	GetSimpleRoute(data SimpleRouteRequest) (SimpleRouteResponse, error)
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

func (s *Service) CalculateDistancesBetweenPointsWithRiskAvoidance(ctx context.Context, data FrontInfoCEPRequest) (Response, error) {
	log.Printf("🚀 INICIANDO CÁLCULO DE ROTA COM EVITAMENTO DE ZONAS DE RISCO")
	log.Printf("📍 CEPs: %v", data.CEPs)

	if len(data.CEPs) < 2 {
		return Response{}, fmt.Errorf("é necessário pelo menos dois pontos para calcular distâncias")
	}

	riskZones, err := s.getActiveRiskZones(ctx)
	if err != nil {
		log.Printf("❌ Erro ao buscar zonas de risco: %v", err)
		log.Printf("🔄 Fallback para função original sem desvios")
		return s.CalculateDistancesBetweenPoints(ctx, data)
	}

	log.Printf("📊 Total de zonas de risco encontradas: %d", len(riskZones))

	client := http.Client{Timeout: 60 * time.Second}
	var resultRoutes []DetailedRoute
	var totalDistance float64
	var totalDuration float64

	for i := 0; i < len(data.CEPs)-1; i++ {
		originCEP := data.CEPs[i]
		destCEP := data.CEPs[i+1]

		log.Printf("\n🔄 Processando segmento %d: %s → %s", i+1, originCEP, destCEP)

		originLat, originLon, err := s.getCoordByCEP(ctx, originCEP)
		if err != nil {
			return Response{}, fmt.Errorf("erro ao buscar coordenadas da origem %s: %w", originCEP, err)
		}
		destLat, destLon, err := s.getCoordByCEP(ctx, destCEP)
		if err != nil {
			return Response{}, fmt.Errorf("erro ao buscar coordenadas do destino %s: %w", destCEP, err)
		}

		log.Printf("📍 Coordenadas: origem(%.6f, %.6f) → destino(%.6f, %.6f)",
			originLat, originLon, destLat, destLon)

		originAddress, _ := s.reverseGeocode(originLat, originLon)
		destAddress, _ := s.reverseGeocode(destLat, destLon)

		originGeocode, _ := s.getGeocodeAddress(ctx, originAddress)
		destGeocode, _ := s.getGeocodeAddress(ctx, destAddress)

		// Verificar se há zonas de risco no caminho direto
		log.Printf("🔍 Verificando zonas de risco para este segmento...")
		directRouteHasRisk, locationHisk := s.CheckRouteForRiskZones(riskZones, originLat, originLon, destLat, destLon)

		var summaries []RouteSummary

		log.Printf("🎯 Resultado da verificação: directRouteHasRisk = %v", directRouteHasRisk)
		if directRouteHasRisk {
			log.Printf("🚨 ROTA TEM RISCO - Calculando desvio...")
			// Se há risco, calcular rota alternativa com desvio
			summaries = s.calculateAlternativeRouteWithAvoidance(ctx, client, riskZones, originLat, originLon, destLat, destLon, originGeocode, destGeocode, data)
		} else {
			log.Printf("✅ ROTA SEGURA - Usando rota direta...")
			// Se não há risco, usar rota direta
			summaries = s.calculateDirectRoute(ctx, client, originLat, originLon, destLat, destLon, originGeocode, destGeocode, data)
		}

		if len(summaries) == 0 {
			return Response{}, fmt.Errorf("não foi possível calcular rota entre %s e %s", originCEP, destCEP)
		}

		// Usar a primeira rota para cálculos totais
		route := summaries[0]
		totalDistance += route.Distance.Value
		totalDuration += route.Duration.Value

		log.Printf("📏 Segmento %d: %s, %s", i+1, route.Distance.Text, route.Duration.Text)

		resultRoutes = append(resultRoutes, DetailedRoute{
			LocationOrigin: AddressInfo{
				Location: Location{Latitude: originLat, Longitude: originLon},
				Address:  originGeocode.FormattedAddress,
			},
			LocationDestination: AddressInfo{
				Location: Location{Latitude: destLat, Longitude: destLon},
				Address:  destGeocode.FormattedAddress,
			},
			HasRisk:      directRouteHasRisk,
			LocationHisk: locationHisk,
			Summaries:    summaries,
		})
	}

	log.Printf("\n🔄 Calculando rota total com desvios...")
	// Calcular rota total com desvios
	totalRoute := s.calculateTotalRouteWithAvoidance(ctx, client, riskZones, data.CEPs, totalDistance, totalDuration, data)

	distText, _ := formatDistance(totalDistance)
	durText, _ := formatDuration(totalDuration)
	log.Printf("✅ CÁLCULO CONCLUÍDO - Total: %s, %s", distText, durText)

	return Response{
		Routes:     resultRoutes,
		TotalRoute: totalRoute,
	}, nil
}

func (s *Service) getActiveRiskZones(ctx context.Context) ([]RiskZone, error) {
	if s.RiskZonesRepository == nil {
		log.Printf("⚠️  RiskZonesRepository é nil - retornando lista vazia")
		return []RiskZone{}, nil
	}

	dbZones, err := s.RiskZonesRepository.GetAllZonasRiscoService(ctx)
	if err != nil {
		log.Printf("❌ Erro ao buscar zonas de risco: %v", err)
		return nil, fmt.Errorf("erro ao buscar zonas de risco: %w", err)
	}

	var riskZones []RiskZone
	for _, dbZone := range dbZones {
		riskZones = append(riskZones, RiskZone{
			ID:     dbZone.ID,
			Name:   dbZone.Name,
			Cep:    dbZone.Cep,
			Lat:    dbZone.Lat,
			Lng:    dbZone.Lng,
			Radius: dbZone.Radius,
			Status: dbZone.Status,
		})
	}

	log.Printf("✅ Buscou %d zonas de risco ativas", len(riskZones))
	for _, zone := range riskZones {
		log.Printf("   - %s: lat=%.6f, lng=%.6f, raio=%dm", zone.Name, zone.Lat, zone.Lng, zone.Radius)
	}

	return riskZones, nil
}

// RiskZone representa uma zona de risco
type RiskZone struct {
	ID     int64   `json:"id"`
	Name   string  `json:"name"`
	Cep    string  `json:"cep"`
	Lat    float64 `json:"lat"`
	Lng    float64 `json:"lng"`
	Radius int64   `json:"radius"`
	Status bool    `json:"status"`
}

// CheckRouteForRiskZones verifica se uma rota passa por zonas de risco
func (s *Service) CheckRouteForRiskZones(riskZones []RiskZone, originLat, originLon, destLat, destLon float64) (bool, LocationHisk) {
	log.Printf("🔍 Verificando rota: origem(%.6f, %.6f) → destino(%.6f, %.6f)", originLat, originLon, destLat, destLon)

	if len(riskZones) == 0 {
		log.Printf("ℹ️  Nenhuma zona de risco para verificar")
		return false, LocationHisk{}
	}

	// Primeiro, calcular a rota real com OSRM para verificar todos os pontos
	client := http.Client{Timeout: 30 * time.Second}
	coords := fmt.Sprintf("%f,%f;%f,%f", originLon, originLat, destLon, destLat)
	url := fmt.Sprintf("http://34.207.174.233:5001/route/v1/driving/%s?overview=full&steps=true", url.PathEscape(coords))

	log.Printf("🌐 Calculando rota real com OSRM: %s", url)

	resp, err := client.Get(url)
	if err != nil {
		log.Printf("❌ Erro ao calcular rota OSRM: %v", err)
		// Fallback para verificação de linha reta
		return s.checkRouteForRiskZonesFallback(riskZones, originLat, originLon, destLat, destLon), LocationHisk{}
	}
	defer resp.Body.Close()

	var osrmResp OSRMResponse
	if err := json.NewDecoder(resp.Body).Decode(&osrmResp); err != nil {
		log.Printf("❌ Erro ao decodificar resposta OSRM: %v", err)
		// Fallback para verificação de linha reta
		return s.checkRouteForRiskZonesFallback(riskZones, originLat, originLon, destLat, destLon), LocationHisk{}
	}

	if len(osrmResp.Routes) == 0 {
		log.Printf("❌ Nenhuma rota retornada pelo OSRM")
		// Fallback para verificação de linha reta
		return s.checkRouteForRiskZonesFallback(riskZones, originLat, originLon, destLat, destLon), LocationHisk{}
	}

	route := osrmResp.Routes[0]
	distText, _ := formatDistance(route.Distance)
	durText, _ := formatDuration(route.Duration)
	log.Printf("✅ Rota OSRM calculada: %s, %s", distText, durText)

	// Verificar se a rota real passa por zonas de risco
	return s.checkRouteGeometryForRiskZones(riskZones, route.Geometry, originLat, originLon, destLat, destLon)
}

// isPointInRiskZone verifica se um ponto está dentro de uma zona de risco
func (s *Service) isPointInRiskZone(lat, lng float64, zone RiskZone) bool {
	distance := s.haversineDistance(lat, lng, zone.Lat, zone.Lng)
	isInside := distance <= float64(zone.Radius)

	if isInside {
		log.Printf("   🎯 Ponto (%.6f, %.6f) está dentro da zona %s: distância=%.1fm, raio=%dm",
			lat, lng, zone.Name, distance, zone.Radius)
	} else {
		log.Printf("   ✅ Ponto (%.6f, %.6f) está fora da zona %s: distância=%.1fm, raio=%dm",
			lat, lng, zone.Name, distance, zone.Radius)
	}

	return isInside
}

// doesRouteCrossRiskZone verifica se uma rota cruza uma zona de risco
func (s *Service) doesRouteCrossRiskZone(originLat, originLon, destLat, destLon float64, zone RiskZone) bool {
	// Distância do centro da zona até a linha da rota
	distanceToRoute := s.distancePointToLine(zone.Lat, zone.Lng, originLat, originLon, destLat, destLon)

	// Se a distância for menor que o raio, a rota cruza a zona
	crossesZone := distanceToRoute <= float64(zone.Radius)

	if crossesZone {
		log.Printf("   🚨 Rota cruza zona %s: distância até rota=%.1fm, raio=%dm",
			zone.Name, distanceToRoute, zone.Radius)
	} else {
		log.Printf("   ✅ Rota não cruza zona %s: distância até rota=%.1fm, raio=%dm",
			zone.Name, distanceToRoute, zone.Radius)
	}

	return crossesZone
}

// distancePointToLine calcula a distância de um ponto até uma linha
func (s *Service) distancePointToLine(pointLat, pointLng, lineLat1, lineLng1, lineLat2, lineLng2 float64) float64 {
	// Fórmula para distância de ponto até linha
	A := pointLat - lineLat1
	B := pointLng - lineLng1
	C := lineLat2 - lineLat1
	D := lineLng2 - lineLng1

	dot := A*C + B*D
	lenSq := C*C + D*D

	if lenSq == 0 {
		return s.haversineDistance(pointLat, pointLng, lineLat1, lineLng1)
	}

	param := dot / lenSq

	var xx, yy float64
	if param < 0 {
		xx = lineLat1
		yy = lineLng1
	} else if param > 1 {
		xx = lineLat2
		yy = lineLng2
	} else {
		xx = lineLat1 + param*C
		yy = lineLng1 + param*D
	}

	return s.haversineDistance(pointLat, pointLng, xx, yy)
}

// haversineDistance calcula a distância entre dois pontos usando fórmula de Haversine
func (s *Service) haversineDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371000 // Raio da Terra em metros

	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLng := (lng2 - lng1) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLng/2)*math.Sin(deltaLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

// calculateAlternativeRouteWithAvoidance calcula rota alternativa evitando zonas de risco
func (s *Service) calculateAlternativeRouteWithAvoidance(ctx context.Context, client http.Client, riskZones []RiskZone, originLat, originLon, destLat, destLon float64, originGeocode, destGeocode GeocodeResult, data FrontInfoCEPRequest) []RouteSummary {
	log.Printf("🔄 Calculando rota alternativa com desvio...")
	var summaries []RouteSummary

	// Estratégia 1: Tentar rota com waypoints intermediários para desviar
	waypoints := s.generateAvoidanceWaypoints(riskZones, originLat, originLon, destLat, destLon)

	log.Printf("📍 Waypoints de desvio gerados: %d", len(waypoints))
	for i, wp := range waypoints {
		log.Printf("   %d. (%.6f, %.6f)", i+1, wp.Latitude, wp.Longitude)
	}

	if len(waypoints) > 0 {
		// Construir URL com waypoints de desvio
		coords := fmt.Sprintf("%f,%f", originLon, originLat)
		for _, wp := range waypoints {
			coords += fmt.Sprintf(";%f,%f", wp.Longitude, wp.Latitude)
		}
		coords += fmt.Sprintf(";%f,%f", destLon, destLat)

		baseURL := "http://34.207.174.233:5001/route/v1/driving/" + url.PathEscape(coords)
		log.Printf("🌐 URL OSRM com desvio: %s", baseURL)

		// Fazer requisição para rota com desvio
		resp, err := client.Get(baseURL + "?alternatives=1&steps=true&overview=full&continue_straight=false")
		if err == nil {
			defer resp.Body.Close()
			var osrmResp OSRMResponse
			if err := json.NewDecoder(resp.Body).Decode(&osrmResp); err == nil && len(osrmResp.Routes) > 0 {
				route := osrmResp.Routes[0]
				distText, _ := formatDistance(route.Distance)
				durText, _ := formatDuration(route.Duration)
				log.Printf("✅ Rota alternativa calculada: %s, %s", distText, durText)
				summary := s.createRouteSummary(route, "desvio_seguro", originGeocode, destGeocode, data, nil)
				summaries = append(summaries, summary)
			} else {
				log.Printf("❌ Erro ao decodificar resposta OSRM: %v", err)
			}
		} else {
			log.Printf("❌ Erro na requisição OSRM: %v", err)
		}
	}

	// Se não conseguiu rota alternativa, tentar rota direta com aviso
	if len(summaries) == 0 {
		log.Printf("⚠️  Não foi possível calcular rota alternativa, criando aviso de risco...")
		summary := s.createRouteSummaryWithRiskWarning(originLat, originLon, destLat, destLon, originGeocode, destGeocode, data, riskZones)
		summaries = append(summaries, summary)
	}

	log.Printf("📊 Total de rotas alternativas: %d", len(summaries))
	return summaries
}

// generateAvoidanceWaypoints gera waypoints para desviar de zonas de risco
func (s *Service) generateAvoidanceWaypoints(riskZones []RiskZone, originLat, originLon, destLat, destLon float64) []Location {
	log.Printf("🔄 Gerando waypoints de desvio...")
	var waypoints []Location

	// 1) Verificação baseada na geometria real da rota (OSRM)
	hasRisk, locationHisk := s.CheckRouteForRiskZones(riskZones, originLat, originLon, destLat, destLon)
	if hasRisk {
		// Encontrar a zona correspondente pelo CEP ou por proximidade
		if zone, ok := s.findRiskZoneByHint(riskZones, locationHisk); ok {
			log.Printf("🚨 Rota passa pela zona %s - gerando waypoints de desvio", zone.Name)
			// Gerar DOIS waypoints tangenciais (entrada e saída) para forçar o contorno
			wpA, wpB := s.computeBypassWaypoints(originLat, originLon, destLat, destLon, zone)
			waypoints = append(waypoints, wpA, wpB)
			log.Printf("📍 Waypoint A: (%.6f, %.6f)", wpA.Latitude, wpA.Longitude)
			log.Printf("📍 Waypoint B: (%.6f, %.6f)", wpB.Latitude, wpB.Longitude)
		}
	}

	// 2) Fallback: verificação pela linha direta entre origem e destino (menos preciso)
	if len(waypoints) == 0 {
		for _, zone := range riskZones {
			if !zone.Status {
				log.Printf("⏭️  Zona %s está inativa, pulando", zone.Name)
				continue
			}

			log.Printf("🔍 (fallback) Verificando se reta cruza zona %s", zone.Name)
			if s.doesRouteCrossRiskZone(originLat, originLon, destLat, destLon, zone) {
				log.Printf("🚨 (fallback) Reta cruza zona %s - gerando waypoints", zone.Name)
				wpA, wpB := s.computeBypassWaypoints(originLat, originLon, destLat, destLon, zone)
				waypoints = append(waypoints, wpA, wpB)
				log.Printf("📍 Waypoint A (reta): (%.6f, %.6f)", wpA.Latitude, wpA.Longitude)
				log.Printf("📍 Waypoint B (reta): (%.6f, %.6f)", wpB.Latitude, wpB.Longitude)
			}
		}
	}

	log.Printf("📊 Total de waypoints de desvio gerados: %d", len(waypoints))
	return waypoints
}

// findRiskZoneByHint localiza a RiskZone com base no CEP ou pela proximidade com o centro da zona
func (s *Service) findRiskZoneByHint(riskZones []RiskZone, hint LocationHisk) (RiskZone, bool) {
	// 1) Tentar casar por CEP
	if hint.CEP != "" {
		for _, z := range riskZones {
			if strings.EqualFold(strings.TrimSpace(z.Cep), strings.TrimSpace(hint.CEP)) {
				return z, true
			}
		}
	}
	// 2) Escolher a mais próxima do hint
	var best RiskZone
	bestDist := math.MaxFloat64
	for _, z := range riskZones {
		d := s.haversineDistance(hint.Latitude, hint.Longitude, z.Lat, z.Lng)
		if d < bestDist {
			bestDist = d
			best = z
		}
	}
	if best.ID != 0 {
		return best, true
	}
	return RiskZone{}, false
}

// computeBypassWaypoints retorna dois waypoints tangenciais para contornar a zona
func (s *Service) computeBypassWaypoints(originLat, originLon, destLat, destLon float64, zone RiskZone) (Location, Location) {
	// Baseado no perpendicular calculado em metros, cria dois pontos a uma distância segura
	// e ordena-os pelo avanço ao longo da rota (proximidade da origem)

	// Reutiliza a lógica de cálculo do perpendicular para obter vetores em metros
	midLat := (originLat + destLat) / 2.0
	metersPerDegLat := 111320.0
	metersPerDegLon := 111320.0 * math.Cos(midLat*math.Pi/180.0)

	dNorth := (destLat - originLat) * metersPerDegLat
	dEast := (destLon - originLon) * metersPerDegLon
	pNorth, pEast := -dEast, dNorth
	mag := math.Hypot(pNorth, pEast)
	if mag == 0 {
		pNorth, pEast, mag = 0, 1, 1
	}
	pNorth /= mag
	pEast /= mag

	safe := float64(zone.Radius) + 1200.0 // 1.2km para maior margem

	// Ponto no eixo do centro da zona alinhado à rota (projeção aproximada)
	baseNorth := (zone.Lat - midLat) * metersPerDegLat
	baseEast := (zone.Lng - ((originLon + destLon) / 2.0)) * metersPerDegLon

	// Dois pontos tangenciais ao redor do centro
	aNorth := baseNorth + pNorth*safe
	aEast := baseEast + pEast*safe
	bNorth := baseNorth - pNorth*safe
	bEast := baseEast - pEast*safe

	aLat := midLat + aNorth/metersPerDegLat
	aLon := ((originLon + destLon) / 2.0) + aEast/metersPerDegLon
	bLat := midLat + bNorth/metersPerDegLat
	bLon := ((originLon + destLon) / 2.0) + bEast/metersPerDegLon

	wpA := Location{Latitude: aLat, Longitude: aLon}
	wpB := Location{Latitude: bLat, Longitude: bLon}

	// Ordenar pelo avanço ao longo da rota (distância até a origem)
	dA := s.haversineDistance(originLat, originLon, wpA.Latitude, wpA.Longitude)
	dB := s.haversineDistance(originLat, originLon, wpB.Latitude, wpB.Longitude)
	if dA <= dB {
		return wpA, wpB
	}
	return wpB, wpA
}

// calculateAvoidancePoint calcula um ponto de desvio para evitar zona de risco
func (s *Service) calculateAvoidancePoint(originLat, originLon, destLat, destLon float64, zone RiskZone) Location {
	log.Printf("🔄 Calculando ponto de desvio para zona %s", zone.Name)

	// Vetor da rota em graus
	dLatDeg := destLat - originLat
	dLonDeg := destLon - originLon

	// Converter graus -> metros (aproximação local)
	midLat := (originLat + destLat) / 2.0
	metersPerDegLat := 111320.0
	metersPerDegLon := 111320.0 * math.Cos(midLat*math.Pi/180.0)

	dNorth := dLatDeg * metersPerDegLat // metros para norte
	dEast := dLonDeg * metersPerDegLon  // metros para leste

	// Vetor perpendicular em metros (90 graus)
	// perpendicular para a direita: (-dNorth, dEast)
	pNorth := -dEast
	pEast := dNorth

	// Normalizar
	mag := math.Hypot(pNorth, pEast)
	if mag == 0 {
		// rota degenerada, empurra para leste 1km
		pNorth, pEast = 0, 1
		mag = 1
	}
	pNorth /= mag
	pEast /= mag

	// Distância segura: raio + 1km
	safeDistanceMeters := float64(zone.Radius) + 1000.0
	log.Printf("   📏 Distância de segurança: %.0fm (raio: %dm + 1000m)", safeDistanceMeters, zone.Radius)

	// Candidato A (lado +) e B (lado -)
	candANorth := pNorth * safeDistanceMeters
	candAEast := pEast * safeDistanceMeters
	candBNorth := -candANorth
	candBEast := -candAEast

	// Aplicar ao ponto médio em metros e converter de volta para graus
	midLon := (originLon + destLon) / 2.0
	candALat := midLat + (candANorth / metersPerDegLat)
	candALon := midLon + (candAEast / metersPerDegLon)
	candBLat := midLat + (candBNorth / metersPerDegLat)
	candBLon := midLon + (candBEast / metersPerDegLon)

	// Escolher o candidato mais distante do centro da zona
	distA := s.haversineDistance(candALat, candALon, zone.Lat, zone.Lng)
	distB := s.haversineDistance(candBLat, candBLon, zone.Lat, zone.Lng)

	var chosen Location
	if distA >= distB {
		chosen = Location{Latitude: candALat, Longitude: candALon}
	} else {
		chosen = Location{Latitude: candBLat, Longitude: candBLon}
	}

	log.Printf("   📍 Ponto médio da rota: (%.6f, %.6f)", midLat, midLon)
	log.Printf("   🎯 Ponto de desvio escolhido: (%.6f, %.6f) — dist %.1fm do centro",
		chosen.Latitude, chosen.Longitude,
		s.haversineDistance(chosen.Latitude, chosen.Longitude, zone.Lat, zone.Lng))

	return chosen
}

// checkRouteForRiskZonesFallback verificação de fallback usando linha reta
func (s *Service) checkRouteForRiskZonesFallback(riskZones []RiskZone, originLat, originLon, destLat, destLon float64) bool {
	log.Printf("🔄 Usando verificação de fallback (linha reta)")

	for _, zone := range riskZones {
		if !zone.Status {
			continue
		}

		log.Printf("🔍 Verificando zona: %s (lat=%.6f, lng=%.6f, raio=%dm)", zone.Name, zone.Lat, zone.Lng, zone.Radius)

		// Verificar se a origem ou destino estão dentro da zona de risco
		originInZone := s.isPointInRiskZone(originLat, originLon, zone)
		destInZone := s.isPointInRiskZone(destLat, destLon, zone)

		if originInZone || destInZone {
			log.Printf("🚨 PONTO DENTRO DA ZONA DE RISCO: %s", zone.Name)
			return true
		}

		// Verificar se a linha reta entre origem e destino cruza a zona de risco
		routeCrossesZone := s.doesRouteCrossRiskZone(originLat, originLon, destLat, destLon, zone)
		if routeCrossesZone {
			log.Printf("🚨 ROTA CRUZA ZONA DE RISCO: %s", zone.Name)
			return true
		}

		log.Printf("✅ Zona %s não representa risco para esta rota", zone.Name)
	}

	log.Printf("✅ Nenhuma zona de risco encontrada para esta rota")
	return false
}

// checkRouteGeometryForRiskZones verifica se a geometria da rota OSRM passa por zonas de risco
func (s *Service) checkRouteGeometryForRiskZones(riskZones []RiskZone, geometry string, originLat, originLon, destLat, destLon float64) (bool, LocationHisk) {
	log.Printf("🔍 Verificando geometria da rota OSRM para zonas de risco")

	// Decodificar a geometria Polyline do OSRM
	coordinates, err := s.decodePolylineOSRM(geometry)
	if err != nil {
		log.Printf("❌ Erro ao decodificar geometria: %v", err)
		return s.checkRouteForRiskZonesFallback(riskZones, originLat, originLon, destLat, destLon), LocationHisk{}
	}

	log.Printf("📍 Rota OSRM tem %d pontos para verificar", len(coordinates))

	// Verificar cada segmento da rota
	for i := 0; i < len(coordinates)-1; i++ {
		point1 := coordinates[i]
		point2 := coordinates[i+1]

		// Verificar se este segmento passa por alguma zona de risco
		for _, zone := range riskZones {
			if !zone.Status {
				continue
			}

			// Verificar se algum dos pontos do segmento está dentro da zona
			point1InZone := s.isPointInRiskZone(point1.Latitude, point1.Longitude, zone)
			point2InZone := s.isPointInRiskZone(point2.Latitude, point2.Longitude, zone)

			if point1InZone || point2InZone {
				log.Printf("🚨 PONTO DA ROTA DENTRO DA ZONA DE RISCO: %s", zone.Name)
				if point1InZone {
					log.Printf("   - Ponto %d (%.6f, %.6f) está dentro da zona", i+1, point1.Latitude, point1.Longitude)
				}
				if point2InZone {
					log.Printf("   - Ponto %d (%.6f, %.6f) está dentro da zona", i+1, point2.Latitude, point2.Longitude)
				}
				return true, LocationHisk{
					CEP:       zone.Cep,
					Latitude:  zone.Lat,
					Longitude: zone.Lng,
				}
			}

			// Verificar se o segmento cruza a zona de risco
			segmentCrossesZone := s.doesRouteCrossRiskZone(point1.Latitude, point1.Longitude, point2.Latitude, point2.Longitude, zone)
			if segmentCrossesZone {
				log.Printf("🚨 SEGMENTO DA ROTA CRUZA ZONA DE RISCO: %s", zone.Name)
				log.Printf("   - Segmento %d: (%.6f, %.6f) → (%.6f, %.6f)",
					i+1, point1.Latitude, point1.Longitude, point2.Latitude, point2.Longitude)
				return true, LocationHisk{
					CEP:       zone.Cep,
					Latitude:  zone.Lat,
					Longitude: zone.Lng,
				}
			}
		}
	}

	log.Printf("✅ Nenhuma zona de risco encontrada na rota OSRM")
	return false, LocationHisk{}
}

// decodePolylineOSRM decodifica a geometria Polyline do OSRM
func (s *Service) decodePolylineOSRM(encoded string) ([]Location, error) {
	// Usar a função existente do helper.go
	latLngPoints, err := decodePolyline(encoded)
	if err != nil {
		return nil, err
	}

	var coordinates []Location
	for _, point := range latLngPoints {
		coordinates = append(coordinates, Location{
			Latitude:  point.Lat,
			Longitude: point.Lng,
		})
	}

	return coordinates, nil
}

// createRouteSummary cria um resumo de rota
func (s *Service) createRouteSummary(route OSRMRoute, routeType string, originGeocode, destGeocode GeocodeResult, data FrontInfoCEPRequest, tolls []Toll) RouteSummary {
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

	var totalTollCost float64
	if tolls != nil {
		for _, toll := range tolls {
			totalTollCost += toll.CashCost
		}
	}

	return RouteSummary{
		RouteType:     routeType,
		HasTolls:      tolls != nil && len(tolls) > 0,
		Distance:      Distance{Text: distText, Value: distVal},
		Duration:      Duration{Text: durText, Value: durVal},
		URL:           googleURL,
		URLWaze:       wazeURL,
		TotalFuelCost: totalFuelCost,
		Tolls:         tolls,
		TotalTolls:    math.Round(totalTollCost*100) / 100,
		Polyline:      route.Geometry,
	}
}

// createRouteSummaryWithRiskWarning cria resumo de rota com aviso de risco
func (s *Service) createRouteSummaryWithRiskWarning(originLat, originLon, destLat, destLon float64, originGeocode, destGeocode GeocodeResult, data FrontInfoCEPRequest, riskZones []RiskZone) RouteSummary {
	// Calcular distância direta
	distance := s.haversineDistance(originLat, originLon, destLat, destLon)
	// Estimativa de tempo baseada na velocidade média (60 km/h)
	duration := distance / 16.67 // 60 km/h = 16.67 m/s

	distText, distVal := formatDistance(distance)
	durText, durVal := formatDuration(duration)

	avgConsumption := (data.ConsumptionCity + data.ConsumptionHwy) / 2
	totalKm := distance / 1000
	totalFuelCost := math.Round((data.Price / avgConsumption) * totalKm)

	googleURL := fmt.Sprintf("https://www.google.com/maps/dir/?api=1&origin=%s&destination=%s",
		neturl.QueryEscape(originGeocode.FormattedAddress),
		neturl.QueryEscape(destGeocode.FormattedAddress),
	)

	wazeURL := fmt.Sprintf("https://www.waze.com/pt-BR/live-map/directions/br?to=place.%s&from=place.%s&reverse=yes",
		neturl.QueryEscape(destGeocode.PlaceID),
		neturl.QueryEscape(originGeocode.PlaceID),
	)

	return RouteSummary{
		RouteType:     "rota_direta_com_aviso",
		HasTolls:      false,
		Distance:      Distance{Text: distText, Value: distVal},
		Duration:      Duration{Text: durText, Value: durVal},
		URL:           googleURL,
		URLWaze:       wazeURL,
		TotalFuelCost: totalFuelCost,
		Tolls:         nil,
		TotalTolls:    0,
		Polyline:      "", // Sem polyline para rota direta
	}
}

// calculateDirectRoute calcula rota direta sem desvios
func (s *Service) calculateDirectRoute(ctx context.Context, client http.Client, originLat, originLon, destLat, destLon float64, originGeocode, destGeocode GeocodeResult, data FrontInfoCEPRequest) []RouteSummary {
	coords := fmt.Sprintf("%f,%f;%f,%f",
		originLon, originLat,
		destLon, destLat,
	)
	baseURL := "http://34.207.174.233:5001/route/v1/driving/" + url.PathEscape(coords)

	var summaries []RouteSummary

	// Fazer requisição para rota direta
	resp, err := client.Get(baseURL + "?alternatives=1&steps=true&overview=full&continue_straight=false")
	if err == nil {
		defer resp.Body.Close()
		var osrmResp OSRMResponse
		if err := json.NewDecoder(resp.Body).Decode(&osrmResp); err == nil && len(osrmResp.Routes) > 0 {
			route := osrmResp.Routes[0]

			// Buscar pedágios na rota
			tolls, _ := s.findTollsOnRoute(ctx, route.Geometry, data.Type, float64(data.Axles))

			summary := s.createRouteSummary(route, "rota_direta", originGeocode, destGeocode, data, tolls)
			summaries = append(summaries, summary)
		}
	}

	return summaries
}

// calculateTotalRouteWithAvoidance calcula rota total com desvios
func (s *Service) calculateTotalRouteWithAvoidance(ctx context.Context, client http.Client, riskZones []RiskZone, ceps []string, totalDistance, totalDuration float64, data FrontInfoCEPRequest) TotalSummary {
	var allCoords []string
	var waypoints []string
	var originLocation, destinationLocation Location

	for idx, cep := range ceps {
		coordLat, coordLon, err := s.getCoordByCEP(ctx, cep)
		if err != nil {
			continue
		}

		allCoords = append(allCoords, fmt.Sprintf("%f,%f", coordLon, coordLat))

		reverse, _ := s.reverseGeocode(coordLat, coordLon)
		geocode, _ := s.getGeocodeAddress(ctx, reverse)
		waypoints = append(waypoints, geocode.FormattedAddress)

		if idx == 0 {
			originLocation = Location{Latitude: coordLat, Longitude: coordLon}
		}
		if idx == len(ceps)-1 {
			destinationLocation = Location{Latitude: coordLat, Longitude: coordLon}
		}
	}

	// Verificar se há zonas de risco no caminho total
	hasRisk, _ := s.CheckRouteForRiskZones(riskZones, originLocation.Latitude, originLocation.Longitude, destinationLocation.Latitude, destinationLocation.Longitude)

	var totalRoute TotalSummary

	if hasRisk {
		// Gerar waypoints de desvio para rota total
		avoidanceWaypoints := s.generateAvoidanceWaypoints(riskZones, originLocation.Latitude, originLocation.Longitude, destinationLocation.Latitude, destinationLocation.Longitude)

		if len(avoidanceWaypoints) > 0 {
			// Construir URL com waypoints de desvio
			coordsStr := allCoords[0]
			for _, wp := range avoidanceWaypoints {
				coordsStr += fmt.Sprintf(";%f,%f", wp.Longitude, wp.Latitude)
			}
			coordsStr += ";" + allCoords[len(allCoords)-1]

			urlTotal := fmt.Sprintf("http://34.207.174.233:5001/route/v1/driving/%s?alternatives=0&steps=true&overview=full&continue_straight=false", url.PathEscape(coordsStr))

			resp, err := client.Get(urlTotal)
			if err == nil {
				defer resp.Body.Close()
				var osrmResp OSRMResponse
				if err := json.NewDecoder(resp.Body).Decode(&osrmResp); err == nil && len(osrmResp.Routes) > 0 {
					route := osrmResp.Routes[0]

					// Montar lista de waypoints para URL do Google Maps incluindo desvios (via:lat,lng)
					waypointsForURL := make([]string, 0, len(waypoints)+len(avoidanceWaypoints))
					if len(waypoints) > 0 {
						waypointsForURL = append(waypointsForURL, waypoints[0])
					}
					// Inserir os waypoints de desvio como "via:lat,lng"
					for _, wp := range avoidanceWaypoints {
						waypointsForURL = append(waypointsForURL, fmt.Sprintf("via:%f,%f", wp.Latitude, wp.Longitude))
					}
					// Inserir eventuais pontos intermediários originais (se houver)
					if len(waypoints) > 2 {
						waypointsForURL = append(waypointsForURL, waypoints[1:len(waypoints)-1]...)
					}
					if len(waypoints) > 1 {
						waypointsForURL = append(waypointsForURL, waypoints[len(waypoints)-1])
					}

					totalRoute = s.createTotalSummary(route, originLocation, destinationLocation, waypointsForURL, data)
				}
			}
		}
	}

	// Se não conseguiu rota com desvio, tentar obter a rota completa padrão (sem desvios) para obter polyline
	if totalRoute.TotalDistance.Value == 0 {
		// Tentar calcular a rota total via OSRM sem desvios
		baseCoords := strings.Join(allCoords, ";")
		urlTotal := fmt.Sprintf("http://34.207.174.233:5001/route/v1/driving/%s?alternatives=0&steps=true&overview=full&continue_straight=false", url.PathEscape(baseCoords))
		if resp, err := client.Get(urlTotal); err == nil {
			defer resp.Body.Close()
			var osrmResp OSRMResponse
			if err := json.NewDecoder(resp.Body).Decode(&osrmResp); err == nil && len(osrmResp.Routes) > 0 {
				route := osrmResp.Routes[0]
				totalRoute = s.createTotalSummary(route, originLocation, destinationLocation, waypoints, data)
			}
		}
	}

	// Se ainda não conseguiu rota com polyline, usar dados agregados (fallback)
	if totalRoute.TotalDistance.Value == 0 {
		distText, distVal := formatDistance(totalDistance)
		durText, durVal := formatDuration(totalDuration)

		avgConsumption := (data.ConsumptionCity + data.ConsumptionHwy) / 2
		totalKm := totalDistance / 1000
		totalFuelCost := math.Round((data.Price / avgConsumption) * totalKm)

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

		currentTimeMillis := (time.Now().UnixNano() + int64(totalDuration*float64(time.Second))) / int64(time.Millisecond)
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
			Tolls:         nil,
			TotalTolls:    0,
			Polyline:      "",
			TotalFuelCost: totalFuelCost,
		}
	}

	return totalRoute
}

// createTotalSummary cria um resumo total da rota
func (s *Service) createTotalSummary(route OSRMRoute, originLocation, destinationLocation Location, waypoints []string, data FrontInfoCEPRequest) TotalSummary {
	distText, distVal := formatDistance(route.Distance)
	durText, durVal := formatDuration(route.Duration)

	avgConsumption := (data.ConsumptionCity + data.ConsumptionHwy) / 2
	totalKm := route.Distance / 1000
	totalFuelCost := math.Round((data.Price / avgConsumption) * totalKm)

	tolls, _ := s.findTollsOnRoute(context.Background(), route.Geometry, data.Type, float64(data.Axles))
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

	return TotalSummary{
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
