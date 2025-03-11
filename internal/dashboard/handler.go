package dashboard

import (
	"fmt"
	"geolocation/internal/get_token"
	"github.com/labstack/echo/v4"
	"net/http"
)

type Handler struct {
	InterfaceService InterfaceService
}

func NewDashboardHandler(InterfaceService InterfaceService) *Handler {
	return &Handler{InterfaceService}
}

// GetDashboardHandler godoc
// @Summary Get Driver.
// @Description Get Driver.
// @Tags Drivers
// @Accept json
// @Produce json
// @Param id path string true "Driver id"
// @Success 200 {object} DriverResponse "Driver Info"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /driver/list [put]
// @Security ApiKeyAuth
func (p *Handler) GetDashboardHandler(c echo.Context) error {
	payload := get_token.GetUserPayloadToken(c)

	fmt.Println(payload.ID)
	result, err := p.InterfaceService.GetDashboardService(c.Request().Context(), payload.ID, payload.ProfileID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}
