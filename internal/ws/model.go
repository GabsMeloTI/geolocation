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

func (r CreateChatRoomRequest) ParseToCreateChatRoomResponse(room db.ChatRoom) CreateChatRoomResponse {
	return CreateChatRoomResponse{
		ID:                  room.ID,
		AdvertisementID:     room.AdvertisementID,
		AdvertisementUserID: room.AdvertisementUserID,
		InterestedUserID:    room.InterestedUserID,
		CreatedAt:           room.CreatedAt,
	}
}
