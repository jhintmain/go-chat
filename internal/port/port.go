package port

import "go-chat/internal/dto"

type Client interface {
	ID() string
	SendMessage(dto.Message) error
	Close()
}

type RoomService interface {
	HandleMessage(dto.Message)
}
