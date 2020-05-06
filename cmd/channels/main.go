package main

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"channels/api"
	"channels/auth"
	"channels/irc"
	"channels/storage"
)

func main() {
	var flagConfig, flagLoglevel string
	var cfg *config
	var store storage.Backend
	var tokenStore storage.TokenBackend
	var err error

	var rootCmd = &cobra.Command{
		Use: "channels",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			switch flagLoglevel {
			case "debug":
				logrus.SetLevel(logrus.DebugLevel)
			case "info":
				logrus.SetLevel(logrus.InfoLevel)
			case "warn":
				logrus.SetLevel(logrus.WarnLevel)
			case "error":
				logrus.SetLevel(logrus.ErrorLevel)
			}

			cfg = readConfig(flagConfig)

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
		},
	}
	rootCmd.Flags().StringVarP(&flagConfig, "config", "c", "config.json", "config file to load")
	rootCmd.Flags().StringVarP(&flagLoglevel, "log", "l", "info", "loglevel: debug, info, warn, error")

	var cmdIRC = &cobra.Command{
		Use:   "irc",
		Short: "Run the irc server",
		Run: func(cmd *cobra.Command, args []string) {
			if cfg.AuthIRC == "ldap" {
				irc.RunServer(cfg.IRC, cfg.LDAP, store)
			} else {
				irc.RunServer(cfg.IRC, &auth.Anonymous{}, store)
			}
		},
	}

	var cmdAPI = &cobra.Command{
		Use:   "api",
		Short: "Run the api server",
		Run: func(cmd *cobra.Command, args []string) {
			var webhookAuth auth.Plugin
			if cfg.AuthWebhook == "token" {
				webhookAuth = &auth.TokenAuth{Store: tokenStore}
			} else {
				webhookAuth = &auth.Anonymous{}
			}
			api.RunServer(cfg.APIPort, &auth.Anonymous{}, webhookAuth, store)
		},
	}

	rootCmd.AddCommand(cmdIRC, cmdAPI)
	rootCmd.Execute()
}
