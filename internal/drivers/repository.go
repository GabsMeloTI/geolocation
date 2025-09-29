package drivers

import (
	"context"
	"database/sql"

	db "geolocation/db/sqlc"
)

type InterfaceRepository interface {
	CreateDriver(ctx context.Context, arg db.CreateDriverParams) (db.Driver, error)
	UpdateDriver(ctx context.Context, arg db.UpdateDriverParams) (db.Driver, error)
	DeleteDriver(ctx context.Context, arg db.DeleteDriverParams) error
	GetDriverById(ctx context.Context, arg int64) (db.Driver, error)
	GetDriverByUserId(ctx context.Context, arg int64) ([]db.Driver, error)
	GetProfileById(ctx context.Context, profileId int64) (db.Profile, error)
	CreateUserToCarrier(ctx context.Context, arg db.CreateUserParams) (db.User, error)
	GetUserByEmail(ctx context.Context, email string) (db.User, error)
	GetOneDriverByUserId(ctx context.Context, arg int64) (db.Driver, error)
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

func (r *Repository) CreateDriver(
	ctx context.Context,
	arg db.CreateDriverParams,
) (db.Driver, error) {
	return r.Queries.CreateDriver(ctx, arg)
}

func (r *Repository) UpdateDriver(
	ctx context.Context,
	arg db.UpdateDriverParams,
) (db.Driver, error) {
	return r.Queries.UpdateDriver(ctx, arg)
}

func (r *Repository) DeleteDriver(ctx context.Context, arg db.DeleteDriverParams) error {
	return r.Queries.DeleteDriver(ctx, arg)
}

func (r *Repository) GetDriverById(ctx context.Context, arg int64) (db.Driver, error) {
	return r.Queries.GetDriverById(ctx, arg)
}

func (r *Repository) GetDriverByUserId(ctx context.Context, arg int64) ([]db.Driver, error) {
	return r.Queries.GetDriverByUserId(ctx, arg)
}

func (r *Repository) GetProfileById(ctx context.Context, profileId int64) (db.Profile, error) {
	return r.Queries.GetProfileById(ctx, profileId)
}

func (r *Repository) CreateUserToCarrier(
	ctx context.Context,
	arg db.CreateUserParams,
) (db.User, error) {
	return r.Queries.CreateUser(ctx, arg)
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	return r.Queries.GetUserByEmail(ctx, email)
}
func (r *Repository) GetOneDriverByUserId(ctx context.Context, arg int64) (db.Driver, error) {
	return r.Queries.GetOneDriverByUserId(ctx, arg)
}
