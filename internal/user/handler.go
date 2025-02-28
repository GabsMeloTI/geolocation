package user

import (
	"fmt"
	"geolocation/internal/get_token"
	"geolocation/pkg/sso"
	"geolocation/validation"
	"github.com/labstack/echo/v4"
	"net/http"
	"net/mail"
	"strings"
)

type Handler struct {
	InterfaceService InterfaceService
	GoogleClientId   string
}

func NewUserHandler(InterfaceService InterfaceService, GoogleClientId string) *Handler {
	return &Handler{InterfaceService, GoogleClientId}
}

func (h *Handler) CreateUser(c echo.Context) error {
	var req CreateUserRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	if err := validation.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	if ok := validation.ValidatePassword(req.Password); !ok {
		return c.JSON(http.StatusBadRequest, "invalid password")
	}

	if ok := req.Password == req.ConfirmPassword; !ok {
		return c.JSON(http.StatusBadRequest, "password and confirm password are different")
	}

	_, err := mail.ParseAddress(req.Email)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "invalid email address")
	}

	if ok := validation.ValidatePhone(req.Phone); !ok {
		return c.JSON(http.StatusBadRequest, "invalid phone number")
	}

	if req.Provider == "google" {
		authorization := c.Request().Header.Get("Authorization")

		token := strings.Replace(authorization, "Bearer ", "", 1)

		payload, err := sso.ValidateGoogleToken(token)

		if err != nil {
			return c.JSON(http.StatusUnauthorized, err.Error())
		}

		if h.GoogleClientId != payload.Audience {
			fmt.Println("invalid client")
		}

	}

	if !validation.ValidatePassword(req.Password) {
		return c.JSON(http.StatusBadRequest, "invalid password")
	}

	res, err := h.InterfaceService.CreateUserService(c.Request().Context(), req)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, res)
}
func (h *Handler) UserLogin(c echo.Context) error {
	var req LoginRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	if req.Provider == "google" {
		authorization := c.Request().Header.Get("Authorization")
		token := strings.Replace(authorization, "Bearer ", "", 1)

		payload, err := sso.ValidateGoogleToken(token)

		if err != nil {
			return c.JSON(http.StatusUnauthorized, err.Error())
		}

		if h.GoogleClientId != payload.Audience {
			fmt.Println("invalid client")
		}
	}

	res, err := h.InterfaceService.UserLoginService(c.Request().Context(), req)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, res)

}
func (h *Handler) DeleteUser(c echo.Context) error {
	payload := get_token.GetUserPayloadToken(c)
	err := h.InterfaceService.DeleteUserService(c.Request().Context(), payload)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, echo.Map{
		"message": "success",
	})
}
func (h *Handler) UpdateUser(c echo.Context) error {

	var req UpdateUserRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	payload := get_token.GetUserPayloadToken(c)

	data := UpdateUserDTO{
		Request: req,
		Payload: payload,
	}

	res, err := h.InterfaceService.UpdateUserService(c.Request().Context(), data)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, res)
}
