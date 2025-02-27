package routes

import (
	"context"
	"database/sql"
	db "geolocation/db/sqlc"
	"strconv"
)

type InterfaceRepository interface {
	GetTollsByLonAndLat(ctx context.Context) ([]db.Toll, error)
	CreateGasStations(ctx context.Context, arg db.CreateGasStationsParams) (db.GasStation, error)
	GetGasStation(ctx context.Context, arg db.GetGasStationParams) ([]db.GetGasStationRow, error)
	GetTollTags(ctx context.Context) ([]db.TollTag, error)
	CreateSavedRoutes(ctx context.Context, arg db.CreateSavedRoutesParams) (db.SavedRoute, error)
	GetSavedRoutes(ctx context.Context, arg db.GetSavedRoutesParams) (db.SavedRoute, error)
	GetSavedRouteById(ctx context.Context, arg int32) (db.SavedRoute, error)
	GetBalanca(ctx context.Context) ([]db.Balanca, error)
	CreateFavoriteRoute(ctx context.Context, arg db.CreateFavoriteRouteParams) (db.FavoriteRoute, error)
	GetTokenHist(ctx context.Context, arg int64) (db.TokenHist, error)
	UpdateNumberOfRequest(ctx context.Context, arg db.UpdateNumberOfRequestParams) error
	CreateRouteHist(ctx context.Context, arg db.CreateRouteHistParams) (db.RouteHist, error)
	GetFreightLoadAll(ctx context.Context) ([]db.FreightLoad, error)
	GetGasStationsByBoundingBox(ctx context.Context, arg db.GetGasStationsByBoundingBoxParams) ([]db.GetGasStationsByBoundingBoxRow, error)
	GetRouteHistByUnique(ctx context.Context, arg db.GetRouteHistByUniqueParams) (db.RouteHist, error)
	UpdateNumberOfRequestRequest(ctx context.Context, arg db.UpdateNumberOfRequestParams) error
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
func (r *Repository) CreateSavedRoutes(ctx context.Context, arg db.CreateSavedRoutesParams) (db.SavedRoute, error) {
	return r.Queries.CreateSavedRoutes(ctx, arg)
}
func (r *Repository) GetSavedRoutes(ctx context.Context, arg db.GetSavedRoutesParams) (db.SavedRoute, error) {
	return r.Queries.GetSavedRoutes(ctx, arg)
}
func (r *Repository) GetSavedRouteById(ctx context.Context, arg int32) (db.SavedRoute, error) {
	return r.Queries.GetSavedRouteById(ctx, arg)
}
func (r *Repository) GetBalanca(ctx context.Context) ([]db.Balanca, error) {
	return r.Queries.GetBalanca(ctx)
}
func (r *Repository) CreateFavoriteRoute(ctx context.Context, arg db.CreateFavoriteRouteParams) (db.FavoriteRoute, error) {
	return r.Queries.CreateFavoriteRoute(ctx, arg)
}
func (r *Repository) GetTokenHist(ctx context.Context, arg int64) (db.TokenHist, error) {
	return r.Queries.GetTokenHist(ctx, strconv.FormatInt(arg, 10))
}
func (r *Repository) UpdateNumberOfRequest(ctx context.Context, arg db.UpdateNumberOfRequestParams) error {
	return r.Queries.UpdateNumberOfRequest(ctx, arg)
}
func (r *Repository) CreateRouteHist(ctx context.Context, arg db.CreateRouteHistParams) (db.RouteHist, error) {
	return r.Queries.CreateRouteHist(ctx, arg)
}
func (r *Repository) GetFreightLoadAll(ctx context.Context) ([]db.FreightLoad, error) {
	return r.Queries.GetFreightLoadAll(ctx)
}
func (r *Repository) GetGasStationsByBoundingBox(ctx context.Context, arg db.GetGasStationsByBoundingBoxParams) ([]db.GetGasStationsByBoundingBoxRow, error) {
	return r.Queries.GetGasStationsByBoundingBox(ctx, arg)
}
func (r *Repository) UpdateNumberOfRequestRequest(ctx context.Context, arg db.UpdateNumberOfRequestParams) error {
	return r.Queries.UpdateNumberOfRequest(ctx, arg)
}
func (r *Repository) GetRouteHistByUnique(ctx context.Context, arg db.GetRouteHistByUniqueParams) (db.RouteHist, error) {
	return r.Queries.GetRouteHistByUnique(ctx, arg)
}
