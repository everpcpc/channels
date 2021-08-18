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

type Config struct {
	Name         string
	Token        string
	SignedSecret string
	Proxy        string

	BotGravatarMail   string
	HumanGravatarMail string

	JoinChannels []string

	Forwards map[string]struct {
		Token           string
		ForwardChannels []struct {
			Source string
			Target string
		}
	}
}

type ForwardClient struct {
	api *slack.Client

	forwardChannels map[string]string
}

type Client struct {
	name  string
	api   *slack.Client
	store storage.Backend
	cache storage.CacheBackend

	joinChannels []string

	signedSecret string

	botGravatarMail   string
	humanGravatarMail string

	forwards map[string]*ForwardClient
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
		forwards:          make(map[string]*ForwardClient),
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
	for forwardName, forwardConfig := range cfg.Forwards {
		fclient := slack.New(forwardConfig.Token, slack.OptionHTTPClient(
			&http.Client{
				Transport: &http.Transport{
					Proxy: http.ProxyURL(proxyUrl),
				},
			},
		))
		c.forwards[forwardName] = &ForwardClient{
			api:             fclient,
			forwardChannels: make(map[string]string),
		}
		for _, forwardChannel := range forwardConfig.ForwardChannels {
			c.forwards[forwardName].forwardChannels[forwardChannel.Source] = forwardChannel.Target
		}
	}
	return
}

func (c *Client) Run() {
	st := state.New(c.name, c.store, &auth.Anonymous{})
	go st.Pulling()

	var cursor string
	for {
		chs, cursor, err := c.api.GetConversations(&slack.GetConversationsParameters{
			Limit:  1000,
			Cursor: cursor,
			Types:  []string{"public_channel"},
		})
		if err != nil {
			logrus.Errorf("get conversations failed: %s", err)
		}
		for _, ch := range chs {
			if !contains(c.joinChannels, "#"+ch.Name) {
				continue
			}
			_, warnings, _, err := c.api.JoinConversation(ch.ID)
			if err != nil {
				logrus.Warnf("join channel %s failed: %v, warning: %v", ch, err, warnings)
				continue
			}
		}
		if cursor == "" {
			break
		}
	}

	for name, fc := range c.forwards {
		for {
			chs, cursor, err := fc.api.GetConversations(&slack.GetConversationsParameters{
				Limit:  1000,
				Cursor: cursor,
				Types:  []string{"public_channel"},
			})
			if err != nil {
				logrus.Errorf("get conversations for %s failed: %s", name, err)
			}
			for _, ch := range chs {
				if !containsValue(fc.forwardChannels, "#"+ch.Name) {
					continue
				}
				_, warnings, _, err := fc.api.JoinConversation(ch.ID)
				if err != nil {
					logrus.Warnf("join channel %s in %s failed: %v, warning: %v", ch.Name, name, err, warnings)
					continue
				}
			}
			if cursor == "" {
				break
			}
		}
	}

	for _, ch := range c.joinChannels {
		if !strings.HasPrefix(ch, "#") {
			continue
		}

		channel := st.NewChannel(ch)
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
			for fname, forward := range c.forwards {
				target := forward.forwardChannels[channel.GetName()]
				if target == "" {
					continue
				}
				if _, _, _, err := forward.api.SendMessage(target, content,
					slack.MsgOptionAsUser(false),
					slack.MsgOptionIconURL(iconURL),
					slack.MsgOptionUsername(msg.From),
					slack.MsgOptionTS(msg.SlackThreadTimeStamp),
				); err != nil {
					logrus.Errorf("send msg to %s in %s failed: %v", channel.GetName(), fname, err)
				}
			}
		})
	}

	select {}
}
