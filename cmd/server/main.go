package main

import (
	"github.com/gorilla/websocket"
	pkgClient "go-chat/internal/client"
	"go-chat/internal/dto"
	pkgResponse "go-chat/internal/response"
	pkgRoom "go-chat/internal/room"
	"log"
	"net/http"
)

// websocket 업그레이더
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 모든 접속 허용
	},
}
var roomManager = pkgRoom.NewRoomManager()
var clientManager = pkgClient.NewClientManager()

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// http > socket으로
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	nickname := r.URL.Query().Get("nickname")
	c := clientManager.CreateClient(conn, nickname)
	go c.Write()
	go c.Read(roomManager)

	_ = c.SendMessage(dto.Message{
		Div:  "UUID",
		Text: c.ID(), // 또는 UUID 전용 필드 추가
	})
}

func handleJoinRoom(w http.ResponseWriter, r *http.Request) {
	clientUUID := r.Header.Get("X-Client-UUID")
	roomID := r.URL.Query().Get("roomID")

	c := clientManager.Get(clientUUID)
	if c == nil {
		http.Error(w, "client not found", http.StatusBadRequest)
		return
	}

	room := roomManager.GetRoom(roomID)
	room.Join(c)
	log.Println(room.ID, room.Clients)
	pkgResponse.Success(w, "joined room")
}

func main() {
	// asdfasdf
	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/join", handleJoinRoom)
	http.Handle("/", http.FileServer(http.Dir("./html")))

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
