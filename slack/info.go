package slack

import "strings"

func (c *Client) GetUserName(uid string) (string, error) {
	profile, err := c.api.GetUserProfile(uid, false)
	if err != nil {
		return "", err
	}
	emailPrefix := strings.SplitN(profile.Email, "@", 2)[0]
	return strings.SplitN(emailPrefix, "+", 2)[0], nil
}

func (c *Client) GetChannelName(cid string) (string, error) {
	channel, err := c.api.GetChannelInfo(cid)
	if err != nil {
		return "", err
	}
	return "#" + channel.Name, nil
}
