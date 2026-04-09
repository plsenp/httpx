package httpx

import (
	"context"
	"net/http"
)

// RouteBuilder is a generic route builder supporting method chaining
type RouteBuilder[Req any, Resp any] struct {
	router      *Router
	method      string
	path        string
	handler     func(ctx context.Context, req *Req) (*Resp, error)
	middlewares []Middleware

	// OpenAPI metadata
	summary     string
	description string
	tags        []string
}

// Summary sets the route summary, returns self for chaining
func (b *RouteBuilder[Req, Resp]) Summary(summary string) *RouteBuilder[Req, Resp] {
	b.summary = summary
	return b
}

// Description sets the route description, returns self for chaining
func (b *RouteBuilder[Req, Resp]) Description(description string) *RouteBuilder[Req, Resp] {
	b.description = description
	return b
}

// Tags sets the route tags, returns self for chaining
func (b *RouteBuilder[Req, Resp]) Tags(tags ...string) *RouteBuilder[Req, Resp] {
	b.tags = tags
	return b
}

// Middlewares sets the route middlewares, returns self for chaining
func (b *RouteBuilder[Req, Resp]) Middlewares(mws ...Middleware) *RouteBuilder[Req, Resp] {
	b.middlewares = mws
	return b
}

// Register registers the route to the router (end of chain)
func (b *RouteBuilder[Req, Resp]) Register() {
	b.register()
}

// httpMethod defines supported HTTP methods and their registration functions
type httpMethod struct {
	name      string
	registrar func(*Router, string, HandlerFunc, ...Middleware)
}

var methods = map[string]httpMethod{
	http.MethodGet: {
		name: http.MethodGet,
		registrar: func(r *Router, path string, h HandlerFunc, mws ...Middleware) {
			r.GET(path, h, mws...)
		},
	},
	http.MethodPost: {
		name: http.MethodPost,
		registrar: func(r *Router, path string, h HandlerFunc, mws ...Middleware) {
			r.POST(path, h, mws...)
		},
	},
	http.MethodPut: {
		name: http.MethodPut,
		registrar: func(r *Router, path string, h HandlerFunc, mws ...Middleware) {
			r.PUT(path, h, mws...)
		},
	},
	http.MethodDelete: {
		name: http.MethodDelete,
		registrar: func(r *Router, path string, h HandlerFunc, mws ...Middleware) {
			r.DELETE(path, h, mws...)
		},
	},
	http.MethodPatch: {
		name: http.MethodPatch,
		registrar: func(r *Router, path string, h HandlerFunc, mws ...Middleware) {
			r.PATCH(path, h, mws...)
		},
	},
	http.MethodHead: {
		name: http.MethodHead,
		registrar: func(r *Router, path string, h HandlerFunc, mws ...Middleware) {
			r.HEAD(path, h, mws...)
		},
	},
	http.MethodOptions: {
		name: http.MethodOptions,
		registrar: func(r *Router, path string, h HandlerFunc, mws ...Middleware) {
			r.OPTIONS(path, h, mws...)
		},
	},
}

// register registers the route to the router and OpenAPI
func (b *RouteBuilder[Req, Resp]) register() {
	// Register OpenAPI documentation (if spec provided)
	if b.router.server.openAPISpec != nil {
		var reqExemplar Req
		var respExemplar Resp
		b.router.server.openAPISpec.RegisterOperation(
			b.method, b.path,
			b.summary, b.description, b.tags,
			reqExemplar, respExemplar,
		)
	}

	// Wrap handler
	fn := func(c *Ctx) error {
		var req Req
		if err := c.Bind(&req); err != nil {
			return err
		}
		// Validate
		if err := c.Validate(&req); err != nil {
			return err
		}
		resp, err := b.handler(c.Request().Context(), &req)
		if err != nil {
			return err
		}
		return c.Render(http.StatusOK, resp)
	}

	// Register to router based on HTTP method
	if m, ok := methods[b.method]; ok {
		m.registrar(b.router, b.path, fn, b.middlewares...)
	}
}

// GET creates a GET route builder (not auto-registered, call Register needed)
func GET[Req any, Resp any](
	router *Router,
	path string,
	handler func(ctx context.Context, req *Req) (*Resp, error),
) *RouteBuilder[Req, Resp] {
	return &RouteBuilder[Req, Resp]{
		router:  router,
		method:  http.MethodGet,
		path:    path,
		handler: handler,
	}
}

// POST creates a POST route builder (not auto-registered, call Register needed)
func POST[Req any, Resp any](
	router *Router,
	path string,
	handler func(ctx context.Context, req *Req) (*Resp, error),
) *RouteBuilder[Req, Resp] {
	return &RouteBuilder[Req, Resp]{
		router:  router,
		method:  http.MethodPost,
		path:    path,
		handler: handler,
	}
}

// PUT creates a PUT route builder (not auto-registered, call Register needed)
func PUT[Req any, Resp any](
	router *Router,
	path string,
	handler func(ctx context.Context, req *Req) (*Resp, error),
) *RouteBuilder[Req, Resp] {
	return &RouteBuilder[Req, Resp]{
		router:  router,
		method:  http.MethodPut,
		path:    path,
		handler: handler,
	}
}

// DELETE creates a DELETE route builder (not auto-registered, call Register needed)
func DELETE[Req any, Resp any](
	router *Router,
	path string,
	handler func(ctx context.Context, req *Req) (*Resp, error),
) *RouteBuilder[Req, Resp] {
	return &RouteBuilder[Req, Resp]{
		router:  router,
		method:  http.MethodDelete,
		path:    path,
		handler: handler,
	}
}

// PATCH creates a PATCH route builder (not auto-registered, call Register needed)
func PATCH[Req any, Resp any](
	router *Router,
	path string,
	handler func(ctx context.Context, req *Req) (*Resp, error),
) *RouteBuilder[Req, Resp] {
	return &RouteBuilder[Req, Resp]{
		router:  router,
		method:  http.MethodPatch,
		path:    path,
		handler: handler,
	}
}

// HEAD creates a HEAD route builder (not auto-registered, call Register needed)
func HEAD[Req any, Resp any](
	router *Router,
	path string,
	handler func(ctx context.Context, req *Req) (*Resp, error),
) *RouteBuilder[Req, Resp] {
	return &RouteBuilder[Req, Resp]{
		router:  router,
		method:  http.MethodHead,
		path:    path,
		handler: handler,
	}
}

// OPTIONS creates an OPTIONS route builder (not auto-registered, call Register needed)
func OPTIONS[Req any, Resp any](
	router *Router,
	path string,
	handler func(ctx context.Context, req *Req) (*Resp, error),
) *RouteBuilder[Req, Resp] {
	return &RouteBuilder[Req, Resp]{
		router:  router,
		method:  http.MethodOptions,
		path:    path,
		handler: handler,
	}
}
