package user

import (
	"context"
	"database/sql"
	"errors"

	"geolocation/internal/get_token"
)

type InterfaceService interface {
	DeleteUserService(ctx context.Context, payload get_token.PayloadUserDTO) error
	UpdateUserService(ctx context.Context, data UpdateUserDTO) (UpdateUserResponse, error)
	UpdateUserPersonalInfoService(
		ctx context.Context,
		data UpdateUserPersonalInfoRequest,
	) (UpdateUserPersonalInfoResponse, error)
	UpdateUserAddressService(
		ctx context.Context,
		data UpdateUserAddressRequest,
	) (UpdateUserAddressResponse, error)
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

func (s *Service) DeleteUserService(ctx context.Context, payload get_token.PayloadUserDTO) error {
	err := s.InterfaceService.DeleteUserByIdRepository(ctx, payload.ID)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) UpdateUserService(
	ctx context.Context,
	data UpdateUserDTO,
) (UpdateUserResponse, error) {
	u, err := s.InterfaceService.UpdateUserByIdRepository(ctx, data.ParseToUpdateUserByIdParams())
	if err != nil {
		return UpdateUserResponse{}, err
	}

	return data.ParseToUpdateUserResponse(u), nil
}

func (p *Service) UpdateUserPersonalInfoService(
	ctx context.Context,
	data UpdateUserPersonalInfoRequest,
) (UpdateUserPersonalInfoResponse, error) {
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

	updateUserService := UpdateUserPersonalInfoResponse{}.ParseToUpdateUserPersonalInfoResponse(
		result,
	)

	return updateUserService, nil
}

func (p *Service) UpdateUserAddressService(
	ctx context.Context,
	data UpdateUserAddressRequest,
) (UpdateUserAddressResponse, error) {
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
