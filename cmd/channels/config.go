package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/sirupsen/logrus"

	"channels/auth"
	"channels/irc"
	"channels/slack"
	"channels/storage"
)

type config struct {
	IRC         *irc.Config
	Slack       *slack.Config
	AuthIRC     string `json:"auth.irc"`
	AuthWebhook string `json:"auth.webhook"`
	LDAP        *auth.LDAPAuth
	APIPort     int `json:"api.port"`
	Storage     string
	Redis       *storage.RedisConfig
}

func readConfig(f string) *config {
	cfgFile, err := os.Open(f)
	if err != nil {
		logrus.Fatalf("read config file error: %v", err)
	}
	defer cfgFile.Close()

	var cfg config
	content, _ := ioutil.ReadAll(cfgFile)
	err = json.Unmarshal([]byte(content), &cfg)
	if err != nil {
		logrus.Fatalf("read config error: %v", err)
	}

	return &cfg
}
