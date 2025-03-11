package payment

import (
	"context"
	"database/sql"
	db "geolocation/db/sqlc"
)

type InterfaceRepository interface {
	CreatePaymentHist(ctx context.Context, arg db.CreatePaymentHistParams) (db.PaymentHist, error)
	GetPaymentHist(ctx context.Context, arg int64) ([]db.PaymentHist, error)
}
type Repository struct {
	Conn    *sql.DB
	DBtx    db.DBTX
	Queries *db.Queries
	SqlConn *sql.DB
}

func NewPaymentRepository(Conn *sql.DB) *Repository {
	q := db.New(Conn)
	return &Repository{
		Conn:    Conn,
		DBtx:    Conn,
		Queries: q,
		SqlConn: Conn,
	}
}

func (r *Repository) CreatePaymentHist(ctx context.Context, arg db.CreatePaymentHistParams) (db.PaymentHist, error) {
	return r.Queries.CreatePaymentHist(ctx, arg)
}
func (r *Repository) GetPaymentHist(ctx context.Context, arg int64) ([]db.PaymentHist, error) {
	return r.Queries.GetPaymentHist(ctx, arg)
}
