package advertisement

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"geolocation/internal/get_token"
	"geolocation/validation"
)

type Handler struct {
	InterfaceService InterfaceService
}

func NewAdvertisementHandler(InterfaceService InterfaceService) *Handler {
	return &Handler{InterfaceService}
}

// CreateAdvertisementHandler godoc
// @Summary Criar um Anúncio.
// @Description Cria um anúncio.
// @Tags Anúncio
// @Accept json
// @Produce json
// @Param request body CreateAdvertisementRequest true "Requisição de Anúncio"
// @Success 200 {object} AdvertisementResponse "Informações do Anúncio"
// @Failure 400 {string} string "Requisição Inválida"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /advertisement/create [post]
// @Security ApiKeyAuth
func (p *Handler) CreateAdvertisementHandler(c echo.Context) error {
	var request CreateAdvertisementRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	payload := get_token.GetUserPayloadToken(c)
	data := CreateAdvertisementDto{
		CreateAdvertisementRequest: request,
		UserID:                     payload.ID,
		CreatedWho:                 payload.Name,
	}
	fmt.Println(payload.Name)
	result, err := p.InterfaceService.CreateAdvertisementService(
		c.Request().Context(),
		data,
		payload.ProfileID,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

func (p *Handler) UpdatedAdvertisementFinishedCreate(c echo.Context) error {
	var request UpdatedAdvertisementFinishedCreate
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	payload := get_token.GetUserPayloadToken(c)
	request.UserID = payload.ID
	result, err := p.InterfaceService.UpdatedAdvertisementFinishedCreate(
		c.Request().Context(),
		request,
		payload.ProfileID,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

// UpdateAdvertisementHandler godoc
// @Summary Atualizar um Anúncio.
// @Description Atualiza um anúncio.
// @Tags Anúncio
// @Accept json
// @Produce json
// @Param user body UpdateAdvertisementRequest true "Requisição de Anúncio"
// @Success 200 {object} AdvertisementResponse "Informações do Anúncio"
// @Failure 400 {string} string "Requisição Inválida"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /advertisement/update [put]
// @Security ApiKeyAuth
func (p *Handler) UpdateAdvertisementHandler(c echo.Context) error {
	var request UpdateAdvertisementRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	payload := get_token.GetUserPayloadToken(c)
	data := UpdateAdvertisementDto{
		UpdateAdvertisementRequest: request,
		UserID:                     payload.ID,
		UpdatedWho: sql.NullString{
			String: payload.Name,
			Valid:  true,
		},
	}

	result, err := p.InterfaceService.UpdateAdvertisementService(
		c.Request().Context(),
		data,
		payload.ProfileID,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

// DeleteAdvertisementHandler godoc
// @Summary Excluir um Anúncio.
// @Description Exclui um anúncio.
// @Tags Anúncio
// @Accept json
// @Produce json
// @Param id path string true "ID do Anúncio"
// @Success 200
// @Failure 400 {string} string "Requisição Inválida"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /advertisement/delete/{id} [put]
// @Security ApiKeyAuth
func (p *Handler) DeleteAdvertisementHandler(c echo.Context) error {
	idStr := c.Param("id")
	id, err := validation.ParseStringToInt64(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	payload := get_token.GetUserPayloadToken(c)
	data := DeleteAdvertisementRequest{
		ID:     id,
		UserID: payload.ID,
		UpdatedWho: sql.NullString{
			String: payload.Name,
			Valid:  true,
		},
	}
	err = p.InterfaceService.DeleteAdvertisementService(c.Request().Context(), data)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, "Success")
}

// GetAllAdvertisementHandler godoc
// @Summary Obter Todos os Anúncios.
// @Description Recupera todos os anúncios.
// @Tags Anúncio
// @Accept json
// @Produce json
// @Success 200 {object} []AdvertisementResponseAll "Lista de Anúncios"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /advertisement/list [get]
// @Security ApiKeyAuth
func (p *Handler) GetAllAdvertisementHandler(c echo.Context) error {
	result, err := p.InterfaceService.GetAllAdvertisementUser(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

func (p *Handler) GetAllAdvertisementByUserHandler(c echo.Context) error {
	payload := get_token.GetUserPayloadToken(c)

	result, err := p.InterfaceService.GetAllAdvertisementByUser(c.Request().Context(), payload.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

// GetAllAdvertisementPublicHandler godoc
// @Summary Obter Todos os Anúncios (Público).
// @Description Recupera todos os anúncios públicos.
// @Tags Anúncio
// @Accept json
// @Produce json
// @Success 200 {object} []AdvertisementResponseNoUser "Lista de Anúncios"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /public/advertisement/list [get]
func (p *Handler) GetAllAdvertisementPublicHandler(c echo.Context) error {
	result, err := p.InterfaceService.GetAllAdvertisementPublic(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

// UpdateAdsRouteChoose godoc
// @Summary Atualizar Escolha de Rota do Anúncio.
// @Description Atualiza a rota escolhida no anúncio.
// @Tags Anúncio
// @Accept json
// @Produce json
// @Param user body UpdateAdsRouteChooseRequest true "Requisição para escolha de rota do anúncio"
// @Success 200 {string} string "Sucesso"
// @Failure 400 {string} string "Requisição Inválida"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /advertisement/route [put]
// @Security ApiKeyAuth
func (p *Handler) UpdateAdsRouteChoose(c echo.Context) error {
	var request UpdateAdsRouteChooseRequest

	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	payload := get_token.GetUserPayloadToken(c)

	data := UpdateAdsRouteChooseDTO{
		Request: request,
		UserID:  payload.ID,
	}

	err := p.InterfaceService.UpdateAdsRouteChooseService(c.Request().Context(), data)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, "success")
}

// GetAdvertisementByIDService godoc
// @Summary Obter Anúncio por ID.
// @Description Recupera um anúncio pelo seu ID.
// @Tags Anúncio
// @Accept json
// @Produce json
// @Success 200 {object} []AdvertisementResponseNoUser "Informações do Anúncio"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /public/advertisement/list [get]
func (p *Handler) GetAdvertisementByIDService(c echo.Context) error {
	idStr := c.Param("id")
	id, err := validation.ParseStringToInt64(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	result, err := p.InterfaceService.GetAdvertisementByIDService(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

// GetAdvertisementByIDPublicService godoc
// @Summary Obter Anúncio por ID (Público).
// @Description Recupera um anúncio pelo seu ID (público).
// @Tags Anúncio
// @Accept json
// @Produce json
// @Success 200 {object} []AdvertisementResponseNoUser "Informações do Anúncio"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /public/advertisement/list [get]
func (p *Handler) GetAdvertisementByIDPublicService(c echo.Context) error {
	idStr := c.Param("id")
	id, err := validation.ParseStringToInt64(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	result, err := p.InterfaceService.GetAdvertisementByIDPublicService(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}
