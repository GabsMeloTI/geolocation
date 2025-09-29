package new_routes

import (
	"context"
	"encoding/json"
	"fmt"
	db "geolocation/db/sqlc"
	"geolocation/validation"
	"math"
	"net/http"
	neturl "net/url"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

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
		// Valor incorreto
	}

	return calculation, nil
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

type GoogleDirectionsResponse struct {
	Routes []struct {
		Legs []struct {
			Duration struct {
				Text  string `json:"text"`
				Value int64  `json:"value"`
			} `json:"duration"`
			DurationInTraffic struct {
				Text  string `json:"text"`
				Value int64  `json:"value"`
			} `json:"duration_in_traffic"`
			Distance struct {
				Text  string `json:"text"`
				Value int64  `json:"value"`
			} `json:"distance"`
		} `json:"legs"`
	} `json:"routes"`
	Status string `json:"status"`
}

func GetGoogleDurationWithTraffic(ctx context.Context, googleMapsAPIKey, origin, destination string, waypoints []string) (string, int64, error) {
	baseURL := "https://maps.googleapis.com/maps/api/directions/json"

	params := neturl.Values{}
	params.Set("origin", origin)
	params.Set("destination", destination)
	if len(waypoints) > 0 {
		params.Set("waypoints", strings.Join(waypoints, "|"))
	}
	params.Set("departure_time", "now")
	params.Set("key", googleMapsAPIKey)

	url := baseURL + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", 0, err
	}

	client := http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	var gdResp GoogleDirectionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&gdResp); err != nil {
		return "", 0, err
	}

	if gdResp.Status != "OK" || len(gdResp.Routes) == 0 || len(gdResp.Routes[0].Legs) == 0 {
		return "", 0, fmt.Errorf("Google Directions API retornou erro: %s", gdResp.Status)
	}

	leg := gdResp.Routes[0].Legs[0]
	// Prioriza duration_in_traffic se disponível
	if leg.DurationInTraffic.Value > 0 {
		return leg.DurationInTraffic.Text, leg.DurationInTraffic.Value, nil
	}

	return leg.Duration.Text, leg.Duration.Value, nil
}
