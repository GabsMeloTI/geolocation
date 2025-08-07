package address

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	db "geolocation/db/sqlc"
	meiliaddress "geolocation/internal/meili_address"
	"golang.org/x/text/unicode/norm"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

type InterfaceService interface {
	FindAddressesByQueryService(context.Context, string) ([]AddressResponse, error)
	FindAddressesByQueryV2Service(context.Context, string) ([]AddressResponse, error)
	FindUniqueAddressesByCEPService(context.Context, string) ([]AddressResponse, error)
	FindAddressesByCEPService(ctx context.Context, query string) (AddressCEPResponse, error)
	FindStateAll(context.Context) ([]StateResponse, error)
	FindCityAll(context.Context, int32) ([]CityResponse, error)
}

type Service struct {
	InterfaceService InterfaceRepository
	MeiliRepository  meiliaddress.InterfaceRepository
	GoogleMapsAPIKey string
}

func NewAddressService(InterfaceService InterfaceRepository, MeiliRepository meiliaddress.InterfaceRepository, GoogleMapsAPIKey string) *Service {
	return &Service{InterfaceService: InterfaceService, MeiliRepository: MeiliRepository, GoogleMapsAPIKey: GoogleMapsAPIKey}
}

func (s *Service) FindAddressesByCEPService(ctx context.Context, query string) (AddressCEPResponse, error) {
	cepRegex := regexp.MustCompile(`^\d{8}$`)
	normalizedQuery := strings.ReplaceAll(strings.ReplaceAll(query, "-", ""), " ", "")
	isCEP := cepRegex.MatchString(normalizedQuery)
	if !isCEP {
		return AddressCEPResponse{}, errors.New("CEP inválido")
	}

	addr, err := s.InterfaceService.FindAddressGroupedByCEPRepository(ctx, normalizedQuery)
	if err != nil {
		log.Println("buscando cep pelo API Brasil")
		address, apiErr := findCEPByAPIBrasil(ctx, normalizedQuery)
		if apiErr != nil {
			log.Println("erro ao buscar CEP em ambas base de dados:", apiErr)
			return AddressCEPResponse{}, nil
		}
		return address, nil
	}

	response := AddressCEPResponse{
		CEP:              normalizedQuery,
		Type:             "street",
		CityName:         addr.CityName.String,
		StateUf:          addr.StateUf.String,
		NeighborhoodName: addr.NeighborhoodName.String,
		StreetName:       addr.StreetName.String,
		Latitude:         addr.Latitude.Float64,
		Longitude:        addr.Longitude.Float64,
	}

	return response, nil
}

func (s *Service) FindUniqueAddressesByCEPService(ctx context.Context, query string) ([]AddressResponse, error) {
	var addressResponses []AddressResponse
	normalizedQuery := strings.ReplaceAll(strings.ReplaceAll(query, "-", ""), " ", "")

	addressCEP, err := s.InterfaceService.FindUniqueAddressesByCEPRepository(ctx, normalizedQuery)
	if err != nil {

		return nil, err
	}
	addressResponses, err = ParseFromUniqueCEPRow(addressCEP)
	if err != nil {

		return nil, err
	}
	return addressResponses, nil
}

