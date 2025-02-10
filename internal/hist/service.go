package routes

import "context"

type InterfaceService interface {
}

type Service struct {
	InterfaceService InterfaceRepository
}

func NewRoutesService(InterfaceService InterfaceRepository) *Service {
	return &Service{InterfaceService}
}

func (s *Service) GetPublicToken(ctx context.Context, ip string) (Response, error) {

}
