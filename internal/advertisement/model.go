package advertisement

import (
	"database/sql"
	db "geolocation/db/sqlc"
	"time"
)

type CreateAdvertisementRequest struct {
	UserID           int64     `json:"user_id"`
	Destination      string    `json:"destination"`
	Origin           string    `json:"origin"`
	DestinationLat   float64   `json:"destination_lat"`
	DestinationLng   float64   `json:"destination_lng"`
	OriginLat        float64   `json:"origin_lat"`
	OriginLng        float64   `json:"origin_lng"`
	Distance         int64     `json:"distance"`
	PickupDate       time.Time `json:"pickup_date"`
	DeliveryDate     time.Time `json:"delivery_date"`
	ExpirationDate   time.Time `json:"expiration_date"`
	Title            string    `json:"title"`
	CargoType        string    `json:"cargo_type"`
	CargoSpecies     string    `json:"cargo_species"`
	CargoWeight      float64   `json:"cargo_weight"`
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
	Price            float64   `json:"price"`
	CreatedWho       string    `json:"created_who"`
	State            string    `json:"state"`
	City             string    `json:"city"`
	Complement       string    `json:"complement"`
	Neighborhood     string    `json:"neighborhood"`
	Street           string    `json:"street"`
	StreetNumber     string    `json:"street_number"`
	CEP              string    `json:"cep"`
}

type CreateAdvertisementDto struct {
	CreateAdvertisementRequest CreateAdvertisementRequest
	UserID                     int64  `json:"user_id"`
	CreatedWho                 string `json:"created_who"`
}

type UpdateAdvertisementRequest struct {
	Destination      string    `json:"destination"`
	Origin           string    `json:"origin"`
	DestinationLat   float64   `json:"destination_lat"`
	DestinationLng   float64   `json:"destination_lng"`
	OriginLat        float64   `json:"origin_lat"`
	OriginLng        float64   `json:"origin_lng"`
	Distance         int64     `json:"distance"`
	PickupDate       time.Time `json:"pickup_date"`
	DeliveryDate     time.Time `json:"delivery_date"`
	ExpirationDate   time.Time `json:"expiration_date"`
	Title            string    `json:"title"`
	CargoType        string    `json:"cargo_type"`
	CargoSpecies     string    `json:"cargo_species"`
	CargoWeight      float64   `json:"cargo_weight"`
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
	Price            float64   `json:"price"`
	ID               int64     `json:"id"`
	State            string    `json:"state"`
	City             string    `json:"city"`
	Complement       string    `json:"complement"`
	Neighborhood     string    `json:"neighborhood"`
	Street           string    `json:"street"`
	StreetNumber     string    `json:"street_number"`
	CEP              string    `json:"cep"`
}

type UpdateAdvertisementDto struct {
	UpdateAdvertisementRequest UpdateAdvertisementRequest
	UserID                     int64          `json:"user_id"`
	UpdatedWho                 sql.NullString `json:"updated_who"`
}

type DeleteAdvertisementRequest struct {
	ID         int64          `json:"id"`
	UserID     int64          `json:"user_id"`
	UpdatedWho sql.NullString `json:"updated_who"`
}

