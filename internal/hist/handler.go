package hist

import (
	"errors"
	"geolocation/internal/get_token"
	"github.com/labstack/echo/v4"
	"net/http"
)

type Handler struct {
	InterfaceService InterfaceService
}

func NewHistHandler(InterfaceService InterfaceService) *Handler {
	return &Handler{InterfaceService}
}

func (h *Handler) GetPublicToken(e echo.Context) error {
	ip := e.Param("ip")
	if err := e.Bind(&ip); err != nil {
		return e.JSON(http.StatusBadRequest, err.Error())
	}

	payload := get_token.GetPublicPayloadToken(e)
	data := Request{
		ID:             payload.ID,
		IP:             payload.IP,
		NumberRequests: payload.NumberRequests,
		Valid:          payload.Valid,
		ExpiredAt:      payload.ExpiredAt,
	}
	result, err := h.InterfaceService.GetPublicToken(e.Request().Context(), ip, data)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, echo.ErrNotFound) {
			statusCode = http.StatusNotFound
		}
		return e.JSON(statusCode, err.Error())
	}

	return e.JSON(http.StatusOK, result)
}
