package user

import (
	"database/sql"
	"time"

	db "geolocation/db/sqlc"
	"geolocation/internal/get_token"
)

type UpdateUserRequest struct {
	Name           string `json:"name"`
	ProfilePicture string `json:"profile_picture"`
	State          string `json:"state"`
	City           string `json:"city"`
	Neighborhood   string `json:"neighborhood"`
	Street         string `json:"street"`
	StreetNumber   string `json:"street_number"`
	Phone          string `json:"phone"`
}

type UpdateUserDTO struct {
	Request UpdateUserRequest
	Payload get_token.PayloadUserDTO
}

type UpdateUserResponse struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	Email          string `json:"email"`
	ProfileID      int64  `json:"profile_id,omitempty"`
	Document       string `json:"document,omitempty"`
	State          string `json:"state,omitempty"`
	City           string `json:"city,omitempty"`
	Neighborhood   string `json:"neighborhood,omitempty"`
	Street         string `json:"street,omitempty"`
	StreetNumber   string `json:"street_number,omitempty"`
	Phone          string `json:"phone,omitempty"`
	GoogleID       string `json:"google_id,omitempty"`
	ProfilePicture string `json:"profile_picture,omitempty"`
}

type UpdateUserAddressRequest struct {
	Complement   string `json:"complement"`
	State        string `json:"state"`
	City         string `json:"city"`
	Neighborhood string `json:"neighborhood"`
	Street       string `json:"street"`
	StreetNumber string `json:"street_number"`
	Cep          string `json:"cep"`
	ID           int64  `json:"id"`
}

type UpdateUserPersonalInfoRequest struct {
	Name             string    `json:"name"`
	Document         string    `json:"document"`
	Email            string    `json:"email"`
	Phone            string    `json:"phone"`
	DateOfBirth      time.Time `json:"date_of_birth"`
	SecondaryContact string    `json:"secondary_contact"`
	ID               int64     `json:"id"`
}

type UpdateUserPersonalInfoResponse struct {
	ID               int64     `json:"id"`
	Name             string    `json:"name"`
	Email            string    `json:"email"`
	Document         string    `json:"document"`
	DateOfBirth      time.Time `json:"date_of_birth"`
	SecondaryContact string    `json:"secondary_contact"`
	Phone            string    `json:"phone"`
}

type UpdateUserAddressResponse struct {
	ID           int64  `json:"id"`
	State        string `json:"state"`
	City         string `json:"city"`
	Neighborhood string `json:"neighborhood"`
	Street       string `json:"street"`
	StreetNumber string `json:"street_number"`
	Cep          string `json:"cep"`
	Complement   string `json:"complement"`
}

type UpdatePasswordRequest struct {
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
	OldPassword     string `json:"old_password"`
}

func (u UpdateUserDTO) ParseToUpdateUserByIdParams() db.UpdateUserByIdParams {
	return db.UpdateUserByIdParams{
		Name: u.Request.Name,
		ProfilePicture: sql.NullString{
			String: u.Request.ProfilePicture,
			Valid:  true,
		},
		State: sql.NullString{
			String: u.Request.State,
			Valid:  true,
		},
		City: sql.NullString{
			String: u.Request.City,
			Valid:  true,
		},
		Neighborhood: sql.NullString{
			String: u.Request.Neighborhood,
			Valid:  true,
		},
		Street: sql.NullString{
			String: u.Request.Street,
			Valid:  true,
		},
		StreetNumber: sql.NullString{
			String: u.Request.StreetNumber,
			Valid:  true,
		},
		Phone: sql.NullString{
			String: u.Request.Phone,
			Valid:  true,
		},
		ID: u.Payload.ID,
	}
}

func (u UpdateUserAddressRequest) ParseToUpdateUserAddressParams() db.UpdateUserAddressParams {
	return db.UpdateUserAddressParams{
		Complement: sql.NullString{
			String: u.Complement,
			Valid:  true,
		},
		State: sql.NullString{
			String: u.State,
			Valid:  true,
		},
		City: sql.NullString{
			String: u.City,
			Valid:  true,
		},
		Neighborhood: sql.NullString{
			String: u.Neighborhood,
			Valid:  true,
		},
		Street: sql.NullString{
			String: u.Street,
			Valid:  true,
		},
		Cep: sql.NullString{
			String: u.Cep,
			Valid:  true,
		},
		ID: u.ID,
		StreetNumber: sql.NullString{
			String: u.StreetNumber,
			Valid:  true,
		},
	}
}

