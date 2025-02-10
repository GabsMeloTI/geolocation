package hist

//
//import (
//	"errors"
//	"geolocation/validation"
//	"github.com/labstack/echo/v4"
//	"net/http"
//)
//
//type Handler struct {
//	InterfaceService InterfaceService
//}
//
//func NewRoutesHandler(InterfaceService InterfaceService) *Handler {
//	return &Handler{InterfaceService}
//}
//
//func (h *Handler) GetPublicToken(e echo.Context) error {
//	var ip string
//	if err := e.Bind(&frontInfo); err != nil {
//		return e.JSON(http.StatusBadRequest, err.Error())
//	}
//
//	result, err := h.InterfaceService.GetPublicToken(e.Request().Context(), frontInfo)
//	if err != nil {
//		statusCode := http.StatusInternalServerError
//		if errors.Is(err, echo.ErrNotFound) {
//			statusCode = http.StatusNotFound
//		}
//		return e.JSON(statusCode, err.Error())
//	}
//
//	return e.JSON(http.StatusOK, result)
//}
