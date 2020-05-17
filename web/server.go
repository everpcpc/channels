package web

import (
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"

	"channels/auth"
	"channels/slack"
	"channels/state"
	"channels/storage"
)

type Config struct {
	Listen      string
	WebhookAuth string
	APIAuth     string
}

type Server struct {
	state      state.State
	store      storage.Backend
	tokenStore storage.TokenBackend
	engine     *gin.Engine
}

func NewServer(cfg *Config, store storage.Backend, tokenStore storage.TokenBackend) *Server {
	r := gin.New()
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/ping"},
	}))
	r.Use(gin.Recovery())

	r.Use(sentrygin.New(sentrygin.Options{
		Repanic: true,
	}))

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	server := &Server{
		store:      store,
		tokenStore: tokenStore,
		engine:     r,
	}
	return server
}

func (s *Server) WithWebhook(authMethod string) {
	var webhookAuth gin.HandlerFunc
	if authMethod == "token" {
		webhookAuth = authToken(&auth.TokenAuth{Store: s.tokenStore})
	} else {
		webhookAuth = authAnoymous()
	}
	webhook := s.engine.Group("/webhook", webhookAuth)
	{
		webhook.POST("/message/:token", s.postMessage)
		webhook.POST("/sentry/:token", s.webhookSentry)
		webhook.POST("/github/:token", s.webhookGitHub)
		webhook.POST("/alertmanager/:token", s.webhookAlertManager)
	}
}
func (s *Server) WithSlack(api *slack.Client) {
	rSlack := s.engine.Group("/slack")
	rSlack.POST("/events", s.slackEvents(api))
}

func (s *Server) Run(listen string) {
	s.engine.Run(listen)
}
