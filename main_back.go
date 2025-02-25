package main

//
//import (
//	"fmt"
//	"github.com/gorilla/websocket"
//	"net/http"
//	"sync"
//)
//
//type Type string
//
//const (
//	MESSAGE Type = "message"
//	SYSTEM  Type = "system"
//	USERS   Type = "users"
//)
//
//type Client struct {
//	conn     *websocket.Conn
//	nickname string
//}
//type Message struct {
//	Type     Type     `json:"type"`
//	NickName string   `json:"nickname"`
//	Text     string   `json:"text"`
//	UserList []string `json:"userList,omitempty"`
//}
//
//func (m *Message) ToSystemMessage(div string) *Message {
//	convertMsg := Message{
//		Type:     SYSTEM,
//		NickName: m.NickName,
//	}
//
//	switch div {
//	case "join":
//		convertMsg.Text = fmt.Sprintf("%s 님이 입장하셨습니다.", m.NickName)
//		break
//	case "leave":
//		convertMsg.Text = fmt.Sprintf("%s 님이 떠났습니다.", m.NickName)
//		break
//	}
//	return &convertMsg
//}
//
//var (
//	//Request > Websocket
//	upgrader = websocket.Upgrader{
//		CheckOrigin: func(r *http.Request) bool {
//			return true
//		},
//	}
//	nicknames = make(map[string]bool)
//	// 클라이언드 목록 생성
//	clients = make(map[*Client]bool)
//	// 메세지 체널 생성
//	broadcast = make(chan Message)
//	mutex     = sync.Mutex{}
//)
//
//// Websocket connection
//func handleConnection(w http.ResponseWriter, r *http.Request) {
//	conn, err := upgrader.Upgrade(w, r, nil)
//	if err != nil {
//		fmt.Println("Web socket Upgrade fail", err)
//		return
//	}
//
//	defer conn.Close()
//
//	var msg Message
//	if err := conn.ReadJSON(&msg); err != nil {
//		fmt.Println("Web socket Read Message fail", err)
//		return
//	}
//	fmt.Println(msg)
//
//	client := &Client{conn: conn, nickname: msg.NickName}
//
//	mutex.Lock()
//	clients[client] = true
//	nicknames[msg.NickName] = true
//	broadcast <- *msg.ToSystemMessage("join")
//	mutex.Unlock()
//
//	sendUserList()
//
//	for {
//		if err := conn.ReadJSON(&msg); err != nil {
//			fmt.Println("Web socket Read Message fail", err)
//			break
//		}
//		msg.NickName = client.nickname
//		msg.Type = MESSAGE
//		broadcast <- msg
//	}
//
//	mutex.Lock()
//	delete(clients, client)
//	delete(nicknames, client.nickname)
//	mutex.Unlock()
//
//	broadcast <- *msg.ToSystemMessage("leave")
//}
//
//// broadcast message
//func handleMessages() {
//	for {
//		msg := <-broadcast
//
//		mutex.Lock()
//		for client := range clients {
//			if err := client.conn.WriteJSON(msg); err != nil {
//				fmt.Println("Write message fail", err)
//				client.conn.Close()
//				delete(clients, client)
//				delete(nicknames, client.nickname)
//			}
//		}
//		mutex.Unlock()
//	}
//}
//
//func sendUserList() {
//	mutex.Lock()
//	var users []string
//	for nickname := range nicknames {
//		users = append(users, nickname)
//	}
//	mutex.Unlock()
//
//	broadcast <- Message{
//		Type:     USERS,
//		UserList: users,
//	}
//}
//
//func main() {
//	go handleMessages()
//
//	http.HandleFunc("/ws", handleConnection)
//
//	http.Handle("/", http.FileServer(http.Dir("./html")))
//
//	fmt.Println("Run WebSocket Server")
//
//	err := http.ListenAndServe(":8080", nil)
//	if err != nil {
//		fmt.Println(err)
//	}
//}
