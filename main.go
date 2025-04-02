package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
)

var makeRoom sync.Mutex
var roomList Rooms

type Message struct {
	Div     string      `json:"div"` // CHAT, SYSTEM, ROOMLIST, USERLIST
	Content interface{} `json:"content"`
}
type Client struct {
	Conn     *websocket.Conn
	NickName string
	Send     chan Message
	Mutex    sync.Mutex
}

func (c *Client) joinRoom(room *Room) error {
	room.Mutex.Lock()
	room.Clients[c] = true
	defer room.Mutex.Unlock()

	room.Broadcast <- Message{
		Div:     "SYSTEM",
		Content: c.NickName + " joined the room",
	}

	room.SendUserList()
	return nil
}

func (c *Client) leaveRoom(room *Room) {
	room.Mutex.Lock()
	delete(room.Clients, c)
	defer room.Mutex.Unlock()
	room.Broadcast <- Message{
		Div:     "SYSTEM",
		Content: c.NickName + " left the room",
	}
	room.SendUserList()
}

func (c *Client) write() {
	for msg := range c.Send {
		if err := c.Conn.WriteJSON(&msg); err != nil {
			break
		}
	}
}
func (c *Client) read(room *Room) {
	defer func() {
		room.Mutex.Lock()
		delete(room.Clients, c) // room에서 client 제거
		room.Mutex.Unlock()
		close(c.Send) // 채널닫기
		c.Conn.Close()
	}()

	for {
		var msg Message
		if err := c.Conn.ReadJSON(&msg); err != nil {
			log.Println("c.conn.ReadMessage err:", err)
			break
		}
		room.Broadcast <- msg
	}
}

type Rooms map[string]*Room

type Room struct {
	ID        string
	Clients   map[*Client]bool
	Broadcast chan Message
	Mutex     sync.Mutex
}

// websocket 업그레이더
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 모든 접속 허용
	},
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	roomID := query.Get("room")
	nickName := query.Get("nickname")

	if roomID == "" {
		http.Error(w, "room id is required", http.StatusBadRequest)
		return
	}
	if nickName == "" {
		http.Error(w, "nickName is required", http.StatusBadRequest)
		return
	}

	// http > socket으로
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	client := Client{
		Conn:     conn,
		NickName: nickName,
		Send:     make(chan Message),
	}
	room := GetRoom(roomID)
	if err := client.joinRoom(room); err != nil {
		http.Error(w, "room not found", http.StatusNotFound)
		return
	}

	go client.write()
	client.SendRoomList()
	client.read(room)
}

func main() {
	http.HandleFunc("/ws", handleWebSocket)
	http.Handle("/", http.FileServer(http.Dir("./public")))

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

func GetRoom(roomID string) *Room {
	makeRoom.Lock()
	defer makeRoom.Unlock()

	if _, exist := roomList[roomID]; exist == false {
		roomList[roomID] = &Room{
			ID:        roomID,
			Clients:   make(map[*Client]bool),
			Broadcast: make(chan Message),
		}
		go roomList[roomID].run()
	}
	roomList[roomID].SendUserList()
	return roomList[roomID]
}

func (r *Room) run() {
	for msg := range r.Broadcast {
		r.Mutex.Lock()

		switch msg.Div {
		case "CHAT":
			// CHAT 메시지 처리
			r.handleChat(msg)

		case "SYSTEM":
			// SYSTEM 메시지 처리
			r.handleSystem(msg)

		case "USERLIST":
			// USERLIST 메시지 처리
			r.handleUserList(msg)

		default:
			// 그 외 메시지 처리
			r.handleDefault(msg)
		}

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

// "SYSTEM" 메시지 처리
func (r *Room) handleSystem(msg Message) {
	// 시스템 메시지일 때 모든 클라이언트에게 시스템 메시지 보내기
	for clientID, client := range r.Clients {
		select {
		case client.Send <- msg:
		default:
			close(client.Send)
			delete(r.Clients, clientID)
		}
	}
}

// "USERLIST" 메시지 처리
func (r *Room) handleUserList(msg Message) {
	// 유저 리스트 메시지일 때, 방의 유저 목록을 클라이언트들에게 전송
	userList := getUserList(r) // 방의 유저 목록을 가져오는 함수 호출
	msg.Content = userList     // 유저 목록을 메시지 내용에 추가

	// 방에 있는 모든 클라이언트에게 유저 리스트 메시지 보내기
	for clientID, client := range r.Clients {
		select {
		case client.Send <- msg:
		default:
			close(client.Send)
			delete(r.Clients, clientID)
		}
	}
}

func (r *Room) SendUserList() {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	var list []string
	for client, _ := range r.Clients {
		list = append(list, client.NickName)
	}

	msg := Message{
		Div:     "USERLIST",
		Content: list,
	}

	for client, _ := range r.Clients {
		select {
		case client.Send <- msg:
		default:
			close(client.Send)
			delete(r.Clients, client)
		}
	}
}

func (c *Client) SendRoomList() {
	var list []string
	for roomID, _ := range roomList {
		list = append(list, roomID)
	}

	msg := Message{
		Div:     "ROOMLIST",
		Content: list,
	}

	c.Send <- msg
}
