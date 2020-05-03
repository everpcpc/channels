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
		ircCmd := flag.NewFlagSet("irc", flag.ExitOnError)
		ircPort := ircCmd.Int("port", 6667, "listen port for irc server")
		ircCmd.Parse(os.Args[2:])

		log.SetFlags(log.Ldate | log.Ltime)
		irc.SetLogLevel(4)

		cfg := irc.Config{
			Port:           *ircPort,
			PingFrequency:  30,
			PongMaxLatency: 5,
		}
		irc.RunServer(cfg)

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
