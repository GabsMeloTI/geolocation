package address

import (
	"context"
	"errors"
	meiliaddress "geolocation/internal/meili_address"
	"log"
	"regexp"
	"strconv"
	"strings"
)

type InterfaceService interface {
	FindAddressesByQueryService(context.Context, string) ([]AddressResponse, error)
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
	var number string

	terms := strings.Split(query, ",")
	for _, term := range terms {
		if _, err := strconv.Atoi(term); err == nil {
			number = term
			continue
		}
	}

	addressesQuery, err := s.MeiliRepository.FindAddresses(ctx, query, 100)
	if err != nil {
		return nil, err
	}
	addressResponses, err = ParseQueryMeiliRow(addressesQuery, number)
	if err != nil {
		return nil, err
	}
	log.Println("searching finished")

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
