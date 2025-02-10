package hist

import (
	"context"
	db "geolocation/db/sqlc"
)

type InterfaceService interface {
	GetPublicToken(ctx context.Context, ip string) (Response, error)
	UpdateNumberOfRequest(ctx context.Context, id int64) error
}

type Service struct {
	InterfaceService InterfaceRepository
}

func NewRoutesService(InterfaceService InterfaceRepository) *Service {
	return &Service{InterfaceService}
}

func (s *Service) GetPublicToken(ctx context.Context, ip string) (Response, error) {
	resultToken, errToken := s.InterfaceService.CreateTokenHist(ctx, db.CreateTokenHistParams{
		Ip:            ip,
		NumberRequest: 0,
	})
	if errToken != nil {
		return Response{}, errToken
	}

	response := Response{
		ID: resultToken.ID,
		IP: resultToken.Ip,
	}

	return response, nil
}

func (s *Service) UpdateNumberOfRequest(ctx context.Context, id int64) error {
	err := s.InterfaceService.UpdateNumberOfRequest(ctx, db.UpdateNumberOfRequestParams{
		NumberRequest: +1,
		ID:            id,
	})
	if err != nil {
		return err
	}

	return nil
}
