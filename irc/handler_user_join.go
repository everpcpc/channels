package irc

import (
	"strings"
)

func (h *userHandler) handleCmdJoin(state state, user *user, conn connection, msg message) handler {
	if len(msg.params) == 0 {
		sendNumeric(state, user, errorNeedMoreParams)
		return h
	}
	channels := strings.Split(msg.params[0], ",")

	for i := 0; i < len(channels); i++ {
		name := channels[i]
		channel := state.getChannel(name)
		if channel == nil {
			channel = state.newChannel(name)
			defer state.recycleChannel(channel)
		}

		if channel == nil {
			sendNumeric(state, user, errorNoSuchChannel, name)
			continue
		}

		state.joinChannel(channel, user)
	}

	return h
}
