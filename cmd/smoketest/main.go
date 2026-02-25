package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type healthResp struct {
	Status string `json:"status"`
}

type createRoomResp struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type msgResp struct {
	ID      int64  `json:"id"`
	RoomID  int64  `json:"roomId"`
	Content string `json:"content"`
}

func main() {
	baseURL := env("BASE_URL", "http://localhost:8080")
	wsBase := env("WS_URL", "ws://localhost:8081/ws")
	timeoutSec := envInt("SMOKE_TIMEOUT_SEC", 20)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
	defer cancel()

	c := &http.Client{Timeout: 6 * time.Second}

	fmt.Println("[smoke] 1) GET /healthz")
	mustHealth(ctx, c, baseURL+"/healthz", "ok")

	fmt.Println("[smoke] 2) GET /readyz")
	mustHealthAny(ctx, c, baseURL+"/readyz", []string{"ready", "ok"})

	roomName := fmt.Sprintf("smoke-room-%d", time.Now().Unix())
	fmt.Println("[smoke] 3) POST /v1/rooms:", roomName)
	roomID := mustCreateRoom(ctx, c, baseURL+"/v1/rooms", roomName)

	fmt.Println("[smoke] 4) GET /v1/rooms (roomId 포함 여부 확인)")
	mustListRoomsContains(ctx, c, baseURL+"/v1/rooms", roomID)

	fmt.Println("[smoke] 5) WebSocket connect + send + receive")
	wsURL := wsBase + "?roomId=" + strconv.FormatInt(roomID, 10)
	mustWebSocketEcho(wsURL, "ws-hello")

	fmt.Println("[smoke] 6) POST /v1/rooms/{id}/messages")
	mustPostMessage(ctx, c, fmt.Sprintf("%s/v1/rooms/%d/messages", baseURL, roomID), "rest-hello")

	fmt.Println("[smoke] 7) GET /v1/rooms/{id}/messages (rest-hello 포함 여부)")
	mustListMessagesContains(ctx, c, fmt.Sprintf("%s/v1/rooms/%d/messages?limit=50", baseURL, roomID), "rest-hello")

	fmt.Println("[smoke] ✅ ALL PASSED")
}

func mustHealth(ctx context.Context, c *http.Client, url string, expected string) {
	var out healthResp
	getJSON(ctx, c, url, &out)
	if out.Status != expected {
		fatalf("health status mismatch: got=%q expected=%q (url=%s)", out.Status, expected, url)
	}
}

func mustHealthAny(ctx context.Context, c *http.Client, url string, expected []string) {
	var out healthResp
	getJSON(ctx, c, url, &out)
	for _, e := range expected {
		if out.Status == e {
			return
		}
	}
	fatalf("ready status mismatch: got=%q expected one of=%v (url=%s)", out.Status, expected, url)
}

func mustCreateRoom(ctx context.Context, c *http.Client, url, name string) int64 {
	body := map[string]any{"name": name}
	b, _ := json.Marshal(body)

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	res, err := c.Do(req)
	if err != nil {
		fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		fatalf("create room failed: status=%d url=%s", res.StatusCode, url)
	}

	var out createRoomResp
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		fatal(err)
	}
	if out.ID <= 0 {
		fatalf("invalid room id: %+v", out)
	}
	return out.ID
}

func mustListRoomsContains(ctx context.Context, c *http.Client, url string, roomID int64) {
	var rooms []createRoomResp
	getJSON(ctx, c, url, &rooms)
	for _, r := range rooms {
		if r.ID == roomID {
			return
		}
	}
	fatalf("rooms list does not contain roomId=%d", roomID)
}

func mustPostMessage(ctx context.Context, c *http.Client, url, content string) {
	body := map[string]any{"content": content}
	b, _ := json.Marshal(body)

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	res, err := c.Do(req)
	if err != nil {
		fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		fatalf("post message failed: status=%d url=%s", res.StatusCode, url)
	}
}

func mustListMessagesContains(ctx context.Context, c *http.Client, url, expectedContent string) {
	var msgs []msgResp
	getJSON(ctx, c, url, &msgs)
	for _, m := range msgs {
		if strings.Contains(m.Content, expectedContent) {
			return
		}
	}
	fatalf("messages do not contain content=%q", expectedContent)
}

func mustWebSocketEcho(wsURL string, content string) {
	d := websocket.Dialer{HandshakeTimeout: 4 * time.Second}
	conn, _, err := d.Dial(wsURL, nil)
	if err != nil {
		fatalf("ws dial failed: %v (url=%s)", err, wsURL)
	}
	defer conn.Close()

	_ = conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
	if err := conn.WriteJSON(map[string]any{"content": content}); err != nil {
		fatalf("ws write failed: %v", err)
	}

	_ = conn.SetReadDeadline(time.Now().Add(4 * time.Second))
	var got map[string]any
	if err := conn.ReadJSON(&got); err != nil {
		fatalf("ws read failed: %v", err)
	}

	if v, ok := got["content"]; ok {
		if s, _ := v.(string); strings.Contains(s, content) {
			return
		}
	}
	b, _ := json.Marshal(got)
	fatal(errors.New("ws message does not contain expected content: " + string(b)))
}

func getJSON(ctx context.Context, c *http.Client, url string, out any) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	res, err := c.Do(req)
	if err != nil {
		fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		fatalf("GET failed: status=%d url=%s", res.StatusCode, url)
	}

	if err := json.NewDecoder(res.Body).Decode(out); err != nil {
		fatal(err)
	}
}

func env(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}

func envInt(k string, def int) int {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "[smoke] ERROR:", err)
	os.Exit(1)
}

func fatalf(format string, args ...any) {
	fmt.Fprintln(os.Stderr, "[smoke] ERROR:", fmt.Sprintf(format, args...))
	os.Exit(1)
}
