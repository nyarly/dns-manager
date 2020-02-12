package server

import (
	"context"
	"fmt"
	"io"
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
	zone := getZoneName(rw, req)
	if zone == "" {
		return
	}

	ctx := req.Context()
	proxyAPIResponse(s.getZoneAPI(ctx, rw, zone))
}

func (s *Server) updateZone(rw http.ResponseWriter, req *http.Request) {
	zone := getZoneName(rw, req)
	if zone == "" {
		return
	}

	present, err := s.recordZone(zone)
	if err != nil {
		rw.WriteHeader(503)
		fmt.Fprintf(rw, "problem recording zone: %v", err)
	}

	ctx := req.Context()
	if present {
		proxyAPIResponse(s.updateZoneAPI(ctx, rw, zone))
	} else {
		proxyAPIResponse(s.createZoneAPI(ctx, rw, zone))
	}
}

func (s *Server) deleteZone(rw http.ResponseWriter, req *http.Request) {
	zone := getZoneName(rw, req)
	if zone == "" {
		return
	}

	ctx := req.Context()
  proxyAPIResponse(s.deleteZoneAPI(ctx, rw, zone))
}

func proxyAPIResponse(rw http.ResponseWriter, rz *http.Response, err error) {
	if rz == nil {
		rw.WriteHeader(503)
	} else {
		rw.WriteHeader(rz.StatusCode)
	}
	if err != nil {
		fmt.Fprintf(rw, "problem updating NS1: %v", err)
	}

	io.Copy(rw, rz.Body)
}

func (s *Server) recordZone(zone string) (bool, error) {
	return s.storage.RecordZone(zone)
}

func (s *Server) getZoneAPI(ctx context.Context, rw http.ResponseWriter, zone string) (http.ResponseWriter, *http.Response, error) {
	_, rz, err := s.ns1Client(ctx).Zones.Get(zone)
	return rw, rz, err
}

func (s *Server) createZoneAPI(ctx context.Context, rw http.ResponseWriter, zone string) (http.ResponseWriter, *http.Response, error) {
	z := dns.NewZone(zone)
	rz, err := s.ns1Client(ctx).Zones.Create(z)
	return rw, rz, err
}

func (s *Server) updateZoneAPI(ctx context.Context, rw http.ResponseWriter, zone string) (http.ResponseWriter, *http.Response, error) {
	z := dns.NewZone(zone)
	rz, err := s.ns1Client(ctx).Zones.Update(z)
	return rw, rz, err
}

func (s *Server) deleteZoneAPI(ctx context.Context, rw http.ResponseWriter, zone string) (http.ResponseWriter, *http.Response, error) {
	rz, err := s.ns1Client(ctx).Zones.Delete(zone)
	return rw, rz, err
}
