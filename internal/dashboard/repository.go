package dashboard

import (
	"context"
	"database/sql"
	db "geolocation/db/sqlc"
)

type InterfaceRepository interface {
	GetDashboardDriver(ctx context.Context, arg db.GetDashboardDriverParams) (db.GetDashboardDriverRow, error)
	GetProfileById(ctx context.Context, arg int64) (db.Profile, error)
	GetDashboardHist(ctx context.Context, arg db.GetDashboardHistParams) ([]db.GetDashboardHistRow, error)
	GetDashboardFuture(ctx context.Context, arg db.GetDashboardFutureParams) ([]db.GetDashboardFutureRow, error)
	GetDashboardCalendar(ctx context.Context, arg db.GetDashboardCalendarParams) ([]db.GetDashboardCalendarRow, error)
	GetDashboardFaturamento(ctx context.Context, arg db.GetDashboardFaturamentoParams) ([]db.GetDashboardFaturamentoRow, error)
	GetDashboardDriverEnterprise(ctx context.Context, arg int64) ([]db.GetDashboardDriverEnterpriseRow, error)
	GetDashboardTrailerEnterprise(ctx context.Context, arg db.GetDashboardTrailerEnterpriseParams) ([]db.GetDashboardTrailerEnterpriseRow, error)
	GetDashboardTractUnitEnterprise(ctx context.Context, arg db.GetDashboardTractUnitEnterpriseParams) ([]db.GetDashboardTractUnitEnterpriseRow, error)
	GetOffersForDashboard(ctx context.Context, arg db.GetOffersForDashboardParams) (db.GetOffersForDashboardRow, error)
}
type Repository struct {
	Conn    *sql.DB
	DBtx    db.DBTX
	Queries *db.Queries
	SqlConn *sql.DB
}

func NewDashboardRepository(Conn *sql.DB) *Repository {
	q := db.New(Conn)
	return &Repository{
		Conn:    Conn,
		DBtx:    Conn,
		Queries: q,
		SqlConn: Conn,
	}
}

func (r *Repository) GetDashboardDriver(ctx context.Context, arg db.GetDashboardDriverParams) (db.GetDashboardDriverRow, error) {
	return r.Queries.GetDashboardDriver(ctx, arg)
}
func (r *Repository) GetDashboardHist(ctx context.Context, arg db.GetDashboardHistParams) ([]db.GetDashboardHistRow, error) {
	return r.Queries.GetDashboardHist(ctx, arg)
}
func (r *Repository) GetDashboardFuture(ctx context.Context, arg db.GetDashboardFutureParams) ([]db.GetDashboardFutureRow, error) {
	return r.Queries.GetDashboardFuture(ctx, arg)
}
func (r *Repository) GetDashboardCalendar(ctx context.Context, arg db.GetDashboardCalendarParams) ([]db.GetDashboardCalendarRow, error) {
	return r.Queries.GetDashboardCalendar(ctx, arg)
}
func (r *Repository) GetDashboardFaturamento(ctx context.Context, arg db.GetDashboardFaturamentoParams) ([]db.GetDashboardFaturamentoRow, error) {
	return r.Queries.GetDashboardFaturamento(ctx, arg)
}
func (r *Repository) GetProfileById(ctx context.Context, arg int64) (db.Profile, error) {
	return r.Queries.GetProfileById(ctx, arg)
}
func (r *Repository) GetDashboardDriverEnterprise(ctx context.Context, arg int64) ([]db.GetDashboardDriverEnterpriseRow, error) {
	return r.Queries.GetDashboardDriverEnterprise(ctx, arg)
}
func (r *Repository) GetDashboardTrailerEnterprise(ctx context.Context, arg db.GetDashboardTrailerEnterpriseParams) ([]db.GetDashboardTrailerEnterpriseRow, error) {
	return r.Queries.GetDashboardTrailerEnterprise(ctx, arg)
}
func (r *Repository) GetDashboardTractUnitEnterprise(ctx context.Context, arg db.GetDashboardTractUnitEnterpriseParams) ([]db.GetDashboardTractUnitEnterpriseRow, error) {
	return r.Queries.GetDashboardTractUnitEnterprise(ctx, arg)
}
func (r *Repository) GetOffersForDashboard(ctx context.Context, arg db.GetOffersForDashboardParams) (db.GetOffersForDashboardRow, error) {
	return r.Queries.GetOffersForDashboard(ctx, arg)
}
