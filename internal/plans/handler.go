package plans

import (
	"geolocation/infra/token"
	"geolocation/internal/get_token"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	InterfaceService InterfaceService
	tokenMaker       token.PasetoMaker
}

func NewUserPlanHandler(InterfaceService InterfaceService) *Handler {
	return &Handler{
		InterfaceService,
		token.PasetoMaker{},
	}
}

// CreateUserPlanHandler godoc
// @Summary Criar um Plano de Usuário.
// @Description Atribui um usuário a um plano selecionado.
// @Tags Usuários
// @Accept json
// @Produce json
// @Param id path int true "ID do Plano"
// @Success 200 {object} UserPlanResponse "Informações do Plano do Usuário"
// @Failure 400 {string} string "Requisição Inválida"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /user/plan [post]
// @Security ApiKeyAuth
func (p *Handler) CreateUserPlanHandler(c echo.Context) error {
	var request CreateUserPlanRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	payload := get_token.GetUserPayloadToken(c)
	request.IDUser = payload.ID
	result, err := p.InterfaceService.CreateUserPlanService(c.Request().Context(), request)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

func (p *Handler) GetTokenUserHandler(c echo.Context) error {
	payload := get_token.GetUserPayloadToken(c)

	newToken, err := p.InterfaceService.GenerateUserToken(payload.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
	}

	return c.JSON(http.StatusOK, map[string]string{"token": newToken})
}
