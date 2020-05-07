package api

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"channels/storage"
)

// sentryMessage from:
// https://github.com/getsentry/sentry/blob/master/src/sentry/plugins/sentry_webhooks/plugin.py#L97
type sentryMessage struct {
	ID              string
	Project         string
	ProjectName     string `json:"project_name"`
	ProjectSlug     string `json:"project_slug"`
	Logger          string
	Level           string
	Culprit         string
	Message         string
	URL             string
	TriggeringRules []string `json:"triggering_rules"`
}

// webhookSentry handles request from sentry as a webhook
func (e *env) webhookSentry(c *gin.Context) {
	caller, ok := e.checkToken(c)
	if !ok {
		return
	}

	var msg sentryMessage
	if err := c.BindJSON(&msg); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}
	if len(caller.Caps) != 1 {
		c.AbortWithStatusJSON(500, gin.H{"error": "caps invalid"})
		return
	}

	msgToSend := &storage.Message{
		From: caller.Name,
		To:   caller.Caps[0],
		Text: fmt.Sprintf("[%s] %s (%s)", msg.Project, msg.Message, msg.URL),
	}
	if err := e.store.Save(msgToSend); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "success"})

}
