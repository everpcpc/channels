package state

type User struct {
	name     string
	channels map[*Channel]bool
	sink     Sink
}

func (u *User) GetName() string {
	return u.name
}

func (u *User) Send(msg SinkMessage) {
	u.sink.Send(msg)
}

func (u *User) AddSink(sink Sink) {
	u.sink = sink
}

// forChannels iterates over all of the channels that the user has joined and
// passes a pointer to each to the supplied callback.
func (u *User) forChannels(callback func(*Channel)) {
	for ch := range u.channels {
		callback(ch)
	}
}
