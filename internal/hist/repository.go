package hist

import (
	"context"
	"database/sql"
	db "geolocation/db/sqlc"
)

type InterfaceRepository interface {
	CreateTokenHist(ctx context.Context, arg db.CreateTokenHistParams) (db.TokenHist, error)
	GetTokenHistExist(ctx context.Context, arg string) (bool, error)
	GetTokenHist(ctx context.Context, arg string) (db.TokenHist, error)
	UpdateTokenHist(ctx context.Context, arg db.UpdateTokenHistParams) (db.UpdateTokenHistRow, error)
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
func (r *Repository) GetTokenHistExist(ctx context.Context, arg string) (bool, error) {
	return r.Queries.GetTokenHistExist(ctx, arg)
}
func (r *Repository) GetTokenHist(ctx context.Context, arg string) (db.TokenHist, error) {
	return r.Queries.GetTokenHist(ctx, arg)
}
func (r *Repository) UpdateTokenHist(ctx context.Context, arg db.UpdateTokenHistParams) (db.UpdateTokenHistRow, error) {
	return r.Queries.UpdateTokenHist(ctx, arg)
}
