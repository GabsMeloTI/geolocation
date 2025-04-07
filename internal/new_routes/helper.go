package new_routes

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	neturl "net/url"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

func geocodeAddress(address string) (AddressInfo, error) {
	baseURL := "https://nominatim.openstreetmap.org/search"
	params := url.Values{}
	params.Set("q", address)
	params.Set("format", "json")
	params.Set("limit", "1")
	fullURL := baseURL + "?" + params.Encode()

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return AddressInfo{}, err
	}
	req.Header.Set("User-Agent", "GoGeocoder/1.0")
	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return AddressInfo{}, err
	}
	defer resp.Body.Close()

	var results []NominatimResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return AddressInfo{}, err
	}
	if len(results) == 0 {
		return AddressInfo{}, fmt.Errorf("nenhum resultado para o endereço: %s", address)
	}
	lat, err := strconv.ParseFloat(results[0].Lat, 64)
	if err != nil {
		return AddressInfo{}, err
	}
	lon, err := strconv.ParseFloat(results[0].Lon, 64)
	if err != nil {
		return AddressInfo{}, err
	}
	return AddressInfo{
		Location: Location{Latitude: lat, Longitude: lon},
		Address:  results[0].DisplayName,
	}, nil
}

func formatDuration(seconds float64) (string, float64) {
	dur := time.Duration(seconds * float64(time.Second))
	h := int(dur.Hours())
	m := int(dur.Minutes()) % 60
	s := int(dur.Seconds()) % 60
	text := fmt.Sprintf("%dh%dm%ds", h, m, s)
	return text, seconds
}

func formatDistance(meters float64) (string, float64) {
	km := meters / 1000
	return fmt.Sprintf("%.0f km", km), meters
}

func selectImage(instruction string) string {
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
	return valueImg
}

func buildGoogleURL(origin, destination string, waypoints []string) string {
	googleURL := fmt.Sprintf("https://www.google.com/maps/dir/?api=1&origin=%s&destination=%s",
		neturl.QueryEscape(origin),
		neturl.QueryEscape(destination))
	if len(waypoints) > 0 {
		googleURL += "&waypoints=" + neturl.QueryEscape(strings.Join(waypoints, "|"))
	}

	return googleURL
}

func buildWazeURL(origin, destination string, lastLeg time.Duration) string {
	currentTimeMillis := (time.Now().UnixNano() + lastLeg.Nanoseconds()) / int64(time.Millisecond)
	wazeURL := fmt.Sprintf(
		"https://www.waze.com/pt-BR/live-map/directions/br?to=place.%s&from=place.%s&time=%d&reverse=yes",
		destination,
		origin,
		currentTimeMillis,
	)
	return wazeURL
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[size:]
}

func translateInstruction(step OSRMStep) string {
	typ := strings.ToLower(step.Maneuver.Type)
	modifier := strings.ToLower(step.Maneuver.Modifier)
	street := strings.TrimSpace(step.Name)

	switch typ {
	case "depart":
		if street != "" {
			return fmt.Sprintf("Inicie sua viagem na %s", street)
		}
		return "Inicie sua viagem"
	case "turn":
		switch modifier {
		case "left":
			if street != "" {
				return fmt.Sprintf("Vire à esquerda na %s", street)
			}
			return "Vire à esquerda"
		case "right":
			if street != "" {
				return fmt.Sprintf("Vire à direita na %s", street)
			}
			return "Vire à direita"
		case "sharp left":
			if street != "" {
				return fmt.Sprintf("Vire fortemente à esquerda na %s", street)
			}
			return "Vire fortemente à esquerda"
		case "sharp right":
			if street != "" {
				return fmt.Sprintf("Vire fortemente à direita na %s", street)
			}
			return "Vire fortemente à direita"
		case "slight left":
			if street != "" {
				return fmt.Sprintf("Vire suavemente à esquerda na %s", street)
			}
			return "Vire suavemente à esquerda"
		case "slight right":
			if street != "" {
				return fmt.Sprintf("Vire suavemente à direita na %s", street)
			}
			return "Vire suavemente à direita"
		case "Rotary":
			if street != "" {
				return fmt.Sprintf("Rotatória %s", street)
			}
			return "Rotatória"
		case "Exit":
			if street != "" {
				return fmt.Sprintf("Exit %s", street)
			}
			return "Exit"
		default:
			if street != "" {
				return fmt.Sprintf("Vire na direção de %s", street)
			}
			return "Vire"
		}
	case "new name":
		if street != "" {
			return fmt.Sprintf("Continue na %s", street)
		}
		return "Continue em frente"
	case "roundabout":
		if street != "" {
			return fmt.Sprintf("Na rotatória, pegue a primeira saída para a %s", street)
		}
		return "Na rotatória, pegue a primeira saída"
	case "exit roundabout":
		if street != "" {
			return fmt.Sprintf("Saia da rotatória em direção à %s", street)
		}
		return "Saia da rotatória"
	case "end of road":
		if street != "" {
			return fmt.Sprintf("No final da estrada, siga para a %s", street)
		}
		return "No final da estrada, siga em frente"
	case "fork":
		if street != "" {
			return fmt.Sprintf("Na bifurcação, siga em direção à %s", street)
		}
		return "Na bifurcação, siga em frente"
	case "on ramp":
		if street != "" {
			return fmt.Sprintf("Pegue a rampa de entrada para a %s", street)
		}
		return "Pegue a rampa de entrada"
	case "off ramp":
		if street != "" {
			return fmt.Sprintf("Pegue a rampa de saída para a %s", street)
		}
		return "Pegue a rampa de saída"
	case "merge":
		if street != "" {
			return fmt.Sprintf("Faça a fusão para a %s", street)
		}
		return "Faça a fusão com a via"
	case "arrive":
		if street != "" {
			return fmt.Sprintf("Chegue à %s", street)
		}
		return "Você chegou ao destino"
	case "continue":
		if street != "" {
			return fmt.Sprintf("Continue na %s", street)
		}
		return "Continue em frente"
	default:
		if street != "" {
			return fmt.Sprintf("%s na %s", capitalize(typ), street)
		}
		return capitalize(typ)
	}
}

