package bak2

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

var chatRoom = make(map[string]*ChatRoom)

var createRoomMutex sync.Mutex

var userList = make(map[string][]string)

var roomList []string

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
			default:
				delete(room.clients, client)
				close(client.send)
			}
		}
		room.mutex.Unlock()
	}
}

type Client struct {
	conn     *websocket.Conn
	send     chan Message
	nickname string
}

func (c *Client) writePump() {
	for msg := range c.send {
		if err := c.conn.WriteJSON(msg); err != nil {
			break
		}
	}
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
		room.broadcast <- msg
	}
}

type Message struct {
	Room        string              `json:"room,required"`
	Nickname    string              `json:"nickname,required"`
	UserList    map[string][]string `json:"userList,omitempty"`
	RoomList    []string            `json:"roomList,omitempty"`
	MessageType string              `json:"messageType"`
	Text        string              `json:"text"`
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrader.Upgrade err:", err)
	}
	defer conn.Close()

	var conInfo Message
	if err := conn.ReadJSON(&conInfo); err != nil {
		log.Println("conn.ReadJSON err:", err)
	}

	client := &Client{
		conn:     conn,
		send:     make(chan Message),
		nickname: conInfo.Nickname,
	}

	var room = getRoom(conInfo.Room)

	room.mutex.Lock()
	room.clients[client] = true
	room.mutex.Unlock()

	room.broadcast <- Message{
		Room:        conInfo.Room,
		Nickname:    conInfo.Nickname,
		UserList:    userList,
		RoomList:    roomList,
		MessageType: "system",
		Text:        client.nickname + "님 입장",
	}

	go client.writePump()
	client.readPump(room)
}

func getRoom(roomID string) *ChatRoom {
	// 방생성 lock & unlock (중복된 방 생성 방지)
	createRoomMutex.Lock()
	defer createRoomMutex.Unlock()

	// 방 미존재시 생성 후 return, 존재시 존재하는 값 return
	if _, exist := chatRoom[roomID]; exist == false {
		chatRoom[roomID] = &ChatRoom{
			clients:   make(map[*Client]bool),
			broadcast: make(chan Message),
		}
		userList[roomID] = make([]string, 0)
		roomList = append(roomList, roomID)
		go chatRoom[roomID].run()
	}
	return chatRoom[roomID]
}

func main() {
	http.HandleFunc("/ws", handleConnections)
	http.Handle("/", http.FileServer(http.Dir("./html")))

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
