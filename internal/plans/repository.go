package plans

import (
	"context"
	"database/sql"
	db "geolocation/db/sqlc"
)

type InterfaceRepository interface {
	CreatePlans(ctx context.Context, arg db.CreatePlansParams) (db.Plan, error)
	UpdatePlans(ctx context.Context, arg db.UpdatePlansParams) (db.Plan, error)
	CreateUserPlans(ctx context.Context, arg db.CreateUserPlansParams) (db.UserPlan, error)
}
type Repository struct {
	Conn    *sql.DB
	DBtx    db.DBTX
	Queries *db.Queries
	SqlConn *sql.DB
}

func NewPlansRepository(Conn *sql.DB) *Repository {
	q := db.New(Conn)
	return &Repository{
		Conn:    Conn,
		DBtx:    Conn,
		Queries: q,
		SqlConn: Conn,
	}
}

func (r *Repository) CreatePlans(ctx context.Context, arg db.CreatePlansParams) (db.Plan, error) {
	return r.Queries.CreatePlans(ctx, arg)
}
func (r *Repository) UpdatePlans(ctx context.Context, arg db.UpdatePlansParams) (db.Plan, error) {
	return r.Queries.UpdatePlans(ctx, arg)
}
func (r *Repository) CreateUserPlans(ctx context.Context, arg db.CreateUserPlansParams) (db.UserPlan, error) {
	return r.Queries.CreateUserPlans(ctx, arg)
}
