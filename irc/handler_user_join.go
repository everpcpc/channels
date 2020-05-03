package irc

import (
	"mcdc/state"
	"strings"
)

func (h *userHandler) handleCmdJoin(s state.State, user *state.User, conn connection, msg message) handler {
	if len(msg.params) == 0 {
		sendNumeric(s, user, errorNeedMoreParams)
		return h
	}
	channels := strings.Split(msg.params[0], ",")

	for i := 0; i < len(channels); i++ {
		name := channels[i]
		channel := s.GetChannel(name)
		if channel == nil {
			channel = s.NewChannel(name)
			defer s.RecycleChannel(channel)
		}

		if channel == nil {
			sendNumeric(s, user, errorNoSuchChannel, name)
			continue
		}

		s.JoinChannel(channel, user)
	}

	return h
}
