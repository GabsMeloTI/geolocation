package ws

import "sync"

type Room struct {
	ID              int64 `json:"id"`
	AdvertisementId int64 `json:"announcement_id"`
	Participants    map[int64]bool
}

type Hub struct {
	Rooms      map[int64]*Room
	Clients    map[int64]*Client
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan *OutgoingMessage
	Mu         *sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		Rooms:      make(map[int64]*Room),
		Clients:    make(map[int64]*Client),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan *OutgoingMessage, 5),
		Mu:         &sync.RWMutex{},
	}
}

func (h *Hub) Run() {
	for {
		select {
		case cl := <-h.Register:
			h.Mu.Lock()
			if _, ok := h.Clients[cl.UserId]; !ok {
				h.Clients[cl.UserId] = cl
			}
			h.Mu.Unlock()

		case cl := <-h.Unregister:
			h.Mu.Lock()
			if _, ok := h.Clients[cl.UserId]; ok {
				delete(h.Clients, cl.UserId)
				close(cl.Message)
			}
			h.Mu.Unlock()

		case m := <-h.Broadcast:
			h.Mu.Lock()
			if _, ok := h.Rooms[m.RoomId]; ok {
			}
			h.Mu.Unlock()

		}

	}
}
