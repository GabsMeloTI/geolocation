package meiliaddress

type MeiliStreets struct {
	StreetID         int32   `json:"street_id"`
	StreetName       string  `json:"street"`
	NeighborhoodName string  `json:"neighborhood"`
	NeighborhoodLat  float64 `json:"neighborhood_lat"`
	NeighborhoodLon  float64 `json:"neighborhood_lon"`
	CityName         string  `json:"city"`
	CityLat          float64 `json:"city_lat"`
	CityLon          float64 `json:"city_lon"`
	StateUf          string  `json:"uf"`
	StateName        string  `json:"state"`
	StateLat         float64 `json:"state_lat"`
	StateLon         float64 `json:"state_lon"`
}

type MeiliAddress struct {
	StreetID         int32   `json:"street_id"`
	StreetName       string  `json:"street"`
	NeighborhoodName string  `json:"neighborhood"`
	NeighborhoodLat  float64 `json:"neighborhood_lat"`
	NeighborhoodLon  float64 `json:"neighborhood_lon"`
	CityName         string  `json:"city"`
	CityLat          float64 `json:"city_lat"`
	CityLon          float64 `json:"city_lon"`
	StateUf          string  `json:"uf"`
	StateName        string  `json:"state"`
	StateLat         float64 `json:"state_lat"`
	StateLon         float64 `json:"state_lon"`
	AddressID        int32   `json:"id"`
	Number           string  `json:"number"`
	Cep              string  `json:"cep"`
	Lat              float64 `json:"lat"`
	Lon              float64 `json:"lon"`
}
