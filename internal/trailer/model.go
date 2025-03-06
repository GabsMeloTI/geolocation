package trailer

import (
	"database/sql"
	db "geolocation/db/sqlc"
	"time"
)

type CreateTrailerRequest struct {
	LicensePlate string  `json:"license_plate"`
	Chassis      string  `json:"chassis"`
	BodyType     string  `json:"body_type"`
	LoadCapacity float64 `json:"load_capacity"`
	Length       float64 `json:"length"`
	Width        float64 `json:"width"`
	Height       float64 `json:"height"`
	State        string  `json:"state"`
	Renavan      string  `json:"renavan"`
	Axles        int64   `json:"axles"`
}

type CreateTrailerDto struct {
	CreateTrailerRequest CreateTrailerRequest
	UserID               int64 `json:"user_id"`
}

type UpdateTrailerRequest struct {
	ID           int64   `json:"id"`
	LicensePlate string  `json:"license_plate"`
	Chassis      string  `json:"chassis"`
	BodyType     string  `json:"body_type" validate:"oneof=open chest bulk_carrier sider"`
	LoadCapacity float64 `json:"load_capacity"`
	Length       float64 `json:"length"`
	Width        float64 `json:"width"`
	Height       float64 `json:"height"`
	State        string  `json:"state"`
	Renavan      string  `json:"renavan"`
	Axles        int64   `json:"axles"`
}

type UpdateTrailerDto struct {
	UpdateTrailerRequest UpdateTrailerRequest
	UserID               int64 `json:"user_id"`
}

type TrailerResponse struct {
	ID           int64      `json:"id"`
	UserId       int64      `json:"user_id"`
	LicensePlate string     `json:"license_plate"`
	Chassis      string     `json:"chassis"`
	BodyType     string     `json:"body_type"`
	LoadCapacity float64    `json:"load_capacity"`
	Length       float64    `json:"length"`
	Width        float64    `json:"width"`
	Height       float64    `json:"height"`
	State        string     `json:"state"`
	Renavan      string     `json:"renavan"`
	Axles        int64      `json:"axles"`
	Status       bool       `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
}

func (p *CreateTrailerDto) ParseCreateToTrailer() db.CreateTrailerParams {
	arg := db.CreateTrailerParams{
		LicensePlate: p.CreateTrailerRequest.LicensePlate,
		UserID:       p.UserID,
		Chassis:      p.CreateTrailerRequest.Chassis,
		BodyType: sql.NullString{
			String: p.CreateTrailerRequest.BodyType,
			Valid:  true,
		},
		LoadCapacity: sql.NullFloat64{
			Float64: p.CreateTrailerRequest.LoadCapacity,
			Valid:   true,
		},
		Length: sql.NullFloat64{
			Float64: p.CreateTrailerRequest.Length,
			Valid:   true,
		},
		Width: sql.NullFloat64{
			Float64: p.CreateTrailerRequest.Height,
			Valid:   true,
		},
		Height: sql.NullFloat64{
			Float64: p.CreateTrailerRequest.Width,
			Valid:   true,
		},
		Axles:   p.CreateTrailerRequest.Axles,
		State:   p.CreateTrailerRequest.State,
		Renavan: p.CreateTrailerRequest.Renavan,
	}
	return arg
}

func (p *UpdateTrailerDto) ParseUpdateToTrailer() db.UpdateTrailerParams {
	arg := db.UpdateTrailerParams{
		LicensePlate: p.UpdateTrailerRequest.LicensePlate,
		Chassis:      p.UpdateTrailerRequest.Chassis,
		BodyType: sql.NullString{
			String: p.UpdateTrailerRequest.BodyType,
			Valid:  true,
		},
		LoadCapacity: sql.NullFloat64{
			Float64: p.UpdateTrailerRequest.LoadCapacity,
			Valid:   true,
		},
		Length: sql.NullFloat64{
			Float64: p.UpdateTrailerRequest.Length,
			Valid:   true,
		},
		Width: sql.NullFloat64{
			Float64: p.UpdateTrailerRequest.Height,
			Valid:   true,
		},
		Height: sql.NullFloat64{
			Float64: p.UpdateTrailerRequest.Width,
			Valid:   true,
		},
		Axles:   p.UpdateTrailerRequest.Axles,
		UserID:  p.UserID,
		State:   p.UpdateTrailerRequest.State,
		Renavan: p.UpdateTrailerRequest.Renavan,
		ID:      p.UpdateTrailerRequest.ID,
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
	p.Axles = result.Axles
	p.State = result.State
	p.Renavan = result.Renavan
	p.Status = result.Status
	p.CreatedAt = result.CreatedAt
	if result.UpdatedAt.Valid {
		p.UpdatedAt = &result.UpdatedAt.Time
	}
}
