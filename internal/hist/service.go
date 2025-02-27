package hist

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	db "geolocation/db/sqlc"
	"geolocation/infra/token"
	"geolocation/validation"
	"strings"
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
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return "", errors.New("IP não pode estar vazio")
	}
	if !validation.IsValidIP(ip) {
		return "", errors.New("IP inválido")
	}

	now := time.Now().UTC()
	tokenHist, err := s.InterfaceService.GetTokenHist(ctx, ip)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("falha ao verificar token no histórico: %w", err)
	}

	// Se o ID for 0, assumimos que o registro não existe
	if tokenHist.ID != 0 {
		if tokenHist.ExpritedAt.After(now) {
			return "", errors.New("Token já gerado para este IP")
		}
		updatedRow, err := s.InterfaceService.UpdateTokenHist(ctx, db.UpdateTokenHistParams{
			ID:            tokenHist.ID,
			NumberRequest: 0,
			ExpritedAt:    now.Add(24 * time.Hour),
		})
		if err != nil {
			return "", fmt.Errorf("falha ao atualizar token no histórico: %w", err)
		}
		tokenHist = db.TokenHist{
			ID:            updatedRow.ID,
			Ip:            updatedRow.Ip,
			NumberRequest: updatedRow.NumberRequest,
			Valid:         updatedRow.Valid,
			ExpritedAt:    updatedRow.ExpritedAt,
		}
	} else {
		tokenHist, err = s.InterfaceService.CreateTokenHist(ctx, db.CreateTokenHistParams{
			Ip:            ip,
			NumberRequest: 0,
			ExpritedAt:    now.Add(24 * time.Hour),
		})
		if err != nil {
			return "", fmt.Errorf("falha ao criar token no histórico: %w", err)
		}
	}

	arg := Request{
		ID:             tokenHist.ID,
		IP:             tokenHist.Ip,
		NumberRequests: tokenHist.NumberRequest,
		Valid:          tokenHist.Valid.Bool,
		ExpiredAt:      tokenHist.ExpritedAt,
	}

	strToken, errT := s.createToken(arg)
	if errT != nil {
		return "", fmt.Errorf("falha ao criar token: %w", errT)
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
