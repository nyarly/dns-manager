package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var recordDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete a record",
	RunE:  recordDeleteFn,
	Args:  cobra.ExactArgs(2),
}

func recordDeleteFn(cmd *cobra.Command, args []string) error {
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

	if zone == "" {
		idx := strings.Index(name, ".")
		if idx == -1 {
			return errors.New("no dots in name")
		}
		zone = name[idx+1 : len(name)]
		fmt.Printf("Using %q as zone\n", zone)
	}

	query := map[string]string{
		"zone":   zone, // underflow should be guarded by Cobra
		"domain": name,
		"type":   kind,
	}

	if err := doRequest("DELETE", addr, "/record", query, nil, nil); err != nil {
		fmt.Println(err)
		return nil
	}

	fmt.Println("Deleted")
	return nil
}
