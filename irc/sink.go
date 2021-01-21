package irc

import (
	"strings"
	"time"

	"channels/state"
	"channels/storage"
)

const (
	MaxMessageLength    = 512
	MaxMessageTagLength = 8191
)

type sink func(message)

func messageSink(conn connection, caps state.Capbilities) func(msg *storage.Message) {
	return func(msg *storage.Message) {
		// ignore message from self
		if msg.Source == storage.MessageSourceIRC {
			return
		}
		target := msg.GetTarget()
		text := strings.NewReplacer("\n", " ", "\r", " ").Replace(msg.Text)

		// :nick PRIVMSG #channel :text\r\n
		// with a safety buffer of 10
		maxLength := MaxMessageLength - len(msg.From) - len(target) - 14 - 10

		send := func(s string) {
			msgToSend := cmdPrivMsg.
				withMessageTag(msg, caps).
				withPrefix(msg.From).
				withParams(target).
				withTrailing(s)
			conn.send(msgToSend)
		}

		if len(text) <= maxLength {
			send(text)
			return
		}

		chunks := len(text) / maxLength
		for i := 0; i < chunks; i++ {
			send(text[i*maxLength : (i+1)*maxLength])
		}
		send(text[chunks*maxLength:])
	}
}

func sendMessageBack(s state.State, user *state.User, ircMsg *message, target, text string) error {
	m := storage.Message{
		Source:    storage.MessageSourceIRC,
		From:      user.GetName(),
		To:        target,
		Text:      text,
		Timestamp: time.Now().UnixNano(),
		IsHuman:   true,
	}

	// caps handler to handle caps for storage
	caps := user.GetCaps()
	for c, _ := range caps {
		handler := supportedCaps[c]
		handler.toStorageMsg(ircMsg, &m, caps)
	}

	return s.SendMessage(&m)
}
