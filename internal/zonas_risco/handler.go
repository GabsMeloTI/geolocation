package zonas_risco

import (
	"database/sql"
	"fmt"
	"geolocation/validation"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	InterfaceService InterfaceService
}

func NewZonasRiscoHandler(InterfaceService InterfaceService) *Handler {
	return &Handler{InterfaceService}
}

// CreateZonaRiscoHandler godoc
// @Summary Criar Zona de Risco
// @Description Cria uma nova zona de risco
// @Tags ZonasRisco
// @Accept json
// @Produce json
// @Param request body CreateZonaRiscoRequest true "Requisição de Zona de Risco"
// @Success 200 {object} ZonaRiscoResponse "Zona de Risco criada"
// @Failure 400 {string} string "Requisição Inválida"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /zonas-risco/create [post]
func (h *Handler) CreateZonaRiscoHandler(c echo.Context) error {
	var req CreateZonaRiscoRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	result, err := h.InterfaceService.CreateZonaRiscoService(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, result)
}

// UpdateZonaRiscoHandler godoc
// @Summary Atualizar Zona de Risco
// @Description Atualiza uma zona de risco existente
// @Tags ZonasRisco
// @Accept json
// @Produce json
// @Param request body UpdateZonaRiscoRequest true "Requisição de Atualização"
// @Success 200 {object} ZonaRiscoResponse "Zona de Risco atualizada"
// @Failure 400 {string} string "Requisição Inválida"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /zonas-risco/update [put]
func (h *Handler) UpdateZonaRiscoHandler(c echo.Context) error {
	var req UpdateZonaRiscoRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	result, err := h.InterfaceService.UpdateZonaRiscoService(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, result)
}

// DeleteZonaRiscoHandler godoc
// @Summary Deletar Zona de Risco
// @Description Deleta logicamente uma zona de risco
// @Tags ZonasRisco
// @Accept json
// @Produce json
// @Param id path int true "ID da Zona de Risco"
// @Success 200 {string} string "Sucesso"
// @Failure 400 {string} string "Requisição Inválida"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /zonas-risco/delete/{id} [put]
func (h *Handler) DeleteZonaRiscoHandler(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	err = h.InterfaceService.DeleteZonaRiscoService(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, "Sucesso")
}

// GetZonaRiscoByIdHandler godoc
// @Summary Obter Zona de Risco por ID
// @Description Recupera uma zona de risco pelo ID
// @Tags ZonasRisco
// @Accept json
// @Produce json
// @Param id path int true "ID da Zona de Risco"
// @Success 200 {object} ZonaRiscoResponse "Zona de Risco"
// @Failure 400 {string} string "Requisição Inválida"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /zonas-risco/{id} [get]
func (h *Handler) GetZonaRiscoByIdHandler(c echo.Context) error {
	id, err := parseIDParam(c, "id")
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	result, err := h.InterfaceService.GetZonaRiscoByIdService(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, result)
}

// GetAllZonasRiscoHandler godoc
// @Summary Listar Zonas de Risco
// @Description Lista todas as zonas de risco
// @Tags ZonasRisco
// @Accept json
// @Produce json
// @Success 200 {array} ZonaRiscoResponse "Lista de Zonas de Risco"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /zonas-risco/list [get]
func (h *Handler) GetAllZonasRiscoHandler(c echo.Context) error {
	idStr := c.Param("id")
	id, err := validation.ParseStringToInt64(idStr)
	result, err := h.InterfaceService.GetAllZonasRiscoService(c.Request().Context(), sql.NullInt64{
		Int64: id,
		Valid: true,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, result)
}

// Função utilitária para parsear o ID do path
func parseIDParam(c echo.Context, param string) (int64, error) {
	idStr := c.Param(param)
	var id int64
	_, err := fmt.Sscan(idStr, &id)
	if err != nil {
		return 0, err
	}
	return id, nil
}
