package login

import (
	"geolocation/validation"
	"github.com/labstack/echo/v4"
	"net/http"
	"net/mail"
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

func (h *Handler) CreateUser(e echo.Context) error {
	var request RequestCreateUser
	if err := e.Bind(&request); err != nil {
		return e.JSON(http.StatusBadRequest, err.Error())
	}

	if err := validation.Validate(request); err != nil {
		return e.JSON(http.StatusBadRequest, err.Error())
	}

	_, err := mail.ParseAddress(request.Email)
	if err != nil {
		return e.JSON(http.StatusBadRequest, "invalid email address")
	}

	if request.Token != "" {
		if ok := validation.ValidatePassword(request.Password); !ok {
			return e.JSON(http.StatusBadRequest, "invalid password")
		}

		if ok := request.Password == request.ConfirmPassword; !ok {
			return e.JSON(http.StatusBadRequest, "password and confirm password are different")
		}
	}

	result, err := h.service.CreateUser(e.Request().Context(), request)
	if err != nil {
		return e.JSON(http.StatusInternalServerError, err.Error())
	}

	return e.JSON(http.StatusOK, result)
}
