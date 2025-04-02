package main

import (
	"github.com/gorilla/websocket"
	internalClient "go-chat/internal/client"
	"go-chat/internal/dto"
	internalResponse "go-chat/internal/response"
	internalRoom "go-chat/internal/room"
	"net/http"
)

// websocket 업그레이더
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 모든 접속 허용
	},
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// http > socket으로
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	// client
	nickname := r.URL.Query().Get("nickname")
	client := internalClient.NewClient(conn, nickname)

	w.Header().Set("X-Client-UUID", client.UUID) // 응답 헤더에 UUID 추가
	internalResponse.Success(w, "connect socket")
	//go client.write()
	//client.read(room)
}

func handleCreateRoom(w http.ResponseWriter, r *http.Request) {
	clientUUID := r.Header.Get("X-Client-UUID") // 클라이언트에서 UUID 보내는지 확인
	if clientUUID == "" {
		http.Error(w, "uuid required", http.StatusUnauthorized)
		return
	}
	if _, exist := internalClient.ClientList[clientUUID]; exist == false {
		http.Error(w, "client not found", http.StatusBadRequest)
		return
	}

	roomID := r.URL.Query().Get("roomID")
	room := internalRoom.GetRoom(roomID)
	client := internalClient.ClientList[clientUUID]
	room.Clients[clientUUID] = client

	room.Broadcast <- dto.Message{
		Div:     "SYSTEM",
		Content: client.NickName + " 님이 입장하였습니다",
	}

	internalResponse.Success(w, "create room")
}

func handleJoinRoom(w http.ResponseWriter, r *http.Request) {
	clientUUID := r.Header.Get("X-Client-UUID") // 클라이언트에서 UUID 보내는지 확인
	if clientUUID == "" {
		http.Error(w, "uuid required", http.StatusUnauthorized)
		return
	}
	if _, exist := internalClient.ClientList[clientUUID]; exist == false {
		http.Error(w, "client not found", http.StatusBadRequest)
		return
	}

	roomID := r.URL.Query().Get("roomID")
	if _, exist := internalRoom.RoomList[roomID]; exist == false {
		http.Error(w, "room not found", http.StatusBadRequest)
		return
	}

	room := internalRoom.GetRoom(roomID)
	client := internalClient.ClientList[clientUUID]
	room.Clients[clientUUID] = client

	room.Broadcast <- dto.Message{
		Div:     "SYSTEM",
		Content: client.NickName + " 님이 입장하였습니다",
	}

	internalResponse.Success(w, "join room")
}

func main() {
	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/createRoom", handleCreateRoom)
	http.HandleFunc("/joinRoom", handleJoinRoom)
	http.Handle("/", http.FileServer(http.Dir("./public")))

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
