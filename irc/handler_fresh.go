package irc

import (
	"github.com/sirupsen/logrus"
	"strings"

	"channels/state"
)

// freshHandler is a handler for a brand new connection that has not been
// registered yet.
type freshHandler struct {
	state       chan state.State
	pass        string
	v3CapClient bool
	capEnd      bool
	caps        state.Capbilities
}

func newFreshHandler(s chan state.State) handler {
	return &freshHandler{state: s, caps: make(map[string]struct{})}
}

func (h *freshHandler) handle(conn connection, msg message) handler {
	if msg.command == cmdQuit.command {
		conn.kill()
		return nullHandler{}
	}
	switch msg.command {
	case cmdNick.command:
		return h.handleNick(conn, msg)
	case cmdPass.command:
		return h.handlePass(conn, msg)
	case cmdCap.command:
		return h.handleCap(conn, msg)
	default:
		return h
	}
}

func (_ *freshHandler) closed(c connection) {
	c.kill()
}

func (h *freshHandler) handleCap(conn connection, msg message) handler {
	s := <-h.state
	defer func() { h.state <- s }()

	if len(msg.params) < 1 {
		sendNumeric(s, conn.send, errorNeedMoreParams)
	} else {
		logrus.Debugf("get msg: %v", msg)
		h.v3CapClient = true
		h.capEnd = false
		switch msg.params[0] {
		case "LS":
			SendServerCap(conn.send, msg, ServerCapsLsResp, "*", "LS")
		case "REQ":
			requiredCaps := strings.Split(msg.trailing, " ")
			if len(requiredCaps) < 1 {
				sendNumeric(s, conn.send, errorInvalidCapCmd)
			}
			var ackCaps []string
			var nakCaps []string
			for _, cap := range requiredCaps {
				if _, ok := supportedCaps[cap]; ok {
					ackCaps = append(ackCaps, cap)
				} else {
					nakCaps = append(nakCaps, cap)
				}
			}

			// cap grant
			if len(ackCaps) > 0 {
				SendServerCap(conn.send, msg, strings.Join(ackCaps, " "), "*", "ACK")
				for _, cap := range ackCaps {
					h.caps[cap] = struct{}{}
				}
			}
			// cap denied
			if len(nakCaps) > 0 {
				SendServerCap(conn.send, msg, strings.Join(nakCaps, " "), "*", "NAK")
			}
		case "END":
			h.capEnd = true
		default:
			sendNumeric(s, conn.send, errorInvalidCapCmd)
		}
	}
	return h
}

func (h *freshHandler) handlePass(conn connection, msg message) handler {
	s := <-h.state
	defer func() { h.state <- s }()

	if len(msg.params) < 1 {
		sendNumeric(s, conn.send, errorNeedMoreParams)
	} else {
		h.pass = msg.params[0]
	}
	return h
}

func (h *freshHandler) handleNick(conn connection, msg message) handler {
	s := <-h.state
	defer func() { h.state <- s }()

	if len(msg.params) < 1 {
		sendNumeric(s, conn.send, errorNoNicknameGiven)
		return h
	}
	nick := msg.params[0]
	if h.pass == "" {
		sendNumeric(s, conn.send, errorPasswdMismatch)
		return h
	}

	caller, err := s.Auth(nick, h.pass)
	if err != nil {
		logrus.Debugf("login failed %s: %v", nick, err)
		sendNumeric(s, conn.send, errorPasswdMismatch)
		return h
	}
	if caller.Name != nick {
		sendNumeric(s, conn.send, errorNickCollision)
		return h
	}

	user := s.NewUser(nick)
	if user == nil {
		sendNumeric(s, conn.send, errorNicknameInUse)
		return h
	}

	if h.capEnd {
		for c, _ := range h.caps {
			err := s.SetUserCap(user, c)
			if err != nil {
				logrus.Warnf("set cap %s for user %s error: %v", c, user.GetName(), err)
			}
		}
	}

	user.AddRoles(caller.Roles...)
	user.SetSendFn(messageSink(conn, user.GetCaps()))

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

	sendIntro(s, h.user, conn.send)

	return newUserHandler(h.state, h.user.GetName())
}
