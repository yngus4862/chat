package ws

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/yngus4862/chat/internal/store"
)

type Hub struct {
	st *store.Store
	ps *RedisPubSub

	mu    sync.RWMutex
	rooms map[int64]map[*Client]struct{}
	subs  map[int64]func()
}

type Client struct {
	conn   *websocket.Conn
	roomID int64
	send   chan []byte
	hub    *Hub
}

type inbound struct {
	Content     string `json:"content"`
	ClientMsgID string `json:"clientMsgId,omitempty"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func NewHub(st *store.Store, ps *RedisPubSub) *Hub {
	return &Hub{
		st:    st,
		ps:    ps,
		rooms: make(map[int64]map[*Client]struct{}),
		subs:  make(map[int64]func()),
	}
}

func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	roomID, err := strconv.ParseInt(r.URL.Query().Get("roomId"), 10, 64)
	if err != nil || roomID <= 0 {
		http.Error(w, "invalid roomId", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	c := &Client{
		conn:   conn,
		roomID: roomID,
		send:   make(chan []byte, 256),
		hub:    h,
	}

	h.join(c)
	go c.writePump()
	c.readPump()
}

func (h *Hub) BroadcastMessage(ctx context.Context, msg store.Message) {
	// publish to redis (so other instances can deliver), and also deliver locally
	if h.ps != nil {
		_ = h.ps.PublishMessage(ctx, msg)
	}
	h.deliver(msg)
}

func (h *Hub) join(c *Client) {
	h.mu.Lock()
	set, ok := h.rooms[c.roomID]
	if !ok {
		set = make(map[*Client]struct{})
		h.rooms[c.roomID] = set

		// start redis subscription for room (once)
		if h.ps != nil {
			ch, cancel, err := h.ps.SubscribeRoom(c.roomID)
			if err == nil {
				h.subs[c.roomID] = cancel
				go func(roomID int64, in <-chan store.Message) {
					for msg := range in {
						h.deliver(msg)
					}
				}(c.roomID, ch)
			}
		}
	}
	set[c] = struct{}{}
	h.mu.Unlock()
}

func (h *Hub) leave(c *Client) {
	h.mu.Lock()
	set, ok := h.rooms[c.roomID]
	if ok {
		delete(set, c)
		if len(set) == 0 {
			delete(h.rooms, c.roomID)
			if cancel, ok := h.subs[c.roomID]; ok {
				cancel()
				delete(h.subs, c.roomID)
			}
		}
	}
	h.mu.Unlock()
}

func (h *Hub) deliver(msg store.Message) {
	b, err := json.Marshal(msg)
	if err != nil {
		return
	}
	h.mu.RLock()
	set := h.rooms[msg.RoomID]
	for c := range set {
		select {
		case c.send <- b:
		default:
			// drop slow client
		}
	}
	h.mu.RUnlock()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.leave(c)
		_ = c.conn.Close()
	}()

	_ = c.conn.SetReadLimit(1 << 20)
	_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
		var in inbound
		if err := json.Unmarshal(data, &in); err != nil {
			continue
		}
		if len(in.Content) == 0 {
			continue
		}

		// Persist message then broadcast
		msg, err := c.hub.st.CreateMessage(context.Background(), c.roomID, in.Content, "ws", in.ClientMsgID)
		if err != nil {
			continue
		}
		c.hub.BroadcastMessage(context.Background(), msg)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
