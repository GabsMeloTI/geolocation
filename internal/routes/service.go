package routes

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	db "geolocation/db/sqlc"
	cache "geolocation/pkg"
	"geolocation/validation"
	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	"golang.org/x/net/html"
	"googlemaps.github.io/maps"
	"math"
	neturl "net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type InterfaceService interface {
	CheckRouteTolls(ctx context.Context, frontInfo FrontInfo, id int64) (Response, error)
}

type Service struct {
	InterfaceService InterfaceRepository
	GoogleMapsAPIKey string
}

func NewRoutesService(interfaceService InterfaceRepository, googleMapsAPIKey string) *Service {
	return &Service{
		InterfaceService: interfaceService,
		GoogleMapsAPIKey: googleMapsAPIKey,
	}
}

func (s *Service) CheckRouteTolls(ctx context.Context, frontInfo FrontInfo, id int64) (Response, error) {
	if frontInfo.PublicOrPrivate == "public" {
		if err := s.updateNumberOfRequest(ctx, id); err != nil {
			return Response{}, err
		}
	}

	cacheKey := fmt.Sprintf("route:%s:%s:%s",
		strings.ToLower(frontInfo.Origin),
		strings.ToLower(frontInfo.Destination),
		strings.ToLower(strings.Join(frontInfo.Waypoints, ",")),
	)

	cached, err := cache.Rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		var cachedResponse Response
		if json.Unmarshal([]byte(cached), &cachedResponse) == nil {
			return RecalculateCosts(cachedResponse, frontInfo), nil
		}
	} else if !errors.Is(err, redis.Nil) {
		return Response{}, err
	}

	origin, err := s.getGeocodeAddress(ctx, frontInfo.Origin)
	if err != nil {
		return Response{}, err
	}
	destination, err := s.getGeocodeAddress(ctx, frontInfo.Destination)
	if err != nil {
		return Response{}, err
	}

	uniqueWaypoints := make(map[string]bool)
	var waypoints []string
	for _, waypoint := range frontInfo.Waypoints {
		normalized := strings.ToLower(strings.TrimSpace(waypoint))
		if !uniqueWaypoints[normalized] {
			uniqueWaypoints[normalized] = true
			result, err := s.getGeocodeAddress(ctx, waypoint)
			if err != nil {
				return Response{}, err
			}
			waypoints = append(waypoints, result.FormattedAddress)
		}
	}

	routeRequest := &maps.DirectionsRequest{
		Origin:       origin.FormattedAddress,
		Destination:  destination.FormattedAddress,
		Mode:         maps.TravelModeDriving,
		Waypoints:    waypoints,
		Alternatives: true,
		Region:       "br",
		Language:     "pt-BR",
	}

	requestJSON, _ := json.Marshal(routeRequest)
	savedRoute, err := s.InterfaceService.GetSavedRoutes(ctx, db.GetSavedRoutesParams{
		Origin:      origin.FormattedAddress,
		Destination: destination.FormattedAddress,
		Waypoints: sql.NullString{
			String: strings.ToLower(strings.Join(frontInfo.Waypoints, ",")),
			Valid:  true,
		},
		Request: requestJSON,
	})
	if err == nil && savedRoute.ExpiredAt.After(time.Now()) {
		var dbResponse Response
		if json.Unmarshal(savedRoute.Response, &dbResponse) == nil {
			cache.Rdb.Set(ctx, cacheKey, savedRoute.Response, 30*24*time.Hour)
			return RecalculateCosts(dbResponse, frontInfo), nil
		}
	}

	client, err := maps.NewClient(maps.WithAPIKey(s.GoogleMapsAPIKey))
	if err != nil {
		return Response{}, err
	}

	routes, _, err := client.Directions(ctx, routeRequest)
	if err != nil {
		return Response{}, err
	}
	if len(routes) == 0 {
		return Response{}, echo.ErrNotFound
	}

	var allRoutes []Route
	var summaryRoute SummaryRoute
	var maxTollCost float64
	minTollCost := math.MaxFloat64

	for _, route := range routes {
		foundTolls, _ := s.findTollsInRoute(ctx, []maps.Route{route}, origin.FormattedAddress, frontInfo.Type, float64(frontInfo.Axles))
		foundBalancas, _ := s.findBalancaInRoute(ctx, []maps.Route{route})
		foundGasStations, _ := s.findGasStationsAlongAllRoutes(ctx, client, []maps.Route{route})
		foundTolls = sortByProximity(origin.Location, foundTolls, func(toll Toll) Location {
			return Location{Latitude: toll.Latitude, Longitude: toll.Longitude}
		})
		foundBalancas = sortByProximity(origin.Location, foundBalancas, func(balanca Balanca) Location {
			return Location{Latitude: balanca.Lat, Longitude: balanca.Lng}
		})
		foundGasStations = sortByProximity(origin.Location, foundGasStations, func(gas GasStation) Location {
			return Location{Latitude: gas.Location.Latitude, Longitude: gas.Location.Longitude}
		})

		var totalDistance int
		var totalDuration time.Duration
		var totalTollCost float64
		var locations []PrincipalRoute
		var instructions []string

		for _, leg := range route.Legs {
			totalDistance += leg.Distance.Meters
			totalDuration += leg.Duration
			locations = append(locations, PrincipalRoute{
				Location: Location{
					Latitude:  leg.StartLocation.Lat,
					Longitude: leg.StartLocation.Lng,
				},
				Address: leg.StartAddress,
			})
			for _, step := range leg.Steps {
				instr := validation.RemoveHTMLTags(step.HTMLInstructions)
				instr = html.UnescapeString(instr)
				instructions = append(instructions, instr)
			}
		}
		lastLeg := route.Legs[len(route.Legs)-1]
		locations = append(locations, PrincipalRoute{
			Location: Location{
				Latitude:  lastLeg.EndLocation.Lat,
				Longitude: lastLeg.EndLocation.Lng,
			},
			Address: lastLeg.EndAddress,
		})
		if len(locations) <= 2 {
			locations = []PrincipalRoute{}
		}

		for _, toll := range foundTolls {
			totalTollCost += toll.CashCost
		}
		if totalTollCost > maxTollCost {
			maxTollCost = totalTollCost
		}
		if totalTollCost < minTollCost {
			minTollCost = totalTollCost
		}
		fuelCostCity := math.Round((frontInfo.Price / frontInfo.ConsumptionCity) * (float64(totalDistance) / 1000))
		fuelCostHwy := math.Round((frontInfo.Price / frontInfo.ConsumptionHwy) * (float64(totalDistance) / 1000))

		googleURL := fmt.Sprintf("https://www.google.com/maps/dir/?api=1&origin=%s&destination=%s",
			neturl.QueryEscape(origin.FormattedAddress),
			neturl.QueryEscape(destination.FormattedAddress))
		if len(frontInfo.Waypoints) > 0 {
			googleURL += "&waypoints=" + neturl.QueryEscape(strings.Join(frontInfo.Waypoints, "|"))
		}
		currentTimeMillis := (time.Now().UnixNano() + lastLeg.Duration.Nanoseconds()) / int64(time.Millisecond)
		wazeURL := fmt.Sprintf(
			"https://www.waze.com/pt-BR/live-map/directions/br?to=place.%s&from=place.%s&time=%d&reverse=yes",
			destination.PlaceID,
			origin.PlaceID,
			currentTimeMillis,
		)
		if len(frontInfo.Waypoints) > 0 {
			wazeURL += "&via=place." + neturl.QueryEscape(frontInfo.Waypoints[0])
		}

		var finalInstruction []Instructions
		for _, instruction := range instructions {
			instructionLower := strings.ToLower(instruction)
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
			case strings.Contains(instructionLower, "continue"), strings.Contains(instructionLower, "siga"), strings.Contains(instructionLower, "pegue"):
				valueImg = "https://plates-routes.s3.us-east-1.amazonaws.com/reto.png"
			case strings.Contains(instructionLower, "rotatória"), strings.Contains(instructionLower, "rotatoria"), strings.Contains(instructionLower, "retorno"):
				valueImg = "https://plates-routes.s3.us-east-1.amazonaws.com/rotatoria.png"
			case strings.Contains(instructionLower, "voltar"), strings.Contains(instructionLower, "volta"):
				valueImg = "https://plates-routes.s3.us-east-1.amazonaws.com/voltar.png"
			}
			finalInstruction = append(finalInstruction, Instructions{
				Text: instruction,
				Img:  valueImg,
			})
		}

		kms := totalDistance / 1000
		resultFreight, err := s.getAllFreight(ctx, frontInfo.Axles, float64(kms))
		if err != nil {
			return Response{}, err
		}

		allRoutes = append(allRoutes, Route{
			Summary: Summary{
				HasTolls:  len(foundTolls) > 0,
				RouteType: "",
				Distance: Distance{
					Text:  fmt.Sprintf("%d km", totalDistance/1000),
					Value: totalDistance,
				},
				Duration: Duration{
					Text:  totalDuration.String(),
					Value: totalDuration.Seconds(),
				},
				URL:     googleURL,
				URLWaze: wazeURL,
			},
			Costs: Costs{
				TagAndCash:      totalTollCost,
				FuelInTheCity:   fuelCostCity,
				FuelInTheHwy:    fuelCostHwy,
				Tag:             totalTollCost - (totalTollCost * 0.05),
				Cash:            totalTollCost,
				PrepaidCard:     totalTollCost,
				MaximumTollCost: maxTollCost,
				MinimumTollCost: minTollCost,
				Axles:           frontInfo.Axles,
			},
			Tolls:        foundTolls,
			Balanca:      foundBalancas,
			GasStations:  foundGasStations,
			Polyline:     route.OverviewPolyline.Points,
			Instructions: finalInstruction,
			FreightLoad:  resultFreight,
		})

		summaryRoute = SummaryRoute{
			RouteOrigin: PrincipalRoute{
				Location: Location{
					Latitude:  routes[0].Legs[0].StartLocation.Lat,
					Longitude: routes[0].Legs[0].StartLocation.Lng,
				},
				Address: routes[0].Legs[0].StartAddress,
			},
			RouteDestination: PrincipalRoute{
				Location: Location{
					Latitude:  routes[0].Legs[len(routes[0].Legs)-1].EndLocation.Lat,
					Longitude: routes[0].Legs[len(routes[0].Legs)-1].EndLocation.Lng,
				},
				Address: routes[0].Legs[len(routes[0].Legs)-1].EndAddress,
			},
			AllWayPoints: locations,
			FuelPrice: FuelPrice{
				Value:    frontInfo.Price,
				Currency: "BRL",
				Units:    "km",
				FuelUnit: "liter",
			},
			FuelEfficiency: FuelEfficiency{
				City:     frontInfo.ConsumptionCity,
				Hwy:      frontInfo.ConsumptionHwy,
				Units:    "km",
				FuelUnit: "liter",
			},
		}
	}

	var fastestRoute, cheapestRoute, efficientRoute Route
	if len(allRoutes) >= 3 {
		sortedByTime := make([]Route, len(allRoutes))
		copy(sortedByTime, allRoutes)
		sort.Slice(sortedByTime, func(i, j int) bool {
			return sortedByTime[i].Summary.Duration.Value < sortedByTime[j].Summary.Duration.Value
		})
		fastestRoute = sortedByTime[0]

		sortedByCost := make([]Route, len(allRoutes))
		copy(sortedByCost, allRoutes)
		sort.Slice(sortedByCost, func(i, j int) bool {
			costI := sortedByCost[i].Costs.TagAndCash + sortedByCost[i].Costs.FuelInTheCity
			costJ := sortedByCost[j].Costs.TagAndCash + sortedByCost[j].Costs.FuelInTheCity
			return costI < costJ
		})
		cheapestRoute = sortedByCost[0]
		if cheapestRoute.Polyline == fastestRoute.Polyline {
			for _, candidate := range sortedByCost[1:] {
				if candidate.Polyline != fastestRoute.Polyline {
					cheapestRoute = candidate
					break
				}
			}
		}

		var candidateRoutes []Route
		for _, r := range allRoutes {
			if r.Polyline != fastestRoute.Polyline && r.Polyline != cheapestRoute.Polyline {
				candidateRoutes = append(candidateRoutes, r)
			}
		}
		if len(candidateRoutes) > 0 {
			efficientRoute = selectEfficientRoute(candidateRoutes, 0.5)
		} else {
			efficientRoute = selectEfficientRoute(allRoutes, 0.5)
		}
	} else {
		fastestRoute = selectFastestRoute(allRoutes)
		cheapestRoute = selectCheapestRoute(allRoutes)
		efficientRoute = selectEfficientRoute(allRoutes, 0.5)
	}

	for i, r := range allRoutes {
		switch {
		case r.Polyline == fastestRoute.Polyline:
			allRoutes[i].Summary.RouteType = "fastest"
		case r.Polyline == cheapestRoute.Polyline:
			allRoutes[i].Summary.RouteType = "cheapest"
		case r.Polyline == efficientRoute.Polyline:
			allRoutes[i].Summary.RouteType = "efficient"
		default:
			allRoutes[i].Summary.RouteType = ""
		}
	}

	response := Response{
		SummaryRoute: summaryRoute,
		Routes:       allRoutes,
	}

	responseJSON, _ := json.Marshal(response)
	if err := cache.Rdb.Set(cache.Ctx, cacheKey, responseJSON, 30*24*time.Hour).Err(); err != nil {
		fmt.Printf("Erro ao salvar cache do Redis (CheckRouteTolls): %v\n", err)
		return Response{}, errors.New("erro ao salvar cache do Redis")
	}

	if frontInfo.PublicOrPrivate == "public" {
		if err := s.createRouteHist(ctx, id, frontInfo, responseJSON); err != nil {
			return Response{}, err
		}
	}

	waypointsStr := strings.ToLower(strings.Join(frontInfo.Waypoints, ","))
	_, err = s.InterfaceService.CreateSavedRoutes(ctx, db.CreateSavedRoutesParams{
		Origin:      origin.FormattedAddress,
		Destination: destination.FormattedAddress,
		Waypoints: sql.NullString{
			String: waypointsStr,
			Valid:  true,
		},
		Request:   requestJSON,
		Response:  responseJSON,
		ExpiredAt: time.Now().Add(30 * 24 * time.Hour),
	})

	return response, nil
}

