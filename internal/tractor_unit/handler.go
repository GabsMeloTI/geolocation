package tractor_unit

import (
	"geolocation/internal/get_token"
	"geolocation/validation"
	"github.com/labstack/echo/v4"
	"net/http"
)

type Handler struct {
	InterfaceService InterfaceService
}

func NewTractorUnitsHandler(InterfaceService InterfaceService) *Handler {
	return &Handler{InterfaceService}
}

// CreateTractorUnitHandler godoc
// @Summary Criar uma Unidade Tratora.
// @Description Cria uma unidade tratora.
// @Tags Cavalo
// @Accept json
// @Produce json
// @Param request body CreateTractorUnitRequest true "Requisição de Unidade Tratora"
// @Success 200 {object} TractorUnitResponse "Informações da Unidade Tratora"
// @Failure 400 {string} string "Requisição Inválida"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /tractor-unit/create [post]
// @Security ApiKeyAuth
func (p *Handler) CreateTractorUnitHandler(c echo.Context) error {
	var request CreateTractorUnitRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	payload := get_token.GetUserPayloadToken(c)
	data := CreateTractorUnitDto{
		CreateTractorUnitRequest: request,
		UserID:                   payload.ID,
	}

	result, err := p.InterfaceService.CreateTractorUnitService(c.Request().Context(), data)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

// UpdateTractorUnitHandler godoc
// @Summary Atualizar uma Unidade Tratora.
// @Description Atualiza os dados de uma unidade tratora.
// @Tags Cavalo
// @Accept json
// @Produce json
// @Param user body UpdateTractorUnitRequest true "Requisição de Unidade Tratora"
// @Success 200 {object} TractorUnitResponse "Informações da Unidade Tratora"
// @Failure 400 {string} string "Requisição Inválida"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /tractor-unit/update [put]
// @Security ApiKeyAuth
func (p *Handler) UpdateTractorUnitHandler(c echo.Context) error {
	var request UpdateTractorUnitRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	payload := get_token.GetUserPayloadToken(c)
	data := UpdateTractorUnitDto{
		UpdateTractorUnitRequest: request,
		UserID:                   payload.ID,
	}

	result, err := p.InterfaceService.UpdateTractorUnitService(c.Request().Context(), data)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

// DeleteTractorUnitHandler godoc
// @Summary Excluir uma Unidade Tratora.
// @Description Exclui uma unidade tratora.
// @Tags Cavalo
// @Accept json
// @Produce json
// @Param id path string true "ID da Unidade Tratora"
// @Success 200 {string} string "Sucesso"
// @Failure 400 {string} string "Requisição Inválida"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /tractor-unit/delete/{id} [put]
// @Security ApiKeyAuth
func (p *Handler) DeleteTractorUnitHandler(c echo.Context) error {
	idStr := c.Param("id")
	id, err := validation.ParseStringToInt64(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	payload := get_token.GetUserPayloadToken(c)
	err = p.InterfaceService.DeleteTractorUnitService(c.Request().Context(), id, payload.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, "Sucesso")
}

// GetTractorUnitHandler godoc
// @Summary Obter Unidade Tratora.
// @Description Recupera as informações da unidade tratora.
// @Tags Cavalo
// @Accept json
// @Produce json
// @Param id path string true "ID da Unidade Tratora"
// @Success 200 {object} TractorUnitResponse "Informações da Unidade Tratora"
// @Failure 400 {string} string "Requisição Inválida"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /tractor-unit/list [get]
// @Security ApiKeyAuth
func (p *Handler) GetTractorUnitHandler(c echo.Context) error {
	payload := get_token.GetUserPayloadToken(c)

	result, err := p.InterfaceService.GetTractorUnitService(c.Request().Context(), payload.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

// GetTractorUnitByIdHandler godoc
// @Summary Obter Unidade Tratora por ID.
// @Description Recupera as informações da unidade tratora a partir do ID.
// @Tags Cavalo
// @Accept json
// @Produce json
// @Param id path string true "ID da Unidade Tratora"
// @Success 200 {object} TractorUnitResponse "Informações da Unidade Tratora"
// @Failure 400 {string} string "Requisição Inválida"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /tractor-unit/list/{id} [get]
// @Security ApiKeyAuth
func (p *Handler) GetTractorUnitByIdHandler(c echo.Context) error {
	idStr := c.Param("id")
	id, err := validation.ParseStringToInt64(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	result, err := p.InterfaceService.GetTractorUnitByIdService(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

// CheckPlateHandler godoc
// @Summary Verificar Placa.
// @Description Verifica se a placa existe ou já está cadastrada.
// @Tags Cavalo
// @Accept json
// @Produce json
// @Param plate path string true "Placa do Veículo"
// @Success 200 {object} interface{} "Resultado da verificação da placa"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /tractor-unit/check-plate/{plate} [get]
// @Security ApiKeyAuth
func (p *Handler) CheckPlateHandler(c echo.Context) error {
	plate := c.Param("plate")

	result, err := p.InterfaceService.CheckPlate(plate)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}