func (u UpdateUserPersonalInfoRequest) ParseToUpdateUserPersonalInfoParams() db.UpdateUserPersonalInfoParams {
	return db.UpdateUserPersonalInfoParams{
		Name: u.Name,
		Document: sql.NullString{
			String: u.Document,
			Valid:  true,
		},
		Email: u.Email,
		Phone: sql.NullString{
			String: u.Phone,
			Valid:  true,
		},
		ID: u.ID,
		SecondaryContact: sql.NullString{
			String: u.SecondaryContact,
			Valid:  true,
		},
		DateOfBirth: sql.NullTime{
			Time:  u.DateOfBirth,
			Valid: true,
		},
	}
}

func (u UpdateUserDTO) ParseToUpdateUserResponse(user db.User) UpdateUserResponse {
	return UpdateUserResponse{
		ID:             user.ID,
		Name:           user.Name,
		Email:          user.Email,
		ProfileID:      user.ProfileID.Int64,
		Document:       user.Document.String,
		State:          user.State.String,
		City:           user.City.String,
		Neighborhood:   user.Neighborhood.String,
		Street:         user.Street.String,
		StreetNumber:   user.StreetNumber.String,
		Phone:          user.Phone.String,
		GoogleID:       user.GoogleID.String,
		ProfilePicture: user.ProfilePicture.String,
	}
}

func (u UpdateUserPersonalInfoResponse) ParseToUpdateUserPersonalInfoResponse(
	user db.User,
) UpdateUserPersonalInfoResponse {
	return UpdateUserPersonalInfoResponse{
		ID:               user.ID,
		Name:             user.Name,
		Email:            user.Email,
		Document:         user.Document.String,
		Phone:            user.Phone.String,
		DateOfBirth:      user.DateOfBirth.Time,
		SecondaryContact: user.SecondaryContact.String,
	}
}

func (u UpdateUserAddressResponse) ParseToUpdateUserAddressResponse(
	user db.User,
) UpdateUserAddressResponse {
	return UpdateUserAddressResponse{
		ID:           user.ID,
		State:        user.State.String,
		City:         user.City.String,
		Neighborhood: user.Neighborhood.String,
		Street:       user.Street.String,
		StreetNumber: user.StreetNumber.String,
		Cep:          user.Cep.String,
		Complement:   user.Complement.String,
	}
}

type GetUserResponse struct {
	ID               int64     `json:"id"`
	Name             string    `json:"name"`
	Email            string    `json:"email"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	ProfileID        int64     `json:"profile_id"`
	Document         string    `json:"document"`
	State            string    `json:"state"`
	City             string    `json:"city"`
	Neighborhood     string    `json:"neighborhood"`
	Street           string    `json:"street"`
	StreetNumber     string    `json:"street_number"`
	Phone            string    `json:"phone"`
	ProfilePicture   string    `json:"profile_picture"`
	Cep              string    `json:"cep"`
	Complement       string    `json:"complement"`
	DateOfBirth      time.Time `json:"date_of_birth"`
	SecondaryContact string    `json:"secondary_contact"`
}

func (u GetUserResponse) ParseFromDbUser(user db.User) GetUserResponse {
	return GetUserResponse{
		ID:               user.ID,
		Name:             user.Name,
		Email:            user.Email,
		CreatedAt:        user.CreatedAt.Time,
		UpdatedAt:        user.UpdatedAt.Time,
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
}

type RecoverPasswordRequest struct {
	Email string `json:"email"`
}

type ConfirmRecoverPasswordRequest struct {
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
}

type ConfirmRecoverPasswordDTO struct {
	Request ConfirmRecoverPasswordRequest
	Token   string
	UserID  int64
}

type UserExitsRequest struct {
	Email string `json:"email"`
}
