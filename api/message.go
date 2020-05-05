package api

import (
	"github.com/gin-gonic/gin"

	"channels/storage"
)

func (e *env) postMessage(c *gin.Context) {
	token, ok := c.Params.Get("token")
	if !ok {
		c.AbortWithStatusJSON(400, gin.H{"error": "token error"})
		return
	}
	caller, err := e.webhookAuth.Authenticate("token", token)
	if err != nil {
		c.AbortWithStatusJSON(403, gin.H{"error": "auth failed"})
		return
	}

	var msg storage.Message
	err = c.BindJSON(&msg)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}
	if !caller.IsCapable(&msg) {
		c.AbortWithStatusJSON(403, gin.H{"error": "scope failed"})
		return
	}

	err = e.store.Save(msg)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "success"})
}
