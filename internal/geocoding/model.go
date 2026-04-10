package geocoding

// RequestLocationDTO representa os dados de entrada para a consulta de localização (geocodificação)
type RequestLocationDTO struct {
	Origin      string `json:"origin"      binding:"required"`
	Destination string `json:"destination" binding:"required"`
	Costing     string `json:"costing"     binding:"required,oneof=auto bicycle pedestrian truck bus motorcycle"`
}

// ResponseLocationDTO representa o retorno consolidado com informações de endereço e coordenadas
type ResponseLocationDTO struct {
	Origin      LocationInfoAddress `json:"origin"`
	Destination LocationInfoAddress `json:"destination"`
	RoutesLink  RoutesLink          `json:"routes_link"`
	TripInfo    []TripInfo          `json:"trip_info"`
}

// LocationInfoAddress contém os detalhes geográficos de um ponto específico
type LocationInfoAddress struct {
	Address string `json:"address"` // Endereço formatado (RUA, BAIRRO, CIDADE/UF)
	Lat     string `json:"lat"`     // Latitude em string com 6 casas decimais
	Lon     string `json:"lon"`     // Longitude em string com 6 casas decimais
}

type RoutesLink struct {
	GoogleMapsLink string `json:"google_maps_link"`
	WazeLink       string `json:"waze_link"`
}

type TripInfo struct {
	Provider          string   `json:"provider"` // "OSRM" ou "VALHALLA"
	DistanceKm        float64  `json:"distance_km"`
	DurationMn        float64  `json:"duration_minutes"`
	Polyline          string   `json:"polyline"`
	PolylinePrecision int      `json:"polyline_precision"` // 5 para OSRM, 6 para VALHALLA
	Instructions      []string `json:"instructions"`
}
