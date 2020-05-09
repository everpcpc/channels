package state

import (
	"fmt"

	"channels/storage"
)

type Channel struct {
	name string

	users map[*User]bool

	send func(*storage.Message)
}

func (ch *Channel) String() string {
	return fmt.Sprintf("<Channel:%s>", ch.name)
}

func (ch *Channel) GetName() string {
	return ch.name
}

func (ch *Channel) SetSendFn(fn func(*storage.Message)) {
	ch.send = fn
}

func (ch *Channel) SendUsers(msg *storage.Message) {
	for user := range ch.users {
		user.send(msg)
	}
}

func (ch *Channel) HasUser(user *User) bool {
	return ch.users[user]
}
