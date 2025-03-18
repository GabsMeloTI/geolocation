package tractor_unit

import (
	"geolocation/internal/get_token"
	"geolocation/validation"
	"github.com/labstack/echo/v4"
	"net/http"
)

type Handler struct {
	InterfaceService InterfaceService
}

func NewTractorUnitsHandler(InterfaceService InterfaceService) *Handler {
	return &Handler{InterfaceService}
}

// CreateTractorUnitHandler godoc
// @Summary Create a TractorUnit.
// @Description Create a TractorUnit.
// @Tags TractorUnits
// @Accept json
// @Produce json
// @Param request body CreateTractorUnitRequest true "TractorUnit Request"
// @Success 200 {object} TractorUnitResponse "TractorUnit Info"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /tractor-unit/create [post]
// @Security ApiKeyAuth
func (p *Handler) CreateTractorUnitHandler(c echo.Context) error {
	var request CreateTractorUnitRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	payload := get_token.GetUserPayloadToken(c)
	data := CreateTractorUnitDto{
		CreateTractorUnitRequest: request,
		UserID:                   payload.ID,
	}

	result, err := p.InterfaceService.CreateTractorUnitService(c.Request().Context(), data)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

// UpdateTractorUnitHandler godoc
// @Summary Update a TractorUnit.
// @Description Update a TractorUnit.
// @Tags TractorUnits
// @Accept json
// @Produce json
// @Param user body UpdateTractorUnitRequest true "TractorUnit Request"
// @Success 200 {object} TractorUnitResponse "TractorUnit Info"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /tractor-unit/update [put]
// @Security ApiKeyAuth
func (p *Handler) UpdateTractorUnitHandler(c echo.Context) error {
	var request UpdateTractorUnitRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	payload := get_token.GetUserPayloadToken(c)
	data := UpdateTractorUnitDto{
		UpdateTractorUnitRequest: request,
		UserID:                   payload.ID,
	}

	result, err := p.InterfaceService.UpdateTractorUnitService(c.Request().Context(), data)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

// DeleteTractorUnitHandler godoc
// @Summary Delete TractorUnit.
// @Description Delete TractorUnit.
// @Tags TractorUnits
// @Accept json
// @Produce json
// @Param id path string true "TractorUnit id"
// @Success 200
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /tractor-unit/delete/{id} [put]
// @Security ApiKeyAuth
func (p *Handler) DeleteTractorUnitHandler(c echo.Context) error {
	idStr := c.Param("id")
	id, err := validation.ParseStringToInt64(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	payload := get_token.GetUserPayloadToken(c)
	err = p.InterfaceService.DeleteTractorUnitService(c.Request().Context(), id, payload.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, "Success")
}

// GetTractorUnitHandler godoc
// @Summary Get Tractor Unit.
// @Description Get Tractor Unit.
// @Tags TractorUnits
// @Accept json
// @Produce json
// @Param id path string true "TractorUnit id"
// @Success 200
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /tractor-unit/list [get]
// @Security ApiKeyAuth
func (p *Handler) GetTractorUnitHandler(c echo.Context) error {
	payload := get_token.GetUserPayloadToken(c)

	result, err := p.InterfaceService.GetTractorUnitService(c.Request().Context(), payload.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

// GetTractorUnitByIdHandler godoc
// @Summary Get Tractor Unit.
// @Description Get Tractor Unit.
// @Tags TractorUnits
// @Accept json
// @Produce json
// @Param id path string true "TractorUnit id"
// @Success 200
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /tractor-unit/list/{id} [get]
// @Security ApiKeyAuth
func (p *Handler) GetTractorUnitByIdHandler(c echo.Context) error {
	idStr := c.Param("id")
	id, err := validation.ParseStringToInt64(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	result, err := p.InterfaceService.GetTractorUnitByIdService(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}
