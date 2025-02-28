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
	GetDriverChatRoomsRepository(ctx context.Context, id int64) ([]db.GetDriverChatRoomsRow, error)
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

func (r *Repository) GetDriverChatRoomsRepository(ctx context.Context, id int64) ([]db.GetDriverChatRoomsRow, error) {
	return r.Queries.GetDriverChatRooms(ctx, id)
}
