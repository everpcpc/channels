package irc

import (
	"fmt"

	"mcdc/storage"
)

type sink func(message)

func messageSink(conn connection) func(msg *storage.Message) {
	return func(msg *storage.Message) {
		msgToSend := cmdPrivMsg.
			withPrefix(fmt.Sprintf("%s!%s", msg.From, msg.From)).
			withParams(msg.Channel).
			withTrailing(msg.Text)

		conn.send(msgToSend)

	}
}
