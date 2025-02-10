package routes

import (
	"context"
	db "geolocation/db/sqlc"
	"time"
)

type InterfaceService interface {
}

type Service struct {
	InterfaceService InterfaceRepository
}

func NewRoutesService(InterfaceService InterfaceRepository) *Service {
	return &Service{InterfaceService}
}

func (s *Service) GetPublicToken(ctx context.Context, ip string) (Response, error) {
	s.InterfaceService.CreateTokenHist(ctx, db.CreateTokenHistParams{
		Ip:            ip,
		NumberRequest: 0,
	})

	s.InterfaceService.CreateRouteHist(ctx, db.CreateTokenHistParams{})
}
