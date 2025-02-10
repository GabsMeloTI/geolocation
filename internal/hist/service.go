package routes

type InterfaceService interface {
}

type Service struct {
	InterfaceService InterfaceRepository
}

func NewRoutesService(InterfaceService InterfaceRepository) *Service {
	return &Service{InterfaceService}
}

//func (s *Service) GetPublicToken(ctx context.Context, ip string, info routes.FrontInfo) (Response, error) {
//	resultToken, errToken := s.InterfaceService.CreateTokenHist(ctx, db.CreateTokenHistParams{
//		Ip:            ip,
//		NumberRequest: 0,
//	})
//	if errToken != nil {
//		return Response{}, errToken
//	}
//
//	resultRoute, errRoute := s.InterfaceService.CreateRouteHist(ctx, db.CreateRouteHistParams{
//		IDTokenHist: resultToken.ID,
//		Origin:      info.Origin,
//		Destination: info.Destination,
//		Waypoints: sql.NullString{
//			String: info.Waypoints,
//			Valid:  true,
//		},
//	})
//	if errRoute != nil {
//		return Response{}, errRoute
//	}
//
//}
