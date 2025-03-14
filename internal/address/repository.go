package address

import (
	"context"
	"database/sql"
	db "geolocation/db/sqlc"
)

type InterfaceRepository interface {
	FindAddressesByQueryRepository(context.Context, db.FindAddressesByQueryParams) ([]db.FindAddressesByQueryRow, error)
	FindAddressesByCEPRepository(context.Context, string) ([]db.FindAddressesByCEPRow, error)
	FindAddressesByLatLonRepository(context.Context, db.FindAddressesByLatLonParams) ([]db.FindAddressesByLatLonRow, error)
	FindAddressGroupedByCEPRepository(ctx context.Context, arg string) ([]db.FindAddressGroupedByCEPRow, error)
	IsStateRepository(context.Context, string) (bool, error)
	IsCityRepository(context.Context, string) (bool, error)
	IsNeighborhoodRepository(context.Context, string) (bool, error)
	FindStateAll(context.Context) ([]db.State, error)
	FindCityAll(context.Context, int32) ([]db.City, error)
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

func (r *Repository) FindAddressGroupedByCEPRepository(ctx context.Context, arg string) ([]db.FindAddressGroupedByCEPRow, error) {
	return r.Queries.FindAddressGroupedByCEP(ctx, arg)
}

func (r *Repository) FindStateAll(ctx context.Context) ([]db.State, error) {
	return r.Queries.FindStateAll(ctx)
}

func (r *Repository) FindCityAll(ctx context.Context, arg int32) ([]db.City, error) {
	return r.Queries.FindCityAll(ctx, arg)
}
