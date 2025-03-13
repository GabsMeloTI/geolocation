package advertisement

import (
	"context"
	"database/sql"

	db "geolocation/db/sqlc"
)

type InterfaceRepository interface {
	CreateAdvertisement(
		ctx context.Context,
		arg db.CreateAdvertisementParams,
	) (db.Advertisement, error)
	UpdateAdvertisement(
		ctx context.Context,
		arg db.UpdateAdvertisementParams,
	) (db.Advertisement, error)
	DeleteAdvertisement(ctx context.Context, arg db.DeleteAdvertisementParams) error
	GetAdvertisementById(ctx context.Context, arg int64) (db.Advertisement, error)
	GetAllAdvertisementUsers(ctx context.Context) ([]db.GetAllAdvertisementUsersRow, error)
	GetAllAdvertisementPublic(ctx context.Context) ([]db.GetAllAdvertisementPublicRow, error)
	CountAdvertisementByUserID(ctx context.Context, arg int64) (int64, error)
	GetProfileById(ctx context.Context, arg int64) (db.Profile, error)
	UpdatedAdvertisementFinishedCreate(
		ctx context.Context,
		arg db.UpdatedAdvertisementFinishedCreateParams,
	) (db.UpdatedAdvertisementFinishedCreateRow, error)
	CreateAdvertisementRoute(
		ctx context.Context,
		arg db.CreateAdvertisementRouteParams,
	) (db.AdvertisementRoute, error)
	GetAllAdvertisementByUser(
		ctx context.Context,
		arg int64,
	) ([]db.GetAllAdvertisementByUserRow, error)
	UpdateAdvertismentRouteChoose(
		ctx context.Context,
		arg db.UpdateAdsRouteChooseByUserIdParams,
	) error
	GetAdvertisementExist(ctx context.Context, arg db.GetAdvertisementExistParams) (db.AdvertisementRoute, error)
}
type Repository struct {
	Conn    *sql.DB
	DBtx    db.DBTX
	Queries *db.Queries
	SqlConn *sql.DB
}

func NewAdvertisementsRepository(Conn *sql.DB) *Repository {
	q := db.New(Conn)
	return &Repository{
		Conn:    Conn,
		DBtx:    Conn,
		Queries: q,
		SqlConn: Conn,
	}
}

func (r *Repository) CreateAdvertisement(
	ctx context.Context,
	arg db.CreateAdvertisementParams,
) (db.Advertisement, error) {
	return r.Queries.CreateAdvertisement(ctx, arg)
}

func (r *Repository) UpdateAdvertisement(
	ctx context.Context,
	arg db.UpdateAdvertisementParams,
) (db.Advertisement, error) {
	return r.Queries.UpdateAdvertisement(ctx, arg)
}

func (r *Repository) DeleteAdvertisement(
	ctx context.Context,
	arg db.DeleteAdvertisementParams,
) error {
	return r.Queries.DeleteAdvertisement(ctx, arg)
}

func (r *Repository) GetAdvertisementById(
	ctx context.Context,
	arg int64,
) (db.Advertisement, error) {
	return r.Queries.GetAdvertisementById(ctx, arg)
}

func (r *Repository) GetAllAdvertisementUsers(
	ctx context.Context,
) ([]db.GetAllAdvertisementUsersRow, error) {
	return r.Queries.GetAllAdvertisementUsers(ctx)
}

func (r *Repository) GetAllAdvertisementPublic(
	ctx context.Context,
) ([]db.GetAllAdvertisementPublicRow, error) {
	return r.Queries.GetAllAdvertisementPublic(ctx)
}

func (r *Repository) CountAdvertisementByUserID(ctx context.Context, arg int64) (int64, error) {
	return r.Queries.CountAdvertisementByUserID(ctx, arg)
}

func (r *Repository) GetProfileById(ctx context.Context, arg int64) (db.Profile, error) {
	return r.Queries.GetProfileById(ctx, arg)
}

func (r *Repository) UpdatedAdvertisementFinishedCreate(
	ctx context.Context, arg db.UpdatedAdvertisementFinishedCreateParams) (db.UpdatedAdvertisementFinishedCreateRow, error) {
	return r.Queries.UpdatedAdvertisementFinishedCreate(ctx, arg)
}

func (r *Repository) CreateAdvertisementRoute(ctx context.Context, arg db.CreateAdvertisementRouteParams) (db.AdvertisementRoute, error) {
	return r.Queries.CreateAdvertisementRoute(ctx, arg)
}

func (r *Repository) GetAllAdvertisementByUser(ctx context.Context, arg int64) ([]db.GetAllAdvertisementByUserRow, error) {
	return r.Queries.GetAllAdvertisementByUser(ctx, arg)
}

func (r *Repository) UpdateAdvertismentRouteChoose(ctx context.Context, arg db.UpdateAdsRouteChooseByUserIdParams) error {
	return r.Queries.UpdateAdsRouteChooseByUserId(ctx, arg)
}

func (r *Repository) GetAdvertisementExist(ctx context.Context, arg db.GetAdvertisementExistParams) (db.AdvertisementRoute, error) {
	return r.Queries.GetAdvertisementExist(ctx, arg)
}
