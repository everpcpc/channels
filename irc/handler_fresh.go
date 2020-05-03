package irc

import (
	"mcdc/state"

	"github.com/sirupsen/logrus"
)

// freshHandler is a handler for a brand new connection that has not been
// registered yet.
type freshHandler struct {
	state chan state.State
}

func newFreshHandler(s chan state.State) handler {
	return &freshHandler{state: s}
}

func (h *freshHandler) handle(conn connection, msg message) handler {
	if msg.command == cmdQuit.command {
		conn.kill()
		return nullHandler{}
	}
	if msg.command != cmdNick.command {
		return h
	}
	return h.handleNick(conn, msg)
}

func (_ *freshHandler) closed(c connection) {
	c.kill()
}

func (h *freshHandler) handleNick(conn connection, msg message) handler {
	s := <-h.state
	defer func() { h.state <- s }()

	if len(msg.params) < 1 {
		sendNumeric(s, conn.send, errorNoNicknameGiven)
		return h
	}
	nick := msg.params[0]

	user := s.NewUser(nick)
	if user == nil {
		sendNumeric(s, conn.send, errorNicknameInUse)
		return h
	}

	user.SetSendFn(messageSink(conn))

	return &freshUserHandler{state: h.state, user: user}
}

// freshUserHandler is a handler for a brand new connection that is in the
// process of registering and has successfully set a nickname.
type freshUserHandler struct {
	user  *state.User
	state chan state.State
}

func (h *freshUserHandler) handle(conn connection, msg message) handler {
	if msg.command == cmdQuit.command {
		s := <-h.state
		s.RemoveUser(h.user)
		h.state <- s
		conn.kill()
		return nullHandler{}
	}
	if msg.command != cmdUser.command {
		return h
	}
	return h.handleUser(conn, msg)
}

func (h *freshUserHandler) closed(c connection) {
	s := <-h.state
	defer func() { h.state <- s }()

	s.RemoveUser(h.user)
	c.kill()
}

func (h *freshUserHandler) handleUser(conn connection, msg message) handler {
	s := <-h.state
	defer func() { h.state <- s }()

	var trailing = msg.laxTrailing(3)
	if len(msg.params) < 3 || trailing == "" {
		sendNumericUser(s, h.user, conn.send, errorNeedMoreParams)
		return h
	}

	logrus.Debugf("handleUser: %+v", msg)

	sendMOTD(s, h.user, conn.send)

	return newUserHandler(h.state, h.user.GetName())
}
