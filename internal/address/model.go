package address

import (
	"fmt"
	db "geolocation/db/sqlc"
	"sort"
)

type AddressResponse struct {
	IDStreet     int32           `json:"id_street"`
	Street       string          `json:"street"`
	Neighborhood string          `json:"neighborhood,omitempty"`
	City         string          `json:"city"`
	State        string          `json:"state"`
	Latitude     float64         `json:"latitude,omitempty"`
	Longitude    float64         `json:"longitude,omitempty"`
	Addresses    []AddressDetail `json:"addresses"`
}

type AddressDetail struct {
	IDAddress int32   `json:"id_address"`
	Number    string  `json:"number"`
	CEP       string  `json:"cep"`
	IsExactly bool    `json:"is_exactly"`
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
}

func ParseFromLatLonRow(results []db.FindAddressesByLatLonRow) ([]AddressResponse, error) {
	if len(results) == 0 {
		return nil, fmt.Errorf("query returned nil result")
	}

	grouped := make(map[int32]*AddressResponse)

	for _, result := range results {
		if _, exists := grouped[result.StreetID]; !exists {
			grouped[result.StreetID] = &AddressResponse{
				IDStreet:     result.StreetID,
				Street:       result.StreetName,
				Neighborhood: result.NeighborhoodName.String,
				City:         result.CityName,
				State:        result.StateUf,
				Addresses:    []AddressDetail{},
			}
		}

		addressDetail := AddressDetail{
			IDAddress: result.AddressID,
			Number:    result.Number.String,
			CEP:       result.Cep,
			IsExactly: false,
		}

		if result.Lat.Valid {
			addressDetail.Latitude = result.Lat.Float64
		}
		if result.Lon.Valid {
			addressDetail.Longitude = result.Lon.Float64
		}

		grouped[result.StreetID].Addresses = append(grouped[result.StreetID].Addresses, addressDetail)
	}

	return calculateGroupedLatitudes(grouped), nil
}

func ParseFromCEPRow(results []db.FindAddressesByCEPRow) ([]AddressResponse, error) {
	if len(results) == 0 {
		return nil, fmt.Errorf("query returned nil result")
	}

	grouped := make(map[int32]*AddressResponse)

	for _, result := range results {
		if _, exists := grouped[result.StreetID]; !exists {
			grouped[result.StreetID] = &AddressResponse{
				IDStreet:     result.StreetID,
				Street:       result.StreetName,
				Neighborhood: result.NeighborhoodName.String,
				City:         result.CityName,
				State:        result.StateUf,
				Addresses:    []AddressDetail{},
			}
		}

		addressDetail := AddressDetail{
			IDAddress: result.AddressID,
			Number:    result.Number.String,
			CEP:       result.Cep,
			IsExactly: false,
		}

		if result.Lat.Valid {
			addressDetail.Latitude = result.Lat.Float64
		}
		if result.Lon.Valid {
			addressDetail.Longitude = result.Lon.Float64
		}

		grouped[result.StreetID].Addresses = append(grouped[result.StreetID].Addresses, addressDetail)
	}

	return calculateGroupedLatitudes(grouped), nil
}

func ParseFromQueryRow(results []db.FindAddressesByQueryRow, numero string) ([]AddressResponse, error) {
	if len(results) == 0 {
		return nil, fmt.Errorf("query returned nil result")
	}

	grouped := make(map[int32]*AddressResponse)

	for _, result := range results {
		if _, exists := grouped[result.StreetID]; !exists {
			grouped[result.StreetID] = &AddressResponse{
				IDStreet:     result.StreetID,
				Street:       result.StreetName,
				Neighborhood: result.NeighborhoodName.String,
				City:         result.CityName,
				State:        result.StateUf,
				Addresses:    []AddressDetail{},
			}
		}

		addressDetail := AddressDetail{
			IDAddress: result.AddressID.Int32,
			Number:    result.Number.String,
			CEP:       result.Cep.String,
			IsExactly: result.Number.String == numero,
		}

		if result.Lat.Valid {
			addressDetail.Latitude = result.Lat.Float64
		}
		if result.Lon.Valid {
			addressDetail.Longitude = result.Lon.Float64
		}

		grouped[result.StreetID].Addresses = append(grouped[result.StreetID].Addresses, addressDetail)
	}

	addressResponses := calculateGroupedLatitudes(grouped)

	for _, response := range addressResponses {
		sort.Slice(response.Addresses, func(i, j int) bool {
			return response.Addresses[i].IsExactly && !response.Addresses[j].IsExactly
		})
	}

	return calculateGroupedLatitudes(grouped), nil
}

func calculateGroupedLatitudes(grouped map[int32]*AddressResponse) []AddressResponse {
	var addressResponses []AddressResponse

	for _, addressResponse := range grouped {
		var lat, lon float64

		if len(addressResponse.Addresses) == 0 {
			lat, lon = 0, 0
		} else if len(addressResponse.Addresses) == 1 {
			lat, lon = addressResponse.Addresses[0].Latitude, addressResponse.Addresses[0].Longitude
		} else {
			latitudes := make([]float64, len(addressResponse.Addresses))
			longitudes := make([]float64, len(addressResponse.Addresses))

			for i, address := range addressResponse.Addresses {
				latitudes[i] = address.Latitude
				longitudes[i] = address.Longitude
			}

			sort.Float64s(latitudes)
			sort.Float64s(longitudes)

			if len(latitudes)%2 == 0 {
				lat = latitudes[len(latitudes)/2-1]
				lon = longitudes[len(longitudes)/2-1]
			} else {
				lat = latitudes[len(latitudes)/2]
				lon = longitudes[len(longitudes)/2]
			}
		}

		response := AddressResponse{
			IDStreet:     addressResponse.IDStreet,
			Street:       addressResponse.Street,
			Neighborhood: addressResponse.Neighborhood,
			City:         addressResponse.City,
			State:        addressResponse.State,
			Latitude:     lat,
			Longitude:    lon,
			Addresses:    addressResponse.Addresses,
		}

		addressResponses = append(addressResponses, response)
	}
	sort.Slice(addressResponses, func(i, j int) bool {
		return addressResponses[i].Addresses[0].IsExactly && !addressResponses[j].Addresses[0].IsExactly
	})

	return addressResponses
}
