package storage

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/ns1/ns1-go.v2/rest/model/dns"
)

func setup(t *testing.T) (Storage, func()) {
	t.Helper()
	dir, err := ioutil.TempDir("", t.Name())
	if err != nil {
		t.Fatalf("Creating tempdir: %v", err)
	}

	storePath := filepath.Join(dir, "storage")
	store := New(storePath)
	return store, func() {
		defer os.RemoveAll(dir)
	}
}

func TestGetZone(t *testing.T) {
	store, cleanup := setup(t)
	defer cleanup()

	zone, err := store.GetZone("example.com")
	if err != nil {
		t.Fatalf("err from GetZone: %v", err)
	}

	if zone != nil {
		t.Fatalf("GetZone returned a zone from empty storage: %v", zone)
	}

	store.RecordZone(*dns.NewZone("example.com"))

	zone, err = store.GetZone("example.com")
	if err != nil {
		t.Fatalf("err from GetZone: %v", err)
	}
	if zone == nil {
		t.Fatalf("GetZone returned nil after storing record")
	}
}

func TestRecordZone(t *testing.T) {
	store, cleanup := setup(t)
	defer cleanup()

  present, err := store.RecordZone(*dns.NewZone("example.com"))
	if err != nil {
		t.Fatalf("err from RecordZone: %v", err)
	}
	if present {
		t.Fatalf("RecordZone returned 'present' after storing record in empty store")
	}

  present, err = store.RecordZone(*dns.NewZone("example.com"))
	if err != nil {
		t.Fatalf("err from RecordZone: %v", err)
	}
	if !present {
		t.Fatalf("RecordZone returned 'not present' after re-storing record")
	}
}

func TestDeleteZone(t *testing.T) {
	store, cleanup := setup(t)
	defer cleanup()

  _, err := store.RecordZone(*dns.NewZone("example.com"))
	if err != nil {
		t.Fatalf("err from RecordZone: %v", err)
	}

  present, err := store.DeleteZone("example.com")
	if err != nil {
		t.Fatalf("err from DeleteZone: %v", err)
	}
	if !present {
		t.Fatalf("DeleteZone returned 'not present' after deleting record")
	}
}

func TestGetRecord(t *testing.T) {
	store, cleanup := setup(t)
	defer cleanup()

	zone, err := store.GetRecord("example.com", "www.example.com", "a")
	if err != nil {
		t.Fatalf("err from GetRecord: %v", err)
	}

	if zone != nil {
		t.Fatalf("GetRecord returned a zone from empty storage: %v", zone)
	}

	store.RecordRecord(*dns.NewRecord("example.com", "www.example.com", "a"))

	zone, err = store.GetRecord("example.com" ,"www.example.com", "a")
	if err != nil {
		t.Fatalf("err from GetRecord: %v", err)
	}
	if zone == nil {
		t.Fatalf("GetRecord returned nil after storing record")
	}
}

func TestRecordRecord(t *testing.T) {
	store, cleanup := setup(t)
	defer cleanup()

  present, err := store.RecordRecord(*dns.NewRecord("example.com", "www.example.com", "a"))
	if err != nil {
		t.Fatalf("err from RecordRecord: %v", err)
	}
	if present {
		t.Fatalf("RecordRecord returned 'present' after storing record in empty store")
	}

  present, err = store.RecordRecord(*dns.NewRecord("example.com", "www.example.com", "a"))
	if err != nil {
		t.Fatalf("err from RecordRecord: %v", err)
	}
	if !present {
		t.Fatalf("RecordRecord returned 'not present' after re-storing record")
	}
}

func TestDeleteRecord(t *testing.T) {
	store, cleanup := setup(t)
	defer cleanup()

  _, err := store.RecordRecord(*dns.NewRecord("example.com", "www.example.com", "a"))
	if err != nil {
		t.Fatalf("err from RecordRecord: %v", err)
	}

  present, err := store.DeleteRecord("example.com", "www.example.com", "a")
	if err != nil {
		t.Fatalf("err from DeleteRecord: %v", err)
	}
	if !present {
		t.Fatalf("DeleteRecord returned 'not present' after deleting record")
	}
}
