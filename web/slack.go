package web

import (
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack/slackevents"

	"channels/slack"
	"channels/storage"
)

func (s *Server) slackEvents(api *slack.Client) func(*gin.Context) {
	return func(c *gin.Context) {
		var err error
		defer func() {
			if err != nil {
				reportError(c, err)
				c.AbortWithStatus(500)
			}
		}()

		var ok bool
		var data []byte
		data, ok, err = api.VerifyWithSignedSecret(c.Request.Header, c.GetRawData)
		if err != nil {
			return
		}
		if !ok {
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
		case slackevents.CallbackEvent:
			ev := event.InnerEvent
			switch ev.Type {
			case slackevents.Message:
				msg := ev.Data.(*slackevents.MessageEvent)
				// NOTE: we only deal with public channel human message
				if msg.ChannelType != "channel" || msg.SubType != "" || msg.BotID != "" {
					return
				}
				var username, channel string

				username, err = api.GetUserName(msg.User)
				if err != nil {
					return
				}
				channel, err = api.GetChannelName(msg.Channel)
				if err != nil {
					return
				}

				m := storage.Message{
					Source:    storage.MessageSourceSlack,
					From:      username,
					To:        channel,
					Text:      msg.Text,
					Timestamp: time.Now().UnixNano(),
				}

				err = s.store.Save(&m)
				if err != nil {
					return
				}

				c.JSON(200, gin.H{"status": "success"})
			}
		}
	}
}
