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
func sendMOTD(s state.State, sink state.Sink) {
	sendNumericTrailing(s, sink, replyMOTDStart, fmt.Sprintf(motdHeader, s.GetName()))
	sendNumericTrailing(s, sink, replyEndOfMOTD, motdFooter)
}
