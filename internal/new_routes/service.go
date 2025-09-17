package new_routes

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	db "geolocation/db/sqlc"
	"geolocation/internal/address"
	"geolocation/internal/get_token"
	"geolocation/internal/route_enterprise"
	"geolocation/internal/routes"
	"geolocation/internal/zonas_risco"
	cache "geolocation/pkg"
	"geolocation/validation"
	"log"
	"math"
	"net/http"
	neturl "net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"googlemaps.github.io/maps"
)

type InterfaceService interface {
	CalculateRoutes(ctx context.Context, frontInfo FrontInfo, idPublicToken int64, idSimp int64) (FinalOutput, error)
	CalculateRoutesWithCEP(ctx context.Context, frontInfo FrontInfoCEP, idPublicToken int64, idSimp int64, payloadSimp get_token.PayloadDTO) (FinalOutput, error)
	CalculateDistancesBetweenPoints(ctx context.Context, data FrontInfoCEPRequest) (Response, error)
	CalculateDistancesBetweenPointsWithRiskAvoidance(ctx context.Context, data FrontInfoCEPRequest) (Response, error)
	CalculateDistancesBetweenPointsWithRiskAvoidanceFromCoordinates(ctx context.Context, data FrontInfoCoordinatesRequest) (Response, error)
	CalculateDistancesFromOrigin(ctx context.Context, data FrontInfoCEPRequest) ([]DetailedRoute, error)
	CalculateRoutesWithCoordinate(ctx context.Context, frontInfo FrontInfoCoordinate, idPublicToken int64, idSimp int64) (FinalOutput, error)
	GetFavoriteRouteService(ctx context.Context, id int64) ([]FavoriteRouteResponse, error)
	RemoveFavoriteRouteService(ctx context.Context, id, idUser int64) error
	GetSimpleRoute(data SimpleRouteRequest) (SimpleRouteResponse, error)
	GetCoordinatesFromAddress(ctx context.Context, street, number, city, state, cep string) (AddressCoordinatesResponse, error)
}

type Service struct {
	InterfaceService         routes.InterfaceRepository
	InterfaceRouteEnterprise route_enterprise.InterfaceRepository
	GoogleMapsAPIKey         string
	RiskZonesRepository      zonas_risco.InterfaceService
	CEPRepository            address.InterfaceRepository
}

func NewRoutesNewService(interfaceService routes.InterfaceRepository, interfaceRouteEnterprise route_enterprise.InterfaceRepository, googleMapsAPIKey string, RiskZonesRepository zonas_risco.InterfaceService, CEPRepository address.InterfaceRepository) *Service {
	return &Service{
		InterfaceService:         interfaceService,
		InterfaceRouteEnterprise: interfaceRouteEnterprise,
		GoogleMapsAPIKey:         googleMapsAPIKey,
		RiskZonesRepository:      RiskZonesRepository,
		CEPRepository:            CEPRepository,
	}
}

// normalizeAddress normaliza endereços removendo caracteres especiais e corrigindo codificação
func normalizeAddress(address string) string {
	if address == "" {
		return address
	}

	// Remove caracteres de controle e normaliza espaços
	address = strings.TrimSpace(address)
	address = regexp.MustCompile(`\s+`).ReplaceAllString(address, " ")

	// Corrige caracteres especiais comuns
	replacements := map[string]string{
		"\\u00e3": "ã", // ã
		"\\u00e1": "á", // á
		"\\u00e9": "é", // é
		"\\u00ed": "í", // í
		"\\u00f3": "ó", // ó
		"\\u00fa": "ú", // ú
		"\\u00e0": "à", // à
		"\\u00e8": "è", // è
		"\\u00ec": "ì", // ì
		"\\u00f2": "ò", // ò
		"\\u00f9": "ù", // ù
		"\\u00e2": "â", // â
		"\\u00ea": "ê", // ê
		"\\u00ee": "î", // î
		"\\u00f4": "ô", // ô
		"\\u00fb": "û", // û
		"\\u00e7": "ç", // ç
		"\\u00c3": "Ã", // Ã
		"\\u00c1": "Á", // Á
		"\\u00c9": "É", // É
		"\\u00cd": "Í", // Í
		"\\u00d3": "Ó", // Ó
		"\\u00da": "Ú", // Ú
		"\\u00c0": "À", // À
		"\\u00c8": "È", // È
		"\\u00cc": "Ì", // Ì
		"\\u00d2": "Ò", // Ò
		"\\u00d9": "Ù", // Ù
		"\\u00c2": "Â", // Â
		"\\u00ca": "Ê", // Ê
		"\\u00ce": "Î", // Î
		"\\u00d4": "Ô", // Ô
		"\\u00db": "Û", // Û
		"\\u00c7": "Ç", // Ç
	}

	for old, new := range replacements {
		address = strings.ReplaceAll(address, old, new)
	}

	// Remove caracteres não-ASCII problemáticos e mantém apenas caracteres válidos
	var result strings.Builder
	for _, r := range address {
		if unicode.IsPrint(r) && !unicode.IsControl(r) {
			result.WriteRune(r)
		} else if r == '\n' || r == '\r' || r == '\t' {
			result.WriteRune(' ')
		}
	}

	return strings.TrimSpace(result.String())
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
			// Erro ao deserializar o cache
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
				// Erro ao salvar rota/favorita (cache)
			}

			cachedOutput.Summary.RouteHistID = routeHistID
			return cachedOutput, nil
		}
	} else if !errors.Is(err, redis.Nil) {
		// Erro ao recuperar cache do Redis
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
	baseOSRMURL := "http://34.207.174.233:5000/route/v1/driving/" + neturl.PathEscape(coords)
	client := http.Client{Timeout: 120 * time.Second}

	osrmURLFast := baseOSRMURL + "?" + neturl.Values{
		"alternatives":      {"3"},
		"steps":             {"true"},
		"overview":          {"full"},
		"continue_straight": {"false"},
	}.Encode()

	osrmURLNoTolls := baseOSRMURL + "?" + neturl.Values{
		"alternatives": {"3"},
		"steps":        {"true"},
		"overview":     {"full"},
		"exclude":      {"toll"},
	}.Encode()

	osrmURLEfficient := baseOSRMURL + "?" + neturl.Values{
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
		neturl.QueryEscape(normalizeAddress(origin.FormattedAddress)),
		neturl.QueryEscape(normalizeAddress(destination.FormattedAddress)))
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
					Address: normalizeAddress(origin.FormattedAddress),
				},
				LocationDestination: AddressInfo{
					Location: Location{
						Latitude:  destination.Location.Latitude,
						Longitude: destination.Location.Longitude,
					},
					Address: normalizeAddress(destination.FormattedAddress),
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
				Address: normalizeAddress(origin.FormattedAddress),
			},
			LocationDestination: AddressInfo{
				Location: Location{
					Latitude:  destination.Location.Latitude,
					Longitude: destination.Location.Longitude,
				},
				Address: normalizeAddress(destination.FormattedAddress),
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
			// Erro ao salvar cache do Redis
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
					// Erro ao salvar rota/favorita (cache)
				}

				cachedOutput.Summary.RouteHistID = routeHistID
				return cachedOutput, nil
			}
		}
	} else if !errors.Is(err, redis.Nil) {
		// Erro ao recuperar cache do Redis
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
	baseOSRMURL := "http://34.207.174.233:5000/route/v1/driving/" + neturl.PathEscape(coords)
	client := http.Client{Timeout: 120 * time.Second}

	osrmURLFast := baseOSRMURL + "?" + neturl.Values{
		"alternatives":      {"3"},
		"steps":             {"true"},
		"overview":          {"full"},
		"continue_straight": {"false"},
	}.Encode()

	osrmURLNoTolls := baseOSRMURL + "?" + neturl.Values{
		"alternatives": {"3"},
		"steps":        {"true"},
		"overview":     {"full"},
		"exclude":      {"toll"},
	}.Encode()

	osrmURLEfficient := baseOSRMURL + "?" + neturl.Values{
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
		neturl.QueryEscape(normalizeAddress(origin.FormattedAddress)),
		neturl.QueryEscape(normalizeAddress(destination.FormattedAddress)))

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
					Address: normalizeAddress(origin.FormattedAddress),
				},
				LocationDestination: AddressInfo{
					Location: Location{
						Latitude:  destination.Location.Latitude,
						Longitude: destination.Location.Longitude,
					},
					Address: normalizeAddress(destination.FormattedAddress),
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
				Address: normalizeAddress(origin.FormattedAddress),
			},
			LocationDestination: AddressInfo{
				Location: Location{
					Latitude:  destination.Location.Latitude,
					Longitude: destination.Location.Longitude,
				},
				Address: normalizeAddress(destination.FormattedAddress),
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
			// Erro ao salvar cache do Redis
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

	client := http.Client{Timeout: 30 * time.Second}
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
		baseURL := "http://34.207.174.233:5000/route/v1/driving/" + neturl.PathEscape(coords)

		type osrmResult struct {
			resp     OSRMResponse
			category string
			err      error
		}
		resultsCh := make(chan osrmResult, 3)

		makeRequest := func(params neturl.Values, category string) {
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
			go makeRequest(neturl.Values{
				"alternatives":      {"3"},
				"steps":             {"true"},
				"overview":          {"full"},
				"continue_straight": {"false"},
			}, "fastest")
			go makeRequest(neturl.Values{
				"alternatives": {"3"},
				"steps":        {"true"},
				"overview":     {"full"},
				"exclude":      {"toll"},
			}, "cheapest")
			go makeRequest(neturl.Values{
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
				makeRequest(neturl.Values{
					"alternatives":      {"3"},
					"steps":             {"true"},
					"overview":          {"full"},
					"continue_straight": {"false"},
				}, "fastest")
			case "barata", "cheapest":
				makeRequest(neturl.Values{
					"alternatives": {"3"},
					"steps":        {"true"},
					"overview":     {"full"},
					"exclude":      {"toll"},
				}, "cheapest")
			case "eficiente", "efficient":
				makeRequest(neturl.Values{
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
				neturl.QueryEscape(normalizeAddress(originGeocode.FormattedAddress)),
				neturl.QueryEscape(normalizeAddress(destGeocode.FormattedAddress)),
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
				Address:  normalizeAddress(originGeocode.FormattedAddress),
			},
			LocationDestination: AddressInfo{
				Location: Location{Latitude: destLat, Longitude: destLon},
				Address:  normalizeAddress(destGeocode.FormattedAddress),
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
	urlTotal := fmt.Sprintf("http://34.207.174.233:5000/route/v1/driving/%s?alternatives=0&steps=true&overview=full&continue_straight=false", neturl.PathEscape(coordsStr))

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
					Address:  normalizeAddress(destAddress),
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

	client := http.Client{Timeout: 30 * time.Second}
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
		baseURL := fmt.Sprintf("http://34.207.174.233:5000/route/v1/driving/%s?alternatives=0&steps=true&overview=full&continue_straight=false", neturl.PathEscape(coords))

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
			neturl.QueryEscape(normalizeAddress(originGeocode.FormattedAddress)),
			neturl.QueryEscape(normalizeAddress(destGeocode.FormattedAddress)),
		)

		currentTimeMillis := (time.Now().UnixNano() + int64(route.Duration*float64(time.Second))) / int64(time.Millisecond)
		wazeURL := fmt.Sprintf("https://www.waze.com/pt-BR/live-map/directions/br?to=%s&from=%s&time=%d&reverse=yes",
			neturl.QueryEscape(normalizeAddress(originGeocode.FormattedAddress)),
			neturl.QueryEscape(normalizeAddress(destGeocode.FormattedAddress)),
			currentTimeMillis,
		)

		results = append(results, DetailedRoute{
			LocationOrigin: AddressInfo{
				Location: Location{Latitude: originLat, Longitude: originLon},
				Address:  normalizeAddress(originGeocode.FormattedAddress),
			},
			LocationDestination: AddressInfo{
				Location: Location{Latitude: destLat, Longitude: destLon},
				Address:  normalizeAddress(destGeocode.FormattedAddress),
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
					// Erro ao salvar rota/favorita (cache)
				}

				cachedOutput.Summary.RouteHistID = routeHistID
				return cachedOutput, nil
			}
		}
	} else if !errors.Is(err, redis.Nil) {
		// Erro ao recuperar cache do Redis
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
	baseOSRMURL := "http://34.207.174.233:5000/route/v1/driving/" + neturl.PathEscape(coords)
	client := http.Client{Timeout: 120 * time.Second}

	osrmURLFast := baseOSRMURL + "?" + neturl.Values{
		"alternatives":      {"3"},
		"steps":             {"true"},
		"overview":          {"full"},
		"continue_straight": {"false"},
	}.Encode()

	osrmURLNoTolls := baseOSRMURL + "?" + neturl.Values{
		"alternatives": {"3"},
		"steps":        {"true"},
		"overview":     {"full"},
		"exclude":      {"toll"},
	}.Encode()

	osrmURLEfficient := baseOSRMURL + "?" + neturl.Values{
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
		neturl.QueryEscape(normalizeAddress(origin.FormattedAddress)),
		neturl.QueryEscape(normalizeAddress(destination.FormattedAddress)))

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

	routeType := "efficient"
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
				RouteType:     routeType,
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
					Address: normalizeAddress(origin.FormattedAddress),
				},
				LocationDestination: AddressInfo{
					Location: Location{
						Latitude:  destination.Location.Latitude,
						Longitude: destination.Location.Longitude,
					},
					Address: normalizeAddress(destination.FormattedAddress),
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
				Address: normalizeAddress(origin.FormattedAddress),
			},
			LocationDestination: AddressInfo{
				Location: Location{
					Latitude:  destination.Location.Latitude,
					Longitude: destination.Location.Longitude,
				},
				Address: normalizeAddress(destination.FormattedAddress),
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
			// Erro ao salvar cache do Redis
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
	baseOSRMURL := "http://34.207.174.233:5000/route/v1/driving/" + neturl.PathEscape(coords)
	client := http.Client{Timeout: 120 * time.Second}

	osrmURL := baseOSRMURL + "?" + neturl.Values{
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
				Address: normalizeAddress(originAddress),
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
	// Implementar cache para coordenadas
	cacheKey := fmt.Sprintf("reverse_geocode:%.6f,%.6f", lat, lng)
	cached, err := cache.Rdb.Get(cache.Ctx, cacheKey).Result()
	if err == nil {
		return cached, nil
	} else if !errors.Is(err, redis.Nil) {
		log.Printf("Erro ao recuperar cache do Redis (reverse_geocode): %v", err)
	}

	geocodeURL := fmt.Sprintf("https://nominatim.openstreetmap.org/reverse?format=json&lat=%f&lon=%f", lat, lng)
	client := http.Client{Timeout: 5 * time.Second}

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

	// Salvar no cache para futuras consultas
	if err := cache.Rdb.Set(cache.Ctx, cacheKey, result.DisplayName, 7*24*time.Hour).Err(); err != nil {
		log.Printf("Erro ao salvar cache do Redis (reverse_geocode): %v", err)
	}

	return normalizeAddress(result.DisplayName), nil
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
	// Implementar cache para evitar chamadas repetidas
	cacheKey := fmt.Sprintf("geocode:%s", address)
	cached, err := cache.Rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		var result GeocodeResult
		if json.Unmarshal([]byte(cached), &result) == nil {
			return result, nil
		}
	} else if !errors.Is(err, redis.Nil) {
		log.Printf("Erro ao recuperar cache do Redis (geocode): %v", err)
	}

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
		FormattedAddress: normalizeAddress(results[0].FormattedAddress),
		PlaceID:          results[0].PlaceID,
		Location: Location{
			Latitude:  results[0].Geometry.Location.Lat,
			Longitude: results[0].Geometry.Location.Lng,
		},
	}

	// Salvar no cache para futuras consultas
	data, err := json.Marshal(result)
	if err == nil {
		if err := cache.Rdb.Set(ctx, cacheKey, data, 30*24*time.Hour).Err(); err != nil {
			log.Printf("Erro ao salvar cache do Redis (geocode): %v", err)
		}
	}
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

func (s *Service) getCoordByCEP(ctx context.Context, cep string) (lat float64, lon float64, error error) {
	cepRegex := regexp.MustCompile(`^\d{8}$`)
	normalizedQuery := strings.ReplaceAll(strings.ReplaceAll(cep, "-", ""), " ", "")
	isCEP := cepRegex.MatchString(normalizedQuery)
	if !isCEP {
		return 0, 0, errors.New("CEP inválido")
	}

	// Implementar cache para CEPs
	cacheKey := fmt.Sprintf("cep_coords:%s", normalizedQuery)
	cached, err := cache.Rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		var coords struct {
			Lat float64 `json:"lat"`
			Lon float64 `json:"lon"`
		}
		if json.Unmarshal([]byte(cached), &coords) == nil {
			return coords.Lat, coords.Lon, nil
		}
	} else if !errors.Is(err, redis.Nil) {
		log.Printf("Erro ao recuperar cache do Redis (cep_coords): %v", err)
	}

	infoCep, err := s.CEPRepository.FindAddressGroupedByCEPRepository(ctx, normalizedQuery)
	if err != nil {
		address, apiErr := address.FindCEPByAPIBrasil(ctx, normalizedQuery)
		if apiErr != nil {
			log.Printf("erro ao buscar CEP em ambas base de dados: %v", apiErr)
			return 0, 0, err
		}

		// Salvar no cache
		coords := struct {
			Lat float64 `json:"lat"`
			Lon float64 `json:"lon"`
		}{address.Latitude, address.Longitude}

		if data, err := json.Marshal(coords); err == nil {
			cache.Rdb.Set(ctx, cacheKey, data, 30*24*time.Hour)
		}

		return address.Latitude, address.Longitude, nil
	}

	// Salvar no cache
	coords := struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	}{infoCep.Latitude.Float64, infoCep.Longitude.Float64}

	if data, err := json.Marshal(coords); err == nil {
		cache.Rdb.Set(ctx, cacheKey, data, 30*24*time.Hour)
	}

	return infoCep.Latitude.Float64, infoCep.Longitude.Float64, nil
}

// -- new function
func (s *Service) CalculateDistancesBetweenPointsWithRiskAvoidance(ctx context.Context, data FrontInfoCEPRequest) (Response, error) {

	if len(data.CEPs) < 2 {
		return Response{}, fmt.Errorf("é necessário pelo menos dois pontos para calcular distâncias")
	}

	// Buscar zonas de risco uma única vez
	riskZones, err := s.getActiveRiskZones(ctx, data.OrganizationID)
	if err != nil {
		log.Printf("Erro ao buscar zonas de risco: %v", err)
		return s.CalculateDistancesBetweenPoints(ctx, data)
	}

	// Pool de clientes HTTP reutilizáveis para melhor performance
	clientPool := &http.Client{
		Timeout: 10 * time.Second, // Timeout ainda mais agressivo
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 20, // Mais conexões por host
			IdleConnTimeout:     30 * time.Second,
			DisableCompression:  true,  // OSRM não usa compressão
			DisableKeepAlives:   false, // Manter conexões ativas
			MaxConnsPerHost:     50,    // Mais conexões simultâneas
		},
	}
	var resultRoutes []DetailedRoute
	var totalDistance float64
	var totalDuration float64

	// Pré-processar todos os CEPs para evitar chamadas repetidas
	cepCoordinates := make(map[string]struct {
		lat, lon float64
		address  string
		geocode  GeocodeResult
	})

	// Processar CEPs em batch para melhor performance
	var wgCEP sync.WaitGroup
	cepChan := make(chan struct {
		cep     string
		lat     float64
		lon     float64
		address string
		geocode GeocodeResult
		err     error
	}, len(data.CEPs))

	for _, cep := range data.CEPs {
		if _, exists := cepCoordinates[cep]; exists {
			continue
		}

		wgCEP.Add(1)
		go func(cep string) {
			defer wgCEP.Done()

			lat, lon, err := s.getCoordByCEP(ctx, cep)
			if err != nil {
				cepChan <- struct {
					cep     string
					lat     float64
					lon     float64
					address string
					geocode GeocodeResult
					err     error
				}{cep: cep, err: err}
				return
			}

			// Geocodificação reversa e geocodificação em paralelo
			addressChan := make(chan string, 1)
			geocodeChan := make(chan GeocodeResult, 1)
			errorChan := make(chan error, 2)

			go func(lat, lon float64) {
				address, err := s.reverseGeocode(lat, lon)
				if err != nil {
					errorChan <- err
					return
				}
				addressChan <- address
			}(lat, lon)

			go func() {
				geocode, err := s.getGeocodeAddress(ctx, fmt.Sprintf("%f,%f", lat, lon))
				if err != nil {
					errorChan <- err
					return
				}
				geocodeChan <- geocode
			}()

			// Aguardar resultados com timeout
			select {
			case address := <-addressChan:
				select {
				case geocode := <-geocodeChan:
					cepChan <- struct {
						cep     string
						lat     float64
						lon     float64
						address string
						geocode GeocodeResult
						err     error
					}{cep, lat, lon, address, geocode, nil}
				case err := <-errorChan:
					cepChan <- struct {
						cep     string
						lat     float64
						lon     float64
						address string
						geocode GeocodeResult
						err     error
					}{cep, lat, lon, "", GeocodeResult{}, err}
				case <-time.After(5 * time.Second):
					cepChan <- struct {
						cep     string
						lat     float64
						lon     float64
						address string
						geocode GeocodeResult
						err     error
					}{cep, lat, lon, "", GeocodeResult{}, fmt.Errorf("timeout geocodificação")}
				}
			case err := <-errorChan:
				cepChan <- struct {
					cep     string
					lat     float64
					lon     float64
					address string
					geocode GeocodeResult
					err     error
				}{cep, lat, lon, "", GeocodeResult{}, err}
			case <-time.After(5 * time.Second):
				cepChan <- struct {
					cep     string
					lat     float64
					lon     float64
					address string
					geocode GeocodeResult
					err     error
				}{cep, lat, lon, "", GeocodeResult{}, fmt.Errorf("timeout geocodificação")}
			}
		}(cep)
	}

	// Aguardar todos os CEPs serem processados
	go func() {
		wgCEP.Wait()
		close(cepChan)
	}()

	// Coletar resultados dos CEPs
	for result := range cepChan {
		if result.err != nil {
			log.Printf("Erro na geocodificação do CEP %s: %v", result.cep, result.err)
			continue
		}

		cepCoordinates[result.cep] = struct {
			lat, lon float64
			address  string
			geocode  GeocodeResult
		}{result.lat, result.lon, result.address, result.geocode}
	}

	// Processar segmentos em paralelo quando possível
	type segmentResult struct {
		index   int
		route   DetailedRoute
		hasRisk bool
	}

	segmentChan := make(chan segmentResult, len(data.CEPs)-1)
	var wg sync.WaitGroup

	// Early exit se não há zonas de risco
	if len(riskZones) == 0 {
		// Processar todos os segmentos sem verificação de risco
		for i := 0; i < len(data.CEPs)-1; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()

				originCEP := data.CEPs[i]
				destCEP := data.CEPs[i+1]

				originData := cepCoordinates[originCEP]
				destData := cepCoordinates[destCEP]

				// Sem zonas de risco, usar rota direta
				summaries := s.calculateDirectRoute(ctx, *clientPool, originData.lat, originData.lon, destData.lat, destData.lon, originData.geocode, destData.geocode, data)

				if len(summaries) == 0 {
					fb := s.createDirectEstimateSummary(originData.lat, originData.lon, destData.lat, destData.lon, originData.geocode, destData.geocode, data)
					summaries = []RouteSummary{fb}
				}

				route := summaries[0]
				totalDistance += route.Distance.Value
				totalDuration += route.Duration.Value

				segmentChan <- segmentResult{
					index: i,
					route: DetailedRoute{
						LocationOrigin:      AddressInfo{Location: Location{Latitude: originData.lat, Longitude: originData.lon}, Address: originData.geocode.FormattedAddress},
						LocationDestination: AddressInfo{Location: Location{Latitude: destData.lat, Longitude: destData.lon}, Address: destData.geocode.FormattedAddress},
						HasRisk:             false,
						LocationHisk:        []LocationHisk{},
						Summaries:           summaries,
					},
					hasRisk: false,
				}
			}(i)
		}
	} else {
		// Com zonas de risco, verificar cada segmento
		for i := 0; i < len(data.CEPs)-1; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()

				originCEP := data.CEPs[i]
				destCEP := data.CEPs[i+1]

				originData := cepCoordinates[originCEP]
				destData := cepCoordinates[destCEP]

				// Verificar se há zonas de risco no caminho direto
				offs, hasAny := s.CheckRouteForAllRiskZones(riskZones, originData.lat, originData.lon, destData.lat, destData.lon)

				locations := make([]LocationHisk, 0, len(offs))
				for _, o := range offs {
					locations = append(locations, LocationHisk{
						CEP:       o.Zone.Cep,
						Latitude:  o.Zone.Lat,
						Longitude: o.Zone.Lng,
					})
				}

				var summaries []RouteSummary
				if hasAny {
					summaries = s.calculateAlternativeRouteWithAvoidance(ctx, *clientPool, riskZones, originData.lat, originData.lon, destData.lat, destData.lon, originData.geocode, destData.geocode, data)
				} else {
					summaries = s.calculateDirectRoute(ctx, *clientPool, originData.lat, originData.lon, destData.lat, destData.lon, originData.geocode, destData.geocode, data)
				}

				if len(summaries) == 0 {
					fb := s.createDirectEstimateSummary(originData.lat, originData.lon, destData.lat, destData.lon, originData.geocode, destData.geocode, data)
					summaries = []RouteSummary{fb}
				}

				// Usar a primeira rota para cálculos totais
				route := summaries[0]
				totalDistance += route.Distance.Value
				totalDuration += route.Duration.Value

				segmentChan <- segmentResult{
					index: i,
					route: DetailedRoute{
						LocationOrigin:      AddressInfo{Location: Location{Latitude: originData.lat, Longitude: originData.lon}, Address: originData.geocode.FormattedAddress},
						LocationDestination: AddressInfo{Location: Location{Latitude: destData.lat, Longitude: destData.lon}, Address: destData.geocode.FormattedAddress},
						HasRisk:             hasAny,
						LocationHisk:        locations,
						Summaries:           summaries,
					},
					hasRisk: hasAny,
				}
			}(i)
		}
	}

	// Aguardar todos os segmentos
	go func() {
		wg.Wait()
		close(segmentChan)
	}()

	// Coletar resultados na ordem correta
	segmentResults := make([]segmentResult, len(data.CEPs)-1)
	for result := range segmentChan {
		segmentResults[result.index] = result
	}

	// Montar resultado final na ordem correta
	for _, result := range segmentResults {
		resultRoutes = append(resultRoutes, result.route)
	}

	// Calcular rota total com desvios
	totalRoute := s.calculateTotalRouteWithAvoidance(ctx, *clientPool, riskZones, data.CEPs, totalDistance, totalDuration, data)

	return Response{
		Routes:     resultRoutes,
		TotalRoute: totalRoute,
	}, nil
}

