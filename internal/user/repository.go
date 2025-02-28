package user

import (
	"context"
	"database/sql"
	db "geolocation/db/sqlc"
)

type InterfaceRepository interface {
	CreateUserRepository(ctx context.Context, arg db.CreateUserParams) (db.User, error)
	GetUserByEmailRepository(ctx context.Context, email string) (db.User, error)
	DeleteUserByIdRepository(ctx context.Context, id int64) error
	UpdateUserByIdRepository(ctx context.Context, arg db.UpdateUserByIdParams) (db.User, error)
	UpdateUserPasswordRepository(ctx context.Context, arg db.UpdateUserPasswordParams) error
}

type Repository struct {
	Conn    *sql.DB
	DBtx    db.DBTX
	Queries *db.Queries
	SqlConn *sql.DB
}

func NewUserRepository(Conn *sql.DB) *Repository {
	q := db.New(Conn)
	return &Repository{
		Conn:    Conn,
		DBtx:    Conn,
		Queries: q,
		SqlConn: Conn,
	}
}

func (r *Repository) CreateUserRepository(ctx context.Context, arg db.CreateUserParams) (db.User, error) {
	return r.Queries.CreateUser(ctx, arg)
}

func (r *Repository) GetUserByEmailRepository(ctx context.Context, email string) (db.User, error) {
	return r.Queries.GetUserByEmail(ctx, email)
}

func (r *Repository) DeleteUserByIdRepository(ctx context.Context, id int64) error {
	return r.Queries.DeleteUserById(ctx, id)
}

func (r *Repository) UpdateUserByIdRepository(ctx context.Context, arg db.UpdateUserByIdParams) (db.User, error) {
	return r.Queries.UpdateUserById(ctx, arg)
}
func (r *Repository) UpdateUserPasswordRepository(ctx context.Context, arg db.UpdateUserPasswordParams) error {
	return r.Queries.UpdateUserPassword(ctx, arg)
}
