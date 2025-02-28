package login

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

type Handler struct {
	service ServiceInterface
}

func NewHandler(service ServiceInterface) *Handler {
	return &Handler{service}
}

func (h *Handler) Login(e echo.Context) error {
	var request RequestLogin
	if err := e.Bind(&request); err != nil {
		return e.JSON(http.StatusBadRequest, err.Error())
	}

	result, err := h.service.Login(e.Request().Context(), request)
	if err != nil {
		return e.JSON(http.StatusInternalServerError, err.Error())
	}

	return e.JSON(http.StatusOK, result)
}