type AdvertisementResponse struct {
	ID               int64      `json:"id"`
	UserID           int64      `json:"user_id"`
	Destination      string     `json:"destination"`
	Origin           string     `json:"origin"`
	DestinationLat   float64    `json:"destination_lat"`
	DestinationLng   float64    `json:"destination_lng"`
	OriginLat        float64    `json:"origin_lat"`
	OriginLng        float64    `json:"origin_lng"`
	Distance         int64      `json:"distance"`
	PickupDate       time.Time  `json:"pickup_date"`
	DeliveryDate     time.Time  `json:"delivery_date"`
	ExpirationDate   time.Time  `json:"expiration_date"`
	Title            string     `json:"title"`
	CargoType        string     `json:"cargo_type"`
	CargoSpecies     string     `json:"cargo_species"`
	CargoWeight      float64    `json:"cargo_weight"`
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
	Price            float64    `json:"price"`
	State            string     `json:"state"`
	City             string     `json:"city"`
	Complement       string     `json:"complement"`
	Neighborhood     string     `json:"neighborhood"`
	Street           string     `json:"street"`
	StreetNumber     string     `json:"street_number"`
	CEP              string     `json:"cep"`
	Status           bool       `json:"status"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        *time.Time `json:"updated_at"`
	CreatedWho       string     `json:"created_who"`
	UpdatedWho       *string    `json:"updated_who"`
}

type AdvertisementResponseAll struct {
	ID               int64      `json:"id"`
	UserID           int64      `json:"user_id"`
	UserName         string     `json:"user_name"`
	ActiveThere      time.Time  `json:"active_there"`
	ActiveDuration   string     `json:"active_duration"`
	UserCity         string     `json:"user_city"`
	UserState        string     `json:"user_state"`
	UserPhone        string     `json:"user_phone"`
	UserEmail        string     `json:"user_email"`
	Destination      string     `json:"destination"`
	Origin           string     `json:"origin"`
	DestinationLat   float64    `json:"destination_lat"`
	DestinationLng   float64    `json:"destination_lng"`
	OriginLat        float64    `json:"origin_lat"`
	OriginLng        float64    `json:"origin_lng"`
	Distance         int64      `json:"distance"`
	PickupDate       time.Time  `json:"pickup_date"`
	DeliveryDate     time.Time  `json:"delivery_date"`
	ExpirationDate   time.Time  `json:"expiration_date"`
	Title            string     `json:"title"`
	CargoType        string     `json:"cargo_type"`
	CargoSpecies     string     `json:"cargo_species"`
	CargoWeight      float64    `json:"cargo_weight"`
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
	ActiveFreight    int64      `json:"active_freight"`
	Price            float64    `json:"price"`
	State            string     `json:"state"`
	City             string     `json:"city"`
	Complement       string     `json:"complement"`
	Neighborhood     string     `json:"neighborhood"`
	Street           string     `json:"street"`
	StreetNumber     string     `json:"street_number"`
	CEP              string     `json:"cep"`
	CreatedAt        time.Time  `json:"created_at"`
	CreatedWho       string     `json:"created_who"`
	UpdatedAt        *time.Time `json:"updated_at,omitempty"`
	UpdatedWho       *string    `json:"updated_who,omitempty"`
}

type AdvertisementResponseNoUser struct {
	ID               int64     `json:"id"`
	Destination      string    `json:"destination"`
	Origin           string    `json:"origin"`
	PickupDate       time.Time `json:"pickup_date"`
	DeliveryDate     time.Time `json:"delivery_date"`
	ExpirationDate   time.Time `json:"expiration_date"`
	Title            string    `json:"title"`
	CargoType        string    `json:"cargo_type"`
	CargoSpecies     string    `json:"cargo_species"`
	CargoWeight      float64   `json:"cargo_weight"`
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
	State            string    `json:"state"`
	City             string    `json:"city"`
	Complement       string    `json:"complement"`
	Neighborhood     string    `json:"neighborhood"`
	Street           string    `json:"street"`
	StreetNumber     string    `json:"street_number"`
	CEP              string    `json:"cep"`
	CreatedAt        time.Time `json:"created_at"`
}

func (p *CreateAdvertisementDto) ParseCreateToAdvertisement() db.CreateAdvertisementParams {
	arg := db.CreateAdvertisementParams{
		UserID:           p.UserID,
		Destination:      p.CreateAdvertisementRequest.Destination,
		Origin:           p.CreateAdvertisementRequest.Origin,
		DestinationLat:   p.CreateAdvertisementRequest.DestinationLat,
		DestinationLng:   p.CreateAdvertisementRequest.DestinationLng,
		OriginLat:        p.CreateAdvertisementRequest.OriginLat,
		OriginLng:        p.CreateAdvertisementRequest.OriginLng,
		Distance:         p.CreateAdvertisementRequest.Distance,
		PickupDate:       p.CreateAdvertisementRequest.PickupDate,
		DeliveryDate:     p.CreateAdvertisementRequest.DeliveryDate,
		ExpirationDate:   p.CreateAdvertisementRequest.ExpirationDate,
		Title:            p.CreateAdvertisementRequest.Title,
		CargoType:        p.CreateAdvertisementRequest.CargoType,
		CargoSpecies:     p.CreateAdvertisementRequest.CargoSpecies,
		CargoWeight:      p.CreateAdvertisementRequest.CargoWeight,
		VehiclesAccepted: p.CreateAdvertisementRequest.VehiclesAccepted,
		Trailer:          p.CreateAdvertisementRequest.Trailer,
		RequiresTarp:     p.CreateAdvertisementRequest.RequiresTarp,
		Tracking:         p.CreateAdvertisementRequest.Tracking,
		Agency:           p.CreateAdvertisementRequest.Agency,
		Description:      p.CreateAdvertisementRequest.Description,
		PaymentType:      p.CreateAdvertisementRequest.PaymentType,
		Advance:          p.CreateAdvertisementRequest.Advance,
		Toll:             p.CreateAdvertisementRequest.Toll,
		Situation:        p.CreateAdvertisementRequest.Situation,
		Price:            p.CreateAdvertisementRequest.Price,
		State:            p.CreateAdvertisementRequest.State,
		City:             p.CreateAdvertisementRequest.City,
		Complement:       p.CreateAdvertisementRequest.Complement,
		Neighborhood:     p.CreateAdvertisementRequest.Neighborhood,
		Street:           p.CreateAdvertisementRequest.Street,
		StreetNumber:     p.CreateAdvertisementRequest.StreetNumber,
		Cep:              p.CreateAdvertisementRequest.CEP,
		CreatedWho:       p.CreatedWho,
	}
	return arg
}

func (p *UpdateAdvertisementDto) ParseUpdateToAdvertisement() db.UpdateAdvertisementParams {
	arg := db.UpdateAdvertisementParams{
		ID:               p.UpdateAdvertisementRequest.ID,
		UserID:           p.UserID,
		Destination:      p.UpdateAdvertisementRequest.Destination,
		Origin:           p.UpdateAdvertisementRequest.Origin,
		DestinationLat:   p.UpdateAdvertisementRequest.DestinationLat,
		DestinationLng:   p.UpdateAdvertisementRequest.DestinationLng,
		OriginLat:        p.UpdateAdvertisementRequest.OriginLat,
		OriginLng:        p.UpdateAdvertisementRequest.OriginLng,
		Distance:         p.UpdateAdvertisementRequest.Distance,
		PickupDate:       p.UpdateAdvertisementRequest.PickupDate,
		DeliveryDate:     p.UpdateAdvertisementRequest.DeliveryDate,
		ExpirationDate:   p.UpdateAdvertisementRequest.ExpirationDate,
		Title:            p.UpdateAdvertisementRequest.Title,
		CargoType:        p.UpdateAdvertisementRequest.CargoType,
		CargoSpecies:     p.UpdateAdvertisementRequest.CargoSpecies,
		CargoWeight:      p.UpdateAdvertisementRequest.CargoWeight,
		VehiclesAccepted: p.UpdateAdvertisementRequest.VehiclesAccepted,
		Trailer:          p.UpdateAdvertisementRequest.Trailer,
		RequiresTarp:     p.UpdateAdvertisementRequest.RequiresTarp,
		Tracking:         p.UpdateAdvertisementRequest.Tracking,
		Agency:           p.UpdateAdvertisementRequest.Agency,
		Description:      p.UpdateAdvertisementRequest.Description,
		PaymentType:      p.UpdateAdvertisementRequest.PaymentType,
		Advance:          p.UpdateAdvertisementRequest.Advance,
		Toll:             p.UpdateAdvertisementRequest.Toll,
		Situation:        p.UpdateAdvertisementRequest.Situation,
		Price:            p.UpdateAdvertisementRequest.Price,
		State:            p.UpdateAdvertisementRequest.State,
		City:             p.UpdateAdvertisementRequest.City,
		Complement:       p.UpdateAdvertisementRequest.Complement,
		Neighborhood:     p.UpdateAdvertisementRequest.Neighborhood,
		Street:           p.UpdateAdvertisementRequest.Street,
		StreetNumber:     p.UpdateAdvertisementRequest.StreetNumber,
		Cep:              p.UpdateAdvertisementRequest.CEP,
		UpdatedWho:       p.UpdatedWho,
	}
	return arg
}

func (p *DeleteAdvertisementRequest) ParseDeleteToAdvertisement() db.DeleteAdvertisementParams {
	arg := db.DeleteAdvertisementParams{
		ID:         p.ID,
		UserID:     p.UserID,
		UpdatedWho: p.UpdatedWho,
	}
	return arg
}

func (p *AdvertisementResponse) ParseFromAdvertisementObject(result db.Advertisement) {
	p.ID = result.ID
	p.UserID = result.UserID
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
	p.Price = result.Price
	p.State = result.State
	p.City = result.City
	p.Complement = result.Complement
	p.Neighborhood = result.Neighborhood
	p.Street = result.Street
	p.StreetNumber = result.StreetNumber
	p.CEP = result.Cep
	p.Status = result.Status
	p.CreatedAt = result.CreatedAt
	if result.UpdatedAt.Valid {
		p.UpdatedAt = &result.UpdatedAt.Time
	}
	p.CreatedWho = result.CreatedWho
	if result.UpdatedWho.Valid {
		p.UpdatedWho = &result.UpdatedWho.String
	}
}

func (p *AdvertisementResponseAll) ParseFromAdvertisementObject(result db.GetAllAdvertisementUsersRow) {
	p.ID = result.ID
	p.UserID = result.UserID
	p.UserName = result.UserName
	p.ActiveThere = result.ActiveThere.Time
	p.UserCity = result.UserCity.String
	p.UserState = result.UserState.String
	p.UserPhone = result.UserPhone.String
	p.UserEmail = result.UserEmail
	p.Destination = result.Destination
	p.Origin = result.Origin
	p.DestinationLat = result.DestinationLat
	p.DestinationLng = result.DestinationLng
	p.OriginLat = result.OriginLat
	p.OriginLng = result.OriginLng
	p.Distance = result.Distance
	p.PickupDate = result.PickupDate
	p.DeliveryDate = result.DeliveryDate
	p.ExpirationDate = result.ExpirationDate
	p.Title = result.Title
	p.CargoType = result.CargoType
	p.CargoSpecies = result.CargoSpecies
	p.CargoWeight = result.CargoWeight
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
	p.Price = result.Price
	p.State = result.State
	p.City = result.City
	p.Complement = result.Complement
	p.Neighborhood = result.Neighborhood
	p.Street = result.Street
	p.StreetNumber = result.StreetNumber
	p.CEP = result.Cep
	p.CreatedAt = result.CreatedAt
	p.CreatedWho = result.CreatedWho
	if result.UpdatedAt.Valid {
		p.UpdatedAt = &result.UpdatedAt.Time
	}
	if result.UpdatedWho.Valid {
		p.UpdatedWho = &result.UpdatedWho.String
	}
}

func (p *AdvertisementResponseNoUser) ParseFromAdvertisementObject(result db.GetAllAdvertisementPublicRow) {
	p.ID = result.ID
	p.Destination = result.Destination
	p.Origin = result.Origin
	p.PickupDate = result.PickupDate
	p.DeliveryDate = result.DeliveryDate
	p.ExpirationDate = result.ExpirationDate
	p.Title = result.Title
	p.CargoType = result.CargoType
	p.CargoSpecies = result.CargoSpecies
	p.CargoWeight = result.CargoWeight
	p.VehiclesAccepted = result.VehiclesAccepted
	p.Trailer = result.Trailer
	p.RequiresTarp = result.RequiresTarp
	p.Tracking = result.Tracking
	p.Agency = result.Agency
	p.Description = result.Description
	p.PaymentType = result.PaymentType
	p.Advance = result.Advance
	p.Toll = result.Toll
	p.State = result.State
	p.City = result.City
	p.Complement = result.Complement
	p.Neighborhood = result.Neighborhood
	p.Street = result.Street
	p.StreetNumber = result.StreetNumber
	p.CEP = result.Cep
	p.Situation = result.Situation
	p.CreatedAt = result.CreatedAt
}
