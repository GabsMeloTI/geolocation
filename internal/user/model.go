package user

import (
	"database/sql"
	db "geolocation/db/sqlc"
	"geolocation/internal/get_token"
)

type CreateUserRequest struct {
	Email           string `json:"email" validate:"required"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
	Name            string `json:"name" validate:"required"`
	ProfilePicture  string `json:"profile_picture"`
	Provider        string `json:"provider"`
	GoogleID        string `json:"google_id"`
	Phone           string `json:"phone" validate:"required"`
	Document        string `json:"document" validate:"required"`
}

func (u CreateUserRequest) ParseToCreateUserParams(hash string) db.CreateUserParams {
	return db.CreateUserParams{
		Name:  u.Name,
		Email: u.Email,
		Password: sql.NullString{
			String: hash,
			Valid:  u.Provider != "google",
		},
		GoogleID: sql.NullString{
			String: u.GoogleID,
			Valid:  u.Provider == "google",
		},
		ProfilePicture: sql.NullString{
			String: u.ProfilePicture,
			Valid:  u.Provider == "google",
		},
		Phone: sql.NullString{
			String: u.Phone,
			Valid:  true,
		},
		Document: sql.NullString{
			String: u.Document,
			Valid:  true,
		},
	}
}

func (u CreateUserRequest) ParseToCreateUserResponse(user db.User) CreateUserResponse {
	return CreateUserResponse{
		Name:  user.Name,
		Email: user.Email,
		ID:    user.ID,
	}
}

type CreateUserResponse struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Provider string `json:"provider"`
}

type LoginUserResponse struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	Email          string `json:"email"`
	ProfilePicture string `json:"profile_picture"`
	ProfileId      int64  `json:"profile_id"`
	Document       string `json:"document"`
	Token          string `json:"token"`
}

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
