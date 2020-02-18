package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var zoneDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete a zone",
	RunE:  zoneDeleteFn,
	Args:  cobra.ExactArgs(1),
}

func zoneDeleteFn(cmd *cobra.Command, args []string) error {
	addr, err := cmd.Flags().GetString("address")
	if err != nil {
		return err
	}

	query := map[string]string{
		"name": args[0], // underflow should be guarded by Cobra
	}

	if err := doRequest("DELETE", addr, "/zone", query, nil, nil); err != nil {
		fmt.Println(err)
		return nil
	}

	fmt.Println("Deleted.")
	return nil
}
