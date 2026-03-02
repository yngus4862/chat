package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/yngus4862/chat/internal/store"
)

type RedisPubSub struct {
	client *redis.Client
	mu     sync.Mutex
	subs   map[int64]*redis.PubSub
}

func NewRedisPubSub(addr string) *RedisPubSub {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &RedisPubSub{client: rdb, subs: make(map[int64]*redis.PubSub)}
}

func (r *RedisPubSub) Close() error {
	if r == nil || r.client == nil {
		return nil
	}
	r.mu.Lock()
	for _, ps := range r.subs {
		_ = ps.Close()
	}
	r.subs = map[int64]*redis.PubSub{}
	r.mu.Unlock()
	return r.client.Close()
}

func (r *RedisPubSub) Ping(ctx context.Context) error {
	if r == nil || r.client == nil {
		return nil
	}
	return r.client.Ping(ctx).Err()
}

func roomChannel(roomID int64) string {
	return fmt.Sprintf("room:%d", roomID)
}

func (r *RedisPubSub) PublishMessage(ctx context.Context, msg store.Message) error {
	if r == nil || r.client == nil {
		return nil
	}
	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return r.client.Publish(ctx, roomChannel(msg.RoomID), b).Err()
}

func (r *RedisPubSub) SubscribeRoom(roomID int64) (<-chan store.Message, func(), error) {
	if r == nil || r.client == nil {
		ch := make(chan store.Message)
		close(ch)
		return ch, func() {}, nil
	}

	r.mu.Lock()
	ps, ok := r.subs[roomID]
	if !ok {
		ps = r.client.Subscribe(context.Background(), roomChannel(roomID))
		r.subs[roomID] = ps
	}
	r.mu.Unlock()

	out := make(chan store.Message, 128)
	done := make(chan struct{})

	go func() {
		defer close(out)
		// Wait for subscription
		_, _ = ps.ReceiveTimeout(context.Background(), 2*time.Second)
		ch := ps.Channel()
		for {
			select {
			case <-done:
				return
			case m, ok := <-ch:
				if !ok {
					return
				}
				var msg store.Message
				if err := json.Unmarshal([]byte(m.Payload), &msg); err != nil {
					continue
				}
				out <- msg
			}
		}
	}()

	cancel := func() {
		close(done)
	}

	return out, cancel, nil
}
