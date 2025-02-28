package appointments

import (
	"database/sql"
	db "geolocation/db/sqlc"
	"geolocation/internal/get_token"
	"time"
)

type CreateAppointmentRequest struct {
	UserID          int64 `json:"user_id"`
	TruckID         int64 `json:"truck_id"`
	AdvertisementID int64 `json:"advertisement_id"`
}

type CreateAppointmentDTO struct {
	Request CreateAppointmentRequest
	Payload get_token.PayloadUserDTO
}

type UpdateAppointmentRequest struct {
	ID        int64  `json:"id"`
	Situation string `json:"situation"`
}

type UpdateAppointmentDTO struct {
	Request UpdateAppointmentRequest
	Payload get_token.PayloadUserDTO
}

//Todo: descobrir quais campos aparecem no agendamento

type AppointmentResponseList struct {
	ID              int64      `json:"id"`
	UserID          int64      `json:"user_id"`
	TruckID         int64      `json:"truck_id"`
	AdvertisementID int64      `json:"advertisement_id"`
	Status          bool       `json:"status"`
	CreatedWho      string     `json:"created_who"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedWho      string     `json:"updated_who"`
	UpdatedAt       *time.Time `json:"updated_at"`
}

type AppointmentResponse struct {
	ID              int64      `json:"id"`
	UserID          int64      `json:"user_id"`
	TruckID         int64      `json:"truck_id"`
	AdvertisementID int64      `json:"advertisement_id"`
	Status          bool       `json:"status"`
	CreatedWho      string     `json:"created_who"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedWho      string     `json:"updated_who"`
	UpdatedAt       *time.Time `json:"updated_at"`
}

func (p *CreateAppointmentDTO) ParseCreateToAppointment() db.CreateAppointmentParams {
	arg := db.CreateAppointmentParams{
		UserID:          p.Request.UserID,
		TruckID:         p.Request.TruckID,
		AdvertisementID: p.Request.AdvertisementID,
		CreatedWho:      p.Payload.Name,
	}
	return arg
}

func (p *UpdateAppointmentDTO) ParseUpdateToAppointment() db.UpdateAppointmentSituationParams {
	arg := db.UpdateAppointmentSituationParams{
		Situation: p.Request.Situation,
		UpdatedWho: sql.NullString{
			String: p.Payload.Name,
			Valid:  true,
		},
		ID: p.Request.ID,
	}
	return arg
}

func (p *AppointmentResponse) ParseFromAppointmentObject(result db.Appointment) {
	p.ID = result.ID
	p.UserID = result.UserID
	p.TruckID = result.TruckID
	p.AdvertisementID = result.AdvertisementID
	p.Status = result.Status
	p.CreatedWho = result.CreatedWho
	p.CreatedAt = result.CreatedAt
	p.UpdatedWho = result.UpdatedWho.String
	if result.UpdatedAt.Valid {
		p.UpdatedAt = &result.UpdatedAt.Time
	}
}

func (p *AppointmentResponseList) ParseFromAppointmentListObject(result db.GetListAppointmentByUserIDRow) {
	p.ID = result.ID
	p.UserID = result.UserID
	p.TruckID = result.TruckID
	p.AdvertisementID = result.AdvertisementID
	p.Status = result.Status
	p.CreatedWho = result.CreatedWho
	p.CreatedAt = result.CreatedAt
	p.UpdatedWho = result.UpdatedWho.String
	if result.UpdatedAt.Valid {
		p.UpdatedAt = &result.UpdatedAt.Time
	}
}
