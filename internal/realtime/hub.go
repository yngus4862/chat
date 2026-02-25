package realtime

import (
	"sync"

	"github.com/gorilla/websocket"
	"github.com/yngus4862/chat/internal/model"
)

type Hub struct {
	mu    sync.RWMutex
	rooms map[int64]map[*websocket.Conn]struct{}
}

func NewHub() *Hub {
	return &Hub{
		rooms: make(map[int64]map[*websocket.Conn]struct{}),
	}
}

func (h *Hub) Add(roomID int64, c *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.rooms[roomID]; !ok {
		h.rooms[roomID] = make(map[*websocket.Conn]struct{})
	}
	h.rooms[roomID][c] = struct{}{}
}

func (h *Hub) Remove(roomID int64, c *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if m, ok := h.rooms[roomID]; ok {
		delete(m, c)
		if len(m) == 0 {
			delete(h.rooms, roomID)
		}
	}
}

func (h *Hub) Broadcast(roomID int64, msg model.Message) {
	h.mu.RLock()
	conns := make([]*websocket.Conn, 0, len(h.rooms[roomID]))
	for c := range h.rooms[roomID] {
		conns = append(conns, c)
	}
	h.mu.RUnlock()

	for _, c := range conns {
		if err := c.WriteJSON(msg); err != nil {
			// best-effort cleanup
			_ = c.Close()
			h.Remove(roomID, c)
		}
	}
}