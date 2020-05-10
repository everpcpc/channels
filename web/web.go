package web

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
	r := gin.New()
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/ping"},
	}))
	r.Use(gin.Recovery())

	e := &env{
		store:       store,
		webhookAuth: webhookAuth,
	}

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	webhook := r.Group("/webhook")
	{
		webhook.POST("/message/:token", e.postMessage)
		webhook.POST("/sentry/:token", e.webhookSentry)
		webhook.POST("/github/:token", e.webhookGitHub)
		webhook.POST("/alertmanager/:token", e.webhookAlertManager)
	}

	r.Run(fmt.Sprintf(":%d", port))
}

func (e *env) checkToken(c *gin.Context) (*auth.Caller, bool) {
	token, ok := c.Params.Get("token")
	if !ok {
		c.AbortWithStatusJSON(400, gin.H{"error": "token error"})
		return nil, false
	}
	caller, err := e.webhookAuth.Authenticate("token", token)
	if err != nil {
		c.AbortWithStatusJSON(403, gin.H{"error": "auth failed"})
		return nil, false
	}
	return caller, true
}
