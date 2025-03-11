package advertisement

import (
	"database/sql"
	"geolocation/internal/get_token"
	"geolocation/validation"
	"github.com/labstack/echo/v4"
	"net/http"
)

type Handler struct {
	InterfaceService InterfaceService
}

func NewAdvertisementHandler(InterfaceService InterfaceService) *Handler {
	return &Handler{InterfaceService}
}

// CreateAdvertisementHandler godoc
// @Summary Create a Advertisement.
// @Description Create a Advertisement.
// @Tags Advertisement
// @Accept json
// @Produce json
// @Param request body CreateAdvertisementRequest true "Advertisement Request"
// @Success 200 {object} AdvertisementResponse "Advertisement Info"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /advertisement/create [post]
// @Security ApiKeyAuth
func (p *Handler) CreateAdvertisementHandler(c echo.Context) error {
	var request CreateAdvertisementRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	payload := get_token.GetUserPayloadToken(c)
	data := CreateAdvertisementDto{
		CreateAdvertisementRequest: request,
		UserID:                     payload.ID,
		CreatedWho:                 payload.Name,
	}

	result, err := p.InterfaceService.CreateAdvertisementService(c.Request().Context(), data, payload.ProfileID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

func (p *Handler) UpdatedAdvertisementFinishedCreate(c echo.Context) error {
	var request UpdatedAdvertisementFinishedCreate
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	payload := get_token.GetUserPayloadToken(c)
	result, err := p.InterfaceService.UpdatedAdvertisementFinishedCreate(c.Request().Context(), request, payload.ProfileID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

// UpdateAdvertisementHandler godoc
// @Summary Update a Advertisement.
// @Description Update a Advertisement.
// @Tags Advertisement
// @Accept json
// @Produce json
// @Param user body UpdateAdvertisementRequest true "Advertisement Request"
// @Success 200 {object} AdvertisementResponse "Advertisement Info"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /advertisement/update [put]
// @Security ApiKeyAuth
func (p *Handler) UpdateAdvertisementHandler(c echo.Context) error {
	var request UpdateAdvertisementRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	payload := get_token.GetUserPayloadToken(c)
	data := UpdateAdvertisementDto{
		UpdateAdvertisementRequest: request,
		UserID:                     payload.ID,
		UpdatedWho: sql.NullString{
			String: payload.Name,
			Valid:  true,
		},
	}

	result, err := p.InterfaceService.UpdateAdvertisementService(c.Request().Context(), data, payload.ProfileID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

// DeleteAdvertisementHandler godoc
// @Summary Delete Advertisement.
// @Description Delete Advertisement.
// @Tags Advertisement
// @Accept json
// @Produce json
// @Param id path string true "Advertisement id"
// @Success 200
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /advertisement/delete/{id} [put]
// @Security ApiKeyAuth
func (p *Handler) DeleteAdvertisementHandler(c echo.Context) error {
	idStr := c.Param("id")
	id, err := validation.ParseStringToInt64(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	payload := get_token.GetUserPayloadToken(c)
	data := DeleteAdvertisementRequest{
		ID:     id,
		UserID: payload.ID,
		UpdatedWho: sql.NullString{
			String: payload.Name,
			Valid:  true,
		},
	}
	err = p.InterfaceService.DeleteAdvertisementService(c.Request().Context(), data)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, "Success")
}

// GetAllAdvertisementHandler godoc
// @Summary Get All Advertisement
// @Description Retrieve all Advertisement
// @Tags Advertisement
// @Accept json
// @Produce json
// @Success 200 {object} []AdvertisementResponseAll "List of Advertisement"
// @Failure 500 {string} string "Internal Server Error"
// @Router /advertisement/list [get]
// @Security ApiKeyAuth
func (p *Handler) GetAllAdvertisementHandler(c echo.Context) error {
	result, err := p.InterfaceService.GetAllAdvertisementUser(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

// GetAllAdvertisementPublicHandler godoc
// @Summary Get All Advertisement
// @Description Retrieve all Advertisement
// @Tags Advertisement
// @Accept json
// @Produce json
// @Success 200 {object} []AdvertisementResponseNoUser "List of Advertisement"
// @Failure 500 {string} string "Internal Server Error"
// @Router /public/advertisement/list [get]
func (p *Handler) GetAllAdvertisementPublicHandler(c echo.Context) error {
	result, err := p.InterfaceService.GetAllAdvertisementPublic(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}
