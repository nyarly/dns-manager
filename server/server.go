package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/nyarly/dns-manager/storage"
	ns1 "gopkg.in/ns1/ns1-go.v2/rest"
	"gopkg.in/ns1/ns1-go.v2/rest/model/dns"
)

// Server represents the DNSManager server itself.
type Server struct {
	address  string
	storage  storage.Storage
	key      string
	clientFn func(context.Context) ns1.Doer
}

// New constructs a Server object from
//   address:      the address to listen on
//   storage:      a persistence engine
//   httpClient:   a properly configured http.Client to talk to NS1 with
//   key:          an NS1 API Key
func New(address string, storage storage.Storage, key string, httpClientFn func(context.Context) ns1.Doer) *Server {
	return &Server{
		address:  address,
		storage:  storage,
		key:      key,
		clientFn: httpClientFn,
	}
}

func (s Server) ns1Client(ctx context.Context) *ns1.Client {
	return ns1.NewClient(s.clientFn(ctx), ns1.SetAPIKey(s.key))
}

// Start commands a Server to start serving HTTP
func (s *Server) Start(ctx context.Context) error {
	server := http.Server{
		Addr:        s.address,
		Handler:     s.buildRouter(),
		BaseContext: s.baseContext(ctx),
		ConnContext: s.connContext(),
	}
	return server.ListenAndServe()
}

func (s *Server) baseContext(ctx context.Context) func(net.Listener) context.Context {
	return func(_ net.Listener) context.Context {
		return ctx
	}
}

func (s *Server) connContext() func(ctx context.Context, conn net.Conn) context.Context {
	return func(ctx context.Context, conn net.Conn) context.Context {
		return ctx
	}
}

func (s *Server) buildRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/zone", func(rw http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case "GET":
			s.getZone(rw, req)
		case "PUT":
			s.updateZone(rw, req)
		case "DELETE":
			s.deleteZone(rw, req)
		default:
			methodNotAllowed(rw)
		}
	})
	mux.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case "GET":
			s.indexPage(rw, req)
		default:
			methodNotAllowed(rw)
		}
	})
	return mux
}

func (s *Server) indexPage(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(rw, "/zone{?name} Zone manipulation")
}

func methodNotAllowed(rw http.ResponseWriter) {
	rw.WriteHeader(405)
}

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

	present, err := s.storage.CheckZone(name)
	if err != nil {
		rw.WriteHeader(503)
		fmt.Fprintf(rw, "problem checking for zone: %v", err)
	}

	ctx := req.Context()

	var rz *http.Response
	var zone *dns.Zone
	if present {
		zone, rz, err = s.updateZoneAPI(ctx, name)
	} else {
		zone, rz, err = s.createZoneAPI(ctx, name)
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
