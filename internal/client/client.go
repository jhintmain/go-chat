package client

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go-chat/internal/dto"
	"log"
	"sync"
)

var ClientList map[string]*Client

type Client struct {
	UUID     string
	Conn     *websocket.Conn
	Send     chan dto.Message
	NickName string
	Mutex    *sync.Mutex
}

func NewClient(conn *websocket.Conn, nickname string) Client {
	client := Client{
		UUID:     uuid.NewString(),
		Conn:     conn,
		Send:     make(chan dto.Message),
		NickName: nickname,
		Mutex:    &sync.Mutex{},
	}
	ClientList[client.UUID] = &client
	return client
}

func (c *Client) write() {
	for msg := range c.Send {
		if err := c.Conn.WriteJSON(&msg); err != nil {
			break
		}
	}
}
func (c *Client) read() {
	for {
		var msg dto.Message
		if err := c.Conn.ReadJSON(&msg); err != nil {
			log.Println("c.conn.ReadMessage err:", err)
			break
		}
		// todo : broadcase
	}
}