func (s *Service) timeWithClient(ctx context.Context, client *maps.Client, origin, destination string) (Arrival, error) {
	cacheKey := fmt.Sprintf("timeWithClient:%s|%s", origin, destination)
	cached, err := cache.Rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		var arrival Arrival
		if json.Unmarshal([]byte(cached), &arrival) == nil {
			return arrival, nil
		}
	} else if !errors.Is(err, redis.Nil) {
		return Arrival{}, err
	}

	routeRequest := &maps.DirectionsRequest{
		Origin:       origin,
		Destination:  destination,
		Alternatives: false,
		Mode:         maps.TravelModeDriving,
		Region:       "br",
	}
	routes, _, err := client.Directions(ctx, routeRequest)
	if err != nil {
		return Arrival{}, err
	}
	if len(routes) == 0 || len(routes[0].Legs) == 0 {
		return Arrival{}, fmt.Errorf("rota ou trecho não encontrado para origem %s e destino %s", origin, destination)
	}

	leg := routes[0].Legs[0]
	arrival := Arrival{
		Distance: leg.Distance.HumanReadable,
		Time:     leg.Duration,
	}
	data, err := json.Marshal(arrival)
	if err == nil {
		if err := cache.Rdb.Set(ctx, cacheKey, data, 30*24*time.Hour).Err(); err != nil {
			fmt.Printf("Erro ao salvar cache no Redis (timeWithClient): %v\n", err)
		}
	}
	return arrival, nil
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
	fmt.Println(result)
	fmt.Println(results[0].Geometry.Location.Lat)
	fmt.Println(results[0].Geometry.Location.Lng)

	data, err := json.Marshal(result)
	if err == nil {
		if err := cache.Rdb.Set(cache.Ctx, cacheKey, data, 30*24*time.Hour).Err(); err != nil {
			fmt.Printf("Erro ao salvar cache do Redis (geocode): %v\n", err)
		}
	}
	return result, nil
}

