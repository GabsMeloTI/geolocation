package routes

import (
	"context"
	"database/sql"
	db "geolocation/db/sqlc"
)

type InterfaceRepository interface {
	CreateTolls(ctx context.Context, arg db.CreateTollsParams) error
	GetTollsByLonAndLat(ctx context.Context) ([]db.Toll, error)
	CreateGasStations(ctx context.Context, arg db.CreateGasStationsParams) (db.GasStation, error)
	GetGasStation(ctx context.Context, arg db.GetGasStationParams) ([]db.GetGasStationRow, error)
	GetTollTags(ctx context.Context) ([]db.TollTag, error)
}

type Repository struct {
	Conn    *sql.DB
	DBtx    db.DBTX
	Queries *db.Queries
	SqlConn *sql.DB
}

func NewTollsRepository(Conn *sql.DB) *Repository {
	q := db.New(Conn)
	return &Repository{
		Conn:    Conn,
		DBtx:    Conn,
		Queries: q,
		SqlConn: Conn,
	}
}

func (r *Repository) CreateTolls(ctx context.Context, arg db.CreateTollsParams) error {
	return r.Queries.CreateTolls(ctx, arg)
}
func (r *Repository) GetTollsByLonAndLat(ctx context.Context) ([]db.Toll, error) {
	return r.Queries.GetTollsByLonAndLat(ctx)
}
func (r *Repository) CreateGasStations(ctx context.Context, arg db.CreateGasStationsParams) (db.GasStation, error) {
	return r.Queries.CreateGasStations(ctx, arg)
}
func (r *Repository) GetGasStation(ctx context.Context, arg db.GetGasStationParams) ([]db.GetGasStationRow, error) {
	return r.Queries.GetGasStation(ctx, arg)
}
func (r *Repository) GetTollTags(ctx context.Context) ([]db.TollTag, error) {
	return r.Queries.GetTollTags(ctx)
}
