package ws

import (
	"context"
	"encoding/json"
	"geolocation/internal/get_token"
	"github.com/gorilla/websocket"
	"log"
	"time"
)

type Client struct {
	Conn           *websocket.Conn          `json:"conn"`
	Message        chan *OutgoingMessage    `json:"message"`
	UserId         int64                    `json:"user_id"`
	Name           string                   `json:"name"`
	ProfilePicture string                   `json:"profile_picture"`
	Payload        get_token.PayloadUserDTO `json:"payload"`
}

type Message struct {
	RoomId      int64  `json:"room_id"`
	Content     string `json:"content"`
	TypeMessage string `json:"type_message"`
}

type OutgoingMessage struct {
	MessageId      int64      `json:"message_id"`
	RoomId         int64      `json:"room_id"`
	UserId         int64      `json:"user_id"`
	Content        string     `json:"content,omitempty"`
	Name           string     `json:"name,omitempty"`
	ProfilePicture string     `json:"profile_picture,omitempty"`
	CreatedAt      *time.Time `json:"created_at,omitempty"`
	TypeMessage    string     `json:"type_message,omitempty"`
	IsAccepted     bool       `json:"is_accepted,omitempty"`
}

type UpdateFreightMessage struct {
	AdvertisementId         int64   `json:"advertisement_id"`
	Latitude                float64 `json:"latitude"`
	Longitude               float64 `json:"longitude"`
	DurationText            string  `json:"duration"`
	DistanceText            string  `json:"distance"`
	DriverName              string  `json:"driver_name"`
	TractorUnitLicensePlate string  `json:"tractor_unit_license_plate"`
	TrailerLicensePlate     string  `json:"trailer_license_p_late"`
	TypeMessage             string  `json:"type_message"`
}

func (c *Client) writeMessage() {
	defer func() {
		err := c.Conn.Close()
		if err != nil {
			return
		}
	}()

	for {
		message, ok := <-c.Message
		if !ok {
			return
		}

		err := c.Conn.WriteJSON(message)
		if err != nil {
			return
		}
	}
}

func (c *Client) readMessage(hub *Hub, s InterfaceService) {
	defer func() {
		hub.Unregister <- c
		err := c.Conn.Close()
		if err != nil {
			return
		}
	}()

	for {
		_, m, err := c.Conn.ReadMessage()
		var msg *Message

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {

			}
			break
		}

		err = json.Unmarshal(m, &msg)

		if err != nil {
			log.Println("error to unmarshal message")
			continue
		}

		if _, ok := hub.Rooms[msg.RoomId]; !ok {
			var room Room
			room, err = s.GetRoomService(context.Background(), msg.RoomId)

			if err != nil {
				continue
			}
			hub.Rooms[msg.RoomId] = &room
		}

		if _, ok := hub.Rooms[msg.RoomId].Participants[c.UserId]; !ok {
			continue
		}

		data, err := s.CreateChatMessageService(context.Background(), msg, c)

		if err != nil {
			log.Println("error in create chat message service")
			continue
		}

		outgoingMessage := &OutgoingMessage{
			MessageId:      data.ID,
			RoomId:         msg.RoomId,
			UserId:         c.UserId,
			Content:        msg.Content,
			Name:           c.Name,
			ProfilePicture: c.ProfilePicture,
			CreatedAt:      &data.CreatedAt,
			TypeMessage:    data.TypeMessage.String,
		}

		hub.Broadcast <- outgoingMessage
	}
}
