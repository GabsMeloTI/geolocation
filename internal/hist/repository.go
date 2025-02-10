package hist

import (
	"context"
	"database/sql"
	db "geolocation/db/sqlc"
)

type InterfaceRepository interface {
	CreateTokenHist(ctx context.Context, arg db.CreateTokenHistParams) (db.TokenHist, error)
	CreateRouteHist(ctx context.Context, arg db.CreateRouteHistParams) (db.RouteHist, error)
	UpdateNumberOfRequest(ctx context.Context, arg db.UpdateNumberOfRequestParams) error
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

func (r *Repository) CreateTokenHist(ctx context.Context, arg db.CreateTokenHistParams) (db.TokenHist, error) {
	return r.Queries.CreateTokenHist(ctx, arg)
}
func (r *Repository) CreateRouteHist(ctx context.Context, arg db.CreateRouteHistParams) (db.RouteHist, error) {
	return r.Queries.CreateRouteHist(ctx, arg)
}
func (r *Repository) UpdateNumberOfRequest(ctx context.Context, arg db.UpdateNumberOfRequestParams) error {
	return r.Queries.UpdateNumberOfRequest(ctx, arg)
}
