package plans

import (
	"context"
)

type InterfaceService interface {
	CreatePlansService(ctx context.Context, data CreatePlansRequest) (PlansResponse, error)
	UpdatePlansService(ctx context.Context, data UpdatePlansRequest) (PlansResponse, error)
}

type Service struct {
	InterfaceService InterfaceRepository
}

func NewPlansService(InterfaceService InterfaceRepository) *Service {
	return &Service{InterfaceService}
}

func (p *Service) CreatePlansService(ctx context.Context, data CreatePlansRequest) (PlansResponse, error) {
	arg := data.ParseCreateToPlans()

	result, err := p.InterfaceService.CreatePlans(ctx, arg)
	if err != nil {
		return PlansResponse{}, err
	}

	createPlansService := PlansResponse{}
	createPlansService.ParseFromPlansObject(result)

	return createPlansService, nil
}

func (p *Service) UpdatePlansService(ctx context.Context, data UpdatePlansRequest) (PlansResponse, error) {
	arg := data.ParseUpdateToPlans()

	result, err := p.InterfaceService.UpdatePlans(ctx, arg)
	if err != nil {
		return PlansResponse{}, err
	}

	updatePlansService := PlansResponse{}
	updatePlansService.ParseFromPlansObject(result)

	return updatePlansService, nil
}
