package state

import (
	"fmt"

	"mcdc/storage"
)

type User struct {
	name     string
	channels map[*Channel]bool
	send     func(*storage.Message)
}

func (u *User) String() string {
	return fmt.Sprintf("<User:%s>", u.name)
}

func (u *User) GetName() string {
	return u.name
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
