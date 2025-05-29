package location

import (
	"context"
	"database/sql"
	db "geolocation/db/sqlc"
)

type InterfaceRepository interface {
	CreateLocation(ctx context.Context, arg db.CreateLocationParams) (db.Location, error)
	UpdateLocation(ctx context.Context, arg db.UpdateLocationParams) (db.Location, error)
	DeleteLocation(ctx context.Context, arg db.DeleteLocationParams) error
	GetLocationByOrg(ctx context.Context, arg db.GetLocationByOrgParams) ([]db.Location, error)
	GetAreasByOrg(ctx context.Context, arg int64) ([]db.GetAreasByOrgRow, error)
	CreateArea(ctx context.Context, arg db.CreateAreaParams) (db.Area, error)
	UpdateArea(ctx context.Context, arg db.UpdateAreaParams) (db.Area, error)
	DeleteArea(ctx context.Context, arg int64) error
	GetLocationByName(ctx context.Context, arg db.GetLocationByNameParams) ([]db.Location, error)
	GetLocationByNameExcludingID(ctx context.Context, arg db.GetLocationByNameExcludingIDParams) (db.Location, error)
}
type Repository struct {
	Conn    *sql.DB
	DBtx    db.DBTX
	Queries *db.Queries
	SqlConn *sql.DB
}

func NewLocationsRepository(Conn *sql.DB) *Repository {
	q := db.New(Conn)
	return &Repository{
		Conn:    Conn,
		DBtx:    Conn,
		Queries: q,
		SqlConn: Conn,
	}
}

func (r *Repository) CreateLocation(ctx context.Context, arg db.CreateLocationParams) (db.Location, error) {
	return r.Queries.CreateLocation(ctx, arg)
}

func (r *Repository) UpdateLocation(ctx context.Context, arg db.UpdateLocationParams) (db.Location, error) {
	return r.Queries.UpdateLocation(ctx, arg)
}

func (r *Repository) DeleteLocation(ctx context.Context, arg db.DeleteLocationParams) error {
	return r.Queries.DeleteLocation(ctx, arg)
}

func (r *Repository) GetLocationByOrg(ctx context.Context, arg db.GetLocationByOrgParams) ([]db.Location, error) {
	return r.Queries.GetLocationByOrg(ctx, arg)
}

func (r *Repository) GetAreasByOrg(ctx context.Context, arg int64) ([]db.GetAreasByOrgRow, error) {
	return r.Queries.GetAreasByOrg(ctx, arg)
}

func (r *Repository) CreateArea(ctx context.Context, arg db.CreateAreaParams) (db.Area, error) {
	return r.Queries.CreateArea(ctx, arg)
}

func (r *Repository) UpdateArea(ctx context.Context, arg db.UpdateAreaParams) (db.Area, error) {
	return r.Queries.UpdateArea(ctx, arg)
}

func (r *Repository) DeleteArea(ctx context.Context, arg int64) error {
	return r.Queries.DeleteArea(ctx, arg)
}

func (r *Repository) GetLocationByName(ctx context.Context, arg db.GetLocationByNameParams) ([]db.Location, error) {
	return r.Queries.GetLocationByName(ctx, arg)
}
func (r *Repository) GetLocationByNameExcludingID(ctx context.Context, arg db.GetLocationByNameExcludingIDParams) (db.Location, error) {
	return r.Queries.GetLocationByNameExcludingID(ctx, arg)
}
