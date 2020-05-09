package slack

import (
	"crypto/md5"
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
	Proxy        string
	GravatarMail string
	JoinChannels []string
}

func Run(cfg *Config, store storage.Backend) {
	st := state.New(cfg.Name, store, &auth.Anonymous{})
	go st.Pulling()

	var client *slack.Client

	if cfg.Proxy == "" {
		client = slack.New(cfg.Token)
	} else {
		proxyUrl, err := url.Parse(cfg.Proxy)
		if err != nil {
			logrus.Fatalf("proxy url error: %v", err)
		}
		client = slack.New(cfg.Token, slack.OptionHTTPClient(
			&http.Client{
				Transport: &http.Transport{
					Proxy: http.ProxyURL(proxyUrl),
				},
			},
		))
	}

	for _, ch := range cfg.JoinChannels {
		if !strings.HasPrefix(ch, "#") {
			continue
		}
		// TODO: support private channels
		_, err := client.JoinChannel(ch)

		if err != nil {
			logrus.Warnf("join channel %s failed: %v", ch, err)
			continue
		}
		channel := st.NewChannel(ch)
		channel.SetSendFn(func(msg *storage.Message) {
			iconURL := "https://www.gravatar.com/avatar/"
			h := md5.New()
			if _, err := io.WriteString(h, fmt.Sprintf(cfg.GravatarMail, msg.From)); err == nil {
				iconURL += string(h.Sum(nil))
			}
			var content slack.MsgOption
			if msg.Markdown != "" {
				content = slack.MsgOptionText(msg.Markdown, false)
			} else {
				content = slack.MsgOptionText(msg.Text, false)
			}
			if _, _, _, err := client.SendMessage(channel.GetName(), content,
				slack.MsgOptionAsUser(false),
				slack.MsgOptionIconURL(iconURL),
				slack.MsgOptionUsername(msg.From),
			); err != nil {
				logrus.Errorf("send msg to %s failed: %v", channel.GetName(), err)
			}
		})
	}

	select {}
}
