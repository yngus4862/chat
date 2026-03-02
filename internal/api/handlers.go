package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yngus4862/chat/internal/store"
	"github.com/yngus4862/chat/internal/ws"
)

type Handlers struct {
	Store *store.Store
	Hub   *ws.Hub
}

type createRoomReq struct {
	Name string `json:"name"`
}

type postMessageReq struct {
	Content     string `json:"content"`
	ClientMsgID string `json:"clientMsgId"`
}

type listMessagesResp struct {
	Items      []store.Message `json:"items"`
	NextCursor string          `json:"nextCursor,omitempty"`
	HasMore    bool            `json:"hasMore"`
}

func (h *Handlers) CreateRoom(c *gin.Context) {
	var req createRoomReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" || len([]rune(name)) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name required (<=200)"})
		return
	}

	r, err := h.Store.CreateRoom(c.Request.Context(), name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, r)
}

func (h *Handlers) ListRooms(c *gin.Context) {
	limit := parseInt(c.Query("limit"), 50)
	rooms, err := h.Store.ListRooms(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rooms)
}

func (h *Handlers) PostMessage(c *gin.Context) {
	roomID, err := strconv.ParseInt(c.Param("roomId"), 10, 64)
	if err != nil || roomID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid roomId"})
		return
	}
	var req postMessageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	content := strings.TrimSpace(req.Content)
	if content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content required"})
		return
	}
	if len([]rune(content)) > 5000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content too long (<=5000)"})
		return
	}
	clientMsgID := strings.TrimSpace(req.ClientMsgID)

	msg, err := h.Store.CreateMessage(c.Request.Context(), roomID, content, "rest", clientMsgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if h.Hub != nil {
		h.Hub.BroadcastMessage(c.Request.Context(), msg)
	}

	c.JSON(http.StatusCreated, msg)
}

func (h *Handlers) ListMessages(c *gin.Context) {
	roomID, err := strconv.ParseInt(c.Param("roomId"), 10, 64)
	if err != nil || roomID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid roomId"})
		return
	}
	limit := parseInt(c.Query("limit"), 50)
	cursor := parseInt64(c.Query("cursor"), 0)

	items, nextCursor, err := h.Store.ListMessages(c.Request.Context(), roomID, cursor, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	resp := listMessagesResp{Items: items}
	if nextCursor > 0 {
		resp.NextCursor = strconv.FormatInt(nextCursor, 10)
	}
	resp.HasMore = len(items) == limit

	c.JSON(http.StatusOK, resp)
}

func parseInt(v string, def int) int {
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	if i <= 0 {
		return def
	}
	if i > 200 {
		return 200
	}
	return i
}

func parseInt64(v string, def int64) int64 {
	if v == "" {
		return def
	}
	i, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return def
	}
	return i
}