func (s *Service) FindAddressesByQueryService(ctx context.Context, query string) ([]AddressResponse, error) {
	var addressResponses []AddressResponse
	var neighborhoodsResponses []AddressResponse

	cepRegex := regexp.MustCompile(`^\d{8}$`)
	normalizedQuery := strings.ReplaceAll(strings.ReplaceAll(query, "-", ""), " ", "")
	isCEP := cepRegex.MatchString(normalizedQuery)

	latLonRegex := regexp.MustCompile(`^\s*-?\d{1,2}(\.\d+)?\s*[, ]\s*-?\d{1,3}(\.\d+)?\s*$`)
	isLatLon := latLonRegex.MatchString(query)

	if isLatLon {
		coords := strings.SplitN(query, ",", 2)
		lat, _ := strconv.ParseFloat(strings.TrimSpace(coords[0]), 64)
		lng, _ := strconv.ParseFloat(strings.TrimSpace(coords[1]), 64)

		log.Println("searching by latitude and longitude")
		addressLatLon, err := s.InterfaceService.FindAddressesByLatLonRepository(ctx, db.FindAddressesByLatLonParams{
			Lat: sql.NullFloat64{
				Float64: lat,
			},
			Lon: sql.NullFloat64{
				Float64: lng,
			},
		})
		if err != nil {
			return nil, err
		}

		addressResponses, err = ParseFromLatLonRow(addressLatLon)
		if err != nil {
			return nil, err
		}

		return addressResponses, nil
	}

	if isCEP {
		log.Println("searching by cep")
		addressCEP, err := s.InterfaceService.FindAddressesByCEPRepository(ctx, normalizedQuery)
		if err != nil {
			return nil, err
		}
		addressResponses, err = ParseFromCEPRow(addressCEP)
		if err != nil {
			return nil, err
		}
		return addressResponses, nil
	}

	terms := strings.Split(query, ",")
	var rua, numero, cidade, estado, bairro string
	ruaDetected := false

	for _, term := range terms {
		term = strings.TrimSpace(term)

		normalized := norm.NFD.String(term)
		var result strings.Builder
		for _, r := range normalized {
			if unicode.Is(unicode.Mn, r) {
				continue
			}
			result.WriteRune(r)
		}

		term = strings.ToLower(result.String())

		if _, err := strconv.Atoi(term); err == nil {
			numero = term
			continue
		}

		if strings.Contains(term, "rua") || strings.Contains(term, "avenida") ||
			strings.Contains(term, "estrada") || strings.Contains(term, "rodovia") {
			rua = term
			ruaDetected = true
			continue
		}

		neighborhoods, err := s.InterfaceService.FindNeighborhoodsByNameRepository(ctx, term)
		if err == nil && len(neighborhoods) > 0 {
			if !ruaDetected {
				for _, neighborhood := range neighborhoods {
					neighborhoodsResponses = append(neighborhoodsResponses, AddressResponse{
						Neighborhood: neighborhood.Name,
						City:         neighborhood.City,
						State:        neighborhood.Uf,
						Latitude:     neighborhood.Lat.Float64,
						Longitude:    neighborhood.Lon.Float64,
					})
				}
			}
			bairro = term
		}

		cities, err := s.InterfaceService.FindCitiesByNameRepository(ctx, term)
		if err == nil && len(cities) > 0 {
			if !ruaDetected {
				var responses []AddressResponse
				for _, city := range cities {
					responses = append(responses, AddressResponse{
						City:      city.Name,
						State:     city.Uf,
						Latitude:  city.Lat.Float64,
						Longitude: city.Lon.Float64,
					})
				}
				return append(responses, neighborhoodsResponses...), nil
			}
			if bairro == "" {
				cidade = term
				continue
			}
		}

		states, err := s.InterfaceService.FindStateByNameRepository(ctx, term)
		if err == nil && len(states) > 0 {
			if !ruaDetected {
				var responses []AddressResponse
				for _, state := range states {
					responses = append(responses, AddressResponse{
						State:     state.Uf,
						Latitude:  state.Lat.Float64,
						Longitude: state.Lon.Float64,
					})
				}
				return append(responses, neighborhoodsResponses...), nil
			}
			if bairro == "" {
				estado = term
				continue
			}
		}
		if bairro == "" {
			rua = term
			ruaDetected = true
			continue
		}
	}

	addressesQuery, err := s.InterfaceService.FindAddressesByQueryRepository(ctx, db.FindAddressesByQueryParams{
		State:        estado,
		City:         cidade,
		Neighborhood: bairro,
		Street:       rua,
		Number:       numero,
	})
	if err != nil {
		return nil, err
	}
	addressResponses, err = ParseFromQueryRow(addressesQuery, numero)
	if err != nil {
		return nil, err
	}

	return addressResponses, nil
}

