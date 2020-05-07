package api

import (
	"github.com/gin-gonic/gin"

	"channels/storage"
)

func (e *env) postMessage(c *gin.Context) {
	caller, ok := e.checkToken(c)
	if !ok {
		return
	}

	var msg storage.Message
	if err := c.BindJSON(&msg); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}
	if !caller.IsCapable(&msg) {
		c.AbortWithStatusJSON(403, gin.H{"error": "scope failed"})
		return
	}

	if err := e.store.Save(&msg); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "success"})
}
