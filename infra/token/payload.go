package token

import (
	"errors"
	"github.com/google/uuid"
	"time"
)

var ErrExpiredToken = errors.New("token has expired")
var ErrInvalidToken = errors.New("token is invalid")
var ErrLimitRequests = errors.New("you have reached the limit of requests per day")

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

type PayloadUser struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	ProfileID int64     `json:"profile_id"`
	Document  string    `json:"document"`
	GoogleID  string    `json:"google_id"`
	ExpireAt  time.Time `json:"expire_at"`
}

func (payload *PayloadSimp) valid() error {
	if time.Now().After(payload.ExpiredAt) {
		return ErrExpiredToken
	}
	return nil
}
func (payload *Payload) validPublic() error {
	if time.Now().After(payload.ExpiredAt) {
		return ErrExpiredToken
	} else if payload.NumberRequests >= 5 {
		return ErrLimitRequests
	} else if !payload.Valid {
		return ErrInvalidToken
	}
	return nil
}

func (payload *PayloadUser) valid() error {
	if time.Now().After(payload.ExpireAt) {
		return ErrExpiredToken
	}
	return nil
}

func NewPayload(tokenHistID int64, ip string, numberRequests int64, valid bool, expiredAt time.Time) (*Payload, error) {
	payload := &Payload{
		ID:             tokenHistID,
		IP:             ip,
		NumberRequests: numberRequests,
		Valid:          valid,
		ExpiredAt:      expiredAt,
	}

	return payload, nil
}

func NewPayloadUser(id int64, name string, email string, profileId int64, document string, googleId string, expireAt time.Time) (*PayloadUser, error) {
	return &PayloadUser{
		ID:        id,
		Name:      name,
		Email:     email,
		ProfileID: profileId,
		Document:  document,
		GoogleID:  googleId,
		ExpireAt:  expireAt,
	}, nil
}
