package main

import (
  "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
  Use: "dns-manager",
  Short: "A management tool for NS1 records.",
}


