package storage

import "strings"

const (
	MessageSourceSlack   = "slack"
	MessageSourceIRC     = "irc"
	MessageSourceWebhook = "webhook"
)

type Message struct {
	Source               string
	From                 string
	To                   string
	Text                 string
	Title                string
	Link                 string
	Color                string
	Markdown             string
	SlackThreadTimeStamp string
	SlackTimeStamp       string
	Timestamp            int64
	IsHuman              bool
}

func (m *Message) GetTarget() string {
	return strings.TrimLeft(m.To, "@")
}
