package slack

import (
	"fmt"
	"regexp"
	"strings"

	"channels/storage"
)

var (
	reMention = regexp.MustCompile(`<(@|#)([0-9A-Z]+)>`)
)

func (c *Client) GetUserName(uid string) (string, error) {
	return storage.Cached(c.cache, "slack:user:"+uid, func() (string, error) {
		profile, err := c.api.GetUserProfile(uid, false)
		if err != nil {
			return "", err
		}
		emailPrefix := strings.SplitN(profile.Email, "@", 2)[0]
		username := strings.SplitN(emailPrefix, "+", 2)[0]
		return username, nil
	})
}

func (c *Client) GetChannelName(cid string) (string, error) {
	return storage.Cached(c.cache, "slack:channel:"+cid, func() (string, error) {
		channel, err := c.api.GetChannelInfo(cid)
		if err != nil {
			return "", err
		}
		return "#" + channel.Name, nil
	})
}

func (c *Client) TranslateMentions(text string) string {
	return replaceAllStringSubmatchFunc(reMention, text, func(groups []string) string {
		indicator := groups[1]
		name := groups[2]
		switch indicator {
		case "#":
			if n, err := c.GetChannelName(name); err == nil {
				name = n
			}
		case "@":
			if n, err := c.GetUserName(name); err == nil {
				name = n
			}
		}
		return fmt.Sprintf("<%s%s>", indicator, name)
	})
}