func (s *Service) calculateTollsArrivalTimes(ctx context.Context, origin string, tolls []Toll) (map[int64]Arrival, error) {
	var destinations []string
	for _, toll := range tolls {
		dest := fmt.Sprintf("%.6f,%.6f", toll.Latitude, toll.Longitude)
		destinations = append(destinations, dest)
	}

	client, err := maps.NewClient(maps.WithAPIKey(s.GoogleMapsAPIKey))
	if err != nil {
		return nil, err
	}
	matrixReq := &maps.DistanceMatrixRequest{
		Origins:      []string{origin},
		Destinations: destinations,
		Mode:         maps.TravelModeDriving,
	}

	matrixResp, err := client.DistanceMatrix(ctx, matrixReq)
	if err != nil {
		return nil, err
	}

	arrivalMap := make(map[int64]Arrival)
	if len(matrixResp.Rows) > 0 {
		row := matrixResp.Rows[0]
		for i, elem := range row.Elements {
			tollID := tolls[i].ID
			arrivalMap[int64(tollID)] = Arrival{
				Distance: elem.Distance.HumanReadable,
				Time:     elem.Duration,
			}
		}
	}

	return arrivalMap, nil
}

func (s *Service) findTollsInRoute(ctx context.Context, routes []maps.Route, origin string, vehicle string, axes float64) ([]Toll, error) {
	tollsDB, err := s.InterfaceService.GetTollsByLonAndLat(ctx)
	if err != nil {
		return nil, err
	}

	resultTags, err := s.InterfaceService.GetTollTags(ctx)
	if err != nil {
		return nil, err
	}

	var routeIsCrescente bool
	if len(routes) > 0 && len(routes[0].Legs) > 0 {
		startLat := routes[0].Legs[0].StartLocation.Lat
		endLat := routes[0].Legs[len(routes[0].Legs)-1].EndLocation.Lat
		routeIsCrescente = startLat < endLat
	}

	uniqueTolls := make(map[int64]bool)
	var candidateTolls []Toll

	for _, route := range routes {
		for _, leg := range route.Legs {
			for _, step := range leg.Steps {
				polyPoints := DecodePolyline(step.Polyline.Points)
				if len(polyPoints) < 2 {
					continue
				}

				for _, dbToll := range tollsDB {
					lat, latErr := validation.ParseNullStringToFloat(dbToll.Latitude)
					lng, lngErr := validation.ParseNullStringToFloat(dbToll.Longitude)
					if latErr != nil || lngErr != nil {
						continue
					}

					tollPos := maps.LatLng{Lat: lat, Lng: lng}
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

					if dbToll.Sentido.String != "" {
						if dbToll.Sentido.String == "Crescente" && !routeIsCrescente {
							continue
						}
						if dbToll.Sentido.String == "Decrescente" && routeIsCrescente {
							continue
						}
					}

					imgConcession := getConcessionImage(dbToll.Concessionaria.String)
					if !uniqueTolls[dbToll.ID] {
						uniqueTolls[dbToll.ID] = true

						candidateTolls = append(candidateTolls, Toll{
							ID:            int(dbToll.ID),
							Latitude:      lat,
							Longitude:     lng,
							Name:          validation.GetStringFromNull(dbToll.PracaDePedagio),
							Concession:    dbToll.Concessionaria.String,
							ConcessionImg: imgConcession,
							Road:          validation.GetStringFromNull(dbToll.Rodovia),
							State:         validation.GetStringFromNull(dbToll.Uf),
							Country:       "Brasil",
							Type:          "Pedágio",
							FreeFlow:      dbToll.FreeFlow.Bool,
							PayFreeFlow:   dbToll.PayFreeFlow.String,
						})
					}
				}
			}
		}
	}

	arrivalMap, err := s.calculateTollsArrivalTimes(ctx, origin, candidateTolls)
	if err != nil {
		return nil, err
	}

	for i, cToll := range candidateTolls {
		arr := arrivalMap[int64(cToll.ID)]
		formattedTime := arr.Time.Round(time.Second).String()

		tarifaFloat := 0.0
		if tollsDB[i].Tarifa.Valid {
			tf, _ := strconv.ParseFloat(tollsDB[i].Tarifa.String, 64)
			tarifaFloat = tf
		}
		totalToll, errPrice := PriceTollsFromVehicle(strings.ToLower(vehicle), tarifaFloat, axes)
		if errPrice != nil {
			continue
		}

		var tags []string
		concession := validation.GetStringFromNull(tollsDB[i].Concessionaria)
		for _, tagRecord := range resultTags {
			acceptedList := strings.Split(tagRecord.DealershipAccepts, ",")
			for _, accepted := range acceptedList {
				if strings.TrimSpace(accepted) == concession {
					tags = append(tags, tagRecord.Name)
					break
				}
			}
		}

		candidateTolls[i].ArrivalResponse = ArrivalResponse{
			Distance: arr.Distance,
			Time:     formattedTime,
		}
		candidateTolls[i].TagCost = math.Round(totalToll - (totalToll * 0.05))
		candidateTolls[i].CashCost = totalToll
		candidateTolls[i].Currency = "BRL"
		candidateTolls[i].PrepaidCardCost = math.Round(totalToll - (totalToll * 0.05))
		candidateTolls[i].TagPrimary = tags
	}

	sort.Slice(candidateTolls, func(i, j int) bool {
		d1Str := strings.TrimSuffix(candidateTolls[i].ArrivalResponse.Distance, " km")
		d2Str := strings.TrimSuffix(candidateTolls[j].ArrivalResponse.Distance, " km")
		d1Str = strings.ReplaceAll(d1Str, ",", "")
		d2Str = strings.ReplaceAll(d2Str, ",", "")

		d1, err1 := strconv.ParseFloat(d1Str, 64)
		d2, err2 := strconv.ParseFloat(d2Str, 64)
		if err1 != nil || err2 != nil {
			return candidateTolls[i].ArrivalResponse.Distance < candidateTolls[j].ArrivalResponse.Distance
		}
		return d1 < d2
	})

	return candidateTolls, nil
}

