package server

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/dnaeon/go-vcr/cassette"
	govcr "github.com/dnaeon/go-vcr/recorder"
	"github.com/nyarly/dns-manager/storage"
	"github.com/nyarly/spies"
	ns1 "gopkg.in/ns1/ns1-go.v2/rest"
)

type harness struct {
	mux     *http.ServeMux
	store   *storage.Spy
	stopVCR func()
}

func testHarness(t *testing.T) harness {
	t.Helper()
	vcr, err := govcr.New(fmt.Sprintf("testdata/cassettes/%s", t.Name()))
	if err != nil {
		t.Fatalf("Error with VCR: %v", err)
	}
	vcr.AddFilter(func(i *cassette.Interaction) error {
		delete(i.Request.Headers, "X-Nsone-Key")
		return nil
	})
	key := os.Getenv("NS1_APIKEY")
	if key == "" {
		t.Fatal("Test needs NS1_APIKEY environment variable to run")
	}

	vcrClient := &http.Client{Transport: vcr}

	store := storage.NewSpy()
	server := New("example.com:80", store, key, func(_ context.Context) ns1.Doer {
		return vcrClient
	})

	return harness{
		mux:   server.buildRouter(),
		store: store,
    stopVCR: func() { vcr.Stop() },
	}
}

// TODO assertions about storage use
// Body contents

func TestGetZone(t *testing.T) {
	recorder := httptest.NewRecorder()
	harness := testHarness(t)
	defer harness.stopVCR()

	req := httptest.NewRequest("GET", "/zone", nil)
	req.URL.RawQuery = "name=jdl-example.com"
	harness.mux.ServeHTTP(recorder, req)
	rz := recorder.Result()

	if rz.StatusCode != 200 {
		t.Errorf("Expected 200 response, but status was %s \n%s", rz.Status, recorder.Body.String())
	}
}

func TestUpdateZone(t *testing.T) {
	recorder := httptest.NewRecorder()
	harness := testHarness(t)
	harness.store.MatchMethod("RecordZone", spies.AnyArgs, false, nil)
	defer harness.stopVCR()

	req := httptest.NewRequest("PUT", "/zone", nil)
	req.URL.RawQuery = "name=jdl-example.com"
	harness.mux.ServeHTTP(recorder, req)
	rz := recorder.Result()

	if rz.StatusCode != 200 {
		t.Errorf("Expected 200 response, but status was %s \n%s", rz.Status, recorder.Body.String())
	}
}

func TestUpdateExistingZone(t *testing.T) {
	recorder := httptest.NewRecorder()
	harness := testHarness(t)
	harness.store.MatchMethod("RecordZone", spies.AnyArgs, true, nil)
	defer harness.stopVCR()

	req := httptest.NewRequest("PUT", "/zone", nil)
	req.URL.RawQuery = "name=jdl-example.com"
	harness.mux.ServeHTTP(recorder, req)
	rz := recorder.Result()

	if rz.StatusCode != 200 {
		t.Errorf("Expected 200 response, but status was %s \n%s", rz.Status, recorder.Body.String())
	}
}

func TestDeleteZone(t *testing.T) {
	recorder := httptest.NewRecorder()
	harness := testHarness(t)
	defer harness.stopVCR()

	req := httptest.NewRequest("DELETE", "/zone", nil)
	req.URL.RawQuery = "name=jdl-example.com"
	harness.mux.ServeHTTP(recorder, req)
	rz := recorder.Result()

	if rz.StatusCode != 200 {
		t.Errorf("Expected 200 response, but status was %s \n%s", rz.Status, recorder.Body.String())
	}
}
