package hist

import (
	"context"
	"database/sql"
	db "geolocation/db/sqlc"
)

type InterfaceRepository interface {
	CreateTokenHist(ctx context.Context, arg db.CreateTokenHistParams) (db.TokenHist, error)
}

type Repository struct {
	Conn    *sql.DB
	DBtx    db.DBTX
	Queries *db.Queries
	SqlConn *sql.DB
}

func NewHistRepository(Conn *sql.DB) *Repository {
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
