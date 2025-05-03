package client

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go-chat/internal/dto"
	"go-chat/internal/port"
	"log"
	"sync"
)

type ChatClient struct {
	uuid     string
	nickname string
	conn     *websocket.Conn
	send     chan dto.Message
	rooms    map[string]bool
	mutex    sync.Mutex
}

func (c *ChatClient) ID() string {
	return c.uuid
}

func (c *ChatClient) Nickname() string {
	return c.nickname
}

func (c *ChatClient) SendMessage(msg dto.Message) error {
	c.send <- msg
	return nil
}

func (c *ChatClient) Close() {
	c.conn.Close()
}

func (c *ChatClient) Write() {
	for msg := range c.send {
		c.conn.WriteJSON(msg)
	}
}

func (c *ChatClient) Read(rm port.RoomService) {
	for {
		var msg dto.Message
		if err := c.conn.ReadJSON(&msg); err != nil {
			break
		}
		msg.Text = c.nickname + ": " + msg.Text
		rm.HandleMessage(msg)
	}
}

// client manager
var clients = make(map[string]*ChatClient)
var mtx sync.Mutex

func NewClientManager() *ClientManager {
	return &ClientManager{}
}

type ClientManager struct{}

func (cm *ClientManager) CreateClient(conn *websocket.Conn, nickname string) *ChatClient {
	c := &ChatClient{
		uuid:     uuid.NewString(),
		nickname: nickname,
		conn:     conn,
		send:     make(chan dto.Message),
		rooms:    make(map[string]bool),
	}
	mtx.Lock()
	clients[c.uuid] = c
	log.Println("create clients", clients)
	mtx.Unlock()
	return c
}

func (cm *ClientManager) Get(uuid string) *ChatClient {
	mtx.Lock()
	defer mtx.Unlock()
	return clients[uuid]
}

func AnnounceAllClient(message dto.Message) {
	for _, client := range clients {
		_ = client.SendMessage(message)
	}
}
