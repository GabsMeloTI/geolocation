package location

import (
	"database/sql"
	db "geolocation/db/sqlc"
	"geolocation/internal/get_token"
	"github.com/google/uuid"
	"time"
)

type CreateLocationRequest struct {
	Type    string              `json:"type"`
	Address string              `json:"address"`
	Area    []CreateAreaRequest `json:"area"`
}

type CreateLocationDTO struct {
	CreateLocationRequest CreateLocationRequest
	Payload               get_token.PayloadDTO
}

type CreateAreaRequest struct {
	LocationsID int64  `json:"locations_id"`
	Latitude    string `json:"latitude"`
	Longitude   string `json:"longitude"`
	Description string `json:"description"`
}

type DeleteLocationRequest struct {
	ID       int64     `json:"id"`
	AccessID int64     `json:"access_id"`
	TenantID uuid.UUID `json:"tenant_id"`
}

type GetAreasResponse struct {
	ID          int64  `json:"id"`
	Latitude    string `json:"latitude"`
	Longitude   string `json:"longitude"`
	Description string `json:"description"`
}

type LocationResponse struct {
	ID        int64              `json:"id"`
	Type      string             `json:"type"`
	Address   string             `json:"address"`
	Area      []GetAreasResponse `json:"area"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
}

type UpdateAreaRequest struct {
	ID          int64  `json:"id"`
	LocationsID int64  `json:"locations_id"`
	Latitude    string `json:"latitude"`
	Longitude   string `json:"longitude"`
	Description string `json:"description"`
}

type UpdateLocationRequest struct {
	ID      int64               `json:"id"`
	Type    string              `json:"type"`
	Address string              `json:"address"`
	Areas   []UpdateAreaRequest `json:"area"`
}

type UpdateLocationDTO struct {
	UpdateLocationRequest UpdateLocationRequest
	Payload               get_token.PayloadDTO
}

func (p *CreateLocationRequest) ParseCreateToLocation() db.CreateLocationParams {
	return db.CreateLocationParams{
		Type: p.Type,
		Address: sql.NullString{
			String: p.Address,
			Valid:  true,
		},
	}
}

func (p *CreateAreaRequest) ParseCreateToArea() db.CreateAreaParams {
	return db.CreateAreaParams{
		LocationsID: p.LocationsID,
		Latitude:    p.Latitude,
		Longitude:   p.Longitude,
		Description: p.Description,
	}
}

func (p *UpdateLocationDTO) ParseUpdateToLocation() db.UpdateLocationParams {
	return db.UpdateLocationParams{
		Type:     p.UpdateLocationRequest.Type,
		Address:  sql.NullString{String: p.UpdateLocationRequest.Address, Valid: true},
		ID:       p.UpdateLocationRequest.ID,
		AccessID: p.Payload.AccessID,
		TenantID: p.Payload.TenantID,
	}
}

func (p *LocationResponse) ParseFromPlansObject(result db.Location, areas []GetAreasResponse) {
	p.ID = result.ID
	p.Type = result.Type
	p.Address = result.Address.String
	p.Area = areas
	p.CreatedAt = result.CreatedAt
	p.UpdatedAt = result.UpdatedAt.Time
}
