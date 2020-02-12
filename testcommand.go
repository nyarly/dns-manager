package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
	api "gopkg.in/ns1/ns1-go.v2/rest"
)

// 100 empty Zones.List() calls took 12s
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Quick test of NS1 integration",
	RunE:  runTest,
}

func runTest(cmd *cobra.Command, args []string) error {
	k := os.Getenv("NS1_APIKEY")
	if k == "" {
		return fmt.Errorf("NS1_APIKEY environment variable is not set, giving up")
	}

	httpClient := &http.Client{Timeout: time.Second * 10}
	client := api.NewClient(httpClient, api.SetAPIKey(k))

	zones, _, err := client.Zones.List()
	if err != nil {
		return err
	}

	for _, z := range zones {
		fmt.Println(z.Zone)
	}

	now := time.Now()
	log.Printf("Starting benchmark: %v", now)
	for i := 0; i < 100; i++ {
		_, _, err := client.Zones.List()
		if err != nil {
			return err
		}
	}
	then := time.Now()
	log.Printf("Done benchmark: %v %v", then, then.Sub(now))

	return nil
}
