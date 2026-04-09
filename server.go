package httpx

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/plsenp/httpx/openapi"
)

type (
	BindFunc   func(r *http.Request, v any) error
	RenderFunc func(w http.ResponseWriter, r *http.Request, status int, v any) error
)

type Server struct {
	hs          *http.Server
	mux         *http.ServeMux
	hostMode    bool
	errHandler  func(w http.ResponseWriter, r *http.Request, err error)
	middlewares []Middleware
	addr        string
	tlsConfig   *tls.Config

	// Timeouts
	readTimeout       time.Duration
	readHeaderTimeout time.Duration
	writeTimeout      time.Duration
	idleTimeout       time.Duration

	// binders
	queryBind  BindFunc
	pathBind   BindFunc
	headerBind BindFunc
	bodyBind   BindFunc

	// validator
	validate Validator

	// render
	renderFunc RenderFunc

	// openapi spec reference (optional)
	openAPISpec *openapi.Spec

	// openapi cache
	openAPICache func() []byte
}

func NewServer(opts ...ServerOption) *Server {
	srv := &Server{
		hs:                &http.Server{},
		addr:              "0.0.0.0:8080",
		mux:               http.NewServeMux(),
		readTimeout:       60 * time.Second,
		readHeaderTimeout: 10 * time.Second,
		writeTimeout:      60 * time.Second,
		idleTimeout:       120 * time.Second,

		queryBind:  defaultQueryBind,
		pathBind:   defaultPathBind,
		headerBind: defaultHeaderBind,
		bodyBind:   defaultBodyBind,
		errHandler: defaultErrorHandler,
		renderFunc: defaultRenderFunc,
		validate:   defaultValidator,
	}
	for _, opt := range opts {
		opt(srv)
	}
	srv.hs.Handler = srv.mux
	srv.hs.ReadTimeout = srv.readTimeout
	srv.hs.ReadHeaderTimeout = srv.readHeaderTimeout
	srv.hs.WriteTimeout = srv.writeTimeout
	srv.hs.IdleTimeout = srv.idleTimeout

	return srv
}

func (s *Server) Start() error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	if s.tlsConfig != nil {
		lis = tls.NewListener(lis, s.tlsConfig)
	}
	defer lis.Close()
	if s.openAPISpec != nil {
		s.registerOpenAPISpec()
	}

	return s.hs.Serve(lis)
}

// registerOpenAPISpec registers the OpenAPI spec with the server.
func (s *Server) registerOpenAPISpec() {
	// Initialize the cached OpenAPI JSON generator
	s.openAPICache = sync.OnceValue(func() []byte {
		data, err := s.openAPISpec.MarshalJSON()
		if err != nil {
			return []byte(`{"error": "failed to generate OpenAPI spec"}`)
		}
		return data
	})

	// Register OpenAPI JSON endpoint
	s.mux.HandleFunc("/docs/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(s.openAPICache())
	})

	// Register Swagger UI endpoint
	s.mux.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data, err := openapi.SwaggerUIHandler()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(data)
	})
}

func (s *Server) Stop(ctx context.Context) error {
	return s.hs.Shutdown(ctx)
}

func (s *Server) Route() *Router {
	r := &Router{
		server: s,
		md:     s.middlewares,
	}
	if !s.hostMode {
		r.prefix = "/"
	}
	return r
}
