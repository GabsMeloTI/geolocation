package attachment

import (
	"context"
	"database/sql"
	db "geolocation/db/sqlc"
)

type InterfaceRepository interface {
	GetAttachmentById(ctx context.Context, arg db.GetAttachmentByIdParams) (db.Attachment, error)
	CreateAttachments(ctx context.Context, arg db.CreateAttachmentsParams) (db.Attachment, error)
	UpdateAttachmentLogicDelete(ctx context.Context, arg db.UpdateAttachmentLogicDeleteParams) error
	UpdateProfilePictureByUserId(ctx context.Context, arg db.UpdateProfilePictureByUserIdParams) error
	GetAllAttachmentById(ctx context.Context, arg db.GetAllAttachmentByIdParams) ([]db.Attachment, error)
}
type Repository struct {
	Conn    *sql.DB
	DBtx    db.DBTX
	Queries *db.Queries
	SqlConn *sql.DB
}

func NewAttachmentRepository(Conn *sql.DB) *Repository {
	q := db.New(Conn)
	return &Repository{
		Conn:    Conn,
		DBtx:    Conn,
		Queries: q,
		SqlConn: Conn,
	}
}

func (r *Repository) GetAttachmentById(ctx context.Context, arg db.GetAttachmentByIdParams) (db.Attachment, error) {
	return r.Queries.GetAttachmentById(ctx, arg)
}
func (r *Repository) CreateAttachments(ctx context.Context, arg db.CreateAttachmentsParams) (db.Attachment, error) {
	return r.Queries.CreateAttachments(ctx, arg)
}
func (r *Repository) UpdateAttachmentLogicDelete(ctx context.Context, arg db.UpdateAttachmentLogicDeleteParams) error {
	return r.Queries.UpdateAttachmentLogicDelete(ctx, arg)
}
func (r *Repository) UpdateProfilePictureByUserId(ctx context.Context, arg db.UpdateProfilePictureByUserIdParams) error {
	return r.Queries.UpdateProfilePictureByUserId(ctx, arg)
}
func (r *Repository) GetAllAttachmentById(ctx context.Context, arg db.GetAllAttachmentByIdParams) ([]db.Attachment, error) {
	return r.Queries.GetAllAttachmentById(ctx, arg)
}
