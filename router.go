package httpx

import (
	"fmt"
	"log/slog"
	"net/http"
	"path"
	"strings"
)

// HandlerFunc is the handler function signature
type HandlerFunc func(ctx *Ctx) error

type Router struct {
	prefix string
	md     []Middleware
	server *Server
}

func (r *Router) Use(middleware ...Middleware) {
	r.md = append(r.md, middleware...)
}

func (r *Router) Group(prefix string, middleware ...Middleware) *Router {
	return &Router{
		prefix: r.calculateAbsolutePath(prefix),
		md:     append(r.md, middleware...),
		server: r.server,
	}
}

func (r *Router) Mount(prefix string, handler http.Handler, middleware ...Middleware) {
	fullPath := r.calculateAbsolutePath(prefix)
	// TODO: maybe don't need chain r.md
	handler = chainMiddleware(middleware...)(handler)
	handler = chainMiddleware(r.md...)(handler)
	prefixToStrip := fullPath
	if len(prefixToStrip) > 0 && prefixToStrip[len(prefixToStrip)-1] == '/' {
		prefixToStrip = prefixToStrip[:len(prefixToStrip)-1]
	}
	r.server.mux.Handle(fullPath, http.StripPrefix(prefixToStrip, handler))
}

func (r *Router) handle(method string, relativePath string, handler HandlerFunc, middleware ...Middleware) {
	fullPath := r.calculateAbsolutePath(relativePath)
	pattern := fmt.Sprintf("%s %s", method, fullPath)
	next := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		rw, ok := w.(*ResponseWriter)
		if !ok {
			rw = NewResponseWriter(w)
		}
		ctx := &Ctx{
			req:    req,
			writer: rw,
			server: r.server,
		}
		if err := handler(ctx); err != nil {
			r.server.errHandler(rw, req, err)
		}
	}))
	next = chainMiddleware(middleware...)(next)
	next = chainMiddleware(r.md...)(next)
	wrapped := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		rw := NewResponseWriter(w)
		next.ServeHTTP(rw, req)
		if !rw.Written() {
			rw.WriteHeader(http.StatusOK)
		}
		rw.flush()
	})
	r.server.mux.Handle(pattern, wrapped)
	slog.Info("[Router]",
		slog.String("method", method),
		slog.String("path", fullPath),
		slog.Int("middlewares", len(r.md)+len(middleware)),
	)
}

func (r *Router) calculateAbsolutePath(relativePath string) string {
	if !r.server.hostMode {
		return path.Join(r.prefix, relativePath)
	}
	if strings.HasPrefix(r.prefix, "/") {
		return path.Join(r.prefix, relativePath)
	}
	return strings.TrimLeft(path.Join(r.prefix, relativePath), "/")
}

func (r *Router) GET(relativePath string, handler HandlerFunc, middleware ...Middleware) {
	r.handle(http.MethodGet, relativePath, handler, middleware...)
}

func (r *Router) POST(relativePath string, handler HandlerFunc, middleware ...Middleware) {
	r.handle(http.MethodPost, relativePath, handler, middleware...)
}

func (r *Router) PUT(relativePath string, handler HandlerFunc, middleware ...Middleware) {
	r.handle(http.MethodPut, relativePath, handler, middleware...)
}

func (r *Router) PATCH(relativePath string, handler HandlerFunc, middleware ...Middleware) {
	r.handle(http.MethodPatch, relativePath, handler, middleware...)
}

func (r *Router) DELETE(relativePath string, handler HandlerFunc, middleware ...Middleware) {
	r.handle(http.MethodDelete, relativePath, handler, middleware...)
}

func (r *Router) OPTIONS(relativePath string, handler HandlerFunc, middleware ...Middleware) {
	r.handle(http.MethodOptions, relativePath, handler, middleware...)
}

func (r *Router) HEAD(relativePath string, handler HandlerFunc, middleware ...Middleware) {
	r.handle(http.MethodHead, relativePath, handler, middleware...)
}

func (r *Router) CONNECT(relativePath string, handler HandlerFunc, middleware ...Middleware) {
	r.handle(http.MethodConnect, relativePath, handler, middleware...)
}

func (r *Router) TRACE(relativePath string, handler HandlerFunc, middleware ...Middleware) {
	r.handle(http.MethodTrace, relativePath, handler, middleware...)
}
