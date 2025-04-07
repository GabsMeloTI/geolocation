package trailer

import (
	"geolocation/internal/get_token"
	"geolocation/validation"
	"github.com/labstack/echo/v4"
	"net/http"
)

type Handler struct {
	InterfaceService InterfaceService
}

func NewTrailersHandler(InterfaceService InterfaceService) *Handler {
	return &Handler{InterfaceService}
}

// CreateTrailerHandler godoc
// @Summary Create a Trailer.
// @Description Create a Trailer.
// @Tags Trailer
// @Accept json
// @Produce json
// @Param request body CreateTrailerRequest true "Trailer Request"
// @Success 200 {object} TrailerResponse "Trailer Info"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /trailer/create [post]
// @Security ApiKeyAuth
func (p *Handler) CreateTrailerHandler(c echo.Context) error {
	var request CreateTrailerRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	payload := get_token.GetUserPayloadToken(c)
	data := CreateTrailerDto{
		CreateTrailerRequest: request,
		UserID:               payload.ID,
	}

	result, err := p.InterfaceService.CreateTrailerService(c.Request().Context(), data)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

// UpdateTrailerHandler godoc
// @Summary Update a Trailer.
// @Description Update a Trailer.
// @Tags Trailer
// @Accept json
// @Produce json
// @Param user body UpdateTrailerRequest true "Trailer Request"
// @Success 200 {object} TrailerResponse "Trailer Info"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /trailer/update [put]
// @Security ApiKeyAuth
func (p *Handler) UpdateTrailerHandler(c echo.Context) error {
	var request UpdateTrailerRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	payload := get_token.GetUserPayloadToken(c)
	data := UpdateTrailerDto{
		UpdateTrailerRequest: request,
		UserID:               payload.ID,
	}

	result, err := p.InterfaceService.UpdateTrailerService(c.Request().Context(), data)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

// DeleteTrailerHandler godoc
// @Summary Delete Trailer.
// @Description Delete Trailer.
// @Tags Trailer
// @Accept json
// @Produce json
// @Param id path string true "Trailer id"
// @Success 200
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /trailer/delete/{id} [put]
// @Security ApiKeyAuth
func (p *Handler) DeleteTrailerHandler(c echo.Context) error {
	idStr := c.Param("id")
	id, err := validation.ParseStringToInt64(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	payload := get_token.GetUserPayloadToken(c)
	err = p.InterfaceService.DeleteTrailerService(c.Request().Context(), id, payload.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, "Success")
}

// GetTrailerHandler godoc
// @Summary Get Trailer.
// @Description Get Trailer.
// @Tags Trailer
// @Accept json
// @Produce json
// @Param id path string true "Trailer id"
// @Success 200
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /trailer/list [put]
// @Security ApiKeyAuth
func (p *Handler) GetTrailerHandler(c echo.Context) error {
	payload := get_token.GetUserPayloadToken(c)

	result, err := p.InterfaceService.GetTrailerService(c.Request().Context(), payload.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

// GetTrailerByIdHandler godoc
// @Summary Get Tractor Unit.
// @Description Get Tractor Unit.
// @Tags Trailer
// @Accept json
// @Produce json
// @Param id path string true "Trailer id"
// @Success 200
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /tractor-unit/list/{id} [get]
// @Security ApiKeyAuth
func (p *Handler) GetTrailerByIdHandler(c echo.Context) error {
	idStr := c.Param("id")
	id, err := validation.ParseStringToInt64(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	result, err := p.InterfaceService.GetTrailerByIdService(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}
