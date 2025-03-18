package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	db "geolocation/db/sqlc"
	"geolocation/infra/token"
	"geolocation/internal/get_token"
	"geolocation/pkg/crypt"
	"geolocation/pkg/email"
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
	RecoverPasswordService(ctx context.Context, data RecoverPasswordRequest) error
	ConfirmRecoverPasswordService(
		ctx context.Context,
		data ConfirmRecoverPasswordDTO,
	) error
	GetUserByEmailService(
		ctx context.Context,
		email string,
	) (GetUserResponse, error)
	UpdateUserPassword(ctx context.Context, userId int64, data UpdatePasswordRequest) error
}

type Service struct {
	InterfaceService InterfaceRepository
	maker            token.Maker
	sendEmail        *email.SendEmail
}

func NewUserService(
	interfaceService InterfaceRepository,
	maker token.Maker,
	sendEmail *email.SendEmail,
) *Service {
	return &Service{
		InterfaceService: interfaceService,
		maker:            maker,
		sendEmail:        sendEmail,
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

func (s *Service) RecoverPasswordService(ctx context.Context, data RecoverPasswordRequest) error {
	user, err := s.InterfaceService.GetUserByEmailRepository(ctx, data.Email)
	if err != nil {
		return err
	}

	token, err := s.maker.CreateTokenUser(
		user.ID,
		user.Name,
		user.Email,
		user.ProfileID.Int64,
		user.Document.String,
		user.GoogleID.String,
		time.Now().Add(24*time.Hour).UTC(),
	)
	if err != nil {
		return err
	}

	urlBase := "https://easyfrete.com.br/"
	linkProvider := fmt.Sprintf("%spassword_reset?token=%s", urlBase, token)

	tmp, err := s.sendEmail.NewTemplate(email.EmailPlaceHolder{
		NameProvider: user.Name,
		Link:         linkProvider,
	}, "recover_password.html")
	if err != nil {
		return err
	}

	err = s.sendEmail.SendEmailNew(
		*tmp,
		user.Email,
		"Redefinição de senha",
	)
	if err != nil {
		fmt.Println(err)
		return err
	}
	err = s.InterfaceService.CreateHistoryRecoverPasswordRepository(
		ctx,
		db.CreateHistoryRecoverPasswordParams{
			UserID: user.ID,
			Email:  user.Email,
			Token:  token,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) ConfirmRecoverPasswordService(ctx context.Context, data ConfirmRecoverPasswordDTO) error {
	_, err := s.InterfaceService.GetUserById(ctx, data.UserID)
	if err != nil {
		return err
	}

	hashedPassword, err := crypt.HashPassword(data.Request.Password)
	if err != nil {
		return err
	}

	err = s.InterfaceService.UpdatePasswordByUserIdRepository(ctx, db.UpdatePasswordByUserIdParams{
		Password: sql.NullString{
			String: hashedPassword,
			Valid:  true,
		},
		ID: data.UserID,
	})
	if err != nil {
		return err
	}

	err = s.InterfaceService.UpdateHistoryPasswordRecoverRepository(ctx, data.Token)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) GetUserByEmailService(ctx context.Context, email string) (GetUserResponse, error) {
	user, err := s.InterfaceService.GetUserByEmailRepository(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return GetUserResponse{}, errors.New("user not found")
		}
		return GetUserResponse{}, err
	}

	res := GetUserResponse{
		ID:               user.ID,
		Name:             user.Name,
		Email:            email,
		CreatedAt:        user.CreatedAt.Time,
		ProfileID:        user.ProfileID.Int64,
		Document:         user.Document.String,
		State:            user.State.String,
		City:             user.City.String,
		Neighborhood:     user.Neighborhood.String,
		Street:           user.Street.String,
		StreetNumber:     user.StreetNumber.String,
		Phone:            user.Phone.String,
		ProfilePicture:   user.ProfilePicture.String,
		Cep:              user.Cep.String,
		Complement:       user.Complement.String,
		DateOfBirth:      user.DateOfBirth.Time,
		SecondaryContact: user.SecondaryContact.String,
	}

	return res, nil
}

func (s *Service) UpdateUserPassword(ctx context.Context, userId int64, data UpdatePasswordRequest) error {
	user, err := s.InterfaceService.GetUserById(ctx, userId)
	if err != nil {
		return fmt.Errorf("erro ao buscar usuário: %v", err)
	}

	if !crypt.CheckPasswordHash(data.OldPassword, user.Password.String) {
		return fmt.Errorf("senha antiga incorreta")
	}

	if data.Password != data.ConfirmPassword {
		return fmt.Errorf("a nova senha e a confirmação não coincidem")
	}

	if data.OldPassword == data.ConfirmPassword {
		return fmt.Errorf("a nova senha não pode ser igual à senha antiga")
	}

	hashedPassword, err := crypt.HashPassword(data.Password)
	if err != nil {
		return err
	}

	arg := db.UpdateUserPasswordByIdParams{
		ID: userId,
		Password: sql.NullString{
			String: hashedPassword,
			Valid:  true,
		},
	}
	err = s.InterfaceService.UpdateUserPasswordById(ctx, arg)
	if err != nil {
		return err
	}

	return nil
}
