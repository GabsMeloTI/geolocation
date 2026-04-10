package geocoding

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Handler gerencia as requisições HTTP relacionadas a geocodificação
type Handler struct {
	service ServiceInterface
}

// NewHandler cria uma nova instância do Handler de geocodificação
func NewHandler(service ServiceInterface) *Handler {
	return &Handler{service}
}

// LocationHandler processa a requisição para obter coordenadas de origem e destino baseadas em CEP
func (h *Handler) LocationHandler(c echo.Context) error {
	var req RequestLocationDTO
	// Faz o bind dos dados da requisição (JSON) para o DTO
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	// Chama o serviço para processar a localização
	result, err := h.service.LocationService(c.Request().Context(), req)
	if err != nil {
		// Retorna erro 500 caso ocorra alguma falha no serviço/fallback
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	// Retorna os dados formatados com sucesso
	return c.JSON(http.StatusOK, result)
}
