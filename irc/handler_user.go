package irc

type commandMap map[string]func(state, *user, connection, message) handler

// userHandler is a handler that handles messages coming from a user connection
// that has successfully associated with the client.
type userHandler struct {
	state    chan state
	nick     string
	commands commandMap
}

func newUserHandler(state chan state, nick string) handler {
	handler := &userHandler{
		state: state,
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
	state := <-h.state
	defer func() { h.state <- state }()

	state.removeUser(state.getUser(h.nick))
	conn.kill()
}

func (h *userHandler) handle(conn connection, msg message) handler {
	state := <-h.state
	defer func() { h.state <- state }()

	command := h.commands[msg.command]
	if command == nil {
		logf(warn, "unknown command from %s: %s\n", h.nick, msg.command)
		return h
	}

	user := state.getUser(h.nick)
	newHandler := command(state, user, conn, msg)
	h.nick = user.nick
	return newHandler
}

func (h *userHandler) handleCmdDummy(state state, user *user, conn connection, msg message) handler {
	return h
}
