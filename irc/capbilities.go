package irc

import (
	"fmt"
	"strings"
	"strconv"

        "channels/storage"
	"channels/state"
	"github.com/sirupsen/logrus"
)

const (
	CAP_MSG_TAG_PREFIX = "channels"
	CAP_SLACK_THREAD_TS_MSG_TAG_NAME = "channels/slack_thread_ts"
)

type capbilityHandler interface {
	handle(*storage.Message, *message) *message
	toString(*message) string
	toStorageMsg(*message, *storage.Message, state.Capbilities) *storage.Message
}

var (
	capMsgTag = "message-tag"
)

type capMsgTagHandler struct {
	CapName string
}

func (c *capMsgTagHandler) toString(msg *message) string {
	msgTagsKV := msg.msgTag
	
	if len(msgTagsKV) == 0 {
		return ""
	}
	
	msgTags := make([]string, 0)
	for k, v := range(msgTagsKV) {
		msgTags = append(msgTags, fmt.Sprintf("%s=%s", k, v))
	}
	tags := fmt.Sprintf("@%s ", strings.Join(msgTags, ";"))
	
	if len(tags) <= IRC_MAX_MESSAGE_TAG_LENGTH {
		return tags
	} else {
		shrinkedMsgTag := strings.Split(tags, ";")
		logrus.Warnf("msg tag len %d overflow limit, will shrink to %d", len(tags), IRC_MAX_MESSAGE_TAG_LENGTH)
		totalLen := 0
		for index, tag := range(shrinkedMsgTag) {
			// 1 = len(";")
			totalLen += len(tag) + 1
			if totalLen > IRC_MAX_MESSAGE_TAG_LENGTH {
				if (index - 1) <= 0 {
					return fmt.Sprintf("@%s/error:MsgTagsTooLong ", CAP_MSG_TAG_PREFIX)
				} else {
					return strings.Join(shrinkedMsgTag[0:index-1], ";") + " "
				}
			}
		}
	}
	return tags
}

func (c *capMsgTagHandler) handle(msg *storage.Message, ircMsg *message) *message {
	tagsMap := make(map[string]string)
	tagsMap[fmt.Sprintf("%s/slack_thread_ts", CAP_MSG_TAG_PREFIX)] = msg.SlackThreadTimeStamp
	tagsMap[fmt.Sprintf("%s/slack_msg_ts", CAP_MSG_TAG_PREFIX)] = msg.SlackTimeStamp
	tagsMap[fmt.Sprintf("%s/source", CAP_MSG_TAG_PREFIX)] = msg.Source
	(*ircMsg).msgTag = tagsMap
	
	return ircMsg
}

func (c *capMsgTagHandler) toStorageMsg(ircMsg *message, msg *storage.Message, caps state.Capbilities) *storage.Message {
	// if cap message-tag and with channels/slack_thread_ts
	// send with thread_ts field as thread reply
	if _, ok := caps[capMsgTag]; ok {
		if v, exists := ircMsg.msgTag[CAP_SLACK_THREAD_TS_MSG_TAG_NAME]; exists {
			if _, err := strconv.ParseFloat(v, 64); err == nil {
				(*msg).SlackThreadTimeStamp = v
			} else {
				logrus.Warnf("msg tag %s value: %s invalid", CAP_SLACK_THREAD_TS_MSG_TAG_NAME, v)
			}
		} else {
			logrus.Debugf("msg tag %s not exists in irc msg", CAP_SLACK_THREAD_TS_MSG_TAG_NAME)
		}
	} else {
		logrus.Debugf("client cap message-tag not supported")
	}
	
	return msg
}

func GetServerCaps(m map[string]capbilityHandler) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return strings.Join(keys, " ")
}

var (
	supportedCaps = map[string]capbilityHandler{
		capMsgTag: &capMsgTagHandler{
			CapName: capMsgTag,
		},
	}
	SERVER_CAPS_LS_RESP = GetServerCaps(supportedCaps)
)
