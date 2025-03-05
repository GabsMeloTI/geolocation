package trailer

import (
	"context"
	"database/sql"
	db "geolocation/db/sqlc"
)

type InterfaceRepository interface {
	CreateTrailer(ctx context.Context, arg db.CreateTrailerParams) (db.Trailer, error)
	UpdateTrailer(ctx context.Context, arg db.UpdateTrailerParams) (db.Trailer, error)
	DeleteTrailer(ctx context.Context, arg int64) error
	GetTrailerById(ctx context.Context, arg int64) (db.Trailer, error)
	GetTrailerByUserId(ctx context.Context, arg int64) ([]db.Trailer, error)
}
type Repository struct {
	Conn    *sql.DB
	DBtx    db.DBTX
	Queries *db.Queries
	SqlConn *sql.DB
}

func NewTrailersRepository(Conn *sql.DB) *Repository {
	q := db.New(Conn)
	return &Repository{
		Conn:    Conn,
		DBtx:    Conn,
		Queries: q,
		SqlConn: Conn,
	}
}

func (r *Repository) CreateTrailer(ctx context.Context, arg db.CreateTrailerParams) (db.Trailer, error) {
	return r.Queries.CreateTrailer(ctx, arg)
}
func (r *Repository) UpdateTrailer(ctx context.Context, arg db.UpdateTrailerParams) (db.Trailer, error) {
	return r.Queries.UpdateTrailer(ctx, arg)
}
func (r *Repository) DeleteTrailer(ctx context.Context, arg int64) error {
	return r.Queries.DeleteTrailer(ctx, arg)
}
func (r *Repository) GetTrailerById(ctx context.Context, arg int64) (db.Trailer, error) {
	return r.Queries.GetTrailerById(ctx, arg)
}
func (r *Repository) GetTrailerByUserId(ctx context.Context, arg int64) ([]db.Trailer, error) {
	return r.Queries.GetTrailerByUserId(ctx, arg)
}
