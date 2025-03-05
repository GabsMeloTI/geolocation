package plans

import (
	"context"
	"database/sql"
	db "geolocation/db/sqlc"
)

type InterfaceRepository interface {
	CreateUserPlans(ctx context.Context, arg db.CreateUserPlansParams) (db.UserPlan, error)
	GetPlansById(ctx context.Context, arg int64) (db.Plan, error)
	GetUserPlanByIdUser(ctx context.Context, arg db.GetUserPlanByIdUserParams) (db.UserPlan, error)
	UpdateUserPlan(ctx context.Context, arg db.UpdateUserPlanParams) error
}
type Repository struct {
	Conn    *sql.DB
	DBtx    db.DBTX
	Queries *db.Queries
	SqlConn *sql.DB
}

func NewUserPlanRepository(Conn *sql.DB) *Repository {
	q := db.New(Conn)
	return &Repository{
		Conn:    Conn,
		DBtx:    Conn,
		Queries: q,
		SqlConn: Conn,
	}
}

func (r *Repository) CreateUserPlans(ctx context.Context, arg db.CreateUserPlansParams) (db.UserPlan, error) {
	return r.Queries.CreateUserPlans(ctx, arg)
}
func (r *Repository) GetPlansById(ctx context.Context, arg int64) (db.Plan, error) {
	return r.Queries.GetPlansById(ctx, arg)
}
func (r *Repository) GetUserPlanByIdUser(ctx context.Context, arg db.GetUserPlanByIdUserParams) (db.UserPlan, error) {
	return r.Queries.GetUserPlanByIdUser(ctx, arg)
}
func (r *Repository) UpdateUserPlan(ctx context.Context, arg db.UpdateUserPlanParams) error {
	return r.Queries.UpdateUserPlan(ctx, arg)
}
