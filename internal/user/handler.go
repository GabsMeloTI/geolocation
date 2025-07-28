package user

import (
	"errors"
	"geolocation/pkg/gpt"
	"geolocation/pkg/plate"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"geolocation/internal/get_token"
	"geolocation/validation"
)

type Handler struct {
	InterfaceService InterfaceService
	GoogleClientId   string
}

func NewUserHandler(InterfaceService InterfaceService, GoogleClientId string) *Handler {
	return &Handler{InterfaceService, GoogleClientId}
}

// DeleteUser godoc
// @Summary Excluir um Usuário.
// @Description Exclui a conta do usuário autenticado.
// @Tags Usuários
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string "Mensagem de sucesso"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /user/delete [delete]
// @Security ApiKeyAuth
func (h *Handler) DeleteUser(c echo.Context) error {
	payload := get_token.GetUserPayloadToken(c)
	err := h.InterfaceService.DeleteUserService(c.Request().Context(), payload)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, echo.Map{
		"message": "sucesso",
	})
}

// UpdateUser godoc
// @Summary Atualizar Informações do Usuário.
// @Description Atualiza as informações do perfil do usuário autenticado.
// @Tags Usuários
// @Accept json
// @Produce json
// @Param request body UpdateUserRequest true "Requisição de Atualização do Usuário"
// @Success 200 {object} UpdateUserResponse "Informações Atualizadas do Usuário"
// @Failure 400 {string} string "Requisição Inválida"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /user/update [put]
// @Security ApiKeyAuth
func (h *Handler) UpdateUser(c echo.Context) error {
	var req UpdateUserRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	payload := get_token.GetUserPayloadToken(c)

	data := UpdateUserDTO{
		Request: req,
		Payload: payload,
	}

	res, err := h.InterfaceService.UpdateUserService(c.Request().Context(), data)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, res)
}

// UpdateUserAddress godoc
// @Summary Atualizar Endereço do Usuário.
// @Description Atualiza os detalhes do endereço do usuário.
// @Tags Usuários
// @Accept json
// @Produce json
// @Param request body UpdateUserAddressRequest true "Requisição de Atualização de Endereço do Usuário"
// @Success 200 {object} UpdateUserResponse "Informações Atualizadas do Endereço"
// @Failure 400 {string} string "Requisição Inválida"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /user/address/update [put]
// @Security ApiKeyAuth
func (h *Handler) UpdateUserAddress(c echo.Context) error {
	var req UpdateUserAddressRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	res, err := h.InterfaceService.UpdateUserAddressService(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, res)
}

// UpdateUserPersonalInfo godoc
// @Summary Atualizar Informações Pessoais do Usuário.
// @Description Atualiza os dados pessoais do usuário.
// @Tags Usuários
// @Accept json
// @Produce json
// @Param request body UpdateUserPersonalInfoRequest true "Requisição de Atualização de Informações Pessoais do Usuário"
// @Success 200 {object} UpdateUserResponse "Informações Pessoais Atualizadas"
// @Failure 400 {string} string "Requisição Inválida"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /user/personal/update [put]
// @Security ApiKeyAuth
func (h *Handler) UpdateUserPersonalInfo(c echo.Context) error {
	var req UpdateUserPersonalInfoRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	res, err := h.InterfaceService.UpdateUserPersonalInfoService(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, res)
}

// GetUserInfo godoc
// @Summary Obter Informações do Usuário.
// @Description Recupera todos os dados do usuário.
// @Tags Usuários
// @Accept json
// @Produce json
// @Success 200 {object} GetUserResponse "Informações do Usuário"
// @Failure 400 {string} string "Requisição Inválida"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /user/info [get]
// @Security ApiKeyAuth
func (h *Handler) GetUserInfo(c echo.Context) error {
	payload := get_token.GetUserPayloadToken(c)
	res, err := h.InterfaceService.GetUserService(c.Request().Context(), payload.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, res)
}