func decodePolyline(encoded string) ([]LatLng, error) {
	var points []LatLng
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
		points = append(points, LatLng{
			Lat: float64(lat) / 1e5,
			Lng: float64(lng) / 1e5,
		})
	}
	return points, nil
}

func distancePointToSegment(p, v, w LatLng) float64 {
	const latFactor = 111320.0
	lngFactor := 111320.0 * math.Cos(v.Lat*math.Pi/180)

	dx := (w.Lng - v.Lng) * lngFactor
	dy := (w.Lat - v.Lat) * latFactor

	dxp := (p.Lng - v.Lng) * lngFactor
	dyp := (p.Lat - v.Lat) * latFactor

	segLenSq := dx*dx + dy*dy
	if segLenSq == 0 {
		return math.Sqrt(dxp*dxp + dyp*dyp)
	}

	dot := dxp*dx + dyp*dy
	t := dot / segLenSq

	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}

	projX := dx * t
	projY := dy * t

	distX := dxp - projX
	distY := dyp - projY

	return math.Sqrt(distX*distX + distY*distY)
}

func queryFuelStations(polyline string) ([]GasStation, error) {
	points, err := decodePolyline(polyline)
	if err != nil {
		return nil, err
	}
	minLat, minLon := points[0].Lat, points[0].Lng
	maxLat, maxLon := points[0].Lat, points[0].Lng
	for _, p := range points[1:] {
		if p.Lat < minLat {
			minLat = p.Lat
		}
		if p.Lng < minLon {
			minLon = p.Lng
		}
		if p.Lat > maxLat {
			maxLat = p.Lat
		}
		if p.Lng > maxLon {
			maxLon = p.Lng
		}
	}
	padding := 1.0
	minLat -= padding
	minLon -= padding
	maxLat += padding
	maxLon += padding

	overpassQuery := fmt.Sprintf(`
		[out:json][timeout:25];
		(
		  node["amenity"="fuel"](%f,%f,%f,%f);
		);
		out body;
		>;
		out skel qt;
	`, minLat, minLon, maxLat, maxLon)

	overpassURL := "http://overpass-api.de/api/interpreter"
	resp, err := http.PostForm(overpassURL, url.Values{"data": {overpassQuery}})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var overpassResp struct {
		Elements []struct {
			ID   int64             `json:"id"`
			Lat  float64           `json:"lat"`
			Lon  float64           `json:"lon"`
			Tags map[string]string `json:"tags"`
		} `json:"elements"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&overpassResp); err != nil {
		return nil, err
	}

	tolerance := 50.0
	var stations []GasStation
	for _, el := range overpassResp.Elements {
		stationPos := LatLng{Lat: el.Lat, Lng: el.Lon}
		minDist := math.MaxFloat64
		for _, p := range points {
			d := math.Hypot(stationPos.Lat-p.Lat, stationPos.Lng-p.Lng) * 111320.0
			if d < minDist {
				minDist = d
			}
		}
		if minDist <= tolerance {
			station := GasStation{
				Name:    el.Tags["name"],
				Address: el.Tags["addr:full"],
				Location: Location{
					Latitude:  el.Lat,
					Longitude: el.Lon,
				},
			}
			stations = append(stations, station)
		}
	}
	return stations, nil
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

func sortByProximity[T any](origin Location, points []T, getLocation func(T) Location) []T {
	sort.Slice(points, func(i, j int) bool {
		locI := getLocation(points[i])
		locJ := getLocation(points[j])
		distI := haversineDistance(origin, locI)
		distJ := haversineDistance(origin, locJ)
		return distI < distJ
	})
	return points
}

func haversineDistance(loc1, loc2 Location) float64 {
	const R = 6371
	deltaLat := (loc2.Latitude - loc1.Latitude) * (math.Pi / 180)
	deltaLon := (loc2.Longitude - loc1.Longitude) * (math.Pi / 180)
	lat1 := loc1.Latitude * (math.Pi / 180)
	lat2 := loc2.Latitude * (math.Pi / 180)
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Sin(deltaLon/2)*math.Sin(deltaLon/2)*math.Cos(lat1)*math.Cos(lat2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

func haversineDistanceTolls(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371000
	φ1 := lat1 * math.Pi / 180
	φ2 := lat2 * math.Pi / 180
	Δφ := (lat2 - lat1) * math.Pi / 180
	Δλ := (lng2 - lng1) * math.Pi / 180

	a := math.Sin(Δφ/2)*math.Sin(Δφ/2) +
		math.Cos(φ1)*math.Cos(φ2)*math.Sin(Δλ/2)*math.Sin(Δλ/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

func isAllRouteOptionsDisabled(options RouteOptions) bool {
	return !options.IncludeFuelStations &&
		!options.IncludeRouteMap &&
		!options.IncludeTollCosts &&
		!options.IncludeWeighStations &&
		!options.IncludeFreightCalc &&
		!options.IncludePolyline
}
