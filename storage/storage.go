package storage

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"

	"gopkg.in/ns1/ns1-go.v2/rest/model/dns"
)

// Storage is an interface for managing local persistence for DNSManager
type Storage interface {
	// Check Zone reports whether a zone name exists in the store already
	CheckZone(string) (bool, error)
	// RecordZone persists a zone. Returns true if the zone was already persisted
	RecordZone(dns.Zone) (bool, error)
	// RecordZone persists a zone name. Returns true if the zone was already persisted
	DeleteZone(string) (bool, error)
}

type textFile struct {
	path string
}

// Stored is the format for the textFile persistence layer
type Stored struct {
	Zones []dns.Zone
}

// New constructs an on-disk Storage at the given path
func New(path string) Storage {
	return &textFile{path: path}
}

func (tf textFile) load() (*Stored, error) {
	f, err := os.Open(tf.path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	stored := Stored{}

	if err := dec.Decode(&stored); err != nil {
		return nil, err
	}
	return &stored, nil
}

func (tf textFile) store(stored *Stored) error {
	f, err := os.Create(tf.path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	return enc.Encode(stored)
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

func (tf textFile) CheckZone(name string) (bool, error) {
	stored, err := tf.load()
	if err != nil {
		return false, err
	}

	zones := stored.Zones

	for _, z := range zones {
		if name == z.Zone {
			return true, nil
		}
	}

	return false, nil
}

func (tf textFile) RecordZone(zone dns.Zone) (bool, error) {
	stored, err := tf.load()
	if err != nil {
		return false, err
	}

	zones := stored.Zones

	found := false
	for i, z := range zones {
		if zone.Zone == z.Zone {
			zones[i] = zone
			found = true
			break
		}
	}
	if !found {
		zones = append(zones, zone)
	}

	stored.Zones = zones
	err = tf.store(stored)
	return found, err
}

func (tf textFile) DeleteZone(name string) (bool, error) {
	stored, err := tf.load()
	if err != nil {
		return false, err
	}

	zones := stored.Zones

	found := false
	for i, z := range zones {
		if name == z.Zone {
			chopped := len(zones) - 1
			zones[i] = zones[chopped]
			zones = zones[:chopped]
			found = true
			break
		}
	}
	if !found {
		return false, nil
	}

	stored.Zones = zones
	err = tf.store(stored)
	return found, err
}