// ---- new function for coordinates
func (s *Service) CalculateDistancesBetweenPointsWithRiskAvoidanceFromCoordinates(ctx context.Context, data FrontInfoCoordinatesRequest) (Response, error) {

	if len(data.Coordinates) < 2 {
		return Response{}, fmt.Errorf("é necessário pelo menos dois pontos para calcular distâncias")
	}

	// Buscar zonas de risco uma única vez
	riskZones, err := s.getActiveRiskZones(ctx, data.OrganizationID)
	if err != nil {
		log.Printf("Erro ao buscar zonas de risco: %v", err)
		return s.CalculateDistancesBetweenPointsFromCoordinates(ctx, data)
	}

	// Buscar balanças disponíveis
	balancas, err := s.InterfaceService.GetBalanca(ctx)
	if err != nil {
		log.Printf("Erro ao obter balanças: %v", err)
		balancas = nil
	}

	// Pool de clientes HTTP reutilizáveis para melhor performance
	clientPool := &http.Client{
		Timeout: 10 * time.Second, // Timeout ainda mais agressivo
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 20, // Mais conexões por host
			IdleConnTimeout:     30 * time.Second,
			DisableCompression:  true,  // OSRM não usa compressão
			DisableKeepAlives:   false, // Manter conexões ativas
			MaxConnsPerHost:     50,    // Mais conexões simultâneas
		},
	}
	var resultRoutes []DetailedRoute
	var totalDistance float64
	var totalDuration float64

	// Pré-processar todas as coordenadas para evitar chamadas repetidas
	coordinateData := make(map[string]struct {
		lat, lon float64
		address  string
		geocode  GeocodeResult
	})

	// Processar coordenadas em batch para melhor performance
	var wgCoord sync.WaitGroup
	coordChan := make(chan struct {
		coord   string
		lat     float64
		lon     float64
		address string
		geocode GeocodeResult
		err     error
	}, len(data.Coordinates))

	for _, coord := range data.Coordinates {
		coordKey := fmt.Sprintf("%s,%s", coord.Lat, coord.Lng)
		if _, exists := coordinateData[coordKey]; exists {
			continue
		}

		wgCoord.Add(1)
		go func(coord Coordinate) {
			defer wgCoord.Done()

			// Converter strings para float64
			lat, err1 := strconv.ParseFloat(strings.ReplaceAll(coord.Lat, ",", "."), 64)
			lon, err2 := strconv.ParseFloat(strings.ReplaceAll(coord.Lng, ",", "."), 64)

			if err1 != nil || err2 != nil {
				coordChan <- struct {
					coord   string
					lat     float64
					lon     float64
					address string
					geocode GeocodeResult
					err     error
				}{coord: coordKey, err: fmt.Errorf("erro ao converter coordenadas: %v, %v", err1, err2)}
				return
			}

			// Geocodificação reversa e geocodificação em paralelo
			addressChan := make(chan string, 1)
			geocodeChan := make(chan GeocodeResult, 1)
			errorChan := make(chan error, 2)

			go func(lat, lon float64) {
				address, err := s.reverseGeocode(lat, lon)
				if err != nil {
					errorChan <- err
					return
				}
				addressChan <- address
			}(lat, lon)

			go func() {
				geocode, err := s.getGeocodeAddress(ctx, fmt.Sprintf("%f,%f", lat, lon))
				if err != nil {
					errorChan <- err
					return
				}
				geocodeChan <- geocode
			}()

			// Aguardar resultados com timeout
			select {
			case address := <-addressChan:
				select {
				case geocode := <-geocodeChan:
					coordChan <- struct {
						coord   string
						lat     float64
						lon     float64
						address string
						geocode GeocodeResult
						err     error
					}{coordKey, lat, lon, address, geocode, nil}
				case err := <-errorChan:
					coordChan <- struct {
						coord   string
						lat     float64
						lon     float64
						address string
						geocode GeocodeResult
						err     error
					}{coordKey, lat, lon, "", GeocodeResult{}, err}
				case <-time.After(5 * time.Second):
					coordChan <- struct {
						coord   string
						lat     float64
						lon     float64
						address string
						geocode GeocodeResult
						err     error
					}{coordKey, lat, lon, "", GeocodeResult{}, fmt.Errorf("timeout geocodificação")}
				}
			case err := <-errorChan:
				coordChan <- struct {
					coord   string
					lat     float64
					lon     float64
					address string
					geocode GeocodeResult
					err     error
				}{coordKey, lat, lon, "", GeocodeResult{}, err}
			case <-time.After(5 * time.Second):
				coordChan <- struct {
					coord   string
					lat     float64
					lon     float64
					address string
					geocode GeocodeResult
					err     error
				}{coordKey, lat, lon, "", GeocodeResult{}, fmt.Errorf("timeout geocodificação")}
			}
		}(coord)
	}

	// Aguardar todas as coordenadas serem processadas
	go func() {
		wgCoord.Wait()
		close(coordChan)
	}()

	// Coletar resultados das coordenadas
	for result := range coordChan {
		if result.err != nil {
			log.Printf("Erro na geocodificação da coordenada %s: %v", result.coord, result.err)
			continue
		}

		coordinateData[result.coord] = struct {
			lat, lon float64
			address  string
			geocode  GeocodeResult
		}{result.lat, result.lon, result.address, result.geocode}
	}

	// Processar segmentos em paralelo quando possível
	type segmentResult struct {
		index   int
		route   DetailedRoute
		hasRisk bool
	}

	segmentChan := make(chan segmentResult, len(data.Coordinates)-1)
	var wg sync.WaitGroup

	// Early exit se não há zonas de risco
	if len(riskZones) == 0 {
		// Processar todos os segmentos sem verificação de risco
		for i := 0; i < len(data.Coordinates)-1; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()

				originCoord := data.Coordinates[i]
				destCoord := data.Coordinates[i+1]
				originKey := fmt.Sprintf("%s,%s", originCoord.Lat, originCoord.Lng)
				destKey := fmt.Sprintf("%s,%s", destCoord.Lat, destCoord.Lng)

				originData := coordinateData[originKey]
				destData := coordinateData[destKey]

				// Sem zonas de risco, usar rota direta
				summaries := s.calculateDirectRoute(ctx, *clientPool, originData.lat, originData.lon, destData.lat, destData.lon, originData.geocode, destData.geocode, s.convertCoordinatesToCEPRequest(data))

				if len(summaries) == 0 {
					fb := s.createDirectEstimateSummary(originData.lat, originData.lon, destData.lat, destData.lon, originData.geocode, destData.geocode, s.convertCoordinatesToCEPRequest(data))
					summaries = []RouteSummary{fb}
				}

				route := summaries[0]
				totalDistance += route.Distance.Value
				totalDuration += route.Duration.Value

				segmentChan <- segmentResult{
					index: i,
					route: DetailedRoute{
						LocationOrigin:      AddressInfo{Location: Location{Latitude: originData.lat, Longitude: originData.lon}, Address: originData.geocode.FormattedAddress},
						LocationDestination: AddressInfo{Location: Location{Latitude: destData.lat, Longitude: destData.lon}, Address: destData.geocode.FormattedAddress},
						HasRisk:             false,
						LocationHisk:        []LocationHisk{},
						Summaries:           summaries,
					},
					hasRisk: false,
				}
			}(i)
		}
	} else {
		// Com zonas de risco, verificar cada segmento
		for i := 0; i < len(data.Coordinates)-1; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()

				originCoord := data.Coordinates[i]
				destCoord := data.Coordinates[i+1]
				originKey := fmt.Sprintf("%s,%s", originCoord.Lat, originCoord.Lng)
				destKey := fmt.Sprintf("%s,%s", destCoord.Lat, destCoord.Lng)

				originData := coordinateData[originKey]
				destData := coordinateData[destKey]

				// Verificar se há zonas de risco no caminho direto
				offs, hasAny := s.CheckRouteForAllRiskZones(riskZones, originData.lat, originData.lon, destData.lat, destData.lon)

				locations := make([]LocationHisk, 0, len(offs))
				for _, o := range offs {
					locations = append(locations, LocationHisk{
						CEP:       o.Zone.Cep,
						Latitude:  o.Zone.Lat,
						Longitude: o.Zone.Lng,
					})
				}

				var summaries []RouteSummary
				if hasAny {
					summaries = s.calculateAlternativeRouteWithAvoidance(ctx, *clientPool, riskZones, originData.lat, originData.lon, destData.lat, destData.lon, originData.geocode, destData.geocode, s.convertCoordinatesToCEPRequest(data))
				} else {
					summaries = s.calculateDirectRoute(ctx, *clientPool, originData.lat, originData.lon, destData.lat, destData.lon, originData.geocode, destData.geocode, s.convertCoordinatesToCEPRequest(data))
				}

				if len(summaries) == 0 {
					fb := s.createDirectEstimateSummary(originData.lat, originData.lon, destData.lat, destData.lon, originData.geocode, destData.geocode, s.convertCoordinatesToCEPRequest(data))
					summaries = []RouteSummary{fb}
				}

				// Usar a primeira rota para cálculos totais
				route := summaries[0]
				totalDistance += route.Distance.Value
				totalDuration += route.Duration.Value

				segmentChan <- segmentResult{
					index: i,
					route: DetailedRoute{
						LocationOrigin:      AddressInfo{Location: Location{Latitude: originData.lat, Longitude: originData.lon}, Address: originData.geocode.FormattedAddress},
						LocationDestination: AddressInfo{Location: Location{Latitude: destData.lat, Longitude: destData.lon}, Address: destData.geocode.FormattedAddress},
						HasRisk:             hasAny,
						LocationHisk:        locations,
						Summaries:           summaries,
					},
					hasRisk: hasAny,
				}
			}(i)
		}
	}

	// Aguardar todos os segmentos
	go func() {
		wg.Wait()
		close(segmentChan)
	}()

	// Coletar resultados na ordem correta
	segmentResults := make([]segmentResult, len(data.Coordinates)-1)
	for result := range segmentChan {
		segmentResults[result.index] = result
	}

	// Montar resultado final na ordem correta
	for _, result := range segmentResults {
		resultRoutes = append(resultRoutes, result.route)
	}

	// Calcular rota total com desvios - agora retorna múltiplas opções
	totalRoute, allTotalRoutes := s.calculateTotalRouteWithAvoidanceFromCoordinates(ctx, *clientPool, riskZones, data.Coordinates, totalDistance, totalDuration, data)

	// Filtrar balanças para a rota total se disponível
	var routeBalancas interface{}
	if balancas != nil && totalRoute.Polyline != "" {
		filteredBalancas, err := s.findBalancaOnRoute(totalRoute.Polyline, balancas)
		if err != nil {
			log.Printf("Erro ao filtrar balanças: %v", err)
			routeBalancas = nil
		} else {
			routeBalancas = filteredBalancas
		}
	}

	// Retorna tanto a melhor rota quanto todas as opções
	response := Response{
		Routes:      resultRoutes,
		TotalRoute:  totalRoute,
		TotalRoutes: allTotalRoutes, // Sempre inclui todas as rotas disponíveis
		Balances:    routeBalancas,
	}

	return response, nil
}

