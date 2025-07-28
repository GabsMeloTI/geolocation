package new_routes

import (
	"errors"
	"geolocation/internal/get_token"
	"geolocation/validation"
	"github.com/labstack/echo/v4"
	"net/http"
)

type Handler struct {
	InterfaceService InterfaceService
}

func NewRoutesNewHandler(InterfaceService InterfaceService) *Handler {
	return &Handler{InterfaceService}
}

// CalculateRoutes godoc
// @Summary Calcular rotas possíveis.
// @Description Calcula as melhores opções de rota a partir de uma origem e destino.
// @Description
// @Description Campos esperados no body:
// @Description - origin: "São Paulo" (Local de chegada)
// @Description - destination: "Salvador" (Local de saída)
// @Description - axles: 2 (Quantidade de eixos, possível somente: 2, 4, 6, 8, 9)
// @Description - consumptionCity: 20 (Consumo de combustível na cidade)
// @Description - consumptionHwy: 22 (Consumo de combustível na estrada)
// @Description - price: 6.20 (Preço da gasolina)
// @Description - waypoints: ["Rio de Janeiro", "Vitória da Conquista"] (Lista de pontos de parada)
// @Description - favorite: true (Se deseja favoritar essa rota)
// @Description - type: "Auto" (Tipo do automóvel, possíveis: Truck, Bus, Auto, Motorcycle)
// @Description - typeRoute: "eficiente" (Caso queira apenas uma rota: eficiente, rápida ou barata)
// @Description - route_options: {
// @Description       include_fuel_stations: false, (traz postos de gasolina)
// @Description       include_route_map: false, (traz rotograma da rota)
// @Description       include_toll_costs: false, (traz pedágios e os custos gerais)
// @Description       include_weigh_stations: false, (traz balanças)
// @Description       include_freight_calc: false, (traz frestes, segundo a ANTT calculados)
// @Description       include_polyline: false (traz polyline para renderizar em mapas)
// @Description   } (Opções adicionais para a rota)
// @Tags Routes
// @Accept json
// @Produce json
// @Param request body FrontInfo true "Requisição para cálculo de rota"
// @Success 200 {object} FinalOutput "Informações calculadas da rota"
// @Failure 400 {string} string "Requisição Inválida"
// @Failure 404 {string} string "Não Encontrado"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /check-route-tolls-easy [post]
// @Security ApiKeyAuth
func (h *Handler) CalculateRoutes(e echo.Context) error {
	var frontInfo FrontInfo
	if err := e.Bind(&frontInfo); err != nil {
		return e.JSON(http.StatusBadRequest, err.Error())
	}

	err := validation.Validate(frontInfo)
	if err != nil {
		return e.JSON(http.StatusBadRequest, err.Error())
	}

	payloadPublic := get_token.GetPublicPayloadToken(e)
	payload := get_token.GetUserPayloadToken(e)
	result, err := h.InterfaceService.CalculateRoutes(e.Request().Context(), frontInfo, payloadPublic.ID, payload.ID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, echo.ErrNotFound) {
			statusCode = http.StatusNotFound
		}
		return e.JSON(statusCode, err.Error())
	}

	return e.JSON(http.StatusOK, result)
}

