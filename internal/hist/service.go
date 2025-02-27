package hist

import (
	"context"
	"errors"
	db "geolocation/db/sqlc"
	"geolocation/infra/token"
	"time"
)

type InterfaceService interface {
	GetPublicToken(ctx context.Context, ip string) (string, error)
}

type Service struct {
	InterfaceService InterfaceRepository
	SignatureString  string
}

func NewHistService(InterfaceService InterfaceRepository, SignatureString string) *Service {
	return &Service{InterfaceService, SignatureString}
}

func (s *Service) GetPublicToken(ctx context.Context, ip string) (string, error) {
	resultToken, errToken := s.InterfaceService.CreateTokenHist(ctx, db.CreateTokenHistParams{
		Ip:            ip,
		NumberRequest: 0,
		ExpritedAt:    time.Now().Add(24 * time.Hour).UTC(),
	})
	if errToken != nil {
		return "", errToken
	}

	arg := Request{
		ID:             resultToken.ID,
		IP:             resultToken.Ip,
		NumberRequests: resultToken.NumberRequest,
		Valid:          resultToken.Valid.Bool,
		ExpiredAt:      resultToken.ExpritedAt,
	}
	strToken, errT := s.createToken(arg)
	if errT != nil {
		return "", errT
	}

	return strToken, nil
}

func (s *Service) createToken(data Request) (string, error) {
	maker, err := token.NewPasetoMaker(s.SignatureString)
	if err != nil {
		return "", errors.New("failed")
	}

	strToken, err := maker.CreateToken(
		data.ID,
		data.IP,
		data.NumberRequests,
		data.Valid,
		data.ExpiredAt,
	)

	if err != nil {
		return "", errors.New("failed")
	}

	return strToken, nil
}
