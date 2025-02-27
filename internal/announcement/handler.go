package announcement

import (
	"geolocation/validation"
	"github.com/labstack/echo/v4"
	"net/http"
)

type Handler struct {
	InterfaceService InterfaceService
}

func NewAnnouncementHandler(InterfaceService InterfaceService) *Handler {
	return &Handler{InterfaceService}
}

// CreateAnnouncementHandler godoc
// @Summary Create a Announcement.
// @Description Create a Announcement.
// @Tags Announcement
// @Accept json
// @Produce json
// @Param request body CreateAnnouncementRequest true "Announcement Request"
// @Success 200 {object} AnnouncementResponse "Announcement Info"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /trailer/create [post]
// @Security ApiKeyAuth
func (p *Handler) CreateAnnouncementHandler(c echo.Context) error {
	var request CreateAnnouncementRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	result, err := p.InterfaceService.CreateAnnouncementService(c.Request().Context(), request)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

// UpdateAnnouncementHandler godoc
// @Summary Update a Announcement.
// @Description Update a Announcement.
// @Tags Announcement
// @Accept json
// @Produce json
// @Param user body UpdateAnnouncementRequest true "Announcement Request"
// @Success 200 {object} AnnouncementResponse "Announcement Info"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /trailer/update [put]
// @Security ApiKeyAuth
func (p *Handler) UpdateAnnouncementHandler(c echo.Context) error {
	var request UpdateAnnouncementRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	result, err := p.InterfaceService.UpdateAnnouncementService(c.Request().Context(), request)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

// DeleteAnnouncementHandler godoc
// @Summary Delete Announcement.
// @Description Delete Announcement.
// @Tags Announcement
// @Accept json
// @Produce json
// @Param id path string true "Announcement id"
// @Success 200
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /trailer/delete/{id} [put]
func (p *Handler) DeleteAnnouncementHandler(c echo.Context) error {
	idStr := c.Param("id")
	id, err := validation.ParseStringToInt64(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	err = p.InterfaceService.DeleteAnnouncementService(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, "Success")
}
