package state

import (
	"fmt"
	"strings"

	"mcdc/storage"
)

type User struct {
	name     string
	channels map[*Channel]bool
	roles    []string

	send func(*storage.Message)
}

func (u *User) String() string {
	return fmt.Sprintf("<User:%s@%s>", u.name, strings.Join(u.roles, ","))
}

func (u *User) GetName() string {
	return u.name
}

func (u *User) AddRoles(roles ...string) {
	u.roles = append(u.roles, roles...)
}

func (u *User) GetChannels() []*Channel {
	channels := make([]*Channel, len(u.channels))
	for ch := range u.channels {
		channels = append(channels, ch)
	}
	return channels
}

func (u *User) SetSendFn(fn func(*storage.Message)) {
	u.send = fn
}

// forChannels iterates over all of the channels that the user has joined and
// passes a pointer to each to the supplied callback.
func (u *User) forChannels(callback func(*Channel)) {
	for ch := range u.channels {
		callback(ch)
	}
}
