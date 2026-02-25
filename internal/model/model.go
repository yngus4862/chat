package model

import "time"

type Room struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

type Message struct {
	ID        int64     `json:"id"`
	RoomID    int64     `json:"roomId"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

type CreateRoomRequest struct {
	Name string `json:"name" binding:"required"`
}

type CreateMessageRequest struct {
	Content string `json:"content" binding:"required"`
}

// WebSocket inbound payload (minimal)
type WSIncoming struct {
	Content string `json:"content"`
}