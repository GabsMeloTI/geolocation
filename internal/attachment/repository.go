package attachment

import (
	"context"
	"database/sql"
	db "geolocation/db/sqlc"
)

type InterfaceRepository interface {
	GetAttachmentById(ctx context.Context, arg int64) (db.Attachment, error)
	CreateAttachments(ctx context.Context, arg db.CreateAttachmentsParams) (db.Attachment, error)
	UpdateAttachmentLogicDelete(ctx context.Context, arg int64) error
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

func (r *Repository) GetAttachmentById(ctx context.Context, arg int64) (db.Attachment, error) {
	return r.Queries.GetAttachmentById(ctx, arg)
}
func (r *Repository) CreateAttachments(ctx context.Context, arg db.CreateAttachmentsParams) (db.Attachment, error) {
	return r.Queries.CreateAttachments(ctx, arg)
}
func (r *Repository) UpdateAttachmentLogicDelete(ctx context.Context, arg int64) error {
	return r.Queries.UpdateAttachmentLogicDelete(ctx, arg)
}
