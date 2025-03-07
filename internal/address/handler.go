package address

import (
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
// @Description Encontra endereço por busca, pode ser 1. CEP, 2.Latidude, Longitude ou 3.Endereço (Rua, bairro, número).
// @Tags Address
// @Accept json
// @Produce json
// @Param q query string true "Address Query"
// @Success 200 {object} AddressResponse[] "Address Info"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /address/find [get]
// @Security ApiKeyAuth
func (p *Handler) FindAddressByQueryHandler(c echo.Context) error {
	q := c.QueryParam("q")
	if q == "" {
		return c.JSON(http.StatusBadRequest, "Query parameter 'q' is required")
	}
	result, err := p.InterfaceService.FindAddressesByQueryService(c.Request().Context(), q)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, result)
}
