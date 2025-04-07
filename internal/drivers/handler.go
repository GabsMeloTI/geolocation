package drivers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"geolocation/internal/get_token"
	"geolocation/validation"
)

type Handler struct {
	InterfaceService InterfaceService
}

func NewDriversHandler(InterfaceService InterfaceService) *Handler {
	return &Handler{InterfaceService}
}

// CreateDriverHandler godoc
// @Summary Criar um Motorista.
// @Description Cria um motorista.
// @Tags Motoristas
// @Accept json
// @Produce json
// @Param request body CreateDriverRequest true "Requisição de Motorista"
// @Success 200 {object} DriverResponse "Informações do Motorista"
// @Failure 400 {string} string "Requisição Inválida"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /driver/create [post]
// @Security ApiKeyAuth
func (p *Handler) CreateDriverHandler(c echo.Context) error {
	var request CreateDriverRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	if !validation.ValidateCPF(request.Cpf) {
		return c.JSON(http.StatusBadRequest, "CPF inválido")
	}

	if !validation.ValidatePhone(request.Phone) {
		return c.JSON(http.StatusBadRequest, "Telefone inválido")
	}

	// if !validation.ValidateCNH(request.LicenseNumber) {
	//     return c.JSON(http.StatusBadRequest, "CNH inválida")
	// }

	payload := get_token.GetUserPayloadToken(c)
	data := CreateDriverDto{
		CreateDriverRequest: request,
		UserID:              payload.ID,
		ProfileId:           payload.ProfileID,
	}

	result, err := p.InterfaceService.CreateDriverService(c.Request().Context(), data)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

// UpdateDriverHandler godoc
// @Summary Atualizar um Motorista.
// @Description Atualiza os dados de um motorista.
// @Tags Motoristas
// @Accept json
// @Produce json
// @Param user body UpdateDriverRequest true "Requisição de Motorista"
// @Success 200 {object} DriverResponse "Informações do Motorista"
// @Failure 400 {string} string "Requisição Inválida"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /driver/update [put]
// @Security ApiKeyAuth
func (p *Handler) UpdateDriverHandler(c echo.Context) error {
	var request UpdateDriverRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	if !validation.ValidatePhone(request.Phone) {
		return c.JSON(http.StatusBadRequest, "Telefone inválido")
	}

	payload := get_token.GetUserPayloadToken(c)
	data := UpdateDriverDto{
		UpdateDriverRequest: request,
		UserID:              payload.ID,
	}

	result, err := p.InterfaceService.UpdateDriverService(c.Request().Context(), data)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

// DeleteDriversHandler godoc
// @Summary Excluir um Motorista.
// @Description Exclui um motorista.
// @Tags Motoristas
// @Accept json
// @Produce json
// @Param id path string true "ID do Motorista"
// @Success 200 {string} string "Sucesso"
// @Failure 400 {string} string "Requisição Inválida"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /driver/delete/{id} [put]
// @Security ApiKeyAuth
func (p *Handler) DeleteDriversHandler(c echo.Context) error {
	idStr := c.Param("id")
	id, err := validation.ParseStringToInt64(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	payload := get_token.GetUserPayloadToken(c)
	err = p.InterfaceService.DeleteDriverService(c.Request().Context(), id, payload.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, "Sucesso")
}

// GetDriverHandler godoc
// @Summary Obter Motorista.
// @Description Recupera as informações do motorista.
// @Tags Motoristas
// @Accept json
// @Produce json
// @Success 200 {object} DriverResponse "Informações do Motorista"
// @Failure 400 {string} string "Requisição Inválida"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /driver/list [get]
// @Security ApiKeyAuth
func (p *Handler) GetDriverHandler(c echo.Context) error {
	payload := get_token.GetUserPayloadToken(c)

	result, err := p.InterfaceService.GetDriverService(c.Request().Context(), payload.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

// GetDriverByIdHandler godoc
// @Summary Obter Motorista por ID.
// @Description Recupera as informações do motorista a partir do ID.
// @Tags Motoristas
// @Accept json
// @Produce json
// @Param id path string true "ID do Motorista"
// @Success 200 {object} DriverResponse "Informações do Motorista"
// @Failure 400 {string} string "Requisição Inválida"
// @Failure 500 {string} string "Erro Interno do Servidor"
// @Router /driver/list/{id} [get]
// @Security ApiKeyAuth
func (p *Handler) GetDriverByIdHandler(c echo.Context) error {
	idStr := c.Param("id")
	id, err := validation.ParseStringToInt64(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	result, err := p.InterfaceService.GetDriverByIdService(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}
