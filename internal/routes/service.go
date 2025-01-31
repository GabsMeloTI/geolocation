package routes

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/labstack/echo/v4"
	"googlemaps.github.io/maps"
	"math"
	"strconv"
	"sync"
	"time"
)

var (
	googleMapsClient *maps.Client
	once             sync.Once
	geoCache         = sync.Map{}
)

func initGoogleMapsClient() (*maps.Client, error) {
	var err error
	once.Do(func() {
		googleMapsClient, err = maps.NewClient(maps.WithAPIKey("AIzaSyAvLoyVe2LlazHJfT0Kan5ZyX7dDb0exyQ"))
	})

	return googleMapsClient, err
}

type InterfaceService interface {
	CheckRouteTolls(ctx context.Context, frontInfo FrontInfo) (Response, error)
	GetExactPlace(ctx context.Context, placeRequest PlaceRequest) (PlaceResponse, error)
}

type Service struct {
	InterfaceService InterfaceRepository
}

func NewRoutesService(InterfaceService InterfaceRepository) *Service {
	return &Service{InterfaceService}
}

func getCachedGeocodeAddress(ctx context.Context, address string) (string, error) {
	if cached, found := geoCache.Load(address); found {
		return cached.(string), nil
	}

	client, err := initGoogleMapsClient()
	if err != nil {
		return "", err
	}

	req := &maps.GeocodingRequest{
		Address: address,
		Region:  "br",
	}

	results, err := client.Geocode(ctx, req)
	if err != nil || len(results) == 0 {
		return "", fmt.Errorf("endereço não encontrado para: %s", address)
	}

	formattedAddress := results[0].FormattedAddress
	geoCache.Store(address, formattedAddress)
	return formattedAddress, nil
}

func (s *Service) CheckRouteTolls(ctx context.Context, frontInfo FrontInfo) (Response, error) {
	client, err := initGoogleMapsClient()
	if err != nil {
		return Response{}, err
	}

	origin, err := getCachedGeocodeAddress(ctx, frontInfo.Origin)
	if err != nil {
		return Response{}, err
	}

	destination, err := getCachedGeocodeAddress(ctx, frontInfo.Destination)
	if err != nil {
		return Response{}, err
	}

	routeRequest := &maps.DirectionsRequest{
		Origin:       origin,
		Destination:  destination,
		Mode:         maps.TravelModeDriving,
		Waypoints:    frontInfo.Waypoints,
		Alternatives: false,
		Region:       "br",
	}

	routes, _, err := client.Directions(ctx, routeRequest)
	if err != nil {
		return Response{}, err
	}

	if len(routes) == 0 {
		return Response{}, fmt.Errorf("nenhuma rota encontrada")
	}

	mainRoute := routes[0]

	foundTolls := s.findTollsInRoute([]maps.Route{mainRoute}, ctx, frontInfo.Origin)

	//gasStations, err := s.findGasStationsAlongRoute(ctx, client, mainRoute)

	totalDistance := 0
	totalDuration := time.Duration(0)
	totalTollCost := 0.0

	for _, leg := range mainRoute.Legs {
		totalDistance += leg.Distance.Meters
		totalDuration += leg.Duration
	}

	for _, toll := range foundTolls {
		totalTollCost += toll.CashCost
	}

	return Response{
		SummaryRoute: SummaryRoute{
			RouteOrigin: PrincipalRoute{
				Location: Location{
					Latitude:  mainRoute.Legs[0].StartLocation.Lat,
					Longitude: mainRoute.Legs[0].StartLocation.Lng,
				},
				Address: mainRoute.Legs[0].StartAddress,
			},
			RouteDestination: PrincipalRoute{
				Location: Location{
					Latitude:  mainRoute.Legs[len(mainRoute.Legs)-1].EndLocation.Lat,
					Longitude: mainRoute.Legs[len(mainRoute.Legs)-1].EndLocation.Lng,
				},
				Address: mainRoute.Legs[len(mainRoute.Legs)-1].EndAddress,
			},
		},
		Routes: []Route{
			{
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
				},
				Costs: Costs{
					TagAndCash: totalTollCost,
				},
				Tolls:       foundTolls,
				GasStations: nil,
			},
		},
	}, nil
}

