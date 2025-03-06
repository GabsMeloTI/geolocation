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

// Login godoc
// @Summary Authenticate a user.
// @Description Authenticate a user by email and password.
// @Tags Users
// @Accept json
// @Produce json
// @Param request body RequestLogin true "Login Request"
// @Success 200 {object} ResponseLogin "Authenticated User Info"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /v2/login [post]
// @Security ApiKeyAuth
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

// CreateUser godoc
// @Summary Create a new user.
// @Description Register a new user in the system.
// @Tags Users
// @Accept json
// @Produce json
// @Param request body RequestCreateUser true "User Creation Request"
// @Success 200 {object} ResponseCreateUser "Created User Info"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /v2/create [post]
// @Security ApiKeyAuth
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
