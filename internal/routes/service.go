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
	GetExactPlace(ctx context.Context, placeRequest PlaceRequest) (PlaceResponse, error)
}

type Service struct {
	InterfaceService InterfaceRepository
}

func NewRoutesService(InterfaceService InterfaceRepository) *Service {
	return &Service{InterfaceService}
}

func (s *Service) CheckRouteTolls(ctx context.Context, frontInfo FrontInfo) (Response, error) {
	apiKey := "AIzaSyAvLoyVe2LlazHJfT0Kan5ZyX7dDb0exyQ"

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

	for _, route := range routes {
		foundTolls := s.findTollsInRoute([]maps.Route{route}, ctx, frontInfo.Origin)

		var totalDistance int
		var totalDuration time.Duration

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
				TagAndCash:      1.0,
				Fuel:            1.0,
				Tag:             1.0,
				Cash:            1.0,
				PrepaidCard:     1.0,
				MaximumTollCost: 1.0,
				MinimumTollCost: 1.0,
			},
			Tolls:    foundTolls,
			Polyline: route.OverviewPolyline.Points,
		})

		summaryRoute = SummaryRoute{
			RouteOrigin: PrincipalRoute{
				Location: Location{
					Latitude:  route.Legs[0].StartLocation.Lat,
					Longitude: route.Legs[0].StartLocation.Lng,
				},
				Address: route.Legs[0].StartAddress,
			},
			RouteDestination: PrincipalRoute{
				Location: Location{
					Latitude:  lastLeg.EndLocation.Lat,
					Longitude: lastLeg.EndLocation.Lng,
				},
				Address: lastLeg.EndAddress,
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

func getGeocodeAddress(ctx context.Context, address string) (string, error) {
	apiKey := "AIzaSyAvLoyVe2LlazHJfT0Kan5ZyX7dDb0exyQ"

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

func (s *Service) time(ctx context.Context, origin, destination string) ([]time.Time, error) {
	apiKey := "AIzaSyAvLoyVe2LlazHJfT0Kan5ZyX7dDb0exyQ"

	client, err := maps.NewClient(maps.WithAPIKey(apiKey))
	if err != nil {
		return []time.Time{}, err
	}

	routeRequest := &maps.DirectionsRequest{
		Origin:       origin,
		Destination:  destination,
		Alternatives: true,
		Mode:         maps.TravelModeDriving,
		Region:       "br",
	}
	routes, _, err := client.Directions(ctx, routeRequest)
	if err != nil {
		return []time.Time{}, err
	}

	if len(routes) == 0 {
		return []time.Time{}, echo.ErrNotFound
	}

	var allRoutes []time.Time
	for _, route := range routes {
		leg := route.Legs[0]

		allRoutes = append(allRoutes, leg.ArrivalTime)
	}

	return allRoutes, nil
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

								arrivalTimes, err := s.time(ctx, origin, fmt.Sprintf("%f,%f", latitude, longitude))
								var arrivalTime time.Time
								if err == nil && len(arrivalTimes) > 0 {
									arrivalTime = arrivalTimes[0]
								} else {
									arrivalTime = time.Now()
								}

								foundTolls = append(foundTolls, Toll{
									ID:              int(dbToll.ID),
									Latitude:        latitude,
									Longitude:       longitude,
									Name:            getStringFromNull(dbToll.PracaDePedagio),
									Road:            getStringFromNull(dbToll.Rodovia),
									State:           getStringFromNull(dbToll.Uf),
									Country:         "Brasil",
									Type:            "Pedágio",
									TagCost:         10,
									CashCost:        10,
									Currency:        "BRL",
									PrepaidCardCost: "1.0",
									Arrival: Arrival{
										Distance: 10,
										Time:     arrivalTime,
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