// RecoverPassword godoc
// @Summary Recuperar Senha.
// @Description Inicia o processo de recuperação de senha para o usuário.
// @Tags Usuários
// @Accept json
// @Produce json
// @Param request body RecoverPasswordRequest true "Requisição de Recuperação de Senha"
// @Success 200 {string} string "Sucesso"
// @Failure 400 {string} string "Requisição Inválida"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /user/recover [post]
func (h *Handler) RecoverPassword(c echo.Context) error {
	var req RecoverPasswordRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	err := h.InterfaceService.RecoverPasswordService(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, "Sucesso")
}

// ConfirmRecoverPassword godoc
// @Summary Confirmar Recuperação de Senha.
// @Description Confirma a recuperação de senha utilizando o token recebido.
// @Tags Usuários
// @Accept json
// @Produce json
// @Param request body ConfirmRecoverPasswordRequest true "Requisição de Confirmação de Recuperação de Senha"
// @Success 200 {string} string "Sucesso"
// @Failure 400 {string} string "Requisição Inválida"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /user/recover/confirm [post]
// @Security ApiKeyAuth
func (h *Handler) ConfirmRecoverPassword(c echo.Context) error {
	var req ConfirmRecoverPasswordRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	if ok := validation.ValidatePassword(req.Password); !ok {
		return c.JSON(http.StatusBadRequest, errors.New("senha inválida"))
	}

	if req.ConfirmPassword != req.Password {
		return c.JSON(
			http.StatusBadRequest,
			errors.New("a senha e a confirmação são diferentes"),
		)
	}
	payload := get_token.GetUserPayloadToken(c)
	bearerToken := c.Request().Header.Get("Authorization")
	tokenStr := strings.Replace(bearerToken, "Bearer ", "", 1)

	data := ConfirmRecoverPasswordDTO{
		Request: req,
		Token:   tokenStr,
		UserID:  payload.ID,
	}

	err := h.InterfaceService.ConfirmRecoverPasswordService(c.Request().Context(), data)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, "Sucesso")
}

// UserExists godoc
// @Summary Verificar Existência do Usuário.
// @Description Verifica se o usuário existe a partir do e-mail fornecido.
// @Tags Usuários
// @Accept json
// @Produce json
// @Param request body UserExitsRequest true "Requisição para Verificar Existência do Usuário"
// @Success 200 {object} GetUserResponse "Dados do Usuário"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /user/exists [post]
func (h *Handler) UserExists(c echo.Context) error {
	var req UserExitsRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	user, err := h.InterfaceService.GetUserByEmailService(c.Request().Context(), req.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, user)
}

// UpdateUserPassword godoc
// @Summary Atualizar Senha do Usuário.
// @Description Atualiza a senha do usuário autenticado.
// @Tags Usuários
// @Accept json
// @Produce json
// @Param request body UpdatePasswordRequest true "Requisição de Atualização de Senha"
// @Success 200 {string} string "Sucesso"
// @Failure 400 {string} string "Requisição Inválida"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /user/password/update [put]
// @Security ApiKeyAuth
func (h *Handler) UpdateUserPassword(c echo.Context) error {
	var request UpdatePasswordRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	if ok := validation.ValidatePassword(request.Password); !ok {
		return c.JSON(http.StatusBadRequest, errors.New("senha inválida").Error())
	}

	payload := get_token.GetUserPayloadToken(c)
	err := h.InterfaceService.UpdateUserPassword(c.Request().Context(), payload.ID, request)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, "Sucesso")
}

func (h *Handler) InfoCaminhao(c echo.Context) error {
	modelo := c.Param("modelo")
	if modelo == "" {
		return errors.New("modelo invalido")
	}

	result, err := gpt.PerguntarAoGptEstruturado(modelo)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

func (h *Handler) ConsultarPlaca(c echo.Context) error {
	placa := c.Param("placa")
	if placa == "" {
		return errors.New("placa invalido")
	}

	result, err := plate.ConsultarPlaca(placa)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}
