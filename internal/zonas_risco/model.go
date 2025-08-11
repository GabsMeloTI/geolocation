package zonas_risco

import (
	db "geolocation/db/sqlc"
)

type CreateZonaRiscoRequest struct {
	Name   string  `json:"name"`
	Cep    string  `json:"cep"`
	Lat    float64 `json:"lat"`
	Lng    float64 `json:"lng"`
	Radius int64   `json:"radius"`
}

type UpdateZonaRiscoRequest struct {
	ID     int64   `json:"id"`
	Name   string  `json:"name"`
	Cep    string  `json:"cep"`
	Lat    float64 `json:"lat"`
	Lng    float64 `json:"lng"`
	Radius int64   `json:"radius"`
	Status bool    `json:"status"`
}

type ZonaRiscoResponse struct {
	ID     int64   `json:"id"`
	Name   string  `json:"name"`
	Cep    string  `json:"cep"`
	Lat    float64 `json:"lat"`
	Lng    float64 `json:"lng"`
	Radius int64   `json:"radius"`
	Status bool    `json:"status"`
}

func (r *ZonaRiscoResponse) ParseFromDb(result db.ZonasRisco) {
	r.ID = result.ID
	r.Name = result.Name
	r.Cep = result.Cep
	r.Lat = result.Lat
	r.Lng = result.Lng
	r.Radius = result.Radius
	r.Status = result.Status
}
