package irc

import "mcdc/state"

type commandMap map[string]func(state.State, *state.User, connection, message) handler

// userHandler is a handler that handles messages coming from a user connection
// that has successfully associated with the client.
type userHandler struct {
	state    chan state.State
	nick     string
	commands commandMap
}

func newUserHandler(s chan state.State, nick string) handler {
	handler := &userHandler{
		state: s,
		nick:  nick,
	}
	handler.commands = commandMap{
		cmdJoin.command: handler.handleCmdJoin,
		cmdPart.command: handler.handleCmdPart,
		cmdPing.command: handler.handleCmdPing,
		cmdQuit.command: handler.handleCmdQuit,

		cmdWho.command:     handler.handleCmdDummy,
		cmdPong.command:    handler.handleCmdDummy,
		cmdAway.command:    handler.handleCmdDummy,
		cmdNames.command:   handler.handleCmdDummy,
		cmdPrivMsg.command: handler.handleCmdDummy,
	}
	return handler
}

func (h *userHandler) closed(conn connection) {
	s := <-h.state
	defer func() { h.state <- s }()

	s.RemoveUser(s.GetUser(h.nick))
	conn.kill()
}

func (h *userHandler) handle(conn connection, msg message) handler {
	s := <-h.state
	defer func() { h.state <- s }()

	command := h.commands[msg.command]
	if command == nil {
		logf(warn, "unknown command from %s: %s\n", h.nick, msg.command)
		return h
	}

	user := s.GetUser(h.nick)
	newHandler := command(s, user, conn, msg)
	h.nick = user.GetName()
	return newHandler
}

func (h *userHandler) handleCmdDummy(s state.State, user *state.User, conn connection, msg message) handler {
	return h
}
