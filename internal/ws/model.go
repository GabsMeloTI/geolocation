package ws

import (
	db "geolocation/db/sqlc"
	"time"
)

type CreateChatRoomRequest struct {
	AdvertisementID int64 `json:"advertisement_id"`
}

type CreateChatRoomResponse struct {
	ID                  int64     `json:"id"`
	AdvertisementID     int64     `json:"advertisement_id"`
	AdvertisementUserID int64     `json:"advertisement_user_id"`
	InterestedUserID    int64     `json:"interested_user_id"`
	CreatedAt           time.Time `json:"created_at"`
}

type HomeResponse struct {
	Advertisement []RoomResponse `json:"advertisements"`
	Interested    []RoomResponse `json:"interested"`
}

type RoomResponse struct {
	RoomId              int64            `json:"room_id"`
	Title               string           `json:"title"`
	Origin              string           `json:"origin"`
	Destination         string           `json:"destination"`
	AdvertisementUserId int64            `json:"advertisement_user_id"`
	AdvertisementId     int64            `json:"advertisement_id"`
	Distance            int64            `json:"distance"`
	LastMessage         *MessageResponse `json:"last_message,omitempty"`
}

type MessageResponse struct {
	MessageId      int64     `json:"message_id"`
	RoomId         int64     `json:"room_id"`
	UserId         int64     `json:"user_id"`
	Content        string    `json:"content"`
	Name           string    `json:"name"`
	ProfilePicture string    `json:"profile_picture"`
	CreatedAt      time.Time `json:"created_at"`
}

func (r CreateChatRoomRequest) ParseToCreateChatRoomResponse(room db.ChatRoom) CreateChatRoomResponse {
	return CreateChatRoomResponse{
		ID:                  room.ID,
		AdvertisementID:     room.AdvertisementID,
		AdvertisementUserID: room.AdvertisementUserID,
		InterestedUserID:    room.InterestedUserID,
		CreatedAt:           room.CreatedAt,
	}
}
