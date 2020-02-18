package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/ns1/ns1-go.v2/rest/model/dns"
)

var recordAddCmd = &cobra.Command{
	Use:   "add",
	Short: "add a record",
	RunE:  recordAddFn,
	Args:  cobra.MinimumNArgs(3),
}

func recordAddFn(cmd *cobra.Command, args []string) error {
	addr, err := cmd.Flags().GetString("address")
	if err != nil {
		return err
	}

	zone, err := cmd.Flags().GetString("zone")
	if err != nil {
		return err
	}

	name := args[0]
	kind := args[1]
  answer := args[2:len(args)]

	if zone == "" {
		idx := strings.Index(name, ".")
		if idx == -1 {
			return errors.New("no dots in name")
		}
    zone = name[idx+1:len(name)]
		fmt.Printf("Using %q as zone\n", zone)
	}

	record := &dns.Record{}
	query := map[string]string{
		"zone":   zone, // underflow should be guarded by Cobra
		"domain": name,
		"type":   kind,
	}

  if err := doRequest("PUT", addr, "/record", query, [][]string{answer}, record); err != nil {
		fmt.Println(err)
		return nil
	}

	fmt.Println("Added")
	return nil
}
