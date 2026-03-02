package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yngus4862/chat/internal/health"
)

type Deps struct {
	Handlers *Handlers
	ReadyFn  func() health.Result
}

func NewRouter(d Deps) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.GET("/readyz", func(c *gin.Context) {
		res := d.ReadyFn()
		if res.Status == "ready" {
			c.JSON(http.StatusOK, res)
			return
		}
		c.JSON(http.StatusServiceUnavailable, res)
	})

	v1 := r.Group("/v1")
	{
		v1.POST("/rooms", d.Handlers.CreateRoom)
		v1.GET("/rooms", d.Handlers.ListRooms)
		v1.POST("/rooms/:roomId/messages", d.Handlers.PostMessage)
		v1.GET("/rooms/:roomId/messages", d.Handlers.ListMessages)
	}

	return r
}
