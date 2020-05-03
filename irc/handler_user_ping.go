package irc

import "mcdc/state"

func (h *userHandler) handleCmdPing(s state.State, user *state.User, conn connection, msg message) handler {
	name := s.GetName()
	conn.Send(cmdPong.withPrefix(name).withParams(name).withTrailing(name))
	return h
}