// CalculateRoutesWithCoordinate godoc
// @Summary Calcular rotas com base em coordenadas.
// @Description Calcula as melhores opções de rota a partir de uma latitude e longitude de origem e destino.
// @Description
// @Description Campos esperados no body:
// @Description - origin_lat: \"-25.550520\" (Latitude do local de saída)
// @Description - origin_lng: \"-48.633309\" (Longitude do local de saída)
// @Description - destination_lat: \"-31.0368176\" (Latitude do local de chegada)
// @Description - destination_lng: \"-51.0368176\" (Longitude do local de chegada)
// @Description - axles: 2 (Quantidade de eixos, possível somente: 2, 4, 6, 8, 9)
// @Description - consumptionCity: 20 (Consumo de combustível na cidade)
// @Description - consumptionHwy: 22 (Consumo de combustível na estrada)
// @Description - price: 6.20 (Preço da gasolina)
// @Description - waypoints: [{\"lat\": \"-23.223701\",\"lng\": \"-45.900907\"},{\"lat\": \"-22.755611\",\"lng\": \"-44.168869\"}] (Lista de pontos de parada, definida pelas coordenadas)
// @Description - favorite: true (Se deseja favoritar essa rota)
// @Description - type: \"Auto\" (Tipo do automóvel, possível: Truck, Bus, Auto, Motorcycle)
// @Description - typeRoute: \"eficiente\" (Caso queira apenas uma rota: eficiente, rápida ou barata)
// @Description - route_options: {
// @Description       include_fuel_stations: false, (traz postos de gasolina)
// @Description       include_route_map: false, (traz rotograma da rota)
// @Description       include_toll_costs: false, (traz pedágios e os custos gerais)
// @Description       include_weigh_stations: false, (traz balanças)
// @Description       include_freight_calc: false, (traz frestes, segundo a ANTT calculados)
// @Description       include_polyline: false (traz polyline para renderizar em mapas)
// @Description   } (Opções adicionais para a rota)
// @Tags Routes
// @Accept json
// @Produce json
// @Param request body FrontInfoCoordinate true "Requisição para cálculo de rota por coordenadas"
// @Success 200 {object} FinalOutput "Informações calculadas da rota"
// @Failure 400 {string} string "Requisição Inválida"
// @Failure 404 {string} string "Não Encontrado"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /check-route-tolls-coordinate [post]
// @Security ApiKeyAuth
func (h *Handler) CalculateRoutesWithCoordinate(e echo.Context) error {
	var frontInfo FrontInfoCoordinate
	if err := e.Bind(&frontInfo); err != nil {
		return e.JSON(http.StatusBadRequest, err.Error())
	}

	err := validation.Validate(frontInfo)
	if err != nil {
		return e.JSON(http.StatusBadRequest, err.Error())
	}

	payloadPublic := get_token.GetPublicPayloadToken(e)
	payload := get_token.GetUserPayloadToken(e)
	result, err := h.InterfaceService.CalculateRoutesWithCoordinate(e.Request().Context(), frontInfo, payloadPublic.ID, payload.ID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, echo.ErrNotFound) {
			statusCode = http.StatusNotFound
		}
		return e.JSON(statusCode, err.Error())
	}

	return e.JSON(http.StatusOK, result)
}

// CalculateRoutesWithCEP godoc
// @Summary Calcular rotas com base em CEP.
// @Description Calcula as melhores opções de rota a partir dos CEPs de origem e destino.
// @Description
// @Description Campos esperados no body:
// @Description   - origin_cep: \"01001000\" (CEP de origem)
// @Description   - destination_cep: \"20040002\" (CEP de destino)
// @Description   - consumptionCity: 20         (Consumo de combustível na cidade, em km/l)
// @Description   - consumptionHwy: 22         (Consumo de combustível na estrada, em km/l)
// @Description   - price: 6.20                (Preço do combustível em BRL)
// @Description   - axles: 2                  (Quantidade de eixos: 2, 4, 6, 8, 9)
// @Description   - waypoints: [\"01310940\",\"20050013\"] (Lista de CEPs para pontos de parada)
// @Description   - public_or_private: \"public\" | \"private\" (Define se conta na cota pública ou privada)
// @Description   - favorite: true             (Se deseja favoritar essa rota)
// @Description   - type: \"Auto\"              (Tipo de veículo: Truck, Bus, Auto, Motorcycle)
// @Description   - typeRoute: \"eficiente\"    (Rota específica: eficiente, rápida ou barata)
// @Description   - route_options: {
// @Description         include_fuel_stations: false,   (retorna postos de gasolina)
// @Description         include_route_map: false,        (retorna rotograma da rota)
// @Description         include_toll_costs: false,       (retorna pedágios e custos gerais)
// @Description         include_weigh_stations: false,   (retorna balanças)
// @Description         include_freight_calc: false,     (retorna cálculo de frete ANTT)
// @Description         include_polyline: false          (retorna polyline para mapas)
// @Description     } (Opções adicionais para a rota)
// @Tags Routes
// @Accept json
// @Produce json
// @Param request body FrontInfoCEP true "Requisição para cálculo de rota por CEP"
// @Success 200 {object} FinalOutput "Informações calculadas da rota"
// @Failure 400 {string} string "Requisição Inválida"
// @Failure 404 {string} string "Não Encontrado"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /check-route-tolls-cep [post]
// @Security ApiKeyAuth
func (h *Handler) CalculateRoutesWithCEP(e echo.Context) error {
	var frontInfo FrontInfoCEP
	if err := e.Bind(&frontInfo); err != nil {
		return e.JSON(http.StatusBadRequest, err.Error())
	}

	err := validation.Validate(frontInfo)
	if err != nil {
		return e.JSON(http.StatusBadRequest, err.Error())
	}

	payloadPublic := get_token.GetPublicPayloadToken(e)
	payloadSimp := get_token.GetPayloadToken(e)
	payload := get_token.GetUserPayloadToken(e)
	result, err := h.InterfaceService.CalculateRoutesWithCEP(e.Request().Context(), frontInfo, payloadPublic.ID, payload.ID, payloadSimp)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, echo.ErrNotFound) {
			statusCode = http.StatusNotFound
		}
		return e.JSON(statusCode, err.Error())
	}

	return e.JSON(http.StatusOK, result)
}

