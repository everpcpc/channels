package web

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack/slackevents"
)

func (e *env) slackEvents(token string) func(*gin.Context) {
	verifyer := slackevents.OptionVerifyToken(
		&slackevents.TokenComparator{VerificationToken: token},
	)
	return func(c *gin.Context) {
		data, err := c.GetRawData()
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
			return
		}
		event, err := slackevents.ParseEvent(json.RawMessage(data), verifyer)
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
			return
		}
		switch event.Type {
		case slackevents.URLVerification:
			var r *slackevents.ChallengeResponse
			err := json.Unmarshal(data, &r)
			if err != nil {
				c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
				return
			}
			c.JSON(200, gin.H{"challenge": r.Challenge})
		default:
			c.Status(200)
		}
	}
}
