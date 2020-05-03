package irc

import "mcdc/state"

func (h *userHandler) handleCmdQuit(s state.State, user *state.User, conn connection, msg message) handler {
	s.RemoveUser(user)
	conn.kill()
	return nullHandler{}
}
