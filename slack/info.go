package slack

import (
	"strings"

	"channels/storage"
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
