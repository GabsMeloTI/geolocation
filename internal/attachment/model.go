package attachment

import "github.com/google/uuid"

type AttachRequestCreate struct {
	Description string `form:"description"`
	Type        string `form:"type"`
}

type DeleteAttachByCodeIdAndOrigin struct {
	CodeId     int64     `json:"code_id"`
	Origin     string    `json:"origin"`
	CreatedWho string    `json:"created_who"`
	AccessID   int64     `json:"access_id"`
	TenantID   uuid.UUID `json:"tenant_id"`
}

type Attachment struct {
	UserID int64  `db:"user_id"`
	URL    string `db:"url"`
}
