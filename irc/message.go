package irc

import (
	"strings"

	"channels/state"
	"channels/storage"
)

type message struct {
	msgTag   map[string]string
	prefix   string
	command  string
	params   []string
	trailing string
}

func (m message) withMessageTag(msg *storage.Message, caps state.Capbilities) message {
	handler := supportedCaps[capMsgTag]
	if _, ok := caps[capMsgTag]; ok {
		return *handler.handle(msg, &m)
	} else {
		m.msgTag = nil
	}
	return m
}

// withParams creates a new copy of a message with the given parameters.
func (m message) withParams(params ...string) message {
	m.params = params
	return m
}

// withTrailing creates a new copy of a message with the given parameters.
func (m message) withTrailing(trailing string) message {
	m.trailing = trailing
	return m
}

// withPrefix creates a new copy of a message with the given prefix.
func (m message) withPrefix(prefix string) message {
	m.prefix = prefix
	return m
}

// laxTrailing returns the trailing portion of an IRC message or the last
// parameter.
func (m message) laxTrailing(minIndex int) string {
	if m.trailing != "" {
		return m.trailing
	}

	l := len(m.params)
	if l <= minIndex {
		return ""
	}

	return m.params[l-1]
}

// toString serializes a Message to an IRC protocol compatible string.
func (m message) toString() (string, bool) {
	if m.command == "" {
		return "", false
	}

	var msg string

	if m.msgTag != nil {
		handler := supportedCaps[capMsgTag]
		msg = handler.toString(&m)
	} else {
		msg = ""
	}

	if len(m.prefix) > 0 {
		msg += ":" + m.prefix + " "
	}

	msg += m.command

	for i := 0; i < len(m.params); i++ {
		param := m.params[i]
		if strings.Contains(param, " ") {
			return "", false
		}
		msg += " " + param
	}

	if m.trailing != "" {
		msg += " :" + m.trailing
	}

	msg += "\r\n"

	return msg, true
}
