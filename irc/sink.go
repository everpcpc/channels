package irc

import (
	"strings"

	"channels/storage"
)

const (
	IRC_MAX_MESSAGE_LENGTH = 512
)

type sink func(message)

func messageSink(conn connection) func(msg *storage.Message) {
	return func(msg *storage.Message) {
		target := msg.GetTarget()
		text := strings.NewReplacer("\n", " ", "\r", " ").Replace(msg.Text)

		// :nick PRIVMSG #channel :text\r\n
		// with a safety buffer of 10
		maxLength := IRC_MAX_MESSAGE_LENGTH - len(msg.From) - len(target) - 14 - 10

		send := func(s string) {
			msgToSend := cmdPrivMsg.
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
