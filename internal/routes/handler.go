package routes

import (
	"errors"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
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

	if err := validateVehicleTypeAndAxles(frontInfo); err != nil {
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

func (h *Handler) GetExactPlaceHandler(e echo.Context) error {
	latStr := e.QueryParam("lat")
	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		return e.JSON(http.StatusBadRequest, err.Error())
	}
	lngStr := e.QueryParam("lng")
	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		return e.JSON(http.StatusBadRequest, err.Error())
	}

	arg := PlaceRequest{
		Latitude:  lat,
		Longitude: lng,
	}
	response, err := h.InterfaceService.GetExactPlace(e.Request().Context(), arg)
	if err != nil {
		return e.JSON(http.StatusInternalServerError, err.Error())
	}

	return e.JSON(http.StatusOK, response)
}
