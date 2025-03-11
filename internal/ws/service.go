package ws

import (
	"context"
	"database/sql"
	"errors"
	db "geolocation/db/sqlc"
	"geolocation/internal/advertisement"
	"geolocation/internal/get_token"
	new_routes "geolocation/internal/new_routes"
)

type InterfaceService interface {
	CreateChatRoomService(ctx context.Context, data CreateChatRoomRequest, userId int64) (CreateChatRoomResponse, error)
	CreateChatMessageService(ctx context.Context, msg *Message, cl *Client) (db.ChatMessage, error)
	GetRoomService(ctx context.Context, id int64) (Room, error)
	GetHomeService(ctx context.Context, payload get_token.PayloadUserDTO) (HomeResponse, error)
	GetChatMessagesByRoomIdService(ctx context.Context, roomId int64, userId int64) ([]MessageResponse, error)
	UpdateMessageOfferService(ctx context.Context, data UpdateOfferDTO, hub *Hub) error
	FreightLocationDetailsService(ctx context.Context, data UpdateFreightData, userId int64) (FreightLocationDetailsResponse, error)
}

type Service struct {
	InterfaceService       InterfaceRepository
	ServiceRoutes          new_routes.InterfaceService
	InterfaceAdvertisement advertisement.InterfaceRepository
}

func NewWsService(interfaceService InterfaceRepository, InterfaceAdvertisement advertisement.InterfaceRepository, ServiceRoutes new_routes.InterfaceService,
) *Service {
	return &Service{
		InterfaceService:       interfaceService,
		InterfaceAdvertisement: InterfaceAdvertisement,
		ServiceRoutes:          ServiceRoutes,
	}
}

func (s *Service) CreateChatRoomService(ctx context.Context, data CreateChatRoomRequest, userId int64) (CreateChatRoomResponse, error) {

	a, err := s.InterfaceAdvertisement.GetAdvertisementById(ctx, data.AdvertisementID)

	if err != nil {
		return CreateChatRoomResponse{}, err
	}

	if a.UserID == userId {
		return CreateChatRoomResponse{}, errors.New("you cannot create a room to your own advertisement")
	}

	ok, err := s.InterfaceService.GetChatRoomByAdvertisementAndInterestedUserRepository(ctx, db.GetChatRoomByAdvertisementAndInterestedUserParams{
		AdvertisementID:  data.AdvertisementID,
		InterestedUserID: userId,
	})

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return CreateChatRoomResponse{}, err
		}
	}

	if ok.ID != 0 {
		return CreateChatRoomResponse{}, errors.New("chat room already exists")
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
		TypeMessage: sql.NullString{
			String: msg.TypeMessage,
			Valid:  msg.TypeMessage != "",
		},
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
		var lastMessage *MessageResponse
		if l, ok := lastMessagesMap[i.RoomID]; ok {
			lastMessage = &l
		}
		res.Interested = append(res.Interested, RoomResponse{
			RoomId:              i.RoomID,
			Title:               i.Title,
			Origin:              i.Origin,
			Destination:         i.Destination,
			AdvertisementUserId: i.AdvertisementUserID,
			AdvertisementId:     i.AdvertisementID,
			Distance:            i.Distance,
			LastMessage:         lastMessage,
		})
	}

	advertisementChatRooms, err := s.InterfaceService.GetAdvertisementChatRoomsRepository(ctx, payload.ID)

	if err != nil {
		return res, err
	}

	for _, a := range advertisementChatRooms {
		var lastMessage *MessageResponse
		if l, ok := lastMessagesMap[a.RoomID]; ok {
			lastMessage = &l
		}
		res.Advertisement = append(res.Advertisement, RoomResponse{
			RoomId:              a.RoomID,
			Title:               a.Title,
			Origin:              a.Origin,
			Destination:         a.Destination,
			AdvertisementUserId: a.AdvertisementUserID,
			AdvertisementId:     a.AdvertisementID,
			Distance:            a.Distance,
			LastMessage:         lastMessage,
		})
	}

	activeFreights, err := s.InterfaceService.GetAllActiveFreightsRepository(ctx, payload.ID)

	if err != nil {
		return res, err
	}

	if len(activeFreights) > 0 {
		for _, f := range activeFreights {
			res.ActiveFreights = append(res.ActiveFreights, ActiveFreight{
				AdvertisementId:         f.AdvertisementID,
				Latitude:                f.Latitude,
				Longitude:               f.Longitude,
				DurationText:            f.Duration,
				DistanceText:            f.Distance,
				DriverName:              f.DriverName,
				TractorUnitLicensePlate: f.TractorUnitLicensePlate.String,
				TrailerLicensePlate:     f.TrailerLicensePlate.String,
			})
		}
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
			TypeMessage:    m.TypeMessage.String,
		}
	}

	return messages, nil
}

