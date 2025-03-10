package plans

import (
	"fmt"
	"geolocation/infra/token"
	"geolocation/internal/get_token"
	"github.com/labstack/echo/v4"
	"net/http"
)

type Handler struct {
	InterfaceService InterfaceService
	tokenMaker       token.PasetoMaker
}

func NewUserPlanHandler(InterfaceService InterfaceService) *Handler {
	return &Handler{
		InterfaceService,
		token.PasetoMaker{},
	}
}

// CreateUserPlanHandler godoc
// @Summary Create a User Plan.
// @Description Assigns a user to a selected plan.
// @Tags Users
// @Accept json
// @Produce json
// @Param id path int true "Plan ID"
// @Success 200 {object} UserPlanResponse "User Plan Info"
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

func (p *Handler) GetTokenUserHandler(c echo.Context) error {
	payload := get_token.GetUserPayloadToken(c)
	fmt.Println(payload.ID)

	newToken, err := p.InterfaceService.GenerateUserToken(payload.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
	}

	return c.JSON(http.StatusOK, map[string]string{"token": newToken})
}
