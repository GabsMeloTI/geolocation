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

// HandleWs godoc
// @Summary Handle WebSocket connection.
// @Description Establishes a WebSocket connection for real-time communication.
// @Tags WebSocket
// @Accept json
// @Produce json
// @Success 101 {string} string "Switching Protocols"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /ws [get]
// @Security ApiKeyAuth
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

// CreateChatRoom godoc
// @Summary Create a new chat room.
// @Description Creates a chat room associated with an advertisement.
// @Tags WebSocket
// @Accept json
// @Produce json
// @Param request body CreateChatRoomRequest true "Chat Room Request"
// @Success 200 {object} CreateChatRoomResponse "Chat Room Info"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /chat/create-room [post]
// @Security ApiKeyAuth
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

// GetMessagesByRoomId godoc
// @Summary Retrieve messages from a chat room.
// @Description Fetches chat messages by the specified room ID.
// @Tags WebSocket
// @Accept json
// @Produce json
// @Param room_id path int true "Chat Room ID"
// @Success 200 {array} MessageResponse "List of chat messages"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /chat/messages/:room_id [get]
// @Security ApiKeyAuth
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

func (h *Handler) UpdateMessageOffer(c echo.Context) error {
	var request UpdateOfferRequest

	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	payload := get_token.GetUserPayloadToken(c)

	data := UpdateOfferDTO{
		Request: request,
		Payload: payload,
	}

	err := h.InterfaceService.UpdateMessageOfferService(c.Request().Context(), data, h.hub)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.String(http.StatusOK, "Success")
}

func (h *Handler) UpdateFreightLocation(c echo.Context) error {
	var request UpdateFreightData

	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	res, err := h.InterfaceService.FreightLocationDetailsService(c.Request().Context(), request)

	if err != nil {
		return err
	}
	updateFreightMessage := &UpdateFreightMessage{
		AdvertisementId:         request.AdvertisementId,
		Latitude:                request.OriginLatitude,
		Longitude:               request.OriginLongitude,
		DurationText:            res.DurationText,
		DistanceText:            res.DistanceText,
		DriverName:              res.DriverName,
		TractorUnitLicensePlate: res.TractorUnitLicensePlate,
		TrailerLicensePlate:     res.TrailerLicensePlate,
		TypeMessage:             "update_freight",
	}

	if cl, ok := h.hub.Clients[res.AdvertisementUserId]; ok {
		err = cl.Conn.WriteJSON(updateFreightMessage)
	}

	return c.JSON(http.StatusOK, res)
}
