package new_routes

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

type AddressCoordinatesResponse struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   string  `json:"address"`
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
	RouteHistID         int64          `json:"route_hist_id"`
	RouteOptions        RouteOptions   `json:"route_options"`
}

type Response struct {
	Routes      []DetailedRoute `json:"routes"`
	TotalRoute  TotalSummary    `json:"total_route"`
	TotalRoutes []TotalSummary  `json:"total_routes_all,omitempty"`
}

type DetailedRoute struct {
	LocationOrigin      AddressInfo    `json:"location_origin"`
	LocationDestination AddressInfo    `json:"location_destination"`
	HasRisk             bool           `json:"has_risk"`
	LocationHisk        []LocationHisk `json:"location_hisk"`
	Summaries           []RouteSummary `json:"summaries"`
}

type LocationHisk struct {
	CEP       string  `json:"cep"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type TotalSummary struct {
	LocationOrigin      AddressInfo        `json:"location_origin"`
	LocationDestination AddressInfo        `json:"location_destination"`
	TotalDistance       Distance           `json:"distance"`
	TotalDuration       Duration           `json:"duration"`
	URL                 string             `json:"url"`
	URLWaze             string             `json:"url_waze"`
	TotalTolls          float64            `json:"total_tolls"`
	TotalFuelCost       float64            `json:"total_fuel_cost"`
	Tolls               []Toll             `json:"tolls"`
	Balances            interface{}        `json:"balances"`
	Polyline            string             `json:"polyline"`
	Instructions        []Instruction      `json:"instructions,omitempty"`
	AttentionZones      *AttentionZoneInfo `json:"attention_zones"`
	RouteType           string             `json:"route_type,omitempty"`
}
type SummaryResponse struct {
	LocationOrigin      AddressInfo    `json:"location_origin"`
	LocationDestination AddressInfo    `json:"location_destination"`
	FuelPrice           FuelPrice      `json:"fuel_price"`
	FuelEfficiency      FuelEfficiency `json:"fuel_efficiency"`
	RouteOptions        RouteOptions   `json:"route_options"`
}

type RiskOffsets struct {
	Zone      RiskZone `json:"zone"`
	Entry     Location `json:"entry"`
	Exit      Location `json:"exit"`
	Before5km Location `json:"before_5km"`
	After5km  Location `json:"after_5km"`
	EntryCum  float64  `json:"entry_cum_m"`
	ExitCum   float64  `json:"exit_cum_m"`
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

type RiskZone struct {
	ID             int64   `json:"id"`
	Name           string  `json:"name"`
	Cep            string  `json:"cep"`
	Lat            float64 `json:"lat"`
	Lng            float64 `json:"lng"`
	Radius         int64   `json:"radius"`
	Status         bool    `json:"status"`
	ZonasAtencao   bool    `json:"zonas_atencao"`
	OrganizationID int64   `json:"organization_id"`
}

type RouteSummary struct {
	RouteType      string             `json:"route_type"`
	HasTolls       bool               `json:"hasTolls"`
	Distance       Distance           `json:"distance"`
	Duration       Duration           `json:"duration"`
	URL            string             `json:"url"`
	URLWaze        string             `json:"url_waze"`
	TotalFuelCost  float64            `json:"total_fuel_cost,omitempty"`
	Tolls          []Toll             `json:"tolls,omitempty"`
	TotalTolls     float64            `json:"total_tolls,omitempty"`
	Polyline       string             `json:"polyline,omitempty"`
	Instructions   []Instruction      `json:"instructions,omitempty"`
	AttentionZones *AttentionZoneInfo `json:"attention_zones"`
	RiskInfo       *RiskOffsets       `json:"risk_info,omitempty"`
	Detour         *DetourPlan        `json:"detour,omitempty"`
}
type DetourPlan struct {
	Source string        `json:"source"`
	Points []DetourPoint `json:"points"`
}

type DetourPoint struct {
	Name     string   `json:"name"`
	Location Location `json:"location"`
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
	Costs        *Costs                 `json:"costs,omitempty"`
	Tolls        []Toll                 `json:"tolls,omitempty"`
	Balances     interface{}            `json:"balances,omitempty"`
	GasStations  []GasStation           `json:"gas_stations,omitempty"`
	Instructions []Instruction          `json:"instructions,omitempty"`
	FreightLoad  map[string]interface{} `json:"freight_load,omitempty"`
	Polyline     string                 `json:"polyline,omitempty"`
}

type RouteOutputSimp struct {
	Summary RouteSummary `json:"summary"`
}

type GasStation struct {
	Name     string   `json:"name"`
	Address  string   `json:"address"`
	Location Location `json:"location"`
}

type FinalOutput struct {
	Summary           Summary       `json:"summary"`
	RouteEnterpriseId int64         `json:"route_enterprise_id,omitempty"`
	Routes            []RouteOutput `json:"routes"`
}

type FinalOutputSimp struct {
	Summary Summary           `json:"summary"`
	Routes  []RouteOutputSimp `json:"routes"`
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

type RouteOptions struct {
	IncludeFuelStations  bool `json:"include_fuel_stations"`  // Se deve incluir postos de combustível
	IncludeRouteMap      bool `json:"include_route_map"`      // Se deve incluir rotograma
	IncludeTollCosts     bool `json:"include_toll_costs"`     // Se deve incluir pedágios
	IncludeWeighStations bool `json:"include_weigh_stations"` // Se deve incluir balanças
	IncludeFreightCalc   bool `json:"include_freight_calc"`   // Se deve calcular frete
	IncludePolyline      bool `json:"include_polyline"`       // Se deve retornar polyline
}

type FrontInfo struct {
	Origin          string       `json:"origin" validate:"required"`
	Destination     string       `json:"destination" validate:"required"`
	ConsumptionCity float64      `json:"consumptionCity"`
	ConsumptionHwy  float64      `json:"consumptionHwy"`
	Price           float64      `json:"price"`
	Axles           int64        `json:"axles"`
	Type            string       `json:"type" validate:"required,oneof=Truck Bus Auto Motorcycle truck bus auto motorcycle"`
	Waypoints       []string     `json:"waypoints"`
	TypeRoute       string       `json:"typeRoute"`
	PublicOrPrivate string       `json:"public_or_private"`
	Favorite        bool         `json:"favorite"`
	RouteOptions    RouteOptions `json:"route_options"`
}

type FrontInfoCEP struct {
	OriginCEP       string       `json:"origin_cep" validate:"required"`
	DestinationCEP  string       `json:"destination_cep" validate:"required"`
	ConsumptionCity float64      `json:"consumptionCity"`
	ConsumptionHwy  float64      `json:"consumptionHwy"`
	Price           float64      `json:"price"`
	Axles           int64        `json:"axles"`
	Type            string       `json:"type" validate:"required,oneof=Truck Bus Auto Motorcycle truck bus auto motorcycle"`
	WaypointsCEP    []string     `json:"waypoints"`
	TypeRoute       string       `json:"typeRoute"`
	PublicOrPrivate string       `json:"public_or_private"`
	Favorite        bool         `json:"favorite"`
	RouteOptions    RouteOptions `json:"route_options"`
	Enterprise      bool         `json:"enterprise"`
}

type FrontInfoCEPRequest struct {
	CEPs            []string     `json:"ceps"`
	ConsumptionCity float64      `json:"consumptionCity"`
	ConsumptionHwy  float64      `json:"consumptionHwy"`
	Price           float64      `json:"price"`
	Axles           int64        `json:"axles"`
	Type            string       `json:"type" validate:"required,oneof=Truck Bus Auto Motorcycle truck bus auto motorcycle"`
	TypeRoute       string       `json:"typeRoute"`
	RouteOptions    RouteOptions `json:"route_options"`
	Waypoints       []Coordinate `json:"waypoints"`
	OrganizationID  int64        `json:"organization_id" validate:"required"`
}

type FrontInfoCoordinate struct {
	OriginLng       string       `json:"origin_lng" validate:"required"`
	OriginLat       string       `json:"origin_lat" validate:"required"`
	DestinationLat  string       `json:"destination_lat" validate:"required"`
	DestinationLng  string       `json:"destination_lng" validate:"required"`
	ConsumptionCity float64      `json:"consumptionCity"`
	ConsumptionHwy  float64      `json:"consumptionHwy"`
	Price           float64      `json:"price"`
	Axles           int64        `json:"axles"`
	Type            string       `json:"type" validate:"required,oneof=Truck Bus Auto Motorcycle truck bus auto motorcycle"`
	Waypoints       []Coordinate `json:"waypoints"`
	TypeRoute       string       `json:"typeRoute"`
	PublicOrPrivate string       `json:"public_or_private"`
	Favorite        bool         `json:"favorite"`
	RouteOptions    RouteOptions `json:"route_options"`
}

type FrontInfoCoordinatesRequest struct {
	Coordinates     []Coordinate `json:"coordinates" validate:"required,min=2"`
	ConsumptionCity float64      `json:"consumptionCity"`
	ConsumptionHwy  float64      `json:"consumptionHwy"`
	Price           float64      `json:"price"`
	Axles           int64        `json:"axles"`
	Type            string       `json:"type" validate:"required,oneof=Truck Bus Auto Motorcycle truck bus auto motorcycle"`
	TypeRoute       string       `json:"typeRoute"`
	RouteOptions    RouteOptions `json:"route_options"`
	Waypoints       []Coordinate `json:"waypoints"`
	OrganizationID  int64        `json:"organization_id" validate:"required"`
}

type Coordinate struct {
	Lat string `json:"lat"`
	Lng string `json:"lng"`
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

type SimpleRouteRequest struct {
	OriginLat float64 `json:"origin_lat"`
	OriginLng float64 `json:"origin_lng"`
	DestLat   float64 `json:"destination_lat"`
	DestLng   float64 `json:"destination_lng"`
}

type SimpleSummary struct {
	LocationOrigin      AddressInfo        `json:"location_origin"`
	LocationDestination AddressInfo        `json:"location_destination"`
	SimpleRoute         SimpleRouteSummary `json:"routes"`
}

type SimpleRouteResponse struct {
	Summary SimpleSummary `json:"summary"`
}

type SimpleRouteSummary struct {
	Distance Distance `json:"distance"`
	Duration Duration `json:"duration"`
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

type APIBrasilResponse struct {
	Error    bool   `json:"error"`
	Message  string `json:"message"`
	Response struct {
		CEP struct {
			CEP        string `json:"cep"`
			Tipo       string `json:"tipo"`
			Logradouro string `json:"logradouro"`
			Estado     string `json:"estado"`
			Latitude   string `json:"latitude"`
			Longitude  string `json:"longitude"`
			Cidade     struct {
				Cidade string `json:"cidade"`
			} `json:"cidade"`
			Bairro struct {
				Bairro string `json:"bairro"`
			} `json:"bairro"`
		} `json:"cep"`
	} `json:"response"`
}
type APIBrasilCoordenada struct {
	Latitude  float64 `json:"laatitude"`
	Longitude float64 `json:"longitude"`
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

// Ponto de entrada/saída da zona de atenção
type AttentionZoneEvent struct {
	Type          string   `json:"type"`           // "entry" ou "exit"
	ZoneName      string   `json:"zone_name"`      // Nome da zona
	ZoneID        int64    `json:"zone_id"`        // ID da zona
	Distance      float64  `json:"distance"`       // Distância do início da rota até este ponto
	Coordinates   Location `json:"coordinates"`    // Coordenadas onde acontece a entrada/saída
	Message       string   `json:"message"`        // Mensagem a ser exibida
	DetectionType string   `json:"detection_type"` // "area" ou "street" - como foi detectada
	StreetName    string   `json:"street_name"`    // Nome da rua (se detectada por rua)
}

// Informações sobre zonas de atenção encontradas na rota
type AttentionZoneInfo struct {
	HasAttentionZones bool                 `json:"has_attention_zones"`
	Events            []AttentionZoneEvent `json:"events"`
	TotalDistance     float64              `json:"total_distance_in_zones"`
	ZoneNames         []string             `json:"zone_names"`
}
