package server

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/dnaeon/go-vcr/cassette"
	govcr "github.com/dnaeon/go-vcr/recorder"
	"github.com/nyarly/dns-manager/storage"
	"github.com/nyarly/spies"
	ns1 "gopkg.in/ns1/ns1-go.v2/rest"
	"gopkg.in/ns1/ns1-go.v2/rest/model/dns"
)

var recordMode = flag.Bool("record", false, "update VCR files")

type harness struct {
	mux     *http.ServeMux
	store   *storage.Spy
	stopVCR func()
}

func testHarness(t *testing.T) harness {
	t.Helper()
	vcrMode := govcr.ModeReplaying
	if *recordMode {
		vcrMode = govcr.ModeRecording
	}
	vcr, err := govcr.NewAsMode(fmt.Sprintf("testdata/cassettes/%s", t.Name()), vcrMode, nil)
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
		mux:     server.buildRouter(),
		store:   store,
		stopVCR: func() { vcr.Stop() },
	}
}

func buildBody(t *testing.T, v interface{}) io.Reader {
	t.Helper()
	b := bytes.Buffer{}
	if err := json.NewEncoder(&b).Encode(v); err != nil {
		t.Fatal(err)
	}
	return &b
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
	if strings.Index(recorder.Body.String(), "jdl-example.com") == -1 {
		t.Errorf("Body doesn't include zone name: %q", recorder.Body.String())
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
	if strings.Index(recorder.Body.String(), "jdl-example.com") == -1 {
		t.Errorf("Body doesn't include zone name: %q", recorder.Body.String())
	}
}

func TestUpdateExistingZone(t *testing.T) {
	recorder := httptest.NewRecorder()
	harness := testHarness(t)
	harness.store.MatchMethod("RecordZone", spies.AnyArgs, true, nil)
	harness.store.MatchMethod("GetZone", spies.AnyArgs, dns.NewZone("jdl-example.com"), nil)
	defer harness.stopVCR()

	req := httptest.NewRequest("PUT", "/zone", nil)
	req.URL.RawQuery = "name=jdl-example.com"
	harness.mux.ServeHTTP(recorder, req)
	rz := recorder.Result()

	if rz.StatusCode != 200 {
		t.Errorf("Expected 200 response, but status was %s \n%s", rz.Status, recorder.Body.String())
	}
	if strings.Index(recorder.Body.String(), "jdl-example.com") == -1 {
		t.Errorf("Body doesn't include zone name: %q", recorder.Body.String())
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
	if len(recorder.Body.String()) > 0 {
		t.Errorf("Body is not empty: %q", recorder.Body.String())
	}
}

func TestCreateRecord(t *testing.T) {
	recorder := httptest.NewRecorder()
	harness := testHarness(t)
	defer harness.stopVCR()

	req := httptest.NewRequest("PUT", "/record", buildBody(t, [][]string{[]string{"1.2.3.4"}}))
	req.URL.RawQuery = "zone=jdl-example.com&domain=somewhere.jdl-example.com&type=A"

	harness.mux.ServeHTTP(recorder, req)
	rz := recorder.Result()

	if rz.StatusCode != 200 {
		t.Errorf("Expected 200 response, but status was %s \n%s", rz.Status, recorder.Body.String())
	}
	if strings.Index(recorder.Body.String(), "jdl-example.com") == -1 {
		t.Errorf("Body doesn't include zone name: %q", recorder.Body.String())
	}
}
