package geocoding

import (
	"context"
	"fmt"
	db "geolocation/db/sqlc"
	"log"
	"strings"
)

// formatAddress transforma os dados do banco (UniqueCep)
// em uma string de endereço formatada: RUA, BAIRRO, CIDADE/UF em CAIXA ALTA
func (s *Service) formatAddress(data db.UniqueCep) string {
	address := ""
	if data.StreetName.Valid {
		address += data.StreetName.String
	}
	if data.NeighborhoodName.Valid {
		if address != "" {
			address += ", "
		}
		address += data.NeighborhoodName.String
	}
	if data.CityName.Valid {
		if address != "" {
			address += ", "
		}
		address += data.CityName.String
	}
	if data.StateUf.Valid {
		if address != "" {
			address += "/"
		}
		address += data.StateUf.String
	}
	return strings.ToUpper(address)
}

// resolveCEP é o orquestrador de busca de coordenadas.
// Tenta primeiro a base local, e se falhar, inicia o processo de "enriquecimento" via APIs externas.
func (s *Service) resolveCEP(ctx context.Context, cep string) (LocationInfoAddress, error) {
	// 1. Tenta Banco Interno (Fonte primária, rápida e sem custo)
	origin, err := s.repository.ConsultCEP(ctx, cep)
	if err == nil && origin.Lat.Valid && origin.Lat.Float64 != 0 {
		return LocationInfoAddress{
			Address: s.formatAddress(origin),
			Lat:     fmt.Sprintf("%.6f", origin.Lat.Float64),
			Lon:     fmt.Sprintf("%.6f", origin.Lon.Float64),
		}, nil
	}

	log.Printf("[Fallback] CEP %s não encontrado ou sem coordenadas. Iniciando busca externa.", cep)

	// 2. Tenta APIs de Endereço (ViaCEP -> BrasilAPI) para obter o nome da rua/bairro
	var street, neighborhood, city, state string
	v, errV := s.fetchViaCEP(ctx, cep)
	if errV == nil {
		street, neighborhood, city, state = v.Logradouro, v.Bairro, v.Localidade, v.UF
	} else {
		b, errB := s.fetchBrasilAPI(ctx, cep)
		if errB == nil {
			street, neighborhood, city, state = b.Street, b.Neighborhood, b.City, b.State
		}
	}

	// 3. Se obteve o endereço, busca a Latitude/Longitude no Nominatim (OpenStreetMap)
	if street != "" {
		lat, lon, errGeo := s.getLatLonByAddress(ctx, street, city, state)
		if errGeo == nil {
			log.Printf("[Fallback] Sucesso! Coordenadas obtidas via Nominatim para o CEP: %s", cep)

			// Retorna o endereço formatado e as coordenadas encontradas externamente
			return LocationInfoAddress{
				Address: strings.ToUpper(fmt.Sprintf("%s, %s, %s/%s", street, neighborhood, city, state)),
				Lat:     lat,
				Lon:     lon,
			}, nil
		}
	}

	return LocationInfoAddress{}, fmt.Errorf("não foi possível localizar coordenadas para o CEP %s após todas as tentativas", cep)
}

func (s *Service) translateManeuver(mType, modifier string) string {
	types := map[string]string{
		"turn":            "Vire",
		"new name":        "Continue na",
		"depart":          "Siga em direção a",
		"arrive":          "Chegada",
		"merge":           "Acesse a",
		"on ramp":         "Pegue a rampa para",
		"off ramp":        "Saia pela rampa para",
		"end of road":     "No fim da rua, vire",
		"fork":            "Mantenha-se na",
		"roundabout":      "Na rotatória, pegue a saída para",
		"exit roundabout": "Saia da rotatória para",
		"rotary":          "Entre na rotatória e saia na",
		"roundabout turn": "Na rotatória, vire",
		"notification":    "Atenção:",
	}
	modifiers := map[string]string{
		"right":        "à direita",
		"left":         "à esquerda",
		"slight right": "levemente à direita",
		"slight left":  "levemente à esquerda",
		"sharp right":  "fechado à direita",
		"sharp left":   "fechado à esquerda",
		"straight":     "em frente",
		"uturn":        "faça o retorno",
	}

	acao := types[mType]
	if acao == "" {
		acao = mType
	}
	if mType == "arrive" {
		return "Você chegou ao seu destino"
	}
	if mType == "depart" && modifier == "straight" {
		return "Siga em frente"
	}

	direcao := modifiers[modifier]
	if direcao != "" {
		return fmt.Sprintf("%s %s", acao, direcao)
	}

	return acao
}
