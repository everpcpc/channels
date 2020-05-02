package irc

type user struct {
	nick string
	user string

	channels map[*channel]bool

	sink sink
}

func (u user) send(msg message) {
	u.sink.send(msg)
}

// forChannels iterates over all of the channels that the user has joined and
// passes a pointer to each to the supplied callback.
func (u user) forChannels(callback func(*channel)) {
	for ch := range u.channels {
		callback(ch)
	}
}
