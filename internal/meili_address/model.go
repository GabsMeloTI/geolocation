package meiliaddress

type MeiliAddress struct {
	StreetName       string        `json:"street"`
	NeighborhoodName string        `json:"neighborhood"`
	NeighborhoodLat  NullableFloat `json:"neighborhood_lat"`
	NeighborhoodLon  NullableFloat `json:"neighborhood_lon"`
	CityName         string        `json:"city"`
	CityLat          NullableFloat `json:"city_lat"`
	CityLon          NullableFloat `json:"city_lon"`
	StateUf          string        `json:"uf"`
	StateName        string        `json:"state"`
	StateLat         NullableFloat `json:"state_lat"`
	StateLon         NullableFloat `json:"state_lon"`
	AddressID        int32         `json:"id"`
	Number           string        `json:"number"`
	Cep              string        `json:"cep"`
	Lat              NullableFloat `json:"lat"`
	Lon              NullableFloat `json:"lon"`
}

type NullableFloat struct {
	Float64 float64 `json:"Float64"`
	Valid   bool    `json:"Valid"`
}
