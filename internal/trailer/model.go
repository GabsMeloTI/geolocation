package trailer

import (
	"database/sql"
	db "geolocation/db/sqlc"
	"time"
)

type CreateTrailerRequest struct {
	UserId       int64   `json:"user_id"`
	LicensePlate string  `json:"license_plate"`
	Chassis      string  `json:"chassis"`
	BodyType     string  `json:"body_type"`
	LoadCapacity float64 `json:"load_capacity"`
	Length       float64 `json:"length"`
	Width        float64 `json:"width"`
	Height       float64 `json:"height"`
}

type UpdateTrailerRequest struct {
	ID           int64   `json:"id"`
	UserId       int64   `json:"userId"`
	LicensePlate string  `json:"license_plate"`
	Chassis      string  `json:"chassis"`
	BodyType     string  `json:"body_type" validate:"oneof=open chest bulk_carrier sider"`
	LoadCapacity float64 `json:"load_capacity"`
	Length       float64 `json:"length"`
	Width        float64 `json:"width"`
	Height       float64 `json:"height"`
}

type TrailerResponse struct {
	ID           int64      `json:"id"`
	UserId       int64      `json:"userId"`
	LicensePlate string     `json:"license_plate"`
	Chassis      string     `json:"chassis"`
	BodyType     string     `json:"body_type"`
	LoadCapacity float64    `json:"load_capacity"`
	Length       float64    `json:"length"`
	Width        float64    `json:"width"`
	Height       float64    `json:"height"`
	Status       bool       `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
}

func (p *CreateTrailerRequest) ParseCreateToTrailer() db.CreateTrailerParams {
	arg := db.CreateTrailerParams{
		UserID:       p.UserId,
		LicensePlate: p.LicensePlate,
		Chassis:      p.Chassis,
		BodyType: sql.NullString{
			String: p.BodyType,
			Valid:  true,
		},
		LoadCapacity: sql.NullFloat64{
			Float64: p.LoadCapacity,
			Valid:   true,
		},
		Length: sql.NullFloat64{
			Float64: p.Length,
			Valid:   true,
		},
		Width: sql.NullFloat64{
			Float64: p.Height,
			Valid:   true,
		},
		Height: sql.NullFloat64{
			Float64: p.Width,
			Valid:   true,
		},
	}
	return arg
}

func (p *UpdateTrailerRequest) ParseUpdateToTrailer() db.UpdateTrailerParams {
	arg := db.UpdateTrailerParams{
		ID:           p.ID,
		UserID:       p.UserId,
		LicensePlate: p.LicensePlate,
		Chassis:      p.Chassis,
		BodyType: sql.NullString{
			String: p.BodyType,
			Valid:  true,
		},
		LoadCapacity: sql.NullFloat64{
			Float64: p.LoadCapacity,
			Valid:   true,
		},
		Length: sql.NullFloat64{
			Float64: p.Length,
			Valid:   true,
		},
		Width: sql.NullFloat64{
			Float64: p.Height,
			Valid:   true,
		},
		Height: sql.NullFloat64{
			Float64: p.Width,
			Valid:   true,
		},
	}
	return arg
}

func (p *TrailerResponse) ParseFromTrailerObject(result db.Trailer) {
	p.ID = result.ID
	p.UserId = result.UserID
	p.LicensePlate = result.LicensePlate
	p.Chassis = result.Chassis
	p.BodyType = result.BodyType.String
	p.LoadCapacity = result.LoadCapacity.Float64
	p.Length = result.Length.Float64
	p.Width = result.Width.Float64
	p.Height = result.Height.Float64
	p.Status = result.Status
	p.CreatedAt = result.CreatedAt
	if result.UpdatedAt.Valid {
		p.UpdatedAt = &result.UpdatedAt.Time
	}
}
