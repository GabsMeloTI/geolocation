package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"

	"geolocation/internal/get_token"
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
	RoomId       int64  `json:"room_id"`
	Content      string `json:"content"`
	TypeMessage  string `json:"type_message"`
	FirstMessage bool   `json:"first_message"`
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

type ReadNotification struct {
	RoomId      int64     `json:"room_id"`
	UserId      int64     `json:"user_id"`
	TypeMessage string    `json:"type_message"`
	ReadAt      time.Time `json:"read_at"`
}

type OfferContent struct {
	TruckId         int64   `json:"truck_id"`
	DriverId        int64   `json:"driver_id"`
	Price           float64 `json:"price"`
	AdvertisementId int64   `json:"advertisement_id"`
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
			if websocket.IsUnexpectedCloseError(
				err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
			) {
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

		if msg.TypeMessage == "count" {
			readAt, err := s.ReadMessagesService(context.Background(), msg, c)
			if err != nil {
				log.Println("error to read msg")
				continue
			}

			for id := range hub.Rooms[msg.RoomId].Participants {
				if id != c.UserId {
					notification := ReadNotification{
						RoomId:      msg.RoomId,
						UserId:      c.UserId,
						TypeMessage: "message_read",
						ReadAt:      readAt,
					}
					err = hub.Clients[id].Conn.WriteJSON(notification)
					if err != nil {
						log.Println("error to write read notification message")
					}
				}
			}
			continue
		}

		if msg.FirstMessage {
			for id := range hub.Rooms[msg.RoomId].Participants {
				if id != c.UserId {
					home, err := s.GetHomeService(context.Background(), hub.Clients[id].Payload)
					if err != nil {
						log.Println("error to get user home")
						continue
					}
					err = hub.Clients[id].Conn.WriteJSON(home)
					if err != nil {
						return
					}
				}
			}
		}

		data, err := s.CreateChatMessageService(context.Background(), msg, c)
		if err != nil {
			log.Println("error in create chat message service")
			continue
		}

		if msg.TypeMessage == "offer" {
			err := s.CreateOfferService(context.Background(), msg, c)
			if err != nil {
				fmt.Println(err)
				log.Println("error to create offer")
				continue
			}
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
