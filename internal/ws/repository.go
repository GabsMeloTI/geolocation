package ws

import (
	"context"
	"database/sql"
	db "geolocation/db/sqlc"
)

type InterfaceRepository interface {
	CreateChatRoomRepository(ctx context.Context, params db.CreateChatRoomParams) (db.ChatRoom, error)
	CreateChatMessageRepository(ctx context.Context, params db.CreateChatMessageParams) (db.ChatMessage, error)
	GetChatRoomByIdRepository(ctx context.Context, id int64) (db.GetChatRoomByIdRow, error)
	GetInterestedChatRoomsRepository(ctx context.Context, id int64) ([]db.GetInterestedChatRoomsRow, error)
	GetAdvertisementChatRoomsRepository(ctx context.Context, arg int64) ([]db.GetAdvertisementChatRoomsRow, error)
	GetChatMessagesByRoomIdRepository(ctx context.Context, arg db.GetChatMessagesByRoomIdParams) ([]db.GetChatMessagesByRoomIdRow, error)
	GetLastChatMessageRepository(ctx context.Context, userId int64) ([]db.GetLastMessageByRoomIdRow, error)
	GetRoomByMessageIdRepository(ctx context.Context, messageId int64) (db.GetRoomByMessageIdRow, error)
	UpdateMessageStatusRepository(ctx context.Context, arg db.UpdateMessageStatusParams) error
	CreateOfferRepository(ctx context.Context, arg db.CreateOfferParams) (db.Offer, error)
	UpdateAdvertisementSituationRepository(ctx context.Context, arg db.UpdateAdvertisementSituationParams) error
	CreateTruckRepository(ctx context.Context, arg db.CreateTruckParams) (db.Truck, error)
}

type Repository struct {
	Conn    *sql.DB
	DBtx    db.DBTX
	Queries *db.Queries
	SqlConn *sql.DB
}

func NewWsRepository(Conn *sql.DB) *Repository {
	q := db.New(Conn)
	return &Repository{
		Conn:    Conn,
		DBtx:    Conn,
		Queries: q,
		SqlConn: Conn,
	}
}

func (r *Repository) CreateChatRoomRepository(ctx context.Context, params db.CreateChatRoomParams) (db.ChatRoom, error) {
	return r.Queries.CreateChatRoom(ctx, params)
}

func (r *Repository) CreateChatMessageRepository(ctx context.Context, params db.CreateChatMessageParams) (db.ChatMessage, error) {
	return r.Queries.CreateChatMessage(ctx, params)
}

func (r *Repository) GetChatRoomByIdRepository(ctx context.Context, id int64) (db.GetChatRoomByIdRow, error) {
	return r.Queries.GetChatRoomById(ctx, id)
}

func (r *Repository) GetInterestedChatRoomsRepository(ctx context.Context, id int64) ([]db.GetInterestedChatRoomsRow, error) {
	return r.Queries.GetInterestedChatRooms(ctx, id)
}

func (r *Repository) GetAdvertisementChatRoomsRepository(ctx context.Context, arg int64) ([]db.GetAdvertisementChatRoomsRow, error) {
	return r.Queries.GetAdvertisementChatRooms(ctx, arg)
}

func (r *Repository) GetChatMessagesByRoomIdRepository(ctx context.Context, arg db.GetChatMessagesByRoomIdParams) ([]db.GetChatMessagesByRoomIdRow, error) {
	return r.Queries.GetChatMessagesByRoomId(ctx, arg)
}

func (r *Repository) GetLastChatMessageRepository(ctx context.Context, userId int64) ([]db.GetLastMessageByRoomIdRow, error) {
	return r.Queries.GetLastMessageByRoomId(ctx, userId)
}

func (r *Repository) GetRoomByMessageIdRepository(ctx context.Context, messageId int64) (db.GetRoomByMessageIdRow, error) {
	return r.Queries.GetRoomByMessageId(ctx, messageId)
}

func (r *Repository) UpdateMessageStatusRepository(ctx context.Context, arg db.UpdateMessageStatusParams) error {
	return r.Queries.UpdateMessageStatus(ctx, arg)
}

func (r *Repository) CreateOfferRepository(ctx context.Context, arg db.CreateOfferParams) (db.Offer, error) {
	return r.Queries.CreateOffer(ctx, arg)
}

func (r *Repository) UpdateAdvertisementSituationRepository(ctx context.Context, arg db.UpdateAdvertisementSituationParams) error {
	return r.Queries.UpdateAdvertisementSituation(ctx, arg)
}

func (r *Repository) CreateTruckRepository(ctx context.Context, arg db.CreateTruckParams) (db.Truck, error) {
	return r.Queries.CreateTruck(ctx, arg)
}
