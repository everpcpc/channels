package irc

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	"channels/state"
)

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
		cmdJoin.command:  handler.handleCmdJoin,
		cmdNames.command: handler.handleCmdNames,
		cmdPart.command:  handler.handleCmdPart,
		cmdPing.command:  handler.handleCmdPing,
		cmdQuit.command:  handler.handleCmdQuit,
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
		return h
	}

	logrus.Debugf("command: %+v", msg)

	user := s.GetUser(h.nick)

	newHandler := command(s, user, conn, msg)
	h.nick = user.GetName()
	return newHandler
}

func (h *userHandler) handleCmdDummy(s state.State, user *state.User, conn connection, msg message) handler {
	return h
}

func (h *userHandler) handleCmdPing(s state.State, user *state.User, conn connection, msg message) handler {
	name := s.GetName()
	conn.send(cmdPong.withPrefix(name).withParams(name).withTrailing(name))
	return h
}

func (h *userHandler) handleCmdJoin(s state.State, user *state.User, conn connection, msg message) handler {
	if len(msg.params) == 0 {
		sendNumericUser(s, user, conn.send, errorNeedMoreParams)
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
			sendNumericUser(s, user, conn.send, errorNoSuchChannel, name)
			continue
		}

		s.JoinChannel(channel, user)

		conn.send(cmdJoin.withPrefix(fmt.Sprintf(
			"%s!%s@%s", user.GetName(), user.GetName(), s.GetName(),
		)).withParams(channel.GetName()))
		// sendNumericUser(s, user, conn.send, replyNoTopic.withTrailing("no topic"), channel.GetName())
		sendNames(s, user, conn.send, channel)
	}

	return h
}

func (h *userHandler) handleCmdNames(s state.State, user *state.User, conn connection, msg message) handler {
	if len(msg.params) == 0 {
		sendNames(s, user, conn.send, user.GetChannels()...)
		return h
	}
	names := strings.Split(msg.params[0], ",")
	for _, name := range names {
		channel := s.GetChannel(name)
		if channel == nil {
			continue
		}
		sendNames(s, user, conn.send, channel)
	}
	return h
}

func (h *userHandler) handleCmdPart(s state.State, user *state.User, conn connection, msg message) handler {
	if len(msg.params) == 0 {
		sendNumericUser(s, user, conn.send, errorNeedMoreParams)
		return h
	}

	reason := msg.laxTrailing(1)
	channels := strings.Split(msg.params[0], ",")
	for i := 0; i < len(channels); i++ {
		name := channels[i]
		channel := s.GetChannel(name)

		if channel == nil {
			sendNumericUser(s, user, conn.send, errorNoSuchChannel)
			continue
		}

		if !channel.HasUser(user) {
			sendNumericUser(s, user, conn.send, errorNotOnChannel)
			continue
		}

		conn.send(cmdPart.withPrefix(user.GetName()).withParams(channel.GetName()).withTrailing(reason))
		s.PartChannel(channel, user, reason)
	}
	return h
}

func (h *userHandler) handleCmdQuit(s state.State, user *state.User, conn connection, msg message) handler {
	s.RemoveUser(user)
	conn.kill()
	return nullHandler{}
}
