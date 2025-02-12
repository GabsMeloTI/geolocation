package routes

import (
	"fmt"
	"googlemaps.github.io/maps"
	"math"
	"strings"
)

func DecodePolyline(encoded string) []maps.LatLng {
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

func StateToCapital(address string) string {
	state := strings.ToLower(strings.TrimSpace(address))

	if strings.Contains(state, ",") {
		parts := strings.Split(state, ",")
		state = strings.TrimSpace(parts[0])
	}

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
		return "Praça da sé, São Paulo"
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

func IsNearby(lat1, lng1, lat2, lng2, radius float64) bool {
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

func RecalculateCosts(response Response, frontInfo FrontInfo) Response {
	for i := range response.Routes {
		route := &response.Routes[i]
		totalDistance := float64(route.Summary.Distance.Value) / 1000

		fuelCostCity := math.Round((frontInfo.Price / frontInfo.ConsumptionCity) * totalDistance)
		fuelCostHwy := math.Round((frontInfo.Price / frontInfo.ConsumptionHwy) * totalDistance)

		route.Costs.FuelInTheCity = fuelCostCity
		route.Costs.FuelInTheHwy = fuelCostHwy
		route.Costs.Axles = frontInfo.Axles

		response.SummaryRoute.FuelEfficiency.Hwy = frontInfo.ConsumptionHwy
		response.SummaryRoute.FuelEfficiency.City = frontInfo.ConsumptionCity
		response.SummaryRoute.FuelPrice.Value = frontInfo.Price
	}

	return response
}

func PriceTollsFromVehicle(vehicle string, price, axes float64) (float64, error) {
	var calculation float64
	switch os := vehicle; os {
	case "motorcycle":
		calculation = price / 2
		return calculation, nil
	case "auto":
		if int(axes)%2 != 0 {
			price = price / 2
		}
		calculation = price
		return calculation, nil
	case "bus":
		if int(axes)%2 != 0 {
			price = price / 2
		}
		calculation = price * axes
		return calculation, nil
	case "truck":
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

func RoundCoord(coord float64) float64 {
	return math.Round(coord*1000) / 1000
}

func selectFastestRoute(routes []Route) Route {
	fastest := routes[0]
	for _, r := range routes {
		if r.Summary.Duration.Value < fastest.Summary.Duration.Value {
			fastest = r
		}
	}
	return fastest
}
func selectCheapestRoute(routes []Route) Route {
	cheapest := routes[0]
	for _, r := range routes {
		custoTotal := r.Costs.TagAndCash + r.Costs.FuelInTheCity
		custoCheapest := cheapest.Costs.TagAndCash + cheapest.Costs.FuelInTheCity
		if custoTotal < custoCheapest {
			cheapest = r
		}
	}
	return cheapest
}
func selectEfficientRoute(routes []Route, alpha float64) Route {
	var maxTime, minTime float64
	var maxCost, minCost float64

	if len(routes) == 0 {
		return Route{}
	}
	maxTime = routes[0].Summary.Duration.Value
	minTime = routes[0].Summary.Duration.Value
	initialCost := routes[0].Costs.TagAndCash + routes[0].Costs.FuelInTheCity
	maxCost = initialCost
	minCost = initialCost

	for _, r := range routes {
		timeVal := r.Summary.Duration.Value
		costVal := r.Costs.TagAndCash + r.Costs.FuelInTheCity
		if timeVal > maxTime {
			maxTime = timeVal
		}
		if timeVal < minTime {
			minTime = timeVal
		}
		if costVal > maxCost {
			maxCost = costVal
		}
		if costVal < minCost {
			minCost = costVal
		}
	}

	bestScore := math.MaxFloat64
	var bestRoute Route

	for _, r := range routes {
		timeVal := r.Summary.Duration.Value
		costVal := r.Costs.TagAndCash + r.Costs.FuelInTheCity

		var normalizedTime float64
		var normalizedCost float64

		if maxTime != minTime {
			normalizedTime = (timeVal - minTime) / (maxTime - minTime)
		} else {
			normalizedTime = 0
		}
		if maxCost != minCost {
			normalizedCost = (costVal - minCost) / (maxCost - minCost)
		} else {
			normalizedCost = 0
		}

		score := alpha*normalizedTime + (1-alpha)*normalizedCost
		if score < bestScore {
			bestScore = score
			bestRoute = r
		}
	}

	return bestRoute
}

func CalcFreight(axles string, km, price float64) float64 {
	switch axles {
	case "two_axes":
		return km * price
	case "three_axes":
		return km * price
	case "four_axes":
		return km * price
	case "five_axes":
		return km * price
	case "six_axes":
		return km * price
	case "seven_axes":
		return km * price
	case "eight_axes":
		return km * price
	case "nine_axes":
		return km * price
	default:
		return 0.00
	}
}
