package main

import "log"

func main() {
  setup()
  if err := rootCmd.Execute(); err != nil {
    log.Fatal(err)
  }
}

func setup() {
  rootCmd.AddCommand(testCmd)

}
