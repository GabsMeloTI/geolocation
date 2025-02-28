package ws

import (
	"geolocation/internal/get_token"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"strconv"
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
		Message: make(chan *OutgoingMessage, 10),
		UserId:  payload.ID,
		Name:    payload.Name,
		Payload: payload,
	}

	h.hub.Register <- cl

	home, err := h.InterfaceService.GetHomeService(c.Request().Context(), payload)

	if err != nil {
		log.Println(err)
	}

	if err = conn.WriteJSON(home); err != nil {
		log.Println(err)
	}

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

func (h *Handler) GetMessagesByRoomId(c echo.Context) error {
	roomIdStr := c.Param("room_id")
	roomId, err := strconv.ParseInt(roomIdStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	payload := get_token.GetUserPayloadToken(c)

	res, err := h.InterfaceService.GetChatMessagesByRoomIdService(c.Request().Context(), roomId, payload.ID)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, res)
}
