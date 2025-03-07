package address

import (
	"context"
	"database/sql"
	db "geolocation/db/sqlc"
	"regexp"
	"strconv"
	"strings"
)

type InterfaceService interface {
	FindAddressesByQueryService(ctx context.Context, query string) ([]AddressResponse, error)
}

type Service struct {
	InterfaceService InterfaceRepository
}

func NewAddresssService(InterfaceService InterfaceRepository) *Service {
	return &Service{InterfaceService}
}

func (p *Service) FindAddressesByQueryService(ctx context.Context, query string) ([]AddressResponse, error) {
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

		addressLatLon, err := p.InterfaceService.FindAddressesByLatLonRepository(ctx, db.FindAddressesByLatLonParams{
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
		for _, result := range addressLatLon {
			var addressResponse AddressResponse
			err = addressResponse.ParseFromLatLonRow(&result)
			if err != nil {
				return nil, err
			}
			addressResponses = append(addressResponses, addressResponse)
		}
		return addressResponses, nil
	}

	if isCEP {
		addressCEP, err := p.InterfaceService.FindAddressesByCEPRepository(ctx, normalizedQuery)
		if err != nil {
			return nil, err
		}
		for _, result := range addressCEP {
			var addressResponse AddressResponse
			err = addressResponse.ParseFromCEPRow(&result)
			if err != nil {
				return nil, err
			}
			addressResponses = append(addressResponses, addressResponse)
		}
		return addressResponses, nil
	}

	terms := strings.Split(query, ",")
	var rua, numero, cidade, estado, bairro string

	for _, term := range terms {
		term = strings.TrimSpace(term)
		if _, err := strconv.Atoi(term); err == nil {
			numero = term
			continue
		}

		if isBairro, err := p.InterfaceService.IsNeighborhoodRepository(ctx, term); err == nil && isBairro {
			bairro = term
			continue
		}

		if isCidade, err := p.InterfaceService.IsCityRepository(ctx, term); err == nil && isCidade {
			cidade = term
			continue
		}

		if isEstado, err := p.InterfaceService.IsStateRepository(ctx, term); err == nil && isEstado {
			estado = term
			continue
		}

		rua = term
	}

	addressesQuery, err := p.InterfaceService.FindAddressesByQueryRepository(ctx, db.FindAddressesByQueryParams{
		Column1: rua,
		Column2: cidade,
		Column3: estado,
		Column4: bairro,
		Column5: numero,
	})
	if err != nil {
		return nil, err
	}

	for _, result := range addressesQuery {
		var addressResponse AddressResponse
		err = addressResponse.ParseFromQueryRow(&result)
		if err != nil {
			return nil, err
		}
		addressResponses = append(addressResponses, addressResponse)
	}
	return addressResponses, nil
}
