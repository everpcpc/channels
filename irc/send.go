package irc

import (
	"fmt"

	"mcdc/state"
)

// sendNumeric sends a numeric response to the given sink.
func sendNumeric(s state.State, sink sink, msg message, extra ...string) {
	sink(msg.withPrefix(s.GetName()).withParams(extra...))
}

func sendNumericUser(s state.State, user *state.User, sink sink, msg message, extra ...string) {
	params := make([]string, len(extra)+1)
	params = append(params, user.GetName())
	params = append(params, extra...)

	sendNumeric(s, sink, msg, params...)
}

// sendIntro sends all of the welcome messages that clients expect to receive
// after joining the server.
// TODO: allow customize
func sendIntro(s state.State, user *state.User, sink sink) {
	sendNumericUser(s, user, sink, replyWelcome.withTrailing("welcome"))
	sendNumericUser(s, user, sink, replyYourHost.withTrailing("your host"))

	sendNumericUser(s, user, sink, replyMOTDStart.withTrailing(fmt.Sprintf("- %s Message of the day - ", s.GetName())))
	sendNumericUser(s, user, sink, replyMOTDStart.withTrailing("--------------------------"))
}

// sendNames sends the messages associated with a NAMES request.
func sendNames(s state.State, user *state.User, sink sink, channels ...*state.Channel) {
	for _, channel := range channels {
		params := []string{"=", channel.GetName()}
		sendNumericUser(s, user, sink, replyNamReply.withTrailing(user.GetName()), params...)
		sendNumericUser(s, user, sink, replyEndOfNames.withTrailing("End NAMES"), channel.GetName())
	}
}
