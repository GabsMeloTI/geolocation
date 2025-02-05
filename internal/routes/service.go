package routes

import (
	"context"
	"database/sql"
	_ "encoding/json"
	"errors"
	"fmt"
	db "geolocation/db/sqlc"
	"github.com/labstack/echo/v4"
	"googlemaps.github.io/maps"
	"html"
	"math"
	neturl "net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type InterfaceService interface {
	CheckRouteTolls(ctx context.Context, frontInfo FrontInfo) (Response, error)
	AddSavedRoutesFavorite(ctx context.Context, id int32) error
}

type Service struct {
	InterfaceService InterfaceRepository
	GoogleMapsAPIKey string
}

func NewRoutesService(InterfaceService InterfaceRepository, GoogleMapsAPIKey string) *Service {
	return &Service{InterfaceService, GoogleMapsAPIKey}
}

func (s *Service) CheckRouteTolls(ctx context.Context, frontInfo FrontInfo) (Response, error) {
	client, err := maps.NewClient(maps.WithAPIKey(s.GoogleMapsAPIKey))
	if err != nil {
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

	routeRequest := &maps.DirectionsRequest{
		Origin:       origin.FormattedAddress,
		Destination:  destination.FormattedAddress,
		Mode:         maps.TravelModeDriving,
		Waypoints:    frontInfo.Waypoints,
		Alternatives: true,
		Region:       "br",
		Language:     "pt-BR",
	}

	if strings.ToUpper(frontInfo.Type) == "BARATO" {
		routeRequest.Avoid = []maps.Avoid{"tolls"}
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
	var minTollCost = math.MaxFloat64

	for _, route := range routes {
		foundTolls, _ := s.findTollsInRoute(ctx, client, []maps.Route{route}, origin.FormattedAddress, frontInfo.Type, float64(frontInfo.Axles))
		findGasStationsAlongRoute, _ := s.findGasStationsAlongAllRoutes(ctx, client, []maps.Route{route})

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
				instr := removeHTMLTags(step.HTMLInstructions)
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

		fuelCost := math.Round((float64(totalDistance)/1000.0/frontInfo.ConsumptionHwy*frontInfo.Price)*100) / 100

		url := fmt.Sprintf("https://www.google.com/maps/dir/?api=1&origin=%s&destination=%s",
			neturl.QueryEscape(origin.FormattedAddress), neturl.QueryEscape(destination.FormattedAddress),
		)
		if len(frontInfo.Waypoints) > 0 {
			url += "&waypoints=" + neturl.QueryEscape(strings.Join(frontInfo.Waypoints, "|"))
		}

		currentTimeMillis := (time.Now().UnixNano() + lastLeg.Duration.Nanoseconds()) / int64(time.Millisecond)
		wazeURL := fmt.Sprintf(
			"https://www.waze.com/pt-BR/live-map/directions/br/mt/cuiaba?to=place.%s&from=place.%s&time=%d&reverse=yes",
			destination.PlaceID,
			origin.PlaceID,
			currentTimeMillis,
		)

		polyline := route.OverviewPolyline.Points
		rotogramaURL := fmt.Sprintf(
			"https://maps.googleapis.com/maps/api/staticmap?size=600x400&path=enc:%s&markers=color:green%%7Clabel:S%%7C%s&markers=color:red%%7Clabel:E%%7C%s&key=%s",
			polyline,
			neturl.QueryEscape(origin.FormattedAddress),
			neturl.QueryEscape(destination.FormattedAddress),
			s.GoogleMapsAPIKey,
		)

		allRoutes = append(allRoutes, Route{
			Summary: Summary{
				HasTolls: len(foundTolls) > 0,
				Distance: Distance{
					Text:  fmt.Sprintf("%d km", totalDistance/1000),
					Value: totalDistance,
				},
				Duration: Duration{
					Text:  totalDuration.String(),
					Value: totalDuration.Seconds(),
				},
				URL:     url,
				URLWaze: wazeURL,
			},
			Costs: Costs{
				TagAndCash:      totalTollCost,
				Fuel:            fuelCost,
				Tag:             totalTollCost,
				Cash:            totalTollCost,
				PrepaidCard:     totalTollCost,
				MaximumTollCost: maxTollCost,
				MinimumTollCost: minTollCost,
			},
			Tolls:        foundTolls,
			Polyline:     route.OverviewPolyline.Points,
			GasStations:  findGasStationsAlongRoute,
			Rotograma:    rotogramaURL,
			Instructions: instructions,
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

	selectedRoute := selectBestRoute(allRoutes, frontInfo.TypeRoute)

	response := Response{
		SummaryRoute: summaryRoute,
		Routes:       []Route{selectedRoute},
	}

	return response, nil
}

func removeHTMLTags(s string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(s, " ")
}

func selectBestRoute(routes []Route, routeType string) Route {
	if len(routes) == 0 {
		return Route{}
	}

	selected := routes[0]
	switch strings.ToUpper(routeType) {
	case "RÁPIDA":
		for _, r := range routes {
			if r.Summary.Duration.Value < selected.Summary.Duration.Value {
				selected = r
			}
		}
	case "BARATO":
		for _, r := range routes {
			if r.Costs.TagAndCash < selected.Costs.TagAndCash {
				selected = r
			}
		}
	case "EFICIENTE":
		for _, r := range routes {
			if (r.Costs.Fuel + r.Costs.TagAndCash) < (selected.Costs.Fuel + selected.Costs.TagAndCash) {
				selected = r
			}
		}
	default:
		for _, r := range routes {
			if r.Summary.Duration.Value < selected.Summary.Duration.Value {
				selected = r
			}
		}
	}
	return selected
}

func (s *Service) timeWithClient(ctx context.Context, client *maps.Client, origin, destination string) (Arrival, error) {
	cacheKey := fmt.Sprintf("%s|%s", origin, destination)

	timeCacheMutex.RLock()
	if arrival, exists := timeCache[cacheKey]; exists {
		timeCacheMutex.RUnlock()
		return arrival, nil
	}
	timeCacheMutex.RUnlock()

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
		return Arrival{}, fmt.Errorf("no route/leg found for origin %s to destination %s", origin, destination)
	}

	leg := routes[0].Legs[0]
	arrival := Arrival{
		Distance: leg.Distance.HumanReadable,
		Time:     leg.Duration,
	}

	timeCacheMutex.Lock()
	timeCache[cacheKey] = arrival
	timeCacheMutex.Unlock()

	return arrival, nil
}

//func (s *Service) findTollsInRoute(ctx context.Context, routes []maps.Route, origin, vehicle string, axes float64) ([]Toll, error) {
//	var foundTolls []Toll
//	uniquePoints := make(map[string]bool)
//	uniqueTolls := make(map[int64]bool)
//
//	tolls, err := s.InterfaceService.GetTollsByLonAndLat(ctx)
//	if err != nil {
//		return nil, err
//	}
//
//	resultTags, err := s.InterfaceService.GetTollTags(ctx)
//	if err != nil {
//		return nil, err
//	}
//
//	var consolidatedPoints []maps.LatLng
//	for _, route := range routes {
//		for _, leg := range route.Legs {
//			for _, step := range leg.Steps {
//				pointKey := fmt.Sprintf("%f,%f", step.StartLocation.Lat, step.StartLocation.Lng)
//				if !uniquePoints[pointKey] {
//					uniquePoints[pointKey] = true
//					consolidatedPoints = append(consolidatedPoints, step.StartLocation)
//				}
//			}
//		}
//	}
//
//	for _, point := range consolidatedPoints {
//		for _, dbToll := range tolls {
//			latitude, latErr := parseNullStringToFloat(dbToll.Latitude)
//			longitude, lonErr := parseNullStringToFloat(dbToll.Longitude)
//			if latErr != nil || lonErr != nil {
//				continue
//			}
//
//			if isNearby(point.Lat, point.Lng, latitude, longitude, 5.0) {
//				if !uniqueTolls[dbToll.ID] {
//					uniqueTolls[dbToll.ID] = true
//
//					arrivalTimes, _ := s.time(ctx, origin, fmt.Sprintf("%f,%f", latitude, longitude))
//					formattedTime := arrivalTimes.Time.Round(time.Second).String()
//
//					concession := getStringFromNull(dbToll.Concessionaria)
//					var tags []string
//					for _, tagRecord := range resultTags {
//						acceptedList := strings.Split(tagRecord.DealershipAccepts, ",")
//						for _, accepted := range acceptedList {
//							if strings.TrimSpace(accepted) == concession {
//								tags = append(tags, tagRecord.Name)
//								break
//							}
//						}
//					}
//
//					totalToll, err := priceTollsFromVehicle(vehicle, dbToll.Tarifa.Float64, axes)
//					if err != nil {
//						return nil, err
//					}
//
//					foundTolls = append(foundTolls, Toll{
//						ID:              int(dbToll.ID),
//						Latitude:        latitude,
//						Longitude:       longitude,
//						Name:            getStringFromNull(dbToll.PracaDePedagio),
//						Concession:      dbToll.Concessionaria.String,
//						Road:            getStringFromNull(dbToll.Rodovia),
//						State:           getStringFromNull(dbToll.Uf),
//						Country:         "Brasil",
//						Type:            "Pedágio",
//						TagCost:         math.Round(totalToll - (totalToll * 0.05)),
//						CashCost:        totalToll,
//						Currency:        "BRL",
//						PrepaidCardCost: math.Round(totalToll - (totalToll * 0.05)),
//						ArrivalResponse: ArrivalResponse{
//							Distance: arrivalTimes.Distance,
//							Time:     formattedTime,
//						},
//						TagPrimary: tags,
//					})
//				}
//			}
//		}
//	}
//
//	return foundTolls, nil
//}

func (s *Service) findTollsInRoute(ctx context.Context, client *maps.Client, routes []maps.Route, origin, vehicle string, axes float64) ([]Toll, error) {
	var foundTolls []Toll
	uniqueTolls := make(map[int64]bool)

	tolls, err := s.InterfaceService.GetTollsByLonAndLat(ctx)
	if err != nil {
		return foundTolls, nil
	}

	resultTags, err := s.InterfaceService.GetTollTags(ctx)
	if err != nil {
		return nil, err
	}

	uniquePoints := make(map[string]maps.LatLng)
	for _, route := range routes {
		for _, leg := range route.Legs {
			for _, step := range leg.Steps {
				polyPoints := decodePolyline(step.Polyline.Points)
				for _, point := range polyPoints {
					key := fmt.Sprintf("%f,%f", roundCoord(point.Lat), roundCoord(point.Lng))
					uniquePoints[key] = point
				}
			}
		}
	}

	localTimeCache := make(map[string]Arrival)
	for _, point := range uniquePoints {
		for _, dbToll := range tolls {
			latitude, latErr := parseNullStringToFloat(dbToll.Latitude)
			longitude, lonErr := parseNullStringToFloat(dbToll.Longitude)
			if latErr != nil || lonErr != nil {
				continue
			}

			if isNearby(point.Lat, point.Lng, latitude, longitude, 5.0) {
				if !uniqueTolls[dbToll.ID] {
					uniqueTolls[dbToll.ID] = true

					tollKey := fmt.Sprintf("%s|%0.3f,%0.3f", origin, roundCoord(latitude), roundCoord(longitude))
					var arrivalTimes Arrival
					var ok bool
					if arrivalTimes, ok = localTimeCache[tollKey]; !ok {
						dest := fmt.Sprintf("%.6f,%.6f", latitude, longitude)
						var err error
						arrivalTimes, err = s.timeWithClient(ctx, client, origin, dest)
						if err != nil {
							fmt.Printf("Error calling timeWithClient for tollKey %s: %v\n", tollKey, err)
							continue
						}
						localTimeCache[tollKey] = arrivalTimes
					}
					formattedTime := arrivalTimes.Time.Round(time.Second).String()

					concession := getStringFromNull(dbToll.Concessionaria)
					var tags []string
					for _, tagRecord := range resultTags {
						acceptedList := strings.Split(tagRecord.DealershipAccepts, ",")
						for _, accepted := range acceptedList {
							if strings.TrimSpace(accepted) == concession {
								tags = append(tags, tagRecord.Name)
								break
							}
						}
					}

					totalToll, err := priceTollsFromVehicle(vehicle, dbToll.Tarifa.Float64, axes)
					if err != nil {
						return nil, err
					}

					foundTolls = append(foundTolls, Toll{
						ID:              int(dbToll.ID),
						Latitude:        latitude,
						Longitude:       longitude,
						Name:            getStringFromNull(dbToll.PracaDePedagio),
						Concession:      dbToll.Concessionaria.String,
						Road:            getStringFromNull(dbToll.Rodovia),
						State:           getStringFromNull(dbToll.Uf),
						Country:         "Brasil",
						Type:            "Pedágio",
						TagCost:         math.Round(totalToll - (totalToll * 0.05)),
						CashCost:        totalToll,
						Currency:        "BRL",
						PrepaidCardCost: math.Round(totalToll - (totalToll * 0.05)),
						ArrivalResponse: ArrivalResponse{
							Distance: arrivalTimes.Distance,
							Time:     formattedTime,
						},
						TagPrimary: tags,
					})
				}
			}
		}
	}
	return foundTolls, nil
}

var (
	timeCache      = make(map[string]Arrival)
	timeCacheMutex = sync.RWMutex{}
)

func (s *Service) time(ctx context.Context, origin, destination string) (Arrival, error) {
	cacheKey := fmt.Sprintf("%s|%s", origin, destination)
	timeCacheMutex.RLock()
	if arrival, exists := timeCache[cacheKey]; exists {
		timeCacheMutex.RUnlock()
		return arrival, nil
	}
	timeCacheMutex.RUnlock()

	client, err := maps.NewClient(maps.WithAPIKey(s.GoogleMapsAPIKey))
	if err != nil {
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
	if len(routes) == 0 {
		return Arrival{}, echo.ErrNotFound
	}

	leg := routes[0].Legs[0]
	arrival := Arrival{
		Distance: leg.Distance.HumanReadable,
		Time:     leg.Duration,
	}

	timeCacheMutex.Lock()
	timeCache[cacheKey] = arrival
	timeCacheMutex.Unlock()

	return arrival, nil
}

//var (
//	timeCache      = make(map[string]Arrival)
//	timeCacheMutex = sync.RWMutex{}
//)
//
//func (s *Service) time(ctx context.Context, origin, destination string) (Arrival, error) {
//	cacheKey := fmt.Sprintf("%s|%s", origin, destination)
//
//	timeCacheMutex.RLock()
//	if arrival, exists := timeCache[cacheKey]; exists {
//		timeCacheMutex.RUnlock()
//		return arrival, nil
//	}
//	timeCacheMutex.RUnlock()
//
//	client, err := maps.NewClient(maps.WithAPIKey(s.GoogleMapsAPIKey))
//	if err != nil {
//		return Arrival{}, err
//	}
//
//	routeRequest := &maps.DirectionsRequest{
//		Origin:       origin,
//		Destination:  destination,
//		Alternatives: false,
//		Mode:         maps.TravelModeDriving,
//		Region:       "br",
//	}
//	routes, _, err := client.Directions(ctx, routeRequest)
//	if err != nil {
//		return Arrival{}, err
//	}
//	if len(routes) == 0 {
//		return Arrival{}, echo.ErrNotFound
//	}
//
//	leg := routes[0].Legs[0]
//	arrival := Arrival{
//		Distance: leg.Distance.HumanReadable,
//		Time:     leg.Duration,
//	}
//
//	timeCacheMutex.Lock()
//	timeCache[cacheKey] = arrival
//	timeCacheMutex.Unlock()
//
//	return arrival, nil
//}

var (
	gasStationsCache      = make(map[string][]GasStation)
	gasStationsCacheMutex = sync.RWMutex{}
)

func roundCoord(coord float64) float64 {
	return math.Round(coord*1000) / 1000
}

func getCacheKeyForPoint(lat, lng float64) string {
	return fmt.Sprintf("%f,%f", roundCoord(lat), roundCoord(lng))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (s *Service) findGasStationsAlongAllRoutes(ctx context.Context, client *maps.Client, routes []maps.Route) ([]GasStation, error) {
	var gasStations []GasStation
	uniquePoints := make(map[string]bool)
	uniqueGasStations := make(map[string]bool)

	var consolidatedPoints []maps.LatLng
	for _, route := range routes {
		for _, leg := range route.Legs {
			for _, step := range leg.Steps {
				pointKey := getCacheKeyForPoint(step.StartLocation.Lat, step.StartLocation.Lng)
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
		wg.Add(1)
		go func(points []maps.LatLng) {
			defer wg.Done()

			for _, point := range points {
				cacheKey := getCacheKeyForPoint(point.Lat, point.Lng)

				gasStationsCacheMutex.RLock()
				cachedStations, found := gasStationsCache[cacheKey]
				gasStationsCacheMutex.RUnlock()
				if found {
					mu.Lock()
					for _, gs := range cachedStations {
						if !uniqueGasStations[gs.Name] {
							uniqueGasStations[gs.Name] = true
							gasStations = append(gasStations, gs)
						}
					}
					mu.Unlock()
					continue
				}

				dbGasStations, err := s.InterfaceService.GetGasStation(ctx, db.GetGasStationParams{
					Column1: point.Lat,
					Column2: point.Lng,
					Column3: 0.05,
				})
				if err != nil {
					fmt.Printf("Erro ao consultar banco: %v\n", err)
					continue
				}

				var cachedResult []GasStation
				if len(dbGasStations) > 0 {
					mu.Lock()
					for _, dbStation := range dbGasStations {
						if !uniqueGasStations[dbStation.SpecificPoint] {
							uniqueGasStations[dbStation.SpecificPoint] = true
							gs := GasStation{
								Name:    dbStation.SpecificPoint,
								Address: dbStation.AddressName,
								Location: Location{
									Latitude:  point.Lat,
									Longitude: point.Lng,
								},
							}
							gasStations = append(gasStations, gs)
							cachedResult = append(cachedResult, gs)
						}
					}
					mu.Unlock()
					gasStationsCacheMutex.Lock()
					gasStationsCache[cacheKey] = cachedResult
					gasStationsCacheMutex.Unlock()
					continue
				}

				placesRequest := &maps.NearbySearchRequest{
					Location: &maps.LatLng{Lat: point.Lat, Lng: point.Lng},
					Radius:   10000,
					Type:     "gas_station",
				}
				placesResponse, err := client.NearbySearch(ctx, placesRequest)
				if err != nil {
					fmt.Printf("Erro na NearbySearch: %v\n", err)
					continue
				}

				mu.Lock()
				for _, result := range placesResponse.Results {
					if !uniqueGasStations[result.Name] {
						uniqueGasStations[result.Name] = true

						gs := GasStation{
							Name:    result.Name,
							Address: result.Vicinity,
							Location: Location{
								Latitude:  result.Geometry.Location.Lat,
								Longitude: result.Geometry.Location.Lng,
							},
						}
						gasStations = append(gasStations, gs)
						cachedResult = append(cachedResult, gs)

						_, err := s.InterfaceService.CreateGasStations(ctx, db.CreateGasStationsParams{
							Name:          result.Name,
							Latitude:      fmt.Sprintf("%f", result.Geometry.Location.Lat),
							Longitude:     fmt.Sprintf("%f", result.Geometry.Location.Lng),
							AddressName:   result.Vicinity,
							Municipio:     result.FormattedAddress,
							SpecificPoint: result.PlaceID,
						})
						if err != nil {
							fmt.Printf("Erro ao salvar posto: %v\n", err)
						}
					}
				}
				mu.Unlock()

				gasStationsCacheMutex.Lock()
				gasStationsCache[cacheKey] = cachedResult
				gasStationsCacheMutex.Unlock()
			}
		}(consolidatedPoints[i:min(i+chunkSize, len(consolidatedPoints))])
	}

	wg.Wait()
	return gasStations, nil
}

var (
	geocodeCache      = make(map[string]GeocodeResult)
	geocodeCacheMutex = sync.RWMutex{}
)

func (s *Service) getGeocodeAddress(ctx context.Context, address string) (GeocodeResult, error) {
	address = stateToCapital(address)

	geocodeCacheMutex.RLock()
	if cachedResult, found := geocodeCache[address]; found {
		geocodeCacheMutex.RUnlock()
		return cachedResult, nil
	}
	geocodeCacheMutex.RUnlock()

	client, err := maps.NewClient(maps.WithAPIKey(s.GoogleMapsAPIKey))
	if err != nil {
		return GeocodeResult{}, fmt.Errorf("erro ao criar cliente Google Maps: %v", err)
	}

	if strings.ToLower(address) == "bahia" {
		address = "Salvador, Bahia"
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
		return GeocodeResult{}, fmt.Errorf("Endereço não encontrado para: %s. Verifique se a pesquisa está escrita corretamente ou seja mais específico(Ex: %s, São Paulo).", address, address)
	}

	result := GeocodeResult{
		FormattedAddress: results[0].FormattedAddress,
		PlaceID:          results[0].PlaceID,
	}

	geocodeCacheMutex.Lock()
	geocodeCache[address] = result
	geocodeCacheMutex.Unlock()

	return result, nil
}

func stateToCapital(address string) string {
	state := strings.ToLower(strings.TrimSpace(address))

	switch state {
	case "acre":
		return "Rio Branco, Acre"
	case "alagoas":
		return "Maceió, Alagoas"
	case "amapá", "amapa":
		return "Macapá, Amapá"
	case "amazonas":
		return "Manaus, Amazonas"
	case "bahia":
		return "Salvador, Bahia"
	case "ceará", "ceara":
		return "Fortaleza, Ceará"
	case "espírito santo", "espirito santo":
		return "Vitória, Espírito Santo"
	case "goiás", "goias":
		return "Goiânia, Goiás"
	case "maranhão", "maranhao":
		return "São Luís, Maranhão"
	case "mato grosso":
		return "Cuiabá, Mato Grosso"
	case "mato grosso do sul":
		return "Campo Grande, Mato Grosso do Sul"
	case "minas gerais":
		return "Belo Horizonte, Minas Gerais"
	case "pará", "para":
		return "Belém, Pará"
	case "paraíba", "paraiba":
		return "João Pessoa, Paraíba"
	case "paraná", "parana":
		return "Curitiba, Paraná"
	case "pernambuco":
		return "Recife, Pernambuco"
	case "piauí", "piaui":
		return "Teresina, Piauí"
	case "rio de janeiro":
		return "Rio de Janeiro, Rio de Janeiro"
	case "rio grande do norte":
		return "Natal, Rio Grande do Norte"
	case "rio grande do sul":
		return "Porto Alegre, Rio Grande do Sul"
	case "rondônia", "rondonia":
		return "Porto Velho, Rondônia"
	case "roraima":
		return "Boa Vista, Roraima"
	case "santa catarina":
		return "Florianópolis, Santa Catarina"
	case "são paulo", "sao paulo":
		return "São Paulo, São Paulo"
	case "sergipe":
		return "Aracaju, Sergipe"
	case "tocantins":
		return "Palmas, Tocantins"
	case "distrito federal":
		return "Brasília, Distrito Federal"
	default:
		return address
	}
}

func parseNullStringToFloat(nullString sql.NullString) (float64, error) {
	if nullString.Valid {
		return strconv.ParseFloat(nullString.String, 64)
	}
	return 0, fmt.Errorf("valor nulo")
}

func getStringFromNull(nullString sql.NullString) string {
	if nullString.Valid {
		return nullString.String
	}
	return ""
}

func isNearby(lat1, lng1, lat2, lng2, radius float64) bool {
	const earthRadius = 6371

	dLat := (lat2 - lat1) * (math.Pi / 180)
	dLng := (lng2 - lng1) * (math.Pi / 180)

	lat1Rad := lat1 * (math.Pi / 180)
	lat2Rad := lat2 * (math.Pi / 180)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLng/2)*math.Sin(dLng/2)*math.Cos(lat1Rad)*math.Cos(lat2Rad)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	distance := earthRadius * c
	return distance <= radius
}

func (s *Service) AddSavedRoutesFavorite(ctx context.Context, id int32) error {
	_, err := s.InterfaceService.GetSavedRouteById(ctx, id)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return errors.New("route not found")
		}
		return err
	}

	err = s.InterfaceService.AddSavedRoutesFavorite(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

func priceTollsFromVehicle(vehicle string, price, axes float64) (float64, error) {
	var calculation float64
	switch os := vehicle; os {
	case "Motorcycle":
		calculation = price / 2
		return calculation, nil
	case "Auto":
		if int(axes)%2 != 0 {
			price = price / 2
		}
		calculation = price * axes
		return calculation, nil
	case "Bus":
		if int(axes)%2 != 0 {
			price = price / 2
		}
		calculation = price * axes
		return calculation, nil
	case "Truck":
		if int(axes)%2 != 0 {
			price = price / 2
		}
		calculation = price * axes
		return calculation, nil
	default:
		fmt.Printf("incoorect value")
	}

	return calculation, nil
}

func decodePolyline(encoded string) []maps.LatLng {
	var points []maps.LatLng
	index, lat, lng := 0, 0, 0
	for index < len(encoded) {
		var result, shift uint
		for {
			b := encoded[index] - 63
			index++
			result |= uint(b&0x1F) << shift
			shift += 5
			if b < 0x20 {
				break
			}
		}
		dlat := int(result)
		if dlat&1 != 0 {
			dlat = ^(dlat >> 1)
		} else {
			dlat = dlat >> 1
		}
		lat += dlat
		shift, result = 0, 0
		for {
			b := encoded[index] - 63
			index++
			result |= uint(b&0x1F) << shift
			shift += 5
			if b < 0x20 {
				break
			}
		}
		dlng := int(result)
		if dlng&1 != 0 {
			dlng = ^(dlng >> 1)
		} else {
			dlng = dlng >> 1
		}
		lng += dlng
		points = append(points, maps.LatLng{
			Lat: float64(lat) / 1e5,
			Lng: float64(lng) / 1e5,
		})
	}
	return points
}
