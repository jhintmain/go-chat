// internal/room/room.go
package room

import (
	"go-chat/internal/dto"
	"go-chat/internal/port"
	"sync"
)

type Room struct {
	ID        string
	Clients   map[string]port.Client
	Broadcast chan dto.Message
	Mutex     *sync.Mutex
}

type RoomManager struct {
	Rooms map[string]*Room
	Mutex *sync.Mutex
}

func NewRoomManager() *RoomManager {
	return &RoomManager{
		Rooms: make(map[string]*Room),
		Mutex: &sync.Mutex{},
	}
}

func (rm *RoomManager) GetRoomList() (roomList []string) {
	for roomID, _ := range rm.Rooms {
		roomList = append(roomList, roomID)
	}
	return roomList
}

func (rm *RoomManager) GetRoom(roomID string) (isNewRoom bool, room *Room) {
	rm.Mutex.Lock()
	defer rm.Mutex.Unlock()

	if room, exists := rm.Rooms[roomID]; exists {
		return false, room
	}
	isNewRoom = true
	room = &Room{
		ID:        roomID,
		Clients:   make(map[string]port.Client),
		Broadcast: make(chan dto.Message),
		Mutex:     &sync.Mutex{},
	}
	rm.Rooms[roomID] = room
	go room.run()
	return isNewRoom, room
}

func (r *Room) run() {
	for msg := range r.Broadcast {
		r.Mutex.Lock()
		for uuid, client := range r.Clients {
			if err := client.SendMessage(msg); err != nil {
				client.Close()
				delete(r.Clients, uuid)
			}
		}
		r.Mutex.Unlock()
	}
}

func (r *Room) Join(c port.Client) {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	r.Clients[c.ID()] = c
}

func (rm *RoomManager) HandleMessage(msg dto.Message) {
	_, room := rm.GetRoom(msg.RoomID)
	room.Broadcast <- msg
}
