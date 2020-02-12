package storage

import "github.com/nyarly/spies"

// Spy is a Storage spy for testing
type Spy struct {
  *spies.Spy
}

// NewSpy constructs a Storage spy for testing
func NewSpy() *Spy {
  return &Spy{ spies.NewSpy() }
}

// RecordZone implements Storage on Spy
func (spy *Spy) RecordZone(zone string) (bool, error) {
  res := spy.Called(zone)
  return res.Bool(0), res.Error(1)
}
