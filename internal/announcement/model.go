package announcement

import (
	db "geolocation/db/sqlc"
	"time"
)

type CreateAnnouncementRequest struct {
	Destination        string    `json:"destination"`
	Origin             string    `json:"origin"`
	DestinationLat     string    `json:"destination_lat"`
	DestinationLng     string    `json:"destination_lng"`
	OriginLat          string    `json:"origin_lat"`
	OriginLng          string    `json:"origin_lng"`
	Description        string    `json:"description"`
	CargoDescription   string    `json:"cargo_description"`
	PaymentDescription string    `json:"payment_description"`
	DeliveryDate       time.Time `json:"delivery_date"`
	PickupDate         time.Time `json:"pickup_date"`
	DeadlineDate       time.Time `json:"deadline_date"`
	Price              string    `json:"price"`
	Vehicle            string    `json:"vehicle"`
	BodyType           string    `json:"body_type"`
	Kilometers         string    `json:"kilometers"`
	CargoNature        string    `json:"cargo_nature"`
	CargoType          string    `json:"cargo_type"`
	CargoWeight        string    `json:"cargo_weight"`
	Tracking           bool      `json:"tracking"`
	RequiresTarp       bool      `json:"requires_tarp"`
}

type UpdateAnnouncementRequest struct {
	ID                 int64     `json:"id"`
	Destination        string    `json:"destination"`
	Origin             string    `json:"origin"`
	DestinationLat     string    `json:"destination_lat"`
	DestinationLng     string    `json:"destination_lng"`
	OriginLat          string    `json:"origin_lat"`
	OriginLng          string    `json:"origin_lng"`
	Description        string    `json:"description"`
	CargoDescription   string    `json:"cargo_description"`
	PaymentDescription string    `json:"payment_description"`
	DeliveryDate       time.Time `json:"delivery_date"`
	PickupDate         time.Time `json:"pickup_date"`
	DeadlineDate       time.Time `json:"deadline_date"`
	Price              string    `json:"price"`
	Vehicle            string    `json:"vehicle"`
	BodyType           string    `json:"body_type"`
	Kilometers         string    `json:"kilometers"`
	CargoNature        string    `json:"cargo_nature"`
	CargoType          string    `json:"cargo_type"`
	CargoWeight        string    `json:"cargo_weight"`
	Tracking           bool      `json:"tracking"`
	RequiresTarp       bool      `json:"requires_tarp"`
}

type AnnouncementResponse struct {
	ID                 int64      `json:"id"`
	Destination        string     `json:"destination"`
	Origin             string     `json:"origin"`
	DestinationLat     string     `json:"destination_lat"`
	DestinationLng     string     `json:"destination_lng"`
	OriginLat          string     `json:"origin_lat"`
	OriginLng          string     `json:"origin_lng"`
	Description        string     `json:"description"`
	CargoDescription   string     `json:"cargo_description"`
	PaymentDescription string     `json:"payment_description"`
	DeliveryDate       time.Time  `json:"delivery_date"`
	PickupDate         time.Time  `json:"pickup_date"`
	DeadlineDate       time.Time  `json:"deadline_date"`
	Price              string     `json:"price"`
	Vehicle            string     `json:"vehicle"`
	BodyType           string     `json:"body_type"`
	Kilometers         string     `json:"kilometers"`
	CargoNature        string     `json:"cargo_nature"`
	CargoType          string     `json:"cargo_type"`
	CargoWeight        string     `json:"cargo_weight"`
	Tracking           bool       `json:"tracking"`
	RequiresTarp       bool       `json:"requires_tarp"`
	Status             bool       `json:"status"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          *time.Time `json:"updated_at"`
}

func (p *CreateAnnouncementRequest) ParseCreateToAnnouncement() db.CreateAnnouncementParams {
	arg := db.CreateAnnouncementParams{
		Destination:        p.Destination,
		Origin:             p.Origin,
		DestinationLat:     p.DestinationLat,
		DestinationLng:     p.DestinationLng,
		OriginLat:          p.OriginLat,
		OriginLng:          p.OriginLng,
		Description:        p.Description,
		CargoDescription:   p.CargoDescription,
		PaymentDescription: p.PaymentDescription,
		DeliveryDate:       p.DeliveryDate,
		PickupDate:         p.PickupDate,
		DeadlineDate:       p.DeadlineDate,
		Price:              p.Price,
		Vehicle:            p.Vehicle,
		BodyType:           p.BodyType,
		Kilometers:         p.Kilometers,
		CargoNature:        p.CargoNature,
		CargoType:          p.CargoType,
		CargoWeight:        p.CargoWeight,
		Tracking:           p.Tracking,
		RequiresTarp:       p.RequiresTarp,
	}
	return arg
}

func (p *UpdateAnnouncementRequest) ParseUpdateToAnnouncement() db.UpdateAnnouncementParams {
	arg := db.UpdateAnnouncementParams{
		ID:                 p.ID,
		Destination:        p.Destination,
		Origin:             p.Origin,
		DestinationLat:     p.DestinationLat,
		DestinationLng:     p.DestinationLng,
		OriginLat:          p.OriginLat,
		OriginLng:          p.OriginLng,
		Description:        p.Description,
		CargoDescription:   p.CargoDescription,
		PaymentDescription: p.PaymentDescription,
		DeliveryDate:       p.DeliveryDate,
		PickupDate:         p.PickupDate,
		DeadlineDate:       p.DeadlineDate,
		Price:              p.Price,
		Vehicle:            p.Vehicle,
		BodyType:           p.BodyType,
		Kilometers:         p.Kilometers,
		CargoNature:        p.CargoNature,
		CargoType:          p.CargoType,
		CargoWeight:        p.CargoWeight,
		Tracking:           p.Tracking,
		RequiresTarp:       p.RequiresTarp,
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
	p.Description = result.Description
	p.CargoDescription = result.CargoDescription
	p.PaymentDescription = result.PaymentDescription
	p.DeliveryDate = result.DeliveryDate
	p.PickupDate = result.PickupDate
	p.DeadlineDate = result.DeadlineDate
	p.Price = result.Price
	p.Vehicle = result.Vehicle
	p.BodyType = result.BodyType
	p.Kilometers = result.Kilometers
	p.CargoNature = result.CargoNature
	p.CargoType = result.CargoType
	p.CargoWeight = result.CargoWeight
	p.Tracking = result.Tracking
	p.RequiresTarp = result.RequiresTarp
	p.Status = result.Status
	p.CreatedAt = result.CreatedAt
	if result.UpdatedAt.Valid {
		p.UpdatedAt = &result.UpdatedAt.Time
	}
}
