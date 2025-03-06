package tractor_unit

import (
	"context"
	"database/sql"
	"errors"
	db "geolocation/db/sqlc"
)

type InterfaceService interface {
	CreateTractorUnitService(ctx context.Context, data CreateTractorUnitDto) (TractorUnitResponse, error)
	UpdateTractorUnitService(ctx context.Context, data UpdateTractorUnitDto) (TractorUnitResponse, error)
	DeleteTractorUnitService(ctx context.Context, id, idUser int64) error
	GetTractorUnitService(ctx context.Context, id int64) ([]TractorUnitResponse, error)
}

type Service struct {
	InterfaceService InterfaceRepository
}

func NewTractorUnitsService(InterfaceService InterfaceRepository) *Service {
	return &Service{InterfaceService}
}

func (p *Service) CreateTractorUnitService(ctx context.Context, data CreateTractorUnitDto) (TractorUnitResponse, error) {
	arg := data.ParseCreateToTractorUnit()

	result, err := p.InterfaceService.CreateTractorUnit(ctx, arg)
	if err != nil {
		return TractorUnitResponse{}, err
	}

	createTractorUnitService := TractorUnitResponse{}
	createTractorUnitService.ParseFromTractorUnitObject(result)

	return createTractorUnitService, nil
}

func (p *Service) UpdateTractorUnitService(ctx context.Context, data UpdateTractorUnitDto) (TractorUnitResponse, error) {
	_, err := p.InterfaceService.GetTractorUnitById(ctx, data.UpdateTractorUnitRequest.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return TractorUnitResponse{}, errors.New("driver not found")
	}
	if err != nil {
		return TractorUnitResponse{}, err
	}

	arg := data.ParseUpdateToTractorUnit()

	result, err := p.InterfaceService.UpdateTractorUnit(ctx, arg)
	if err != nil {
		return TractorUnitResponse{}, err
	}

	updateTractorUnitService := TractorUnitResponse{}
	updateTractorUnitService.ParseFromTractorUnitObject(result)

	return updateTractorUnitService, nil
}

func (p *Service) DeleteTractorUnitService(ctx context.Context, id, idUser int64) error {
	_, err := p.InterfaceService.GetTractorUnitById(ctx, id)
	if err != nil {
		return err
	}

	err = p.InterfaceService.DeleteTractorUnit(ctx, db.DeleteTractorUnitParams{
		ID:     id,
		UserID: idUser,
	})
	if err != nil {
		return err
	}
	return nil
}

func (p *Service) GetTractorUnitService(ctx context.Context, id int64) ([]TractorUnitResponse, error) {
	result, err := p.InterfaceService.GetTractorUnitByUserId(ctx, id)
	if err != nil {
		return []TractorUnitResponse{}, err
	}

	var getAllTractorUnit []TractorUnitResponse
	for _, trailer := range result {
		getTractorUnitResponse := TractorUnitResponse{}
		getTractorUnitResponse.ParseFromTractorUnitObject(trailer)
		getAllTractorUnit = append(getAllTractorUnit, getTractorUnitResponse)
	}

	return getAllTractorUnit, nil
}
