package web

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	"channels/auth"
	"channels/storage"
)

// sentryMessage from:
// https://github.com/getsentry/sentry/blob/master/src/sentry/plugins/sentry_webhooks/plugin.py#L97
type sentryMessage struct {
	ID          string
	Project     string
	ProjectName string `json:"project_name"`
	ProjectSlug string `json:"project_slug"`

	Logger  string
	Level   string
	Culprit string
	Message string
	URL     string

	TriggeringRules []string `json:"triggering_rules"`
}

// webhookSentry handles request from sentry as a webhook
func (e *env) webhookSentry(c *gin.Context) {
	ctxCaller, exists := c.Get("caller")
	if !exists {
		c.AbortWithStatusJSON(403, gin.H{"error": "caller not found"})
		return
	}
	caller := ctxCaller.(*auth.Caller)

	if len(caller.Caps) != 1 {
		c.AbortWithStatusJSON(500, gin.H{"error": "caps invalid"})
		return
	}

	var msg sentryMessage
	if err := c.BindJSON(&msg); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}

	text := fmt.Sprintf("[%s] %s ( %s )",
		msg.Project, msg.Message, msg.URL)

	markdown := fmt.Sprintf("[%s] <%s|%s>",
		msg.Project, msg.URL, msg.Message)

	m := storage.Message{
		From:      caller.Name,
		To:        caller.Caps[0],
		Text:      text,
		Markdown:  markdown,
		Timestamp: time.Now().UnixNano(),
	}
	if err := e.store.Save(&m); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "success"})

}
