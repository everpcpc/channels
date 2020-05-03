package irc

import (
	"reflect"

	"mcdc/state"
)

// sendNumeric sends a numeric response to the given sink. If the sink is of
// type User
func sendNumeric(s state.State, sink state.Sink, msg message, extra ...string) {
	params := make([]string, 0, len(extra)+1)

	// Attempt to add the nick name of the current sink to the error message.
	sinkType := reflect.TypeOf(sink)
	switch sinkType {
	case reflect.TypeOf(&state.User{}):
		params = append(params, sink.(*state.User).GetName())
	}

	params = append(params, extra...)
	sink.Send(msg.withPrefix(s.GetName()).withParams(params...))
}

// sendNumericTrailing sends a numeric response to the given client.
func sendNumericTrailing(s state.State, sink state.Sink, msg message, trailing string, extra ...string) {
	sendNumeric(s, sink, msg.withTrailing(trailing), extra...)
}
