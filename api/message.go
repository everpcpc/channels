package api

import (
	"time"

	"github.com/gin-gonic/gin"

	"channels/storage"
)

type message struct {
	Target string
	Text   string
}

func (e *env) postMessage(c *gin.Context) {
	caller, ok := e.checkToken(c)
	if !ok {
		return
	}

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
		Timestamp: time.Now().UnixNano(),
	}
	if err := e.store.Save(&m); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "success"})
}
