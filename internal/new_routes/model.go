package routes

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
	Summary      RouteSummary             `json:"summary"`
	Costs        Costs                    `json:"costs"`
	Tolls        []Toll                   `json:"tolls,omitempty"`
	Balances     interface{}              `json:"balances"`
	GasStations  []GasStation             `json:"gas_stations"`
	Instructions []Instruction            `json:"instructions"`
	FreightLoad  map[string][]FreightLoad `json:"freight_load"`
	Polyline     string                   `json:"polyline"`
}

type Toll struct {
	ID               int64   `json:"id"`
	Concessionaria   string  `json:"concessionaria"`
	PracaDePedagio   string  `json:"praca_de_pedagio"`
	AnoDoPNVSNV      int64   `json:"ano_do_pnv_snv"`
	Rodovia          string  `json:"rodovia"`
	UF               string  `json:"uf"`
	KmM              string  `json:"km_m"`
	Municipio        string  `json:"municipio"`
	TipoPista        string  `json:"tipo_pista"`
	Sentido          string  `json:"sentido"`
	Situacao         string  `json:"situacao"`
	DataDaInativacao string  `json:"data_da_inativacao"`
	Latitude         float64 `json:"lat"`
	Longitude        float64 `json:"lng"`
	Tarifa           float64 `json:"tarifa"`
	FreeFlow         bool    `json:"free_flow"`
	PayFreeFlow      string  `json:"pay_free_flow"`
}

type GasStation struct {
	Name     string   `json:"name"`
	Address  string   `json:"address"`
	Location Location `json:"location"`
}

type FreightLoad struct {
	Description string  `json:"description"`
	QtdAxle     int     `json:"qtd_axle"`
	TotalValue  float64 `json:"total_value"`
	TypeOfLoad  string  `json:"type_of_load"`
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
