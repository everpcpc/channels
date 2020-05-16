package web

import (
	"github.com/gin-gonic/gin"

	"channels/auth"
	"channels/state"
	"channels/storage"
)

type Config struct {
	Listen      string
	WebhookAuth string
	APIAuth     string
	SlackToken  string
}

type env struct {
	state state.State
	store storage.Backend
}

func RunServer(cfg *Config, store storage.Backend, tokenStore storage.TokenBackend) {
	r := gin.New()
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/ping"},
	}))
	r.Use(gin.Recovery())

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	e := &env{
		store: store,
	}

	var webhookAuth gin.HandlerFunc
	if cfg.WebhookAuth == "token" {
		webhookAuth = authToken(&auth.TokenAuth{Store: tokenStore})
	} else {
		webhookAuth = authAnoymous()
	}
	webhook := r.Group("/webhook", webhookAuth)
	{
		webhook.POST("/message/:token", e.postMessage)
		webhook.POST("/sentry/:token", e.webhookSentry)
		webhook.POST("/github/:token", e.webhookGitHub)
		webhook.POST("/alertmanager/:token", e.webhookAlertManager)
	}

	if cfg.SlackToken != "" {
		slack := r.Group("/slack")
		slack.POST("/events", e.slackEvents(cfg.SlackToken))
	}

	r.Run(cfg.Listen)
}
