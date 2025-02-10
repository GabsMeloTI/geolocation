package hist

import (
	"context"
	"database/sql"
	db "geolocation/db/sqlc"
	"geolocation/internal/routes"
	"strings"
)

type InterfaceService interface {
	GetPublicToken(ctx context.Context, ip string, info routes.FrontInfo) (Response, error)
	UpdateNumberOfRequest(ctx context.Context, id int64) error
}

type Service struct {
	InterfaceService InterfaceRepository
}

func NewRoutesService(InterfaceService InterfaceRepository) *Service {
	return &Service{InterfaceService}
}

func (s *Service) GetPublicToken(ctx context.Context, ip string, info routes.FrontInfo) (Response, error) {
	resultToken, errToken := s.InterfaceService.CreateTokenHist(ctx, db.CreateTokenHistParams{
		Ip:            ip,
		NumberRequest: 0,
	})
	if errToken != nil {
		return Response{}, errToken
	}

	_, errRoute := s.InterfaceService.CreateRouteHist(ctx, db.CreateRouteHistParams{
		IDTokenHist: resultToken.ID,
		Origin:      info.Origin,
		Destination: info.Destination,
		Waypoints: sql.NullString{
			String: strings.ToLower(strings.Join(info.Waypoints, ",")),
			Valid:  true,
		},
	})
	if errRoute != nil {
		return Response{}, errRoute
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
