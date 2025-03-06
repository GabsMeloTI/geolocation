package appointments

import (
	"geolocation/internal/get_token"
	"geolocation/validation"
	"github.com/labstack/echo/v4"
	"net/http"
)

type Handler struct {
	InterfaceService InterfaceService
}

func NewAppointmentHandler(InterfaceService InterfaceService) *Handler {
	return &Handler{InterfaceService}
}

// CreateAppointmentHandler godoc
// @Summary Create an Appointment.
// @Description Create an Appointment.
// @Tags Appointments
// @Accept json
// @Produce json
// @Param request body CreateAppointmentRequest true "Appointment Request"
// @Success 200 {object} AppointmentResponse "Appointment Info"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /appointment/create [post]
// @Security ApiKeyAuth
func (p *Handler) CreateAppointmentHandler(c echo.Context) error {
	var request CreateAppointmentRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	payload := get_token.GetUserPayloadToken(c)

	data := CreateAppointmentDTO{
		Request: request,
		Payload: payload,
	}

	result, err := p.InterfaceService.CreateAppointmentService(c.Request().Context(), data)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

// UpdateAppointmentHandler godoc
// @Summary Update an Appointment.
// @Description Update an Appointment.
// @Tags Appointments
// @Accept json
// @Produce json
// @Param user body UpdateAppointmentRequest true "Appointment Request"
// @Success 200 {string} string "Success"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /appointment/update [put]
// @Security ApiKeyAuth
func (p *Handler) UpdateAppointmentHandler(c echo.Context) error {
	var request UpdateAppointmentRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	payload := get_token.GetUserPayloadToken(c)
	data := UpdateAppointmentDTO{
		Request: request,
		Payload: payload,
	}

	err := p.InterfaceService.UpdateAppointmentSituationService(c.Request().Context(), data)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, "Success")
}

// DeleteAppointmentsHandler godoc
// @Summary Delete Appointment.
// @Description Delete Appointment.
// @Tags Appointments
// @Accept json
// @Produce json
// @Param id path string true "Appointment id"
// @Success 200
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /appointment/delete/{id} [put]
func (p *Handler) DeleteAppointmentsHandler(c echo.Context) error {
	idStr := c.Param("id")
	id, err := validation.ParseStringToInt64(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	err = p.InterfaceService.DeleteAppointmentService(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, "Success")
}

// GetAppointmentByUserIDHandler godoc
// @Summary Get list Appointment.
// @Description Get list Appointment.
// @Tags Appointments
// @Accept json
// @Produce json
// @Param id path string true "User id"
// @Success 200 {object} AppointmentResponseList "Appointment Info"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /appointment/{id} [get]
func (p *Handler) GetAppointmentByUserIDHandler(c echo.Context) error {
	idStr := c.Param("id")
	id, err := validation.ParseStringToInt64(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	result, err := p.InterfaceService.GetAppointmentByUserIDService(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, result)
}
