package get_token

import (
	"github.com/google/uuid"
	"time"
)

type PayloadDTO struct {
	ID           uuid.UUID `json:"id"`
	UserID       string    `json:"user_id"`
	UserNickname string    `json:"user_nickname"`
	ExpiryAt     time.Time `json:"expiry_at"`
	AccessKey    int64     `json:"access_key"`
	AccessID     int64     `json:"access_id"`
	TenantID     uuid.UUID `json:"tenant_id"`
	UserOrgId    int64     `json:"user_org_id"`
	UserEmail    string    `json:"user_email"`
	Document     string    `json:"document"`
	UserName     string    `json:"user_name"`
}

type PublicPayloadDTO struct {
	ID             int64     `json:"id"`
	IP             string    `json:"ip"`
	NumberRequests int64     `json:"number_requests"`
	Valid          bool      `json:"valid"`
	ExpiredAt      time.Time `json:"expired_at"`
}

type PayloadUserDTO struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	ProfileID int64     `json:"profile_id"`
	Document  string    `json:"document"`
	GoogleID  string    `json:"google_id"`
	ExpireAt  time.Time `json:"expire_at"`
}
