package zonas_risco

import (
	"context"
	db "geolocation/db/sqlc"
)

type InterfaceService interface {
	CreateZonaRiscoService(ctx context.Context, data CreateZonaRiscoRequest) (ZonaRiscoResponse, error)
	UpdateZonaRiscoService(ctx context.Context, data UpdateZonaRiscoRequest) (ZonaRiscoResponse, error)
	DeleteZonaRiscoService(ctx context.Context, id int64) error
	GetZonaRiscoByIdService(ctx context.Context, id int64) (ZonaRiscoResponse, error)
	GetAllZonasRiscoService(ctx context.Context) ([]ZonaRiscoResponse, error)
}

type Service struct {
	InterfaceService InterfaceRepository
}

func NewZonasRiscoService(InterfaceService InterfaceRepository) *Service {
	return &Service{InterfaceService}
}

func (s *Service) CreateZonaRiscoService(ctx context.Context, data CreateZonaRiscoRequest) (ZonaRiscoResponse, error) {
	arg := db.CreateZonaRiscoParams{
		Name:   data.Name,
		Cep:    data.Cep,
		Lat:    data.Lat,
		Lng:    data.Lng,
		Radius: data.Radius,
	}
	result, err := s.InterfaceService.CreateZonaRisco(ctx, arg)
	if err != nil {
		return ZonaRiscoResponse{}, err
	}
	resp := ZonaRiscoResponse{}
	resp.ParseFromDb(result)
	return resp, nil
}

func (s *Service) UpdateZonaRiscoService(ctx context.Context, data UpdateZonaRiscoRequest) (ZonaRiscoResponse, error) {
	arg := db.UpdateZonaRiscoParams{
		ID:     data.ID,
		Name:   data.Name,
		Cep:    data.Cep,
		Lat:    data.Lat,
		Lng:    data.Lng,
		Radius: data.Radius,
		Status: data.Status,
	}
	result, err := s.InterfaceService.UpdateZonaRisco(ctx, arg)
	if err != nil {
		return ZonaRiscoResponse{}, err
	}
	resp := ZonaRiscoResponse{}
	resp.ParseFromDb(result)
	return resp, nil
}

func (s *Service) DeleteZonaRiscoService(ctx context.Context, id int64) error {
	return s.InterfaceService.DeleteZonaRisco(ctx, id)
}

func (s *Service) GetZonaRiscoByIdService(ctx context.Context, id int64) (ZonaRiscoResponse, error) {
	result, err := s.InterfaceService.GetZonaRiscoById(ctx, id)
	if err != nil {
		return ZonaRiscoResponse{}, err
	}
	resp := ZonaRiscoResponse{}
	resp.ParseFromDb(result)
	return resp, nil
}

func (s *Service) GetAllZonasRiscoService(ctx context.Context) ([]ZonaRiscoResponse, error) {
	results, err := s.InterfaceService.GetAllZonasRisco(ctx)
	if err != nil {
		return nil, err
	}
	var respList []ZonaRiscoResponse
	for _, r := range results {
		resp := ZonaRiscoResponse{}
		resp.ParseFromDb(r)
		respList = append(respList, resp)
	}
	return respList, nil
}
