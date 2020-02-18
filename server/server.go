package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/nyarly/dns-manager/storage"
	ns1 "gopkg.in/ns1/ns1-go.v2/rest"
)

// Server represents the DNSManager server itself.
type Server struct {
	address  string
	storage  storage.Storage
	key      string
	clientFn func(context.Context) ns1.Doer
}

type contextInjectingClient struct {
	http ns1.Doer
	ctx  context.Context
}

// LiveClient is a suitable implementation for New's httpClientFn
func LiveClient(ctx context.Context) ns1.Doer {
	return contextInjectingClient{
		http: &http.Client{},
		ctx:  ctx,
	}
}

func (c contextInjectingClient) Do(rq *http.Request) (*http.Response, error) {
	return c.http.Do(rq.WithContext(c.ctx))
}

// New constructs a Server object from
//   address:      the address to listen on
//   storage:      a persistence engine
//   key:          an NS1 API Key
//   httpClientFn: a factory function returning a properly configured http.Client to talk to NS1 with
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
	mux.HandleFunc("/record", func(rw http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case "GET":
			s.getRecord(rw, req)
		case "PUT":
			s.updateRecord(rw, req)
		case "DELETE":
			s.deleteRecord(rw, req)
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
	fmt.Fprintln(rw, "/zone{?name} Zone manipulation")
	fmt.Fprintln(rw, "/record{?zone,domain,type} Record manipulation")
}

func methodNotAllowed(rw http.ResponseWriter) {
	rw.WriteHeader(405)
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
