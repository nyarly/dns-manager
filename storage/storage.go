package storage

import (
	"fmt"
	"os"
)

// Storage is an interface for managing local persistence for DNSManager
type Storage interface {
	// RecordZone persists a zone name.
	RecordZone(string) error
}

type textFile struct {
	path string
}

// New constructs an on-disk Storage at the given path
func New(path string) Storage {
	return &textFile{path: path}
}

func (tf textFile) RecordZone(zone string) error {
	f, err := os.OpenFile(tf.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
  _, err = fmt.Fprintln(f, zone)
  return err
}
