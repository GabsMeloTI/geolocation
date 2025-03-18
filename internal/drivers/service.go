package drivers

import (
	"context"
	"database/sql"
	"errors"
	"net/mail"

	db "geolocation/db/sqlc"
)

type InterfaceService interface {
	CreateDriverService(ctx context.Context, data CreateDriverDto) (DriverResponse, error)
	UpdateDriverService(ctx context.Context, data UpdateDriverDto) (DriverResponse, error)
	DeleteDriverService(ctx context.Context, id, idUser int64) error
	GetDriverService(ctx context.Context, id int64) ([]DriverResponse, error)
	GetDriverByIdService(ctx context.Context, id int64) (DriverResponse, error)
}

type Service struct {
	InterfaceService InterfaceRepository
}

func NewDriversService(InterfaceService InterfaceRepository) *Service {
	return &Service{InterfaceService}
}

func (p *Service) CreateDriverService(
	ctx context.Context,
	data CreateDriverDto,
) (DriverResponse, error) {
	profile, err := p.InterfaceService.GetProfileById(ctx, data.ProfileId)
	if err != nil {
		return DriverResponse{}, err
	}

	if profile.Name == "Shipper" || profile.Name == "Carrier Driver" {
		return DriverResponse{}, errors.New("you cannot create a driver")
	}
	u, err := p.InterfaceService.GetUserByEmail(ctx, data.CreateDriverRequest.Email)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return DriverResponse{}, err
		}
	}

	if u.ID != 0 {
		return DriverResponse{}, errors.New("email already in use")
	}

	arg := data.ParseCreateToDriver()

	result, err := p.InterfaceService.CreateDriver(ctx, arg)
	if err != nil {
		return DriverResponse{}, err
	}

	if profile.Name == "Carrier" || profile.Name == "Driver" {
		_, err = mail.ParseAddress(data.CreateDriverRequest.Email)
		if err != nil {
			return DriverResponse{}, err
		}

		_, err = p.InterfaceService.CreateUserToCarrier(
			ctx,
			data.ParseToCreateUserParams(result.ID),
		)
		if err != nil {
			return DriverResponse{}, err
		}
	}

	createDriverService := DriverResponse{}
	createDriverService.ParseFromDriverObject(result)

	return createDriverService, nil
}

func (p *Service) UpdateDriverService(
	ctx context.Context,
	data UpdateDriverDto,
) (DriverResponse, error) {
	_, err := p.InterfaceService.GetDriverById(ctx, data.UpdateDriverRequest.ID)
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

func (p *Service) DeleteDriverService(ctx context.Context, id, idUser int64) error {
	_, err := p.InterfaceService.GetDriverById(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("driver not found")
	}
	if err != nil {
		return err
	}

	err = p.InterfaceService.DeleteDriver(ctx, db.DeleteDriverParams{
		ID:     id,
		UserID: idUser,
	})
	if err != nil {
		return err
	}
	return nil
}

func (p *Service) GetDriverService(ctx context.Context, id int64) ([]DriverResponse, error) {
	result, err := p.InterfaceService.GetDriverByUserId(ctx, id)
	if err != nil {
		return []DriverResponse{}, err
	}

	var getAllDriver []DriverResponse
	for _, trailer := range result {
		getDriverResponse := DriverResponse{}
		getDriverResponse.ParseFromDriverObject(trailer)
		getAllDriver = append(getAllDriver, getDriverResponse)
	}

	return getAllDriver, nil
}

func (p *Service) GetDriverByIdService(ctx context.Context, id int64) (DriverResponse, error) {
	result, err := p.InterfaceService.GetDriverById(ctx, id)
	if err != nil {
		return DriverResponse{}, err
	}

	getDriverResponse := DriverResponse{}
	getDriverResponse.ParseFromDriverObject(result)

	return getDriverResponse, nil
}
