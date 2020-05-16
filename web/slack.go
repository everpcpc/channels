package web

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

func (e *env) slackEvents(secret string) func(*gin.Context) {
	return func(c *gin.Context) {
		var err error
		defer func() {
			if err != nil {
				reportError(c, err)
				c.AbortWithStatus(500)
			}
		}()

		var secretVerifier slack.SecretsVerifier
		secretVerifier, err = slack.NewSecretsVerifier(c.Request.Header, secret)
		if err != nil {
			return
		}

		var data []byte
		data, err = c.GetRawData()
		if err != nil {
			return
		}

		_, err = secretVerifier.Write(data)
		if err != nil {
			return
		}
		err = secretVerifier.Ensure()
		if err != nil {
			err = nil
			c.AbortWithStatusJSON(403, gin.H{"error": "verify failed"})
			return
		}

		var event slackevents.EventsAPIEvent
		event, err = slackevents.ParseEvent(json.RawMessage(data), slackevents.OptionNoVerifyToken())
		if err != nil {
			return
		}

		switch event.Type {
		case slackevents.URLVerification:
			var r *slackevents.ChallengeResponse
			err = json.Unmarshal(data, &r)
			if err != nil {
				return
			}
			c.JSON(200, gin.H{"challenge": r.Challenge})
		default:
			c.Status(200)
		}
	}
}
