package user

import (
	"fmt"
	"net/http"
	"net/mail"
	"strings"

	"github.com/labstack/echo/v4"

	"geolocation/internal/get_token"
	"geolocation/pkg/sso"
	"geolocation/validation"
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

	// if ok := validation.ValidatePassword(req.Password); !ok {
	// return c.JSON(http.StatusBadRequest, "invalid password")
	// }

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

// DeleteUser godoc
// @Summary Delete a User.
// @Description Deletes the authenticated user account.
// @Tags Users
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string "Success message"
// @Failure 500 {string} string "Internal Server Error"
// @Router /user/delete [delete]
// @Security ApiKeyAuth
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

// UpdateUser godoc
// @Summary Update User Information.
// @Description Update the authenticated user’s profile information.
// @Tags Users
// @Accept json
// @Produce json
// @Param request body UpdateUserRequest true "User Update Request"
// @Success 200 {object} UpdateUserResponse "Updated User Info"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /user/update [put]
// @Security ApiKeyAuth
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

// UpdateUserAddress godoc
// @Summary Update User Address.
// @Description Updates the address details of a user.
// @Tags Users
// @Accept json
// @Produce json
// @Param request body UpdateUserAddressRequest true "User Address Update Request"
// @Success 200 {object} UpdateUserResponse "Updated Address Info"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /user/address/update [put]
// @Security ApiKeyAuth
func (h *Handler) UpdateUserAddress(c echo.Context) error {
	var req UpdateUserAddressRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	res, err := h.InterfaceService.UpdateUserAddressService(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, res)
}

// UpdateUserPersonalInfo godoc
// @Summary Update User Personal Info.
// @Description Updates the personal details of a user.
// @Tags Users
// @Accept json
// @Produce json
// @Param request body UpdateUserPersonalInfoRequest true "User Personal Info Update Request"
// @Success 200 {object} UpdateUserResponse "Updated Personal Info"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /user/personal/update [put]
// @Security ApiKeyAuth
func (h *Handler) UpdateUserPersonalInfo(c echo.Context) error {
	var req UpdateUserPersonalInfoRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	res, err := h.InterfaceService.UpdateUserPersonalInfoService(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, res)
}

// GetUserInfo godoc
// @Summary Get User Info.
// @Description Retrieve all user data.
// @Tags Users
// @Accept json
// @Produce json
// @Success 200 {object} GetUserResponse "Get User "
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /user/info [get]
// @Security ApiKeyAuth
func (h *Handler) GetUserInfo(c echo.Context) error {
	payload := get_token.GetUserPayloadToken(c)
	res, err := h.InterfaceService.GetUserService(c.Request().Context(), payload.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, res)
}
