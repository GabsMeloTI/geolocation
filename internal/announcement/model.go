package announcement

import (
	"database/sql"
	db "geolocation/db/sqlc"
	"time"
)

type CreateAnnouncementRequest struct {
	UserID           int64     `json:"user_id"`
	Destination      string    `json:"destination"`
	Origin           string    `json:"origin"`
	DestinationLat   string    `json:"destination_lat"`
	DestinationLng   string    `json:"destination_lng"`
	OriginLat        string    `json:"origin_lat"`
	OriginLng        string    `json:"origin_lng"`
	Distance         int64     `json:"distance"`
	PickupDate       time.Time `json:"pickup_date"`
	DeliveryDate     time.Time `json:"delivery_date"`
	ExpirationDate   time.Time `json:"expiration_date"`
	Title            string    `json:"title"`
	CargoType        string    `json:"cargo_type"`
	CargoSpecies     string    `json:"cargo_species"`
	CargoVolume      string    `json:"cargo_volume"`
	CargoWeight      string    `json:"cargo_weight"`
	VehiclesAccepted string    `json:"vehicles_accepted"`
	Trailer          string    `json:"trailer"`
	RequiresTarp     bool      `json:"requires_tarp"`
	Tracking         bool      `json:"tracking"`
	Agency           bool      `json:"agency"`
	Description      string    `json:"description"`
	PaymentType      string    `json:"payment_type"`
	Advance          string    `json:"advance"`
	Toll             bool      `json:"toll"`
	Situation        string    `json:"situation"`
	CreatedWho       string    `json:"created_who"`
}

type UpdateAnnouncementRequest struct {
	UserID           int64          `json:"user_id"`
	Destination      string         `json:"destination"`
	Origin           string         `json:"origin"`
	DestinationLat   string         `json:"destination_lat"`
	DestinationLng   string         `json:"destination_lng"`
	OriginLat        string         `json:"origin_lat"`
	OriginLng        string         `json:"origin_lng"`
	Distance         int64          `json:"distance"`
	PickupDate       time.Time      `json:"pickup_date"`
	DeliveryDate     time.Time      `json:"delivery_date"`
	ExpirationDate   time.Time      `json:"expiration_date"`
	Title            string         `json:"title"`
	CargoType        string         `json:"cargo_type"`
	CargoSpecies     string         `json:"cargo_species"`
	CargoVolume      string         `json:"cargo_volume"`
	CargoWeight      string         `json:"cargo_weight"`
	VehiclesAccepted string         `json:"vehicles_accepted"`
	Trailer          string         `json:"trailer"`
	RequiresTarp     bool           `json:"requires_tarp"`
	Tracking         bool           `json:"tracking"`
	Agency           bool           `json:"agency"`
	Description      string         `json:"description"`
	PaymentType      string         `json:"payment_type"`
	Advance          string         `json:"advance"`
	Toll             bool           `json:"toll"`
	Situation        string         `json:"situation"`
	UpdatedWho       sql.NullString `json:"updated_who"`
	ID               int64          `json:"id"`
}

type DeleteAnnouncementRequest struct {
	ID         int64          `json:"id"`
	UpdatedWho sql.NullString `json:"updated_who"`
}

