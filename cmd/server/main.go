package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	pkgClient "go-chat/internal/client"
	"go-chat/internal/dto"
	pkgResponse "go-chat/internal/response"
	pkgRoom "go-chat/internal/room"
	"go-chat/internal/storage"
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
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	nickname := r.URL.Query().Get("nickname")

	client := clientManager.CreateClient(conn, nickname)
	go client.Write()
	go client.Read(roomManager)

	_ = client.SendMessage(dto.Message{
		Div:  "UUID",
		Text: client.ID(), // 또는 UUID 전용 필드 추가
	})
}

func handleJoinRoom(w http.ResponseWriter, r *http.Request) {
	clientUUID := r.Header.Get("X-Client-UUID")
	roomID := r.URL.Query().Get("roomID")

	log.Println("join - clientUUID", clientUUID)
	log.Println("join - roomID", roomID)
	client := clientManager.Get(clientUUID)
	if client == nil {
		log.Println("!!!!!!!!!!!")
		http.Error(w, "client not found", http.StatusBadRequest)
		return
	}

	isNewRoom, room := roomManager.GetRoom(roomID)
	room.Join(client)
	log.Println("join - room clients", roomID, room.Clients)

	room.Broadcast <- dto.Message{
		Div:      "CHAT",
		RoomID:   roomID,
		Nickname: client.Nickname(),
		Text:     fmt.Sprintf("[%s] 님 입장", client.Nickname()),
	}

	if isNewRoom {
		// 모든 클라이언트에게 발송
		pkgClient.AnnounceAllClient(dto.Message{
			Div:    "CREATE_ROOM",
			RoomID: roomID,
		})
		pkgResponse.Success(w, "created room")
		return
	}

	pkgResponse.Success(w, "joined room")
}

func handleRoomList(w http.ResponseWriter, r *http.Request) {
	roomList := roomManager.GetRoomList()
	pkgResponse.Success(w, roomList)
}

func getRoomClients(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("roomID")
	_, roomList := roomManager.GetRoom(roomID)
	log.Println("roomList", roomList.Clients)
	pkgResponse.Success(w, roomList)
}

func main() {
	storage.InitRedis()

	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/join", handleJoinRoom)
	http.HandleFunc("/rooms", handleRoomList)
	http.HandleFunc("/info/roomClients", getRoomClients)

	// /js/ -> ./static/js/
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("./static/js"))))

	// /html/ -> ./static/html/
	http.Handle("/html/", http.StripPrefix("/html/", http.FileServer(http.Dir("./static/html"))))

	defer storage.CloseRedis()
	if err := http.ListenAndServe(":1324", nil); err != nil {
		panic(err)
	}
}
