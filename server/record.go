package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"gopkg.in/ns1/ns1-go.v2/rest/model/dns"
)

func getRecordParams(rw http.ResponseWriter, req *http.Request) (string, string, string) {
	query := req.URL.Query()      // TODO handle errors
	zone := query.Get("zone")     // TODO handle errors here
	domain := query.Get("domain") // TODO handle errors here
	kind := query.Get("type")     // TODO handle errors here

	if zone == "" || domain == "" || kind == "" {
		rw.WriteHeader(400)
		fmt.Fprintf(rw, "parameters for zone, domain and kind are all required")
		return "", "", ""
	}

	return zone, domain, kind
}

func (s *Server) getRecord(rw http.ResponseWriter, req *http.Request) {
	name, domain, kind := getRecordParams(rw, req)
	if name == "" {
		return
	}

	existing, err := s.storage.GetRecord(name, domain, kind)
	if err != nil {
		rw.WriteHeader(503)
		fmt.Fprintf(rw, "problem checking for zone: %v", err)
	}

	if existing != nil {
		if err := json.NewEncoder(rw).Encode(existing); err != nil {
			rw.WriteHeader(503)
			fmt.Fprintf(rw, "problem serializing cached zone: %v", err)
		}
		return
	}

	ctx := req.Context()
	zone, rz, err := s.getRecordAPI(ctx, name, domain, kind)
	if _, err := s.storage.RecordRecord(*zone); err != nil {
		rw.WriteHeader(503)
		fmt.Fprintf(rw, "problem recording zone: %v", err)
		return
	}
	proxyAPIResponse(rw, rz, zone, err)
}

func (s *Server) updateRecord(rw http.ResponseWriter, req *http.Request) {
	name, domain, kind := getRecordParams(rw, req)
	if name == "" {
		return
	}

	existing, err := s.storage.GetRecord(name, domain, kind)
	if err != nil {
		rw.WriteHeader(503)
		fmt.Fprintf(rw, "problem checking for zone: %v", err)
		return
	}

	answers := [][]string{}
	if err := json.NewDecoder(req.Body).Decode(&answers); err != nil {
		rw.WriteHeader(400)
		fmt.Fprintf(rw, "body of request ill formed: %v\n", err)
		return
	}
	record := buildRecord(name, domain, kind, answers)

	ctx := req.Context()

	var rz *http.Response
	if existing == nil {
		rz, err = s.createRecordAPI(ctx, record)
	} else {
		rz, err = s.updateRecordAPI(ctx, record)
	}
	if _, err := s.storage.RecordRecord(*record); err != nil {
		rw.WriteHeader(503)
		fmt.Fprintf(rw, "problem recording zone: %v", err)
	}

	proxyAPIResponse(rw, rz, record, err)
}

func buildRecord(name, domain, kind string, answers [][]string) *dns.Record {
	rr := dns.NewRecord(name, domain, kind)
	for _, a := range answers {
		ans := dns.NewAnswer(a)
		rr.AddAnswer(ans)
	}
	return rr
}

func (s *Server) deleteRecord(rw http.ResponseWriter, req *http.Request) {
	name, domain, kind := getRecordParams(rw, req)
	if name == "" {
		return
	}

	ctx := req.Context()
	rz, err := s.deleteRecordAPI(ctx, name, domain, kind)
	proxyAPIResponse(rw, rz, nil, err)
}

func (s *Server) getRecordAPI(ctx context.Context, name, domain, kind string) (*dns.Record, *http.Response, error) {
	zone, rz, err := s.ns1Client(ctx).Records.Get(name, domain, kind)
	return zone, rz, err
}

func (s *Server) createRecordAPI(ctx context.Context, record *dns.Record) (*http.Response, error) {
	rz, err := s.ns1Client(ctx).Records.Create(record)

	return rz, err
}

func (s *Server) updateRecordAPI(ctx context.Context, record *dns.Record) (*http.Response, error) {
	rz, err := s.ns1Client(ctx).Records.Update(record)
	return rz, err
}

func (s *Server) deleteRecordAPI(ctx context.Context, name, domain, kind string) (*http.Response, error) {
	return s.ns1Client(ctx).Records.Delete(name, domain, kind)
}
