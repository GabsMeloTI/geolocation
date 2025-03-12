package advertisement

import (
	"database/sql"
	"time"

	db "geolocation/db/sqlc"
	"geolocation/internal/new_routes"
)

type CreateAdvertisementRequest struct {
	Destination             string    `json:"destination"`
	Origin                  string    `json:"origin"`
	Distance                int64     `json:"distance"`
	PickupDate              time.Time `json:"pickup_date"`
	DeliveryDate            time.Time `json:"delivery_date"`
	ExpirationDate          time.Time `json:"expiration_date"`
	Title                   string    `json:"title"`
	CargoType               string    `json:"cargo_type"`
	CargoSpecies            string    `json:"cargo_species"`
	CargoWeight             float64   `json:"cargo_weight"`
	VehiclesAccepted        string    `json:"vehicles_accepted"`
	Trailer                 string    `json:"trailer"`
	RequiresTarp            bool      `json:"requires_tarp"`
	Tracking                bool      `json:"tracking"`
	Agency                  bool      `json:"agency"`
	Description             string    `json:"description"`
	PaymentType             string    `json:"payment_type"`
	Advance                 string    `json:"advance"`
	Toll                    bool      `json:"toll"`
	Price                   float64   `json:"price"`
	CreatedWho              string    `json:"created_who"`
	StateOrigin             string    `json:"state_origin"`
	CityOrigin              string    `json:"city_origin"`
	ComplementOrigin        string    `json:"complement_origin"`
	NeighborhoodOrigin      string    `json:"neighborhood_origin"`
	StreetOrigin            string    `json:"street_origin"`
	StreetNumberOrigin      string    `json:"street_number_origin"`
	CEPOrigin               string    `json:"cep_origin"`
	StateDestination        string    `json:"state_destination"`
	CityDestination         string    `json:"city_destination"`
	ComplementDestination   string    `json:"complement_destination"`
	NeighborhoodDestination string    `json:"neighborhood_destination"`
	StreetDestination       string    `json:"street_destination"`
	StreetNumberDestination string    `json:"street_number_destination"`
	CEPDestination          string    `json:"cep_destination"`
}

type CreateAdvertisementDto struct {
	CreateAdvertisementRequest CreateAdvertisementRequest
	UserID                     int64  `json:"user_id"`
	CreatedWho                 string `json:"created_who"`
}