// convertCoordinatesToCEPRequest converte FrontInfoCoordinatesRequest para FrontInfoCEPRequest
func (s *Service) convertCoordinatesToCEPRequest(data FrontInfoCoordinatesRequest) FrontInfoCEPRequest {
	// Converter coordenadas para CEPs fictícios para manter compatibilidade
	ceps := make([]string, len(data.Coordinates))
	for i, coord := range data.Coordinates {
		ceps[i] = fmt.Sprintf("%s,%s", coord.Lat, coord.Lng)
	}

	return FrontInfoCEPRequest{
		CEPs:            ceps,
		ConsumptionCity: data.ConsumptionCity,
		ConsumptionHwy:  data.ConsumptionHwy,
		Price:           data.Price,
		Axles:           data.Axles,
		Type:            data.Type,
		TypeRoute:       data.TypeRoute,
		RouteOptions:    data.RouteOptions,
		Waypoints:       data.Waypoints,
		OrganizationID:  data.OrganizationID,
	}
}

// CalculateDistancesBetweenPointsFromCoordinates função auxiliar para fallback
func (s *Service) CalculateDistancesBetweenPointsFromCoordinates(ctx context.Context, data FrontInfoCoordinatesRequest) (Response, error) {
	// Converter para CEPs fictícios e usar a função existente
	cepData := s.convertCoordinatesToCEPRequest(data)
	return s.CalculateDistancesBetweenPoints(ctx, cepData)
}

// calculateTotalRouteWithAvoidanceFromCoordinates calcula rota total com desvios para coordenadas
func (s *Service) calculateTotalRouteWithAvoidanceFromCoordinates(ctx context.Context, client http.Client, riskZones []RiskZone, coordinates []Coordinate, totalDistance, totalDuration float64, data FrontInfoCoordinatesRequest) (TotalSummary, []TotalSummary) {

	// ------------------------------
	// 1) Monta lista base de coords e endereços
	// ------------------------------
	var allCoords []string
	var waypoints []string
	var originLocation, destinationLocation Location

	for idx, coord := range coordinates {
		lat, err1 := strconv.ParseFloat(strings.ReplaceAll(coord.Lat, ",", "."), 64)
		lon, err2 := strconv.ParseFloat(strings.ReplaceAll(coord.Lng, ",", "."), 64)

		if err1 != nil || err2 != nil {
			continue
		}

		allCoords = append(allCoords, fmt.Sprintf("%f,%f", lon, lat))

		reverse, _ := s.reverseGeocode(lat, lon)
		geocode, _ := s.getGeocodeAddress(ctx, reverse)
		waypoints = append(waypoints, geocode.FormattedAddress)

		if idx == 0 {
			originLocation = Location{Latitude: lat, Longitude: lon}
		}
		if idx == len(coordinates)-1 {
			destinationLocation = Location{Latitude: lat, Longitude: lon}
		}
	}

	if len(allCoords) < 2 {
		return TotalSummary{}, []TotalSummary{}
	}

	// ------------------------------
	// 2) Para cada segmento, injeta via-points evitando TODAS as zonas cruzadas
	// ------------------------------
	newCoords := make([]string, 0, len(allCoords)+12) // espaço extra para vias
	newCoords = append(newCoords, allCoords[0])

	var extraWaypointsForURL []string
	var detourPtsTotal []Location // opcional para expor no TotalSummary

	for i := 0; i < len(allCoords)-1; i++ {
		var lon1, lat1, lon2, lat2 float64
		fmt.Sscanf(allCoords[i], "%f,%f", &lon1, &lat1)
		fmt.Sscanf(allCoords[i+1], "%f,%f", &lon2, &lat2)

		// Sempre vamos trabalhar numa lista de via-points ACUMULADOS por segmento
		segWps := []Location{} // via-points inseridos nesse segmento

		// helper para calcular a polyline do segmento COM os via-points atuais
		routeForSegment := func() (OSRMRoute, bool) {
			coords := fmt.Sprintf("%f,%f", lon1, lat1)
			for _, wp := range segWps {
				coords += fmt.Sprintf(";%f,%f", wp.Longitude, wp.Latitude)
			}
			coords += fmt.Sprintf(";%f,%f", lon2, lat2)

			profile := "driving"
			if data.Type != "" {
				profile = data.Type // ex: "car", "truck"
			}

			u := fmt.Sprintf("http://34.207.174.233:5000/route/v1/%s/%s?alternatives=0&steps=true&overview=full&continue_straight=false",
				profile, neturl.PathEscape(coords))

			// Cliente com timeout reduzido para melhor performance
			fastClient := http.Client{Timeout: 15 * time.Second}
			resp, err := fastClient.Get(u)
			if err != nil || resp.StatusCode != 200 {
				u = fmt.Sprintf("http://34.207.174.233:5000/route/v1/driving/%s?alternatives=0&steps=true&overview=full&continue_straight=false",
					neturl.PathEscape(coords))
				resp, err = fastClient.Get(u)
			}

			defer resp.Body.Close()
			var r OSRMResponse
			if json.NewDecoder(resp.Body).Decode(&r) != nil || len(r.Routes) == 0 {
				return OSRMRoute{}, false
			}
			return r.Routes[0], true
		}

		// 2.A) Calcula rota do segmento (sem vias) e coleta TODAS as zonas cruzadas
		if r0, ok := routeForSegment(); ok {
			crossings := s.detectAllCrossingsFromGeometry(r0.Geometry, riskZones, 1000)

			// 2.B) Itera enquanto ainda cruzar alguma zona
			iter := 0
			for len(crossings) > 0 && iter < 15 {
				iter++
				off := crossings[0] // trata a próxima na ordem do percurso

				// ====== Estratégia preferencial: 3 pontos no MESMO lado ======
				latRef, nx, ny := s.awayNormalForSegment(off.Before5km, off.After5km, off.Zone)
				injected := false

				if nx != 0 || ny != 0 {
					baseDist := float64(off.Zone.Radius) + 200.0 // raio + buffer
					mid := Location{
						Latitude:  (off.Entry.Latitude + off.Exit.Latitude) / 2,
						Longitude: (off.Entry.Longitude + off.Exit.Longitude) / 2,
					}
					buildSeq := func(scale float64) []Location {
						d := baseDist * scale
						p1 := s.offsetByNormal(off.Before5km, latRef, nx, ny, d)
						p2 := s.offsetByNormal(mid, latRef, nx, ny, d)
						p3 := s.offsetByNormal(off.After5km, latRef, nx, ny, d)
						return []Location{p1, p2, p3}
					}

					for _, sc := range []float64{1.0, 1.3, 1.6, 2.0} {
						cand := buildSeq(sc)
						// Snap GARANTINDO ficar fora da zona
						cand = s.snapOutsideMany(client, cand, off.Zone)

						// Testa rota do segmento com segWps + cand
						segTest := append(append([]Location{}, segWps...), cand...)
						if r2, ok2 := func() (OSRMRoute, bool) {
							coords := fmt.Sprintf("%f,%f", lon1, lat1)
							for _, wp := range segTest {
								coords += fmt.Sprintf(";%f,%f", wp.Longitude, wp.Latitude)
							}
							coords += fmt.Sprintf(";%f,%f", lon2, lat2)

							profile := "driving"
							if data.Type != "" {
								profile = data.Type // ex: "car", "truck"
							}

							u := fmt.Sprintf("http://34.207.174.233:5000/route/v1/%s/%s?alternatives=0&steps=true&overview=full&continue_straight=false",
								profile, neturl.PathEscape(coords))

							resp, err := client.Get(u)
							if err != nil || resp.StatusCode != 200 {
								u = fmt.Sprintf("http://34.207.174.233:5000/route/v1/driving/%s?alternatives=0&steps=true&overview=full&continue_straight=false",
									neturl.PathEscape(coords))
								resp, err = client.Get(u)
							}

							defer resp.Body.Close()
							var r OSRMResponse
							if json.NewDecoder(resp.Body).Decode(&r) != nil || len(r.Routes) == 0 {
								return OSRMRoute{}, false
							}
							return r.Routes[0], true
						}(); ok2 {
							// rota de teste não pode cruzar ESTA zona
							if hasThis, _ := s.checkRouteGeometryForRiskZones([]RiskZone{off.Zone}, r2.Geometry, lat1, lon1, lat2, lon2); !hasThis {
								segWps = append(segWps, cand...)
								for _, p := range cand {
									extraWaypointsForURL = append(extraWaypointsForURL, fmt.Sprintf("via:%f,%f", p.Latitude, p.Longitude))
									detourPtsTotal = append(detourPtsTotal, p)
								}
								injected = true
								break
							}
						}
					}
				}

				// ====== Fallback: lateral e A/B
				if !injected {
					// menor arco
					seq := s.assembleLateralDetour(off.Entry, off.Exit, off.Zone, 2, 200, 80, false)
					seq = s.snapOutsideMany(client, seq, off.Zone)
					tryInject := func(cand []Location) bool {
						test := append(append([]Location{}, segWps...), cand...)
						if r2, ok2 := func() (OSRMRoute, bool) {
							coords := fmt.Sprintf("%f,%f", lon1, lat1)
							for _, wp := range test {
								coords += fmt.Sprintf(";%f,%f", wp.Longitude, wp.Latitude)
							}
							coords += fmt.Sprintf(";%f,%f", lon2, lat2)

							profile := "driving"
							if data.Type != "" {
								profile = data.Type // ex: "car", "truck"
							}

							u := fmt.Sprintf("http://34.207.174.233:5000/route/v1/%s/%s?alternatives=0&steps=true&overview=full&continue_straight=false",
								profile, neturl.PathEscape(coords))

							resp, err := client.Get(u)
							if err != nil || resp.StatusCode != 200 {
								u = fmt.Sprintf("http://34.207.174.233:5000/route/v1/driving/%s?alternatives=0&steps=true&overview=full&continue_straight=false",
									neturl.PathEscape(coords))
								resp, err = client.Get(u)
							}

							defer resp.Body.Close()
							var r OSRMResponse
							if json.NewDecoder(resp.Body).Decode(&r) != nil || len(r.Routes) == 0 {
								return OSRMRoute{}, false
							}
							return r.Routes[0], true
						}(); ok2 {
							if hasThis, _ := s.checkRouteGeometryForRiskZones([]RiskZone{off.Zone}, r2.Geometry, lat1, lon1, lat2, lon2); !hasThis {
								segWps = append(segWps, cand...)
								for _, p := range cand {
									extraWaypointsForURL = append(extraWaypointsForURL, fmt.Sprintf("via:%f,%f", p.Latitude, p.Longitude))
									detourPtsTotal = append(detourPtsTotal, p)
								}
								return true
							}
						}
						return false
					}

					if !tryInject(seq) {
						// arco oposto
						seq2 := s.assembleLateralDetour(off.Entry, off.Exit, off.Zone, 2, 200, 80, true)
						seq2 = s.snapOutsideMany(client, seq2, off.Zone)
						if !tryInject(seq2) {
							// A/B padrão
							wpA, wpB := s.computeBypassWaypoints(lat1, lon1, lat2, lon2, off.Zone)
							ab := s.snapOutsideMany(client, []Location{wpA, wpB}, off.Zone)
							if !tryInject(ab) {
								// A/B escalado
								for _, sc := range []float64{1.5, 2.0, 3.0} {
									wpA2 := s.pushAwayFromCenter(wpA, off.Zone, float64(off.Zone.Radius)*(sc-1)+500)
									wpB2 := s.pushAwayFromCenter(wpB, off.Zone, float64(off.Zone.Radius)*(sc-1)+500)
									ab2 := s.snapOutsideMany(client, []Location{wpA2, wpB2}, off.Zone)
									if tryInject(ab2) {
										break
									}
								}
							}
						}
					}
				}

				// Recalcula rota do segmento com os via-points acumulados e atualiza a lista de crossings
				if rNow, okNow := routeForSegment(); okNow {
					crossings = s.detectAllCrossingsFromGeometry(rNow.Geometry, riskZones, 1000)
				} else {
					// sem rota após tentativa — interrompe
					break
				}
			} // while cruzar zona
		}

		// Append final do destino do segmento e materializa os via-points em newCoords
		for _, wp := range segWps {
			newCoords = append(newCoords, fmt.Sprintf("%f,%f", wp.Longitude, wp.Latitude))
		}
		newCoords = append(newCoords, allCoords[i+1])
	}

	// ------------------------------
	// 3) Recalcula múltiplas rotas TOTAIS com diferentes parâmetros (como no CalculateRoutes)
	// ------------------------------
	var totalRoute TotalSummary

	coordsStr := strings.Join(newCoords, ";")
	baseOSRMURL := "http://34.207.174.233:5000/route/v1/driving/" + neturl.PathEscape(coordsStr)

	// URLs para diferentes tipos de rota (EXATAMENTE igual ao CalculateRoutes)
	osrmURLFast := baseOSRMURL + "?" + neturl.Values{
		"alternatives":      {"3"},
		"steps":             {"true"},
		"overview":          {"full"},
		"continue_straight": {"false"},
	}.Encode()

	osrmURLNoTolls := baseOSRMURL + "?" + neturl.Values{
		"alternatives": {"3"},
		"steps":        {"true"},
		"overview":     {"full"},
		"exclude":      {"toll"},
	}.Encode()

	osrmURLEfficient := baseOSRMURL + "?" + neturl.Values{
		"alternatives": {"3"},
		"steps":        {"true"},
		"overview":     {"full"},
		"exclude":      {"motorway"},
	}.Encode()

	// URLs adicionais para mais variações de rota
	osrmURLFastNoFerry := baseOSRMURL + "?" + neturl.Values{
		"alternatives":      {"3"},
		"steps":             {"true"},
		"overview":          {"full"},
		"continue_straight": {"false"},
		"exclude":           {"ferry"},
	}.Encode()

	osrmURLBalanced := baseOSRMURL + "?" + neturl.Values{
		"alternatives": {"3"},
		"steps":        {"true"},
		"overview":     {"full"},
		"exclude":      {"ferry,toll"},
	}.Encode()

	// Estrutura para resultados paralelos
	type osrmTotalResult struct {
		resp     OSRMResponse
		err      error
		category string
	}
	totalResultsCh := make(chan osrmTotalResult, 5) // Aumenta para 5 requisições

	// Função para fazer requisições paralelas
	makeTotalOSRMRequest := func(url, category, errMsg string) {
		resp, err := client.Get(url)
		if err != nil {
			totalResultsCh <- osrmTotalResult{err: fmt.Errorf("%s: %w", errMsg, err), category: category}
			return
		}
		defer resp.Body.Close()
		var osrmResp OSRMResponse
		if err := json.NewDecoder(resp.Body).Decode(&osrmResp); err != nil {
			totalResultsCh <- osrmTotalResult{err: fmt.Errorf("erro ao decodificar resposta OSRM (%s): %w", category, err), category: category}
			return
		}
		if osrmResp.Code != "Ok" || len(osrmResp.Routes) == 0 {
			totalResultsCh <- osrmTotalResult{err: fmt.Errorf("OSRM (%s) retornou erro ou nenhuma rota encontrada", category), category: category}
			return
		}
		totalResultsCh <- osrmTotalResult{resp: osrmResp, category: category}
	}

	// Lança requisições paralelas - agora com mais variações
	go makeTotalOSRMRequest(osrmURLFast, "fastest", "erro na requisição OSRM total (rota rápida)")
	go makeTotalOSRMRequest(osrmURLNoTolls, "cheapest", "erro na requisição OSRM total (rota com menos pedágio)")
	go makeTotalOSRMRequest(osrmURLEfficient, "efficient", "erro na requisição OSRM total (rota eficiente)")
	go makeTotalOSRMRequest(osrmURLFastNoFerry, "fastest_no_ferry", "erro na requisição OSRM total (rota rápida sem ferry)")
	go makeTotalOSRMRequest(osrmURLBalanced, "balanced", "erro na requisição OSRM total (rota balanceada)")

	var osrmTotalRespFast, osrmTotalRespNoTolls, osrmTotalRespEfficient, osrmTotalRespFastNoFerry, osrmTotalRespBalanced OSRMResponse

	// Coleta resultados
	for i := 0; i < 5; i++ {
		res := <-totalResultsCh
		if res.err != nil {
			continue // Ignora erros e tenta outras rotas
		}
		switch res.category {
		case "fastest":
			osrmTotalRespFast = res.resp
		case "cheapest":
			osrmTotalRespNoTolls = res.resp
		case "efficient":
			osrmTotalRespEfficient = res.resp
		case "fastest_no_ferry":
			osrmTotalRespFastNoFerry = res.resp
		case "balanced":
			osrmTotalRespBalanced = res.resp
		}
	}

	// Processa múltiplas rotas como no CalculateRoutes
	processTotalRoutes := func(osrmResp OSRMResponse, routeCategory string) []TotalSummary {
		var totalRoutes []TotalSummary
		for _, route := range osrmResp.Routes {
			// Monta waypoints p/ URL do Google (endereços originais + via:lat,lng dos desvios)
			waypointsForURL := make([]string, 0, len(waypoints)+len(extraWaypointsForURL))
			if len(waypoints) > 2 {
				// Insere "via:" antes dos pontos intermediários de endereço
				waypointsForURL = append(waypointsForURL, extraWaypointsForURL...)
				waypointsForURL = append(waypointsForURL, waypoints[1:len(waypoints)-1]...)
			} else {
				waypointsForURL = append(waypointsForURL, extraWaypointsForURL...)
			}

			if len(waypointsForURL) == 0 {
				waypointsForURL = []string{
					fmt.Sprintf("%f,%f", originLocation.Latitude, originLocation.Longitude),
					fmt.Sprintf("%f,%f", destinationLocation.Latitude, destinationLocation.Longitude),
				}
			}

			routeSummary := s.createTotalSummaryWithURLWaypoints(route, originLocation, destinationLocation, waypoints, waypointsForURL, s.convertCoordinatesToCEPRequest(data))

			// Força o uso dos valores acumulados dos segmentos para duração e distância mais precisos
			distText, distVal := formatDistance(totalDistance)
			durText, durVal := formatDuration(totalDuration)
			routeSummary.TotalDistance = Distance{Text: distText, Value: distVal}
			routeSummary.TotalDuration = Duration{Text: durText, Value: durVal}

			// Recalcula custo de combustível com a distância correta
			avgConsumption := (data.ConsumptionCity + data.ConsumptionHwy) / 2
			totalKm := totalDistance / 1000
			routeSummary.TotalFuelCost = math.Round((data.Price / avgConsumption) * totalKm)

			// Adiciona categoria da rota
			routeSummary.RouteType = routeCategory

			totalRoutes = append(totalRoutes, routeSummary)
		}
		return totalRoutes
	}

	// Processa rotas exatamente como no CalculateRoutes
	routesFast := processTotalRoutes(osrmTotalRespFast, "fastest")
	routesNoTolls := processTotalRoutes(osrmTotalRespNoTolls, "cheapest")
	routesEfficient := processTotalRoutes(osrmTotalRespEfficient, "efficient")
	routesFastNoFerry := processTotalRoutes(osrmTotalRespFastNoFerry, "fastest_no_ferry")
	routesBalanced := processTotalRoutes(osrmTotalRespBalanced, "balanced")

	// Combina todas as rotas exatamente como no CalculateRoutes
	var allTotalRoutes []TotalSummary
	switch strings.ToLower(data.TypeRoute) {
	case "efficient", "eficiente":
		if len(routesEfficient) > 0 {
			totalRoute = routesEfficient[0]
		}
		allTotalRoutes = routesEfficient
	case "fastest", "fast", "rapida":
		if len(routesFast) > 0 {
			totalRoute = routesFast[0]
		}
		// Combina rotas rápidas e suas variações
		allTotalRoutes = append(routesFast, routesFastNoFerry...)
	case "cheapest", "cheap", "barata":
		if len(routesNoTolls) > 0 {
			totalRoute = routesNoTolls[0]
		}
		// Combina rotas baratas e suas variações
		allTotalRoutes = append(routesNoTolls, routesBalanced...)
	default:
		// EXATAMENTE como no CalculateRoutes: combina TODAS as rotas
		allTotalRoutes = append(append(append(append(routesFast, routesNoTolls...), routesEfficient...), routesFastNoFerry...), routesBalanced...)
		if len(allTotalRoutes) > 0 {
			totalRoute = allTotalRoutes[0]
		}
	}

	// ------------------------------
	// 4) Fallbacks (rota total sem desvio / agregados)
	// ------------------------------
	if totalRoute.TotalDistance.Value == 0 {
		// tenta rota total padrão para obter polyline
		baseCoords := strings.Join(allCoords, ";")
		urlBase := fmt.Sprintf(
			"http://34.207.174.233:5000/route/v1/driving/%s?alternatives=0&steps=true&overview=full&continue_straight=false",
			neturl.PathEscape(baseCoords),
		)

		if resp, err := client.Get(urlBase); err == nil {
			defer resp.Body.Close()
			var osrmResp OSRMResponse
			if err := json.NewDecoder(resp.Body).Decode(&osrmResp); err == nil && len(osrmResp.Routes) > 0 {
				route := osrmResp.Routes[0]
				totalRoute = s.createTotalSummary(route, originLocation, destinationLocation, waypoints, s.convertCoordinatesToCEPRequest(data))
			}
		}
	}

	if totalRoute.TotalDistance.Value == 0 {
		// último recurso: usa agregados
		distText, distVal := formatDistance(totalDistance)
		durText, durVal := formatDuration(totalDuration)

		avgConsumption := (data.ConsumptionCity + data.ConsumptionHwy) / 2
		totalKm := totalDistance / 1000
		totalFuelCost := math.Round((data.Price / avgConsumption) * totalKm)

		// Buscar pedágios para a rota base (sem desvios)
		baseCoords := strings.Join(allCoords, ";")
		urlBase := fmt.Sprintf(
			"http://34.207.174.233:5000/route/v1/driving/%s?alternatives=0&steps=true&overview=full&continue_straight=false",
			neturl.PathEscape(baseCoords),
		)

		var tolls []Toll
		var totalTollCost float64
		if resp, err := client.Get(urlBase); err == nil {
			defer resp.Body.Close()
			var osrmResp OSRMResponse
			if err := json.NewDecoder(resp.Body).Decode(&osrmResp); err == nil && len(osrmResp.Routes) > 0 {
				route := osrmResp.Routes[0]
				tolls, _ = s.findTollsOnRoute(ctx, route.Geometry, data.Type, float64(data.Axles))
				for _, toll := range tolls {
					totalTollCost += toll.CashCost
				}
			}
		}

		originAddress := waypoints[0]
		destAddress := waypoints[len(waypoints)-1]
		waypointStr := ""
		if len(waypoints) > 2 {
			waypointStr = "&waypoints=" + neturl.QueryEscape(strings.Join(waypoints[1:len(waypoints)-1], "|"))
		}

		googleURL := fmt.Sprintf(
			"https://www.google.com/maps/dir/?api=1&origin=%s&destination=%s%s&travelmode=driving",
			neturl.QueryEscape(originAddress),
			neturl.QueryEscape(destAddress),
			waypointStr,
		)
		currentTimeMillis := (time.Now().UnixNano() + int64(totalDuration*float64(time.Second))) / int64(time.Millisecond)
		wazeURL := fmt.Sprintf(
			"https://www.waze.com/pt-BR/live-map/directions/br?to=%s&from=%s&time=%d&reverse=yes",
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
				Address:  normalizeAddress(destAddress),
			},
			TotalDistance: Distance{Text: distText, Value: distVal},
			TotalDuration: Duration{Text: durText, Value: durVal},
			URL:           googleURL,
			URLWaze:       wazeURL,
			Tolls:         tolls,
			TotalTolls:    math.Round(totalTollCost*100) / 100,
			Polyline:      "",
			TotalFuelCost: totalFuelCost,
		}
	}

	return totalRoute, allTotalRoutes
}

