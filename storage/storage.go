package storage

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// Storage is an interface for managing local persistence for DNSManager
type Storage interface {
	// RecordZone persists a zone name. Returns true if the zone was already persisted
	RecordZone(string) (bool, error)
}

type textFile struct {
	path string
}

// New constructs an on-disk Storage at the given path
func New(path string) Storage {
	return &textFile{path: path}
}

func (tf textFile) listZones() ([]string, error) {
	f, err := os.Open(tf.path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	contents, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(contents), "\n"), nil
}

func (tf textFile) RecordZone(zone string) (bool, error) {
	zones, err := tf.listZones()
	if err != nil {
		return false, err
	}

	for _, z := range zones {
		if zone == z {
      return true, nil
		}
	}

	f, err := os.OpenFile(tf.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return false, err
	}
	defer f.Close()
	_, err = fmt.Fprintln(f, zone)
	return false, err
}
