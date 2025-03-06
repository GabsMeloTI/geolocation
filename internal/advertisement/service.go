package advertisement

import (
	"context"
	"database/sql"
	"errors"
	"geolocation/validation"
	"strings"
)

type InterfaceService interface {
	CreateAdvertisementService(ctx context.Context, data CreateAdvertisementDto, idProfile int64) (AdvertisementResponse, error)
	UpdateAdvertisementService(ctx context.Context, data UpdateAdvertisementDto, idProfile int64) (AdvertisementResponse, error)
	DeleteAdvertisementService(ctx context.Context, data DeleteAdvertisementRequest) error
	GetAllAdvertisementUser(ctx context.Context) ([]AdvertisementResponseAll, error)
	GetAllAdvertisementPublic(ctx context.Context) ([]AdvertisementResponseNoUser, error)
}

type Service struct {
	InterfaceService InterfaceRepository
}

func NewAdvertisementsService(InterfaceService InterfaceRepository) *Service {
	return &Service{InterfaceService}
}

func (p *Service) CreateAdvertisementService(ctx context.Context, data CreateAdvertisementDto, idProfile int64) (AdvertisementResponse, error) {
	if err := data.CreateAdvertisementRequest.ValidateCreate(); err != nil {
		return AdvertisementResponse{}, err
	}

	resultProfile, errProfile := p.InterfaceService.GetProfileById(ctx, idProfile)
	if errProfile != nil {
		return AdvertisementResponse{}, errProfile
	}

	if resultProfile.Name == "Driver" {
		return AdvertisementResponse{}, errors.New("motoristas não podem criar anúncios")
	}

	arg := data.ParseCreateToAdvertisement()
	arg.Situation = strings.ToLower(data.CreateAdvertisementRequest.Situation)
	result, err := p.InterfaceService.CreateAdvertisement(ctx, arg)
	if err != nil {
		return AdvertisementResponse{}, err
	}

	var response AdvertisementResponse
	response.ParseFromAdvertisementObject(result)

	return response, nil
}

func (p *Service) UpdateAdvertisementService(ctx context.Context, data UpdateAdvertisementDto, idProfile int64) (AdvertisementResponse, error) {
	_, err := p.InterfaceService.GetAdvertisementById(ctx, data.UpdateAdvertisementRequest.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return AdvertisementResponse{}, errors.New("anúncio não encontrado")
	}
	if err != nil {
		return AdvertisementResponse{}, err
	}

	if err := data.UpdateAdvertisementRequest.ValidateUpdate(); err != nil {
		return AdvertisementResponse{}, err
	}
	resultProfile, errProfile := p.InterfaceService.GetProfileById(ctx, idProfile)
	if errProfile != nil {
		return AdvertisementResponse{}, errProfile
	}

	if resultProfile.Name == "Driver" {
		return AdvertisementResponse{}, errors.New("motoristas não podem criar anúncios")
	}

	arg := data.ParseUpdateToAdvertisement()
	arg.Situation = strings.ToLower(data.UpdateAdvertisementRequest.Situation)
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

func (p *Service) GetAllAdvertisementPublic(ctx context.Context) ([]AdvertisementResponseNoUser, error) {
	results, err := p.InterfaceService.GetAllAdvertisementPublic(ctx)
	if err != nil {
		return nil, err
	}

	var announcementResponses []AdvertisementResponseNoUser
	for _, result := range results {

		var response AdvertisementResponseNoUser
		response.ParseFromAdvertisementObject(result)

		announcementResponses = append(announcementResponses, response)
	}
	return announcementResponses, nil
}
