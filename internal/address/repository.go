package address

import (
	"context"
	"database/sql"
	db "geolocation/db/sqlc"
)

type InterfaceRepository interface {
	FindAddressesByQueryRepository(ctx context.Context, arg db.FindAddressesByQueryParams) ([]db.FindAddressesByQueryRow, error)
	FindAddressesByCEPRepository(ctx context.Context, arg string) ([]db.FindAddressesByCEPRow, error)
	FindAddressesByLatLonRepository(ctx context.Context, arg db.FindAddressesByLatLonParams) ([]db.FindAddressesByLatLonRow, error)
	IsStateRepository(ctx context.Context, arg string) (bool, error)
	IsCityRepository(ctx context.Context, arg string) (bool, error)
	IsNeighborhoodRepository(ctx context.Context, arg string) (bool, error)
}
type Repository struct {
	Conn    *sql.DB
	DBtx    db.DBTX
	Queries *db.Queries
	SqlConn *sql.DB
}

func NewAddressRepository(Conn *sql.DB) *Repository {
	q := db.New(Conn)
	return &Repository{
		Conn:    Conn,
		DBtx:    Conn,
		Queries: q,
		SqlConn: Conn,
	}
}

func (r *Repository) FindAddressesByQueryRepository(ctx context.Context, arg db.FindAddressesByQueryParams) ([]db.FindAddressesByQueryRow, error) {
	return r.Queries.FindAddressesByQuery(ctx, arg)
}
func (r *Repository) FindAddressesByCEPRepository(ctx context.Context, arg string) ([]db.FindAddressesByCEPRow, error) {
	return r.Queries.FindAddressesByCEP(ctx, arg)
}

func (r *Repository) FindAddressesByLatLonRepository(ctx context.Context, arg db.FindAddressesByLatLonParams) ([]db.FindAddressesByLatLonRow, error) {
	return r.Queries.FindAddressesByLatLon(ctx, arg)
}

func (r *Repository) IsStateRepository(ctx context.Context, arg string) (bool, error) {
	return r.Queries.IsState(ctx, arg)
}

func (r *Repository) IsCityRepository(ctx context.Context, arg string) (bool, error) {
	return r.Queries.IsCity(ctx, arg)
}

func (r *Repository) IsNeighborhoodRepository(ctx context.Context, arg string) (bool, error) {
	return r.Queries.IsNeighborhood(ctx, arg)
}
