package address

import (
	"context"
	"database/sql"
	"errors"
	db "geolocation/db/sqlc"
	meiliaddress "geolocation/internal/meili_address"
	"golang.org/x/text/unicode/norm"
	"log"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

type InterfaceService interface {
	FindAddressesByQueryService(context.Context, string) ([]AddressResponse, error)
	FindAddressesByQueryV2Service(context.Context, string) ([]AddressResponse, error)
	FindAddressesByCEPService(ctx context.Context, query string) (AddressCEPResponse, error)
	FindStateAll(context.Context) ([]StateResponse, error)
	FindCityAll(context.Context, int32) ([]CityResponse, error)
}

type Service struct {
	InterfaceService InterfaceRepository
	MeiliRepository  meiliaddress.InterfaceRepository
}

func NewAddressService(InterfaceService InterfaceRepository, MeiliRepository meiliaddress.InterfaceRepository) *Service {
	return &Service{InterfaceService: InterfaceService, MeiliRepository: MeiliRepository}
}

func (s *Service) FindAddressesByCEPService(ctx context.Context, query string) (AddressCEPResponse, error) {
	cepRegex := regexp.MustCompile(`^\d{8}$`)
	normalizedQuery := strings.ReplaceAll(strings.ReplaceAll(query, "-", ""), " ", "")
	isCEP := cepRegex.MatchString(normalizedQuery)
	if !isCEP {
		return AddressCEPResponse{}, errors.New("CEP inválido")
	}

	addresses, err := s.InterfaceService.FindAddressGroupedByCEPRepository(ctx, normalizedQuery)
	if err != nil {
		return AddressCEPResponse{}, err
	}
	cityName := ""
	stateUf := ""
	neighborhoodSet := make(map[string]bool)
	streetSet := make(map[string]bool)
	var latitude, longitude float64

	log.Println("searching by cep")

	for _, addr := range addresses {
		if cityName == "" {
			cityName = addr.CityName
			stateUf = addr.StateUf
		}

		if addr.NeighborhoodName.Valid {
			neighborhoodSet[addr.NeighborhoodName.String] = true
		}

		streetSet[addr.StreetName] = true

		latitude = addr.Latitude.Float64
		longitude = addr.Longitude.Float64
	}

	responseType := "city"
	neighborhoodName := ""
	streetName := ""

	if len(streetSet) == 1 {
		responseType = "street"
		for s := range streetSet {
			streetName = s
		}
		if len(neighborhoodSet) == 1 {
			for n := range neighborhoodSet {
				neighborhoodName = n
			}
		}
	} else if len(neighborhoodSet) == 1 {
		responseType = "neighborhood"
		for n := range neighborhoodSet {
			neighborhoodName = n
		}
	}

	response := AddressCEPResponse{
		CEP:              normalizedQuery,
		Type:             responseType,
		CityName:         cityName,
		StateUf:          stateUf,
		NeighborhoodName: neighborhoodName,
		StreetName:       streetName,
		Latitude:         latitude,
		Longitude:        longitude,
	}

	return response, nil
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
