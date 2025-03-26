package main

import (
	"github.com/gorilla/websocket"
	"go-chat/server"
	"net/http"
)

// websocket 업그레이더
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 모든 접속 허용
	},
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("room")
	if roomID == "" {
		http.Error(w, "room id is required", http.StatusBadRequest)
		return
	}

	// http > socket으로tteste
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	client := &server.Client{
		Conn:     conn,
		NickName: r.URL.Query().Get("nick"),
		Send:     make(chan []byte),
	}

	room := server.GetRoom(roomID)
	if room == nil {
		http.Error(w, "room not found", http.StatusNotFound)
		return
	}

	room.Clients = map[*server.Client]bool{
		client: true,
	}

	defer conn.Close()
}

func main() {
	http.HandleFunc("/ws", handleWebSocket)
	http.Handle("/", http.FileServer(http.Dir("./html")))

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
