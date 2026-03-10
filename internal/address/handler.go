package address

import (
	"geolocation/validation"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	InterfaceService InterfaceService
}

func NewAddressHandler(InterfaceService InterfaceService) *Handler {
	return &Handler{InterfaceService}
}

// FindAddressByQueryHandler godoc
// @Summary Buscar Endereço por Consulta
// @Description Busca um endereço por pesquisa. Pode ser: 1. Nome de localidade, 2. Latitude e Longitude ou 3. Endereço completo (Rua, bairro, número).
// @Tags Endereços
// @Accept json
// @Produce json
// @Param q query string true "Consulta do Endereço"
// @Success 200 {object} []AddressResponse "Informações do Endereço"
// @Failure 400 {string} string "Requisição Inválida"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /address/find [get]
// @Security ApiKeyAuth
func (h *Handler) FindAddressByQueryHandler(c echo.Context) error {
	q := c.QueryParam("q")
	if q == "" {
		return c.JSON(http.StatusBadRequest, "O parâmetro de consulta 'q' é obrigatório")
	}
	result, err := h.InterfaceService.FindAddressesByQueryService(c.Request().Context(), q)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, result)
}

// FindAddressByQueryV2Handler godoc
// @Summary Buscar Endereço por Consulta
// @Description Busca um endereço por pesquisa. Endereço completo (Rua, bairro, número).
// @Tags Endereços
// @Accept json
// @Produce json
// @Param q query string true "Consulta do Endereço"
// @Success 200 {object} []AddressResponse "Informações do Endereço"
// @Failure 400 {string} string "Requisição Inválida"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /address/find/v2 [get]
// @Security ApiKeyAuth
func (h *Handler) FindAddressByQueryV2Handler(c echo.Context) error {
	q := c.QueryParam("q")
	if q == "" {
		return c.JSON(http.StatusBadRequest, "O parâmetro de consulta 'q' é obrigatório")
	}
	result, err := h.InterfaceService.FindAddressesByQueryV2Service(c.Request().Context(), q)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, result)
}

// FindUniqueAddressesByCEPHandler godoc
// @Summary Buscar Endereço por Cep
// @Description Buscar Endereço por Cep
// @Tags Endereços
// @Accept json
// @Produce json
// @Param q query string true "Consulta do Endereço"
// @Success 200 {object} []AddressResponse "Informações do Endereço"
// @Failure 400 {string} string "Requisição Inválida"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /address/find/cep/v2 [get]
// @Security ApiKeyAuth
func (h *Handler) FindUniqueAddressesByCEPHandler(c echo.Context) error {
	q := c.QueryParam("q")
	if q == "" {
		return c.JSON(http.StatusBadRequest, "O parâmetro de consulta 'q' é obrigatório")
	}
	result, err := h.InterfaceService.FindUniqueAddressesByCEPService(c.Request().Context(), q)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, result)
}

// FindAddressByCEPHandler godoc
// @Summary Buscar Endereço por CEP
// @Description Encontra um endereço pelo CEP, retornando o tipo baseado nas repetições encontradas.
// @Tags Endereços
// @Accept json
// @Produce json
// @Param cep path string true "CEP"
// @Success 200 {object} []AddressCEPResponse "Informações do Endereço"
// @Failure 400 {string} string "Requisição Inválida"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /address/find/{cep} [get]
// @Security ApiKeyAuth
func (h *Handler) FindAddressByCEPHandler(c echo.Context) error {
	cep := c.Param("cep")
	result, err := h.InterfaceService.FindAddressesByCEPService(c.Request().Context(), cep)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) FindTwoRandomCEPsHandler(c echo.Context) error {
	result, err := h.InterfaceService.FindTwoRandomCEPsService(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, result)
}

// FindStateAll godoc
// @Summary Buscar Todos os Estados
// @Description Retorna todos os estados disponíveis.
// @Tags Endereços
// @Accept json
// @Produce json
// @Success 200 {object} []StateResponse "Lista de Estados"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /address/state [get]
// @Security ApiKeyAuth
func (h *Handler) FindStateAll(c echo.Context) error {
	result, err := h.InterfaceService.FindStateAll(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, result)
}

// FindCityAll godoc
// @Summary Buscar Cidades por Estado
// @Description Retorna todas as cidades de um estado específico, utilizando o ID do estado.
// @Tags Endereços
// @Accept json
// @Produce json
// @Param idState path int true "ID do Estado"
// @Success 200 {object} []CityResponse "Lista de Cidades"
// @Failure 400 {string} string "Requisição Inválida"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /address/city/{idState} [get]
// @Security ApiKeyAuth
func (h *Handler) FindCityAll(c echo.Context) error {
	idStr := c.Param("idState")
	id, err := validation.ParseStringToInt32(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	result, err := h.InterfaceService.FindCityAll(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, result)
}
