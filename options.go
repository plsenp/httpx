package httpx

import (
	"time"

	"github.com/plsenp/httpx/openapi"
)

type ServerOption func(*Server)

func WithAddr(addr string) ServerOption {
	return func(s *Server) {
		s.addr = addr
	}
}

func WithMiddleware(md ...Middleware) ServerOption {
	return func(s *Server) {
		s.middlewares = append(s.middlewares, md...)
	}
}

func WithOpenAPISpec(spec openapi.Config) ServerOption {
	return func(s *Server) {
		s.openAPISpec = openapi.NewSpec(spec)
	}
}

func WithReadTimeout(timeout time.Duration) ServerOption {
	return func(s *Server) {
		s.readTimeout = timeout
	}
}

func WithReadHeaderTimeout(timeout time.Duration) ServerOption {
	return func(s *Server) {
		s.readHeaderTimeout = timeout
	}
}

func WithWriteTimeout(timeout time.Duration) ServerOption {
	return func(s *Server) {
		s.writeTimeout = timeout
	}
}

func WithIdleTimeout(timeout time.Duration) ServerOption {
	return func(s *Server) {
		s.idleTimeout = timeout
	}
}