type AnnouncementResponse struct {
	ID               int64      `json:"id"`
	UserID           int64      `json:"user_id"`
	Destination      string     `json:"destination"`
	Origin           string     `json:"origin"`
	DestinationLat   string     `json:"destination_lat"`
	DestinationLng   string     `json:"destination_lng"`
	OriginLat        string     `json:"origin_lat"`
	OriginLng        string     `json:"origin_lng"`
	Distance         int64      `json:"distance"`
	PickupDate       time.Time  `json:"pickup_date"`
	DeliveryDate     time.Time  `json:"delivery_date"`
	ExpirationDate   time.Time  `json:"expiration_date"`
	Title            string     `json:"title"`
	CargoType        string     `json:"cargo_type"`
	CargoSpecies     string     `json:"cargo_species"`
	CargoVolume      string     `json:"cargo_volume"`
	CargoWeight      string     `json:"cargo_weight"`
	VehiclesAccepted string     `json:"vehicles_accepted"`
	Trailer          string     `json:"trailer"`
	RequiresTarp     bool       `json:"requires_tarp"`
	Tracking         bool       `json:"tracking"`
	Agency           bool       `json:"agency"`
	Description      string     `json:"description"`
	PaymentType      string     `json:"payment_type"`
	Advance          string     `json:"advance"`
	Toll             bool       `json:"toll"`
	Situation        string     `json:"situation"`
	Status           bool       `json:"status"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        *time.Time `json:"updated_at"`
	CreatedWho       string     `json:"created_who"`
	UpdatedWho       string     `json:"updated_who"`
}

func (p *CreateAnnouncementRequest) ParseCreateToAnnouncement() db.CreateAnnouncementParams {
	arg := db.CreateAnnouncementParams{
		UserID:           p.UserID,
		Destination:      p.Destination,
		Origin:           p.Origin,
		DestinationLat:   p.DestinationLat,
		DestinationLng:   p.DestinationLng,
		OriginLat:        p.OriginLat,
		OriginLng:        p.OriginLng,
		Distance:         p.Distance,
		PickupDate:       p.PickupDate,
		DeliveryDate:     p.DeliveryDate,
		ExpirationDate:   p.ExpirationDate,
		Title:            p.Title,
		CargoType:        p.CargoType,
		CargoSpecies:     p.CargoSpecies,
		CargoVolume:      p.CargoVolume,
		CargoWeight:      p.CargoWeight,
		VehiclesAccepted: p.VehiclesAccepted,
		Trailer:          p.Trailer,
		RequiresTarp:     p.RequiresTarp,
		Tracking:         p.Tracking,
		Agency:           false,
		Description:      p.Description,
		PaymentType:      p.PaymentType,
		Advance:          p.Advance,
		Toll:             p.Toll,
		Situation:        p.Situation,
		CreatedWho:       p.CreatedWho,
	}
	return arg
}

func (p *UpdateAnnouncementRequest) ParseUpdateToAnnouncement() db.UpdateAnnouncementParams {
	arg := db.UpdateAnnouncementParams{
		ID:               p.ID,
		UserID:           p.UserID,
		Destination:      p.Destination,
		Origin:           p.Origin,
		DestinationLat:   p.DestinationLat,
		DestinationLng:   p.DestinationLng,
		OriginLat:        p.OriginLat,
		OriginLng:        p.OriginLng,
		Distance:         p.Distance,
		PickupDate:       p.PickupDate,
		DeliveryDate:     p.DeliveryDate,
		ExpirationDate:   p.ExpirationDate,
		Title:            p.Title,
		CargoType:        p.CargoType,
		CargoSpecies:     p.CargoSpecies,
		CargoVolume:      p.CargoVolume,
		CargoWeight:      p.CargoWeight,
		VehiclesAccepted: p.VehiclesAccepted,
		Trailer:          p.Trailer,
		RequiresTarp:     p.RequiresTarp,
		Tracking:         p.Tracking,
		Agency:           false,
		Description:      p.Description,
		PaymentType:      p.PaymentType,
		Advance:          p.Advance,
		Toll:             p.Toll,
		Situation:        p.Situation,
		UpdatedWho:       p.UpdatedWho,
	}
	return arg
}

func (p *DeleteAnnouncementRequest) ParseDeleteToAnnouncement() db.DeleteAnnouncementParams {
	arg := db.DeleteAnnouncementParams{
		ID:         p.ID,
		UpdatedWho: p.UpdatedWho,
	}
	return arg
}

func (p *AnnouncementResponse) ParseFromAnnouncementObject(result db.Announcement) {
	p.ID = result.ID
	p.Destination = result.Destination
	p.Origin = result.Origin
	p.DestinationLat = result.DestinationLat
	p.DestinationLng = result.DestinationLng
	p.OriginLat = result.OriginLat
	p.OriginLng = result.OriginLng
	p.Distance = result.Distance
	p.DeliveryDate = result.DeliveryDate
	p.PickupDate = result.PickupDate
	p.ExpirationDate = result.ExpirationDate
	p.Title = result.Title
	p.CargoType = result.CargoType
	p.CargoWeight = result.CargoWeight
	p.CargoSpecies = result.CargoSpecies
	p.CargoVolume = result.CargoVolume
	p.VehiclesAccepted = result.VehiclesAccepted
	p.Trailer = result.Trailer
	p.RequiresTarp = result.RequiresTarp
	p.Tracking = result.Tracking
	p.Agency = result.Agency
	p.Description = result.Description
	p.PaymentType = result.PaymentType
	p.Advance = result.Advance
	p.Toll = result.Toll
	p.Situation = result.Situation
	p.Status = result.Status
	p.CreatedAt = result.CreatedAt
	if result.UpdatedAt.Valid {
		p.UpdatedAt = &result.UpdatedAt.Time
	}
	p.CreatedWho = result.CreatedWho
	p.UpdatedWho = result.UpdatedWho.String
}
