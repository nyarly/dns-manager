package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func getRecordParams(rw http.ResponseWriter, req *http.Request) (string, string, string) {
	query := req.URL.Query()      // TODO handle errors
	zone := query.Get("zone")     // TODO handle errors here
	domain := query.Get("domain") // TODO handle errors here
	kind := query.Get("type")     // TODO handle errors here

	if zone == "" || domain == "" || kind == ""{
		rw.WriteHeader(400)
		fmt.Fprintf(rw, "parameters for zone, domain and kind are all required")
    return "", "", ""
	}

	return zone, domain, kind
}

func (s *Server) getRecord(rw http.ResponseWriter, req *http.Request) {
	zone, domain, kind := getRecordParams(rw, req)
	if zone == "" {
		return
	}

	ctx := req.Context()
	zone, rz, err := s.getRecordAPI(ctx, zone, domain, kind)
	if _, err := s.storage.RecordRecord(*zone); err != nil {
		rw.WriteHeader(503)
		fmt.Fprintf(rw, "problem recording zone: %v", err)
	}
	proxyAPIResponse(rw, rz, zone, err)
}

func (s *Server) updateRecord(rw http.ResponseWriter, req *http.Request) {
	zone, domain, kind := getRecordParams(rw, req)
	if zone == "" {
		return
	}

	existing, err := s.storage.GetRecord(zone, domain, kind)
	if err != nil {
		rw.WriteHeader(503)
		fmt.Fprintf(rw, "problem checking for zone: %v", err)
	}

	ctx := req.Context()

	var rz *http.Response
	var zone *dns.Record
	if existing == nil {
		zone, rz, err = s.createRecordAPI(ctx, zone, domain, kind)
	} else {
		zone, rz, err = s.updateRecordAPI(ctx, zone, domain, kind)
	}
	if _, err := s.storage.RecordRecord(*zone); err != nil {
		rw.WriteHeader(503)
		fmt.Fprintf(rw, "problem recording zone: %v", err)
	}

	proxyAPIResponse(rw, rz, zone, err)
}

func (s *Server) deleteRecord(rw http.ResponseWriter, req *http.Request) {
	zone, domain, kind := getRecordParams(rw, req)
	if zone == "" {
		return
	}

	ctx := req.Context()
	rz, err := s.deleteRecordAPI(ctx, zone, domain, kind)
	proxyAPIResponse(rw, rz, nil, err)
}

func proxyAPIResponse(rw http.ResponseWriter, rz *http.Response, body interface{}, err error) {
	if rz == nil {
		rw.WriteHeader(503)
	} else {
		rw.WriteHeader(rz.StatusCode)
	}

	if err != nil {
		fmt.Fprintf(rw, "problem updating NS1: %v", err)
		return
	}

	if rz.StatusCode != 200 || body == nil {
		return
	}

	if err := json.NewEncoder(rw).Encode(body); err != nil {
		panic(err) // XXX but we already wrote a status...
	}
}

func (s *Server) getRecordAPI(ctx context.Context, zone, domain, kind string) (*dns.Record, *http.Response, error) {
	zone, rz, err := s.ns1Client(ctx).Records.Get(zone, domain, kind)
	return zone, rz, err
}

func (s *Server) createRecordAPI(ctx context.Context, zone string) (*dns.Record, *http.Response, error) {
	z := dns.NewRecord(zone)
	rz, err := s.ns1Client(ctx).Records.Create(z)

	return z, rz, err
}

func (s *Server) updateRecordAPI(ctx context.Context, zone string) (*dns.Record, *http.Response, error) {
	z := dns.NewRecord(zone)
	rz, err := s.ns1Client(ctx).Records.Update(z)
	return z, rz, err
}

func (s *Server) deleteRecordAPI(ctx context.Context, zone string) (*http.Response, error) {
	return s.ns1Client(ctx).Records.Delete(zone)
}
