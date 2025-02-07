package routes

import (
	"context"
	"database/sql"
	"encoding/json"
	_ "encoding/json"
	"errors"
	"fmt"
	db "geolocation/db/sqlc"
	cache "geolocation/pkg"
	"geolocation/validation"
	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	"googlemaps.github.io/maps"
	"html"
	"math"
	neturl "net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type InterfaceService interface {
	CheckRouteTolls(ctx context.Context, frontInfo FrontInfo) (Response, error)
}

type Service struct {
	InterfaceService InterfaceRepository
	GoogleMapsAPIKey string
}

func NewRoutesService(InterfaceService InterfaceRepository, GoogleMapsAPIKey string) *Service {
	return &Service{InterfaceService, GoogleMapsAPIKey}
}

func (s *Service) CheckRouteTolls(ctx context.Context, frontInfo FrontInfo) (Response, error) {
	cacheKey := fmt.Sprintf("route:%s:%s:%s", frontInfo.Origin, frontInfo.Destination, strings.Join(frontInfo.Waypoints, ","))

	cached, err := cache.Rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		var cachedResponse Response
		if json.Unmarshal([]byte(cached), &cachedResponse) == nil {
			return RecalculateCosts(cachedResponse, frontInfo), nil
		}
	} else if !errors.Is(err, redis.Nil) {
		return Response{}, err
	}

	savedRoute, err := s.InterfaceService.GetSavedRoutes(ctx, db.GetSavedRoutesParams{
		Origin:      frontInfo.Origin,
		Destination: frontInfo.Destination,
		Waypoints: sql.NullString{
			String: strings.Join(frontInfo.Waypoints, ","),
			Valid:  true,
		},
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
		foundTolls, _ := s.findTollsInRoute(ctx, []maps.Route{route}, origin.FormattedAddress, frontInfo.Type, float64(frontInfo.Axles))
		foundBalancas, _ := s.findBalancaInRoute(ctx, []maps.Route{route})
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
				FuelInTheCity:   fuelCostCity,
				FuelInTheHwy:    fuelCostHwy,
				Tag:             totalTollCost,
				Cash:            totalTollCost,
				PrepaidCard:     totalTollCost,
				MaximumTollCost: maxTollCost,
				MinimumTollCost: minTollCost,
			},
			Tolls:        foundTolls,
			Balanca:      foundBalancas,
			GasStations:  findGasStationsAlongRoute,
			Polyline:     route.OverviewPolyline.Points,
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

	selectedRoute := SelectBestRoute(allRoutes, frontInfo.TypeRoute)
	response := Response{
		SummaryRoute: summaryRoute,
		Routes:       []Route{selectedRoute},
	}

	responseJSON, _ := json.Marshal(response)
	requestJSON, _ := json.Marshal(routeRequest)

	if err := cache.Rdb.Set(cache.Ctx, cacheKey, responseJSON, 30*24*time.Hour).Err(); err != nil {
		fmt.Printf("Erro ao salvar cache do Redis (CheckRouteTolls): %v\n", err)
		return Response{}, errors.New("Erro ao salvar cache do Redis")
	}

	waypoints := strings.Join(frontInfo.Waypoints, ",")
	_, err = s.InterfaceService.CreateSavedRoutes(ctx, db.CreateSavedRoutesParams{
		Origin:      frontInfo.Origin,
		Destination: frontInfo.Destination,
		Waypoints: sql.NullString{
			String: waypoints,
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
		if err := json.Unmarshal([]byte(cached), &arrival); err == nil {
			return arrival, nil
		}
	} else if err != redis.Nil {
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

func (s *Service) findTollsInRoute(ctx context.Context, routes []maps.Route, origin, vehicle string, axes float64) ([]Toll, error) {
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
				polyPoints := DecodePolyline(step.Polyline.Points)
				for _, point := range polyPoints {
					key := fmt.Sprintf("%f,%f", RoundCoord(point.Lat), RoundCoord(point.Lng))
					uniquePoints[key] = point
				}
			}
		}
	}

	for _, point := range uniquePoints {
		for _, dbToll := range tolls {
			latitude, latErr := validation.ParseNullStringToFloat(dbToll.Latitude)
			longitude, lonErr := validation.ParseNullStringToFloat(dbToll.Longitude)
			if latErr != nil || lonErr != nil {
				continue
			}

			if IsNearby(point.Lat, point.Lng, latitude, longitude, 0.5) {
				if !uniqueTolls[dbToll.ID] {
					uniqueTolls[dbToll.ID] = true

					dest := fmt.Sprintf("%.6f,%.6f", latitude, longitude)
					arrivalTimes, err := s.calculateTimeToToll(ctx, origin, dest)
					if err != nil {
						fmt.Printf("Erro ao obter tempo para origem %s e destino %s: %v\n", origin, dest, err)
						continue
					}
					formattedTime := arrivalTimes.Time.Round(time.Second).String()

					concession := validation.GetStringFromNull(dbToll.Concessionaria)
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

					tarifaFloat, err := strconv.ParseFloat(dbToll.Tarifa, 64)

					totalToll, err := PriceTollsFromVehicle(strings.ToLower(vehicle), tarifaFloat, axes)
					if err != nil {
						return nil, err
					}

					foundTolls = append(foundTolls, Toll{
						ID:              int(dbToll.ID),
						Latitude:        latitude,
						Longitude:       longitude,
						Name:            validation.GetStringFromNull(dbToll.PracaDePedagio),
						Concession:      dbToll.Concessionaria.String,
						Road:            validation.GetStringFromNull(dbToll.Rodovia),
						State:           validation.GetStringFromNull(dbToll.Uf),
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

func (s *Service) findGasStationsAlongAllRoutes(ctx context.Context, client *maps.Client, routes []maps.Route) ([]GasStation, error) {
	var gasStations []GasStation
	uniqueGasStations := make(map[string]bool)
	var consolidatedPoints []maps.LatLng
	uniquePoints := make(map[string]bool)

	for _, route := range routes {
		for _, leg := range route.Legs {
			for _, step := range leg.Steps {
				pointKey := GetCacheKeyForPoint(step.StartLocation.Lat, step.StartLocation.Lng)
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
				cacheKey := fmt.Sprintf("gasStations:%s", GetCacheKeyForPoint(point.Lat, point.Lng))
				cached, err := cache.Rdb.Get(cache.Ctx, cacheKey).Result()
				if err == nil {
					var cachedStations []GasStation
					if err := json.Unmarshal([]byte(cached), &cachedStations); err == nil {
						mu.Lock()
						for _, gs := range cachedStations {
							if !uniqueGasStations[gs.Address] {
								uniqueGasStations[gs.Address] = true
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
				var cachedResult []GasStation
				if err == nil && len(dbGasStations) > 0 {
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
					if len(cachedResult) > 0 {
						data, err := json.Marshal(cachedResult)
						if err == nil {
							if err := cache.Rdb.Set(cache.Ctx, cacheKey, data, 30*24*time.Hour).Err(); err != nil {
								fmt.Printf("Erro ao salvar cache Redis para gasStations: %v\n", err)
							}
						}
					}
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
				if len(cachedResult) > 0 {
					data, err := json.Marshal(cachedResult)
					if err == nil {
						if err := cache.Rdb.Set(cache.Ctx, cacheKey, data, 30*24*time.Hour).Err(); err != nil {
							fmt.Printf("Erro ao salvar cache Redis para gasStations: %v\n", err)
						}
					}
				}
			}
		}(consolidatedPoints[i:end])
	}

	wg.Wait()
	return gasStations, nil
}

func (s *Service) calculateTimeToToll(ctx context.Context, origin, destination string) (Arrival, error) {
	cacheKey := fmt.Sprintf("time:%s|%s", origin, destination)

	cached, err := cache.Rdb.Get(cache.Ctx, cacheKey).Result()
	if err == nil {
		var arrival Arrival
		if err := json.Unmarshal([]byte(cached), &arrival); err == nil {
			return arrival, nil
		}
	} else if err != redis.Nil {
		fmt.Printf("Erro ao recuperar cache do Redis: %v\n", err)
	}

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
	if len(routes) == 0 || len(routes[0].Legs) == 0 {
		return Arrival{}, echo.ErrNotFound
	}

	leg := routes[0].Legs[0]
	arrival := Arrival{
		Distance: leg.Distance.HumanReadable,
		Time:     leg.Duration,
	}

	data, err := json.Marshal(arrival)
	if err == nil {
		if err := cache.Rdb.Set(cache.Ctx, cacheKey, data, 30*24*time.Hour).Err(); err != nil {
			fmt.Printf("Erro ao salvar cache no Redis: %v\n", err)
		}
	}

	return s.timeWithClient(ctx, client, origin, destination)
}

func (s *Service) getGeocodeAddress(ctx context.Context, address string) (GeocodeResult, error) {
	address = StateToCapital(strings.ToLower(address))

	cacheKey := fmt.Sprintf("geocode:%s", address)

	cached, err := cache.Rdb.Get(cache.Ctx, cacheKey).Result()
	if err == nil {
		var result GeocodeResult
		if err := json.Unmarshal([]byte(cached), &result); err == nil {
			return result, nil
		}
	} else if err != redis.Nil {
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
		return GeocodeResult{}, fmt.Errorf("Endereço não encontrado para: %s. Verifique se a pesquisa está escrita corretamente ou seja mais específico(Ex: %s, São Paulo).", address, address)
	}

	result := GeocodeResult{
		FormattedAddress: results[0].FormattedAddress,
		PlaceID:          results[0].PlaceID,
	}

	data, err := json.Marshal(result)
	if err == nil {
		if err := cache.Rdb.Set(cache.Ctx, cacheKey, data, 30*24*time.Hour).Err(); err != nil {
			fmt.Printf("Erro ao salvar cache do Redis (geocode): %v\n", err)
		}
	}

	return result, nil
}