func (h *Handler) CalculateRoutesCEP(e echo.Context) error {
	var frontInfo FrontInfoCEPRequest
	if err := e.Bind(&frontInfo); err != nil {
		return e.JSON(http.StatusBadRequest, err.Error())
	}

	result, err := h.InterfaceService.CalculateDistancesBetweenPoints(e.Request().Context(), frontInfo)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, echo.ErrNotFound) {
			statusCode = http.StatusNotFound
		}
		return e.JSON(statusCode, err.Error())
	}

	return e.JSON(http.StatusOK, result)
}

// GetSimpleRoute godoc
// @Summary Get simple route information.
// @Description Retrieves a simple route with distance and duration.
// @Tags Routes
// @Accept json
// @Produce json
// @Param request body SimpleRouteRequest true "Route calculation request"
// @Success 200 {object} SimpleRouteResponse "Route details"
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /route/simple [get]
// @Security ApiKeyAuth
func (h *Handler) GetSimpleRoute(e echo.Context) error {
	var request SimpleRouteRequest
	if err := e.Bind(&request); err != nil {
		return e.JSON(http.StatusBadRequest, err.Error())
	}

	result, err := h.InterfaceService.GetSimpleRoute(request)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, echo.ErrNotFound) {
			statusCode = http.StatusNotFound
		}
		return e.JSON(statusCode, err.Error())
	}

	return e.JSON(http.StatusOK, result)
}

// GetFavoriteRouteHandler godoc
// @Summary Get FavoriteRoute.
// @Description Get FavoriteRoute.
// @Tags Rotas Favoritas
// @Accept json
// @Produce json
// @Param id path string true "FavoriteRoute id"
// @Success 200 {string} json.RawMessage "Favorite Route Info"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /route/favorite/list [get]
// @Security ApiKeyAuth
func (h *Handler) GetFavoriteRouteHandler(c echo.Context) error {
	payload := get_token.GetUserPayloadToken(c)

	result, err := h.InterfaceService.GetFavoriteRouteService(c.Request().Context(), payload.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

// RemoveFavoriteRouteHandler godoc
// @Summary Get FavoriteRoute.
// @Description Get FavoriteRoute.
// @Tags Rotas Favoritas
// @Accept json
// @Produce json
// @Param id path string true "FavoriteRoute id"
// @Success 200
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /route/favorite/remove/:id [put]
// @Security ApiKeyAuth
func (h *Handler) RemoveFavoriteRouteHandler(c echo.Context) error {
	idStr := c.Param("id")
	id, err := validation.ParseStringToInt64(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	payload := get_token.GetUserPayloadToken(c)
	err = h.InterfaceService.RemoveFavoriteRouteService(c.Request().Context(), id, payload.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, "success")
}
