package user

import (
	"fmt"
	"geolocation/internal/get_token"
	"geolocation/pkg/sso"
	"github.com/labstack/echo/v4"
	"net/http"
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
	var data CreateUserDTO
	var req CreateUserRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	authorization := c.Request().Header.Get("Authorization")

	if authorization != "" {

		token := strings.Replace(authorization, "Bearer ", "", 1)

		payload, err := sso.ValidateGoogleToken(token)

		if err != nil {
			return c.JSON(http.StatusUnauthorized, err.Error())
		}

		if h.GoogleClientId != payload.Audience {
			fmt.Println("invalid client")
		}

		data = CreateUserDTO{
			Request: req,
			Sso:     true,
			Payload: payload,
		}
	}

	data = CreateUserDTO{
		Request: req,
	}

	err := data.Validate()

	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	res, err := h.InterfaceService.CreateUserService(c.Request().Context(), data)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, res)
}

func (h *Handler) UserLogin(c echo.Context) error {
	var req LoginRequest
	var data LoginDTO

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	authorization := c.Request().Header.Get("Authorization")

	if authorization != "" {
		token := strings.Replace(authorization, "Bearer ", "", 1)

		payload, err := sso.ValidateGoogleToken(token)

		if err != nil {
			return c.JSON(http.StatusUnauthorized, err.Error())
		}

		if h.GoogleClientId != payload.Audience {
			fmt.Println("invalid client")
		}
		data = LoginDTO{
			Request: LoginRequest{
				Email: payload.Email,
			},
			Sso: true,
		}
	}

	if !data.Sso {
		data = LoginDTO{
			Request: req,
		}
	}

	res, err := h.InterfaceService.UserLoginService(c.Request().Context(), data)

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
