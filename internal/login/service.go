package login

import (
	"context"
	"database/sql"
	"errors"
	db "geolocation/db/sqlc"
	"geolocation/infra/token"
	"geolocation/pkg/crypt"
	"geolocation/pkg/sso"
	"time"
)

var ErrUserNotFound = errors.New("user not found")
var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrInvalidClientID = errors.New("invalid client ID")

type ServiceInterface interface {
	Login(context.Context, RequestLogin) (ResponseLogin, error)
	CreateUser(context.Context, RequestCreateUser) (ResponseCreateUser, error)
}

type Service struct {
	googleToken    sso.GoogleTokenInterface
	repository     RepositoryInterface
	maker          token.Maker
	googleClientID string
}

func NewService(
	googleToken sso.GoogleTokenInterface,
	repository RepositoryInterface,
	maker token.Maker,
	googleClientID string,
) *Service {
	return &Service{googleToken, repository, maker, googleClientID}
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

		if googleToken.Audience != s.googleClientID {
			return response, ErrInvalidClientID
		}
	}

	result, err := s.repository.GetUser(ctx, db.LoginParams{
		Email:    emailSearch,
		GoogleID: sql.NullString{String: googleIDSearch, Valid: true},
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return response, ErrUserNotFound
		}
		return response, err
	}

	if !crypt.CheckPasswordHash(data.Password, result.Password.String) {
		return response, ErrInvalidCredentials
	}

	tokenStr, err := s.maker.CreateTokenUser(
		result.ID,
		result.Name,
		result.Email,
		result.ProfileID.Int64,
		result.Document.String,
		result.GoogleID.String,
		time.Now().Add(24*time.Hour).UTC(),
	)
	if err != nil {
		return response, err
	}

	profile, err := s.repository.GetProfileById(ctx, result.ProfileID.Int64)

	if err != nil {
		return response, err
	}

	responseData := ResponseLogin{
		ID:      result.ID,
		Name:    result.Name,
		Email:   result.Email,
		Token:   tokenStr,
		Profile: profile.Name,
	}

	return responseData, err
}

func (s *Service) CreateUser(ctx context.Context, data RequestCreateUser) (response ResponseCreateUser, err error) {
	var userEmail, userGoogleID string
	var newPassword sql.NullString
	userEmail = data.Email
	userGoogleID = ""

	hashedPassword, err := crypt.HashPassword(data.Password)
	if err != nil {
		return response, err
	}
	newPassword = sql.NullString{
		String: hashedPassword,
		Valid:  true,
	}

	if data.Token != "" {
		googleToken, err := s.googleToken.Validation(ctx, data.Token)
		if err != nil {
			return response, err
		}

		if googleToken.Audience != s.googleClientID {
			return response, ErrInvalidClientID
		}

		userEmail = googleToken.Email
		userGoogleID = googleToken.UserId
		newPassword = sql.NullString{
			String: "",
			Valid:  false,
		}
	}

	result, err := s.repository.CreateUser(ctx, db.NewCreateUserParams{
		Name:           data.Name,
		Email:          userEmail,
		Password:       newPassword,
		ProfileID:      sql.NullInt64{Int64: data.TypePerson, Valid: true},
		Document:       sql.NullString{String: data.Document, Valid: true},
		Phone:          sql.NullString{String: data.Telephone, Valid: true},
		GoogleID:       sql.NullString{String: userGoogleID, Valid: true},
		ProfilePicture: sql.NullString{String: "", Valid: true},
	})
	if err != nil {
		return response, err
	}

	tokenStr, err := s.maker.CreateTokenUser(
		result.ID,
		result.Name,
		result.Email,
		result.ProfileID.Int64,
		result.Document.String,
		result.GoogleID.String,
		time.Now().Add(24*time.Hour).UTC(),
	)
	if err != nil {
		return response, err
	}

	response.Token = tokenStr
	return response, err
}
