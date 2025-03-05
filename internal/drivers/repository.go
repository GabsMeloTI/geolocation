package drivers

import (
	"context"
	"database/sql"
	db "geolocation/db/sqlc"
)

type InterfaceRepository interface {
	CreateDriver(ctx context.Context, arg db.CreateDriverParams) (db.Driver, error)
	UpdateDriver(ctx context.Context, arg db.UpdateDriverParams) (db.Driver, error)
	DeleteDriver(ctx context.Context, arg int64) error
	GetDriverById(ctx context.Context, arg int64) (db.Driver, error)
	GetDriverByUserId(ctx context.Context, arg int64) (db.Driver, error)
}
type Repository struct {
	Conn    *sql.DB
	DBtx    db.DBTX
	Queries *db.Queries
	SqlConn *sql.DB
}

func NewDriversRepository(Conn *sql.DB) *Repository {
	q := db.New(Conn)
	return &Repository{
		Conn:    Conn,
		DBtx:    Conn,
		Queries: q,
		SqlConn: Conn,
	}
}

func (r *Repository) CreateDriver(ctx context.Context, arg db.CreateDriverParams) (db.Driver, error) {
	return r.Queries.CreateDriver(ctx, arg)
}
func (r *Repository) UpdateDriver(ctx context.Context, arg db.UpdateDriverParams) (db.Driver, error) {
	return r.Queries.UpdateDriver(ctx, arg)
}
func (r *Repository) DeleteDriver(ctx context.Context, arg int64) error {
	return r.Queries.DeleteDriver(ctx, arg)
}
func (r *Repository) GetDriverById(ctx context.Context, arg int64) (db.Driver, error) {
	return r.Queries.GetDriverById(ctx, arg)
}
func (r *Repository) GetDriverByUserId(ctx context.Context, arg int64) (db.Driver, error) {
	return r.Queries.GetDriverByUserId(ctx, arg)
}
