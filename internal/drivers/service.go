package drivers

import (
	"context"
	"database/sql"
	"errors"
)

type InterfaceService interface {
	CreateDriverService(ctx context.Context, data CreateDriverRequest) (DriverResponse, error)
	UpdateDriverService(ctx context.Context, data UpdateDriverRequest) (DriverResponse, error)
	DeleteDriverService(ctx context.Context, id int64) error
}

type Service struct {
	InterfaceService InterfaceRepository
}

func NewDriversService(InterfaceService InterfaceRepository) *Service {
	return &Service{InterfaceService}
}

func (p *Service) CreateDriverService(ctx context.Context, data CreateDriverRequest) (DriverResponse, error) {
	arg := data.ParseCreateToDriver()

	result, err := p.InterfaceService.CreateDriver(ctx, arg)
	if err != nil {
		return DriverResponse{}, err
	}

	createDriverService := DriverResponse{}
	createDriverService.ParseFromDriverObject(result)

	return createDriverService, nil
}

func (p *Service) UpdateDriverService(ctx context.Context, data UpdateDriverRequest) (DriverResponse, error) {
	_, err := p.InterfaceService.GetDriverById(ctx, data.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return DriverResponse{}, errors.New("driver not found")
	}
	if err != nil {
		return DriverResponse{}, err
	}

	arg := data.ParseUpdateToDriver()

	result, err := p.InterfaceService.UpdateDriver(ctx, arg)
	if err != nil {
		return DriverResponse{}, err
	}

	updateDriverService := DriverResponse{}
	updateDriverService.ParseFromDriverObject(result)

	return updateDriverService, nil
}

func (p *Service) DeleteDriverService(ctx context.Context, id int64) error {
	_, err := p.InterfaceService.GetDriverById(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("driver not found")
	}
	if err != nil {
		return err
	}

	err = p.InterfaceService.DeleteDriver(ctx, id)
	if err != nil {
		return err
	}
	return nil
}
