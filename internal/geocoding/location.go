package geocoding

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// ViaCEPResponse estrutura a resposta da API viacep.com.br
type ViaCEPResponse struct {
	Logradouro string `json:"logradouro"`
	Bairro     string `json:"bairro"`
	Localidade string `json:"localidade"`
	UF         string `json:"uf"`
	Erro       bool   `json:"erro"`
}

// BrasilAPIResponse estrutura a resposta da API brasilapi.com.br
type BrasilAPIResponse struct {
	Cep          string `json:"cep"`
	State        string `json:"state"`
	City         string `json:"city"`
	Service      string `json:"service"`
	Street       string `json:"street"`
	Neighborhood string `json:"neighborhood"`
}

// NominatimResponse estrutura a resposta da API Nominatim (OpenStreetMap)
type NominatimResponse struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}

type OSRMResponse struct {
	Routes []struct {
		Distance float64 `json:"distance"`
		Duration float64 `json:"duration"`
		Geometry string  `json:"geometry"`
		Legs     []struct {
			Steps []struct {
				Name     string  `json:"name"` // Nome da rua
				Distance float64 `json:"distance"`
				Maneuver struct {
					Type        string `json:"type"`        // Ex: "turn"
					Modifier    string `json:"modifier"`    // Ex: "right"
					Instruction string `json:"instruction"` // Algumas versões do OSRM já mandam o texto pronto
				} `json:"maneuver"`
			} `json:"steps"`
		} `json:"legs"`
	} `json:"routes"`
}

type ValhallaResponse struct {
	Trip       TripData   `json:"trip"`
	Alternates []TripData `json:"alternates,omitempty"` // Aqui entram as rotas extras
}

type TripData struct {
	Summary struct {
		Length float64 `json:"length"`
		Time   float64 `json:"time"`
	} `json:"summary"`
	Legs []struct {
		Maneuvers []struct {
			Instruction string `json:"instruction"`
		} `json:"maneuvers"`
		Shape string `json:"shape"`
	} `json:"legs"`
}

// fetchViaCEP realiza a consulta de endereço no serviço ViaCEP
func (s *Service) fetchViaCEP(ctx context.Context, cep string) (*ViaCEPResponse, error) {
	url := fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cep)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ViaCEPResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil || result.Erro {
		return nil, fmt.Errorf("cep não encontrado no viacep")
	}
	return &result, nil
}

