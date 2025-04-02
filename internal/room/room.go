package room

import (
	internalClient "go-chat/internal/client"
	"go-chat/internal/dto"
	"sync"
)

var (
	RoomList = make(map[string]*Room)
	mtx      sync.Mutex
)

type Room struct {
	ID        string
	Clients   map[string]*internalClient.Client
	Broadcast chan dto.Message
	Mutex     *sync.Mutex
}

func NewRoom(roomID string) *Room {
	mtx.Lock()
	defer mtx.Unlock()

	room := &Room{
		ID:        roomID,
		Clients:   make(map[string]*internalClient.Client),
		Broadcast: make(chan dto.Message),
		Mutex:     &sync.Mutex{},
	}
	RoomList[room.ID] = room
	go room.run()
	return room
}

func GetRoom(roomID string) *Room {
	if _, exist := RoomList[roomID]; exist == false {
		return NewRoom(roomID)
	}
	return RoomList[roomID]
}

func (r *Room) run() {
	for msg := range r.Broadcast {
		r.Mutex.Lock()
		for uuid, client := range r.Clients {
			select {
			case client.Send <- msg:
			default:
				close(client.Send)
				delete(r.Clients, uuid)
			}
		}
		r.Mutex.Unlock()
	}
}
