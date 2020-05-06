package main

import (
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"channels/storage"
)

func getTokenCommand(store *storage.TokenBackend) *cobra.Command {
	var err error

	var cmd = &cobra.Command{
		Use:   "token",
		Short: "Manage authenticate tokens",
	}

	var cmdList = &cobra.Command{
		Use:   "list",
		Short: "Show a list of current tokens",
		Run: func(cmd *cobra.Command, args []string) {
			tokens, err := (*store).ListTokens()
			if err != nil {
				logrus.Fatalf("list tokens failed: %v", err)
			}
			fmt.Println("tokens:")
			for token, data := range tokens {
				fmt.Printf("%s: %v\n", token, data)
			}
		},
	}
	var cmdAdd = &cobra.Command{
		Use:     "add [user] [scope] [note]",
		Short:   "Generate an new authenticate token",
		Example: "add bot '#testbot,@' 'token for test bot'",
		Args:    cobra.MinimumNArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			b := make([]byte, 16)
			rand.Read(b)
			token := fmt.Sprintf("%x", b)
			data := &storage.TokenData{
				User:      args[0],
				Scope:     strings.Split(args[1], ","),
				Note:      args[2],
				CreatedAt: time.Now().UnixNano(),
			}
			fmt.Printf("adding token %s: %v\n", token, data)
			err = (*store).AddToken(token, data)
			if err != nil {
				logrus.Fatalf("add token failed: %v", err)
			}
		},
	}
	var cmdDel = &cobra.Command{
		Use:   "del [tokens...]",
		Short: "Delete specific authenticate token",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err = (*store).DeleteTokens(args...)
			if err != nil {
				logrus.Fatalf("delete tokens failed: %v", err)
			}
		},
	}
	cmd.AddCommand(cmdList, cmdAdd, cmdDel)

	return cmd
}
