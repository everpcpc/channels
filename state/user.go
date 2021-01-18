package state

import (
	"fmt"
	"strings"

	"channels/storage"
)

type User struct {
	name     string
	channels map[*Channel]bool
	roles    []string
	caps     map[string]struct{}

	send func(*storage.Message, *map[string]struct{})
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

// set client caps for user
// ref: https://ircv3.net/specs/core/capability-negotiation.html
func (u *User) AddCap(cap string) {
	u.caps[cap] = struct{}{}
}

func (u *User) GetCaps() map[string]struct{} {
	return u.caps
}
