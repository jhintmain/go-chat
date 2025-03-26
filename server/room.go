package server

import (
	"github.com/gorilla/websocket"
	"sync"
)

type Client struct {
	Conn     *websocket.Conn
	NickName string
	Send     chan []byte
}

type Room struct {
	ID        string
	Clients   map[*Client]bool
	Broadcast chan []byte
	Mutex     sync.Mutex
}

func (r *Room) run() {
	for msg := range r.Broadcast {
		r.Mutex.Lock()
		for client, _ := range r.Clients {
			select {
			case client.Send <- msg:
			default:
				close(client.Send)
				delete(r.Clients, client)
			}
		}
		r.Mutex.Unlock()
	}
}

var Rooms map[string]*Room
var roomMutex sync.Mutex

func GetRoom(roomID string) *Room {
	roomMutex.Lock()
	defer roomMutex.Unlock()

	if _, exist := Rooms[roomID]; exist == false {
		Rooms[roomID] = &Room{
			ID:        roomID,
			Clients:   make(map[*Client]bool),
			Broadcast: make(chan []byte),
		}
		go Rooms[roomID].run()
	}
	return Rooms[roomID]
}
