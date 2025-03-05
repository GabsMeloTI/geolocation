package routes

import (
	"errors"
	"geolocation/internal/get_token"
	"geolocation/validation"
	"github.com/labstack/echo/v4"
	"net/http"
)

type Handler struct {
	InterfaceService InterfaceService
}

func NewRoutesNewHandler(InterfaceService InterfaceService) *Handler {
	return &Handler{InterfaceService}
}

func (h *Handler) CalculateRoutes(e echo.Context) error {
	var frontInfo FrontInfo
	if err := e.Bind(&frontInfo); err != nil {
		return e.JSON(http.StatusBadRequest, err.Error())
	}

	err := validation.Validate(frontInfo)
	if err != nil {
		return e.JSON(http.StatusBadRequest, err.Error())
	}

	payloadPublic := get_token.GetPublicPayloadToken(e)
	payload := get_token.GetUserPayloadToken(e)
	result, err := h.InterfaceService.CalculateRoutes(e.Request().Context(), frontInfo, payloadPublic.ID, payload.ID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, echo.ErrNotFound) {
			statusCode = http.StatusNotFound
		}
		return e.JSON(statusCode, err.Error())
	}

	return e.JSON(http.StatusOK, result)
}

// GetFavoriteRouteHandler godoc
// @Summary Get FavoriteRoute.
// @Description Get FavoriteRoute.
// @Tags FavoriteRoutes
// @Accept json
// @Produce json
// @Param id path string true "FavoriteRoute id"
// @Success 200
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /route/favorite/list [get]
// @Security ApiKeyAuth
func (h *Handler) GetFavoriteRouteHandler(c echo.Context) error {
	payload := get_token.GetUserPayloadToken(c)

	result, err := h.InterfaceService.GetFavoriteRouteService(c.Request().Context(), payload.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

// RemoveFavoriteRouteHandler godoc
// @Summary Get FavoriteRoute.
// @Description Get FavoriteRoute.
// @Tags FavoriteRoutes
// @Accept json
// @Produce json
// @Param id path string true "FavoriteRoute id"
// @Success 200
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /route/favorite/remove/:id [get]
// @Security ApiKeyAuth
func (h *Handler) RemoveFavoriteRouteHandler(c echo.Context) error {
	idStr := c.Param("id")
	id, err := validation.ParseStringToInt64(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	payload := get_token.GetUserPayloadToken(c)
	err = h.InterfaceService.RemoveFavoriteRouteService(c.Request().Context(), id, payload.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, "success")
}
