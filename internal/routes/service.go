package routes

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/labstack/echo/v4"
	"googlemaps.github.io/maps"
	"math"
	neturl "net/url"
	"strconv"
	"strings"
	"time"
)

type InterfaceService interface {
	CheckRouteTolls(ctx context.Context, frontInfo FrontInfo) (Response, error)
}

type Service struct {
	InterfaceService InterfaceRepository
}

func NewRoutesService(InterfaceService InterfaceRepository) *Service {
	return &Service{InterfaceService}
}

func (s *Service) CheckRouteTolls(ctx context.Context, frontInfo FrontInfo) (Response, error) {
	apiKey := "AIzaSyAbzUZuCp1zcNNkzje_kmwVVqyOI5w8jkQ"

	client, err := maps.NewClient(maps.WithAPIKey(apiKey))
	if err != nil {
		return Response{}, err
	}

	origin, err := getGeocodeAddress(ctx, frontInfo.Origin)
	if err != nil {
		return Response{}, err
	}

	destination, err := getGeocodeAddress(ctx, frontInfo.Destination)
	if err != nil {
		return Response{}, err
	}

	routeRequest := &maps.DirectionsRequest{
		Origin:       origin,
		Destination:  destination,
		Mode:         maps.TravelModeDriving,
		Waypoints:    frontInfo.Waypoints,
		Alternatives: true,
		Region:       "br",
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
		foundTolls, _ := s.findTollsInRoute(ctx, []maps.Route{route}, frontInfo.Origin)
		//findGasStationsAlongRoute, err := s.findGasStationsAlongAllRoutes(ctx, client, []maps.Route{route})
		if err != nil {
			return Response{}, err
		}

		var totalDistance int
		var totalDuration time.Duration
		var totalTollCost float64

		var locations []PrincipalRoute
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
			neturl.QueryEscape(origin), neturl.QueryEscape(destination),
		)
		if len(frontInfo.Waypoints) > 0 {
			url += "&waypoints=" + neturl.QueryEscape(strings.Join(frontInfo.Waypoints, "|"))
		}

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
				URL: url,
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
			Tolls:       foundTolls,
			Polyline:    route.OverviewPolyline.Points,
			GasStations: nil,
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

	return Response{
		SummaryRoute: summaryRoute,
		Routes:       allRoutes,
	}, nil
}

func (s *Service) findTollsInRoute(ctx context.Context, routes []maps.Route, origin string) ([]Toll, error) {
	var foundTolls []Toll
	uniquePoints := make(map[string]bool)
	uniqueTolls := make(map[int64]bool)

	tolls, err := s.InterfaceService.GetTollsByLonAndLat(ctx)
	if err != nil {
		return nil, err
	}

	var consolidatedPoints []maps.LatLng
	for _, route := range routes {
		for _, leg := range route.Legs {
			for _, step := range leg.Steps {
				pointKey := fmt.Sprintf("%f,%f", step.StartLocation.Lat, step.StartLocation.Lng)
				if !uniquePoints[pointKey] {
					uniquePoints[pointKey] = true
					consolidatedPoints = append(consolidatedPoints, step.StartLocation)
				}
			}
		}
	}

	for _, point := range consolidatedPoints {
		for _, dbToll := range tolls {
			latitude, latErr := parseNullStringToFloat(dbToll.Latitude)
			longitude, lonErr := parseNullStringToFloat(dbToll.Longitude)
			if latErr != nil || lonErr != nil {
				continue
			}

			if isNearby(point.Lat, point.Lng, latitude, longitude, 5.0) {
				if !uniqueTolls[dbToll.ID] {
					uniqueTolls[dbToll.ID] = true

					foundTolls = append(foundTolls, Toll{
						ID:        int(dbToll.ID),
						Latitude:  latitude,
						Longitude: longitude,
						Name:      getStringFromNull(dbToll.PracaDePedagio),
						Road:      getStringFromNull(dbToll.Rodovia),
						State:     getStringFromNull(dbToll.Uf),
						CashCost:  dbToll.Tarifa.Float64,
					})
				}
			}
		}
	}

	return foundTolls, nil
}

