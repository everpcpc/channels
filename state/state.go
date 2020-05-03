package state

import "github.com/sirupsen/logrus"

// State represents the state of this server.
// State is not safe for concurrent access.
type State interface {

	// GetName get the server name
	GetName() string

	// GetChannel returns a pointer to the channel struct corresponding to the
	// given channel name.
	GetChannel(string) *Channel

	// GetUser returns a pointer to the user struct corresponding to the given
	// nickname.
	GetUser(string) *User

	// NewUser creates a new user with the given nickname and the appropriate
	// default values.
	NewUser(string) *User

	// RemoveUser removes a user from this server. In addition, it forces the
	// user to part from all channels that they are in.
	RemoveUser(*User)

	// NewChannel creates a new channel with the given name and the appropriate
	// default values.
	NewChannel(string) *Channel

	// RecycleChannel removes a channel if there are no more joined users.
	RecycleChannel(*Channel)

	// JoinChannel adds a user to a channel. It does not perform any permissions
	// checking, it only updates pointers.
	JoinChannel(*Channel, *User)

	// PartChannel removes a user from this channel. It sends a parting message to
	// all remaining members of the channel, and removes the channel if there are
	// no remaining users.
	PartChannel(*Channel, *User, string)

	// RemoveFromChannel silently removes a user from the given channel. It does
	// not send any messages to the channel or user. The channel will also be
	// reaped if there are no active users left.
	RemoveFromChannel(*Channel, *User)
}

// stateImpl is a concrete implementation of the State interface.
type stateImpl struct {
	name     string
	channels map[string]*Channel
	users    map[string]*User
}

func New(name string) State {
	return &stateImpl{
		name:     name,
		channels: make(map[string]*Channel),
		users:    make(map[string]*User),
	}
}

func (s *stateImpl) GetName() string {
	return s.name
}

func (s *stateImpl) GetChannel(name string) *Channel {
	return s.channels[lowercase(name)]
}

func (s *stateImpl) GetUser(name string) *User {
	return s.users[lowercase(name)]
}

func (s *stateImpl) NewUser(name string) *User {
	nameLower := lowercase(name)
	if s.users[nameLower] != nil {
		return nil
	}

	logrus.Debugf("Adding new user %s", name)

	u := &User{
		name:     name,
		channels: make(map[*Channel]bool),
	}
	s.users[nameLower] = u
	return u
}

func (s *stateImpl) RemoveUser(user *User) {
	logrus.Debugf("Removing user %s", user.name)

	user.forChannels(func(ch *Channel) {
		s.PartChannel(ch, user, "QUITing")
	})

	nameLower := lowercase(user.name)
	delete(s.users, nameLower)
}

func (s *stateImpl) NewChannel(name string) *Channel {
	name = lowercase(name)
	if s.channels[name] != nil {
		return nil
	}

	if name[0] != '#' && name[0] != '&' {
		return nil
	}

	ch := &Channel{
		name:  name,
		users: make(map[*User]bool),
	}
	s.channels[name] = ch
	return ch
}

func (s *stateImpl) RecycleChannel(channel *Channel) {
	logrus.Debugf("Recycling channel %+v", channel)

	if channel == nil || len(channel.users) != 0 {
		return
	}
	delete(s.channels, channel.name)
}

func (s *stateImpl) JoinChannel(channel *Channel, user *User) {
	// Don't add a user to a channel more than once.
	if channel.users[user] {
		return
	}

	logrus.Debugf("Adding %+v to %+v", user, channel)

	channel.users[user] = true
	user.channels[channel] = true
}

func (s *stateImpl) PartChannel(channel *Channel, user *User, reason string) {
	s.RemoveFromChannel(channel, user)
}

func (s *stateImpl) RemoveFromChannel(channel *Channel, user *User) {
	logrus.Debugf("Removing %+v from %+v", user, channel)

	delete(user.channels, channel)

	if !channel.users[user] {
		return
	}

	delete(channel.users, user)

	s.RecycleChannel(channel)
}
