package health

import (
	"context"
	"time"

	"github.com/yngus4862/chat/internal/store"
	"github.com/yngus4862/chat/internal/ws"
)

type Result struct {
	Status string `json:"status"`
	DB     string `json:"db,omitempty"`
	Redis  string `json:"redis,omitempty"`
	Error  string `json:"error,omitempty"`
}

func Ready(parent context.Context, st *store.Store, ps *ws.RedisPubSub) Result {
	ctx, cancel := context.WithTimeout(parent, 2*time.Second)
	defer cancel()

	if err := st.Ping(ctx); err != nil {
		return Result{Status: "not_ready", DB: "down", Error: err.Error()}
	}

	if ps != nil {
		if err := ps.Ping(ctx); err != nil {
			return Result{Status: "not_ready", DB: "up", Redis: "down", Error: err.Error()}
		}
		return Result{Status: "ready", DB: "up", Redis: "up"}
	}

	return Result{Status: "ready", DB: "up", Redis: "disabled"}
}
