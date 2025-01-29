package routes

import (
	"googlemaps.github.io/maps"
	"time"
)

type FrontInfo struct {
	Origin          string  `json:"origin"`
	Destination     string  `json:"destination"`
	ConsumptionCity float64 `json:"consumptionCity"`
	ConsumptionHwy  float64 `json:"consumptionHwy"`
	Price           float64 `json:"price"`
	Axles           int     `json:"axles"`
	Type            string  `json:"type"`
}

// FEATURE
//var vehicleAxleMap = map[string]int{
//	"Utilitário":              2,
//	"VUC":                     2,
//	"Caminhão 3/4":            2,
//	"Caminhão Toco":           2,
//	"Caminhão Truck":          3,
//	"Cavalo Mecânico Simples": 2,
//	"Cavalo Mecânico Trucado": 3,
//	"Carreta 2 Eixos":         4,
//	"Carreta 3 Eixos":         5,
//	"Bitrem":                  7,
//	"Bitrenzão":               9,
//	"Rodotrem":                9,
//	"Treminhão":               8,
//	"Tritrem":                 9,
//	"Bitruck":                 4,
//}
//
//func validateVehicleTypeAndAxles(info FrontInfo) error {
//	expectedAxles, exists := vehicleAxleMap[info.Type]
//	if !exists {
//		return errors.New("tipo de veículo inválido")
//	}
//	if info.Axles != expectedAxles {
//		return fmt.Errorf("número de eixos (%d) não corresponde ao esperado (%d) para o tipo de veículo %s", info.Axles, expectedAxles, info.Type)
//	}
//	return nil
//}

type Response struct {
	SummaryRoute SummaryRoute `json:"summary"`
	Routes       []Route      `json:"routes"`
}

type SummaryRoute struct {
	RouteOrigin      PrincipalRoute `json:"location_origin"`
	RouteDestination PrincipalRoute `json:"location_destination"`
	FuelPrice        FuelPrice      `json:"fuel_price"`
	FuelEfficiency   FuelEfficiency `json:"fuel_efficiency"`
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
	Summary  Summary `json:"summary"`
	Costs    Costs   `json:"costs"`
	Tolls    []Toll  `json:"tolls"`
	Polyline string  `json:"polyline"`
}

type Summary struct {
	HasTolls bool     `json:"hasTolls"`
	Distance Distance `json:"distance"`
	Duration Duration `json:"duration"`
	URL      string   `json:"url"`
	Name     string   `json:"name"`
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
	Fuel            float64 `json:"fuel"`
	Tag             float64 `json:"tag"`
	Cash            float64 `json:"cash"`
	LicensePlate    string  `json:"licensePlate"`
	PrepaidCard     float64 `json:"prepaidCard"`
	MaximumTollCost float64 `json:"maximumTollCost"`
	MinimumTollCost float64 `json:"minimumTollCost"`
}

type Toll struct {
	ID              int      `json:"id"`
	Latitude        float64  `json:"lat"`
	Longitude       float64  `json:"lng"`
	Name            string   `json:"name"`
	Road            string   `json:"road"`
	State           string   `json:"state"`
	Country         string   `json:"country"`
	Type            string   `json:"type"`
	TagCost         float64  `json:"tagCost"`
	CashCost        float64  `json:"cashCost"`
	Currency        string   `json:"currency"`
	PrepaidCardCost string   `json:"prepaidCardCost"`
	Arrival         Arrival  `json:"arrival"`
	TagPrimary      []string `json:"tagPrimary"`
}

type Arrival struct {
	Distance float64   `json:"distance"`
	Time     time.Time `json:"time"`
}

type PlaceRequest struct {
	Latitude  float64 `form:"latitude" binding:"required"`
	Longitude float64 `form:"longitude" binding:"required"`
}

type PlaceResponse struct {
	Place maps.PlacesSearchResult `json:"place"`
}
