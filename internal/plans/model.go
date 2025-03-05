package plans

import (
	db "geolocation/db/sqlc"
	"time"
)

type CreatePlansRequest struct {
	Name           string    `json:"name"`
	Price          string    `json:"price"`
	Duration       string    `json:"duration"`
	Annual         bool      `json:"annual"`
	Active         bool      `json:"active"`
	ExpirationDate time.Time `json:"expiration_date"`
}

type UpdatePlansRequest struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Price    string `json:"price"`
	Duration string `json:"duration"`
	Annual   bool   `json:"annual"`
}

type PlansResponse struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	Price          string    `json:"price"`
	Duration       string    `json:"duration"`
	Annual         bool      `json:"annual"`
	Active         bool      `json:"active"`
	ActiveDate     time.Time `json:"active_date"`
	ExpirationDate time.Time `json:"expiration_date"`
}

func (p *CreatePlansRequest) ParseCreateToPlans() db.CreatePlansParams {
	return db.CreatePlansParams{
		Name:           p.Name,
		Price:          p.Price,
		Duration:       p.Duration,
		Annual:         p.Annual,
		Active:         p.Active,
		ExpirationDate: p.ExpirationDate,
	}
}

func (p *UpdatePlansRequest) ParseUpdateToPlans() db.UpdatePlansParams {
	return db.UpdatePlansParams{
		ID:       p.ID,
		Name:     p.Name,
		Price:    p.Price,
		Duration: p.Duration,
		Annual:   p.Annual,
	}
}

func (p *PlansResponse) ParseFromPlansObject(result db.Plan) {
	p.ID = result.ID
	p.Name = result.Name
	p.Price = result.Price
	p.Duration = result.Duration
	p.Annual = result.Annual
	p.Active = result.Active
	p.ActiveDate = result.ActiveDate
	p.ExpirationDate = result.ExpirationDate
}
