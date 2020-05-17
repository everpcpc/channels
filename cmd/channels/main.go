package main

import (
	"github.com/getsentry/sentry-go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"channels/auth"
	"channels/irc"
	"channels/slack"
	"channels/storage"
	"channels/web"
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

			if cfg.SentryDSN != "" {
				err = sentry.Init(sentry.ClientOptions{
					Dsn:              cfg.SentryDSN,
					AttachStacktrace: true,
				})
				if err != nil {
					logrus.Fatal(err)
				}
			}

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
	rootCmd.PersistentFlags().StringVarP(&flagConfig, "config", "c", "config.json", "config file to load")
	rootCmd.PersistentFlags().StringVarP(&flagLoglevel, "log", "l", "info", "loglevel: debug, info, warn, error")

	var cmdIRC = &cobra.Command{
		Use:   "irc",
		Short: "Run the irc server",
		Run: func(cmd *cobra.Command, args []string) {
			if cfg.IRC.Auth == "ldap" {
				irc.RunServer(cfg.IRC, cfg.LDAP, store)
			} else {
				irc.RunServer(cfg.IRC, &auth.Anonymous{}, store)
			}
		},
	}

	var cmdSlack = &cobra.Command{
		Use:   "slack",
		Short: "Run the slack forwarder",
		Run: func(cmd *cobra.Command, args []string) {
			client, err := slack.NewClient(cfg.Slack, store)
			if err != nil {
				logrus.Fatal(err)
			}
			client.Run()
		},
	}

	var cmdWeb = &cobra.Command{
		Use:   "web",
		Short: "Run the web server",
		Run: func(cmd *cobra.Command, args []string) {
			server := web.NewServer(cfg.Web, store, tokenStore)
			server.WithWebhook(cfg.Web.WebhookAuth)
			if cfg.Slack != nil {
				api, err := slack.NewClient(cfg.Slack, store)
				if err != nil {
					logrus.Fatal(err)
				}
				server.WithSlack(api)
			}
			server.Run(cfg.Web.Listen)
		},
	}

	var cmdToken = getTokenCommand(&tokenStore)

	rootCmd.AddCommand(cmdIRC, cmdSlack, cmdWeb, cmdToken)
	rootCmd.Execute()
}
