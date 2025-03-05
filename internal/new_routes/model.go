package routes

import (
	"encoding/json"
	db "geolocation/db/sqlc"
	"time"
)

type NominatimResult struct {
	DisplayName string `json:"display_name"`
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
}

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type AddressInfo struct {
	Location Location `json:"location"`
	Address  string   `json:"address"`
}

type FuelPrice struct {
	Price    float64 `json:"price"`
	Currency string  `json:"currency"`
	Units    string  `json:"units"`
	FuelUnit string  `json:"fuel_unit"`
}

type FuelEfficiency struct {
	City     float64 `json:"city"`
	Hwy      float64 `json:"hwy"`
	Units    string  `json:"units"`
	FuelUnit string  `json:"fuel_unit"`
}

type Summary struct {
	LocationOrigin      AddressInfo    `json:"location_origin"`
	LocationDestination AddressInfo    `json:"location_destination"`
	AllStoppingPoints   []interface{}  `json:"all_stopping_points"`
	FuelPrice           FuelPrice      `json:"fuel_price"`
	FuelEfficiency      FuelEfficiency `json:"fuel_efficiency"`
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

type Distance struct {
	Text  string  `json:"text"`
	Value float64 `json:"value"`
}

type Duration struct {
	Text  string  `json:"text"`
	Value float64 `json:"value"`
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

type GasStation struct {
	Name     string   `json:"name"`
	Address  string   `json:"address"`
	Location Location `json:"location"`
}

type FinalOutput struct {
	Summary Summary       `json:"summary"`
	Routes  []RouteOutput `json:"routes"`
}

type OSRMResponse struct {
	Code   string      `json:"code"`
	Routes []OSRMRoute `json:"routes"`
}

type OSRMRoute struct {
	Distance float64   `json:"distance"`
	Duration float64   `json:"duration"`
	Legs     []OSRMLeg `json:"legs"`
	Geometry string    `json:"geometry"`
}

type OSRMLeg struct {
	Distance float64    `json:"distance"`
	Duration float64    `json:"duration"`
	Steps    []OSRMStep `json:"steps"`
}

type OSRMStep struct {
	Distance float64 `json:"distance"`
	Duration float64 `json:"duration"`
	Name     string  `json:"name"`
	Maneuver struct {
		Location [2]float64 `json:"location"`
		Type     string     `json:"type"`
		Modifier string     `json:"modifier"`
	} `json:"maneuver"`
	Geometry string `json:"geometry"`
}

type LatLng struct {
	Lat float64
	Lng float64
}

type FrontInfo struct {
	Origin          string   `json:"origin" validate:"required"`
	Destination     string   `json:"destination" validate:"required"`
	ConsumptionCity float64  `json:"consumptionCity"`
	ConsumptionHwy  float64  `json:"consumptionHwy"`
	Price           float64  `json:"price"`
	Axles           int64    `json:"axles"`
	Type            string   `json:"type" validate:"required,oneof=Truck Bus Auto Motorcycle truck bus auto motorcycle"`
	Waypoints       []string `json:"waypoints"`
	TypeRoute       string   `json:"typeRoute"`
	PublicOrPrivate string   `json:"public_or_private"`
	Favorite        bool     `json:"favorite"`
}

type GeocodeResult struct {
	FormattedAddress string
	PlaceID          string
	Location         Location
}

type Arrival struct {
	Distance string        `json:"distance"`
	Time     time.Duration `json:"time"`
}
type ArrivalResponse struct {
	Distance string `json:"distance"`
	Time     string `json:"time"`
}

type FreightLoad struct {
	TypeOfLoad  string `json:"type_of_load"`
	TwoAxes     string `json:"two_axes"`
	ThreeAxes   string `json:"three_axes"`
	FourAxes    string `json:"four_axes"`
	FiveAxes    string `json:"five_axes"`
	SixAxes     string `json:"six_axes"`
	SevenAxes   string `json:"seven_axes"`
	NineAxes    string `json:"nine_axes"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (p *FreightLoad) ParseFromFreightObject(result db.FreightLoad) {
	p.TypeOfLoad = result.TypeOfLoad.String
	p.TwoAxes = result.TwoAxes.String
	p.ThreeAxes = result.ThreeAxes.String
	p.FourAxes = result.FourAxes.String
	p.FiveAxes = result.FiveAxes.String
	p.SixAxes = result.SixAxes.String
	p.SevenAxes = result.SevenAxes.String
	p.NineAxes = result.NineAxes.String
	p.ThreeAxes = result.ThreeAxes.String
	p.Name = result.Name.String
	p.Description = result.Description.String
}

type FavoriteRouteResponse struct {
	ID          int64           `json:"id"`
	IDUser      int64           `json:"id_user"`
	Origin      string          `json:"origin"`
	Destination string          `json:"destination"`
	Waypoints   string          `json:"waypoints"`
	Response    json.RawMessage `json:"response"`
	CreatedAt   time.Time       `json:"created_at"`
}

func (p *FavoriteRouteResponse) ParseFromFavoriteRouteObject(result db.FavoriteRoute) {
	p.ID = result.ID
	p.IDUser = result.IDUser
	p.Origin = result.Origin
	p.Destination = result.Destination
	p.Waypoints = result.Waypoints.String
	p.Response = result.Response
	p.CreatedAt = result.CreatedAt
}
