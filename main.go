package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "dns-manager",
		Short: "A management tool for NS1 records.",
	}

	zoneCmd = &cobra.Command{
		Use:   "zone",
		Short: "Zone commands",
	}

	recordCmd = &cobra.Command{
		Use:   "record",
		Short: "Record commands",
	}
)

func main() {
	setup()
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

//go:generate inlinefiles --package=main --vfs=Templates templates templates.go

func setup() {
	rootCmd.AddCommand(serverCmd, zoneCmd, recordCmd)
	zoneCmd.AddCommand(zoneAddCmd, zoneDeleteCmd)
	recordCmd.AddCommand(recordAddCmd, recordDeleteCmd)

	serverCmd.Flags().StringP("listen", "L", "localhost:4444", "the address to listen for client requests on")
	serverCmd.Flags().StringP("store", "s", "manager.cache", "the path to use to store local records of DNS states")

	zoneAddCmd.Flags().StringP("address", "S", "localhost:4444", "the address to talk to the server on")
	zoneDeleteCmd.Flags().StringP("address", "S", "localhost:4444", "the address to talk to the server on")

	recordAddCmd.Flags().StringP("address", "S", "localhost:4444", "the address to talk to the server on")
	recordAddCmd.Flags().StringP("zone", "z", "", "The zone to add the record under - by default we guess from the name")

	recordDeleteCmd.Flags().StringP("address", "S", "localhost:4444", "the address to talk to the server on")
	recordDeleteCmd.Flags().StringP("zone", "z", "", "The zone to add the record under - by default we guess from the name")
}

func doRequest(method, addr, path string, query map[string]string, dtoIn, dtoOut interface{}) error {
	u, err := url.Parse(addr)
	if err != nil {
		return err
	}
	if u.Scheme != "http" { // XXX https
		u, err = url.Parse("http://" + addr)
		if err != nil {
			return err
		}
	}

	u.Path = path
	vs := url.Values{}
	for k, v := range query {
		vs.Set(k, v)
	}
	u.RawQuery = vs.Encode()

	var req *http.Request

	if dtoIn != nil {
		body := &bytes.Buffer{}
		if err := json.NewEncoder(body).Encode(dtoIn); err != nil {
			return err
		}
		req, err = http.NewRequest(method, u.String(), body)
		if err != nil {
			return err
		}
	} else {
		req, err = http.NewRequest(method, u.String(), nil)
		if err != nil {
			return err
		}
	}

	rz, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if rz.StatusCode != 200 {
		body, err := ioutil.ReadAll(rz.Body)
		if err != nil {
			return err
		}

		return errors.New(string(body))
	}

	if dtoOut == nil {
		return nil
	}

	return json.NewDecoder(rz.Body).Decode(dtoOut)
}