func (s *Service) FindAddressesByQueryV2Service(ctx context.Context, query string) ([]AddressResponse, error) {
	var addressResponses []AddressResponse
	var number string

	cepRegex := regexp.MustCompile(`^\d{8}$`)
	normalizedQuery := strings.ReplaceAll(strings.ReplaceAll(query, "-", ""), " ", "")
	isCEP := cepRegex.MatchString(normalizedQuery)

	latLonRegex := regexp.MustCompile(`^\s*-?\d{1,2}(\.\d+)?\s*[, ]\s*-?\d{1,3}(\.\d+)?\s*$`)
	isLatLon := latLonRegex.MatchString(query)

	if isLatLon {
		coords := strings.SplitN(query, ",", 2)
		lat, _ := strconv.ParseFloat(strings.TrimSpace(coords[0]), 64)
		lng, _ := strconv.ParseFloat(strings.TrimSpace(coords[1]), 64)

		log.Println("searching by latitude and longitude")
		addressLatLon, err := s.InterfaceService.FindAddressesByLatLonRepository(ctx, db.FindAddressesByLatLonParams{
			Lat: sql.NullFloat64{
				Float64: lat,
			},
			Lon: sql.NullFloat64{
				Float64: lng,
			},
		})
		if err != nil {
			return nil, err
		}

		addressResponses, err = ParseFromLatLonRow(addressLatLon)
		if err != nil {
			return nil, err
		}

		return addressResponses, nil
	}

	if isCEP {
		log.Println("searching by cep")
		addressCEP, err := s.InterfaceService.FindAddressesByCEPRepository(ctx, normalizedQuery)
		if err != nil {
			return nil, err
		}
		addressResponses, err = ParseFromCEPRow(addressCEP)
		if err != nil {
			return nil, err
		}
		return addressResponses, nil
	}

	terms := strings.Split(query, ",")
	for _, term := range terms {
		if _, err := strconv.Atoi(term); err == nil {
			number = term
			continue
		}
	}

	var meiliAddresses []meiliaddress.MeiliAddress
	streetsMeili, err := s.MeiliRepository.FindMeiliStreetsRepository(ctx, query, 50)
	if err != nil {
		return nil, err
	}

	for _, street := range streetsMeili {
		addressesList, err := s.InterfaceService.FindAddressByStreetIDRepository(ctx, street.StreetID)
		if err != nil {
			return nil, err
		}

		for _, addr := range addressesList {
			meiliAddr := meiliaddress.MeiliAddress{
				StreetID:         street.StreetID,
				StreetName:       street.StreetName,
				NeighborhoodName: street.NeighborhoodName,
				NeighborhoodLat:  street.NeighborhoodLat,
				NeighborhoodLon:  street.NeighborhoodLon,
				CityName:         street.CityName,
				CityLat:          street.CityLat,
				CityLon:          street.CityLon,
				StateUf:          street.StateUf,
				StateName:        street.StateName,
				StateLat:         street.StateLat,
				StateLon:         street.StateLon,
				AddressID:        addr.ID,
				Number:           addr.Number.String,
				Cep:              addr.Cep,
				Lat:              addr.Lat.Float64,
				Lon:              addr.Lon.Float64,
			}

			meiliAddresses = append(meiliAddresses, meiliAddr)
		}
	}
	addressResponses, err = ParseQueryMeiliRow(meiliAddresses, number)
	if err != nil {
		return nil, err
	}

	return addressResponses, nil
}

func (s *Service) FindStateAll(ctx context.Context) ([]StateResponse, error) {
	var stateResponse []StateResponse

	states, err := s.InterfaceService.FindStateAll(ctx)
	if err != nil {
		return nil, err
	}

	for _, state := range states {
		stateResponse = append(stateResponse, StateResponse{
			ID:   state.ID,
			Name: state.Name,
			Uf:   state.Uf,
		})
	}

	return stateResponse, nil
}

func (s *Service) FindCityAll(ctx context.Context, idState int32) ([]CityResponse, error) {
	var cityResponse []CityResponse

	cities, err := s.InterfaceService.FindCityAll(ctx, idState)
	if err != nil {
		return nil, err
	}

	for _, city := range cities {
		cityResponse = append(cityResponse, CityResponse{
			ID:   city.ID,
			Name: city.Name,
		})
	}

	return cityResponse, nil
}

func findCEPByAPIBrasil(ctx context.Context, cep string) (AddressCEPResponse, error) {
	url := "https://gateway.apibrasil.io/api/v2/cep/cep"
	bodyData := map[string]string{"cep": cep}
	bodyJSON, _ := json.Marshal(bodyData)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyJSON))
	if err != nil {
		return AddressCEPResponse{}, err
	}
	bearer := os.Getenv("BEARER_TOKEN")
	device := os.Getenv("DEVICE_TOKEN_CEP")
	req.Header.Set("Authorization", "Bearer "+bearer)
	req.Header.Set("DeviceToken", device)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return AddressCEPResponse{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return AddressCEPResponse{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return AddressCEPResponse{}, fmt.Errorf("erro na APIBrasil: %s", string(body))
	}

	var apiResp APIBrasilResponse
	err = json.Unmarshal(body, &apiResp)
	if err != nil {
		return AddressCEPResponse{}, err
	}

	if apiResp.Error {
		return AddressCEPResponse{}, errors.New(apiResp.Message)
	}

	cepResp := apiResp.Response.CEP
	lat, _ := strconv.ParseFloat(cepResp.Latitude, 64)
	long, _ := strconv.ParseFloat(cepResp.Longitude, 64)

	return AddressCEPResponse{
		CEP:              apiResp.Response.CEP.CEP,
		Type:             strings.ToUpper(apiResp.Response.CEP.Tipo),
		CityName:         strings.ToUpper(apiResp.Response.CEP.Cidade.Cidade),
		StateUf:          strings.ToUpper(apiResp.Response.CEP.Estado),
		NeighborhoodName: strings.ToUpper(apiResp.Response.CEP.Bairro.Bairro),
		StreetName:       strings.ToUpper(apiResp.Response.CEP.Logradouro),
		Latitude:         lat,
		Longitude:        long,
	}, nil
}
