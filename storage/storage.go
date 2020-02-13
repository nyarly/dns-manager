package storage

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"strings"

	"gopkg.in/ns1/ns1-go.v2/rest/model/dns"
)

// Storage is an interface for managing local persistence for DNSManager
type Storage interface {
	// GetZone retreives a zone from the store by name
	GetZone(string) (*dns.Zone, error)
	// RecordZone persists a zone. Returns true if the zone was already persisted
	RecordZone(dns.Zone) (bool, error)
	// DeleteZone removes a zone from storage
	DeleteZone(string) (bool, error)
	// GetRecord retreives a record from the store by name
	GetRecord(string, string, string) (*dns.Record, error)
	// RecordRecord persists a record. Returns true if the record was already persisted
	//   note that it's VerbNoun, not just a repetition in the name
	RecordRecord(dns.Record) (bool, error)
	// DeleteRecord removes a record from storage by name
	DeleteRecord(string, string, string) (bool, error)
}

type textFile struct {
	path string
}

// Stored is the format for the textFile persistence layer
type Stored struct {
	Zones   []dns.Zone
	Records []dns.Record
}

// New constructs an on-disk Storage at the given path
func New(path string) Storage {
	return &textFile{path: path}
}

func (tf textFile) load() (*Stored, error) {
	f, err := os.Open(tf.path)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Stored{}, nil
		}
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

func (tf textFile) GetZone(name string) (*dns.Zone, error) {
	stored, err := tf.load()
	if err != nil {
		return nil, err
	}

	zones := stored.Zones

	for _, z := range zones {
		if name == z.Zone {
			return &z, nil
		}
	}

	return nil, nil
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

func (tf textFile) GetRecord(zone, domain, kind string) (*dns.Record, error) {
	stored, err := tf.load()
	if err != nil {
		return nil, err
	}

	records := stored.Records

	for _, r := range records {
		if r.Zone == zone && r.Domain == domain && r.Type == kind {
			return &r, nil
		}
	}

	return nil, nil
}

func (tf textFile) RecordRecord(record dns.Record) (bool, error) {
	stored, err := tf.load()
	if err != nil {
		return false, err
	}

	records := stored.Records

	found := false
	for i, r := range records {
		if r.Zone == record.Zone && r.Domain == record.Domain && r.Type == record.Type {
			records[i] = record
			found = true
			break
		}
	}
	if !found {
		records = append(records, record)
	}
	stored.Records = records
	err = tf.store(stored)
	return found, err
}

func (tf textFile) DeleteRecord(zone, domain, kind string) (bool, error) {
	stored, err := tf.load()
	if err != nil {
		return false, err
	}

	records := stored.Records

	found := false
	for i, r := range records {
		if r.Zone == zone && r.Domain == domain && r.Type == kind {
			chopped := len(records) - 1
			records[i] = records[chopped]
			records = records[:chopped]
			found = true
			break
		}
	}
	if !found {
		return false, nil
	}
	stored.Records = records
	err = tf.store(stored)
	return found, err
}