func (s *Service) findTollsInRoute(routes []maps.Route, ctx context.Context, origin string) []Toll {
	var foundTolls []Toll
	uniqueTolls := make(map[int64]bool)
	tolls, err := s.InterfaceService.GetTollsByLonAndLat(ctx)
	if err != nil {
		return foundTolls
	}

	for _, route := range routes {
		for _, leg := range route.Legs {
			for _, step := range leg.Steps {
				points := decodePolyline(step.Polyline.Points)
				for _, point := range points {
					for _, dbToll := range tolls {
						latitude, latErr := parseNullStringToFloat(dbToll.Latitude)
						longitude, lonErr := parseNullStringToFloat(dbToll.Longitude)
						if latErr != nil || lonErr != nil {
							continue
						}

						if isNearby(point.Lat, point.Lng, latitude, longitude, 5.0) {
							if !uniqueTolls[dbToll.ID] {
								uniqueTolls[dbToll.ID] = true

								arrivalTimes, _ := s.time(ctx, origin, fmt.Sprintf("%f,%f", latitude, longitude))

								formattedTime := arrivalTimes.Time.Round(time.Second).String()

								foundTolls = append(foundTolls, Toll{
									ID:              int(dbToll.ID),
									Latitude:        latitude,
									Longitude:       longitude,
									Name:            getStringFromNull(dbToll.PracaDePedagio),
									Road:            getStringFromNull(dbToll.Rodovia),
									State:           getStringFromNull(dbToll.Uf),
									Country:         "Brasil",
									Type:            "Pedágio",
									TagCost:         dbToll.Tarifa.Float64,
									CashCost:        dbToll.Tarifa.Float64,
									Currency:        "BRL",
									PrepaidCardCost: dbToll.Tarifa.Float64,
									ArrivalResponse: ArrivalResponse{
										Distance: arrivalTimes.Distance,
										Time:     formattedTime,
									},
									TagPrimary: []string{"Sem Parar", "ConectCar", "Veloe", "Move Mais", "Taggy"},
								})
							}
						}
					}
				}
			}
		}
	}
	return foundTolls
}

//func (s *Service) findGasStationsAlongRoute(ctx context.Context, client *maps.Client, route maps.Route) ([]GasStation, error) {
//	var gasStations []GasStation
//	uniqueGasStation := make(map[string]bool)
//
//	for _, leg := range route.Legs {
//		for _, step := range leg.Steps {
//			placesRequest := &maps.NearbySearchRequest{
//				Location: &maps.LatLng{Lat: step.StartLocation.Lat, Lng: step.StartLocation.Lng},
//				Radius:   10,
//				Type:     "gas_station",
//				Keyword:  "posto de gasolina",
//			}
//			placesResponse, err := client.NearbySearch(ctx, placesRequest)
//			if err != nil {
//				return nil, err
//			}
//			for _, result := range placesResponse.Results {
//				if !uniqueGasStation[result.Name] {
//					uniqueGasStation[result.Name] = true
//
//					gasStations = append(gasStations, GasStation{
//						Name:     result.Name,
//						Address:  result.Vicinity,
//						Location: Location{Latitude: result.Geometry.Location.Lat, Longitude: result.Geometry.Location.Lng},
//					})
//				}
//			}
//		}
//	}
//	return gasStations, nil
//}

