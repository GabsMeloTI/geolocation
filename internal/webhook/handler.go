package webhook

import (
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
)

func WebhookPaymentHandler(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Erro ao ler o corpo da requisição"})
	}

	ProcessWebhookData(body)

	return c.JSON(http.StatusOK, map[string]string{"message": "Recebido com sucesso"})
}
