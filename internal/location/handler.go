package location

import (
	"geolocation/internal/get_token"
	"geolocation/validation"
	"github.com/labstack/echo/v4"
	"net/http"
)

type Handler struct {
	InterfaceService InterfaceService
}

func NewLocationHandler(InterfaceService InterfaceService) *Handler {
	return &Handler{
		InterfaceService,
	}
}

func (h *Handler) CreateLocationHandler(c echo.Context) error {
	var request CreateLocationRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	payload := get_token.GetPayloadToken(c)
	arg := CreateLocationDTO{
		CreateLocationRequest: request,
		Payload:               payload,
	}
	result, err := h.InterfaceService.CreateLocationService(c.Request().Context(), arg)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

func (h *Handler) UpdateLocationHandler(c echo.Context) error {
	var request UpdateLocationRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	payload := get_token.GetPayloadToken(c)
	arg := UpdateLocationDTO{
		UpdateLocationRequest: request,
		Payload:               payload,
	}
	updated, err := h.InterfaceService.UpdateLocationService(c.Request().Context(), arg)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, updated)
}

func (h *Handler) DeleteLocationHandler(c echo.Context) error {
	idStr := c.Param("id")
	id, err := validation.ParseStringToInt64(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	payload := get_token.GetPayloadToken(c)
	if err := h.InterfaceService.DeleteLocationService(c.Request().Context(), id, payload); err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, "Success")
}

func (h *Handler) GetLocationHandler(c echo.Context) error {
	idStr := c.Param("providerId")
	id, err := validation.ParseStringToInt64(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	payload := get_token.GetPayloadToken(c)
	result, err := h.InterfaceService.GetLocationService(c.Request().Context(), id, payload)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}
