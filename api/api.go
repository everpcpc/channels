package api

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"channels/state"
	"channels/storage"
)

type env struct {
	state state.State
	store storage.Backend
}

func RunServer(port int, store storage.Backend) {
	r := gin.Default()

	e := &env{store: store}

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	api := r.Group("/api")
	{
		api.POST("/message", e.postMessage)
	}

	r.Run(fmt.Sprintf(":%d", port))
}
