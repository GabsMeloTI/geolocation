package login

import (
	"context"
	"database/sql"
	db "geolocation/db/sqlc"
)

type RepositoryInterface interface {
	GetUser(context.Context, db.LoginParams) (db.User, error)
}

type Repository struct {
	Conn    *sql.DB
	DBtx    db.DBTX
	Queries *db.Queries
	SqlConn *sql.DB
}

func NewRepository(conn *sql.DB) *Repository {
	q := db.New(conn)
	return &Repository{
		Conn:    conn,
		DBtx:    conn,
		Queries: q,
		SqlConn: conn,
	}
}

func (r *Repository) GetUser(ctx context.Context, arg db.LoginParams) (db.User, error) {
	return r.Queries.Login(ctx, arg)
}
