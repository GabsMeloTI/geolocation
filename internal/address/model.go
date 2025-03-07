package address

import (
	"fmt"
	db "geolocation/db/sqlc"
)

type AddressResponse struct {
	ID           int32   `json:"id"`
	Street       string  `json:"street"`
	Neighborhood string  `json:"neighborhood,omitempty"`
	City         string  `json:"city"`
	State        string  `json:"state"`
	Number       string  `json:"number"`
	CEP          string  `json:"cep"`
	Lat          float64 `json:"latitude,omitempty"`
	Lon          float64 `json:"longitude,omitempty"`
}

func (p *AddressResponse) ParseFromLatLonRow(result *db.FindAddressesByLatLonRow) error {
	if result == nil {
		return fmt.Errorf("query returned nil result")
	}

	p.ID = result.AddressID
	p.Street = result.StreetName
	p.Neighborhood = result.NeighborhoodName.String
	p.City = result.CityName
	p.State = result.StateUf
	p.Number = result.Number.String
	p.CEP = result.Cep
	if result.Lat.Valid {
		p.Lat = result.Lat.Float64
	}
	if result.Lon.Valid {
		p.Lon = result.Lon.Float64
	}
	return nil
}

func (p *AddressResponse) ParseFromCEPRow(result *db.FindAddressesByCEPRow) error {
	if result == nil {
		return fmt.Errorf("query returned nil result")
	}

	p.ID = result.AddressID
	p.Street = result.StreetName
	p.Neighborhood = result.NeighborhoodName.String
	p.City = result.CityName
	p.State = result.StateUf
	p.Number = result.Number.String
	p.CEP = result.Cep
	if result.Lat.Valid {
		p.Lat = result.Lat.Float64
	}
	if result.Lon.Valid {
		p.Lon = result.Lon.Float64
	}
	return nil
}

func (p *AddressResponse) ParseFromQueryRow(result *db.FindAddressesByQueryRow) error {
	if result == nil {
		return fmt.Errorf("query returned nil result")
	}

	p.ID = result.StreetID
	p.Street = result.StreetName
	p.Neighborhood = result.NeighborhoodName.String
	p.City = result.CityName
	p.State = result.StateUf
	p.Number = result.Number.String
	p.CEP = result.Cep.String
	if result.Lat.Valid {
		p.Lat = result.Lat.Float64
	}
	if result.Lon.Valid {
		p.Lon = result.Lon.Float64
	}

	return nil
}
