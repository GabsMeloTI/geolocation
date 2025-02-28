package drivers

import (
	"geolocation/internal/get_token"
	"geolocation/validation"
	"github.com/labstack/echo/v4"
	"net/http"
)

type Handler struct {
	InterfaceService InterfaceService
}

func NewDriversHandler(InterfaceService InterfaceService) *Handler {
	return &Handler{InterfaceService}
}

// CreateDriverHandler godoc
// @Summary Create a Driver.
// @Description Create a Driver.
// @Tags Drivers
// @Accept json
// @Produce json
// @Param request body CreateDriverRequest true "Driver Request"
// @Success 200 {object} DriverResponse "Driver Info"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /driver/create [post]
// @Security ApiKeyAuth
func (p *Handler) CreateDriverHandler(c echo.Context) error {
	var request CreateDriverRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	if !validation.ValidateCPF(request.Cpf) {
		return c.JSON(http.StatusBadRequest, "CPF inválido")
	}

	if !validation.ValidatePhone(request.Phone) {
		return c.JSON(http.StatusBadRequest, "Telefone inválido")
	}

	if !validation.ValidateCNH(request.LicenseNumber) {
		return c.JSON(http.StatusBadRequest, "CNH inválida")
	}

	payload := get_token.GetUserPayloadToken(c)
	request.UserID = payload.ID

	result, err := p.InterfaceService.CreateDriverService(c.Request().Context(), request)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

// UpdateDriverHandler godoc
// @Summary Update a Driver.
// @Description Update a Driver.
// @Tags Drivers
// @Accept json
// @Produce json
// @Param user body UpdateDriverRequest true "Driver Request"
// @Success 200 {object} DriverResponse "Driver Info"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /driver/update [put]
// @Security ApiKeyAuth
func (p *Handler) UpdateDriverHandler(c echo.Context) error {
	var request UpdateDriverRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	if !validation.ValidatePhone(request.Phone) {
		return c.JSON(http.StatusBadRequest, "Telefone inválido")
	}

	result, err := p.InterfaceService.UpdateDriverService(c.Request().Context(), request)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

// DeleteDriversHandler godoc
// @Summary Delete Driver.
// @Description Delete Driver.
// @Tags Drivers
// @Accept json
// @Produce json
// @Param id path string true "Driver id"
// @Success 200
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /driver/delete/{id} [put]
func (p *Handler) DeleteDriversHandler(c echo.Context) error {
	idStr := c.Param("id")
	id, err := validation.ParseStringToInt64(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	err = p.InterfaceService.DeleteDriverService(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, "Success")
}
