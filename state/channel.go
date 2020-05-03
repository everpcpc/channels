package state

type Channel struct {
	name string

	users map[*User]bool
}

func (ch *Channel) Send(msg SinkMessage) {
	for user := range ch.users {
		user.sink.Send(msg)
	}
}

func (ch *Channel) HasUser(user *User) bool {
	return ch.users[user]
}
