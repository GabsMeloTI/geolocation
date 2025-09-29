package trailer

import (
	"geolocation/internal/get_token"
	"geolocation/validation"
	"github.com/labstack/echo/v4"
	"net/http"
)

type Handler struct {
	InterfaceService InterfaceService
}

func NewTrailersHandler(InterfaceService InterfaceService) *Handler {
	return &Handler{InterfaceService}
}

// CreateTrailerHandler godoc
// @Summary Criar um Carroceria.
// @Description Cria um reboque.
// @Tags Carroceria
// @Accept json
// @Produce json
// @Param request body CreateTrailerRequest true "Requisição de Carroceria"
// @Success 200 {object} TrailerResponse "Informações do Carroceria"
// @Failure 400 {string} string "Requisição Inválida"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /trailer/create [post]
// @Security ApiKeyAuth
func (p *Handler) CreateTrailerHandler(c echo.Context) error {
	var request CreateTrailerRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	payload := get_token.GetPayloadToken(c)
	id, err := validation.ParseStringToInt64(payload.UserID)
	data := CreateTrailerDto{
		CreateTrailerRequest: request,
		UserID:               id,
	}

	result, err := p.InterfaceService.CreateTrailerService(c.Request().Context(), data)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

// UpdateTrailerHandler godoc
// @Summary Atualizar um Carroceria.
// @Description Atualiza as informações de um reboque.
// @Tags Carroceria
// @Accept json
// @Produce json
// @Param user body UpdateTrailerRequest true "Requisição de Carroceria"
// @Success 200 {object} TrailerResponse "Informações do Carroceria"
// @Failure 400 {string} string "Requisição Inválida"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /trailer/update [put]
// @Security ApiKeyAuth
func (p *Handler) UpdateTrailerHandler(c echo.Context) error {
	var request UpdateTrailerRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	payload := get_token.GetPayloadToken(c)
	id, err := validation.ParseStringToInt64(payload.UserID)
	data := UpdateTrailerDto{
		UpdateTrailerRequest: request,
		UserID:               id,
	}

	result, err := p.InterfaceService.UpdateTrailerService(c.Request().Context(), data)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

// DeleteTrailerHandler godoc
// @Summary Excluir um Carroceria.
// @Description Exclui um reboque.
// @Tags Carroceria
// @Accept json
// @Produce json
// @Param id path string true "ID do Carroceria"
// @Success 200 {string} string "Sucesso"
// @Failure 400 {string} string "Requisição Inválida"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /trailer/delete/{id} [put]
// @Security ApiKeyAuth
func (p *Handler) DeleteTrailerHandler(c echo.Context) error {
	idStr := c.Param("id")
	id, err := validation.ParseStringToInt64(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	payload := get_token.GetPayloadToken(c)
	idUser, err := validation.ParseStringToInt64(payload.UserID)
	err = p.InterfaceService.DeleteTrailerService(c.Request().Context(), id, idUser)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, "Sucesso")
}

// GetTrailerHandler godoc
// @Summary Obter Carroceria.
// @Description Recupera as informações do reboque.
// @Tags Carroceria
// @Accept json
// @Produce json
// @Param id path string true "ID do Carroceria"
// @Success 200 {object} TrailerResponse "Informações do Carroceria"
// @Failure 400 {string} string "Requisição Inválida"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /trailer/list [get]
// @Security ApiKeyAuth
func (p *Handler) GetTrailerHandler(c echo.Context) error {
	payload := get_token.GetPayloadToken(c)
	idUser, err := validation.ParseStringToInt64(payload.UserID)

	result, err := p.InterfaceService.GetTrailerService(c.Request().Context(), idUser)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

// GetTrailerByIdHandler godoc
// @Summary Obter Carroceria por ID.
// @Description Recupera as informações do reboque a partir do ID.
// @Tags Carroceria
// @Accept json
// @Produce json
// @Param id path string true "ID do Carroceria"
// @Success 200 {object} TrailerResponse "Informações do Carroceria"
// @Failure 400 {string} string "Requisição Inválida"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /tractor-unit/list/{id} [get]
// @Security ApiKeyAuth
func (p *Handler) GetTrailerByIdHandler(c echo.Context) error {
	idStr := c.Param("id")
	id, err := validation.ParseStringToInt64(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	result, err := p.InterfaceService.GetTrailerByIdService(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}
