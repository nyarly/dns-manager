package main

import (
	"fmt"
	"os"

	"github.com/nyarly/inlinefiles/templatestore"
	"github.com/spf13/cobra"
	"gopkg.in/ns1/ns1-go.v2/rest/model/dns"
)

var zoneAddCmd = &cobra.Command{
	Use:   "add",
	Short: "add a zone",
	RunE:  zoneAddFn,
	Args:  cobra.ExactArgs(1),
}

func zoneAddFn(cmd *cobra.Command, args []string) error {
	tmpl, err := templatestore.LoadText(Templates, "zone-add", "zone-add.tmpl")
	if err != nil {
		panic(err)
	}

	addr, err := cmd.Flags().GetString("address")
	if err != nil {
		return err
	}

	zone := &dns.Zone{}
	query := map[string]string{
		"name": args[0], // underflow should be guarded by Cobra
	}

	if err := doRequest("PUT", addr, "/zone", query, nil, zone); err != nil {
		fmt.Println(err)
		return nil
	}

	return tmpl.Execute(os.Stdout, zone)
}
