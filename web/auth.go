package web

import (
	"github.com/gin-gonic/gin"

	"channels/auth"
)

func authToken(tokenPlugin auth.Plugin) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := c.Params.Get("token")
		if !ok {
			c.AbortWithStatusJSON(400, gin.H{"error": "token error"})
			return
		}
		caller, err := tokenPlugin.Authenticate("token", token)
		if err != nil {
			c.AbortWithStatusJSON(403, gin.H{"error": "token failed"})
			return
		}
		c.Set("caller", caller)
		c.Next()
	}
}

func authAnoymous() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("caller", &auth.Caller{
			Name: "anoymous",
			Caps: []string{"#", "@"},
		})
		c.Next()
	}
}
