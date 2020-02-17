package server

import (
	"context"
	"fmt"
	"net/http"

	"gopkg.in/ns1/ns1-go.v2/rest/model/dns"
)

func getZoneName(rw http.ResponseWriter, req *http.Request) string {
	query := req.URL.Query()  // TODO handle errors
	zone := query.Get("name") // TODO handle errors here

	if zone == "" {
		rw.WriteHeader(400)
		fmt.Fprintf(rw, "name parameter is required")
	}

	return zone
}

func (s *Server) getZone(rw http.ResponseWriter, req *http.Request) {
	name := getZoneName(rw, req)
	if name == "" {
		return
	}

	ctx := req.Context()
	zone, rz, err := s.getZoneAPI(ctx, name)
	if _, err := s.storage.RecordZone(*zone); err != nil {
		rw.WriteHeader(503)
		fmt.Fprintf(rw, "problem recording zone: %v", err)
	}
	proxyAPIResponse(rw, rz, zone, err)
}

func (s *Server) updateZone(rw http.ResponseWriter, req *http.Request) {
	name := getZoneName(rw, req)
	if name == "" {
		return
	}

	existing, err := s.storage.GetZone(name)
	if err != nil {
		rw.WriteHeader(503)
		fmt.Fprintf(rw, "problem checking for zone: %v", err)
	}

	ctx := req.Context()

	var rz *http.Response
	var zone *dns.Zone
	if existing == nil {
		zone, rz, err = s.createZoneAPI(ctx, name)
	} else {
		zone, rz, err = s.updateZoneAPI(ctx, name)
	}
	if _, err := s.storage.RecordZone(*zone); err != nil {
		rw.WriteHeader(503)
		fmt.Fprintf(rw, "problem recording zone: %v", err)
	}

	proxyAPIResponse(rw, rz, zone, err)
}

func (s *Server) deleteZone(rw http.ResponseWriter, req *http.Request) {
	name := getZoneName(rw, req)
	if name == "" {
		return
	}

	ctx := req.Context()
	rz, err := s.deleteZoneAPI(ctx, name)
	proxyAPIResponse(rw, rz, nil, err)
}

func (s *Server) getZoneAPI(ctx context.Context, name string) (*dns.Zone, *http.Response, error) {
	zone, rz, err := s.ns1Client(ctx).Zones.Get(name)
	return zone, rz, err
}

func (s *Server) createZoneAPI(ctx context.Context, zone string) (*dns.Zone, *http.Response, error) {
	z := dns.NewZone(zone)
	rz, err := s.ns1Client(ctx).Zones.Create(z)

	return z, rz, err
}

func (s *Server) updateZoneAPI(ctx context.Context, zone string) (*dns.Zone, *http.Response, error) {
	z := dns.NewZone(zone)
	rz, err := s.ns1Client(ctx).Zones.Update(z)
	return z, rz, err
}

func (s *Server) deleteZoneAPI(ctx context.Context, zone string) (*http.Response, error) {
	return s.ns1Client(ctx).Zones.Delete(zone)
}
