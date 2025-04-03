package new_routes

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	db "geolocation/db/sqlc"
	"geolocation/internal/routes"
	cache "geolocation/pkg"
	"geolocation/validation"
	"github.com/go-redis/redis/v8"
	"googlemaps.github.io/maps"
	"log"
	"math"
	"net/http"
	"net/url"
	neturl "net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

type InterfaceService interface {
	CalculateRoutes(ctx context.Context, frontInfo FrontInfo, idPublicToken int64, idSimp int64) (FinalOutput, error)
	CalculateRoutesWithCoordinate(ctx context.Context, frontInfo FrontInfoCoordinate, idPublicToken int64, idSimp int64) (FinalOutput, error)
	GetFavoriteRouteService(ctx context.Context, id int64) ([]FavoriteRouteResponse, error)
	RemoveFavoriteRouteService(ctx context.Context, id, idUser int64) error
	GetSimpleRoute(data SimpleRouteRequest) (SimpleRouteResponse, error)
}

type Service struct {
	InterfaceService routes.InterfaceRepository
	GoogleMapsAPIKey string
}

func NewRoutesNewService(interfaceService routes.InterfaceRepository, googleMapsAPIKey string) *Service {
	return &Service{
		InterfaceService: interfaceService,
		GoogleMapsAPIKey: googleMapsAPIKey,
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
		if json.Unmarshal([]byte(cached), &cachedOutput) == nil {
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
	baseOSRMURL := "http://34.207.174.233:5000/route/v1/driving/" + url.PathEscape(coords)
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
					URL:     googleURL,
					URLWaze: wazeURL,
				},
				Costs: Costs{
					TagAndCash:      totalTollCost,
					FuelInTheCity:   fuelCostCity,
					FuelInTheHwy:    fuelCostHwy,
					Tag:             (totalTollCost - (totalTollCost * 0.05)) * float64(frontInfo.Axles),
					Cash:            totalTollCost * float64(frontInfo.Axles),
					PrepaidCard:     totalTollCost * float64(frontInfo.Axles),
					MaximumTollCost: totalTollCost * float64(frontInfo.Axles),
					MinimumTollCost: totalTollCost * float64(frontInfo.Axles),
					Axles:           int(frontInfo.Axles),
				},
				Tolls:        routeTolls,
				Balances:     routeBalancas,
				GasStations:  routeGasStations,
				Instructions: finalInstructions,
				FreightLoad:  freight,
				Polyline:     route.Geometry,
			})
		}

		sort.Slice(output, func(i, j int) bool {
			return len(output[i].Tolls) < len(output[j].Tolls)
		})
		return output
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
	baseOSRMURL := "http://34.207.174.233:5000/route/v1/driving/" + url.PathEscape(coords)
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
	fmt.Println(origin)
	fmt.Println(origin.PlaceID)
	fmt.Println(destination.PlaceID)
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
					URL:     googleURL,
					URLWaze: wazeURL,
				},
				Costs: Costs{
					TagAndCash:      totalTollCost,
					FuelInTheCity:   fuelCostCity,
					FuelInTheHwy:    fuelCostHwy,
					Tag:             (totalTollCost - (totalTollCost * 0.05)) * float64(frontInfo.Axles),
					Cash:            totalTollCost * float64(frontInfo.Axles),
					PrepaidCard:     totalTollCost * float64(frontInfo.Axles),
					MaximumTollCost: totalTollCost * float64(frontInfo.Axles),
					MinimumTollCost: totalTollCost * float64(frontInfo.Axles),
					Axles:           int(frontInfo.Axles),
				},
				Tolls:        routeTolls,
				Balances:     routeBalancas,
				GasStations:  routeGasStations,
				Instructions: finalInstructions,
				FreightLoad:  freight,
				Polyline:     route.Geometry,
			})
		}

		sort.Slice(output, func(i, j int) bool {
			return len(output[i].Tolls) < len(output[j].Tolls)
		})
		return output
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

	tolerance := 50.0
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
	address = StateToCapital(strings.ToLower(address))
	cacheKey := fmt.Sprintf("geocode:%s", address)
	cached, err := cache.Rdb.Get(cache.Ctx, cacheKey).Result()
	if err == nil {
		var result GeocodeResult
		if json.Unmarshal([]byte(cached), &result) == nil {
			return result, nil
		}
	} else if !errors.Is(err, redis.Nil) {
		fmt.Printf("Erro ao recuperar cache do Redis (geocode): %v\n", err)
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

	data, err := json.Marshal(result)
	if err == nil {
		if err := cache.Rdb.Set(cache.Ctx, cacheKey, data, 30*24*time.Hour).Err(); err != nil {
			fmt.Printf("Erro ao salvar cache do Redis (geocode): %v\n", err)
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
	baseOSRMURL := "http://34.207.174.233:5000/route/v1/driving/" + url.PathEscape(coords)
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
