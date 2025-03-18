package tractor_unit

import (
	"context"
	"database/sql"
	db "geolocation/db/sqlc"
)

type InterfaceRepository interface {
	CreateTractorUnit(ctx context.Context, arg db.CreateTractorUnitParams) (db.TractorUnit, error)
	UpdateTractorUnit(ctx context.Context, arg db.UpdateTractorUnitParams) (db.TractorUnit, error)
	DeleteTractorUnit(ctx context.Context, arg db.DeleteTractorUnitParams) error
	GetTractorUnitById(ctx context.Context, arg int64) (db.TractorUnit, error)
	GetTractorUnitByUserId(ctx context.Context, arg int64) ([]db.TractorUnit, error)
	GetOneTractorUnitByUserId(ctx context.Context, arg int64) (db.TractorUnit, error)
}

type Repository struct {
	Conn    *sql.DB
	DBtx    db.DBTX
	Queries *db.Queries
	SqlConn *sql.DB
}

func NewTractorUnitsRepository(Conn *sql.DB) *Repository {
	q := db.New(Conn)
	return &Repository{
		Conn:    Conn,
		DBtx:    Conn,
		Queries: q,
		SqlConn: Conn,
	}
}

func (r *Repository) CreateTractorUnit(ctx context.Context, arg db.CreateTractorUnitParams) (db.TractorUnit, error) {
	return r.Queries.CreateTractorUnit(ctx, arg)
}
func (r *Repository) UpdateTractorUnit(ctx context.Context, arg db.UpdateTractorUnitParams) (db.TractorUnit, error) {
	return r.Queries.UpdateTractorUnit(ctx, arg)
}
func (r *Repository) DeleteTractorUnit(ctx context.Context, arg db.DeleteTractorUnitParams) error {
	return r.Queries.DeleteTractorUnit(ctx, arg)
}
func (r *Repository) GetTractorUnitById(ctx context.Context, arg int64) (db.TractorUnit, error) {
	return r.Queries.GetTractorUnitById(ctx, arg)
}
func (r *Repository) GetTractorUnitByUserId(ctx context.Context, arg int64) ([]db.TractorUnit, error) {
	return r.Queries.GetTractorUnitByUserId(ctx, arg)
}
func (r *Repository) GetOneTractorUnitByUserId(ctx context.Context, arg int64) (db.TractorUnit, error) {
	return r.Queries.GetOneTractorUnitByUserId(ctx, arg)
}
