package ws

import (
	"geolocation/internal/get_token"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
)

type Handler struct {
	InterfaceService InterfaceService
	hub              *Hub
}

func NewWsHandler(hub *Hub, InterfaceService InterfaceService) *Handler {
	return &Handler{
		hub:              hub,
		InterfaceService: InterfaceService,
	}
}

var upgrade = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (h *Handler) HandleWs(c echo.Context) error {
	payload := get_token.GetUserPayloadToken(c)

	conn, err := upgrade.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Println(err)
	}
	defer func(conn *websocket.Conn) {
		err = conn.Close()
		if err != nil {
		}
	}(conn)

	cl := &Client{
		Conn:    conn,
		Message: make(chan *Message, 10),
		UserId:  payload.ID,
		Name:    payload.Name,
		Payload: payload,
	}

	h.hub.Register <- cl

	go cl.writeMessage()

	cl.readMessage(h.hub, h.InterfaceService)

	return nil
}

func (h *Handler) CreateChatRoom(c echo.Context) error {
	var req CreateChatRoomRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	payload := get_token.GetUserPayloadToken(c)

	res, err := h.InterfaceService.CreateChatRoomService(c.Request().Context(), req, payload.ID)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, res)
}
