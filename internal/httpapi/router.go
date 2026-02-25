package httpapi

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yngus4862/chat/internal/model"
	"github.com/yngus4862/chat/internal/realtime"
	"github.com/yngus4862/chat/internal/store"
)

func NewRouter(st *store.Store, hub *realtime.Hub) http.Handler {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.GET("/readyz", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := st.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	v1 := r.Group("/v1")
	{
		v1.POST("/rooms", func(c *gin.Context) {
			var req model.CreateRoomRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
			defer cancel()
			room, err := st.CreateRoom(ctx, req.Name)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusCreated, room)
		})

		v1.GET("/rooms", func(c *gin.Context) {
			ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
			defer cancel()
			rooms, err := st.ListRooms(ctx)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, rooms)
		})

		v1.POST("/rooms/:roomId/messages", func(c *gin.Context) {
			roomID, err := strconv.ParseInt(c.Param("roomId"), 10, 64)
			if err != nil || roomID <= 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid roomId"})
				return
			}
			var req model.CreateMessageRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
			defer cancel()
			msg, err := st.CreateMessage(ctx, roomID, req.Content)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			// REST로 쏜 메시지도 WS 클라이언트에게 브로드캐스트(최소 스켈레톤)
			hub.Broadcast(roomID, msg)

			c.JSON(http.StatusCreated, msg)
		})

		v1.GET("/rooms/:roomId/messages", func(c *gin.Context) {
			roomID, err := strconv.ParseInt(c.Param("roomId"), 10, 64)
			if err != nil || roomID <= 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid roomId"})
				return
			}
			limit, _ := strconv.Atoi(c.Query("limit"))

			ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
			defer cancel()
			msgs, err := st.ListMessages(ctx, roomID, limit)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, msgs)
		})
	}

	return r
}
