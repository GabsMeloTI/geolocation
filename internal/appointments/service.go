package appointments

import (
	"context"
	"database/sql"
	"errors"
)

type InterfaceService interface {
	CreateAppointmentService(ctx context.Context, data CreateAppointmentDTO) (AppointmentResponse, error)
	UpdateAppointmentSituationService(ctx context.Context, data UpdateAppointmentDTO) error
	DeleteAppointmentService(ctx context.Context, id int64) error
	GetAppointmentByUserIDService(ctx context.Context, userID int64) ([]AppointmentResponseList, error)
}

type Service struct {
	InterfaceService InterfaceRepository
}

func NewAppointmentsService(InterfaceService InterfaceRepository) *Service {
	return &Service{InterfaceService}
}

func (p *Service) CreateAppointmentService(ctx context.Context, data CreateAppointmentDTO) (AppointmentResponse, error) {
	arg := data.ParseCreateToAppointment()

	result, err := p.InterfaceService.CreateAppointment(ctx, arg)
	if err != nil {
		return AppointmentResponse{}, err
	}

	createAppointmentService := AppointmentResponse{}
	createAppointmentService.ParseFromAppointmentObject(result)

	return createAppointmentService, nil
}

func (p *Service) UpdateAppointmentSituationService(ctx context.Context, data UpdateAppointmentDTO) error {
	_, err := p.InterfaceService.GetAppointmentByID(ctx, data.Request.ID)

	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("appointment not found")
	}
	if err != nil {
		return err
	}

	arg := data.ParseUpdateToAppointment()

	err = p.InterfaceService.UpdateAppointmentSituation(ctx, arg)
	if err != nil {
		return err
	}

	return nil
}

func (p *Service) DeleteAppointmentService(ctx context.Context, id int64) error {
	_, err := p.InterfaceService.GetAppointmentByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("appointment not found")
	}
	if err != nil {
		return err
	}

	err = p.InterfaceService.DeleteAppointment(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

func (p *Service) GetAppointmentByUserIDService(ctx context.Context, userID int64) ([]AppointmentResponseList, error) {

	result, err := p.InterfaceService.GetListAppointmentByUserID(ctx, userID)
	if err != nil {
		return []AppointmentResponseList{}, err
	}

	var createAppointmentService []AppointmentResponseList
	for _, ap := range result {
		response := AppointmentResponseList{}
		response.ParseFromAppointmentListObject(ap)
		createAppointmentService = append(createAppointmentService, response)
	}

	return createAppointmentService, nil
}
