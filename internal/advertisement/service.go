package advertisement

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"

	db "geolocation/db/sqlc"
	"geolocation/internal/new_routes"
	"geolocation/validation"
)

type InterfaceService interface {
	CreateAdvertisementService(ctx context.Context, data CreateAdvertisementDto, idProfile int64) (AdvertisementResponse, error)
	UpdateAdvertisementService(ctx context.Context, data UpdateAdvertisementDto, idProfile int64) (AdvertisementResponse, error)
	DeleteAdvertisementService(ctx context.Context, data DeleteAdvertisementRequest) error
	GetAllAdvertisementUser(ctx context.Context) ([]AdvertisementResponseAll, error)
	GetAllAdvertisementPublic(ctx context.Context) ([]AdvertisementResponseNoUser, error)
	UpdatedAdvertisementFinishedCreate(ctx context.Context, data UpdatedAdvertisementFinishedCreate, idProfile int64) (ResponseUpdatedAdvertisementFinishedCreate, error)
	GetAllAdvertisementByUser(ctx context.Context, id int64) ([]AdvertisementResponseAll, error)
	UpdateAdsRouteChooseService(ctx context.Context, data UpdateAdsRouteChooseDTO) error
	GetAdvertisementByIDService(ctx context.Context, id int64) (AdvertisementResponseAll, error)
	GetAdvertisementByIDPublicService(ctx context.Context, id int64) (AdvertisementResponseNoUser, error)
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

	_, errExist := p.InterfaceService.GetAdvertisementExist(ctx, db.GetAdvertisementExistParams{
		RouteHistID:     data.RouteHistID,
		UserID:          data.UserID,
		AdvertisementID: data.ID,
	})
	if errExist == nil {
		return ResponseUpdatedAdvertisementFinishedCreate{}, errors.New("advertisement already has route choose")
	}
	if !errors.Is(errExist, sql.ErrNoRows) {
		return ResponseUpdatedAdvertisementFinishedCreate{}, errExist
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
	response.ParseFromUpdatedAdvertisementFinishedCreateObject(
		resultAdRoute.RouteHistID,
		resultAdRoute.RouteChoose,
		result,
	)

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
		response.RouteIndexChoose = index
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

func (p *Service) GetAllAdvertisementByUser(ctx context.Context, id int64) ([]AdvertisementResponseAll, error) {
	results, err := p.InterfaceService.GetAllAdvertisementByUser(ctx, id)
	if err != nil {
		return nil, err
	}

	var announcementResponses []AdvertisementResponseAll
	for _, result := range results {
		var index int
		if result.RouteChoose.Valid {
			index = int(result.RouteChoose.Int64)
		} else {
			index = -1
		}

		var route new_routes.FinalOutput
		if result.ResponseRoutes.RawMessage != nil && len(result.ResponseRoutes.RawMessage) > 0 {
			if errRoute := json.Unmarshal(result.ResponseRoutes.RawMessage, &route); errRoute != nil {
				return nil, errRoute
			}
		}

		var chosenRoute new_routes.RouteOutput
		if index >= 0 && index < len(route.Routes) {
			chosenRoute = route.Routes[index]
		}

		totalFreights, err := p.InterfaceService.CountAdvertisementByUserID(ctx, result.UserID)
		if err != nil {
			return nil, err
		}

		var response AdvertisementResponseAll
		response.RouteIndexChoose = index
		response.RouteChoose = chosenRoute
		response.ActiveFreight = totalFreights

		if result.ActiveThere.Valid {
			response.ActiveDuration = validation.FormatActiveDuration(result.ActiveThere.Time)
		}

		response.ParseFromAdvertisementByIDObject(result)

		announcementResponses = append(announcementResponses, response)
	}

	return announcementResponses, nil
}

func (p *Service) UpdateAdsRouteChooseService(ctx context.Context, data UpdateAdsRouteChooseDTO) error {
	err := p.InterfaceService.UpdateAdvertismentRouteChoose(ctx, db.UpdateAdsRouteChooseByUserIdParams{
		RouteChoose:     data.Request.NewRoute,
		UserID:          data.UserID,
		AdvertisementID: data.Request.AdvertisementID,
	},
	)
	if err != nil {
		return err
	}

	return nil
}

func (p *Service) GetAdvertisementByIDService(ctx context.Context, id int64) (AdvertisementResponseAll, error) {
	result, err := p.InterfaceService.GetAllAdvertisementById(ctx, id)
	if err != nil {
		return AdvertisementResponseAll{}, err
	}

	getAdvertisementResponse := AdvertisementResponseAll{}
	getAdvertisementResponse.ParseFromAdvertisementObject(db.GetAllAdvertisementUsersRow(result))

	var route new_routes.FinalOutput
	if result.ResponseRoutes != nil && len(result.ResponseRoutes) > 0 {
		if errRoute := json.Unmarshal(result.ResponseRoutes, &route); errRoute != nil {
			return AdvertisementResponseAll{}, errRoute
		}
	}

	index := int(result.RouteChoose)
	var chosenRoute new_routes.RouteOutput
	if index >= 0 && index < len(route.Routes) {
		chosenRoute = route.Routes[index]
	}

	getAdvertisementResponse.RouteIndexChoose = index
	getAdvertisementResponse.RouteChoose = chosenRoute

	return getAdvertisementResponse, nil
}

func (p *Service) GetAdvertisementByIDPublicService(ctx context.Context, id int64) (AdvertisementResponseNoUser, error) {
	result, err := p.InterfaceService.GetAllAdvertisementPublicById(ctx, id)
	if err != nil {
		return AdvertisementResponseNoUser{}, err
	}

	getAdvertisementResponse := AdvertisementResponseNoUser{}
	getAdvertisementResponse.ParseFromAdvertisementObject(db.GetAllAdvertisementPublicRow(result))

	return getAdvertisementResponse, nil
}
