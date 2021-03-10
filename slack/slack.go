package slack

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"

	"channels/auth"
	"channels/state"
	"channels/storage"
)

type Conversation struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type Config struct {
	Name         string
	Token        string
	SignedSecret string
	Proxy        string

	BotGravatarMail   string
	HumanGravatarMail string

	JoinChannels []*Conversation
}

type Client struct {
	name  string
	api   *slack.Client
	store storage.Backend
	cache storage.CacheBackend

	joinChannels []*Conversation

	signedSecret string

	botGravatarMail   string
	humanGravatarMail string
}

func NewClient(cfg *Config, store storage.Backend, cache storage.CacheBackend) (c *Client, err error) {
	c = &Client{
		store: store,
		cache: cache,

		name:              cfg.Name,
		signedSecret:      cfg.SignedSecret,
		botGravatarMail:   cfg.BotGravatarMail,
		humanGravatarMail: cfg.HumanGravatarMail,
		joinChannels:      cfg.JoinChannels,
	}
	if cfg.Proxy == "" {
		c.api = slack.New(cfg.Token)
		return
	}

	proxyUrl, err := url.Parse(cfg.Proxy)
	if err != nil {
		return
	}
	c.api = slack.New(cfg.Token, slack.OptionHTTPClient(
		&http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyUrl),
			},
		},
	))
	return
}

func (c *Client) Run() {
	st := state.New(c.name, c.store, &auth.Anonymous{})
	go st.Pulling()

	for _, ch := range c.joinChannels {
		if !strings.HasPrefix(ch.Name, "#") {
			continue
		}
		_, warnings, _, err := c.api.JoinConversation(ch.ID)

		if err != nil {
			logrus.Warnf("join channel %s failed: %v, warning: %v", ch, err, warnings)
			continue
		}
		channel := st.NewChannel(ch.Name)
		channel.SetSendFn(func(msg *storage.Message) {
			// ignore message from self
			if msg.Source == storage.MessageSourceSlack {
				return
			}
			iconURL := "https://www.gravatar.com/avatar/"
			var mail string
			if msg.IsHuman {
				mail = fmt.Sprintf(c.humanGravatarMail, msg.From)
			} else {
				mail = fmt.Sprintf(c.botGravatarMail, msg.From)
			}

			h := md5.New()
			if _, err := io.WriteString(h, mail); err != nil {
				logrus.Warnf("email md5 failed: %v", err)
			} else {
				iconURL += hex.EncodeToString(h.Sum(nil))
			}
			var content slack.MsgOption
			if msg.Title == "" {
				if msg.Markdown != "" {
					content = slack.MsgOptionText(msg.Markdown, false)
				} else {
					content = slack.MsgOptionText(msg.Text, false)
				}
			} else {
				color := "#5bc0de" // info
				if msg.Color != "" {
					color = msg.Color
				}
				attachment := slack.Attachment{
					Title:     msg.Title,
					TitleLink: msg.Link,
					Text:      msg.Markdown,
					Color:     color,
				}
				content = slack.MsgOptionAttachments(attachment)
			}

			if _, _, _, err := c.api.SendMessage(channel.GetName(), content,
				slack.MsgOptionAsUser(false),
				slack.MsgOptionIconURL(iconURL),
				slack.MsgOptionUsername(msg.From),
				slack.MsgOptionTS(msg.SlackThreadTimeStamp),
			); err != nil {
				logrus.Errorf("send msg to %s failed: %v", channel.GetName(), err)
			}
		})
	}

	select {}
}
