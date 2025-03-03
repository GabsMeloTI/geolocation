package tractor_unit

import (
	"database/sql"
	db "geolocation/db/sqlc"
	"time"
)

type CreateTractorUnitRequest struct {
	LicensePlate    string  `json:"license_plate"`
	DriverID        int64   `json:"driver_id"`
	UserID          int64   `json:"user_id"`
	Chassis         string  `json:"chassis"`
	Brand           string  `json:"brand"`
	Model           string  `json:"model"`
	ManufactureYear int64   `json:"manufacture_year"`
	EnginePower     string  `json:"engine_power"`
	UnitType        string  `json:"unit_type" validate:"oneof=stump truck tractor_unit"`
	CanCouple       bool    `json:"can_couple"`
	Height          float64 `json:"height"`
	State           string  `json:"state"`
	Renavan         string  `json:"renavan"`
	Capacity        string  `json:"capacity"`
	Width           float64 `json:"width"`
	Length          float64 `json:"length"`
	Color           string  `json:"color"`
}

type UpdateTractorUnitRequest struct {
	ID              int64   `json:"id"`
	LicensePlate    string  `json:"license_plate"`
	DriverID        int64   `json:"driver_id"`
	Chassis         string  `json:"chassis"`
	Brand           string  `json:"brand"`
	Model           string  `json:"model"`
	ManufactureYear int64   `json:"manufacture_year"`
	EnginePower     string  `json:"engine_power"`
	UnitType        string  `json:"unit_type"`
	Height          float64 `json:"height"`
	UserID          int64   `json:"user_id"`
	State           string  `json:"state"`
	Renavan         string  `json:"renavan"`
	Capacity        string  `json:"capacity"`
	Width           float64 `json:"width"`
	Length          float64 `json:"length"`
	Color           string  `json:"color"`
}

type TractorUnitResponse struct {
	ID              int64      `json:"id"`
	UserID          int64      `json:"user_id"`
	LicensePlate    string     `json:"license_plate"`
	DriverID        int64      `json:"driver_id"`
	Chassis         string     `json:"chassis"`
	Brand           string     `json:"brand"`
	Model           string     `json:"model"`
	ManufactureYear int64      `json:"manufacture_year"`
	EnginePower     string     `json:"engine_power"`
	UnitType        string     `json:"unit_type"`
	Height          float64    `json:"height"`
	State           string     `json:"state"`
	Renavan         string     `json:"renavan"`
	Capacity        string     `json:"capacity"`
	Width           float64    `json:"width"`
	Length          float64    `json:"length"`
	Color           string     `json:"color"`
	Status          bool       `json:"status"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       *time.Time `json:"updated_at"`
}

func (p *CreateTractorUnitRequest) ParseCreateToTractorUnit() db.CreateTractorUnitParams {
	arg := db.CreateTractorUnitParams{
		LicensePlate: p.LicensePlate,
		DriverID:     p.DriverID,
		UserID:       p.UserID,
		Chassis:      p.Chassis,
		Brand:        p.Brand,
		Model:        p.Model,
		ManufactureYear: sql.NullInt64{
			Int64: p.ManufactureYear,
			Valid: true,
		},
		EnginePower: sql.NullString{
			String: p.EnginePower,
			Valid:  true,
		},
		UnitType: sql.NullString{
			String: p.UnitType,
			Valid:  true,
		},
		CanCouple: sql.NullBool{
			Bool:  p.CanCouple,
			Valid: true,
		},
		Height:   p.Height,
		State:    p.State,
		Renavan:  p.Renavan,
		Capacity: p.Capacity,
		Width:    p.Width,
		Length:   p.Length,
		Color:    p.Color,
	}
	return arg
}

func (p *UpdateTractorUnitRequest) ParseUpdateToTractorUnit() db.UpdateTractorUnitParams {
	arg := db.UpdateTractorUnitParams{
		LicensePlate: p.LicensePlate,
		DriverID:     p.DriverID,
		UserID:       p.UserID,
		Chassis:      p.Chassis,
		Brand:        p.Brand,
		Model:        p.Model,
		ManufactureYear: sql.NullInt64{
			Int64: p.ManufactureYear,
			Valid: true,
		},
		EnginePower: sql.NullString{
			String: p.EnginePower,
			Valid:  true,
		},
		UnitType: sql.NullString{
			String: p.UnitType,
			Valid:  true,
		},
		Height:   p.Height,
		State:    p.State,
		Renavan:  p.Renavan,
		Capacity: p.Capacity,
		Width:    p.Width,
		Length:   p.Length,
		Color:    p.Color,
		ID:       p.ID,
	}
	return arg
}

func (p *TractorUnitResponse) ParseFromTractorUnitObject(result db.TractorUnit) {
	p.ID = result.ID
	p.DriverID = result.DriverID
	p.UserID = result.UserID
	p.LicensePlate = result.LicensePlate
	p.Chassis = result.Chassis
	p.Brand = result.Brand
	p.Model = result.Model
	p.ManufactureYear = result.ManufactureYear.Int64
	p.EnginePower = result.EnginePower.String
	p.UnitType = result.UnitType.String
	p.Height = result.Height
	p.Status = result.Status
	p.CreatedAt = result.CreatedAt
	if result.UpdatedAt.Valid {
		p.UpdatedAt = &result.UpdatedAt.Time
	}
	p.State = result.State
	p.Renavan = result.Renavan
	p.Capacity = result.Capacity
	p.Width = result.Width
	p.Length = result.Length
	p.Color = result.Color
}
