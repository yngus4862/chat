package store

import "time"

type Room struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

type Message struct {
	ID          int64     `json:"id"`
	RoomID      int64     `json:"roomId"`
	Content     string    `json:"content"`
	Source      string    `json:"source"`
	ClientMsgID string    `json:"clientMsgId,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}
