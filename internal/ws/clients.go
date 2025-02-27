package ws

import (
	"encoding/json"
	"geolocation/internal/get_token"
	"github.com/gorilla/websocket"
	"log"
)

type Client struct {
	Conn           *websocket.Conn          `json:"conn"`
	Message        chan *Message            `json:"message"`
	UserId         int64                    `json:"user_id"`
	Name           string                   `json:"name"`
	ProfilePicture string                   `json:"profile_picture"`
	Payload        get_token.PayloadUserDTO `json:"payload"`
}

type Message struct {
	MessageId      int64  `json:"message_id"`
	ChatId         int64  `json:"chat_id"`
	Name           string `json:"name"`
	ProfilePicture string `json:"profile_picture"`
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

func (c *Client) readMessage(hub *Hub) {
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

		if _, ok := hub.Rooms[msg.ChatId]; !ok {

		}

		hub.Broadcast <- msg
	}
}
