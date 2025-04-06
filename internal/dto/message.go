package dto

type Message struct {
	Div    string      `json:"div"` // CHAT, SYSTEM, ROOMLIST, USERLIST
	RoomID string      `json:"roomID"`
	Text   string      `json:"text"`
	Data   interface{} `json:"data"`
}