func (s *Service) getActiveRiskZones(ctx context.Context, id int64) ([]RiskZone, error) {
	if s.RiskZonesRepository == nil {

		return []RiskZone{}, nil
	}

	org := sql.NullInt64{
		Int64: id,
		Valid: true,
	}
	dbZones, err := s.RiskZonesRepository.GetAllZonasRiscoService(ctx, org)
	if err != nil {
		log.Printf("Erro ao buscar zonas de risco: %v", err)
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

	return riskZones, nil
}

func (s *Service) CheckRouteForRiskZones(riskZones []RiskZone, originLat, originLon, destLat, destLon float64) (bool, LocationHisk) {

	if len(riskZones) == 0 {

		return false, LocationHisk{}
	}

	// Primeiro, calcular a rota real com OSRM para verificar todos os pontos
	client := http.Client{Timeout: 15 * time.Second}
	coords := fmt.Sprintf("%f,%f;%f,%f", originLon, originLat, destLon, destLat)
	url := fmt.Sprintf("http://34.207.174.233:5000/route/v1/driving/%s?overview=full&steps=true", neturl.PathEscape(coords))

	resp, err := client.Get(url)
	if err != nil {
		// Fallback para verificação de linha reta
		return s.checkRouteForRiskZonesFallback(riskZones, originLat, originLon, destLat, destLon), LocationHisk{}
	}
	defer resp.Body.Close()

	var osrmResp OSRMResponse
	if err := json.NewDecoder(resp.Body).Decode(&osrmResp); err != nil {
		// Fallback para verificação de linha reta
		return s.checkRouteForRiskZonesFallback(riskZones, originLat, originLon, destLat, destLon), LocationHisk{}
	}

	if len(osrmResp.Routes) == 0 {
		// Fallback para verificação de linha reta
		return s.checkRouteForRiskZonesFallback(riskZones, originLat, originLon, destLat, destLon), LocationHisk{}
	}

	route := osrmResp.Routes[0]

	// Verificar se a rota real passa por zonas de risco
	return s.checkRouteGeometryForRiskZones(riskZones, route.Geometry, originLat, originLon, destLat, destLon)
}

func (s *Service) isPointInRiskZone(lat, lng float64, zone RiskZone) bool {
	distance := s.haversineDistance(lat, lng, zone.Lat, zone.Lng)
	isInside := distance <= float64(zone.Radius)

	return isInside
}

func (s *Service) doesRouteCrossRiskZone(originLat, originLon, destLat, destLon float64, zone RiskZone) bool {
	// Distância do centro da zona até a linha da rota
	distanceToRoute := s.distancePointToLine(zone.Lat, zone.Lng, originLat, originLon, destLat, destLon)

	// Se a distância for menor que o raio, a rota cruza a zona
	crossesZone := distanceToRoute <= float64(zone.Radius)

	return crossesZone
}

func (s *Service) projectToMeters(latRef float64, lat, lon float64) (xEast, yNorth float64) {
	const mPerDegLat = 111320.0
	mPerDegLon := 111320.0 * math.Cos(latRef*math.Pi/180.0)
	yNorth = (lat) * mPerDegLat
	xEast = (lon) * mPerDegLon
	return
}

func (s *Service) unprojectFromMeters(latRef float64, xEast, yNorth float64) (lat, lon float64) {
	const mPerDegLat = 111320.0
	mPerDegLon := 111320.0 * math.Cos(latRef*math.Pi/180.0)
	lat = yNorth / mPerDegLat
	lon = xEast / mPerDegLon
	return
}

// distância de ponto a **segmento** (não linha infinita), em metros
func (s *Service) distancePointToLine(pointLat, pointLng, aLat, aLng, bLat, bLng float64) float64 {
	latRef := (aLat + bLat) / 2.0
	ax, ay := s.projectToMeters(latRef, aLat, aLng)
	bx, by := s.projectToMeters(latRef, bLat, bLng)
	px, py := s.projectToMeters(latRef, pointLat, pointLng)

	vx, vy := bx-ax, by-ay
	wx, wy := px-ax, py-ay

	den := vx*vx + vy*vy
	if den == 0 {
		return math.Hypot(px-ax, py-ay)
	}
	t := (wx*vx + wy*vy) / den
	if t < 0 {
		return math.Hypot(px-ax, py-ay)
	}
	if t > 1 {
		return math.Hypot(px-bx, py-by)
	}
	cx := ax + t*vx
	cy := ay + t*vy
	return math.Hypot(px-cx, py-cy)
}

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

	const arcExtraBuffer = 200.0
	const arcPoints = 2
	const entryExitPush = 80.0

	// -------- util: calcula rota OSRM (ignora risco) com a sequência atual de via-points
	routeRaw := func(wps []Location, tag string) (OSRMRoute, bool) {
		coords := fmt.Sprintf("%f,%f", originLon, originLat)
		for _, wp := range wps {
			coords += fmt.Sprintf(";%f,%f", wp.Longitude, wp.Latitude)
		}
		coords += fmt.Sprintf(";%f,%f", destLon, destLat)

		profile := "driving"
		if data.Type != "" {
			profile = data.Type // ex: "car", "truck"
		}

		u := fmt.Sprintf("http://34.207.174.233:5000/route/v1/%s/%s?alternatives=0&steps=true&overview=full&continue_straight=false",
			profile, neturl.PathEscape(coords))

		// Cliente com timeout reduzido para melhor performance
		fastClient := http.Client{Timeout: 15 * time.Second}
		resp, err := fastClient.Get(u)
		if err != nil || resp.StatusCode != 200 {
			u = fmt.Sprintf("http://34.207.174.233:5000/route/v1/driving/%s?alternatives=0&steps=true&overview=full&continue_straight=false",
				neturl.PathEscape(coords))
			resp, err = fastClient.Get(u)
		}

		defer resp.Body.Close()

		var osrmResp OSRMResponse
		if json.NewDecoder(resp.Body).Decode(&osrmResp) != nil || len(osrmResp.Routes) == 0 {
			return OSRMRoute{}, false
		}
		return osrmResp.Routes[0], true
	}

	// -------- util: valida a rota resultante contra TODAS as zonas (não apenas a atual)
	tryRoute := func(wps []Location, tag string) (OSRMRoute, bool) {
		r, ok := routeRaw(wps, tag)
		if !ok {
			return OSRMRoute{}, false
		}
		if hasAny, _ := s.checkRouteGeometryForRiskZones(riskZones, r.Geometry, originLat, originLon, destLat, destLon); hasAny {

			return OSRMRoute{}, false
		}

		return r, true
	}

	// -------- util: snap simples
	snapMany := func(label string, pts []Location) []Location {
		out := make([]Location, len(pts))
		copy(out, pts)
		for i := range out {
			if lat, lon, ok := s.osrmNearestSnap(client, out[i].Latitude, out[i].Longitude); ok {
				out[i].Latitude, out[i].Longitude = lat, lon
			}
		}
		return out
	}

	// -------- util: snap que GARANTE ficar fora da zona
	snapOutsideMany := func(label string, pts []Location, zone RiskZone) []Location {
		out := make([]Location, len(pts))
		copy(out, pts)
		for i := range out {
			p := out[i]
			for step := 0; step < 6; step++ {
				if lat, lon, ok := s.osrmNearestSnap(client, p.Latitude, p.Longitude); ok {
					if s.haversineDistance(lat, lon, zone.Lat, zone.Lng) > float64(zone.Radius)+5 {
						out[i] = Location{Latitude: lat, Longitude: lon}
						break
					}
				}
				// se falhou snap ou ainda ficou dentro, empurra mais e tenta de novo
				p = s.pushAwayFromCenter(p, zone, 120.0)
			}
		}
		return out
	}

	// -------- util: evitar via-points duplicados (após snap)
	sameLL := func(a, b Location) bool {
		// tolerância ~2m
		return math.Abs(a.Latitude-b.Latitude) < 0.00002 && math.Abs(a.Longitude-b.Longitude) < 0.00002
	}
	appendUnique := func(dst []Location, src ...Location) []Location {
		for _, p := range src {
			dup := false
			for _, q := range dst {
				if sameLL(p, q) {
					dup = true
					break
				}
			}
			if !dup {
				dst = append(dst, p)
			}
		}
		return dst
	}

	// 0) se não tem risco → direta
	if hasRisk, _ := s.CheckRouteForRiskZones(riskZones, originLat, originLon, destLat, destLon); !hasRisk {
		return s.calculateDirectRoute(ctx, client, originLat, originLon, destLat, destLon, originGeocode, destGeocode, data)
	}

	// 1) rota base (sem via-points) e util p/ coletar cruzamentos
	_, ok := routeRaw(nil, "init")
	if !ok {
		return s.calculateDirectRoute(ctx, client, originLat, originLon, destLat, destLon, originGeocode, destGeocode, data)
	}
	collectCrossings := func(geometry string) []RiskOffsets {
		var offs []RiskOffsets
		for _, z := range riskZones {
			if !z.Status {
				continue
			}
			if off, ok := s.computeRiskOffsetsFromGeometry(geometry, z, 1000); ok {
				offs = append(offs, off)
			}
		}
		sort.Slice(offs, func(i, j int) bool { return offs[i].EntryCum < offs[j].EntryCum })
		return offs
	}

	// 2) usar WAYPOINTS do usuário (se houver)
	if len(data.Waypoints) > 0 {
		if userWps, err := s.normalizeUserWaypoints(data.Waypoints); err == nil && len(userWps) > 0 {

			userWps = snapMany("user", userWps)
			if r, ok := tryRoute(userWps, "user"); ok {
				tolls, _ := s.findTollsOnRoute(ctx, r.Geometry, data.Type, float64(data.Axles))
				sum := s.createRouteSummary(r, "desvio_usuario", originGeocode, destGeocode, data, tolls)
				return []RouteSummary{sum}
			}

		}
	}

	// ---------- LOOP MULTI-ZONAS ----------
	accumWps := []Location{}
	var detourPoints []DetourPoint
	maxIters := 20

	for iter := 0; iter < maxIters; iter++ {
		// recalcula rota com via-points já inseridos
		rCurr, ok := routeRaw(accumWps, fmt.Sprintf("iter_%d_curr", iter))
		if !ok {
			break
		}
		crossings := collectCrossings(rCurr.Geometry)
		if len(crossings) == 0 {
			// tenta finalizar (testa globalmente dentro de tryRoute)
			if rFinal, ok := tryRoute(accumWps, "final"); ok {
				tolls, _ := s.findTollsOnRoute(ctx, rFinal.Geometry, data.Type, float64(data.Axles))
				sum := s.createRouteSummary(rFinal, "desvio_multi_zonas", originGeocode, destGeocode, data, tolls)
				if len(detourPoints) > 0 {
					sum.Detour = &DetourPlan{Source: "multi_zonas", Points: detourPoints}
				}
				return []RouteSummary{sum}
			}

			continue
		}

		// pega a PRIMEIRA zona ainda cruzada na polyline atual
		off := crossings[0]

		// ===== Etapa 3: LATERAL (menor arco)
		shortSeq := s.assembleLateralDetour(off.Entry, off.Exit, off.Zone, arcPoints, arcExtraBuffer, entryExitPush, false)
		shortSeq = snapOutsideMany("short_seq", shortSeq, off.Zone)
		if r2, ok := routeRaw(append(append([]Location{}, accumWps...), shortSeq...), fmt.Sprintf("iter_%d_short", iter)); ok {
			if hasThis, _ := s.checkRouteGeometryForRiskZones([]RiskZone{off.Zone}, r2.Geometry, originLat, originLon, destLat, destLon); !hasThis {
				accumWps = appendUnique(accumWps, shortSeq...)
				for i := range shortSeq {
					detourPoints = append(detourPoints, DetourPoint{
						Name:     fmt.Sprintf("short_%d_%d", iter+1, i+1),
						Location: shortSeq[i],
					})
				}
				continue
			}
		}

		// ===== Etapa 4: LATERAL (arco oposto)
		longSeq := s.assembleLateralDetour(off.Entry, off.Exit, off.Zone, arcPoints, arcExtraBuffer, entryExitPush, true)
		longSeq = snapOutsideMany("long_seq", longSeq, off.Zone)
		if r2, ok := routeRaw(append(append([]Location{}, accumWps...), longSeq...), fmt.Sprintf("iter_%d_long", iter)); ok {
			if hasThis, _ := s.checkRouteGeometryForRiskZones([]RiskZone{off.Zone}, r2.Geometry, originLat, originLon, destLat, destLon); !hasThis {
				accumWps = appendUnique(accumWps, longSeq...)
				for i := range longSeq {
					detourPoints = append(detourPoints, DetourPoint{
						Name:     fmt.Sprintf("long_%d_%d", iter+1, i+1),
						Location: longSeq[i],
					})
				}
				continue
			}
		}

		// ===== Etapa 5: Fallback A/B
		wpA, wpB := s.computeBypassWaypoints(originLat, originLon, destLat, destLon, off.Zone)
		abSeq := snapOutsideMany("ab", []Location{wpA, wpB}, off.Zone)
		if r2, ok := routeRaw(append(append([]Location{}, accumWps...), abSeq...), fmt.Sprintf("iter_%d_ab", iter)); ok {
			if hasThis, _ := s.checkRouteGeometryForRiskZones([]RiskZone{off.Zone}, r2.Geometry, originLat, originLon, destLat, destLon); !hasThis {
				accumWps = appendUnique(accumWps, abSeq...)
				for i := range abSeq {
					detourPoints = append(detourPoints, DetourPoint{
						Name:     fmt.Sprintf("ab_%d_%d", iter+1, i+1),
						Location: abSeq[i],
					})
				}
				continue
			}
		}

		// ===== Etapa 6: AB escalado
		scales := []float64{1.5, 2.0, 3.0}
		scaledOK := false
		for _, sc := range scales {
			wpA2 := s.pushAwayFromCenter(wpA, off.Zone, float64(off.Zone.Radius)*(sc-1)+500)
			wpB2 := s.pushAwayFromCenter(wpB, off.Zone, float64(off.Zone.Radius)*(sc-1)+500)
			abScaled := snapOutsideMany(fmt.Sprintf("ab_scale_%.1f", sc), []Location{wpA2, wpB2}, off.Zone)
			if r2, ok := routeRaw(append(append([]Location{}, accumWps...), abScaled...), fmt.Sprintf("iter_%d_ab_scale_%.1f", iter, sc)); ok {
				if hasThis, _ := s.checkRouteGeometryForRiskZones([]RiskZone{off.Zone}, r2.Geometry, originLat, originLon, destLat, destLon); !hasThis {
					accumWps = appendUnique(accumWps, abScaled...)
					for i := range abScaled {
						detourPoints = append(detourPoints, DetourPoint{
							Name:     fmt.Sprintf("ab_scaled_%.1f_%d", sc, i+1),
							Location: abScaled[i],
						})
					}
					scaledOK = true
					break
				}
			}
		}
		if scaledOK {
			continue
		}

		// ===== Etapa 7: se não conseguimos remover ESTA zona, encerra e devolve aviso
		if rTry, ok := routeRaw(accumWps, fmt.Sprintf("iter_%d_curr_post", iter)); ok {
			if hasAny, _ := s.checkRouteGeometryForRiskZones(riskZones, rTry.Geometry, originLat, originLon, destLat, destLon); hasAny {

				sum := s.createRouteSummaryWithRiskWarning(originLat, originLon, destLat, destLon, originGeocode, destGeocode, data)
				return []RouteSummary{sum}
			}
		}
	}

	// ---------- GUARD SWEEP: INÍCIO E CHEGADA ----------
	if rEnd, ok := routeRaw(accumWps, "pre_guard"); ok {
		startWindow := 2200.0 // m próximos ao início
		endWindow := 2200.0   // m próximos ao destino

		offs := s.detectAllCrossingsFromGeometry(rEnd.Geometry, riskZones, 1000)

		// 1) Departure guard (se cruzar muito perto da origem)
		if len(offs) > 0 && offs[0].EntryCum <= startWindow {
			first := offs[0]
			origin := Location{Latitude: originLat, Longitude: originLon}
			for b := 80.0; b <= 500.0; b += 10.0 {
				guard := s.computeArrivalGuardPoint(first.Zone, origin, b)
				guard = snapOutsideMany("dep_guard", []Location{guard}, first.Zone)[0]

				test := append([]Location{guard}, accumWps...)
				if r2, ok2 := routeRaw(test, fmt.Sprintf("dep_guard_%.0f", b)); ok2 {
					if hasAny, _ := s.checkRouteGeometryForRiskZones(riskZones, r2.Geometry, originLat, originLon, destLat, destLon); !hasAny {
						accumWps = test
						break
					}
				}
			}

			if rEnd2, ok2 := routeRaw(accumWps, "post_dep_guard"); ok2 {
				rEnd = rEnd2
			}
		}

		// 2) Arrival guard (se cruzar muito perto do destino)
		offs = s.detectAllCrossingsFromGeometry(rEnd.Geometry, riskZones, 1000)
		if len(offs) > 0 && (rEnd.Distance-offs[len(offs)-1].EntryCum) <= endWindow {
			last := offs[len(offs)-1]
			dest := Location{Latitude: destLat, Longitude: destLon}
			for b := 80.0; b <= 500.0; b += 10.0 {
				guard := s.computeArrivalGuardPoint(last.Zone, dest, b)
				guard = snapOutsideMany("arr_guard", []Location{guard}, last.Zone)[0]

				test := append(append([]Location{}, accumWps...), guard)
				if r2, ok2 := routeRaw(test, fmt.Sprintf("arr_guard_%.0f", b)); ok2 {
					if hasAny, _ := s.checkRouteGeometryForRiskZones(riskZones, r2.Geometry, originLat, originLon, destLat, destLon); !hasAny {
						accumWps = append(accumWps, guard)
						break
					}
				}
			}

		}
	}
	// ---------- fim do guard sweep ----------

	// tenta "best_effort" já com possíveis guards
	if r, ok := tryRoute(accumWps, "best_effort"); ok {
		tolls, _ := s.findTollsOnRoute(ctx, r.Geometry, data.Type, float64(data.Axles))
		sum := s.createRouteSummary(r, "desvio_multi_zonas_best_effort", originGeocode, destGeocode, data, tolls)
		if len(detourPoints) > 0 {
			sum.Detour = &DetourPlan{Source: "multi_zonas", Points: detourPoints}
		}
		return []RouteSummary{sum}
	}

	// fallback final: rota direta com aviso

	return []RouteSummary{
		s.createRouteSummaryWithRiskWarning(originLat, originLon, destLat, destLon, originGeocode, destGeocode, data),
	}
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

func (s *Service) computeBypassWaypoints(originLat, originLon, destLat, destLon float64, zone RiskZone) (Location, Location) {
	// vetor da rota em metros
	latRef := (originLat + destLat) / 2.0
	ax, ay := s.projectToMeters(latRef, originLat, originLon)
	bx, by := s.projectToMeters(latRef, destLat, destLon)
	cx, cy := s.projectToMeters(latRef, zone.Lat, zone.Lng)

	vx, vy := bx-ax, by-ay
	// normal unitária
	nx, ny := -vy, vx
	mag := math.Hypot(nx, ny)
	if mag == 0 {
		nx, ny, mag = 0, 1, 1
	}
	nx /= mag
	ny /= mag

	// empurra para fora do círculo: raio + buffer
	buffer := 500.0 // 500 m adicional
	safe := float64(zone.Radius) + buffer

	// projeta o centro da zona na reta AB para pegar o "miolo" por onde cruzaria
	den := vx*vx + vy*vy
	var tx float64
	if den == 0 {
		tx = 0.5
	} else {
		tx = ((cx-ax)*vx + (cy-ay)*vy) / den
	}
	tx = math.Min(1, math.Max(0, tx))
	midX := ax + tx*vx
	midY := ay + tx*vy

	// dois pontos a lados opostos do círculo, fora da borda
	ax2 := midX + nx*safe
	ay2 := midY + ny*safe
	bx2 := midX - nx*safe
	by2 := midY - ny*safe

	// volta pra lat/lon
	aLatOff, aLonOff := s.unprojectFromMeters(latRef, ax2, ay2)
	bLatOff, bLonOff := s.unprojectFromMeters(latRef, bx2, by2)

	wpA := Location{Latitude: aLatOff + (0), Longitude: aLonOff + (0)}
	wpB := Location{Latitude: bLatOff + (0), Longitude: bLonOff + (0)}

	// ordena pelo avanço a partir da origem
	dA := s.haversineDistance(originLat, originLon, wpA.Latitude, wpA.Longitude)
	dB := s.haversineDistance(originLat, originLon, wpB.Latitude, wpB.Longitude)
	if dA <= dB {
		return wpA, wpB
	}
	return wpB, wpA
}

// computeBypassFromRouteGeometry tenta pegar pontos de tangência usando a polyline real da rota
func (s *Service) computeBypassFromRouteGeometry(originLat, originLon, destLat, destLon float64, zone RiskZone) (Location, Location, bool) {
	// Consulta uma rota OSRM simples entre origem e destino
	coords := fmt.Sprintf("%f,%f;%f,%f", originLon, originLat, destLon, destLat)
	osrmURL := fmt.Sprintf("http://34.207.174.233:5000/route/v1/driving/%s?alternatives=0&steps=true&overview=full", neturl.PathEscape(coords))
	client := http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(osrmURL)
	if err != nil {
		return Location{}, Location{}, false
	}
	defer resp.Body.Close()
	var osrmResp OSRMResponse
	if err := json.NewDecoder(resp.Body).Decode(&osrmResp); err != nil || len(osrmResp.Routes) == 0 {
		return Location{}, Location{}, false
	}
	// Decodificar polyline e encontrar primeiro e último pontos onde a rota toca a borda da zona (entrada e saída)
	points, err := s.decodePolylineOSRM(osrmResp.Routes[0].Geometry)
	if err != nil || len(points) < 3 {
		return Location{}, Location{}, false
	}
	// Varre pontos para encontrar o primeiro que entra no círculo e o primeiro que sai
	inCircle := func(p Location) bool {
		return s.haversineDistance(p.Latitude, p.Longitude, zone.Lat, zone.Lng) <= float64(zone.Radius)
	}
	var idxEnter, idxExit int = -1, -1
	wasIn := false
	for i := 0; i < len(points); i++ {
		inside := inCircle(points[i])
		if !wasIn && inside {
			// borda de entrada: usa o ponto anterior como tangência A se existir
			idxEnter = max(0, i-1)
		}
		if wasIn && !inside {
			// borda de saída: usa este ponto como tangência B
			idxExit = i
			break
		}
		wasIn = inside
	}
	if idxEnter >= 0 && idxExit > idxEnter {
		wpA := points[idxEnter]
		wpB := points[idxExit]
		// Empurra levemente para fora do círculo (50m) para assegurar que o roteador contorne
		wpA = s.pushAwayFromCenter(wpA, zone, 50)
		wpB = s.pushAwayFromCenter(wpB, zone, 50)
		return wpA, wpB, true
	}
	return Location{}, Location{}, false
}

func (s *Service) pushAwayFromCenter(p Location, zone RiskZone, meters float64) Location {
	// vetor do centro até p, normalizado, e empurra para fora
	bearingNorth := (p.Latitude - zone.Lat) * 111320.0
	bearingEast := (p.Longitude - zone.Lng) * 111320.0 * math.Cos(zone.Lat*math.Pi/180.0)
	mag := math.Hypot(bearingNorth, bearingEast)
	if mag == 0 {
		// empurra arbitrariamente para leste
		bearingNorth, bearingEast, mag = 0, 1, 1
	}
	bearingNorth /= mag
	bearingEast /= mag

	metersPerDegLat := 111320.0
	metersPerDegLon := 111320.0 * math.Cos(zone.Lat*math.Pi/180.0)
	return Location{
		Latitude:  p.Latitude + (bearingNorth*meters)/metersPerDegLat,
		Longitude: p.Longitude + (bearingEast*meters)/metersPerDegLon,
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// checkRouteForRiskZonesFallback verificação de fallback usando linha reta
func (s *Service) checkRouteForRiskZonesFallback(riskZones []RiskZone, originLat, originLon, destLat, destLon float64) bool {

	for _, zone := range riskZones {
		if !zone.Status {
			continue
		}

		// Verificar se a origem ou destino estão dentro da zona de risco
		originInZone := s.isPointInRiskZone(originLat, originLon, zone)
		destInZone := s.isPointInRiskZone(destLat, destLon, zone)

		if originInZone || destInZone {
			return true
		}

		// Verificar se a linha reta entre origem e destino cruza a zona de risco
		routeCrossesZone := s.doesRouteCrossRiskZone(originLat, originLon, destLat, destLon, zone)
		if routeCrossesZone {
			return true
		}
	}

	return false
}

// checkRouteGeometryForRiskZones verifica se a geometria da rota OSRM passa por zonas de risco
func (s *Service) checkRouteGeometryForRiskZones(riskZones []RiskZone, geometry string, originLat, originLon, destLat, destLon float64) (bool, LocationHisk) {

	// Decodificar a geometria Polyline do OSRM
	coordinates, err := s.decodePolylineOSRM(geometry)
	if err != nil {
		return s.checkRouteForRiskZonesFallback(riskZones, originLat, originLon, destLat, destLon), LocationHisk{}
	}

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
				return true, LocationHisk{
					CEP:       zone.Cep,
					Latitude:  zone.Lat,
					Longitude: zone.Lng,
				}
			}

			// Verificar se o segmento cruza a zona de risco
			segmentCrossesZone := s.doesRouteCrossRiskZone(point1.Latitude, point1.Longitude, point2.Latitude, point2.Longitude, zone)
			if segmentCrossesZone {
				return true, LocationHisk{
					CEP:       zone.Cep,
					Latitude:  zone.Lat,
					Longitude: zone.Lng,
				}
			}
		}
	}

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
		neturl.QueryEscape(normalizeAddress(originGeocode.FormattedAddress)),
		neturl.QueryEscape(normalizeAddress(destGeocode.FormattedAddress)),
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

	summary := RouteSummary{
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

	// 🔹 Tenta atualizar a duração com o Google Directions API
	if s.GoogleMapsAPIKey != "" {
		if durText, durVal, err := GetGoogleDurationWithTraffic(
			context.Background(),
			s.GoogleMapsAPIKey,
			originGeocode.FormattedAddress,
			destGeocode.FormattedAddress,
			nil, // se tiver waypoints, passe aqui
		); err == nil {
			summary.Duration = Duration{Text: durText, Value: float64(durVal)}
		} else {
			log.Printf("Google Directions API falhou: %v", err)
		}
	}

	return summary
}

// createRouteSummaryWithRiskWarning cria resumo de rota com aviso de risco
func (s *Service) createRouteSummaryWithRiskWarning(originLat, originLon, destLat, destLon float64, originGeocode, destGeocode GeocodeResult, data FrontInfoCEPRequest) RouteSummary {
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
		neturl.QueryEscape(normalizeAddress(originGeocode.FormattedAddress)),
		neturl.QueryEscape(normalizeAddress(destGeocode.FormattedAddress)),
	)

	wazeURL := fmt.Sprintf("https://www.waze.com/pt-BR/live-map/directions/br?to=place.%s&from=place.%s&reverse=yes",
		neturl.QueryEscape(destGeocode.PlaceID),
		neturl.QueryEscape(originGeocode.PlaceID),
	)

	return RouteSummary{
		RouteType:     data.TypeRoute,
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

// calculateTotalRouteWithAvoidance calcula rota total com desvios
func (s *Service) calculateTotalRouteWithAvoidance(ctx context.Context, client http.Client, riskZones []RiskZone, ceps []string, totalDistance, totalDuration float64, data FrontInfoCEPRequest) TotalSummary {

	// ------------------------------
	// 1) Monta lista base de coords e endereços
	// ------------------------------
	var allCoords []string
	var waypoints []string
	var originLocation, destinationLocation Location

	for idx, cep := range ceps {
		lat, lon, err := s.getCoordByCEP(ctx, cep)
		if err != nil {
			continue
		}
		allCoords = append(allCoords, fmt.Sprintf("%f,%f", lon, lat))

		reverse, _ := s.reverseGeocode(lat, lon)
		geocode, _ := s.getGeocodeAddress(ctx, reverse)
		waypoints = append(waypoints, geocode.FormattedAddress)

		if idx == 0 {
			originLocation = Location{Latitude: lat, Longitude: lon}
		}
		if idx == len(ceps)-1 {
			destinationLocation = Location{Latitude: lat, Longitude: lon}
		}
	}

	if len(allCoords) < 2 {
		return TotalSummary{}
	}

	// ------------------------------
	// 2) Para cada segmento, injeta via-points evitando TODAS as zonas cruzadas
	// ------------------------------
	newCoords := make([]string, 0, len(allCoords)+12) // espaço extra para vias
	newCoords = append(newCoords, allCoords[0])

	var extraWaypointsForURL []string
	var detourPtsTotal []Location // opcional para expor no TotalSummary

	for i := 0; i < len(allCoords)-1; i++ {
		var lon1, lat1, lon2, lat2 float64
		fmt.Sscanf(allCoords[i], "%f,%f", &lon1, &lat1)
		fmt.Sscanf(allCoords[i+1], "%f,%f", &lon2, &lat2)

		// Sempre vamos trabalhar numa lista de via-points ACUMULADOS por segmento
		segWps := []Location{} // via-points inseridos nesse segmento

		// helper para calcular a polyline do segmento COM os via-points atuais
		routeForSegment := func() (OSRMRoute, bool) {
			coords := fmt.Sprintf("%f,%f", lon1, lat1)
			for _, wp := range segWps {
				coords += fmt.Sprintf(";%f,%f", wp.Longitude, wp.Latitude)
			}
			coords += fmt.Sprintf(";%f,%f", lon2, lat2)

			profile := "driving"
			if data.Type != "" {
				profile = data.Type // ex: "car", "truck"
			}

			u := fmt.Sprintf("http://34.207.174.233:5000/route/v1/%s/%s?alternatives=0&steps=true&overview=full&continue_straight=false",
				profile, neturl.PathEscape(coords))

			// Cliente com timeout reduzido para melhor performance
			fastClient := http.Client{Timeout: 15 * time.Second}
			resp, err := fastClient.Get(u)
			if err != nil || resp.StatusCode != 200 {
				u = fmt.Sprintf("http://34.207.174.233:5000/route/v1/driving/%s?alternatives=0&steps=true&overview=full&continue_straight=false",
					neturl.PathEscape(coords))
				resp, err = fastClient.Get(u)
			}

			defer resp.Body.Close()
			var r OSRMResponse
			if json.NewDecoder(resp.Body).Decode(&r) != nil || len(r.Routes) == 0 {

				return OSRMRoute{}, false
			}
			return r.Routes[0], true
		}

		// 2.A) Calcula rota do segmento (sem vias) e coleta TODAS as zonas cruzadas
		if r0, ok := routeForSegment(); ok {
			crossings := s.detectAllCrossingsFromGeometry(r0.Geometry, riskZones, 1000)

			// 2.B) Itera enquanto ainda cruzar alguma zona
			iter := 0
			for len(crossings) > 0 && iter < 15 {
				iter++
				off := crossings[0] // trata a próxima na ordem do percurso

				// ====== Estratégia preferencial: 3 pontos no MESMO lado ======
				latRef, nx, ny := s.awayNormalForSegment(off.Before5km, off.After5km, off.Zone)
				injected := false

				if nx != 0 || ny != 0 {
					baseDist := float64(off.Zone.Radius) + 200.0 // raio + buffer
					mid := Location{
						Latitude:  (off.Entry.Latitude + off.Exit.Latitude) / 2,
						Longitude: (off.Entry.Longitude + off.Exit.Longitude) / 2,
					}
					buildSeq := func(scale float64) []Location {
						d := baseDist * scale
						p1 := s.offsetByNormal(off.Before5km, latRef, nx, ny, d)
						p2 := s.offsetByNormal(mid, latRef, nx, ny, d)
						p3 := s.offsetByNormal(off.After5km, latRef, nx, ny, d)
						return []Location{p1, p2, p3}
					}

					for _, sc := range []float64{1.0, 1.3, 1.6, 2.0} {
						cand := buildSeq(sc)
						// Snap GARANTINDO ficar fora da zona
						cand = s.snapOutsideMany(client, cand, off.Zone)

						// Testa rota do segmento com segWps + cand
						segTest := append(append([]Location{}, segWps...), cand...)
						if r2, ok2 := func() (OSRMRoute, bool) {
							coords := fmt.Sprintf("%f,%f", lon1, lat1)
							for _, wp := range segTest {
								coords += fmt.Sprintf(";%f,%f", wp.Longitude, wp.Latitude)
							}
							coords += fmt.Sprintf(";%f,%f", lon2, lat2)

							profile := "driving"
							if data.Type != "" {
								profile = data.Type // ex: "car", "truck"
							}

							u := fmt.Sprintf("http://34.207.174.233:5000/route/v1/%s/%s?alternatives=0&steps=true&overview=full&continue_straight=false",
								profile, neturl.PathEscape(coords))

							resp, err := client.Get(u)
							if err != nil || resp.StatusCode != 200 {
								u = fmt.Sprintf("http://34.207.174.233:5000/route/v1/driving/%s?alternatives=0&steps=true&overview=full&continue_straight=false",
									neturl.PathEscape(coords))
								resp, err = client.Get(u)
							}

							defer resp.Body.Close()
							var r OSRMResponse
							if json.NewDecoder(resp.Body).Decode(&r) != nil || len(r.Routes) == 0 {
								return OSRMRoute{}, false
							}
							return r.Routes[0], true
						}(); ok2 {
							// rota de teste não pode cruzar ESTA zona
							if hasThis, _ := s.checkRouteGeometryForRiskZones([]RiskZone{off.Zone}, r2.Geometry, lat1, lon1, lat2, lon2); !hasThis {
								segWps = append(segWps, cand...)
								for _, p := range cand {
									extraWaypointsForURL = append(extraWaypointsForURL, fmt.Sprintf("via:%f,%f", p.Latitude, p.Longitude))
									detourPtsTotal = append(detourPtsTotal, p)
								}
								injected = true
								break
							}
						}
					}
				}

				// ====== Fallback: lateral e A/B
				if !injected {
					// menor arco
					seq := s.assembleLateralDetour(off.Entry, off.Exit, off.Zone, 2, 200, 80, false)
					seq = s.snapOutsideMany(client, seq, off.Zone)
					tryInject := func(cand []Location) bool {
						test := append(append([]Location{}, segWps...), cand...)
						if r2, ok2 := func() (OSRMRoute, bool) {
							coords := fmt.Sprintf("%f,%f", lon1, lat1)
							for _, wp := range test {
								coords += fmt.Sprintf(";%f,%f", wp.Longitude, wp.Latitude)
							}
							coords += fmt.Sprintf(";%f,%f", lon2, lat2)

							profile := "driving"
							if data.Type != "" {
								profile = data.Type // ex: "car", "truck"
							}

							u := fmt.Sprintf("http://34.207.174.233:5000/route/v1/%s/%s?alternatives=0&steps=true&overview=full&continue_straight=false",
								profile, neturl.PathEscape(coords))

							resp, err := client.Get(u)
							if err != nil || resp.StatusCode != 200 {
								u = fmt.Sprintf("http://34.207.174.233:5000/route/v1/driving/%s?alternatives=0&steps=true&overview=full&continue_straight=false",
									neturl.PathEscape(coords))
								resp, err = client.Get(u)
							}

							defer resp.Body.Close()
							var r OSRMResponse
							if json.NewDecoder(resp.Body).Decode(&r) != nil || len(r.Routes) == 0 {
								return OSRMRoute{}, false
							}
							return r.Routes[0], true
						}(); ok2 {
							if hasThis, _ := s.checkRouteGeometryForRiskZones([]RiskZone{off.Zone}, r2.Geometry, lat1, lon1, lat2, lon2); !hasThis {
								segWps = append(segWps, cand...)
								for _, p := range cand {
									extraWaypointsForURL = append(extraWaypointsForURL, fmt.Sprintf("via:%f,%f", p.Latitude, p.Longitude))
									detourPtsTotal = append(detourPtsTotal, p)
								}
								return true
							}
						}
						return false
					}

					if !tryInject(seq) {
						// arco oposto
						seq2 := s.assembleLateralDetour(off.Entry, off.Exit, off.Zone, 2, 200, 80, true)
						seq2 = s.snapOutsideMany(client, seq2, off.Zone)
						if !tryInject(seq2) {
							// A/B padrão
							wpA, wpB := s.computeBypassWaypoints(lat1, lon1, lat2, lon2, off.Zone)
							ab := s.snapOutsideMany(client, []Location{wpA, wpB}, off.Zone)
							if !tryInject(ab) {
								// A/B escalado
								for _, sc := range []float64{1.5, 2.0, 3.0} {
									wpA2 := s.pushAwayFromCenter(wpA, off.Zone, float64(off.Zone.Radius)*(sc-1)+500)
									wpB2 := s.pushAwayFromCenter(wpB, off.Zone, float64(off.Zone.Radius)*(sc-1)+500)
									ab2 := s.snapOutsideMany(client, []Location{wpA2, wpB2}, off.Zone)
									if tryInject(ab2) {
										break
									}
								}
							}
						}
					}
				}

				// Recalcula rota do segmento com os via-points acumulados e atualiza a lista de crossings
				if rNow, okNow := routeForSegment(); okNow {
					crossings = s.detectAllCrossingsFromGeometry(rNow.Geometry, riskZones, 1000)
				} else {
					// sem rota após tentativa — interrompe
					break
				}
			} // while cruzar zona
		}

		// Append final do destino do segmento e materializa os via-points em newCoords
		for _, wp := range segWps {
			newCoords = append(newCoords, fmt.Sprintf("%f,%f", wp.Longitude, wp.Latitude))
		}
		newCoords = append(newCoords, allCoords[i+1])
	}

	// ------------------------------
	// 3) Recalcula a rota TOTAL com a sequência ajustada
	// ------------------------------
	var totalRoute TotalSummary

	coordsStr := strings.Join(newCoords, ";")
	urlTotal := fmt.Sprintf(
		"http://34.207.174.233:5000/route/v1/driving/%s?alternatives=0&steps=true&overview=full&continue_straight=false",
		neturl.PathEscape(coordsStr),
	)

	if resp, err := client.Get(urlTotal); err == nil {
		defer resp.Body.Close()
		var osrmResp OSRMResponse
		if err := json.NewDecoder(resp.Body).Decode(&osrmResp); err == nil && len(osrmResp.Routes) > 0 {
			route := osrmResp.Routes[0]

			// ---------- GUARD SWEEP NA ROTA TOTAL (INÍCIO E CHEGADA) ----------
			startWindow := 2200.0 // m
			endWindow := 2200.0   // m

			// helpers para injetar/remover
			injectFront := func(seq []Location) { // imediatamente após a origem
				if len(newCoords) >= 1 {
					head := newCoords[0]
					tail := append([]string{}, newCoords[1:]...)
					newCoords = []string{head}
					for _, p := range seq {
						newCoords = append(newCoords, fmt.Sprintf("%f,%f", p.Longitude, p.Latitude))
					}
					newCoords = append(newCoords, tail...)
				}
			}
			removeFront := func(n int) {
				if len(newCoords) >= 1+n {
					head := newCoords[0]
					newCoords = append([]string{head}, newCoords[1+n:]...)
				}
			}
			injectBeforeDest := func(seq []Location) { // imediatamente antes do destino
				if len(newCoords) >= 2 {
					destTail := newCoords[len(newCoords)-1]
					newCoords = newCoords[:len(newCoords)-1]
					for _, p := range seq {
						newCoords = append(newCoords, fmt.Sprintf("%f,%f", p.Longitude, p.Latitude))
					}
					newCoords = append(newCoords, destTail)
				}
			}
			removeBeforeDest := func(n int) {
				if len(newCoords) >= 1+n {
					destTail := newCoords[len(newCoords)-1]
					newCoords = append(newCoords[:len(newCoords)-1-n], destTail)
				}
			}

			// 1) Departure guard sweep
			for range []int{0} {
				cross := s.detectAllCrossingsFromGeometry(route.Geometry, riskZones, 1000)
				if len(cross) == 0 || cross[0].EntryCum > startWindow {
					break
				}
				first := cross[0]
				origin := originLocation
				for b := 80.0; b <= 500.0; b += 10.0 {
					guard := s.computeArrivalGuardPoint(first.Zone, origin, b)
					guard = s.snapOutsideMany(client, []Location{guard}, first.Zone)[0]
					injectFront([]Location{guard})

					coordsStr = strings.Join(newCoords, ";")
					urlTotal = fmt.Sprintf("http://34.207.174.233:5000/route/v1/driving/%s?alternatives=0&steps=true&overview=full&continue_straight=false", neturl.PathEscape(coordsStr))
					if resp2, err2 := client.Get(urlTotal); err2 == nil {
						defer resp2.Body.Close()
						var osrm2 OSRMResponse
						if json.NewDecoder(resp2.Body).Decode(&osrm2) == nil && len(osrm2.Routes) > 0 {
							route = osrm2.Routes[0]
							if has, _ := s.checkRouteGeometryForRiskZones(riskZones, route.Geometry, originLocation.Latitude, originLocation.Longitude, destinationLocation.Latitude, destinationLocation.Longitude); !has {
								break
							}
						}
					}
					removeFront(1)
				}

			}

			// 2) Arrival guard sweep
			for range []int{0} {
				cross := s.detectAllCrossingsFromGeometry(route.Geometry, riskZones, 1000)
				if len(cross) == 0 {
					break
				}
				last := cross[len(cross)-1]
				if route.Distance-last.EntryCum > endWindow {
					break // últimos cruzamentos estão longe do destino
				}
				for b := 80.0; b <= 500.0; b += 10.0 {
					guard := s.computeArrivalGuardPoint(last.Zone, destinationLocation, b)
					guard = s.snapOutsideMany(client, []Location{guard}, last.Zone)[0]
					injectBeforeDest([]Location{guard})

					coordsStr = strings.Join(newCoords, ";")
					urlTotal = fmt.Sprintf("http://34.207.174.233:5000/route/v1/driving/%s?alternatives=0&steps=true&overview=full&continue_straight=false", neturl.PathEscape(coordsStr))
					if resp2, err2 := client.Get(urlTotal); err2 == nil {
						defer resp2.Body.Close()
						var osrm2 OSRMResponse
						if json.NewDecoder(resp2.Body).Decode(&osrm2) == nil && len(osrm2.Routes) > 0 {
							route = osrm2.Routes[0]
							if has, _ := s.checkRouteGeometryForRiskZones(riskZones, route.Geometry, originLocation.Latitude, originLocation.Longitude, destinationLocation.Latitude, destinationLocation.Longitude); !has {
								break
							}
						}
					}
					removeBeforeDest(1)
				}

			}
			// ---------- fim do guard sweep ----------

			// Monta waypoints p/ URL do Google (endereços originais + via:lat,lng dos desvios)
			waypointsForURL := make([]string, 0, len(waypoints)+len(extraWaypointsForURL))
			if len(waypoints) > 2 {
				// Insere "via:" antes dos pontos intermediários de endereço
				waypointsForURL = append(waypointsForURL, extraWaypointsForURL...)
				waypointsForURL = append(waypointsForURL, waypoints[1:len(waypoints)-1]...)
			} else {
				waypointsForURL = append(waypointsForURL, extraWaypointsForURL...)
			}

			if len(waypointsForURL) == 0 {
				waypointsForURL = []string{
					fmt.Sprintf("%f,%f", originLocation.Latitude, originLocation.Longitude),
					fmt.Sprintf("%f,%f", destinationLocation.Latitude, destinationLocation.Longitude),
				}
			}
			totalRoute = s.createTotalSummaryWithURLWaypoints(route, originLocation, destinationLocation, waypoints, waypointsForURL, data)

			// Força o uso dos valores acumulados dos segmentos para duração e distância mais precisos
			distText, distVal := formatDistance(totalDistance)
			durText, durVal := formatDuration(totalDuration)
			totalRoute.TotalDistance = Distance{Text: distText, Value: distVal}
			totalRoute.TotalDuration = Duration{Text: durText, Value: durVal}

			// Recalcula custo de combustível com a distância correta
			avgConsumption := (data.ConsumptionCity + data.ConsumptionHwy) / 2
			totalKm := totalDistance / 1000
			totalRoute.TotalFuelCost = math.Round((data.Price / avgConsumption) * totalKm)

			// Caso precise expor os pontos do desvio, habilite:
			// totalRoute.DetourPoints = detourPtsTotal

		} else {

		}
	} else {

	}

	// ------------------------------
	// 4) Fallbacks (rota total sem desvio / agregados)
	// ------------------------------
	if totalRoute.TotalDistance.Value == 0 {
		// tenta rota total padrão para obter polyline
		baseCoords := strings.Join(allCoords, ";")
		urlBase := fmt.Sprintf(
			"http://34.207.174.233:5000/route/v1/driving/%s?alternatives=0&steps=true&overview=full&continue_straight=false",
			neturl.PathEscape(baseCoords),
		)

		if resp, err := client.Get(urlBase); err == nil {
			defer resp.Body.Close()
			var osrmResp OSRMResponse
			if err := json.NewDecoder(resp.Body).Decode(&osrmResp); err == nil && len(osrmResp.Routes) > 0 {
				route := osrmResp.Routes[0]
				totalRoute = s.createTotalSummary(route, originLocation, destinationLocation, waypoints, data)
			} else {

			}
		} else {

		}
	}

	if totalRoute.TotalDistance.Value == 0 {
		// último recurso: usa agregados
		distText, distVal := formatDistance(totalDistance)
		durText, durVal := formatDuration(totalDuration)

		avgConsumption := (data.ConsumptionCity + data.ConsumptionHwy) / 2
		totalKm := totalDistance / 1000
		totalFuelCost := math.Round((data.Price / avgConsumption) * totalKm)

		// Buscar pedágios para a rota base (sem desvios)
		baseCoords := strings.Join(allCoords, ";")
		urlBase := fmt.Sprintf(
			"http://34.207.174.233:5000/route/v1/driving/%s?alternatives=0&steps=true&overview=full&continue_straight=false",
			neturl.PathEscape(baseCoords),
		)

		var tolls []Toll
		var totalTollCost float64
		if resp, err := client.Get(urlBase); err == nil {
			defer resp.Body.Close()
			var osrmResp OSRMResponse
			if err := json.NewDecoder(resp.Body).Decode(&osrmResp); err == nil && len(osrmResp.Routes) > 0 {
				route := osrmResp.Routes[0]
				tolls, _ = s.findTollsOnRoute(ctx, route.Geometry, data.Type, float64(data.Axles))
				for _, toll := range tolls {
					totalTollCost += toll.CashCost
				}
			}
		}

		originAddress := waypoints[0]
		destAddress := waypoints[len(waypoints)-1]
		waypointStr := ""
		if len(waypoints) > 2 {
			waypointStr = "&waypoints=" + neturl.QueryEscape(strings.Join(waypoints[1:len(waypoints)-1], "|"))
		}

		googleURL := fmt.Sprintf(
			"https://www.google.com/maps/dir/?api=1&origin=%s&destination=%s%s&travelmode=driving",
			neturl.QueryEscape(originAddress),
			neturl.QueryEscape(destAddress),
			waypointStr,
		)
		currentTimeMillis := (time.Now().UnixNano() + int64(totalDuration*float64(time.Second))) / int64(time.Millisecond)
		wazeURL := fmt.Sprintf(
			"https://www.waze.com/pt-BR/live-map/directions/br?to=%s&from=%s&time=%d&reverse=yes",
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
				Address:  normalizeAddress(destAddress),
			},
			TotalDistance: Distance{Text: distText, Value: distVal},
			TotalDuration: Duration{Text: durText, Value: durVal},
			URL:           googleURL,
			URLWaze:       wazeURL,
			Tolls:         tolls,
			TotalTolls:    math.Round(totalTollCost*100) / 100,
			Polyline:      "",
			TotalFuelCost: totalFuelCost,
		}
	}

	return totalRoute
}

type osrmNearestResp struct {
	Waypoints []struct {
		Location [2]float64 `json:"location"`
	} `json:"waypoints"`
}

func (s *Service) snapToRoad(lat, lon float64) (float64, float64, bool) {
	urlStr := fmt.Sprintf("http://34.207.174.233:5000/nearest/v1/driving/%f,%f?number=1", lon, lat)
	client := http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(urlStr)
	if err != nil {
		return 0, 0, false
	}
	defer resp.Body.Close()

	var nr osrmNearestResp
	if err := json.NewDecoder(resp.Body).Decode(&nr); err != nil || len(nr.Waypoints) == 0 {
		return 0, 0, false
	}
	loc := nr.Waypoints[0].Location
	return loc[1], loc[0], true
}

// cria um resumo total da rota
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

	// fallback: se não vierem waypoints, usa coordenadas puras
	if len(waypoints) == 0 {
		originAddress := fmt.Sprintf("%f,%f", originLocation.Latitude, originLocation.Longitude)
		destAddress := fmt.Sprintf("%f,%f", destinationLocation.Latitude, destinationLocation.Longitude)
		waypoints = []string{originAddress, destAddress}
	}

	originAddress := waypoints[0]
	destAddress := waypoints[len(waypoints)-1]
	waypointStr := ""
	if len(waypoints) > 2 {
		waypointStr = "&waypoints=" + neturl.QueryEscape(strings.Join(waypoints[1:len(waypoints)-1], "|"))
	}

	googleURL := fmt.Sprintf(
		"https://www.google.com/maps/dir/?api=1&origin=%s&destination=%s%s&travelmode=driving",
		neturl.QueryEscape(originAddress),
		neturl.QueryEscape(destAddress),
		waypointStr,
	)

	currentTimeMillis := (time.Now().UnixNano() + int64(route.Duration*float64(time.Second))) / int64(time.Millisecond)
	wazeURL := fmt.Sprintf(
		"https://www.waze.com/pt-BR/live-map/directions/br?to=%s&from=%s&time=%d&reverse=yes",
		neturl.QueryEscape(destAddress),
		neturl.QueryEscape(originAddress),
		currentTimeMillis,
	)

	// Monta resumo com duração do OSRM
	summary := TotalSummary{
		LocationOrigin: AddressInfo{
			Location: originLocation,
			Address:  normalizeAddress(originAddress),
		},
		LocationDestination: AddressInfo{
			Location: destinationLocation,
			Address:  normalizeAddress(destAddress),
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

	// 🔹 Atualiza a duração com o Google Directions API (tempo real)
	if s.GoogleMapsAPIKey != "" {
		var waypointsForAPI []string
		if len(waypoints) > 2 {
			waypointsForAPI = waypoints[1 : len(waypoints)-1] // intermediários, se existirem
		}

		if gDurText, gDurVal, err := GetGoogleDurationWithTraffic(
			context.Background(),
			s.GoogleMapsAPIKey,
			originAddress,
			destAddress,
			waypointsForAPI,
		); err == nil {
			summary.TotalDuration = Duration{Text: gDurText, Value: float64(gDurVal)}
		} else {
			log.Printf("⚠️ Google Directions API falhou: %v", err)
		}
	}

	return summary
}

// cria um resumo total da rota com waypoints para URL separados
func (s *Service) createTotalSummaryWithURLWaypoints(route OSRMRoute, originLocation, destinationLocation Location, waypoints []string, waypointsForURL []string, data FrontInfoCEPRequest) TotalSummary {
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

	// fallback: se não vierem waypoints, usa coordenadas puras
	if len(waypoints) == 0 {
		originAddress := fmt.Sprintf("%f,%f", originLocation.Latitude, originLocation.Longitude)
		destAddress := fmt.Sprintf("%f,%f", destinationLocation.Latitude, destinationLocation.Longitude)
		waypoints = []string{originAddress, destAddress}
	}

	originAddress := waypoints[0]
	destAddress := waypoints[len(waypoints)-1]

	// Sempre usa os endereços originais para a URL do Google Maps
	waypointStr := ""
	if len(waypoints) > 2 {
		waypointStr = "&waypoints=" + neturl.QueryEscape(strings.Join(waypoints[1:len(waypoints)-1], "|"))
	}

	googleURL := fmt.Sprintf(
		"https://www.google.com/maps/dir/?api=1&origin=%s&destination=%s%s&travelmode=driving",
		neturl.QueryEscape(waypoints[0]),
		neturl.QueryEscape(waypoints[len(waypoints)-1]),
		waypointStr,
	)

	currentTimeMillis := (time.Now().UnixNano() + int64(route.Duration*float64(time.Second))) / int64(time.Millisecond)
	wazeURL := fmt.Sprintf(
		"https://www.waze.com/pt-BR/live-map/directions/br?to=%s&from=%s&time=%d&reverse=yes",
		neturl.QueryEscape(waypoints[len(waypoints)-1]),
		neturl.QueryEscape(waypoints[0]),
		currentTimeMillis,
	)

	// Monta resumo com duração do OSRM
	summary := TotalSummary{
		LocationOrigin: AddressInfo{
			Location: originLocation,
			Address:  normalizeAddress(originAddress),
		},
		LocationDestination: AddressInfo{
			Location: destinationLocation,
			Address:  normalizeAddress(destAddress),
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

	// Tenta obter duração mais precisa do Google Directions API
	if s.GoogleMapsAPIKey != "" && len(waypoints) >= 2 {
		waypointsForAPI := append([]string{}, waypoints...)

		if gDurText, gDurVal, err := GetGoogleDurationWithTraffic(
			context.Background(),
			s.GoogleMapsAPIKey,

			waypoints[0],
			waypoints[len(waypoints)-1],
			waypointsForAPI,
		); err == nil {
			summary.TotalDuration = Duration{Text: gDurText, Value: float64(gDurVal)}
		} else {
			log.Printf("⚠️ Google Directions API falhou: %v", err)
		}
	}

	return summary
}

func (s *Service) osrmNearestSnap(client http.Client, lat, lon float64) (float64, float64, bool) {
	u := fmt.Sprintf("http://34.207.174.233:5000/nearest/v1/driving/%.6f,%.6f?number=1", lon, lat)

	// Usar timeout reduzido para melhor performance
	fastClient := http.Client{Timeout: 5 * time.Second}
	resp, err := fastClient.Get(u)
	if err != nil {
		return lat, lon, false
	}
	defer resp.Body.Close()
	var r struct {
		Code      string `json:"code"`
		Waypoints []struct {
			Location []float64 `json:"location"`
		} `json:"waypoints"`
	}
	if json.NewDecoder(resp.Body).Decode(&r) != nil || len(r.Waypoints) == 0 {
		return lat, lon, false
	}
	sLon := r.Waypoints[0].Location[0]
	sLat := r.Waypoints[0].Location[1]
	return sLat, sLon, true
}

// NEWS - DEVOLVER PONTOS ANTES DA ROTA
func (s *Service) computeRiskOffsetsFromGeometry(geometry string, zone RiskZone, offsetMeters float64) (RiskOffsets, bool) {
	points, err := s.decodePolylineOSRM(geometry)
	if err != nil || len(points) < 2 {
		return RiskOffsets{}, false
	}

	// Distância acumulada ao longo da rota
	cum := make([]float64, len(points))
	for i := 1; i < len(points); i++ {
		cum[i] = cum[i-1] + s.haversineDistance(
			points[i-1].Latitude, points[i-1].Longitude,
			points[i].Latitude, points[i].Longitude,
		)
	}
	total := cum[len(cum)-1]

	// Descobrir entrada/saída com interseção segmento-círculo (em metros)
	latRef := zone.Lat // boa aproximação local
	var (
		foundEntry bool
		entrySeg   int
		entryT     float64
		exitSeg    int
		exitT      float64
	)

	inside := s.haversineDistance(points[0].Latitude, points[0].Longitude, zone.Lat, zone.Lng) <= float64(zone.Radius)

	for i := 0; i < len(points)-1; i++ {
		ts := s.segmentCircleIntersectionsMeters(points[i], points[i+1], zone, latRef)
		if len(ts) == 0 {
			// Atualiza estado "inside" usando extremo B
			inside = s.haversineDistance(points[i+1].Latitude, points[i+1].Longitude, zone.Lat, zone.Lng) <= float64(zone.Radius)
			continue
		}

		// Ordena e processa na ordem do percurso
		sort.Float64s(ts)

		if !foundEntry {
			if inside {
				// rota começa dentro → primeira interseção é SAÍDA
				exitSeg, exitT = i, ts[0]
				foundEntry = true // marca que já processamos um dos dois
				inside = false
				// ainda pode haver 2ª interseção no mesmo segmento (reentrada)
				if len(ts) > 1 {
					entrySeg, entryT = i, ts[1]
					inside = true
					break
				}
			} else {
				// fora → primeira interseção é ENTRADA
				entrySeg, entryT = i, ts[0]
				foundEntry = true
				inside = true
				// se houver 2ª interseção no mesmo segmento, ela é a SAÍDA
				if len(ts) > 1 {
					exitSeg, exitT = i, ts[1]
					inside = false
					break
				}
			}
		} else {
			// Já tivemos a primeira interseção; a próxima é a complementar
			if inside {
				// estávamos dentro → próxima é SAÍDA
				exitSeg, exitT = i, ts[0]
				inside = false
				break
			} else {
				// estávamos fora → próxima é ENTRADA (caso raro)
				entrySeg, entryT = i, ts[0]
				inside = true
				// se tiver uma segunda ainda neste segmento, já é a SAÍDA
				if len(ts) > 1 {
					exitSeg, exitT = i, ts[1]
					inside = false
					break
				}
			}
		}

		// Atualiza estado para o fim do segmento
		inside = s.haversineDistance(points[i+1].Latitude, points[i+1].Longitude, zone.Lat, zone.Lng) <= float64(zone.Radius)
	}

	// Se não achou par completo, aborta
	// (se só encontrou SAÍDA sem ENTRADA — rota começa dentro — criamos ENTRADA no ponto inicial)
	var entryLoc, exitLoc Location
	var entryCum, exitCum float64
	if entrySeg == 0 && entryT == 0 && !foundEntry {
		return RiskOffsets{}, false
	}

	if entrySeg == 0 && entryT == 0 && inside == false {
		// nada a fazer
	}

	// Se não temos ENTRADA mas temos SAÍDA → define ENTRADA no início da rota
	if entrySeg == 0 && entryT == 0 && exitT != 0 || (entrySeg == 0 && exitSeg != 0 && !pointIsValid(entrySeg, entryT)) {
		entrySeg, entryT = 0, 0
	}

	if !pointIsValid(exitSeg, exitT) && pointIsValid(entrySeg, entryT) {
		// achou apenas a entrada (ficou preso dentro até o final) → define saída no fim
		exitSeg, exitT = len(points)-2, 1
	}

	if !pointIsValid(entrySeg, entryT) && !pointIsValid(exitSeg, exitT) {
		return RiskOffsets{}, false
	}

	// Interpola pontos de entrada/saída e distâncias acumuladas
	entryLoc, entryCum = interpOnSegmentCum(points, cum, entrySeg, entryT, s)
	exitLoc, exitCum = interpOnSegmentCum(points, cum, exitSeg, exitT, s)

	// 5 km antes de entrar e 5 km depois de sair (limitando ao início/fim)
	beforeTarget := entryCum - offsetMeters
	if beforeTarget < 0 {
		beforeTarget = 0
	}
	afterTarget := exitCum + offsetMeters
	if afterTarget > total {
		afterTarget = total
	}

	beforeLoc, _ := pointAtCumDistance(points, cum, beforeTarget, s)
	afterLoc, _ := pointAtCumDistance(points, cum, afterTarget, s)

	return RiskOffsets{
		Zone:      zone,
		Entry:     entryLoc,
		Exit:      exitLoc,
		Before5km: beforeLoc,
		After5km:  afterLoc,
		EntryCum:  entryCum,
		ExitCum:   exitCum,
	}, true
}

func pointIsValid(seg int, t float64) bool {
	return seg >= 0 && t >= 0 && t <= 1
}

// Interseções t em [0,1] entre o segmento AB e o círculo (centro=zone, raio=zone.Radius), tudo em METROS
func (s *Service) segmentCircleIntersectionsMeters(a, b Location, zone RiskZone, latRef float64) []float64 {
	ax, ay := s.projectToMeters(latRef, a.Latitude, a.Longitude)
	bx, by := s.projectToMeters(latRef, b.Latitude, b.Longitude)
	cx, cy := s.projectToMeters(latRef, zone.Lat, zone.Lng)

	// Translada para o centro
	ax -= cx
	ay -= cy
	bx -= cx
	by -= cy

	dx := bx - ax
	dy := by - ay

	A := dx*dx + dy*dy
	if A == 0 {
		return nil
	}
	B := 2 * (ax*dx + ay*dy)
	R := float64(zone.Radius)
	C := ax*ax + ay*ay - R*R

	D := B*B - 4*A*C
	if D < 0 {
		return nil
	}
	sqrtD := math.Sqrt(D)
	t1 := (-B - sqrtD) / (2 * A)
	t2 := (-B + sqrtD) / (2 * A)

	var ts []float64
	if t1 >= 0 && t1 <= 1 {
		ts = append(ts, t1)
	}
	if t2 >= 0 && t2 <= 1 && (len(ts) == 0 || math.Abs(t2-ts[0]) > 1e-9) {
		ts = append(ts, t2)
	}
	sort.Float64s(ts)
	return ts
}

// Interpola ponto e distância acumulada em um segmento i no parâmetro t∈[0,1]
func interpOnSegmentCum(points []Location, cum []float64, seg int, t float64, s *Service) (Location, float64) {
	a := points[seg]
	b := points[seg+1]
	segLen := s.haversineDistance(a.Latitude, a.Longitude, b.Latitude, b.Longitude)
	lat := a.Latitude + (b.Latitude-a.Latitude)*t
	lon := a.Longitude + (b.Longitude-a.Longitude)*t
	return Location{Latitude: lat, Longitude: lon}, cum[seg] + segLen*t
}

// Retorna o ponto na rota no deslocamento acumulado "target" (m), com interpolação
func pointAtCumDistance(points []Location, cum []float64, target float64, s *Service) (Location, bool) {
	if target <= 0 {
		return points[0], true
	}
	lastIdx := len(points) - 1
	total := cum[lastIdx]
	if target >= total {
		return points[lastIdx], true
	}

	// busca linear (pode trocar por busca binária se quiser)
	for i := 1; i < len(points); i++ {
		if target <= cum[i] {
			a := points[i-1]
			b := points[i]
			segLen := s.haversineDistance(a.Latitude, a.Longitude, b.Latitude, b.Longitude)
			if segLen == 0 {
				return a, true
			}
			t := (target - cum[i-1]) / segLen
			lat := a.Latitude + (b.Latitude-a.Latitude)*t
			lon := a.Longitude + (b.Longitude-a.Longitude)*t
			return Location{Latitude: lat, Longitude: lon}, true
		}
	}
	return points[lastIdx], true
}

// Converte []Coordinate (strings) -> []Location (float64) e valida faixa
func (s *Service) normalizeUserWaypoints(in []Coordinate) ([]Location, error) {
	wps := make([]Location, 0, len(in))
	for _, c := range in {
		latStr := strings.TrimSpace(strings.ReplaceAll(c.Lat, ",", "."))
		lngStr := strings.TrimSpace(strings.ReplaceAll(c.Lng, ",", "."))
		lat, err1 := strconv.ParseFloat(latStr, 64)
		lng, err2 := strconv.ParseFloat(lngStr, 64)
		if err1 != nil || err2 != nil {
			continue // ignora inválidos
		}
		if lat < -90 || lat > 90 || lng < -180 || lng > 180 {
			continue
		}
		wps = append(wps, Location{Latitude: lat, Longitude: lng})
	}
	if len(wps) == 0 {
		return nil, fmt.Errorf("nenhum waypoint válido recebido")
	}
	return wps, nil
}

// NEWS - ADICIONAR WAYPOINTS AUTOMATICAMENTE
func normalizeAngle(a float64) float64 {
	for a <= -math.Pi {
		a += 2 * math.Pi
	}
	for a > math.Pi {
		a -= 2 * math.Pi
	}
	return a
}

// Gera n pontos em uma ARCADA entre os ângulos de entrada/saída do círculo,
// usando raio + extraBuffer (em metros). Usa projeção local (mesma de projectToMeters).
func (s *Service) buildArcWaypoints(entry, exit Location, zone RiskZone, n int, extraBuffer float64) []Location {
	latRef := zone.Lat
	// centro em metros
	cx, cy := s.projectToMeters(latRef, zone.Lat, zone.Lng)
	// pontos relativos ao centro
	ex, ey := s.projectToMeters(latRef, entry.Latitude, entry.Longitude)
	ex -= cx
	ey -= cy
	sx, sy := s.projectToMeters(latRef, exit.Latitude, exit.Longitude)
	sx -= cx
	sy -= cy

	theta1 := math.Atan2(ey, ex)
	theta2 := math.Atan2(sy, sx)
	delta := normalizeAngle(theta2 - theta1)

	// usa a menor arcada (mais curta)
	if math.Abs(delta) > math.Pi {
		if delta > 0 {
			delta -= 2 * math.Pi
		} else {
			delta += 2 * math.Pi
		}
	}

	R := float64(zone.Radius) + extraBuffer
	out := make([]Location, 0, n)
	for i := 1; i <= n; i++ {
		t := float64(i) / float64(n+1)
		ang := theta1 + delta*t
		ax := math.Cos(ang) * R
		ay := math.Sin(ang) * R
		lat, lon := s.unprojectFromMeters(latRef, ax+cx, ay+cy)
		out = append(out, Location{Latitude: lat, Longitude: lon})
	}
	return out
}

func (s *Service) buildArcWaypointsDir(entry, exit Location, zone RiskZone, n int, extraBuffer float64, useLongArc bool) []Location {
	latRef := zone.Lat
	cx, cy := s.projectToMeters(latRef, zone.Lat, zone.Lng)

	ex, ey := s.projectToMeters(latRef, entry.Latitude, entry.Longitude)
	ex -= cx
	ey -= cy
	sx, sy := s.projectToMeters(latRef, exit.Latitude, exit.Longitude)
	sx -= cx
	sy -= cy

	theta1 := math.Atan2(ey, ex)
	theta2 := math.Atan2(sy, sx)
	delta := normalizeAngle(theta2 - theta1) // (-π, π]

	// MENOR arco por padrão
	if !useLongArc {
		// já é o menor arco, pois delta ∈ (-π, π]
	} else {
		// lado oposto: mesmo sentido, mas 2π - |delta|
		if delta > 0 {
			delta = delta - 2*math.Pi
		} else {
			delta = delta + 2*math.Pi
		}
	}

	R := float64(zone.Radius) + extraBuffer
	out := make([]Location, 0, n)
	for i := 1; i <= n; i++ {
		t := float64(i) / float64(n+1)
		ang := theta1 + delta*t
		ax := math.Cos(ang) * R
		ay := math.Sin(ang) * R
		lat, lon := s.unprojectFromMeters(latRef, ax+cx, ay+cy)
		out = append(out, Location{Latitude: lat, Longitude: lon})
	}
	return out
}

// monta via-points: [entryOut, arc(midpoints...), exitOut]
func (s *Service) assembleLateralDetour(entry, exit Location, zone RiskZone, arcPoints int, arcExtraBuffer float64, entryExitPush float64, useLongArc bool) []Location {
	// âncoras  (ligeiro empurrão para fora)
	entryOut := s.pushAwayFromCenter(entry, zone, entryExitPush)
	exitOut := s.pushAwayFromCenter(exit, zone, entryExitPush)

	// arcada (mesmo raio+buffer que você já usa)
	var arc []Location
	if useLongArc {
		arc = s.buildArcWaypointsDir(entry, exit, zone, arcPoints, arcExtraBuffer, true)
	} else {
		arc = s.buildArcWaypointsDir(entry, exit, zone, arcPoints, arcExtraBuffer, false)
	}

	wps := make([]Location, 0, 2+len(arc))
	wps = append(wps, entryOut)
	wps = append(wps, arc...)
	wps = append(wps, exitOut)
	return wps
}

func (s *Service) pushToOutside(p Location, zone RiskZone, extra float64) Location {
	d := s.haversineDistance(p.Latitude, p.Longitude, zone.Lat, zone.Lng)
	target := float64(zone.Radius) + extra
	delta := target - d
	if delta < 50 {
		delta = 50
	}
	return s.pushAwayFromCenter(p, zone, delta)
}

func (s *Service) awayNormalForSegment(before, after Location, zone RiskZone) (latRef, nx, ny float64) {
	latRef = (before.Latitude + after.Latitude) / 2
	ax, ay := s.projectToMeters(latRef, before.Latitude, before.Longitude)
	bx, by := s.projectToMeters(latRef, after.Latitude, after.Longitude)
	cx, cy := s.projectToMeters(latRef, zone.Lat, zone.Lng)

	vx, vy := bx-ax, by-ay           // direção do segmento
	lx, ly := -vy, vx                // normal "esquerda" do segmento
	cross := vx*(cy-ay) - vy*(cx-ax) // >0: centro está à esquerda do segmento

	// se o centro está à esquerda, usamos a normal da "direita" (oposta à esquerda)
	if cross > 0 {
		lx, ly = -lx, -ly
	}
	mag := math.Hypot(lx, ly)
	if mag == 0 {
		return latRef, 0, 0
	}
	return latRef, lx / mag, ly / mag
}

// desloca um ponto p pela normal (nx,ny) em 'dist' metros (mantendo o mesmo lado)
func (s *Service) offsetByNormal(p Location, latRef, nx, ny, dist float64) Location {
	px, py := s.projectToMeters(latRef, p.Latitude, p.Longitude)
	px += nx * dist
	py += ny * dist
	lat, lon := s.unprojectFromMeters(latRef, px, py)
	return Location{Latitude: lat, Longitude: lon}
}

// news
// Snap que GARANTE ficar fora da zona. Se o snap cair dentro, empurra p/ fora e tenta de novo.
func (s *Service) snapOutsideMany(client http.Client, pts []Location, zone RiskZone) []Location {
	out := make([]Location, len(pts))
	copy(out, pts)
	for i := range out {
		p := out[i]
		for step := 0; step < 6; step++ {
			lat, lon, ok := s.osrmNearestSnap(client, p.Latitude, p.Longitude)
			if ok && s.haversineDistance(lat, lon, zone.Lat, zone.Lng) > float64(zone.Radius)+5 {
				out[i] = Location{Latitude: lat, Longitude: lon}
				break
			}
			p = s.pushAwayFromCenter(p, zone, 120.0)
		}
	}
	return out
}

// Detecta TODAS as zonas cruzadas pela polyline e devolve ordenadas pela distância acumulada de entrada.
func (s *Service) detectAllCrossingsFromGeometry(geometry string, zones []RiskZone, offsetMeters float64) []RiskOffsets {
	offs := make([]RiskOffsets, 0, len(zones))
	for _, z := range zones {
		if !z.Status {
			continue
		}
		if off, ok := s.computeRiskOffsetsFromGeometry(geometry, z, offsetMeters); ok {
			offs = append(offs, off)
		}
	}
	sort.Slice(offs, func(i, j int) bool { return offs[i].EntryCum < offs[j].EntryCum })
	return offs
}

// Checa rota OSRM real e retorna TODAS as zonas cruzadas (ordenadas). Bool indica se há pelo menos uma.
func (s *Service) CheckRouteForAllRiskZones(riskZones []RiskZone, originLat, originLon, destLat, destLon float64) ([]RiskOffsets, bool) {
	client := http.Client{Timeout: 15 * time.Second}
	coords := fmt.Sprintf("%f,%f;%f,%f", originLon, originLat, destLon, destLat)
	url := fmt.Sprintf("http://34.207.174.233:5000/route/v1/driving/%s?overview=full&steps=true", neturl.PathEscape(coords))

	resp, err := client.Get(url)
	if err != nil {
		// fallback: mantém seu comportamento antigo (linha reta), mas sem offsets detalhados
		if s.checkRouteForRiskZonesFallback(riskZones, originLat, originLon, destLat, destLon) {
			// sem offsets nesse fallback
			return nil, true
		}
		return nil, false
	}
	defer resp.Body.Close()

	var osrmResp OSRMResponse
	if json.NewDecoder(resp.Body).Decode(&osrmResp) != nil || len(osrmResp.Routes) == 0 {
		if s.checkRouteForRiskZonesFallback(riskZones, originLat, originLon, destLat, destLon) {
			return nil, true
		}
		return nil, false
	}

	route := osrmResp.Routes[0]
	offs := s.detectAllCrossingsFromGeometry(route.Geometry, riskZones, 1000)
	return offs, len(offs) > 0
}

// computeArrivalGuardPoint retorna um ponto de "guarda" fora da zona,
// no raio+buffer, na direção do destino. Força a aproximação por fora.
func (s *Service) computeArrivalGuardPoint(zone RiskZone, dest Location, buffer float64) Location {
	latRef := zone.Lat
	cx, cy := s.projectToMeters(latRef, zone.Lat, zone.Lng)
	dx, dy := s.projectToMeters(latRef, dest.Latitude, dest.Longitude)

	vx, vy := dx-cx, dy-cy
	mag := math.Hypot(vx, vy)
	if mag == 0 {
		// destino exatamente no centro (caso patológico) → empurra para leste
		vx, vy, mag = 1, 0, 1
	}
	// coloca o guard na borda externa (raio + buffer) olhando para o destino
	scale := (float64(zone.Radius) + buffer) / mag
	gx := cx + vx*scale
	gy := cy + vy*scale

	lat, lon := s.unprojectFromMeters(latRef, gx, gy)
	return Location{Latitude: lat, Longitude: lon}
}

// Fallback: calcula distância/tempo por Haversine e monta um resumo "ok".
func (s *Service) createDirectEstimateSummary(originLat, originLon, destLat, destLon float64, originGeocode, destGeocode GeocodeResult, data FrontInfoCEPRequest) RouteSummary {

	// distância/tempo estimados
	distance := s.haversineDistance(originLat, originLon, destLat, destLon)
	duration := distance / 16.67 // ~60km/h

	distText, distVal := formatDistance(distance)
	durText, durVal := formatDuration(duration)

	avgConsumption := (data.ConsumptionCity + data.ConsumptionHwy) / 2
	totalKm := distance / 1000
	totalFuelCost := math.Round((data.Price / avgConsumption) * totalKm)

	// Tentar calcular pedágios mesmo no fallback usando uma rota OSRM simples
	var tolls []Toll
	var totalTollCost float64
	coords := fmt.Sprintf("%f,%f;%f,%f", originLon, originLat, destLon, destLat)
	baseURL := "http://34.207.174.233:5000/route/v1/driving/" + neturl.PathEscape(coords) +
		"?alternatives=0&steps=true&overview=full&continue_straight=false"

	client := http.Client{Timeout: 5 * time.Second}
	if resp, err := client.Get(baseURL); err == nil && resp != nil {
		defer resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			var osrmResp OSRMResponse
			if json.NewDecoder(resp.Body).Decode(&osrmResp) == nil && len(osrmResp.Routes) > 0 {
				route := osrmResp.Routes[0]
				tolls, _ = s.findTollsOnRoute(context.Background(), route.Geometry, data.Type, float64(data.Axles))
				for _, toll := range tolls {
					totalTollCost += toll.CashCost
				}
			}
		}
	}

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
		HasTolls:      len(tolls) > 0,
		Distance:      Distance{Text: distText, Value: distVal},
		Duration:      Duration{Text: durText, Value: durVal},
		URL:           googleURL,
		URLWaze:       wazeURL,
		TotalFuelCost: totalFuelCost,
		Tolls:         tolls,
		TotalTolls:    math.Round(totalTollCost*100) / 100,
		Polyline:      "", // sem polyline no fallback
	}
}

func (s *Service) calculateDirectRoute(ctx context.Context, client http.Client, originLat, originLon, destLat, destLon float64, originGeocode, destGeocode GeocodeResult, data FrontInfoCEPRequest) []RouteSummary {

	coords := fmt.Sprintf("%f,%f;%f,%f", originLon, originLat, destLon, destLat)
	baseURL := "http://34.207.174.233:5000/route/v1/driving/" + neturl.PathEscape(coords) +
		"?alternatives=1&steps=true&overview=full&continue_straight=false"

	resp, err := client.Get(baseURL)
	if err == nil && resp != nil {
		defer resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			var osrmResp OSRMResponse
			if json.NewDecoder(resp.Body).Decode(&osrmResp) == nil && len(osrmResp.Routes) > 0 {
				route := osrmResp.Routes[0]
				tolls, _ := s.findTollsOnRoute(ctx, route.Geometry, data.Type, float64(data.Axles))
				return []RouteSummary{
					s.createRouteSummary(route, "rota_direta_com_aviso", originGeocode, destGeocode, data, tolls),
				}
			}
		}
	} else {
		log.Printf("⚠️  Erro HTTP ao consultar OSRM (rota direta): %v. Usando fallback.", err)
	}

	// Fallback local (nunca devolve erro ao front)
	return []RouteSummary{
		s.createDirectEstimateSummary(originLat, originLon, destLat, destLon, originGeocode, destGeocode, data),
	}
}

// GetCoordinatesFromAddress obtém latitude e longitude a partir de um endereço (rua e número)
func (s *Service) GetCoordinatesFromAddress(ctx context.Context, street, number, city, state, cep string) (AddressCoordinatesResponse, error) {
	// Validação dos parâmetros de entrada
	if strings.TrimSpace(street) == "" {
		return AddressCoordinatesResponse{}, fmt.Errorf("rua não pode estar vazia")
	}
	if strings.TrimSpace(number) == "" {
		return AddressCoordinatesResponse{}, fmt.Errorf("número não pode estar vazio")
	}

	// Monta o endereço completo
	// Monta o endereço base (rua + número)
	addressParts := []string{
		strings.TrimSpace(street),
		strings.TrimSpace(number),
	}

	// Adiciona refinamentos se existirem
	if cep != "" {
		addressParts = append(addressParts, cep)
	}
	if city != "" {
		addressParts = append(addressParts, city)
	}
	if state != "" {
		addressParts = append(addressParts, state)
	}

	// Sempre termina com Brasil
	addressParts = append(addressParts, "Brasil")
	address := strings.Join(addressParts, ", ")

	// Chave do cache baseada no endereço
	cacheKey := fmt.Sprintf("geocode_address:%s", address)

	// Tenta buscar no cache primeiro [[memory:7626964]]
	cached, err := cache.Rdb.Get(cache.Ctx, cacheKey).Result()
	if err == nil {
		var location AddressCoordinatesResponse
		if json.Unmarshal([]byte(cached), &location) == nil {
			return location, nil
		}
	} else if !errors.Is(err, redis.Nil) {
		log.Printf("Erro ao buscar no cache: %v", err)
	}

	// Cria cliente do Google Maps
	client, err := maps.NewClient(maps.WithAPIKey(s.GoogleMapsAPIKey))
	if err != nil {
		return AddressCoordinatesResponse{}, fmt.Errorf("erro ao criar cliente Google Maps: %v", err)
	}

	// Primeiro tenta usar Place Autocomplete para melhorar a precisão
	//autoCompleteReq := &maps.PlaceAutocompleteRequest{
	//	Input:    address,
	//	Location: &maps.LatLng{Lat: -14.2350, Lng: -51.9253}, // Centro do Brasil
	//	Radius:   1000000,
	//	Language: "pt-BR",
	//	Types:    "geocode",
	//}
	//fmt.Println(address)
	//
	//autoCompleteResp, autoCompleteErr := client.PlaceAutocomplete(ctx, autoCompleteReq)
	//if autoCompleteErr == nil && len(autoCompleteResp.Predictions) > 0 {
	//	address = autoCompleteResp.Predictions[0].Description
	//} else if autoCompleteErr != nil {
	//	log.Printf("Erro no Place Autocomplete: %v", autoCompleteErr)
	//}

	// Faz a requisição de geocoding
	req := &maps.GeocodingRequest{
		Address: address,
		Region:  "br",
		Components: map[maps.Component]string{
			"country":             "BR",
			"postal_code":         cep,   // se informado
			"administrative_area": state, // UF
			"locality":            city,  // município
		},
	}

	results, err := client.Geocode(ctx, req)
	if err != nil {
		return AddressCoordinatesResponse{}, fmt.Errorf("erro ao consultar Google Maps API: %v", err)
	}

	if len(results) == 0 {
		return AddressCoordinatesResponse{}, fmt.Errorf("endereço não encontrado: %s. Verifique se a rua e número estão corretos", address)
	}

	// Extrai as coordenadas do primeiro resultado
	location := AddressCoordinatesResponse{
		Latitude:  results[0].Geometry.Location.Lat,
		Longitude: results[0].Geometry.Location.Lng,
		Address:   results[0].FormattedAddress,
	}

	// Salva no cache por 30 dias [[memory:7626964]]
	data, err := json.Marshal(location)
	if err == nil {
		if err := cache.Rdb.Set(cache.Ctx, cacheKey, data, 30*24*time.Hour).Err(); err != nil {
			log.Printf("Erro ao salvar no cache: %v", err)
		}
	}

	return location, nil
}
