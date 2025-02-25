package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var room = map[string]*Room{
	"all": {
		clients:   make(map[*Client]bool),
		broadcast: make(chan Message),
	},
}
var roomMutex sync.Mutex

type Room struct {
	clients   map[*Client]bool
	broadcast chan Message
	mutex     sync.Mutex
}

type Client struct {
	conn *websocket.Conn
	send chan Message
	room string
}

type Message struct {
	nickname string
	text     string
}

type ConInfo struct {
	room     string
	nickname string
}

func handleChatRoom(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrader.Upgrade err:", err)
	}
	defer conn.Close()

	var conInfo ConInfo
	if err := conn.ReadJSON(conInfo); err != nil {
		log.Println("conn.ReadJSON err:", err)
	}

	client := &Client{
		conn: conn,
		send: make(chan Message),
		room: conInfo.room,
	}

	chatRoom := getRoom(conInfo.room)

	chatRoom.mutex.Lock()
	chatRoom.clients[client] = true
	chatRoom.mutex.Unlock()

	go client.writePump()
	client.readPump(chatRoom)

}

func (c *Client) writePump() {
	for msg := range c.send {
		if err := c.conn.WriteJSON(msg); err != nil {
			log.Println("c.conn.WriteJSON err:", err)
		}
	}
}
func (c *Client) readPump(chatRoom *Room) {
	defer func() {
		chatRoom.mutex.Lock()
		delete(chatRoom.clients, c)
		chatRoom.mutex.Unlock()
		c.conn.Close()
	}()

	for {
		var msg Message
		if err := c.conn.ReadJSON(&msg); err != nil {
			log.Println("c.conn.ReadJSON err:", err)
			break
		}

		chatRoom.broadcast <- msg
	}
}

func getRoom(roomId string) *Room {
	roomMutex.Lock()
	defer roomMutex.Unlock()

	if _, exist := room[roomId]; exist == false {
		room[roomId] = &Room{
			clients:   make(map[*Client]bool),
			broadcast: make(chan Message),
		}
	}
	return room[roomId]
}

func main() {

	http.HandleFunc("/wc", handleChatRoom)
	http.Handle("/", http.FileServer(http.Dir("./html")))

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
