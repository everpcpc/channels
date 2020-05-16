package web

import (
	"time"

	"github.com/gin-gonic/gin"

	"channels/auth"
	"channels/storage"
)

type message struct {
	Target   string
	Text     string
	Markdown string
}

func (e *env) postMessage(c *gin.Context) {
	ctxCaller, exists := c.Get("caller")
	if !exists {
		c.AbortWithStatusJSON(403, gin.H{"error": "caller not found"})
		return
	}
	caller := ctxCaller.(*auth.Caller)

	var msg message
	if err := c.BindJSON(&msg); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}
	if !caller.IsCapable(msg.Target) {
		c.AbortWithStatusJSON(403, gin.H{"error": "scope failed"})
		return
	}

	m := storage.Message{
		From:      caller.Name,
		To:        msg.Target,
		Text:      msg.Text,
		Markdown:  msg.Markdown,
		Timestamp: time.Now().UnixNano(),
	}
	if err := e.store.Save(&m); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "success"})
}
