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
	payload := get_token.GetPayloadToken(e)
	id, err := validation.ParseStringToInt64(payload.UserID)
	if err != nil {
		return e.JSON(http.StatusBadRequest, err.Error())
	}

	result, err := h.InterfaceService.CalculateRoutes(e.Request().Context(), frontInfo, payloadPublic.ID, id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, echo.ErrNotFound) {
			statusCode = http.StatusNotFound
		}
		return e.JSON(statusCode, err.Error())
	}

	return e.JSON(http.StatusOK, result)
}
