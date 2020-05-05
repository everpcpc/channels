package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"

	"github.com/sirupsen/logrus"

	"channels/api"
	"channels/auth"
	"channels/irc"
	"channels/storage"
)

type config struct {
	IRC         *irc.Config
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

func main() {
	flagConfig := flag.String("config", "config.json", "config file to load")
	flagLogLevel := flag.String("log", "info", "loglevel: debug, info, warn, error")
	flagRun := flag.String("run", "", "choose: irc, api")
	flag.Parse()

	switch *flagLogLevel {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	}

	cfg := readConfig(*flagConfig)

	var store storage.Backend
	var tokenStore storage.TokenBackend
	var err error

	switch cfg.Storage {
	case "redis":
		store, err = storage.NewRedisBackend(cfg.Redis)
		if err != nil {
			logrus.Fatal(err)
		}
		tokenStore, err = storage.NewRedisBackend(cfg.Redis)
		if err != nil {
			logrus.Fatal(err)
		}
	default:
		logrus.Fatalf("storage %s not supported", cfg.Storage)
	}

	switch *flagRun {
	case "irc":
		if cfg.AuthIRC == "ldap" {
			irc.RunServer(cfg.IRC, cfg.LDAP, store)
		} else {
			irc.RunServer(cfg.IRC, &auth.Anonymous{}, store)
		}
	case "api":
		var webhookAuth auth.Plugin
		if cfg.AuthWebhook == "token" {
			webhookAuth = &auth.TokenAuth{Store: tokenStore}
		} else {
			webhookAuth = &auth.Anonymous{}
		}
		api.RunServer(cfg.APIPort, &auth.Anonymous{}, webhookAuth, store)
	default:
		flag.PrintDefaults()
	}
}
