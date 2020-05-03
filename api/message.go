package api

import (
	"time"

	"github.com/gin-gonic/gin"

	"mcdc/storage"
)

type messageParam struct {
	Channel string `json:"channel,,omitempty"`
	Text    string `json:"text,omitempty"`
}

func postMessage(c *gin.Context) {
	var msg messageParam
	var err error
	err = c.BindJSON(&msg)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}
	err = store.Save(storage.Message{
		Channel:   msg.Channel,
		Text:      msg.Text,
		Timestamp: time.Now().UnixNano(),
	})
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "sent succeed"})
}
