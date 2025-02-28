package advertisement

import (
	"context"
	"database/sql"
	"errors"
	"geolocation/validation"
	"strings"
)

type InterfaceService interface {
	CreateAdvertisementService(ctx context.Context, data CreateAdvertisementRequest) (AdvertisementResponse, error)
	UpdateAdvertisementService(ctx context.Context, data UpdateAdvertisementRequest) (AdvertisementResponse, error)
	DeleteAdvertisementService(ctx context.Context, data DeleteAdvertisementRequest) error
	GetAllAdvertisementUser(ctx context.Context) ([]AdvertisementResponseAll, error)
}

type Service struct {
	InterfaceService InterfaceRepository
}

func NewAdvertisementsService(InterfaceService InterfaceRepository) *Service {
	return &Service{InterfaceService}
}

func (p *Service) CreateAdvertisementService(ctx context.Context, data CreateAdvertisementRequest) (AdvertisementResponse, error) {
	if err := data.ValidateCreate(); err != nil {
		return AdvertisementResponse{}, err
	}

	arg := data.ParseCreateToAdvertisement()
	arg.Situation = strings.ToLower(data.Situation)
	result, err := p.InterfaceService.CreateAdvertisement(ctx, arg)
	if err != nil {
		return AdvertisementResponse{}, err
	}

	var response AdvertisementResponse
	response.ParseFromAdvertisementObject(result)

	return response, nil
}

func (p *Service) UpdateAdvertisementService(ctx context.Context, data UpdateAdvertisementRequest) (AdvertisementResponse, error) {
	_, err := p.InterfaceService.GetAdvertisementById(ctx, data.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return AdvertisementResponse{}, errors.New("anúncio não encontrado")
	}
	if err != nil {
		return AdvertisementResponse{}, err
	}

	if err := data.ValidateUpdate(); err != nil {
		return AdvertisementResponse{}, err
	}

	arg := data.ParseUpdateToAdvertisement()
	arg.Situation = strings.ToLower(data.Situation)
	result, err := p.InterfaceService.UpdateAdvertisement(ctx, arg)
	if err != nil {
		return AdvertisementResponse{}, err
	}

	var response AdvertisementResponse
	response.ParseFromAdvertisementObject(result)

	return response, nil
}

func (p *Service) DeleteAdvertisementService(ctx context.Context, data DeleteAdvertisementRequest) error {
	_, err := p.InterfaceService.GetAdvertisementById(ctx, data.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("anúncio não encontrado")
	}
	if err != nil {
		return err
	}

	arg := data.ParseDeleteToAdvertisement()
	err = p.InterfaceService.DeleteAdvertisement(ctx, arg)
	if err != nil {
		return err
	}

	return nil
}

func (p *Service) GetAllAdvertisementUser(ctx context.Context) ([]AdvertisementResponseAll, error) {
	results, err := p.InterfaceService.GetAllAdvertisementUsers(ctx)
	if err != nil {
		return nil, err
	}

	var announcementResponses []AdvertisementResponseAll
	for _, result := range results {
		totalFreights, err := p.InterfaceService.CountAdvertisementByUserID(ctx, result.UserID)
		if err != nil {
			return nil, err
		}

		var response AdvertisementResponseAll
		response.ActiveFreight = totalFreights
		response.ActiveDuration = validation.FormatActiveDuration(response.ActiveThere)
		response.ParseFromAdvertisementObject(result)

		announcementResponses = append(announcementResponses, response)
	}
	return announcementResponses, nil
}