func (s *Service) UpdateMessageOfferService(ctx context.Context, data UpdateOfferDTO, hub *Hub) error {
	r, err := s.InterfaceService.GetRoomByMessageIdRepository(ctx, data.Request.MessageId)

	if err != nil {
		return err
	}

	if r.AdvertisementUserID != data.Payload.ID || r.TypeMessage.String != "offer" || r.IsAccepted.Bool {
		return errors.New("invalid offer")
	}

	err = s.InterfaceService.UpdateMessageStatusRepository(ctx, db.UpdateMessageStatusParams{
		IsAccepted: sql.NullBool{
			Bool:  data.Request.IsAccepted,
			Valid: true,
		},
		ID: data.Request.MessageId,
	})

	if err != nil {
		return err
	}

	offer, err := s.InterfaceService.CreateOfferRepository(ctx, data.ToCreateOfferParams(r.InterestedUserID))

	if err != nil {
		return err
	}

	err = s.InterfaceService.UpdateAdvertisementSituationRepository(ctx, data.ToUpdateAdvertisementSituationParams())

	if err != nil {
		return err
	}

	truck, err := s.InterfaceService.CreateTruckRepository(ctx, data.ToCreateTruckParams())

	if err != nil {
		return err
	}

	_, err = s.InterfaceService.CreateAppointmentRepository(ctx, data.ToCreateAppointmentParams(r.AdvertisementUserID, r.InterestedUserID, offer.ID, truck.ID))

	if err != nil {
		return err
	}

	msg := &OutgoingMessage{
		MessageId:   data.Request.MessageId,
		RoomId:      r.ID,
		UserId:      data.Payload.ID,
		TypeMessage: "offer",
		IsAccepted:  data.Request.IsAccepted,
	}

	if cl, ok := hub.Clients[r.InterestedUserID]; ok {
		cl.Message <- msg
	}

	return nil
}

func (s *Service) FreightLocationDetailsService(ctx context.Context, data UpdateFreightData, userId int64) (FreightLocationDetailsResponse, error) {
	freightDetails, err := s.InterfaceService.GetAppointmentDetailsByAdvertisementIdRepository(ctx, data.AdvertisementId)

	if err != nil {
		return FreightLocationDetailsResponse{}, err
	}

	if freightDetails.InterestedUserID.Int64 != userId {
		return FreightLocationDetailsResponse{}, errors.New("invalid user id")
	}

	route, err := s.ServiceRoutes.GetSimpleRoute(new_routes.SimpleRouteRequest{
		OriginLat: data.OriginLatitude,
		OriginLng: data.OriginLongitude,
		DestLat:   freightDetails.DestinationLat.Float64,
		DestLng:   freightDetails.DestinationLng.Float64,
	})

	if err != nil {
		return FreightLocationDetailsResponse{}, err
	}

	activeFreight, err := s.InterfaceService.GetActiveFreightRepository(ctx, data.AdvertisementId)

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return FreightLocationDetailsResponse{}, err
		}
		if errors.Is(err, sql.ErrNoRows) {
			err = s.InterfaceService.CreateActiveFreightRepository(ctx, data.ToCreateActiveFreightParams(route, freightDetails))

			if err != nil {
				return FreightLocationDetailsResponse{}, err
			}
		}
	}

	if activeFreight.ID != 0 {
		err = s.InterfaceService.UpdateActiveFreightRepository(ctx, data.ToUpdateActiveFreightParams(route, activeFreight.ID))

		if err != nil {
			return FreightLocationDetailsResponse{}, err
		}
	}

	return FreightLocationDetailsResponse{
		DurationText:            route.Summary.SimpleRoute.Duration.Text,
		DistanceText:            route.Summary.SimpleRoute.Distance.Text,
		DriverName:              freightDetails.Name,
		AdvertisementUserId:     freightDetails.AdvertisementUserID.Int64,
		TractorUnitLicensePlate: freightDetails.TractorUnitLicensePlate.String,
		TrailerLicensePlate:     freightDetails.TrailerLicensePlate.String,
	}, nil
}
