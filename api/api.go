package api

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"channels/auth"
	"channels/state"
	"channels/storage"
)

type env struct {
	state       state.State
	store       storage.Backend
	authPlugin  auth.Plugin
	webhookAuth auth.Plugin
}

func RunServer(port int, authPlugin auth.Plugin, webhookAuth auth.Plugin, store storage.Backend) {
	r := gin.Default()

	e := &env{
		store:       store,
		webhookAuth: webhookAuth,
	}

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	api := r.Group("/api")
	{
		api.POST("/message/:token", e.postMessage)
	}

	r.Run(fmt.Sprintf(":%d", port))
}
