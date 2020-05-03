package irc

import (
	"fmt"

	"mcdc/state"
)

const (
	motdHeader = "- %s Message of the day - "
	motdFooter = "--------------------------"
)

// sendMOTD will send the message of the day to a relay.
func sendMOTD(s state.State, user *state.User, sink sink) {
	sendNumericUser(s, user, sink, replyMOTDStart.withTrailing(fmt.Sprintf(motdHeader, s.GetName())))
	sendNumericUser(s, user, sink, replyMOTDStart.withTrailing(motdFooter))
}
