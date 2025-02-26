package main

import (
	"fmt"
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

var room = make(map[string]*ChatRoom)
var roomMutex sync.Mutex

type ChatRoom struct {
	clients   map[*Client]bool
	broadcast chan Message
	mutex     sync.Mutex
}

func (room *ChatRoom) run() {
	for msg := range room.broadcast {

		room.mutex.Lock()
		for client, _ := range room.clients {
			select {
			case client.send <- msg:
				fmt.Printf("client.send [%v] msg :%v", client.room, msg)
			default:
				delete(room.clients, client)
				close(client.send)
			}
		}
		room.mutex.Unlock()
	}
}

type Client struct {
	conn *websocket.Conn
	send chan Message
	room string
}

type Message struct {
	Nickname string `json:"nickname"`
	Text     string `json:"text"`
}

type ConInfo struct {
	Room     string `json:"room"`
	Nickname string `json:"nickname"`
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrader.Upgrade err:", err)
	}
	defer conn.Close()

	var conInfo ConInfo
	if err := conn.ReadJSON(&conInfo); err != nil {
		log.Println("conn.ReadJSON err:", err)
	}

	client := &Client{
		conn: conn,
		send: make(chan Message),
		room: conInfo.Room,
	}
	fmt.Println("ConInfo", conInfo)
	room := getRoom(conInfo.Room)

	room.mutex.Lock()
	room.clients[client] = true
	room.mutex.Unlock()

	go client.writePump()
	client.readPump(room)

}

func getRoom(roomID string) *ChatRoom {
	roomMutex.Lock()
	defer roomMutex.Unlock()

	if _, exist := room[roomID]; exist == false {
		room[roomID] = &ChatRoom{
			clients:   make(map[*Client]bool),
			broadcast: make(chan Message),
		}
		go room[roomID].run()
	}
	return room[roomID]
}

func (c *Client) readPump(room *ChatRoom) {
	defer func() {
		room.mutex.Lock()
		delete(room.clients, c)
		room.mutex.Unlock()
		close(c.send)
		c.conn.Close()
	}()

	for {
		var msg Message
		if err := c.conn.ReadJSON(&msg); err != nil {
			log.Println("c.conn.ReadJSON err:", err)
			break
		}
		fmt.Printf("readPump [%v] msg : %v", c.room, msg)
		room.broadcast <- msg
	}
}

func (c *Client) writePump() {
	for msg := range c.send {
		fmt.Println("writePump msg", msg)
		if err := c.conn.WriteJSON(msg); err != nil {
			log.Println("c.conn.WriteJSON err:", err)
			break
		}
	}
}

func main() {
	http.HandleFunc("/ws", handleConnections)
	http.Handle("/", http.FileServer(http.Dir("./html")))

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
