package login

import (
	"geolocation/internal/get_token"
	"net/http"
	"net/mail"

	"github.com/labstack/echo/v4"

	"geolocation/validation"
)

type Handler struct {
	service ServiceInterface
}

func NewHandler(service ServiceInterface) *Handler {
	return &Handler{service}
}

// Login godoc
// @Summary Autenticar um usuário.
// @Description Autentica um usuário utilizando email e senha.
// @Tags Usuários
// @Accept json
// @Produce json
// @Param request body RequestLogin true "Requisição de Login"
// @Success 200 {object} ResponseLogin "Informações do Usuário Autenticado"
// @Failure 400 {string} string "Requisição Inválida"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /login [post]
func (h *Handler) Login(e echo.Context) error {
	var request RequestLogin
	if err := e.Bind(&request); err != nil {
		return e.JSON(http.StatusBadRequest, err.Error())
	}

	result, err := h.service.Login(e.Request().Context(), request)
	if err != nil {
		return e.JSON(http.StatusInternalServerError, err.Error())
	}

	return e.JSON(http.StatusOK, result)
}

// CreateUser godoc
// @Summary Criar um novo usuário.
// @Description Registra um novo usuário no sistema.
// @Tags Usuários
// @Accept json
// @Produce json
// @Param request body RequestCreateUser true "Requisição para criação de usuário"
// @Success 200 {object} ResponseCreateUser "Informações do Usuário Criado"
// @Failure 400 {string} string "Requisição Inválida"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /create [post]
func (h *Handler) CreateUser(e echo.Context) error {
	var request RequestCreateUser
	if err := e.Bind(&request); err != nil {
		return e.JSON(http.StatusBadRequest, err.Error())
	}

	if err := validation.Validate(request); err != nil {
		return e.JSON(http.StatusBadRequest, err.Error())
	}

	_, err := mail.ParseAddress(request.Email)
	if err != nil {
		return e.JSON(http.StatusBadRequest, "endereço de email inválido")
	}

	if request.Token == "" {
		if ok := validation.ValidatePassword(request.Password); !ok {
			return e.JSON(http.StatusBadRequest, "senha inválida")
		}

		if ok := request.Password == request.ConfirmPassword; !ok {
			return e.JSON(http.StatusBadRequest, "a senha e a confirmação são diferentes")
		}
	}

	result, err := h.service.CreateUser(e.Request().Context(), request)
	if err != nil {
		return e.JSON(http.StatusInternalServerError, err.Error())
	}

	return e.JSON(http.StatusOK, result)
}

// CreateUserClient godoc
// @Summary Criar um novo cliente usuário.
// @Description Registra um novo cliente usuário no sistema.
// @Tags Usuários
// @Accept json
// @Produce json
// @Param request body RequestCreateUser true "Requisição para criação de usuário"
// @Success 200 {object} ResponseCreateUser "Informações do Usuário Criado"
// @Failure 400 {string} string "Requisição Inválida"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /create/client [post]
// @Security ApiKeyAuth
func (h *Handler) CreateUserClient(e echo.Context) error {
	var request RequestCreateUser
	if err := e.Bind(&request); err != nil {
		return e.JSON(http.StatusBadRequest, err.Error())
	}

	if err := validation.Validate(request); err != nil {
		return e.JSON(http.StatusBadRequest, err.Error())
	}

	_, err := mail.ParseAddress(request.Email)
	if err != nil {
		return e.JSON(http.StatusBadRequest, "endereço de email inválido")
	}

	if ok := validation.ValidatePassword(request.Password); !ok {
		return e.JSON(http.StatusBadRequest, "senha inválida")
	}

	if ok := request.Password == request.ConfirmPassword; !ok {
		return e.JSON(http.StatusBadRequest, "a senha e a confirmação são diferentes")
	}

	payload := get_token.GetUserPayloadToken(e)
	result, err := h.service.CreateUserClient(e.Request().Context(), request, payload.ID)
	if err != nil {
		return e.JSON(http.StatusInternalServerError, err.Error())
	}

	return e.JSON(http.StatusOK, result)
}
