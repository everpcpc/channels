package api

import (
	"github.com/gin-gonic/gin"

	"channels/storage"
)

func (e *env) postMessage(c *gin.Context) {
	var msg storage.Message
	var err error
	err = c.BindJSON(&msg)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}
	err = e.store.Save(msg)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "success"})
}
