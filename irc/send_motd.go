package irc

import (
	"fmt"
)

const (
	motdHeader = "- %s Message of the day - "
	motdFooter = "--------------------------"
)

// sendMOTD will send the message of the day to a relay.
func sendMOTD(state state, sink sink) {
	sendNumericTrailing(state, sink, replyMOTDStart,
		fmt.Sprintf(motdHeader, state.getConfig().Name))

	sendNumericTrailing(state, sink, replyEndOfMOTD, motdFooter)
}
