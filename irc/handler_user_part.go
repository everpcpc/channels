package irc

import (
	"mcdc/state"
	"strings"
)

func (h *userHandler) handleCmdPart(s state.State, user *state.User, conn connection, msg message) handler {
	if len(msg.params) == 0 {
		sendNumeric(s, user, errorNeedMoreParams)
		return h
	}

	reason := msg.laxTrailing(1)
	channels := strings.Split(msg.params[0], ",")
	for i := 0; i < len(channels); i++ {
		name := channels[i]
		channel := s.GetChannel(name)

		if channel == nil {
			sendNumeric(s, user, errorNoSuchChannel, name)
			continue
		}

		if !channel.HasUser(user) {
			sendNumeric(s, user, errorNotOnChannel, name)
			continue
		}

		s.PartChannel(channel, user, reason)
	}
	return h
}
