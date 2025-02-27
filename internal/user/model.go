package user

import (
	"database/sql"
	"errors"
	db "geolocation/db/sqlc"
	"geolocation/internal/get_token"
	"google.golang.org/api/oauth2/v2"
	"net/mail"
	"regexp"
	"unicode"
)

type CreateUserRequest struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
	Name            string `json:"name"`
	ProfilePicture  string `json:"profile_picture"`
}

func (u CreateUserDTO) Validate() error {
	_, err := mail.ParseAddress(u.Request.Email)

	if err != nil {
		return errors.New("invalid email address")
	}

	if u.Request.Name == "" {
		return errors.New("name is required")
	}

	if !u.Sso {
		if u.Request.Password != u.Request.ConfirmPassword {
			return errors.New("password does not match")
		}

		if len(u.Request.Password) < 8 {
			return errors.New("password must have at least one uppercase letter, one number and one special character")
		}

		var hasUpper, hasDigit, hasSpecial bool

		for _, c := range u.Request.Password {
			switch {

			case unicode.IsUpper(c):
				hasUpper = true
			case unicode.IsDigit(c):
				hasDigit = true
			}
		}

		specialCharRegex := regexp.MustCompile(`[!@#$%^&*()\-_=+\[\]{}|;:'",.<>?/\\` + "`~]")

		hasSpecial = specialCharRegex.MatchString(u.Request.Password)

		if !hasUpper || !hasDigit || !hasSpecial {
			return errors.New("password must have at least one uppercase letter, one number and one special character")
		}

	}

	return nil
}

func (u CreateUserDTO) ParseToCreateUserParams(hash string) db.CreateUserParams {

	if u.Sso {
		return db.CreateUserParams{
			Name:  u.Request.Name,
			Email: u.Request.Email,
			GoogleID: sql.NullString{
				String: u.Payload.UserId,
				Valid:  true,
			},
			ProfilePicture: sql.NullString{
				String: u.Request.ProfilePicture,
				Valid:  true,
			},
		}
	}

	return db.CreateUserParams{
		Name:  u.Request.Name,
		Email: u.Request.Email,
		Password: sql.NullString{
			String: hash,
			Valid:  true,
		},
	}
}

func (u CreateUserDTO) ParseToCreateUserResponse(user db.User) CreateUserResponse {
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

type CreateUserDTO struct {
	Request CreateUserRequest
	Sso     bool
	Payload *oauth2.Tokeninfo
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginDTO struct {
	Request LoginRequest
	Sso     bool
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