//func (s *Service) findTollsInRoute(ctx context.Context, routes []maps.Route, origin string, vehicle string, axes float64) ([]Toll, error) {
//	var foundTolls []Toll
//	uniqueTolls := make(map[int64]bool)
//
//	tolls, err := s.InterfaceService.GetTollsByLonAndLat(ctx)
//	if err != nil {
//		return foundTolls, nil
//	}
//
//	resultTags, err := s.InterfaceService.GetTollTags(ctx)
//	if err != nil {
//		return nil, err
//	}
//
//	uniquePoints := make(map[string]maps.LatLng)
//	for _, route := range routes {
//		for _, leg := range route.Legs {
//			for _, step := range leg.Steps {
//				polyPoints := DecodePolyline(step.Polyline.Points)
//				for _, point := range polyPoints {
//					key := fmt.Sprintf("%f,%f", RoundCoord(point.Lat), RoundCoord(point.Lng))
//					uniquePoints[key] = point
//				}
//			}
//		}
//	}
//
//	var routeIsCrescente bool
//	if len(routes) > 0 && len(routes[0].Legs) > 0 {
//		startLat := routes[0].Legs[0].StartLocation.Lat
//		endLat := routes[0].Legs[len(routes[0].Legs)-1].EndLocation.Lat
//		routeIsCrescente = startLat < endLat
//	}
//
//	for _, point := range uniquePoints {
//		for _, dbToll := range tolls {
//			latitude, latErr := validation.ParseNullStringToFloat(dbToll.Latitude)
//			longitude, lonErr := validation.ParseNullStringToFloat(dbToll.Longitude)
//			if latErr != nil || lonErr != nil {
//				continue
//			}
//
//			if !IsNearby(point.Lat, point.Lng, latitude, longitude, 0.08) {
//				continue
//			}
//
//			if dbToll.Sentido.String != "" {
//				if dbToll.Sentido.String == "crescente" && !routeIsCrescente {
//					continue
//				}
//				if dbToll.Sentido.String == "decrescente" && routeIsCrescente {
//					continue
//				}
//			}
//
//			if !uniqueTolls[dbToll.ID] {
//				uniqueTolls[dbToll.ID] = true
//				dest := fmt.Sprintf("%.6f,%.6f", latitude, longitude)
//				arrivalTimes, err := s.calculateTimeToToll(ctx, origin, dest)
//				if err != nil {
//					fmt.Printf("Erro ao obter tempo para origem %s e destino %s: %v\n", origin, dest, err)
//					continue
//				}
//				formattedTime := arrivalTimes.Time.Round(time.Second).String()
//				concession := validation.GetStringFromNull(dbToll.Concessionaria)
//				var tags []string
//				for _, tagRecord := range resultTags {
//					acceptedList := strings.Split(tagRecord.DealershipAccepts, ",")
//					for _, accepted := range acceptedList {
//						if strings.TrimSpace(accepted) == concession {
//							tags = append(tags, tagRecord.Name)
//							break
//						}
//					}
//				}
//
//				tarifaFloat := 0.0
//				if dbToll.Tarifa.Valid {
//					tarifaFloat, err = strconv.ParseFloat(dbToll.Tarifa.String, 64)
//					if err != nil {
//						continue
//					}
//				}
//				totalToll, err := PriceTollsFromVehicle(strings.ToLower(vehicle), tarifaFloat, axes)
//				if err != nil {
//					return nil, err
//				}
//				foundTolls = append(foundTolls, Toll{
//					ID:              int(dbToll.ID),
//					Latitude:        latitude,
//					Longitude:       longitude,
//					Name:            validation.GetStringFromNull(dbToll.PracaDePedagio),
//					Concession:      dbToll.Concessionaria.String,
//					Road:            validation.GetStringFromNull(dbToll.Rodovia),
//					State:           validation.GetStringFromNull(dbToll.Uf),
//					Country:         "Brasil",
//					Type:            "Pedágio",
//					TagCost:         math.Round(totalToll - (totalToll * 0.05)),
//					CashCost:        totalToll,
//					Currency:        "BRL",
//					PrepaidCardCost: math.Round(totalToll - (totalToll * 0.05)),
//					ArrivalResponse: ArrivalResponse{
//						Distance: arrivalTimes.Distance,
//						Time:     formattedTime,
//					},
//					TagPrimary:  tags,
//					FreeFlow:    dbToll.FreeFlow.Bool,
//					PayFreeFlow: dbToll.PayFreeFlow.String,
//				})
//			}
//		}
//	}
//
//	sort.Slice(foundTolls, func(i, j int) bool {
//		return foundTolls[i].ArrivalResponse.Distance < foundTolls[j].ArrivalResponse.Distance
//	})
//
//	return foundTolls, nil
//}