// fetchBrasilAPI realiza a consulta de endereço no serviço BrasilAPI (v1)
func (s *Service) fetchBrasilAPI(ctx context.Context, cep string) (*BrasilAPIResponse, error) {
	url := fmt.Sprintf("https://brasilapi.com.br/api/cep/v1/%s", cep)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result BrasilAPIResponse
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cep não encontrado na brasilapi")
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// getLatLonByAddress busca latitude e longitude para um endereço (rua, cidade, estado) usando Nominatim (OSM)
func (s *Service) getLatLonByAddress(ctx context.Context, street, city, state string) (string, string, error) {
	// Formatamos a query para URL seguindo as exigências da API
	query := fmt.Sprintf("street=%s&city=%s&state=%s&format=json&limit=1",
		url.QueryEscape(street),
		url.QueryEscape(city),
		url.QueryEscape(state),
	)

	endpoint := "https://nominatim.openstreetmap.org/search?" + query
	req, _ := http.NewRequestWithContext(ctx, "GET", endpoint, nil)

	// IMPORTANTE: O Nominatim exige um User-Agent descritivo para evitar bloqueios
	req.Header.Set("User-Agent", "Geopy-Logistica-App/1.0 (contato@seudominio.com)")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var results []NominatimResponse
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil || len(results) == 0 {
		return "", "", fmt.Errorf("coordenadas não encontradas para este endereço")
	}

	return results[0].Lat, results[0].Lon, nil
}

func (s *Service) generateNavigationLinks(origin, dest LocationInfoAddress) (string, string) {
	// Google Maps: Parâmetros origin, destination e travelmode
	googleMapsURL := fmt.Sprintf(
		"https://www.google.com/maps/dir/?api=1&origin=%s,%s&destination=%s,%s&travelmode=driving",
		origin.Lat, origin.Lon, dest.Lat, dest.Lon,
	)

	// Waze: ll (latlong do destino), from (origem) e to (destino)
	wazeURL := fmt.Sprintf(
		"https://waze.com/ul?ll=%s,%s&navigate=yes&from=%s,%s&to=%s,%s",
		dest.Lat, dest.Lon, origin.Lat, origin.Lon, dest.Lat, dest.Lon,
	)

	return googleMapsURL, wazeURL
}

func (s *Service) fetchRouteInfo(ctx context.Context, origin, dest LocationInfoAddress) ([]TripInfo, error) {
	// OSRM espera: lon,lat;lon,lat
	// alternatives=true solicita rotas extras (geralmente até 3)
	url := fmt.Sprintf("http://router.project-osrm.org/route/v1/driving/%s,%s;%s,%s?overview=full&geometries=polyline&steps=true&alternatives=true",
		origin.Lon, origin.Lat, dest.Lon, dest.Lat,
	)

	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var osrm OSRMResponse
	if err := json.NewDecoder(resp.Body).Decode(&osrm); err != nil || len(osrm.Routes) == 0 {
		return nil, fmt.Errorf("falha ao calcular rota")
	}

	var trips []TripInfo
	for i, route := range osrm.Routes {
		var instructions []string
		for _, leg := range route.Legs {
			for _, step := range leg.Steps {
				acaoTraduzida := s.translateManeuver(step.Maneuver.Type, step.Maneuver.Modifier)
				distStr := fmt.Sprintf("%.0fm", step.Distance)
				if step.Distance >= 1000 {
					distStr = fmt.Sprintf("%.1fkm", step.Distance/1000)
				}
				var text string
				if step.Name != "" {
					text = fmt.Sprintf("%s na %s (%s)", acaoTraduzida, step.Name, distStr)
				} else {
					text = fmt.Sprintf("%s (%s)", acaoTraduzida, distStr)
				}
				if step.Maneuver.Type == "arrive" {
					text = "Você chegou ao seu destino"
				}
				instructions = append(instructions, text)
			}
		}

		trips = append(trips, TripInfo{
			Provider:          fmt.Sprintf("OSRM-OPT-%d", i+1),
			DistanceKm:        route.Distance / 1000,
			DurationMn:        route.Duration / 60,
			Polyline:          route.Geometry,
			PolylinePrecision: 5, // OSRM padrão é 5
			Instructions:      instructions,
		})
	}

	return trips, nil
}

func (s *Service) fetchValhallaRoute(ctx context.Context, origin, dest LocationInfoAddress, costing string) ([]TripInfo, error) {
	// Configuração para caminhão com idioma em português
	// Alternativas: 1 solicita rotas extras quando disponíveis
	requestBody := fmt.Sprintf(`{
    "locations":[{"lat":%s,"lon":%s},{"lat":%s,"lon":%s}],
    "costing":"%s", 
    "alternates":3,
    "directions_options":{"language":"pt-BR"}
}`, origin.Lat, origin.Lon, dest.Lat, dest.Lon, costing)

	url := "https://valhalla1.openstreetmap.de/route"
	log.Printf("[Valhalla] Chamando POST em: %s", url)

	req, _ := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(requestBody))
	req.Header.Set("User-Agent", "Geopy-Logistica-App/1.0 (contato@seudominio.com)")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("[Valhalla] Erro na requisição HTTP: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[Valhalla] Erro na API (Status %d): %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("Valhalla API error: %d", resp.StatusCode)
	}

	var response ValhallaResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	var allTrips []TripInfo

	// 1. Processar a Trip Principal
	allTrips = append(allTrips, s.mapValhallaTripToInfo(response.Trip, "PRINCIPAL"))

	// 2. Processar as Alternativas (se existirem)
	for i, alt := range response.Alternates {
		label := fmt.Sprintf("ALTERNATIVA-%d", i+1)
		allTrips = append(allTrips, s.mapValhallaTripToInfo(alt, label))
	}

	return allTrips, nil
}

func (s *Service) mapValhallaTripToInfo(data TripData, label string) TripInfo {
	var instructions []string
	var fullShape string

	// No Valhalla, uma trip pode ter várias legs (se houver pontos intermediários)
	// Vamos concatenar as instruções e pegar o shape da primeira leg (ou tratar conforme sua necessidade)
	for _, leg := range data.Legs {
		for _, m := range leg.Maneuvers {
			instructions = append(instructions, m.Instruction)
		}
		if fullShape == "" {
			fullShape = leg.Shape // Para rotas simples A->B, só existe uma leg
		}
	}

	return TripInfo{
		Provider:          fmt.Sprintf("VALHALLA-%s", label),
		DistanceKm:        data.Summary.Length,
		DurationMn:        data.Summary.Time / 60,
		Polyline:          fullShape,
		PolylinePrecision: 6,
		Instructions:      instructions,
	}
}
