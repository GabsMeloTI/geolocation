package announcement

import (
	"context"
	"database/sql"
	db "geolocation/db/sqlc"
)

type InterfaceRepository interface {
	CreateAnnouncement(ctx context.Context, arg db.CreateAnnouncementParams) (db.Announcement, error)
	UpdateAnnouncement(ctx context.Context, arg db.UpdateAnnouncementParams) (db.Announcement, error)
	DeleteAnnouncement(ctx context.Context, arg db.DeleteAnnouncementParams) error
	GetAnnouncementById(ctx context.Context, arg int64) (db.Announcement, error)
	GetAllAnnouncementUsers(ctx context.Context) (db.Announcement, error)
	GetAllAnnouncementPublic(ctx context.Context) (db.GetAllAnnouncementPublicRow, error)
}
type Repository struct {
	Conn    *sql.DB
	DBtx    db.DBTX
	Queries *db.Queries
	SqlConn *sql.DB
}

func NewAnnouncementsRepository(Conn *sql.DB) *Repository {
	q := db.New(Conn)
	return &Repository{
		Conn:    Conn,
		DBtx:    Conn,
		Queries: q,
		SqlConn: Conn,
	}
}

func (r *Repository) CreateAnnouncement(ctx context.Context, arg db.CreateAnnouncementParams) (db.Announcement, error) {
	return r.Queries.CreateAnnouncement(ctx, arg)
}
func (r *Repository) UpdateAnnouncement(ctx context.Context, arg db.UpdateAnnouncementParams) (db.Announcement, error) {
	return r.Queries.UpdateAnnouncement(ctx, arg)
}
func (r *Repository) DeleteAnnouncement(ctx context.Context, arg db.DeleteAnnouncementParams) error {
	return r.Queries.DeleteAnnouncement(ctx, arg)
}
func (r *Repository) GetAnnouncementById(ctx context.Context, arg int64) (db.Announcement, error) {
	return r.Queries.GetAnnouncementById(ctx, arg)
}
func (r *Repository) GetAllAnnouncementUsers(ctx context.Context) (db.Announcement, error) {
	return r.Queries.GetAllAnnouncementUsers(ctx)
}
func (r *Repository) GetAllAnnouncementPublic(ctx context.Context) (db.GetAllAnnouncementPublicRow, error) {
	return r.Queries.GetAllAnnouncementPublic(ctx)
}
