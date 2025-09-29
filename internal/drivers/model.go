package drivers

import (
	"database/sql"
	"strings"
	"time"
	"unicode"

	db "geolocation/db/sqlc"
)

type CreateDriverRequest struct {
	Name                  string    `json:"name"`
	BirthDate             time.Time `json:"birth_date"`
	Cpf                   string    `json:"cpf"`
	LicenseNumber         string    `json:"license_number"`
	LicenseCategory       string    `json:"license_category"        validate:"oneof=a b c d e"`
	LicenseExpirationDate time.Time `json:"license_expiration_date"`
	CEP                   string    `json:"cep"`
	State                 string    `json:"state"`
	City                  string    `json:"city"`
	Neighborhood          string    `json:"neighborhood"`
	Street                string    `json:"street"`
	StreetNumber          string    `json:"street_number"`
	Complement            string    `json:"complement"`
	Phone                 string    `json:"phone"`
	//Email                 string    `json:"email"`
}

type CreateDriverDto struct {
	CreateDriverRequest CreateDriverRequest
	UserID              int64 `json:"user_id"`
	ProfileId           int64 `json:"profile_id"`
}

type UpdateDriverRequest struct {
	Name                  string    `json:"name"`
	BirthDate             time.Time `json:"birth_date"`
	Cpf                   string    `json:"cpf"`
	LicenseNumber         string    `json:"license_number"`
	LicenseCategory       string    `json:"license_category"        validate:"oneof=a b c d e"`
	LicenseExpirationDate time.Time `json:"license_expiration_date"`
	CEP                   string    `json:"cep"`
	State                 string    `json:"state"`
	City                  string    `json:"city"`
	Neighborhood          string    `json:"neighborhood"`
	Street                string    `json:"street"`
	StreetNumber          string    `json:"street_number"`
	Complement            string    `json:"complement"`
	Phone                 string    `json:"phone"`
	ID                    int64     `json:"id"`
}

type UpdateDriverDto struct {
	UpdateDriverRequest UpdateDriverRequest
	UserID              int64 `json:"user_id"`
}