func (s *Service) findBalancaInRoute(ctx context.Context, routes []maps.Route) ([]Balanca, error) {
	var foundBalancas []Balanca
	uniqueBalanca := make(map[int64]bool)

	tolls, err := s.InterfaceService.GetBalanca(ctx)
	if err != nil {
		return foundBalancas, nil
	}

	uniquePoints := make(map[string]maps.LatLng)
	for _, route := range routes {
		for _, leg := range route.Legs {
			for _, step := range leg.Steps {
				polyPoints := DecodePolyline(step.Polyline.Points)
				for _, point := range polyPoints {
					key := fmt.Sprintf("%f,%f", RoundCoord(point.Lat), RoundCoord(point.Lng))
					uniquePoints[key] = point
				}
			}
		}
	}

	for _, point := range uniquePoints {
		for _, dbBalanca := range tolls {
			latitude, latErr := validation.ParseStringToFloat(dbBalanca.Lat)
			longitude, lonErr := validation.ParseStringToFloat(dbBalanca.Lng)
			if latErr != nil || lonErr != nil {
				continue
			}
			if IsNearby(point.Lat, point.Lng, latitude, longitude, 1.0) {
				if !uniqueBalanca[dbBalanca.ID] {
					uniqueBalanca[dbBalanca.ID] = true
					foundBalancas = append(foundBalancas, Balanca{
						ID:             int(dbBalanca.ID),
						Concessionaria: dbBalanca.Concessionaria,
						Km:             dbBalanca.Km,
						Lat:            latitude,
						Lng:            longitude,
						Nome:           dbBalanca.Nome,
						Rodovia:        dbBalanca.Rodovia,
						Sentido:        dbBalanca.Sentido,
						Uf:             dbBalanca.Uf,
					})
				}
			}
		}
	}
	return foundBalancas, nil
}

