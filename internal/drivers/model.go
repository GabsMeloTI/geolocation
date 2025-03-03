package drivers

import (
	"database/sql"
	db "geolocation/db/sqlc"
	"time"
)

type CreateDriverRequest struct {
	UserID                int64     `json:"user_id"`
	Name                  string    `json:"name"`
	BirthDate             time.Time `json:"birth_date"`
	Cpf                   string    `json:"cpf"`
	LicenseNumber         string    `json:"license_number"`
	LicenseCategory       string    `json:"license_category" validate:"oneof=a b c d e"`
	LicenseExpirationDate time.Time `json:"license_expiration_date"`
	CEP                   string    `json:"cep"`
	State                 string    `json:"state"`
	City                  string    `json:"city"`
	Neighborhood          string    `json:"neighborhood"`
	Street                string    `json:"street"`
	StreetNumber          string    `json:"street_number"`
	Complement            string    `json:"complement"`
	Phone                 string    `json:"phone"`
}

type UpdateDriverRequest struct {
	UserID                int64     `json:"user_id"`
	Name                  string    `json:"name"`
	BirthDate             time.Time `json:"birth_date"`
	Cpf                   string    `json:"cpf"`
	LicenseNumber         string    `json:"license_number"`
	LicenseCategory       string    `json:"license_category" validate:"oneof=a b c d e"`
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

func (p *CreateDriverRequest) ParseCreateToDriver() db.CreateDriverParams {
	arg := db.CreateDriverParams{
		UserID:                p.UserID,
		Name:                  p.Name,
		BirthDate:             p.BirthDate,
		Cpf:                   p.Cpf,
		LicenseNumber:         p.LicenseNumber,
		LicenseCategory:       p.LicenseCategory,
		LicenseExpirationDate: p.LicenseExpirationDate,
		State: sql.NullString{
			String: p.State,
			Valid:  true,
		},
		City: sql.NullString{
			String: p.City,
			Valid:  true,
		},
		Neighborhood: sql.NullString{
			String: p.Neighborhood,
			Valid:  true,
		},
		Street: sql.NullString{
			String: p.Street,
			Valid:  true,
		},
		StreetNumber: sql.NullString{
			String: p.StreetNumber,
			Valid:  true,
		},
		Phone: p.Phone,
		Cep:   p.CEP,
		Complement: sql.NullString{
			String: p.Complement,
			Valid:  true,
		},
	}
	return arg
}

func (p *UpdateDriverRequest) ParseUpdateToDriver() db.UpdateDriverParams {
	arg := db.UpdateDriverParams{
		ID:                    p.ID,
		UserID:                p.UserID,
		BirthDate:             p.BirthDate,
		LicenseCategory:       p.LicenseCategory,
		LicenseExpirationDate: p.LicenseExpirationDate,
		State: sql.NullString{
			String: p.State,
			Valid:  true,
		},
		City: sql.NullString{
			String: p.City,
			Valid:  true,
		},
		Neighborhood: sql.NullString{
			String: p.Neighborhood,
			Valid:  true,
		},
		Street: sql.NullString{
			String: p.Street,
			Valid:  true,
		},
		StreetNumber: sql.NullString{
			String: p.StreetNumber,
			Valid:  true,
		},
		Phone: p.Phone,
		Cep:   p.CEP,
		Complement: sql.NullString{
			String: p.Complement,
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
