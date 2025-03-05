package plans

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

type Handler struct {
	InterfaceService InterfaceService
}

func NewPlansHandler(InterfaceService InterfaceService) *Handler {
	return &Handler{InterfaceService}
}

// CreatePlanHandler godoc
// @Summary Create a Plan.
// @Description Create a Plan.
// @Tags Plans
// @Accept json
// @Produce json
// @Param request body CreatePlansRequest true "Plan Request"
// @Success 200 {object} PlansResponse "Plan Info"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /driver/create [post]
// @Security ApiKeyAuth
func (p *Handler) CreatePlanHandler(c echo.Context) error {
	var request CreatePlansRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	result, err := p.InterfaceService.CreatePlansService(c.Request().Context(), request)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

// UpdatePlanHandler godoc
// @Summary Update a Plan.
// @Description Update a Plan.
// @Tags Plans
// @Accept json
// @Produce json
// @Param user body UpdatePlansRequest true "Plan Request"
// @Success 200 {object} PlansResponse "Plan Info"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /driver/update [put]
// @Security ApiKeyAuth
func (p *Handler) UpdatePlanHandler(c echo.Context) error {
	var request UpdatePlansRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	result, err := p.InterfaceService.UpdatePlansService(c.Request().Context(), request)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}
