package ws

import (
	"context"
	"database/sql"
	db "geolocation/db/sqlc"
	"geolocation/internal/advertisement"
	"geolocation/internal/get_token"
)

type InterfaceService interface {
	CreateChatRoomService(ctx context.Context, data CreateChatRoomRequest, userId int64) (CreateChatRoomResponse, error)
	CreateChatMessageService(ctx context.Context, msg *Message, cl *Client) (db.ChatMessage, error)
	GetRoomService(ctx context.Context, id int64) (Room, error)
}

type Service struct {
	InterfaceService       InterfaceRepository
	InterfaceAdvertisement advertisement.InterfaceRepository
}

func NewWsService(interfaceService InterfaceRepository, InterfaceAdvertisement advertisement.InterfaceRepository) *Service {
	return &Service{
		InterfaceService:       interfaceService,
		InterfaceAdvertisement: InterfaceAdvertisement,
	}
}

func (s *Service) CreateChatRoomService(ctx context.Context, data CreateChatRoomRequest, userId int64) (CreateChatRoomResponse, error) {

	a, err := s.InterfaceAdvertisement.GetAdvertisementById(ctx, data.AdvertisementID)

	if err != nil {
		return CreateChatRoomResponse{}, err
	}

	chatRoom, err := s.InterfaceService.CreateChatRoomRepository(ctx, db.CreateChatRoomParams{
		AdvertisementID:     a.ID,
		AdvertisementUserID: a.UserID,
		InterestedUserID:    userId,
	})

	if err != nil {
		return CreateChatRoomResponse{}, err
	}

	res := data.ParseToCreateChatRoomResponse(chatRoom)

	return res, nil
}

func (s *Service) CreateChatMessageService(ctx context.Context, msg *Message, cl *Client) (db.ChatMessage, error) {
	chatMessage, err := s.InterfaceService.CreateChatMessageRepository(ctx, db.CreateChatMessageParams{
		RoomID: sql.NullInt64{
			Int64: msg.RoomId,
			Valid: true,
		},
		UserID: sql.NullInt64{
			Int64: cl.UserId,
			Valid: true,
		},
		Content: msg.Content,
	})

	if err != nil {
		return db.ChatMessage{}, err
	}

	return chatMessage, nil
}

func (s *Service) GetRoomService(ctx context.Context, id int64) (Room, error) {
	chatRoom, err := s.InterfaceService.GetChatRoomByIdRepository(ctx, id)

	if err != nil {
		return Room{}, err
	}

	room := Room{
		ID:              chatRoom.ID,
		AdvertisementId: chatRoom.AdvertisementID,
		Participants:    make(map[int64]bool),
	}

	room.Participants[chatRoom.InterestedUserID] = true
	room.Participants[chatRoom.AdvertisementUserID] = true

	return room, nil

}

func (s *Service) GetHomeService(ctx context.Context, payload get_token.PayloadUserDTO) error {
	
}
