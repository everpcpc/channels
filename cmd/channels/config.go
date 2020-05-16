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
	"channels/web"
)

type config struct {
	IRC   *irc.Config
	Slack *slack.Config
	Web   *web.Config

	LDAP    *auth.LDAPAuth
	Storage string
	Redis   *storage.RedisConfig

	SentryDSN string
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
