package attachment

import (
	"errors"
	"geolocation/internal/get_token"
	"github.com/labstack/echo/v4"
	"net/http"
)

type Handler struct {
	InterfaceService ServiceInterface
}

func NewAttachmentHandler(service ServiceInterface) *Handler {
	return &Handler{
		InterfaceService: service,
	}
}

// CreateAttachHandler godoc
// @Summary Create Attach.
// @Description Create Attach.
// @Tags Attach
// @Accept multipart/form-data
// @Produce json
// @Param files_input formData []file true "File to upload. Only accepts .jpg, .jpeg, .png, and .pdf files."
// @Param code_id formData string true "CodeID"
// @Param origin formData string true "Origin"
// @Param description formData string true "Description"
// @Success 200 {string} string "Success"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /attach/upload [post]
// @Security ApiKeyAuth
func (h *Handler) CreateAttachHandler(c echo.Context) error {
	request, err := c.MultipartForm()
	if err != nil {
		return c.JSON(http.StatusBadRequest, errors.New("failed to parse multipart form"))
	}

	payload := get_token.GetUserPayloadToken(c)
	err = h.InterfaceService.CreateAttachService(c.Request().Context(), request, payload.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, "Success")
}

// UpdateAttachHandler godoc
// @Summary Update Attachment (Logical Delete and Replace)
// @Description Mark an existing attachment as inactive and create a new record with the provided file.
// @Tags Attach
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Attachment ID"
// @Param userId formData string true "User ID"
// @Param description formData string false "Description"
// @Param file formData file true "New attachment file"
// @Success 200 {string} string "Success"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /attach/update [put]
// @Security ApiKeyAuth
func (h *Handler) UpdateAttachHandler(c echo.Context) error {
	form, err := c.MultipartForm()
	if err != nil {
		return c.JSON(http.StatusBadRequest, errors.New("failed to parse multipart form"))
	}

	payload := get_token.GetUserPayloadToken(c)
	err = h.InterfaceService.UpdateAttachService(c.Request().Context(), payload.ID, form)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, "Success")
}
