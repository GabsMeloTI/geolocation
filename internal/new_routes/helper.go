package routes

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
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
	req.Header.Set("User-Agent", "GoGeocoder/1.0 (seuemail@exemplo.com)")
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

func selectImage(modifier string) string {
	switch strings.ToLower(modifier) {
	case "left":
		return "https://plates-routes.s3.us-east-1.amazonaws.com/esquerda.png"
	case "right":
		return "https://plates-routes.s3.us-east-1.amazonaws.com/direita.png"
	default:
		return "https://plates-routes.s3.us-east-1.amazonaws.com/reto.png"
	}
}

func buildGoogleURL(origin, destination string) string {
	return "https://www.google.com/maps/dir/?api=1&origin=" +
		url.QueryEscape(origin) + "&destination=" + url.QueryEscape(destination)
}

func buildWazeURL(origin, destination string) string {
	return "https://www.waze.com/ul?ll=" + url.QueryEscape(destination) +
		"&from=" + url.QueryEscape(origin)
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

func getTollsDB(db *sql.DB) ([]Toll, error) {
	rows, err := db.Query(`SELECT * FROM tolls`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tolls []Toll
	for rows.Next() {
		var t Toll
		var concessionaria, praca, rodovia, uf, km_m, municipio, tipoPista, sentido, situacao, dataInativacao, lat, lon, payFreeFlow sql.NullString
		var ano sql.NullInt64
		var tarifa sql.NullFloat64
		var freeFlow sql.NullBool

		err := rows.Scan(
			&t.ID,
			&concessionaria,
			&praca,
			&ano,
			&rodovia,
			&uf,
			&km_m,
			&municipio,
			&tipoPista,
			&sentido,
			&situacao,
			&dataInativacao,
			&lat,
			&lon,
			&tarifa,
			&freeFlow,
			&payFreeFlow,
		)
		if err != nil {
			return nil, err
		}

		if concessionaria.Valid {
			t.Concessionaria = concessionaria.String
		}
		if praca.Valid {
			t.PracaDePedagio = praca.String
		}
		if ano.Valid {
			t.AnoDoPNVSNV = ano.Int64
		}
		if rodovia.Valid {
			t.Rodovia = rodovia.String
		}
		if uf.Valid {
			t.UF = uf.String
		}
		if km_m.Valid {
			t.KmM = km_m.String
		}
		if municipio.Valid {
			t.Municipio = municipio.String
		}
		if tipoPista.Valid {
			t.TipoPista = tipoPista.String
		}
		if sentido.Valid {
			t.Sentido = sentido.String
		}
		if situacao.Valid {
			t.Situacao = situacao.String
		}
		if dataInativacao.Valid {
			t.DataDaInativacao = dataInativacao.String
		}
		if lat.Valid {
			t.Latitude, _ = strconv.ParseFloat(lat.String, 64)
		}
		if lon.Valid {
			t.Longitude, _ = strconv.ParseFloat(lon.String, 64)
		}
		if tarifa.Valid {
			t.Tarifa = tarifa.Float64
		}
		if freeFlow.Valid {
			t.FreeFlow = freeFlow.Bool
		}
		if payFreeFlow.Valid {
			t.PayFreeFlow = payFreeFlow.String
		}

		tolls = append(tolls, t)
	}
	return tolls, nil
}

func getBalancaDB(db *sql.DB) ([]Balanca, error) {
	rows, err := db.Query(`SELECT * FROM balanca`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var balancas []Balanca
	for rows.Next() {
		var b Balanca
		var concessionaria, km, lat, lng, nome, rodovia, sentido, uf sql.NullString

		err := rows.Scan(
			&b.ID,
			&concessionaria,
			&km,
			&lat,
			&lng,
			&nome,
			&rodovia,
			&sentido,
			&uf,
		)
		if err != nil {
			return nil, err
		}

		if concessionaria.Valid {
			b.Concessionaria = concessionaria.String
		}
		if nome.Valid {
			b.Nome = nome.String
		}
		if km.Valid {
			b.Km = km.String
		}
		if lat.Valid {
			b.Lat, _ = strconv.ParseFloat(lat.String, 64)
		}
		if lng.Valid {
			b.Lng, _ = strconv.ParseFloat(lng.String, 64)
		}
		if rodovia.Valid {
			b.Rodovia = rodovia.String
		}
		if sentido.Valid {
			b.Sentido = sentido.String
		}
		if uf.Valid {
			b.Uf = uf.String
		}
		balancas = append(balancas, b)
	}
	return balancas, nil
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

func findTollsOnRoute(routeGeometry string, tolls []Toll) ([]Toll, error) {
	var foundTolls []Toll

	polyPoints, err := decodePolyline(routeGeometry)
	if err != nil {
		return nil, err
	}
	if len(polyPoints) < 2 {
		return foundTolls, nil
	}

	routeIsCrescente := polyPoints[0].Lat < polyPoints[len(polyPoints)-1].Lat

	for _, toll := range tolls {
		tollPos := LatLng{Lat: toll.Latitude, Lng: toll.Longitude}
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

		if toll.Sentido != "" {
			if toll.Sentido == "Crescente" && !routeIsCrescente {
				continue
			}
			if toll.Sentido == "Decrescente" && routeIsCrescente {
				continue
			}
		}

		foundTolls = append(foundTolls, toll)
	}

	return foundTolls, nil
}

func findBalancaOnRoute(routeGeometry string, balancas []Balanca) ([]Balanca, error) {
	var foundBalancas []Balanca

	polyPoints, err := decodePolyline(routeGeometry)
	if err != nil {
		return nil, err
	}
	if len(polyPoints) < 2 {
		return foundBalancas, nil
	}

	routeIsCrescente := polyPoints[0].Lat < polyPoints[len(polyPoints)-1].Lat

	for _, b := range balancas {
		pos := LatLng{Lat: b.Lat, Lng: b.Lng}
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

func getAllFreight(axles int, kmValue float64, db *sql.DB) (map[string][]FreightLoad, error) {
	grouped := make(map[string][]FreightLoad)

	rows, err := db.Query("SELECT type_of_load, two_axes, three_axes, four_axes, five_axes, six_axes, seven_axes, nine_axes, name, description FROM freight_load")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var fl FreightLoad
		var twoAxes, threeAxes, fourAxes, fiveAxes, sixAxes, sevenAxes, nineAxes, name, description string
		err := rows.Scan(&fl.TypeOfLoad, &twoAxes, &threeAxes, &fourAxes, &fiveAxes, &sixAxes, &sevenAxes, &nineAxes, &name, &description)
		if err != nil {
			continue
		}
		var rateStr string
		switch axles {
		case 2:
			rateStr = twoAxes
		case 3:
			rateStr = threeAxes
		case 4:
			rateStr = fourAxes
		case 5:
			rateStr = fiveAxes
		case 6:
			rateStr = sixAxes
		case 7:
			rateStr = sevenAxes
		case 8:
			rateStr = sevenAxes
		case 9:
			rateStr = nineAxes
		default:
			rateStr = twoAxes
		}
		rateStr = strings.Replace(rateStr, ",", ".", -1)
		rate, err := strconv.ParseFloat(rateStr, 64)
		if err != nil {
			rate = 0
		}
		fl.Description = description
		fl.QtdAxle = axles
		fl.TotalValue = kmValue * rate
		fl.TypeOfLoad = fl.TypeOfLoad

		grouped[name] = append(grouped[name], fl)
	}
	return grouped, nil
}

func queryFuelStations(polyline string) ([]GasStation, error) {
	// Decodifica a polyline para obter os pontos da rota
	points, err := decodePolyline(polyline)
	if err != nil {
		return nil, err
	}
	// Calcula a bounding box que cobre todos os pontos da rota
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
	// (Opcional) Expanda um pouco a bbox para garantir cobertura
	padding := 0.05
	minLat -= padding
	minLon -= padding
	maxLat += padding
	maxLon += padding

	// Cria a query Overpass para buscar nós com amenity=fuel
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

	// Filtra os postos que estão realmente próximos da rota (tolerância em metros)
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
