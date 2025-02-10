package routes

import (
	"encoding/json"
	"googlemaps.github.io/maps"
	"time"
)

type FrontInfo struct {
	Origin          string   `json:"origin" validate:"required"`
	Destination     string   `json:"destination" validate:"required"`
	ConsumptionCity float64  `json:"consumptionCity" validate:"required"`
	ConsumptionHwy  float64  `json:"consumptionHwy" validate:"required"`
	Price           float64  `json:"price" validate:"required"`
	Axles           int      `json:"axles" validate:"required"`
	Type            string   `json:"type" validate:"required,oneof=Truck Bus Auto Motorcycle truck bus auto motorcycle"`
	Waypoints       []string `json:"waypoints"`
	TypeRoute       string   `json:"typeRoute"`
}

type Response struct {
	SummaryRoute   SummaryRoute `json:"summary"`
	Routes         []Route      `json:"routes"`
	FastestRoute   string       `json:"fastest_route"`
	CheapestRoute  string       `json:"cheapest_route"`
	EfficientRoute string       `json:"efficient_route"`
	SelectedRoute  string       `json:"selected_route"`
}

type SummaryRoute struct {
	RouteOrigin      PrincipalRoute   `json:"location_origin"`
	RouteDestination PrincipalRoute   `json:"location_destination"`
	AllWayPoints     []PrincipalRoute `json:"all_stopping_points"`
	FuelPrice        FuelPrice        `json:"fuel_price"`
	FuelEfficiency   FuelEfficiency   `json:"fuel_efficiency"`
}

type PrincipalRoute struct {
	Location Location `json:"location"`
	Address  string   `json:"address"`
}

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type FuelEfficiency struct {
	City     float64 `json:"city"`
	Hwy      float64 `json:"hwy"`
	Units    string  `json:"units"`
	FuelUnit string  `json:"fuel_unit"`
}

type FuelPrice struct {
	Value    float64 `json:"city"`
	Currency string  `json:"hwy"`
	Units    string  `json:"units"`
	FuelUnit string  `json:"fuel_unit"`
}

type Route struct {
	Summary      Summary        `json:"summary"`
	Costs        Costs          `json:"costs"`
	Tolls        []Toll         `json:"tolls"`
	Balanca      []Balanca      `json:"balances"`
	GasStations  []GasStation   `json:"gas_stations"`
	Polyline     string         `json:"polyline"`
	Instructions []Instructions `json:"instructions"`
}

type Instructions struct {
	Text string `json:"text"`
	Img  string `json:"img"`
}

type Balanca struct {
	ID             int     `json:"id"`
	Concessionaria string  `json:"concessionaria"`
	Km             string  `json:"km"`
	Lat            float64 `json:"lat"`
	Lng            float64 `json:"lng"`
	Nome           string  `json:"nome"`
	Rodovia        string  `json:"rodovia"`
	Sentido        string  `json:"sentido"`
	Uf             string  `json:"uf"`
}

type Summary struct {
	RouteType string   `json:"route_type"`
	HasTolls  bool     `json:"hasTolls"`
	Distance  Distance `json:"distance"`
	Duration  Duration `json:"duration"`
	URL       string   `json:"url"`
	URLWaze   string   `json:"url_waze"`
}

type GeocodeResult struct {
	FormattedAddress string
	PlaceID          string
}

type GasStation struct {
	Name     string   `json:"name"`
	Address  string   `json:"address"`
	Location Location `json:"location"`
}

type Distance struct {
	Text  string `json:"text"`
	Value int    `json:"value"`
}

type Duration struct {
	Text  string  `json:"text"`
	Value float64 `json:"value"`
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
}

type Toll struct {
	ID              int             `json:"id"`
	Latitude        float64         `json:"lat"`
	Longitude       float64         `json:"lng"`
	Name            string          `json:"name"`
	Concession      string          `json:"concession"`
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
}

type Arrival struct {
	Distance string        `json:"distance"`
	Time     time.Duration `json:"time"`
}
type ArrivalResponse struct {
	Distance string `json:"distance"`
	Time     string `json:"time"`
}

type PlaceRequest struct {
	Latitude  float64 `form:"latitude" binding:"required"`
	Longitude float64 `form:"longitude" binding:"required"`
}

type PlaceResponse struct {
	Place maps.PlacesSearchResult `json:"place"`
}

type CreateFavoriteRouteRequest struct {
	TollsID          int64           `json:"tolls_id"`
	Response         json.RawMessage `json:"response"`
	UserOrganization string          `json:"user_organization"`
}
