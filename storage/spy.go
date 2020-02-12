package storage

import (
  "github.com/nyarly/spies"
	"gopkg.in/ns1/ns1-go.v2/rest/model/dns"
)

// Spy is a Storage spy for testing
type Spy struct {
	*spies.Spy
}

// NewSpy constructs a Storage spy for testing
func NewSpy() *Spy {
	return &Spy{spies.NewSpy()}
}

// CheckZone implements Storage on Spy
func (spy *Spy) CheckZone(name string) (bool, error) {
	res := spy.Called(name)
	return res.Bool(0), res.Error(1)
}

// RecordZone implements Storage on Spy
func (spy *Spy) RecordZone(zone dns.Zone) (bool, error) {
	res := spy.Called(zone)
	return res.Bool(0), res.Error(1)
}

// DeleteZone implements Storage on Spy
func (spy *Spy) DeleteZone(name string) (bool, error) {
	res := spy.Called(name)
	return res.Bool(0), res.Error(1)
}
