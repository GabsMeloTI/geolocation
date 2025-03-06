package appointments

import (
	"context"
	"database/sql"
	db "geolocation/db/sqlc"
)

type InterfaceRepository interface {
	CreateAppointment(ctx context.Context, arg db.CreateAppointmentParams) (db.Appointment, error)
	UpdateAppointmentSituation(ctx context.Context, arg db.UpdateAppointmentSituationParams) error
	DeleteAppointment(ctx context.Context, arg int64) error
	GetAppointmentByID(ctx context.Context, arg int64) (db.Appointment, error)
	GetListAppointmentByUserID(ctx context.Context, arg int64) ([]db.GetListAppointmentByUserIDRow, error)
	//GetListAppointmentByAdvertiser(ctx context.Context, arg int64) ([]db.GetListAppointmentByAdvertiserRow, error)
}
type Repository struct {
	Conn    *sql.DB
	DBtx    db.DBTX
	Queries *db.Queries
	SqlConn *sql.DB
}

func NewAppointmentsRepository(Conn *sql.DB) *Repository {
	q := db.New(Conn)
	return &Repository{
		Conn:    Conn,
		DBtx:    Conn,
		Queries: q,
		SqlConn: Conn,
	}
}

func (r *Repository) CreateAppointment(ctx context.Context, arg db.CreateAppointmentParams) (db.Appointment, error) {
	return r.Queries.CreateAppointment(ctx, arg)
}
func (r *Repository) UpdateAppointmentSituation(ctx context.Context, arg db.UpdateAppointmentSituationParams) error {
	return r.Queries.UpdateAppointmentSituation(ctx, arg)
}
func (r *Repository) DeleteAppointment(ctx context.Context, arg int64) error {
	return r.Queries.DeleteAppointment(ctx, arg)
}
func (r *Repository) GetAppointmentByID(ctx context.Context, arg int64) (db.Appointment, error) {
	return r.Queries.GetAppointmentByID(ctx, arg)
}

func (r *Repository) GetListAppointmentByUserID(ctx context.Context, arg int64) ([]db.GetListAppointmentByUserIDRow, error) {
	return r.Queries.GetListAppointmentByUserID(ctx, arg)
}

//func (r *Repository) GetListAppointmentByAdvertiser(ctx context.Context, arg int64) ([]db.GetListAppointmentByAdvertiserRow, error) {
//	return r.Queries.GetListAppointmentByAdvertiser(ctx, arg)
//}
