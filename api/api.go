package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"mcdc/storage"
)

var (
	store storage.Backend
)

func RunServer(port int) {
	r := gin.Default()

	var err error
	store, err = storage.New("redis", "localhost:6379")
	if err != nil {
		logrus.Fatal(err)
	}

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// TODO: auth
	api := r.Group("/api")
	{
		api.POST("/message", postMessage)
	}

	r.Run(fmt.Sprintf(":%d", port))
}
