package irc

type channel struct {
	name string

	users map[*user]bool
}

func (ch channel) send(msg message) {
	for user := range ch.users {
		user.sink.send(msg)
	}
}

// forUsers iterates over all of the users in the channel and passes a pointer
// to each to the supplied callback.
func (ch channel) forUsers(callback func(*user)) {
	for u := range ch.users {
		callback(u)
	}
}
