package irc

import (
	"fmt"
	"strings"

	"channels/storage"
)

const (
	IRC_MAX_MESSAGE_LENGTH = 512
)

type sink func(message)

func messageSink(conn connection) func(msg *storage.Message) {
	return func(msg *storage.Message) {
		replacer := strings.NewReplacer("\n", " ", "\r", " ")
		text := replacer.Replace(msg.Text)
		msgToSend := cmdPrivMsg.
			withPrefix(fmt.Sprintf("%s!%s", msg.From, msg.From)).
			withParams(msg.GetTarget()).
			withTrailing(text)

		conn.send(msgToSend)
	}
}
