package main

import "log"

func main() {
  setup()
  if err := rootCmd.Execute(); err != nil {
    log.Fatal(err)
  }
}

func setup() {
  rootCmd.AddCommand(serverCmd)
  serverCmd.Flags().StringP("listen", "L", "localhost:4444", "the address to listen for client requests on")
  serverCmd.Flags().StringP("store", "s", "manager.cache", "the path to use to store local records of DNS states")
}
