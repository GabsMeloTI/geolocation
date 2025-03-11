package advertisement

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"geolocation/internal/new_routes"
	"geolocation/validation"
	"strings"
)

type InterfaceService interface {
	CreateAdvertisementService(ctx context.Context, data CreateAdvertisementDto, idProfile int64) (AdvertisementResponse, error)
	UpdateAdvertisementService(ctx context.Context, data UpdateAdvertisementDto, idProfile int64) (AdvertisementResponse, error)
	DeleteAdvertisementService(ctx context.Context, data DeleteAdvertisementRequest) error
	GetAllAdvertisementUser(ctx context.Context) ([]AdvertisementResponseAll, error)
	GetAllAdvertisementPublic(ctx context.Context) ([]AdvertisementResponseNoUser, error)
	UpdatedAdvertisementFinishedCreate(ctx context.Context, data UpdatedAdvertisementFinishedCreate, idProfile int64) (ResponseUpdatedAdvertisementFinishedCreate, error)
	GetAllAdvertisementUser2(ctx context.Context) ([]AdvertisementResponseAll, error)
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
	result, err := p.InterfaceService.CreateAdvertisement(ctx, arg)
	if err != nil {
		return AdvertisementResponse{}, err
	}

	var response AdvertisementResponse
	response.ParseFromAdvertisementObject(result)

	return response, nil
}

func (p *Service) UpdatedAdvertisementFinishedCreate(ctx context.Context, data UpdatedAdvertisementFinishedCreate, idProfile int64) (ResponseUpdatedAdvertisementFinishedCreate, error) {
	resultProfile, errProfile := p.InterfaceService.GetProfileById(ctx, idProfile)
	if errProfile != nil {
		return ResponseUpdatedAdvertisementFinishedCreate{}, errProfile
	}
	if resultProfile.Name == "Driver" {
		return ResponseUpdatedAdvertisementFinishedCreate{}, errors.New("motoristas não podem criar anúncios")
	}

	arg := data.ParseUpdatedToAdvertisementFinishedCreate()
	result, err := p.InterfaceService.UpdatedAdvertisementFinishedCreate(ctx, arg)
	if err != nil {
		return ResponseUpdatedAdvertisementFinishedCreate{}, err
	}

	argAdRoute := data.ParseCreateToAdvertisementRoute()
	resultAdRoute, errAdRoute := p.InterfaceService.CreateAdvertisementRoute(ctx, argAdRoute)
	if errAdRoute != nil {
		return ResponseUpdatedAdvertisementFinishedCreate{}, errAdRoute
	}

	var response ResponseUpdatedAdvertisementFinishedCreate
	response.ParseFromUpdatedAdvertisementFinishedCreateObject(resultAdRoute.RouteHistID, resultAdRoute.RouteChoose, result)

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
		index := int(result.RouteChoose)
		var route new_routes.FinalOutput
		if index >= 0 && index < len(result.ResponseRoutes) {
			errRoute := json.Unmarshal(result.ResponseRoutes, &route)
			if errRoute != nil {
				return announcementResponses, errRoute
			}
		} else {
			result.ResponseRoutes = nil
		}

		totalFreights, err := p.InterfaceService.CountAdvertisementByUserID(ctx, result.UserID)
		if err != nil {
			return nil, err
		}

		var response AdvertisementResponseAll
		response.RouteChoose = route.Routes[index]
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

func (p *Service) GetAllAdvertisementUser2(ctx context.Context) ([]AdvertisementResponseAll, error) {
	result, err := p.InterfaceService.GetAllAdvertisementUsersNotComplete(ctx)
	if err != nil {
		return nil, err
	}

	var announcementResponses []AdvertisementResponseAll
	for _, item := range result {
		announcementResponses = append(announcementResponses, AdvertisementResponseAll{
			ID:                      item.ID,
			UserID:                  item.UserID,
			UserName:                item.UserName,
			ActiveThere:             item.ActiveThere.Time,
			UserCity:                item.UserCity.String,
			UserState:               item.UserState.String,
			UserPhone:               item.UserPhone.String,
			UserEmail:               item.UserEmail,
			Destination:             item.Destination,
			Origin:                  item.Origin,
			PickupDate:              item.PickupDate,
			DeliveryDate:            item.DeliveryDate,
			ExpirationDate:          item.ExpirationDate,
			Title:                   item.Title,
			CargoType:               item.CargoType,
			CargoSpecies:            item.CargoSpecies,
			CargoWeight:             item.CargoWeight,
			VehiclesAccepted:        item.VehiclesAccepted,
			Trailer:                 item.Trailer,
			RequiresTarp:            item.RequiresTarp,
			Tracking:                item.Tracking,
			Agency:                  item.Agency,
			Description:             item.Description,
			PaymentType:             item.PaymentType,
			Advance:                 item.Advance,
			Toll:                    item.Toll,
			Situation:               item.Situation,
			Price:                   item.Price,
			StateOrigin:             item.StateOrigin,
			CityOrigin:              item.CityOrigin,
			ComplementOrigin:        item.ComplementOrigin,
			NeighborhoodOrigin:      item.NeighborhoodOrigin,
			StreetOrigin:            item.StreetOrigin,
			StreetNumberOrigin:      item.StreetNumberOrigin,
			CEPOrigin:               item.CepOrigin,
			StateDestination:        item.StateDestination,
			CityDestination:         item.CityDestination,
			ComplementDestination:   item.ComplementDestination,
			NeighborhoodDestination: item.NeighborhoodDestination,
			StreetDestination:       item.StreetDestination,
			StreetNumberDestination: item.StreetNumberDestination,
			CEPDestination:          item.CepDestination,
		})
	}

	return announcementResponses, nil
}
