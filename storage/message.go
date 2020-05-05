package storage

import "strings"

type Message struct {
	From string `json:"from,omitempty"`
	To   string `json:"to,omitempty"`
	Text string `json:"text,omitempty"`
}

func (m *Message) IsChannel() bool {
	return strings.HasPrefix(m.To, "#")
}

func (m *Message) IsPrivate() bool {
	return strings.HasPrefix(m.To, "@")
}

func (m *Message) GetTarget() string {
	if m.IsPrivate() {
		return strings.TrimLeft(m.To, "@")
	}
	return m.To
}
