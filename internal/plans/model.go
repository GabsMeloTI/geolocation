package plans

import (
	db "geolocation/db/sqlc"
	"time"
)

type CreateUserPlanRequest struct {
	IDUser         int64     `json:"id_user"`
	IDPlan         int64     `json:"id_plan"`
	Annual         bool      `json:"annual"`
	ExpirationDate time.Time `json:"expiration_date"`
	Price          float64   `json:"price"`
}

type UserPlanResponse struct {
	ID             int64     `json:"id"`
	IDUser         int64     `json:"id_user"`
	IDPlan         int64     `json:"id_plan"`
	Annual         bool      `json:"annual"`
	Active         bool      `json:"active"`
	ActiveDate     time.Time `json:"active_date"`
	ExpirationDate time.Time `json:"expiration_date"`
	Price          float64   `json:"price"`
}

func (p *CreateUserPlanRequest) ParseCreateToUserPlan() db.CreateUserPlansParams {
	arg := db.CreateUserPlansParams{
		IDUser:         p.IDUser,
		IDPlan:         p.IDPlan,
		Annual:         p.Annual,
		ExpirationDate: p.ExpirationDate,
	}
	return arg
}

func (p *UserPlanResponse) ParseFromPlansObject(result db.UserPlan) {
	p.ID = result.ID
	p.IDUser = result.IDUser
	p.IDPlan = result.IDPlan
	p.Annual = result.Annual
	p.Active = result.Active
	p.ActiveDate = result.ActiveDate
	p.ExpirationDate = result.ExpirationDate
}
