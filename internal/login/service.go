package login

import (
	"context"
	"database/sql"
	"errors"
	db "geolocation/db/sqlc"
	"geolocation/infra/token"
	"geolocation/pkg/sso"
	"time"
)

var ErrUserNotFound = errors.New("user not found")

type ServiceInterface interface {
	Login(context.Context, RequestLogin) (ResponseLogin, error)
}

type Service struct {
	googleToken sso.GoogleTokenInterface
	repository  RepositoryInterface
	maker       token.Maker
}

func NewService(googleToken sso.GoogleTokenInterface, repository RepositoryInterface, maker token.Maker) *Service {
	return &Service{googleToken, repository, maker}
}

func (s *Service) Login(ctx context.Context, data RequestLogin) (response ResponseLogin, err error) {
	var emailSearch string
	var googleIDSearch string
	emailSearch = data.Username
	googleIDSearch = ""
	if data.Token != "" {
		googleToken, err := s.googleToken.Validation(ctx, data.Token)
		if err != nil {
			return response, err
		}
		emailSearch = googleToken.Email
		googleIDSearch = googleToken.UserId
	}

	result, err := s.repository.GetUser(ctx, db.LoginParams{
		Email: emailSearch,
		GoogleID: sql.NullString{
			String: googleIDSearch,
			Valid:  true,
		},
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return response, ErrUserNotFound
		}
		return response, err
	}

	tokenStr, err := s.maker.CreateTokenUser(result.ID, result.Name, result.Email, result.ProfileID.Int64, result.Document.String, result.GoogleID.String, time.Now().Add(24*time.Hour).UTC())
	if err != nil {
		return response, err
	}

	response.Token = tokenStr
	return response, err
}
