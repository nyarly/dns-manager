// This file was automatically generated based on the contents of *.tmpl
// If you need to update this file, change the contents of those files
// (or add new ones) and run 'go generate'

package main

import "golang.org/x/tools/godoc/vfs/mapfs"

var Templates = mapfs.New(map[string]string{
	`zone-add.tmpl`: "Zone {{.Zone}} created!\n\nTo publish your zone, you need to configure your registrar to use the following nameservers:\n{{ range .DNSServers -}}\n- {{.}}\n{{ end }}\n",
})
