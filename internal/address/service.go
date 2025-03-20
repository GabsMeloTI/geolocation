package address

import (
	"context"
	"database/sql"
	"errors"
	db "geolocation/db/sqlc"
	"golang.org/x/text/unicode/norm"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

type InterfaceService interface {
	FindAddressesByQueryService(context.Context, string) ([]AddressResponse, error)
	FindAddressesByCEPService(ctx context.Context, query string) (AddressCEPResponse, error)
	FindStateAll(context.Context) ([]StateResponse, error)
	FindCityAll(context.Context, int32) ([]CityResponse, error)
}

type Service struct {
	InterfaceService InterfaceRepository
}

func NewAddresssService(InterfaceService InterfaceRepository) *Service {
	return &Service{InterfaceService}
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

	for _, addr := range addresses {
		if cityName == "" {
			cityName = addr.CityName
			stateUf = addr.StateUf
		}

		if addr.NeighborhoodName.Valid {
			neighborhoodSet[addr.NeighborhoodName.String] = true
		}

		streetSet[addr.StreetName] = true
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
	}

	return response, nil
}

func (s *Service) FindAddressesByQueryService(ctx context.Context, query string) ([]AddressResponse, error) {
	var addressResponses []AddressResponse

	cepRegex := regexp.MustCompile(`^\d{8}$`)
	normalizedQuery := strings.ReplaceAll(strings.ReplaceAll(query, "-", ""), " ", "")
	isCEP := cepRegex.MatchString(normalizedQuery)

	latLonRegex := regexp.MustCompile(`^\s*-?\d{1,2}(\.\d+)?\s*[, ]\s*-?\d{1,3}(\.\d+)?\s*$`)
	isLatLon := latLonRegex.MatchString(query)

	if isLatLon {
		coords := strings.SplitN(query, ",", 2)
		lat, _ := strconv.ParseFloat(strings.TrimSpace(coords[0]), 64)
		lng, _ := strconv.ParseFloat(strings.TrimSpace(coords[1]), 64)

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

		prefixes := []string{"travessa", "rua", "avenida", "estrada", "rodovia"}
		for _, prefix := range prefixes {
			if strings.HasPrefix(term, prefix) {
				rua = term
				break
			}
		}

		if _, err := strconv.Atoi(term); err == nil {
			numero = term
			continue
		}

		if isBairro, err := s.InterfaceService.IsNeighborhoodRepository(ctx, term); err == nil && isBairro {
			bairro = term
			continue
		}

		if isCidade, err := s.InterfaceService.IsCityRepository(ctx, term); err == nil && isCidade {
			cidade = term
			continue
		}

		if isEstado, err := s.InterfaceService.IsStateRepository(ctx, term); err == nil && isEstado {
			estado = term
			continue
		}

		rua = term
	}

	addressesQuery, err := s.InterfaceService.FindAddressesByQueryRepository(ctx, db.FindAddressesByQueryParams{
		Street:       rua,
		City:         cidade,
		State:        estado,
		Neighborhood: bairro,
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