//func (s *Service) findTollsInRoute(routes []maps.Route, ctx context.Context, origin string) []Toll {
//	var foundTolls []Toll
//	uniqueTolls := make(map[int64]bool)
//	tolls, err := s.InterfaceService.GetTollsByLonAndLat(ctx)
//	if err != nil {
//		return foundTolls
//	}
//
//	for _, route := range routes {
//		for _, leg := range route.Legs {
//			for _, step := range leg.Steps {
//				points := decodePolyline(step.Polyline.Points)
//				for _, point := range points {
//					for _, dbToll := range tolls {
//						latitude, latErr := parseNullStringToFloat(dbToll.Latitude)
//						longitude, lonErr := parseNullStringToFloat(dbToll.Longitude)
//						if latErr != nil || lonErr != nil {
//							continue
//						}
//
//						if isNearby(point.Lat, point.Lng, latitude, longitude, 5.0) {
//							if !uniqueTolls[dbToll.ID] {
//								uniqueTolls[dbToll.ID] = true
//
//								arrivalTimes, _ := s.time(ctx, origin, fmt.Sprintf("%f,%f", latitude, longitude))
//
//								formattedTime := arrivalTimes.Time.Round(time.Second).String()
//
//								foundTolls = append(foundTolls, Toll{
//									ID:              int(dbToll.ID),
//									Latitude:        latitude,
//									Longitude:       longitude,
//									Name:            getStringFromNull(dbToll.PracaDePedagio),
//									Road:            getStringFromNull(dbToll.Rodovia),
//									State:           getStringFromNull(dbToll.Uf),
//									Country:         "Brasil",
//									Type:            "Pedágio",
//									TagCost:         dbToll.Tarifa.Float64,
//									CashCost:        dbToll.Tarifa.Float64,
//									Currency:        "BRL",
//									PrepaidCardCost: dbToll.Tarifa.Float64,
//									ArrivalResponse: ArrivalResponse{
//										Distance: arrivalTimes.Distance,
//										Time:     formattedTime,
//									},
//									TagPrimary: []string{"Sem Parar", "ConectCar", "Veloe", "Move Mais", "Taggy"},
//								})
//							}
//						}
//					}
//				}
//			}
//		}
//	}
//	return foundTolls
//}

func (s *Service) time(ctx context.Context, origin, destination string) (Arrival, error) {
	apiKey := "AIzaSyAbzUZuCp1zcNNkzje_kmwVVqyOI5w8jkQ"

	client, err := maps.NewClient(maps.WithAPIKey(apiKey))
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
	return Arrival{
		Distance: leg.Distance.HumanReadable,
		Time:     leg.Duration,
	}, nil
}

//func (s *Service) findGasStationsAlongAllRoutes(ctx context.Context, client *maps.Client, routes []maps.Route) ([]GasStation, error) {
//	var gasStations []GasStation
//	uniquePoints := make(map[string]bool)
//	uniqueGasStations := make(map[string]bool)
//
//	// Extrair pontos principais de todas as rotas
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
//	// Fazer chamadas ao Places API para grupos de pontos
//	var wg sync.WaitGroup
//	var mu sync.Mutex
//	chunkSize := 5 // Tamanho do grupo para evitar excesso de requisições
//	for i := 0; i < len(consolidatedPoints); i += chunkSize {
//		wg.Add(1)
//		go func(points []maps.LatLng) {
//			defer wg.Done()
//
//			for _, point := range points {
//				placesRequest := &maps.NearbySearchRequest{
//					Location: &maps.LatLng{Lat: point.Lat, Lng: point.Lng},
//					Radius:   10000, // Ampliar o raio para evitar duplicação
//					Type:     "gas_station",
//				}
//				placesResponse, err := client.NearbySearch(ctx, placesRequest)
//				if err != nil {
//					continue
//				}
//
//				mu.Lock()
//				for _, result := range placesResponse.Results {
//					if !uniqueGasStations[result.Name] {
//						uniqueGasStations[result.Name] = true
//						gasStations = append(gasStations, GasStation{
//							Name:     result.Name,
//							Address:  result.Vicinity,
//							Location: Location{Latitude: result.Geometry.Location.Lat, Longitude: result.Geometry.Location.Lng},
//						})
//					}
//				}
//				mu.Unlock()
//			}
//		}(consolidatedPoints[i:min(i+chunkSize, len(consolidatedPoints))])
//	}
//
//	wg.Wait()
//	return gasStations, nil
//}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func getGeocodeAddress(ctx context.Context, address string) (string, error) {
	apiKey := "AIzaSyAbzUZuCp1zcNNkzje_kmwVVqyOI5w8jkQ"

	client, err := maps.NewClient(maps.WithAPIKey(apiKey))
	if err != nil {
		return "", fmt.Errorf("erro ao criar cliente Google Maps: %v", err)
	}

	if strings.ToLower(address) == "bahia" {
		address = "Salavador, Bahia"
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
		return "", fmt.Errorf("Endereço não encontrado para: %s. Verifique se a pesquisa está escrita corretamente ou seja mais específico(Como: %s, são paulo). Tente adicionar uma, cidade, um estado ou um CEP.", address, address)
	}

	return results[0].FormattedAddress, nil
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
