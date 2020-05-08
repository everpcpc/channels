package storage

import "strings"

type Message struct {
	From      string
	To        string
	Text      string
	Markdown  string
	Timestamp int64
}

func (m *Message) GetTarget() string {
	return strings.TrimLeft(m.To, "@")
}
