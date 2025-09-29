package payment

import (
	"encoding/json"
	"geolocation/internal/get_token"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	InterfaceService InterfaceService
}

func NewPaymentHandler(InterfaceService InterfaceService) *Handler {
	return &Handler{
		InterfaceService,
	}
}

// StripeWebhookHandler godoc
// @Summary Processar Webhook do Stripe
// @Description Recebe e processa os eventos do webhook do Stripe.
// @Tags Pagamentos
// @Accept json
// @Produce json
// @Success 200 {string} string "Sucesso"
// @Failure 400 {string} string "Requisição Inválida"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /webhook/stripe [post]
// @Security ApiKeyAuth
func (p *Handler) StripeWebhookHandler(c echo.Context) error {
	jsonData, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request body")
	}

	var event map[string]interface{}
	if err := json.Unmarshal(jsonData, &event); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid JSON format")
	}
	eventType, ok := event["type"].(string)
	if !ok {
		return c.JSON(http.StatusBadRequest, "Missing event type")
	}

	result, err := p.InterfaceService.ProcessStripeEvent(c.Request().Context(), eventType, event)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

// GetPaymentHistHandler godoc
// @Summary Obter Histórico de Pagamentos
// @Description Recupera o histórico de pagamentos de um usuário.
// @Tags Pagamentos
// @Accept json
// @Produce json
// @Param id path int true "ID do Usuário"
// @Success 200 {object} []PaymentHistResponse "Lista do Histórico de Pagamentos"
// @Failure 400 {string} string "Requisição Inválida"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /payment-history [get]
// @Security ApiKeyAuth
func (p *Handler) GetPaymentHistHandler(c echo.Context) error {
	payload := get_token.GetUserPayloadToken(c)

	result, err := p.InterfaceService.GetPaymentHistService(c.Request().Context(), payload.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}
