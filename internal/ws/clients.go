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
	RoomId  int64  `json:"room_id"`
	Content string `json:"content"`
	Action  string `json:"action"`
}

type OutgoingMessage struct {
	MessageId      int64     `json:"message_id"`
	RoomId         int64     `json:"room_id"`
	UserId         int64     `json:"user_id"`
	Content        string    `json:"content"`
	Name           string    `json:"name"`
	ProfilePicture string    `json:"profile_picture"`
	CreatedAt      time.Time `json:"created_at"`
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
				log.Println("error in get room service")
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
			CreatedAt:      data.CreatedAt,
		}

		hub.Broadcast <- outgoingMessage
	}
}
