package plans

import (
	"geolocation/internal/get_token"
	"github.com/labstack/echo/v4"
	"net/http"
)

type Handler struct {
	InterfaceService InterfaceService
}

func NewUserPlanHandler(InterfaceService InterfaceService) *Handler {
	return &Handler{InterfaceService}
}

// CreateUserPlanHandler godoc
// @Summary Create a User Plan.
// @Description Assigns a user to a selected plan.
// @Tags User Plans
// @Accept json
// @Produce json
// @Param id path int true "Plan ID"
// @Success 200 {object} PlansResponse "User Plan Info"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /user/plan [post]
// @Security ApiKeyAuth
func (p *Handler) CreateUserPlanHandler(c echo.Context) error {
	var request CreateUserPlanRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	payload := get_token.GetUserPayloadToken(c)
	request.IDUser = payload.ID
	result, err := p.InterfaceService.CreateUserPlanService(c.Request().Context(), request)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}
