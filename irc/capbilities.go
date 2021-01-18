package irc

import (
	"channels/storage"
	"fmt"
)

const CAP_MSG_TAG_PREFIX = "channels"

type capbilityHandler interface {
	handle(*storage.Message) string
}

var (
	capMsgTag = "message-tag"
)

type capMsgTagHandler struct {
	CapName string
}

func (c *capMsgTagHandler) handle(msg *storage.Message) string {
	return fmt.Sprintf(
		"@%s/thread_ts=%s;%s/ts=%s",
		CAP_MSG_TAG_PREFIX,
		msg.SlackThreadTimeStamp,
		CAP_MSG_TAG_PREFIX,
		msg.SlackTimeStamp,
	)
}

var (
	supportedCaps = map[string]capbilityHandler{
		capMsgTag: &capMsgTagHandler{
			CapName: capMsgTag,
		},
	}
)
