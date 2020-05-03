package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"mcdc/api"
	"mcdc/irc"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Println("expected subcommand: irc, api")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "irc":
		var config irc.Config
		ircCmd := flag.NewFlagSet("irc", flag.ExitOnError)
		ircCmd.StringVar(&config.Name, "name", "ircd", "irc server name")
		ircCmd.StringVar(&config.Network, "network", "My IRC Network", "irc network name")
		ircCmd.IntVar(&config.Port, "port", 6667, "listen port for irc server")
		ircCmd.IntVar(&config.PingFrequency, "ping", 30, "ping frequency to client in seconds")
		ircCmd.IntVar(&config.PongMaxLatency, "timeout", 5, "client pong response timeout in seconds")
		ircCmd.Parse(os.Args[2:])

		log.SetFlags(log.Ldate | log.Ltime)
		irc.SetLogLevel(4)

		irc.RunServer(config)

	case "api":
		apiCmd := flag.NewFlagSet("irc", flag.ExitOnError)
		apiPort := apiCmd.Int("port", 8080, "listen port for api server")
		apiCmd.Parse(os.Args[2:])

		api.RunServer(*apiPort)

	default:
		fmt.Println("expected subcommand: irc, api")
		os.Exit(1)
	}
}
