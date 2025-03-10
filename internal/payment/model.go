package payment

import (
	db "geolocation/db/sqlc"
	"geolocation/validation"
	"strconv"
	"time"
)

type CreatePaymentHistRequest struct {
	UserID           string    `json:"token"`
	Email            string    `json:"email"`
	Name             string    `json:"name"`
	Value            float64   `json:"value"`
	Method           string    `json:"method"`
	Automatic        bool      `json:"automatic"`
	PaymentDate      time.Time `json:"payment_date"`
	PaymentExpireted time.Time `json:"payment_expireted"`
	PaymentStatus    string    `json:"payment_status"`
	Currency         string    `json:"currency"`
	Invoice          string    `json:"invoice"`
	Customer         string    `json:"customer"`
	Interval         string    `json:"interval"`
}

type PaymentHistResponse struct {
	ID               int64     `json:"id"`
	UserID           string    `json:"user_id"`
	Email            string    `json:"email"`
	Name             string    `json:"name"`
	Value            float64   `json:"value"`
	Method           string    `json:"method"`
	Automatic        bool      `json:"automatic"`
	PaymentDate      time.Time `json:"payment_date"`
	PaymentExpireted time.Time `json:"payment_expireted"`
	PaymentStatus    string    `json:"payment_status"`
	Currency         string    `json:"currency"`
	Invoice          string    `json:"invoice"`
	Customer         string    `json:"customer"`
	Interval         string    `json:"interval"`
}

func (p *CreatePaymentHistRequest) ParseCreateToPaymentHist() db.CreatePaymentHistParams {
	idNumber, _ := validation.ParseStringToInt64(p.UserID)

	arg := db.CreatePaymentHistParams{
		UserID:           idNumber,
		Email:            p.Email,
		Name:             p.Name,
		Value:            p.Value,
		Method:           p.Method,
		Automatic:        p.Automatic,
		PaymentDate:      p.PaymentDate,
		PaymentExpireted: p.PaymentExpireted,
		PaymentStatus:    p.PaymentStatus,
		Currency:         p.Currency,
		Invoice:          p.Invoice,
		Customer:         p.Customer,
		Interval:         p.Interval,
	}
	return arg
}

func (p *PaymentHistResponse) ParseFromPaymentHistObject(result db.PaymentHist) {
	p.ID = result.ID
	p.UserID = strconv.FormatInt(result.UserID, 10)
	p.Email = result.Email
	p.Name = result.Name
	p.Value = result.Value
	p.Method = result.Method
	p.Automatic = result.Automatic
	p.PaymentDate = result.PaymentDate
	p.PaymentExpireted = result.PaymentExpireted
	p.PaymentStatus = result.PaymentStatus
	p.Currency = result.Currency
	p.Invoice = result.Invoice
	p.Customer = result.Customer
	p.Interval = result.Interval
}
