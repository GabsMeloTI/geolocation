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
