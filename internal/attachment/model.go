package attachment

import "github.com/google/uuid"

type AttachRequestCreate struct {
	UserId      string `form:"user_id"`
	Description string `form:"description"`
}

type DeleteAttachByCodeIdAndOrigin struct {
	CodeId     int64     `json:"code_id"`
	Origin     string    `json:"origin"`
	CreatedWho string    `json:"created_who"`
	AccessID   int64     `json:"access_id"`
	TenantID   uuid.UUID `json:"tenant_id"`
}
