package web

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	"channels/storage"
)

// sentryMessage from:
// https://github.com/getsentry/sentry/blob/master/src/sentry/plugins/sentry_webhooks/plugin.py#L97
type sentryMessage struct {
	ID          string
	Project     string
	ProjectName string `json:"project_name"`
	ProjectSlug string `json:"project_slug"`

	// HACK: add first team in sentry webhook data
	Team  string
	Title string

	Logger  string
	Level   string
	Culprit string
	Message string
	URL     string

	TriggeringRules []string `json:"triggering_rules"`
}

// webhookSentry handles request from sentry as a webhook
func (e *env) webhookSentry(c *gin.Context) {
	caller, ok := e.checkToken(c)
	if !ok {
		return
	}
	if len(caller.Caps) != 1 {
		c.AbortWithStatusJSON(500, gin.H{"error": "caps invalid"})
		return
	}

	var msg sentryMessage
	if err := c.BindJSON(&msg); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}

	var target string
	if caller.Caps[0] == "#" {
		if msg.Team == "" {
			c.AbortWithStatusJSON(400, gin.H{"error": "no target"})
			return
		}
		target = "#" + msg.Team
	} else {
		target = caller.Caps[0]
	}
	text := fmt.Sprintf("[%s] %s-%s ( %s )",
		msg.Project, msg.Title, msg.Message, msg.URL)

	markdown := fmt.Sprintf("[%s] <%s|%s>\n> %s",
		msg.Project, msg.URL, msg.Title, msg.Message)

	m := storage.Message{
		From:      caller.Name,
		To:        target,
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
