package route_enterprise

import (
	"context"
	"database/sql"
	db "geolocation/db/sqlc"
)

type InterfaceRepository interface {
	CreateRouteEnterprise(ctx context.Context, arg db.CreateRouteEnterpriseParams) (db.RouteEnterprise, error)
	DeleteRouteEnterprise(ctx context.Context, arg db.DeleteRouteEnterpriseParams) error
	GetOrganizationByTenant(ctx context.Context, arg db.GetOrganizationByTenantParams) (sql.NullString, error)
}

type Repository struct {
	Conn    *sql.DB
	DBtx    db.DBTX
	Queries *db.Queries
	SqlConn *sql.DB
}

func NewRouteEnterpriseRepository(Conn *sql.DB) *Repository {
	q := db.New(Conn)
	return &Repository{
		Conn:    Conn,
		DBtx:    Conn,
		Queries: q,
		SqlConn: Conn,
	}
}

func (r *Repository) CreateRouteEnterprise(ctx context.Context, arg db.CreateRouteEnterpriseParams) (db.RouteEnterprise, error) {
	return r.Queries.CreateRouteEnterprise(ctx, arg)
}

func (r *Repository) DeleteRouteEnterprise(ctx context.Context, arg db.DeleteRouteEnterpriseParams) error {
	return r.Queries.DeleteRouteEnterprise(ctx, arg)
}

func (r *Repository) GetOrganizationByTenant(ctx context.Context, arg db.GetOrganizationByTenantParams) (sql.NullString, error) {
	return r.Queries.GetOrganizationByTenant(ctx, arg)
}
