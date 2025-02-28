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
	GetHomeService(ctx context.Context, payload get_token.PayloadUserDTO) (HomeResponse, error)
	GetChatMessagesByRoomIdService(ctx context.Context, roomId int64, userId int64) ([]MessageResponse, error)
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

func (s *Service) GetHomeService(ctx context.Context, payload get_token.PayloadUserDTO) (HomeResponse, error) {
	var res HomeResponse

	lastMessages, err := s.InterfaceService.GetLastChatMessageRepository(ctx, payload.ID)

	if err != nil {
		return res, err
	}

	lastMessagesMap := make(map[int64]MessageResponse)

	for _, l := range lastMessages {
		lastMessagesMap[l.RoomID.Int64] = MessageResponse{
			MessageId:      l.ID,
			RoomId:         l.RoomID.Int64,
			UserId:         l.UserID.Int64,
			Content:        l.Content,
			Name:           l.Name,
			ProfilePicture: l.ProfilePicture.String,
			CreatedAt:      l.CreatedAt,
		}
	}

	interestedChatRooms, err := s.InterfaceService.GetInterestedChatRoomsRepository(ctx, payload.ID)

	if err != nil {
		return res, err
	}

	for _, i := range interestedChatRooms {
		res.Interested = append(res.Interested, RoomResponse{
			RoomId:              i.RoomID,
			Title:               i.Title,
			Origin:              i.Origin,
			Destination:         i.Destination,
			AdvertisementUserId: i.AdvertisementUserID,
			AdvertisementId:     i.AdvertisementID,
			Distance:            i.Distance,
			LastMessage:         lastMessagesMap[i.RoomID],
		})
	}

	advertisementChatRooms, err := s.InterfaceService.GetAdvertisementChatRoomsRepository(ctx, payload.ID)

	if err != nil {
		return res, err
	}

	for _, a := range advertisementChatRooms {
		res.Advertisement = append(res.Advertisement, RoomResponse{
			RoomId:              a.RoomID,
			Title:               a.Title,
			Origin:              a.Origin,
			Destination:         a.Destination,
			AdvertisementUserId: a.AdvertisementUserID,
			AdvertisementId:     a.AdvertisementID,
			Distance:            a.Distance,
			LastMessage:         lastMessagesMap[a.RoomID],
		})
	}

	return res, nil
}

func (s *Service) GetChatMessagesByRoomIdService(ctx context.Context, roomId int64, userId int64) ([]MessageResponse, error) {
	chatMessages, err := s.InterfaceService.GetChatMessagesByRoomIdRepository(ctx, db.GetChatMessagesByRoomIdParams{
		RoomID: sql.NullInt64{
			Int64: roomId,
			Valid: true,
		},
		UserID: userId,
	})

	if err != nil {
		return nil, err
	}

	messages := make([]MessageResponse, len(chatMessages))

	for i, m := range chatMessages {
		messages[i] = MessageResponse{
			MessageId:      m.ID,
			RoomId:         m.RoomID.Int64,
			UserId:         m.UserID.Int64,
			Content:        m.Content,
			Name:           m.Name,
			ProfilePicture: m.ProfilePicture.String,
			CreatedAt:      m.CreatedAt,
		}
	}

	return messages, nil
}
