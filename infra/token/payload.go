package token

import (
	"errors"
	"github.com/google/uuid"
	"time"
)

var ErrExpiredToken = errors.New("token has expired")
var ErrInvalidToken = errors.New("token is invalid")

type Payload struct {
	ID             int64     `json:"id"`
	IP             string    `json:"ip"`
	NumberRequests int64     `json:"number_requests"`
	Valid          bool      `json:"valid"`
	ExpiredAt      time.Time `json:"expired_at"`
}

type PayloadSimp struct {
	ID           uuid.UUID `json:"id"`
	UserNickname string    `json:"user_nickname"`
	UserID       string    `json:"user_id"`
	AccessKey    int64     `json:"access_key"`
	AccessID     int64     `json:"access_id"`
	TenantID     string    `json:"tenant_id"`
	IssuedAt     time.Time `json:"issued_at"`
	ExpiredAt    time.Time `json:"expired_at"`
	Document     string    `json:"document"`
	UserOrgId    int64     `json:"user_org_id"`
	UserEmail    string    `json:"user_email"`
	UserName     string    `json:"user_name"`
}

func (payload *PayloadSimp) valid() error {
	if time.Now().After(payload.ExpiredAt) {
		return ErrExpiredToken
	}
	return nil
}

func NewPayload(tokenHistID int64, ip string, numberRequests int64, valid bool, expiredAt time.Duration) (*Payload, error) {
	payload := &Payload{
		ID:             tokenHistID,
		IP:             ip,
		NumberRequests: numberRequests,
		Valid:          valid,
		ExpiredAt:      time.Now().Add(expiredAt),
	}

	return payload, nil
}
