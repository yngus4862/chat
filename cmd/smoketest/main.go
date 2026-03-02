package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type healthResp struct {
	Status string `json:"status"`
}

type room struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type message struct {
	ID      int64  `json:"id"`
	RoomID  int64  `json:"roomId"`
	Content string `json:"content"`
}

type listMessagesResp struct {
	Items      []message `json:"items"`
	NextCursor string    `json:"nextCursor,omitempty"`
	HasMore    bool      `json:"hasMore"`
}

func main() {
	apiBase := flag.String("api", "http://127.0.0.1:8080", "api base url")
	wsBase := flag.String("ws", "ws://127.0.0.1:8081/ws", "ws url")
	timeout := flag.Int("timeout", 20, "timeout seconds")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*timeout)*time.Second)
	defer cancel()

	client := &http.Client{Timeout: 6 * time.Second}

	mustHealth(ctx, client, *apiBase+"/healthz", "ok")
	mustReady(ctx, client, *apiBase+"/readyz")

	roomID := mustCreateRoom(ctx, client, *apiBase+"/v1/rooms", "smoke-room")
	mustListRoomsContains(ctx, client, *apiBase+"/v1/rooms", roomID)

	// WS
	wsURL := *wsBase + "?roomId=" + strconv.FormatInt(roomID, 10)
	mustWebSocket(wsURL, "ws-hello")

	// REST message
	mustPostMessage(ctx, client, fmt.Sprintf("%s/v1/rooms/%d/messages", *apiBase, roomID), "rest-hello")
	mustListMessagesContains(ctx, client, fmt.Sprintf("%s/v1/rooms/%d/messages?limit=50", *apiBase, roomID), "rest-hello")

	fmt.Println("[smoke] ✅ ALL PASSED")
}

func mustHealth(ctx context.Context, c *http.Client, url, expected string) {
	var out healthResp
	getJSON(ctx, c, url, &out)
	if out.Status != expected {
		panic(fmt.Sprintf("health mismatch: got=%q expected=%q", out.Status, expected))
	}
}

func mustReady(ctx context.Context, c *http.Client, url string) {
	var out map[string]any
	getJSONAny(ctx, c, url, &out)
	st, _ := out["status"].(string)
	if st != "ready" && st != "ok" {
		panic(fmt.Sprintf("ready mismatch: got=%v", out))
	}
}

func mustCreateRoom(ctx context.Context, c *http.Client, url, name string) int64 {
	b, _ := json.Marshal(map[string]any{"name": name})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	res, err := c.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		panic(fmt.Sprintf("create room failed status=%d", res.StatusCode))
	}
	var r room
	_ = json.NewDecoder(res.Body).Decode(&r)
	if r.ID <= 0 {
		panic("invalid room id")
	}
	return r.ID
}

func mustListRoomsContains(ctx context.Context, c *http.Client, url string, roomID int64) {
	var rooms []room
	getJSON(ctx, c, url, &rooms)
	for _, r := range rooms {
		if r.ID == roomID {
			return
		}
	}
	panic("rooms list does not contain created room")
}

func mustPostMessage(ctx context.Context, c *http.Client, url, content string) {
	b, _ := json.Marshal(map[string]any{"content": content})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	res, err := c.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		panic(fmt.Sprintf("post message failed status=%d", res.StatusCode))
	}
}

func mustListMessagesContains(ctx context.Context, c *http.Client, url, expected string) {
	var resp listMessagesResp
	getJSON(ctx, c, url, &resp)
	for _, m := range resp.Items {
		if strings.Contains(m.Content, expected) {
			return
		}
	}
	panic("messages list does not contain expected content")
}

func mustWebSocket(wsURL, content string) {
	d := websocket.Dialer{HandshakeTimeout: 4 * time.Second}
	conn, _, err := d.Dial(wsURL, nil)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	_ = conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
	_ = conn.WriteJSON(map[string]any{"content": content})

	_ = conn.SetReadDeadline(time.Now().Add(4 * time.Second))
	var got map[string]any
	if err := conn.ReadJSON(&got); err != nil {
		panic(err)
	}
	if v, ok := got["content"]; ok {
		if s, _ := v.(string); strings.Contains(s, content) {
			return
		}
	}
	b, _ := json.Marshal(got)
	panic("ws response mismatch: " + string(b))
}

func getJSON[T any](ctx context.Context, c *http.Client, url string, out *T) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	res, err := c.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		panic(fmt.Sprintf("GET failed status=%d url=%s", res.StatusCode, url))
	}
	_ = json.NewDecoder(res.Body).Decode(out)
}

func getJSONAny(ctx context.Context, c *http.Client, url string, out *map[string]any) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	res, err := c.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	_ = json.NewDecoder(res.Body).Decode(out)
}
