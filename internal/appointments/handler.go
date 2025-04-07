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

// UpdateAppointmentHandler godoc
// @Summary Atualizar um Agendamento.
// @Description Atualiza um agendamento.
// @Tags Agendamentos
// @Accept json
// @Produce json
// @Param user body UpdateAppointmentRequest true "Requisição de Agendamento"
// @Success 200 {string} string "Sucesso"
// @Failure 400 {string} string "Requisição Inválida"
// @Failure 500 {string} string "Erro Interno do Servidor"
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

	return c.JSON(http.StatusOK, "Sucesso")
}

// DeleteAppointmentsHandler godoc
// @Summary Excluir um Agendamento.
// @Description Exclui um agendamento.
// @Tags Agendamentos
// @Accept json
// @Produce json
// @Param id path string true "ID do Agendamento"
// @Success 200 {string} string "Sucesso"
// @Failure 400 {string} string "Requisição Inválida"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /appointment/delete/{id} [put]
// @Security ApiKeyAuth
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
	return c.JSON(http.StatusOK, "Sucesso")
}

// GetAppointmentByUserIDHandler godoc
// @Summary Obter lista de Agendamentos.
// @Description Recupera a lista de agendamentos para um usuário.
// @Tags Agendamentos
// @Accept json
// @Produce json
// @Param id path string true "ID do Usuário"
// @Success 200 {object} AppointmentResponseList "Informações dos Agendamentos"
// @Failure 400 {string} string "Requisição Inválida"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /appointment/{id} [get]
// @Security ApiKeyAuth
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
