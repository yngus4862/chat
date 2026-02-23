package realtime

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/yngus4862/chat/internal/model"
	"github.com/yngus4862/chat/internal/store"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// DEV ONLY: allow all origins. In production, restrict this.
	CheckOrigin: func(r *http.Request) bool { return true },
}

func WSHandler(hub *Hub, st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomIDStr := r.URL.Query().Get("roomId")
		roomID, err := strconv.ParseInt(roomIDStr, 10, 64)
		if err != nil || roomID <= 0 {
			http.Error(w, "invalid roomId", http.StatusBadRequest)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, "upgrade failed", http.StatusBadRequest)
			return
		}
		defer conn.Close()

		hub.Add(roomID, conn)
		defer hub.Remove(roomID, conn)

		for {
			var in model.WSIncoming
			if err := conn.ReadJSON(&in); err != nil {
				return // disconnected
			}

			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			msg, err := st.CreateMessage(ctx, roomID, in.Content)
			cancel()
			if err != nil {
				// ignore write; keep connection alive
				continue
			}

			hub.Broadcast(roomID, msg)
		}
	}
}