package main

import (
	"context"
	"errors"
	"os"

	"github.com/nyarly/dns-manager/server"
	"github.com/nyarly/dns-manager/storage"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "run the dns-manager HTTP server",
  Long: "Starts an HTTP server to handle requests to manipulate the NS1 DNS service.\n" +
    "  Note: you must set an NS1_APIKEY environment with a key obtained from https://my.nsone.net/#/account/settings",
	RunE:  serverFn,
}

func serverFn(cmd *cobra.Command, args []string) error {
	listen, err := cmd.Flags().GetString("listen")
	if err != nil {
		return err
	}
	storePath, err := cmd.Flags().GetString("store")
	if err != nil {
		return err
	}
	storage := storage.New(storePath)
	key, present := os.LookupEnv("NS1_APIKEY")
	if !present {
		return errors.New("NS1_APIKEY environment variable is required to be set")
	}

	server.New(
		listen,
		storage,
		key,
		server.LiveClient,
	).Start(context.Background())
	return nil
}