type UpdateAdvertisementRequest struct {
	ID                      int64     `json:"id"`
	Destination             string    `json:"destination"`
	Origin                  string    `json:"origin"`
	DestinationLat          float64   `json:"destination_lat"`
	DestinationLng          float64   `json:"destination_lng"`
	OriginLat               float64   `json:"origin_lat"`
	OriginLng               float64   `json:"origin_lng"`
	Distance                int64     `json:"distance"`
	PickupDate              time.Time `json:"pickup_date"`
	DeliveryDate            time.Time `json:"delivery_date"`
	ExpirationDate          time.Time `json:"expiration_date"`
	Title                   string    `json:"title"`
	CargoType               string    `json:"cargo_type"`
	CargoSpecies            string    `json:"cargo_species"`
	CargoWeight             float64   `json:"cargo_weight"`
	VehiclesAccepted        string    `json:"vehicles_accepted"`
	Trailer                 string    `json:"trailer"`
	RequiresTarp            bool      `json:"requires_tarp"`
	Tracking                bool      `json:"tracking"`
	Agency                  bool      `json:"agency"`
	Description             string    `json:"description"`
	PaymentType             string    `json:"payment_type"`
	Advance                 string    `json:"advance"`
	Toll                    bool      `json:"toll"`
	Situation               string    `json:"situation"`
	Price                   float64   `json:"price"`
	StateOrigin             string    `json:"state_origin"`
	CityOrigin              string    `json:"city_origin"`
	ComplementOrigin        string    `json:"complement_origin"`
	NeighborhoodOrigin      string    `json:"neighborhood_origin"`
	StreetOrigin            string    `json:"street_origin"`
	StreetNumberOrigin      string    `json:"street_number_origin"`
	CEPOrigin               string    `json:"cep_origin"`
	StateDestination        string    `json:"state_destination"`
	CityDestination         string    `json:"city_destination"`
	ComplementDestination   string    `json:"complement_destination"`
	NeighborhoodDestination string    `json:"neighborhood_destination"`
	StreetDestination       string    `json:"street_destination"`
	StreetNumberDestination string    `json:"street_number_destination"`
	CEPDestination          string    `json:"cep_destination"`
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
	ID                      int64      `json:"id"`
	UserID                  int64      `json:"user_id"`
	Destination             string     `json:"destination"`
	Origin                  string     `json:"origin"`
	DestinationLat          float64    `json:"destination_lat"`
	DestinationLng          float64    `json:"destination_lng"`
	OriginLat               float64    `json:"origin_lat"`
	OriginLng               float64    `json:"origin_lng"`
	Distance                int64      `json:"distance"`
	PickupDate              time.Time  `json:"pickup_date"`
	DeliveryDate            time.Time  `json:"delivery_date"`
	ExpirationDate          time.Time  `json:"expiration_date"`
	Title                   string     `json:"title"`
	CargoType               string     `json:"cargo_type"`
	CargoSpecies            string     `json:"cargo_species"`
	CargoWeight             float64    `json:"cargo_weight"`
	VehiclesAccepted        string     `json:"vehicles_accepted"`
	Trailer                 string     `json:"trailer"`
	RequiresTarp            bool       `json:"requires_tarp"`
	Tracking                bool       `json:"tracking"`
	Agency                  bool       `json:"agency"`
	Description             string     `json:"description"`
	PaymentType             string     `json:"payment_type"`
	Advance                 string     `json:"advance"`
	Toll                    bool       `json:"toll"`
	Situation               string     `json:"situation"`
	Price                   float64    `json:"price"`
	StateOrigin             string     `json:"state_origin"`
	CityOrigin              string     `json:"city_origin"`
	ComplementOrigin        string     `json:"complement_origin"`
	NeighborhoodOrigin      string     `json:"neighborhood_origin"`
	StreetOrigin            string     `json:"street_origin"`
	StreetNumberOrigin      string     `json:"street_number_origin"`
	CEPOrigin               string     `json:"cep_origin"`
	StateDestination        string     `json:"state_destination"`
	CityDestination         string     `json:"city_destination"`
	ComplementDestination   string     `json:"complement_destination"`
	NeighborhoodDestination string     `json:"neighborhood_destination"`
	StreetDestination       string     `json:"street_destination"`
	StreetNumberDestination string     `json:"street_number_destination"`
	CEPDestination          string     `json:"cep_destination"`
	Status                  bool       `json:"status"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               *time.Time `json:"updated_at"`
	CreatedWho              string     `json:"created_who"`
	UpdatedWho              *string    `json:"updated_who"`
}

type AdvertisementResponseAll struct {
	ID                      int64                  `json:"id"`
	UserID                  int64                  `json:"user_id"`
	UserName                string                 `json:"user_name"`
	ActiveThere             time.Time              `json:"active_there"`
	ActiveDuration          string                 `json:"active_duration"`
	UserCity                string                 `json:"user_city"`
	UserState               string                 `json:"user_state"`
	UserPhone               string                 `json:"user_phone"`
	UserEmail               string                 `json:"user_email"`
	Destination             string                 `json:"destination"`
	Origin                  string                 `json:"origin"`
	DestinationLat          float64                `json:"destination_lat"`
	DestinationLng          float64                `json:"destination_lng"`
	OriginLat               float64                `json:"origin_lat"`
	OriginLng               float64                `json:"origin_lng"`
	Distance                int64                  `json:"distance"`
	PickupDate              time.Time              `json:"pickup_date"`
	DeliveryDate            time.Time              `json:"delivery_date"`
	ExpirationDate          time.Time              `json:"expiration_date"`
	Title                   string                 `json:"title"`
	CargoType               string                 `json:"cargo_type"`
	CargoSpecies            string                 `json:"cargo_species"`
	CargoWeight             float64                `json:"cargo_weight"`
	VehiclesAccepted        string                 `json:"vehicles_accepted"`
	Trailer                 string                 `json:"trailer"`
	RequiresTarp            bool                   `json:"requires_tarp"`
	Tracking                bool                   `json:"tracking"`
	Agency                  bool                   `json:"agency"`
	Description             string                 `json:"description"`
	PaymentType             string                 `json:"payment_type"`
	Advance                 string                 `json:"advance"`
	Toll                    bool                   `json:"toll"`
	Situation               string                 `json:"situation"`
	ActiveFreight           int64                  `json:"active_freight"`
	Price                   float64                `json:"price"`
	StateOrigin             string                 `json:"state_origin"`
	CityOrigin              string                 `json:"city_origin"`
	ComplementOrigin        string                 `json:"complement_origin"`
	NeighborhoodOrigin      string                 `json:"neighborhood_origin"`
	StreetOrigin            string                 `json:"street_origin"`
	StreetNumberOrigin      string                 `json:"street_number_origin"`
	CEPOrigin               string                 `json:"cep_origin"`
	StateDestination        string                 `json:"state_destination"`
	CityDestination         string                 `json:"city_destination"`
	ComplementDestination   string                 `json:"complement_destination"`
	NeighborhoodDestination string                 `json:"neighborhood_destination"`
	StreetDestination       string                 `json:"street_destination"`
	StreetNumberDestination string                 `json:"street_number_destination"`
	CEPDestination          string                 `json:"cep_destination"`
	RouteIndexChoose        int                    `json:"route_index_choose"`
	RouteChoose             new_routes.RouteOutput `json:"route_choose"`
	CreatedAt               time.Time              `json:"created_at"`
	CreatedWho              string                 `json:"created_who"`
	UpdatedAt               *time.Time             `json:"updated_at,omitempty"`
	UpdatedWho              *string                `json:"updated_who,omitempty"`
}

type AdvertisementResponseNoUser struct {
	ID                      int64     `json:"id"`
	UserID                  int64     `json:"user_id"`
	Destination             string    `json:"destination"`
	Origin                  string    `json:"origin"`
	PickupDate              time.Time `json:"pickup_date"`
	DeliveryDate            time.Time `json:"delivery_date"`
	ExpirationDate          time.Time `json:"expiration_date"`
	Title                   string    `json:"title"`
	CargoType               string    `json:"cargo_type"`
	CargoSpecies            string    `json:"cargo_species"`
	CargoWeight             float64   `json:"cargo_weight"`
	VehiclesAccepted        string    `json:"vehicles_accepted"`
	Trailer                 string    `json:"trailer"`
	RequiresTarp            bool      `json:"requires_tarp"`
	Tracking                bool      `json:"tracking"`
	Agency                  bool      `json:"agency"`
	Description             string    `json:"description"`
	PaymentType             string    `json:"payment_type"`
	Advance                 string    `json:"advance"`
	Toll                    bool      `json:"toll"`
	Situation               string    `json:"situation"`
	StateOrigin             string    `json:"state_origin"`
	CityOrigin              string    `json:"city_origin"`
	ComplementOrigin        string    `json:"complement_origin"`
	NeighborhoodOrigin      string    `json:"neighborhood_origin"`
	StreetOrigin            string    `json:"street_origin"`
	StreetNumberOrigin      string    `json:"street_number_origin"`
	CEPOrigin               string    `json:"cep_origin"`
	StateDestination        string    `json:"state_destination"`
	CityDestination         string    `json:"city_destination"`
	ComplementDestination   string    `json:"complement_destination"`
	NeighborhoodDestination string    `json:"neighborhood_destination"`
	StreetDestination       string    `json:"street_destination"`
	StreetNumberDestination string    `json:"street_number_destination"`
	CEPDestination          string    `json:"cep_destination"`
	CreatedAt               time.Time `json:"created_at"`
}

type UpdatedAdvertisementFinishedCreate struct {
	ID             int64   `json:"id"`
	RouteHistID    int64   `json:"route_hist_id"`
	RouteChoose    int64   `json:"route_choose"`
	UserID         int64   `json:"user_id"`
	DestinationLat float64 `json:"destination_lat"`
	DestinationLng float64 `json:"destination_lng"`
	OriginLat      float64 `json:"origin_lat"`
	OriginLng      float64 `json:"origin_lng"`
}

type ResponseUpdatedAdvertisementFinishedCreate struct {
	ID             int64   `json:"id"`
	UserID         int64   `json:"user_id"`
	RouteHistID    int64   `json:"route_hist_id"`
	RouteChoose    int64   `json:"route_choose"`
	DestinationLat float64 `json:"destination_lat"`
	DestinationLng float64 `json:"destination_lng"`
	OriginLat      float64 `json:"origin_lat"`
	OriginLng      float64 `json:"origin_lng"`
	Situation      string  `json:"situation"`
}

type RouteOutput struct {
	Summary      RouteSummary           `json:"summary"`
	Costs        Costs                  `json:"costs"`
	Tolls        []Toll                 `json:"tolls,omitempty"`
	Balances     interface{}            `json:"balances"`
	GasStations  []GasStation           `json:"gas_stations"`
	Instructions []Instruction          `json:"instructions"`
	FreightLoad  map[string]interface{} `json:"freight_load"`
	Polyline     string                 `json:"polyline"`
}

type RouteSummary struct {
	RouteType string   `json:"route_type"`
	HasTolls  bool     `json:"hasTolls"`
	Distance  Distance `json:"distance"`
	Duration  Duration `json:"duration"`
	URL       string   `json:"url"`
	URLWaze   string   `json:"url_waze"`
}

type Costs struct {
	TagAndCash      float64 `json:"tagAndCash"`
	FuelInTheCity   float64 `json:"fuel_in_the_city"`
	FuelInTheHwy    float64 `json:"fuel_in_the_hwy"`
	Tag             float64 `json:"tag"`
	Cash            float64 `json:"cash"`
	PrepaidCard     float64 `json:"prepaidCard"`
	MaximumTollCost float64 `json:"maximumTollCost"`
	MinimumTollCost float64 `json:"minimumTollCost"`
	Axles           int     `json:"axles"`
}

type Instruction struct {
	Text string `json:"text"`
	Img  string `json:"img"`
}

type Toll struct {
	ID              int             `json:"id"`
	Latitude        float64         `json:"lat"`
	Longitude       float64         `json:"lng"`
	Name            string          `json:"name"`
	Concession      string          `json:"concession"`
	ConcessionImg   string          `json:"concession_img"`
	Road            string          `json:"road"`
	State           string          `json:"state"`
	Country         string          `json:"country"`
	Type            string          `json:"type"`
	TagCost         float64         `json:"tagCost"`
	CashCost        float64         `json:"cashCost"`
	Currency        string          `json:"currency"`
	PrepaidCardCost float64         `json:"prepaidCardCost"`
	ArrivalResponse ArrivalResponse `json:"arrival"`
	TagPrimary      []string        `json:"tagPrimary"`
	TagImg          []string        `json:"tagImg"`
	FreeFlow        bool            `json:"free_flow"`
	PayFreeFlow     string          `json:"pay_free_flow"`
}

type GasStation struct {
	Name     string   `json:"name"`
	Address  string   `json:"address"`
	Location Location `json:"location"`
}

type Distance struct {
	Text  string  `json:"text"`
	Value float64 `json:"value"`
}

type Duration struct {
	Text  string  `json:"text"`
	Value float64 `json:"value"`
}

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type ArrivalResponse struct {
	Distance string `json:"distance"`
	Time     string `json:"time"`
}

func (p *CreateAdvertisementDto) ParseCreateToAdvertisement() db.CreateAdvertisementParams {
	arg := db.CreateAdvertisementParams{
		UserID:                  p.UserID,
		Destination:             p.CreateAdvertisementRequest.Destination,
		Origin:                  p.CreateAdvertisementRequest.Origin,
		Distance:                p.CreateAdvertisementRequest.Distance,
		PickupDate:              p.CreateAdvertisementRequest.PickupDate,
		DeliveryDate:            p.CreateAdvertisementRequest.DeliveryDate,
		ExpirationDate:          p.CreateAdvertisementRequest.ExpirationDate,
		Title:                   p.CreateAdvertisementRequest.Title,
		CargoType:               p.CreateAdvertisementRequest.CargoType,
		CargoSpecies:            p.CreateAdvertisementRequest.CargoSpecies,
		CargoWeight:             p.CreateAdvertisementRequest.CargoWeight,
		VehiclesAccepted:        p.CreateAdvertisementRequest.VehiclesAccepted,
		Trailer:                 p.CreateAdvertisementRequest.Trailer,
		RequiresTarp:            p.CreateAdvertisementRequest.RequiresTarp,
		Tracking:                p.CreateAdvertisementRequest.Tracking,
		Agency:                  p.CreateAdvertisementRequest.Agency,
		Description:             p.CreateAdvertisementRequest.Description,
		PaymentType:             p.CreateAdvertisementRequest.PaymentType,
		Advance:                 p.CreateAdvertisementRequest.Advance,
		Toll:                    p.CreateAdvertisementRequest.Toll,
		Price:                   p.CreateAdvertisementRequest.Price,
		StateOrigin:             p.CreateAdvertisementRequest.StateOrigin,
		CityOrigin:              p.CreateAdvertisementRequest.CityOrigin,
		ComplementOrigin:        p.CreateAdvertisementRequest.ComplementOrigin,
		NeighborhoodOrigin:      p.CreateAdvertisementRequest.NeighborhoodOrigin,
		StreetOrigin:            p.CreateAdvertisementRequest.StreetOrigin,
		StreetNumberOrigin:      p.CreateAdvertisementRequest.StreetNumberOrigin,
		CepOrigin:               p.CreateAdvertisementRequest.CEPOrigin,
		StateDestination:        p.CreateAdvertisementRequest.StateDestination,
		CityDestination:         p.CreateAdvertisementRequest.CityDestination,
		ComplementDestination:   p.CreateAdvertisementRequest.ComplementDestination,
		NeighborhoodDestination: p.CreateAdvertisementRequest.NeighborhoodDestination,
		StreetDestination:       p.CreateAdvertisementRequest.StreetDestination,
		StreetNumberDestination: p.CreateAdvertisementRequest.StreetNumberDestination,
		CepDestination:          p.CreateAdvertisementRequest.CEPDestination,
		CreatedWho:              p.CreatedWho,
	}
	return arg
}

func (p *UpdatedAdvertisementFinishedCreate) ParseUpdatedToAdvertisementFinishedCreate() db.UpdatedAdvertisementFinishedCreateParams {
	arg := db.UpdatedAdvertisementFinishedCreateParams{
		ID:     p.ID,
		UserID: p.UserID,
		DestinationLat: sql.NullFloat64{
			Float64: p.DestinationLat,
			Valid:   true,
		},
		DestinationLng: sql.NullFloat64{
			Float64: p.DestinationLng,
			Valid:   true,
		},
		OriginLat: sql.NullFloat64{
			Float64: p.OriginLat,
			Valid:   true,
		},
		OriginLng: sql.NullFloat64{
			Float64: p.OriginLng,
			Valid:   true,
		},
	}
	return arg
}

func (p *UpdatedAdvertisementFinishedCreate) ParseCreateToAdvertisementRoute() db.CreateAdvertisementRouteParams {
	arg := db.CreateAdvertisementRouteParams{
		AdvertisementID: p.ID,
		RouteHistID:     p.RouteHistID,
		UserID:          p.UserID,
		RouteChoose:     p.RouteChoose,
	}
	return arg
}

func (p *UpdateAdvertisementDto) ParseUpdateToAdvertisement() db.UpdateAdvertisementParams {
	arg := db.UpdateAdvertisementParams{
		UserID:      p.UserID,
		Destination: p.UpdateAdvertisementRequest.Destination,
		Origin:      p.UpdateAdvertisementRequest.Origin,
		DestinationLat: sql.NullFloat64{
			Float64: p.UpdateAdvertisementRequest.DestinationLat,
			Valid:   true,
		},
		DestinationLng: sql.NullFloat64{
			Float64: p.UpdateAdvertisementRequest.DestinationLng,
			Valid:   true,
		},
		OriginLat: sql.NullFloat64{
			Float64: p.UpdateAdvertisementRequest.OriginLat,
			Valid:   true,
		},
		OriginLng: sql.NullFloat64{
			Float64: p.UpdateAdvertisementRequest.OriginLng,
			Valid:   true,
		},
		Distance:                p.UpdateAdvertisementRequest.Distance,
		PickupDate:              p.UpdateAdvertisementRequest.PickupDate,
		DeliveryDate:            p.UpdateAdvertisementRequest.DeliveryDate,
		ExpirationDate:          p.UpdateAdvertisementRequest.ExpirationDate,
		Title:                   p.UpdateAdvertisementRequest.Title,
		CargoType:               p.UpdateAdvertisementRequest.CargoType,
		CargoSpecies:            p.UpdateAdvertisementRequest.CargoSpecies,
		CargoWeight:             p.UpdateAdvertisementRequest.CargoWeight,
		VehiclesAccepted:        p.UpdateAdvertisementRequest.VehiclesAccepted,
		Trailer:                 p.UpdateAdvertisementRequest.Trailer,
		RequiresTarp:            p.UpdateAdvertisementRequest.RequiresTarp,
		Tracking:                p.UpdateAdvertisementRequest.Tracking,
		Agency:                  p.UpdateAdvertisementRequest.Agency,
		Description:             p.UpdateAdvertisementRequest.Description,
		PaymentType:             p.UpdateAdvertisementRequest.PaymentType,
		Advance:                 p.UpdateAdvertisementRequest.Advance,
		Toll:                    p.UpdateAdvertisementRequest.Toll,
		Situation:               p.UpdateAdvertisementRequest.Situation,
		Price:                   p.UpdateAdvertisementRequest.Price,
		UpdatedWho:              p.UpdatedWho,
		StateOrigin:             p.UpdateAdvertisementRequest.StateOrigin,
		CityOrigin:              p.UpdateAdvertisementRequest.CityOrigin,
		ComplementOrigin:        p.UpdateAdvertisementRequest.ComplementOrigin,
		NeighborhoodOrigin:      p.UpdateAdvertisementRequest.NeighborhoodOrigin,
		StreetOrigin:            p.UpdateAdvertisementRequest.StreetOrigin,
		StreetNumberOrigin:      p.UpdateAdvertisementRequest.StreetNumberOrigin,
		CepOrigin:               p.UpdateAdvertisementRequest.CEPOrigin,
		StateDestination:        p.UpdateAdvertisementRequest.StateDestination,
		CityDestination:         p.UpdateAdvertisementRequest.CityDestination,
		ComplementDestination:   p.UpdateAdvertisementRequest.ComplementDestination,
		NeighborhoodDestination: p.UpdateAdvertisementRequest.NeighborhoodDestination,
		StreetDestination:       p.UpdateAdvertisementRequest.StreetDestination,
		StreetNumberDestination: p.UpdateAdvertisementRequest.StreetNumberDestination,
		CepDestination:          p.UpdateAdvertisementRequest.CEPDestination,
		ID:                      p.UpdateAdvertisementRequest.ID,
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
	p.DestinationLat = result.DestinationLat.Float64
	p.DestinationLng = result.DestinationLng.Float64
	p.OriginLat = result.OriginLat.Float64
	p.OriginLng = result.OriginLng.Float64
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
	p.StateOrigin = result.StateOrigin
	p.CityOrigin = result.CityOrigin
	p.ComplementOrigin = result.ComplementOrigin
	p.NeighborhoodOrigin = result.NeighborhoodOrigin
	p.StreetOrigin = result.StreetOrigin
	p.StreetNumberOrigin = result.StreetNumberOrigin
	p.CEPOrigin = result.CepOrigin
	p.StateDestination = result.StateDestination
	p.CityDestination = result.CityDestination
	p.ComplementDestination = result.ComplementDestination
	p.NeighborhoodDestination = result.NeighborhoodDestination
	p.StreetDestination = result.StreetDestination
	p.StreetNumberDestination = result.StreetNumberDestination
	p.CEPDestination = result.CepDestination
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

func (p *AdvertisementResponseAll) ParseFromAdvertisementObject(
	result db.GetAllAdvertisementUsersRow,
) {
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
	p.DestinationLat = result.DestinationLat.Float64
	p.DestinationLng = result.DestinationLng.Float64
	p.OriginLat = result.OriginLat.Float64
	p.OriginLng = result.OriginLng.Float64
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
	p.StateOrigin = result.StateOrigin
	p.CityOrigin = result.CityOrigin
	p.ComplementOrigin = result.ComplementOrigin
	p.NeighborhoodOrigin = result.NeighborhoodOrigin
	p.StreetOrigin = result.StreetOrigin
	p.StreetNumberOrigin = result.StreetNumberOrigin
	p.CEPOrigin = result.CepOrigin
	p.StateDestination = result.StateDestination
	p.CityDestination = result.CityDestination
	p.ComplementDestination = result.ComplementDestination
	p.NeighborhoodDestination = result.NeighborhoodDestination
	p.StreetDestination = result.StreetDestination
	p.StreetNumberDestination = result.StreetNumberDestination
	p.CEPDestination = result.CepDestination
	p.CreatedAt = result.CreatedAt
	p.CreatedWho = result.CreatedWho
	if result.UpdatedAt.Valid {
		p.UpdatedAt = &result.UpdatedAt.Time
	}
	if result.UpdatedWho.Valid {
		p.UpdatedWho = &result.UpdatedWho.String
	}
}

func (p *AdvertisementResponseNoUser) ParseFromAdvertisementObject(
	result db.GetAllAdvertisementPublicRow,
) {
	p.ID = result.ID
	p.UserID = result.UserID
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
	p.Situation = result.Situation
	p.StateOrigin = result.StateOrigin
	p.CityOrigin = result.CityOrigin
	p.ComplementOrigin = result.ComplementOrigin
	p.NeighborhoodOrigin = result.NeighborhoodOrigin
	p.StreetOrigin = result.StreetOrigin
	p.StreetNumberOrigin = result.StreetNumberOrigin
	p.CEPOrigin = result.CepOrigin
	p.StateDestination = result.StateDestination
	p.CityDestination = result.CityDestination
	p.ComplementDestination = result.ComplementDestination
	p.NeighborhoodDestination = result.NeighborhoodDestination
	p.StreetDestination = result.StreetDestination
	p.StreetNumberDestination = result.StreetNumberDestination
	p.CEPDestination = result.CepDestination
	p.CreatedAt = result.CreatedAt
}

func (p *AdvertisementResponseAll) ParseFromAdvertisementByIDObject(
	result db.GetAllAdvertisementByUserRow,
) {
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
	p.DestinationLat = result.DestinationLat.Float64
	p.DestinationLng = result.DestinationLng.Float64
	p.OriginLat = result.OriginLat.Float64
	p.OriginLng = result.OriginLng.Float64
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
	p.StateOrigin = result.StateOrigin
	p.CityOrigin = result.CityOrigin
	p.ComplementOrigin = result.ComplementOrigin
	p.NeighborhoodOrigin = result.NeighborhoodOrigin
	p.StreetOrigin = result.StreetOrigin
	p.StreetNumberOrigin = result.StreetNumberOrigin
	p.CEPOrigin = result.CepOrigin
	p.StateDestination = result.StateDestination
	p.CityDestination = result.CityDestination
	p.ComplementDestination = result.ComplementDestination
	p.NeighborhoodDestination = result.NeighborhoodDestination
	p.StreetDestination = result.StreetDestination
	p.StreetNumberDestination = result.StreetNumberDestination
	p.CEPDestination = result.CepDestination
	p.CreatedAt = result.CreatedAt
	p.CreatedWho = result.CreatedWho
	if result.UpdatedAt.Valid {
		p.UpdatedAt = &result.UpdatedAt.Time
	}
	if result.UpdatedWho.Valid {
		p.UpdatedWho = &result.UpdatedWho.String
	}
}

func (p *ResponseUpdatedAdvertisementFinishedCreate) ParseFromUpdatedAdvertisementFinishedCreateObject(
	idRouteHist, idRouteChoose int64,
	result db.UpdatedAdvertisementFinishedCreateRow,
) {
	p.ID = result.ID
	p.UserID = result.UserID
	p.OriginLat = result.OriginLat.Float64
	p.OriginLng = result.OriginLng.Float64
	p.DestinationLat = result.DestinationLat.Float64
	p.DestinationLng = result.DestinationLng.Float64
	p.RouteHistID = idRouteHist
	p.RouteChoose = idRouteChoose
	p.Situation = result.Situation
}

type UpdateAdsRouteChooseRequest struct {
	AdvertisementID int64 `json:"advertisement_id"`
	NewRoute        int64 `json:"new_route"`
}

type UpdateAdsRouteChooseDTO struct {
	Request UpdateAdsRouteChooseRequest `json:"request"`
	UserID  int64                       `json:"user_id"`
}
