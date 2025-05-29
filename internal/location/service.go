package location

import (
	"context"
	"database/sql"
	"errors"
	db "geolocation/db/sqlc"
	"geolocation/internal/get_token"
)

type InterfaceService interface {
	CreateLocationService(ctx context.Context, data CreateLocationDTO) (LocationResponse, error)
	GetLocationService(ctx context.Context, provider int64, payload get_token.PayloadDTO) ([]LocationResponse, error)
	UpdateLocationService(ctx context.Context, data UpdateLocationDTO) (LocationResponse, error)
	DeleteLocationService(ctx context.Context, id int64, payload get_token.PayloadDTO) error
}

type Service struct {
	InterfaceService InterfaceRepository
}

func NewLocationsService(InterfaceService InterfaceRepository) *Service {
	return &Service{InterfaceService}
}

func (s *Service) CreateLocationService(ctx context.Context, data CreateLocationDTO) (LocationResponse, error) {
	_, err := s.InterfaceService.GetLocationByName(ctx, db.GetLocationByNameParams{
		AccessID:       data.Payload.AccessID,
		TenantID:       data.Payload.TenantID,
		IDProviderInfo: data.CreateLocationRequest.ProviderInfoID,
		Type:           data.CreateLocationRequest.Type,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return LocationResponse{}, err
	}
	if err == nil {
		return LocationResponse{}, errors.New("já existe um local deste tipo")
	}

	locParams := data.ParseCreateToLocation()
	locParams.AccessID = data.Payload.AccessID
	locParams.TenantID = data.Payload.TenantID

	loc, err := s.InterfaceService.CreateLocation(ctx, locParams)
	if err != nil {
		return LocationResponse{}, err
	}

	var responses []GetAreasResponse
	for _, a := range data.CreateLocationRequest.Area {
		params := a.ParseCreateToArea()
		params.LocationsID = loc.ID

		created, err := s.InterfaceService.CreateArea(ctx, params)
		if err != nil {
			return LocationResponse{}, err
		}

		responses = append(responses, GetAreasResponse{
			ID:          created.ID,
			Latitude:    created.Latitude,
			Longitude:   created.Longitude,
			Description: created.Description,
		})
	}

	var resp LocationResponse
	resp.ParseFromPlansObject(loc, responses)

	return resp, nil
}

func (s *Service) UpdateLocationService(ctx context.Context, data UpdateLocationDTO) (LocationResponse, error) {
	_, err := s.InterfaceService.GetLocationByNameExcludingID(ctx, db.GetLocationByNameExcludingIDParams{
		AccessID:       data.Payload.AccessID,
		TenantID:       data.Payload.TenantID,
		Type:           data.UpdateLocationRequest.Type,
		ID:             data.UpdateLocationRequest.ID,
		IDProviderInfo: data.UpdateLocationRequest.ProviderInfoID,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return LocationResponse{}, err
	}
	if err == nil {
		return LocationResponse{}, errors.New("já existe um local deste tipo")
	}

	arg := data.ParseUpdateToLocation()
	updatedLoc, err := s.InterfaceService.UpdateLocation(ctx, arg)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return LocationResponse{}, errors.New("location does not exist")
		}
		return LocationResponse{}, err
	}

	err = s.InterfaceService.DeleteArea(ctx, updatedLoc.ID)
	if err != nil {
		return LocationResponse{}, err
	}

	var updatedAreas []GetAreasResponse
	for _, a := range data.UpdateLocationRequest.Areas {
		a.LocationsID = updatedLoc.ID

		argArea := db.CreateAreaParams{
			LocationsID: a.LocationsID,
			Latitude:    a.Latitude,
			Longitude:   a.Longitude,
			Description: a.Description,
		}
		areaResult, err := s.InterfaceService.CreateArea(ctx, argArea)
		if err != nil {
			return LocationResponse{}, err
		}

		updatedAreas = append(updatedAreas, GetAreasResponse{
			ID:          areaResult.ID,
			Latitude:    areaResult.Latitude,
			Longitude:   areaResult.Longitude,
			Description: areaResult.Description,
		})
	}

	var resp LocationResponse
	resp.ParseFromPlansObject(updatedLoc, updatedAreas)
	return resp, nil
}

func (s *Service) DeleteLocationService(ctx context.Context, id int64, payload get_token.PayloadDTO) error {
	arg := db.DeleteLocationParams{
		ID:       id,
		AccessID: payload.AccessID,
		TenantID: payload.TenantID,
	}

	err := s.InterfaceService.DeleteLocation(ctx, arg)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("location does not exist")
		}
		return err
	}

	err = s.InterfaceService.DeleteArea(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("location does not exist")
		}
		return err
	}

	return nil
}

func (s *Service) GetLocationService(ctx context.Context, provider int64, payload get_token.PayloadDTO) ([]LocationResponse, error) {
	locs, err := s.InterfaceService.GetLocationByOrg(ctx, db.GetLocationByOrgParams{
		IDProviderInfo: provider,
		AccessID:       payload.AccessID,
		TenantID:       payload.TenantID,
	})
	if err != nil {
		return nil, err
	}

	var responses []LocationResponse
	for _, loc := range locs {
		areaRows, err := s.InterfaceService.GetAreasByOrg(ctx, loc.ID)
		if err != nil {
			return nil, err
		}

		var areas []GetAreasResponse
		for _, r := range areaRows {
			areas = append(areas, GetAreasResponse{
				ID:          r.ID,
				Latitude:    r.Latitude,
				Longitude:   r.Longitude,
				Description: r.Description,
			})
		}

		var resp LocationResponse
		resp.ParseFromPlansObject(loc, areas)
		responses = append(responses, resp)
	}

	return responses, nil
}