type DriverResponse struct {
	ID                    int64      `json:"id"`
	UserID                int64      `json:"user_id"`
	Name                  string     `json:"name"`
	BirthDate             time.Time  `json:"birth_date"`
	Cpf                   string     `json:"cpf"`
	LicenseNumber         string     `json:"license_number"`
	LicenseCategory       string     `json:"license_category"`
	LicenseExpirationDate time.Time  `json:"license_expiration_date"`
	CEP                   string     `json:"cep"`
	State                 string     `json:"state"`
	City                  string     `json:"city"`
	Neighborhood          string     `json:"neighborhood"`
	Street                string     `json:"street"`
	StreetNumber          string     `json:"street_number"`
	Phone                 string     `json:"phone"`
	Complement            string     `json:"complement"`
	Status                bool       `json:"status"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             *time.Time `json:"updated_at"`
}

func (p *CreateDriverDto) ParseCreateToDriver() db.CreateDriverParams {
	arg := db.CreateDriverParams{
		UserID:                p.UserID,
		Name:                  p.CreateDriverRequest.Name,
		BirthDate:             p.CreateDriverRequest.BirthDate,
		Cpf:                   p.CreateDriverRequest.Cpf,
		LicenseNumber:         p.CreateDriverRequest.LicenseNumber,
		LicenseCategory:       p.CreateDriverRequest.LicenseCategory,
		LicenseExpirationDate: p.CreateDriverRequest.LicenseExpirationDate,
		State: sql.NullString{
			String: p.CreateDriverRequest.State,
			Valid:  true,
		},
		City: sql.NullString{
			String: p.CreateDriverRequest.City,
			Valid:  true,
		},
		Neighborhood: sql.NullString{
			String: p.CreateDriverRequest.Neighborhood,
			Valid:  true,
		},
		Street: sql.NullString{
			String: p.CreateDriverRequest.Street,
			Valid:  true,
		},
		StreetNumber: sql.NullString{
			String: p.CreateDriverRequest.StreetNumber,
			Valid:  true,
		},
		Phone: p.CreateDriverRequest.Phone,
		Cep:   p.CreateDriverRequest.CEP,
		Complement: sql.NullString{
			String: p.CreateDriverRequest.Complement,
			Valid:  true,
		},
	}
	return arg
}

func (p *UpdateDriverDto) ParseUpdateToDriver() db.UpdateDriverParams {
	arg := db.UpdateDriverParams{
		ID:                    p.UpdateDriverRequest.ID,
		UserID:                p.UserID,
		BirthDate:             p.UpdateDriverRequest.BirthDate,
		LicenseCategory:       p.UpdateDriverRequest.LicenseCategory,
		LicenseExpirationDate: p.UpdateDriverRequest.LicenseExpirationDate,
		State: sql.NullString{
			String: p.UpdateDriverRequest.State,
			Valid:  true,
		},
		City: sql.NullString{
			String: p.UpdateDriverRequest.City,
			Valid:  true,
		},
		Neighborhood: sql.NullString{
			String: p.UpdateDriverRequest.Neighborhood,
			Valid:  true,
		},
		Street: sql.NullString{
			String: p.UpdateDriverRequest.Street,
			Valid:  true,
		},
		StreetNumber: sql.NullString{
			String: p.UpdateDriverRequest.StreetNumber,
			Valid:  true,
		},
		Phone: p.UpdateDriverRequest.Phone,
		Cep:   p.UpdateDriverRequest.CEP,
		Complement: sql.NullString{
			String: p.UpdateDriverRequest.Complement,
			Valid:  true,
		},
	}
	return arg
}

func (p *DriverResponse) ParseFromDriverObject(result db.Driver) {
	p.ID = result.ID
	p.Name = result.Name
	p.UserID = result.UserID
	p.BirthDate = result.BirthDate
	p.Cpf = result.Cpf
	p.LicenseNumber = result.LicenseNumber
	p.LicenseCategory = result.LicenseCategory
	p.LicenseExpirationDate = result.LicenseExpirationDate
	p.CEP = result.Cep
	p.State = result.State.String
	p.City = result.City.String
	p.Neighborhood = result.Neighborhood.String
	p.Street = result.Street.String
	p.StreetNumber = result.StreetNumber.String
	p.Complement = result.Complement.String
	p.Phone = result.Phone
	p.Status = result.Status
	p.CreatedAt = result.CreatedAt
	if result.UpdatedAt.Valid {
		p.UpdatedAt = &result.UpdatedAt.Time
	}
}

func (p CreateDriverDto) ParseToCreateUserParams(driverId int64) db.CreateUserParams {
	defaultHash := "$2a$14$E4fX.uo.wKejvb2eq1o3m.IAAYbFxW5nF8fjPo1ESjhpv.eUeia8G"

	var sb strings.Builder

	for _, c := range p.CreateDriverRequest.Cpf {
		if unicode.IsDigit(c) {
			sb.WriteRune(c)
		}
	}
	return db.CreateUserParams{
		Name: p.CreateDriverRequest.Name,
		//Email: p.CreateDriverRequest.Email,
		Password: sql.NullString{
			String: defaultHash,
			Valid:  true,
		},
		Phone: sql.NullString{
			String: p.CreateDriverRequest.Phone,
			Valid:  true,
		},
		Document: sql.NullString{
			String: sb.String(),
			Valid:  true,
		},
		ProfileID: sql.NullInt64{
			Int64: 4,
			Valid: true,
		},
		DriverID: sql.NullInt64{
			Int64: driverId,
			Valid: true,
		},
		GoogleID: sql.NullString{
			String: "",
			Valid:  true,
		},
	}
}
