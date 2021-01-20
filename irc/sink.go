package irc

import (
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"channels/state"
	"channels/storage"
)

const (
	IRC_MAX_MESSAGE_LENGTH = 512
	IRC_MAX_MESSAGE_TAG_LENGTH = 8191
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
		maxLength := IRC_MAX_MESSAGE_LENGTH - len(msg.From) - len(target) - 14 - 10

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
	logrus.Debugf("-- sending back ")
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
	for cap, _ := range caps {
		handler := supportedCaps[cap]
		handler.toStorageMsg(ircMsg, &m, caps)
	}
	logrus.Debugf("-- send msg %v back", m)
	
	return s.SendMessage(&m)
}