//func (s *Service) findGasStationsAlongAllRoutes(ctx context.Context, client *maps.Client, routes []maps.Route) ([]GasStation, error) {
//
//	var gasStations []GasStation
//	uniqueGasStations := make(map[string]bool)
//
//	var points []maps.LatLng
//	uniquePoints := make(map[string]bool)
//
//	for _, route := range routes {
//		for _, leg := range route.Legs {
//			startKey := fmt.Sprintf("%.6f:%.6f", leg.StartLocation.Lat, leg.StartLocation.Lng)
//			if !uniquePoints[startKey] {
//				uniquePoints[startKey] = true
//				points = append(points, leg.StartLocation)
//			}
//			endKey := fmt.Sprintf("%.6f:%.6f", leg.EndLocation.Lat, leg.EndLocation.Lng)
//			if !uniquePoints[endKey] {
//				uniquePoints[endKey] = true
//				points = append(points, leg.EndLocation)
//			}
//		}
//	}
//
//	chunkSize := len(points)
//	if chunkSize == 0 {
//		return nil, nil
//	}
//
//	var wg sync.WaitGroup
//	var mu sync.Mutex
//
//	for i := 0; i < len(points); i += chunkSize {
//		end := i + chunkSize
//		if end > len(points) {
//			end = len(points)
//		}
//		chunk := points[i:end]
//
//		wg.Add(1)
//		go func(chunkPoints []maps.LatLng) {
//			defer wg.Done()
//
//			var sumLat, sumLng float64
//			for _, pt := range chunkPoints {
//				sumLat += pt.Lat
//				sumLng += pt.Lng
//			}
//			avgLat := sumLat / float64(len(chunkPoints))
//			avgLng := sumLng / float64(len(chunkPoints))
//
//			cacheKey := fmt.Sprintf("gasStationsMid:%.6f:%.6f", avgLat, avgLng)
//			cached, err := cache.Rdb.Get(ctx, cacheKey).Result()
//			if err == nil {
//				var cachedStations []GasStation
//				if json.Unmarshal([]byte(cached), &cachedStations) == nil {
//					mu.Lock()
//					for _, gs := range cachedStations {
//						stationKey := gs.Name
//						if stationKey == "" {
//							stationKey = gs.Address
//						}
//						if !uniqueGasStations[stationKey] {
//							uniqueGasStations[stationKey] = true
//							gasStations = append(gasStations, gs)
//						}
//					}
//					mu.Unlock()
//					return
//				}
//			}
//
//			placesRequest := &maps.NearbySearchRequest{
//				Location: &maps.LatLng{Lat: avgLat, Lng: avgLng},
//				Radius:   10000,
//				Type:     "gas_station",
//			}
//			ctxNearby, cancel := context.WithTimeout(ctx, 10*time.Second)
//			defer cancel()
//
//			placesResponse, err := client.NearbySearch(ctxNearby, placesRequest)
//			if err != nil {
//				fmt.Printf("Erro na NearbySearch: %v\n", err)
//				return
//			}
//
//			var resultCached []GasStation
//			mu.Lock()
//			for _, result := range placesResponse.Results {
//				stationName := result.Name
//				if stationName == "" {
//					stationName = result.PlaceID
//				}
//				if !uniqueGasStations[stationName] {
//					uniqueGasStations[stationName] = true
//					gs := GasStation{
//						Name:    stationName,
//						Address: result.Vicinity,
//						Location: Location{
//							Latitude:  result.Geometry.Location.Lat,
//							Longitude: result.Geometry.Location.Lng,
//						},
//					}
//					gasStations = append(gasStations, gs)
//					resultCached = append(resultCached, gs)
//
//					_, dbErr := s.InterfaceService.CreateGasStations(ctx, db.CreateGasStationsParams{
//						Name:          stationName,
//						Latitude:      fmt.Sprintf("%f", result.Geometry.Location.Lat),
//						Longitude:     fmt.Sprintf("%f", result.Geometry.Location.Lng),
//						AddressName:   result.Vicinity,
//						Municipio:     result.FormattedAddress,
//						SpecificPoint: result.PlaceID,
//					})
//					if dbErr != nil {
//						fmt.Printf("Erro ao salvar posto: %v\n", dbErr)
//					}
//				}
//			}
//			mu.Unlock()
//
//			if len(resultCached) > 0 {
//				data, _ := json.Marshal(resultCached)
//				_ = cache.Rdb.Set(ctx, cacheKey, data, 30*24*time.Hour).Err()
//			}
//
//		}(chunk)
//	}
//	wg.Wait()
//
//	return gasStations, nil
//}

