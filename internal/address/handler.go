package address

import (
	"geolocation/validation"
	"github.com/labstack/echo/v4"
	"net/http"
)

type Handler struct {
	InterfaceService InterfaceService
}

func NewAddressHandler(InterfaceService InterfaceService) *Handler {
	return &Handler{InterfaceService}
}

// FindAddressByQueryHandler godoc
// @Summary Find Address By Query
// @Description Find address by search, it can be 1. And, 2. Latitude, Longitude or 3. Address (Street, neighborhood, number).
// @Tags Address
// @Accept json
// @Produce json
// @Param q query string true "Address Query"
// @Success 200 {object} AddressResponse[] "Address Info"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /address/find [get]
// @Security ApiKeyAuth
func (h *Handler) FindAddressByQueryHandler(c echo.Context) error {
	q := c.QueryParam("q")
	if q == "" {
		return c.JSON(http.StatusBadRequest, "Query parameter 'q' is required")
	}
	result, err := h.InterfaceService.FindAddressesByQueryService(c.Request().Context(), q)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, result)
}

// FindAddressByCEPHandler godoc
// @Summary Find Address By CEP
// @Description Finds address by zip code, returns type based on repetitions found
// @Tags Address
// @Accept json
// @Produce json
// @Param CEP path string true "cep"
// @Success 200 {object} AddressCEPResponse[] "Address Info"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /address/find/{cep} [get]
// @Security ApiKeyAuth
func (h *Handler) FindAddressByCEPHandler(c echo.Context) error {
	cep := c.Param("cep")
	result, err := h.InterfaceService.FindAddressesByCEPService(c.Request().Context(), cep)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, result)
}

// FindStateAll godoc
// @Summary Find All States
// @Description Returns all available states.
// @Tags Address
// @Accept json
// @Produce json
// @Success 200 {object} StateResponse[] "List of States"
// @Failure 500 {string} string "Internal Server Error"
// @Router /address/state [get]
// @Security ApiKeyAuth
func (h *Handler) FindStateAll(c echo.Context) error {
	result, err := h.InterfaceService.FindStateAll(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, result)
}

// FindCityAll godoc
// @Summary Find All Cities by State ID
// @Description Returns all cities in a specific state by their ID.
// @Tags Address
// @Accept json
// @Produce json
// @Param idState path int32 true "State ID"
// @Success 200 {object} CityResponse[] "List of Cities"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /address/city/{idState} [get]
// @Security ApiKeyAuth
func (h *Handler) FindCityAll(c echo.Context) error {
	idStr := c.Param("idState")
	id, err := validation.ParseStringToInt32(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	result, err := h.InterfaceService.FindCityAll(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, result)
}
