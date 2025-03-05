package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"geolocation/infra/token"
	"geolocation/internal/get_token"
	"geolocation/pkg/crypt"
	"time"
)

type InterfaceService interface {
	CreateUserService(ctx context.Context, data CreateUserRequest) (CreateUserResponse, error)
	UserLoginService(ctx context.Context, data LoginRequest) (LoginUserResponse, error)
	DeleteUserService(ctx context.Context, payload get_token.PayloadUserDTO) error
	UpdateUserService(ctx context.Context, data UpdateUserDTO) (UpdateUserResponse, error)
	UpdateUserPersonalInfoService(ctx context.Context, data UpdateUserPersonalInfoRequest) (UpdateUserPersonalInfoResponse, error)
	UpdateUserAddressService(ctx context.Context, data UpdateUserAddressRequest) (UpdateUserAddressResponse, error)
	GetUserService(ctx context.Context, userId int64) (GetUserResponse, error)
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

func (s *Service) CreateUserService(ctx context.Context, data CreateUserRequest) (CreateUserResponse, error) {
	u, err := s.InterfaceService.GetUserByEmailRepository(ctx, data.Email)

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return CreateUserResponse{}, err
		}
	}

	if u.ID != 0 {
		return CreateUserResponse{}, errors.New("email already exists")
	}

	_, err = s.InterfaceService.GetProfileByIdRepository(ctx, data.ProfileId)

	if err != nil {
		return CreateUserResponse{}, err
	}

	hashedPassword, err := crypt.HashPassword(data.Password)

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

func (s *Service) UserLoginService(ctx context.Context, data LoginRequest) (LoginUserResponse, error) {
	result, err := s.InterfaceService.GetUserByEmailRepository(ctx, data.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return LoginUserResponse{}, errors.New("invalid credentials")
		}
		return LoginUserResponse{}, err
	}

	if data.Provider != "google" {
		if !crypt.CheckPasswordHash(data.Password, result.Password.String) {
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

func (s *Service) DeleteUserService(ctx context.Context, payload get_token.PayloadUserDTO) error {
	err := s.InterfaceService.DeleteUserByIdRepository(ctx, payload.ID)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) UpdateUserService(ctx context.Context, data UpdateUserDTO) (UpdateUserResponse, error) {
	u, err := s.InterfaceService.UpdateUserByIdRepository(ctx, data.ParseToUpdateUserByIdParams())

	if err != nil {
		return UpdateUserResponse{}, err
	}

	return data.ParseToUpdateUserResponse(u), nil
}

func (p *Service) UpdateUserPersonalInfoService(ctx context.Context, data UpdateUserPersonalInfoRequest) (UpdateUserPersonalInfoResponse, error) {
	_, err := p.InterfaceService.GetUserById(ctx, data.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return UpdateUserPersonalInfoResponse{}, errors.New("user not found")
	}
	if err != nil {
		return UpdateUserPersonalInfoResponse{}, err
	}

	arg := data.ParseToUpdateUserPersonalInfoParams()

	result, err := p.InterfaceService.UpdateUserPersonalInfo(ctx, arg)
	if err != nil {
		return UpdateUserPersonalInfoResponse{}, err
	}

	updateUserService := UpdateUserPersonalInfoResponse{}.ParseToUpdateUserPersonalInfoResponse(result)

	return updateUserService, nil
}

func (p *Service) UpdateUserAddressService(ctx context.Context, data UpdateUserAddressRequest) (UpdateUserAddressResponse, error) {
	_, err := p.InterfaceService.GetUserById(ctx, data.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return UpdateUserAddressResponse{}, errors.New("user not found")
	}
	if err != nil {
		return UpdateUserAddressResponse{}, err
	}

	arg := data.ParseToUpdateUserAddressParams()

	result, err := p.InterfaceService.UpdateUserAddress(ctx, arg)
	if err != nil {
		return UpdateUserAddressResponse{}, err
	}

	updateUserService := UpdateUserAddressResponse{}.ParseToUpdateUserAddressResponse(result)

	return updateUserService, nil
}

func (s *Service) GetUserService(ctx context.Context, userId int64) (GetUserResponse, error) {
	var res GetUserResponse

	user, err := s.InterfaceService.GetUserById(ctx, userId)

	if err != nil {
		return res, err
	}

	return res.ParseFromDbUser(user), nil
}
