package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"geolocation/infra/token"
	"geolocation/pkg/crypt"
	"time"
)

type InterfaceService interface {
	CreateUserService(ctx context.Context, data CreateUserDTO) (CreateUserResponse, error)
	UserLogin(ctx context.Context, data LoginDTO) (LoginUserResponse, error)
}

type Service struct {
	InterfaceService InterfaceRepository
	SignatureString  string
}

func NewUserService(interfaceService InterfaceRepository, SignatureString string) *Service {
	return &Service{
		InterfaceService: interfaceService,
		SignatureString:  SignatureString,
	}
}

func (s *Service) CreateUserService(ctx context.Context, data CreateUserDTO) (CreateUserResponse, error) {
	u, err := s.InterfaceService.GetUserByEmailRepository(ctx, data.Request.Email)

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return CreateUserResponse{}, err
		}
	}

	if u.ID != 0 {
		return CreateUserResponse{}, errors.New("email already exists")
	}

	hashedPassword, err := crypt.HashPassword(data.Request.Password)

	if err != nil {
		return CreateUserResponse{}, err
	}

	user, err := s.InterfaceService.CreateUserRepository(ctx, data.ParseToCreateUserParams(hashedPassword))

	if err != nil {
		return CreateUserResponse{}, err
	}

	res := data.ParseToCreateUserResponse(user)

	return res, nil
}

func (s *Service) UserLogin(ctx context.Context, data LoginDTO) (LoginUserResponse, error) {
	result, err := s.InterfaceService.GetUserByEmailRepository(ctx, data.Request.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return LoginUserResponse{}, errors.New("invalid credentials")
		}
		return LoginUserResponse{}, err
	}

	if !data.Sso {
		if !crypt.CheckPasswordHash(data.Request.Password, result.Password.String) {
			return LoginUserResponse{}, errors.New("invalid credentials")
		}
	}

	symetricKey := s.SignatureString

	maker, err := token.NewPasetoMaker(symetricKey)
	if err != nil {
		return LoginUserResponse{}, fmt.Errorf("token creation error: %w", err)
	}

	tokenStr, err := maker.CreateTokenUser(result.ID, result.Name, result.Email, result.ProfileID.Int64, result.Document.String, result.GoogleID.String, time.Now().Add(24*time.Hour).UTC())
	if err != nil {
		return LoginUserResponse{}, fmt.Errorf("token generation failed: %w", err)
	}

	return LoginUserResponse{
		ID:             result.ID,
		Name:           result.Name,
		Email:          result.Email,
		ProfilePicture: result.ProfilePicture.String,
		ProfileId:      result.ProfileID.Int64,
		Document:       result.Document.String,
		Token:          tokenStr,
	}, nil
}
