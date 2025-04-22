package dashboard

import (
	"geolocation/internal/get_token"
	"github.com/labstack/echo/v4"
	"net/http"
	"time"
)

type Handler struct {
	InterfaceService InterfaceService
}

func NewDashboardHandler(InterfaceService InterfaceService) *Handler {
	return &Handler{InterfaceService}
}

// GetDashboardHandler godoc
// @Summary Obter Dashboard.
// @Description Recupera as informações do Dashboard.
// @Tags Dashboard
// @Accept json
// @Produce json
// @Param id path string true "ID do Dashboard"
// @Param start query string false "Data de início (formato YYYY-MM-DD)"
// @Param end query string false "Data de fim (formato YYYY-MM-DD)"
// @Success 200 {object} Response "Informações do Dashboard"
// @Failure 400 {string} string "Requisição Inválida"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /dashboard/list [get]
// @Security ApiKeyAuth
func (p *Handler) GetDashboardHandler(c echo.Context) error {
	startStr := c.QueryParam("start")
	endStr := c.QueryParam("end")

	var startDate, endDate *time.Time
	if startStr != "" {
		parsedStart, err := time.Parse("2006-01-02", startStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, "Formato de data de início inválido. Use YYYY-MM-DD")
		}
		startDate = &parsedStart
	}
	if endStr != "" {
		parsedEnd, err := time.Parse("2006-01-02", endStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, "Formato de data de fim inválido. Use YYYY-MM-DD")
		}
		endDate = &parsedEnd
	}

	payload := get_token.GetUserPayloadToken(c)
	result, err := p.InterfaceService.GetDashboardService(c.Request().Context(), payload.ID, payload.ProfileID, startDate, endDate)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}
