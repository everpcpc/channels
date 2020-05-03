package irc

import (
	"mcdc/state"
)

// sendNumeric sends a numeric response to the given sink.
func sendNumeric(s state.State, sink sink, msg message, extra ...string) {
	sink(msg.withPrefix(s.GetName()).withParams(extra...))
}

func sendNumericUser(s state.State, user *state.User, sink sink, msg message, extra ...string) {
	extra = append(extra, user.GetName())
	sendNumeric(s, sink, msg, extra...)
}