//func (s *Service) findGasStationsAlongRoute(ctx context.Context, client *maps.Client, route maps.Route) ([]GasStation, error) {
//	var gasStations []GasStation
//	var wg sync.WaitGroup
//	var mu sync.Mutex
//	errChan := make(chan error, 1)
//	defer close(errChan)
//
//	leg := route.Legs[0]
//	wg.Add(1)
//	go func(leg maps.Leg) {
//		defer wg.Done()
//		placesRequest := &maps.NearbySearchRequest{
//			Location: &maps.LatLng{Lat: leg.StartLocation.Lat, Lng: leg.StartLocation.Lng},
//			Radius:   10000,
//			Type:     "gas_station",
//			Keyword:  "posto de gasolina",
//		}
//		placesResponse, err := client.NearbySearch(ctx, placesRequest)
//		if err != nil {
//			errChan <- err
//			return
//		}
//		mu.Lock()
//		for _, result := range placesResponse.Results {
//			gasStations = append(gasStations, GasStation{
//				Name:     result.Name,
//				Address:  result.Vicinity,
//				Location: Location{Latitude: result.Geometry.Location.Lat, Longitude: result.Geometry.Location.Lng},
//			})
//		}
//		mu.Unlock()
//	}(maps.Leg{
//		Steps:             leg.Steps,
//		Distance:          leg.Distance,
//		Duration:          leg.Duration,
//		DurationInTraffic: leg.DurationInTraffic,
//		ArrivalTime:       leg.ArrivalTime,
//		DepartureTime:     leg.DepartureTime,
//		StartLocation:     leg.StartLocation,
//		EndLocation:       leg.EndLocation,
//		StartAddress:      leg.StartAddress,
//		EndAddress:        leg.EndAddress,
//		ViaWaypoint:       leg.ViaWaypoint,
//	})
//
//	wg.Wait()
//	select {
//	case err := <-errChan:
//		return nil, err
//	default:
//		return gasStations, nil
//	}
//}

//func (s *Service) findGasStationsAlongRoute(ctx context.Context, client *maps.Client, route maps.Route) ([]GasStation, error) {
//	var gasStations []GasStation
//	uniqueLocations := make(map[string]bool)
//
//	for _, leg := range route.Legs {
//		for _, step := range leg.Steps {
//			locationKey := fmt.Sprintf("%f,%f", step.StartLocation.Lat, step.StartLocation.Lng)
//			if uniqueLocations[locationKey] {
//				continue
//			}
//
//			uniqueLocations[locationKey] = true
//
//			placesRequest := &maps.NearbySearchRequest{
//				Location: &maps.LatLng{
//					Lat: step.StartLocation.Lat,
//					Lng: step.StartLocation.Lng,
//				},
//				Radius: 5000,
//				Type:   "gas_station",
//			}
//
//			placesResponse, err := client.NearbySearch(ctx, placesRequest)
//			if err != nil {
//				return nil, err
//			}
//
//			for _, result := range placesResponse.Results {
//				gasStations = append(gasStations, GasStation{
//					Name:     result.Name,
//					Address:  result.Vicinity,
//					Location: Location{Latitude: result.Geometry.Location.Lat, Longitude: result.Geometry.Location.Lng},
//				})
//			}
//		}
//	}
//	return gasStations, nil
//}

func (s *Service) time(ctx context.Context, origin, destination string) (Arrival, error) {
	client, err := initGoogleMapsClient()
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

func (s *Service) GetExactPlace(ctx context.Context, placeRequest PlaceRequest) (PlaceResponse, error) {
	apiKey := "AIzaSyAvLoyVe2LlazHJfT0Kan5ZyX7dDb0exyQ"

	client, err := maps.NewClient(maps.WithAPIKey(apiKey))
	if err != nil {
		return PlaceResponse{}, err
	}

	placeSearchRequest := &maps.NearbySearchRequest{
		Location: &maps.LatLng{
			Lat: placeRequest.Latitude,
			Lng: placeRequest.Longitude,
		},
		Radius: 100,
	}

	searchResults, err := client.NearbySearch(ctx, placeSearchRequest)
	if err != nil {
		return PlaceResponse{}, err
	}

	if len(searchResults.Results) == 0 {
		return PlaceResponse{}, nil
	}

	return PlaceResponse{Place: searchResults.Results[0]}, nil
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
