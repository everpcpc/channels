package state

// Sink is a abstraction over network connections that is able to take a message
// and forward it to the appropriated connection.
type Sink interface {
	Send(SinkMessage)
}

// nullSink is an implementation of sink that drops all messages on the floor.
type nullSink struct{}

func (_ nullSink) Send(msg SinkMessage) {}

// sliceSink is an implementation of sink that stores all received messages in a
// slice.
type sliceSink []SinkMessage

func (s *sliceSink) Send(msg SinkMessage) {
	*s = append(*s, msg)
}
