package zonas_risco

import (
	"context"
	"database/sql"
	db "geolocation/db/sqlc"
)

type InterfaceRepository interface {
	CreateZonaRisco(ctx context.Context, arg db.CreateZonaRiscoParams) (db.ZonasRisco, error)
	UpdateZonaRisco(ctx context.Context, arg db.UpdateZonaRiscoParams) (db.ZonasRisco, error)
	DeleteZonaRisco(ctx context.Context, id int64) error
	GetZonaRiscoById(ctx context.Context, id int64) (db.ZonasRisco, error)
	GetAllZonasRisco(ctx context.Context, organization_id sql.NullInt64) ([]db.ZonasRisco, error)
}

type Repository struct {
	Conn    *sql.DB
	DBtx    db.DBTX
	Queries *db.Queries
	SqlConn *sql.DB
}

func NewZonasRiscoRepository(Conn *sql.DB) *Repository {
	q := db.New(Conn)
	return &Repository{
		Conn:    Conn,
		DBtx:    Conn,
		Queries: q,
		SqlConn: Conn,
	}
}

func (r *Repository) CreateZonaRisco(ctx context.Context, arg db.CreateZonaRiscoParams) (db.ZonasRisco, error) {
	return r.Queries.CreateZonaRisco(ctx, arg)
}

func (r *Repository) UpdateZonaRisco(ctx context.Context, arg db.UpdateZonaRiscoParams) (db.ZonasRisco, error) {
	return r.Queries.UpdateZonaRisco(ctx, arg)
}

func (r *Repository) DeleteZonaRisco(ctx context.Context, id int64) error {
	return r.Queries.DeleteZonaRisco(ctx, id)
}

func (r *Repository) GetZonaRiscoById(ctx context.Context, id int64) (db.ZonasRisco, error) {
	return r.Queries.GetZonaRiscoById(ctx, id)
}

func (r *Repository) GetAllZonasRisco(ctx context.Context, organization_id sql.NullInt64) ([]db.ZonasRisco, error) {
	return r.Queries.GetAllZonasRisco(ctx, organization_id)
}
