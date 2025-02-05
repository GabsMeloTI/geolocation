package routes

import (
	"errors"
	"geolocation/validation"
	"github.com/labstack/echo/v4"
	"net/http"
)

type Handler struct {
	InterfaceService InterfaceService
}

func NewRoutesHandler(InterfaceService InterfaceService) *Handler {
	return &Handler{InterfaceService}
}

func (h *Handler) CheckRouteTolls(e echo.Context) error {
	var frontInfo FrontInfo
	if err := e.Bind(&frontInfo); err != nil {
		return e.JSON(http.StatusBadRequest, err.Error())
	}

	err := validation.Validate(frontInfo)
	if err != nil {
		return e.JSON(http.StatusBadRequest, err.Error())
	}

	result, err := h.InterfaceService.CheckRouteTolls(e.Request().Context(), frontInfo)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, echo.ErrNotFound) {
			statusCode = http.StatusNotFound
		}
		return e.JSON(statusCode, err.Error())
	}

	return e.JSON(http.StatusOK, result)
}

func (h *Handler) AddSavedRoutesFavorite(e echo.Context) error {
	strId := e.QueryParam("id")
	id, err := validation.ParseStringToInt32(strId)
	if err != nil {
		return e.JSON(http.StatusBadRequest, err.Error())
	}

	if err := e.Bind(&id); err != nil {
		return e.JSON(http.StatusBadRequest, err.Error())
	}

	err = h.InterfaceService.AddSavedRoutesFavorite(e.Request().Context(), id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, echo.ErrNotFound) {
			statusCode = http.StatusNotFound
		}
		return e.JSON(statusCode, err.Error())
	}

	return e.JSON(http.StatusOK, "Successfully added saved routes favorite")
}