func (s *Service) findGasStationsAlongAllRoutes(ctx context.Context, client *maps.Client, routes []maps.Route) ([]GasStation, error) {
	var gasStations []GasStation
	uniqueGasStations := make(map[string]bool)
	var consolidatedPoints []maps.LatLng
	uniquePoints := make(map[string]bool)

	for _, route := range routes {
		for _, leg := range route.Legs {
			startKey := fmt.Sprintf("%.6f:%.6f", leg.StartLocation.Lat, leg.StartLocation.Lng)
			if !uniquePoints[startKey] {
				uniquePoints[startKey] = true
				consolidatedPoints = append(consolidatedPoints, leg.StartLocation)
			}
			endKey := fmt.Sprintf("%.6f:%.6f", leg.EndLocation.Lat, leg.EndLocation.Lng)
			if !uniquePoints[endKey] {
				uniquePoints[endKey] = true
				consolidatedPoints = append(consolidatedPoints, leg.EndLocation)
			}
			for _, step := range leg.Steps {
				pointKey := fmt.Sprintf("%.6f:%.6f", step.StartLocation.Lat, step.StartLocation.Lng)
				if !uniquePoints[pointKey] {
					uniquePoints[pointKey] = true
					consolidatedPoints = append(consolidatedPoints, step.StartLocation)
				}
			}
		}
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	chunkSize := 5

	for i := 0; i < len(consolidatedPoints); i += chunkSize {
		end := i + chunkSize
		if end > len(consolidatedPoints) {
			end = len(consolidatedPoints)
		}
		wg.Add(1)
		go func(points []maps.LatLng) {
			defer wg.Done()
			for _, point := range points {
				cacheKey := fmt.Sprintf("gasStations:%.6f:%.6f", point.Lat, point.Lng)
				cached, err := cache.Rdb.Get(ctx, cacheKey).Result()
				if err == nil {
					var cachedStations []GasStation
					if json.Unmarshal([]byte(cached), &cachedStations) == nil {
						mu.Lock()
						for _, gs := range cachedStations {
							stationKey := gs.Name
							if stationKey == "" {
								stationKey = gs.Address
							}
							if !uniqueGasStations[stationKey] {
								uniqueGasStations[stationKey] = true
								gasStations = append(gasStations, gs)
							}
						}
						mu.Unlock()
						continue
					}
				} else if err != redis.Nil {
					fmt.Printf("Erro ao recuperar cache Redis para gasStations: %v\n", err)
				}

				dbGasStations, err := s.InterfaceService.GetGasStation(ctx, db.GetGasStationParams{
					Column1: point.Lat,
					Column2: point.Lng,
					Column3: 0.05,
				})
				if err != nil {
					fmt.Printf("Erro ao consultar o banco de dados: %v\n", err)
					continue
				}

				var cachedResult []GasStation
				if len(dbGasStations) > 0 {
					mu.Lock()
					for _, dbStation := range dbGasStations {
						stationName := dbStation.Name
						if stationName == "" {
							stationName = dbStation.SpecificPoint
						}
						if !uniqueGasStations[stationName] {
							uniqueGasStations[stationName] = true
							lat, _ := validation.ParseStringToFloat(dbStation.Latitude)
							lng, _ := validation.ParseStringToFloat(dbStation.Longitude)
							gs := GasStation{
								Name:    stationName,
								Address: dbStation.AddressName,
								Location: Location{
									Latitude:  lat,
									Longitude: lng,
								},
							}
							gasStations = append(gasStations, gs)
							cachedResult = append(cachedResult, gs)
						}
					}
					mu.Unlock()
					if len(cachedResult) > 0 {
						data, err := json.Marshal(cachedResult)
						if err == nil {
							if err := cache.Rdb.Set(ctx, cacheKey, data, 30*24*time.Hour).Err(); err != nil {
								fmt.Printf("Erro ao salvar cache Redis para gasStations: %v\n", err)
							}
						}
					}
				}
			}
		}(consolidatedPoints[i:end])
	}
	wg.Wait()
	return gasStations, nil
}

func (s *Service) updateNumberOfRequest(ctx context.Context, id int64) error {
	result, err := s.InterfaceService.GetTokenHist(ctx, id)
	if err != nil {
		return err
	}
	number := result.NumberRequest + 1
	if number > 5 {
		return errors.New("you have reached the limit of requests per day")
	}
	err = s.InterfaceService.UpdateNumberOfRequest(ctx, db.UpdateNumberOfRequestParams{
		NumberRequest: number,
		ID:            id,
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) createRouteHist(ctx context.Context, idTokenHist int64, info FrontInfo, response json.RawMessage) error {
	waypoints := strings.ToLower(strings.Join(info.Waypoints, ","))
	_, err := s.InterfaceService.CreateRouteHist(ctx, db.CreateRouteHistParams{
		IDTokenHist: idTokenHist,
		Origin:      info.Origin,
		Destination: info.Destination,
		Waypoints: sql.NullString{
			String: waypoints,
			Valid:  true,
		},
		Response: response,
	})
	if err != nil {
		return err
	}
	return nil
}

type FreightItem struct {
	Carga int64   `json:"carga"`
	Valor float64 `json:"valor"`
}

func (s *Service) getAllFreight(ctx context.Context, axles int64, kmValue float64) (map[string]interface{}, error) {
	results, err := s.InterfaceService.GetFreightLoadAll(ctx)
	if err != nil {
		return nil, err
	}

	grouped := make(map[string][]FreightLoad)
	for _, result := range results {
		var fl FreightLoad
		fl.ParseFromNcmObject(result)
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

const EarthRadius = 6371000

func distanceInMeters(a, b maps.LatLng) float64 {
	lat1 := a.Lat * math.Pi / 180
	lng1 := a.Lng * math.Pi / 180
	lat2 := b.Lat * math.Pi / 180
	lng2 := b.Lng * math.Pi / 180

	dLat := lat2 - lat1
	dLng := lng2 - lng1

	h := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Sin(dLng/2)*math.Sin(dLng/2)

	c := 2 * math.Atan2(math.Sqrt(h), math.Sqrt(1-h))
	return EarthRadius * c
}

func distancePointToSegment(p, A, B maps.LatLng) float64 {
	distAB := distanceInMeters(A, B)
	if distAB == 0 {
		return distanceInMeters(p, A)
	}

	Ax, Ay := project(A)
	Bx, By := project(B)
	Px, Py := project(p)

	vx := Bx - Ax
	vy := By - Ay

	wx := Px - Ax
	wy := Py - Ay

	c1 := (wx*vx + wy*vy) / (vx*vx + vy*vy)

	if c1 < 0 {
		return distanceInMeters(p, A)
	} else if c1 > 1 {
		return distanceInMeters(p, B)
	}

	projx := Ax + c1*vx
	projy := Ay + c1*vy

	dx := Px - projx
	dy := Py - projy
	distProj := math.Sqrt(dx*dx + dy*dy)

	return distProj
}

func project(ll maps.LatLng) (float64, float64) {
	latRad := ll.Lat * math.Pi / 180
	lngRad := ll.Lng * math.Pi / 180

	x := EarthRadius * lngRad * math.Cos(latRad)
	y := EarthRadius * latRad
	return x, y
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
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/rodoanel_oeste.png"
	case "CSG":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/csg.png"
	case "ROTA DAS BANDEIRAS":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/rota_das_bandeiras.png"
	case "CONCEF":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/concef.png"
	case "TRIUNFO":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/triunfo.png"
	case "ECO 050":
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/eco_050.png"
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
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/epr_triangulo.png"
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
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/epr_via_mineira.png"
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
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/concebra.png"
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
		return "https://dealership-routes.s3.us-east-1.amazonaws.com/ccr_riosp.png"
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
