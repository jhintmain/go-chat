package dto

type Message struct {
	Div     string      `json:"div"` // CHAT, SYSTEM, ROOMLIST, USERLIST
	RoomId  string      `json:"roomID"`
	Content string      `json:"content"`
	Data    interface{} `json:"data"`
}
