# AGENT.md - httpx Library Usage Guide for AI Assistants

This document provides a comprehensive reference for AI assistants to generate correct, idiomatic code using the httpx library.

## Library Identity

- **Module**: `github.com/plsenp/httpx`
- **Go Version**: 1.24+
- **Type**: Lightweight web framework built on Go standard library `net/http`
- **Design Philosophy**: Standard library first, type-safe, zero magic

## Architecture Overview

```
httpx.NewServer() → *Server
  ├── .Route() → *Router (registers routes)
  │     ├── .GET/.POST/.PUT/.DELETE/.PATCH/.HEAD/.OPTIONS(path, handler, ...middleware)
  │     ├── .Group(prefix, ...middleware) → *Router
  │     └── .Mount(prefix, http.Handler, ...middleware)
  ├── .Start() error
  └── .Stop(ctx) error
```

## Core Types

### Server

```go
srv := httpx.NewServer(
    httpx.WithAddr(":8080"),
    httpx.WithReadTimeout(60*time.Second),
    httpx.WithReadHeaderTimeout(10*time.Second),
    httpx.WithWriteTimeout(60*time.Second),
    httpx.WithIdleTimeout(120*time.Second),
    httpx.WithMiddleware(middleware.Logger, middleware.Recovery),
    httpx.WithOpenAPISpec(openapi.Config{
        Title:   "API Title",
        Version: "1.0.0",
    }),
)
```

Server options (`httpx.With*`):
- `WithAddr(addr string)` — listen address (default `"0.0.0.0:8080"`)
- `WithMiddleware(md ...Middleware)` — global middleware
- `WithOpenAPISpec(spec openapi.Config)` — enable OpenAPI docs
- `WithReadTimeout`, `WithReadHeaderTimeout`, `WithWriteTimeout`, `WithIdleTimeout` — server timeouts

### Ctx (Request Context)

Every handler receives `*httpx.Ctx`. Key methods:

**Parameter Access:**
- `ctx.Param(key)` — path parameter value (string)
- `ctx.Query(key)` — query parameter value (string)
- `ctx.Header(key)` — request header value (string)

**Binding:**
- `ctx.Bind(&v)` — auto-bind (query+path for GET/HEAD/OPTIONS/DELETE; body+path for POST/PUT/PATCH)
- `ctx.BindQuery(&v)` — query only
- `ctx.BindPath(&v)` — path only
- `ctx.BindHeader(&v)` — headers only
- `ctx.BindBody(&v)` — body only (auto-detect content type)
- `ctx.BindJSON(&v)` — JSON body only
- `ctx.BindForm(&v)` — URL-encoded form only
- `ctx.BindMultipart(&v)` — multipart form with file support

**Response:**
- `ctx.JSON(code, v)` — write JSON response
- `ctx.String(code, format, args...)` — write plain text response
- `ctx.Data(code, contentType, data)` — write raw bytes
- `ctx.Render(code, v)` — content-negotiated response (uses Accept header)

**Validation:**
- `ctx.Validate(&v)` — validate struct using go-playground/validator tags

**Context Values:**
- `ctx.Get(key)` / `ctx.Set(key, value)` — request-scoped key-value store

**Underlying:**
- `ctx.Request()` → `*http.Request`
- `ctx.Writer()` → `*httpx.ResponseWriter`
- `ctx.RawWriter()` → `http.ResponseWriter`

### ResponseWriter

Wraps `http.ResponseWriter` with status tracking:
- `.Status()` — response status code
- `.Size()` — bytes written
- `.Written()` — whether headers have been sent
- Supports `http.Hijacker`, `http.Flusher`, `http.Pusher`

### Middleware

Type: `type Middleware func(http.Handler) http.Handler`

Standard `net/http` middleware pattern. Compatible with any third-party middleware that follows this signature.

## Two Handler Styles

### Style 1: Ctx Handler (Flexible)

```go
r.GET("/users/{id}", func(ctx *httpx.Ctx) error {
    id := ctx.Param("id")
    return ctx.JSON(200, map[string]string{"id": id})
})
```

Use when you need fine-grained control over binding, validation, and response.

### Style 2: Type-Safe Endpoint (Recommended)

```go
type GetUserReq struct {
    ID int64 `path:"id"`
}

type UserResp struct {
    ID   int64  `json:"id"`
    Name string `json:"name"`
}

httpx.GET(r, "/users/{id}", func(ctx context.Context, req *GetUserReq) (*UserResp, error) {
    return &UserResp{ID: req.ID, Name: "Alice"}, nil
}).
    Summary("Get a user").
    Tags("users").
    Register()
```

Handler signature: `func(ctx context.Context, req *Req) (*Resp, error)`

The builder automatically:
1. Binds request to `*Req`
2. Validates `*Req`
3. Calls handler
4. Renders `*Resp` (content-negotiated)
5. Handles errors via server error handler
6. Registers OpenAPI operation

Available builders: `httpx.GET`, `httpx.POST`, `httpx.PUT`, `httpx.DELETE`, `httpx.PATCH`, `httpx.HEAD`, `httpx.OPTIONS`

Builder chain methods:
- `.Summary(s)` — OpenAPI summary
- `.Description(s)` — OpenAPI description
- `.Tags(tags...)` — OpenAPI tags
- `.Middlewares(mws...)` — route-level middleware
- `.Register()` — **must call** to register the route

## Struct Tags Reference

| Tag | Source | Example |
|-----|--------|---------|
| `json:"name"` | JSON body field | `json:"username"` |
| `query:"name"` | Query parameter | `query:"page"` |
| `path:"name"` | Path parameter | `path:"id"` |
| `header:"name"` | HTTP header | `header:"X-Token"` |
| `form:"name"` | Form field | `form:"email"` |
| `file:"name"` | Uploaded file | `file:"avatar"` |
| `validate:"rules"` | Validation rules | `validate:"required,email"` |
| `description:"text"` | OpenAPI description | `description:"User name"` |
| `example:"value"` | OpenAPI example | `example:"John"` |

### File Upload Types

```go
type UploadReq struct {
    Avatar  string                    `file:"avatar"`   // *multipart.FileHeader
    Photos  []string                  `file:"photos"`   // []*multipart.FileHeader
}
```

## Routing Patterns

Uses Go 1.22+ `http.ServeMux` pattern syntax:
- `/users/{id}` — path parameter
- `/api/*` — catch-all wildcard
- `GET /users` — method + path (registered internally)

Route groups inherit middleware and prefix:
```go
api := r.Group("/api")
api.Use(AuthMiddleware)
api.GET("/users", listHandler)  // matches GET /api/users
```

## Error Handling

### Error Types

```go
import httpxerrors "github.com/plsenp/httpx/errors"

httpxerrors.NewBindError("message", err)           // 400, type=bind
httpxerrors.NewValidateError("message", err)        // 400, type=validate
httpxerrors.NewBusinessError(code, "message", err)  // custom code, type=business
httpxerrors.NewInternalError("message", err)        // 500, type=internal
```

### Error Type Checking

```go
httpxerrors.IsBindError(err)
httpxerrors.IsValidateError(err)
httpxerrors.IsBusinessError(err)
httpxerrors.IsInternalError(err)
httpxerrors.GetHTTPError(err) // returns *HTTPError or nil
```

### Custom Error Handler

The default error handler writes JSON. To customize, set `srv.errHandler` (unexported, needs ServerOption or modify source).

## Built-in Middleware

Import: `"github.com/plsenp/httpx/middleware"`

### Logger

```go
middleware.Logger  // logs method, path, status, duration
```

### Recovery

```go
middleware.Recovery  // catches panics, returns 500

middleware.RecoveryWithConfig(middleware.RecoveryConfig{
    Logger:   slog.Default(),
    LogStack: true,
    Handler:  func(w http.ResponseWriter, r *http.Request, err any) { ... },
})
```

### CORS

```go
middleware.CORS(middleware.DefaultCORSConfig)  // allow all origins

middleware.CORS(middleware.CORSConfig{
    AllowOrigins:     []string{"https://example.com"},
    AllowMethods:     []string{"GET", "POST"},
    AllowHeaders:     []string{"Content-Type", "Authorization"},
    AllowCredentials: true,
    ExposeHeaders:    []string{"X-Custom-Header"},
    MaxAge:           3600,
})
```

## OpenAPI Generation

### Enable

```go
import "github.com/plsenp/httpx/openapi"

srv := httpx.NewServer(httpx.WithOpenAPISpec(openapi.Config{
    Title:          "API Title",
    Description:    "API Description",
    Version:        "1.0.0",
    OpenAPIVersion: "3.0.3",  // optional, default "3.0.3"
}))
```

### Auto-registered Endpoints

- `GET /docs/openapi.json` — OpenAPI JSON spec
- `GET /docs` — Swagger UI

### How Schemas Are Generated

Type-safe endpoints (`httpx.GET`, `httpx.POST`, etc.) automatically register operations. The OpenAPI spec is generated from:
- Go struct field types → OpenAPI schema types
- `json` tags → property names
- `validate` tags → required fields + description annotations
- `description` tags → schema descriptions
- `example` tags → schema examples
- `query` tags → query parameters (for GET)
- `path` tags → path parameters

## Content Type Support

The `encoding` package registers codecs for content negotiation:

| Content Type | Codec Name | Import to register |
|---|---|---|
| `application/json` | `json` | `_ "github.com/plsenp/httpx/encoding/json"` |
| `application/xml` | `xml` | `_ "github.com/plsenp/httpx/encoding/xml"` |
| `application/yaml` | `yaml` | `_ "github.com/plsenp/httpx/encoding/yaml"` |
| `application/x-www-form-urlencoded` | `form` | `_ "github.com/plsenp/httpx/encoding/form"` |
| `application/protobuf` | `proto` | `_ "github.com/plsenp/httpx/encoding/proto"` |

All codecs are auto-registered via blank imports in `codec.go`.

## Protobuf Code Generation

The `protoc-gen-go-httpx` plugin generates HTTP route registration from `.proto` files with `google.api.http` annotations.

### Proto Definition

```protobuf
syntax = "proto3";
package myapp;
option go_package = "github.com/example/myapp";

import "google/api/annotations.proto";

service UserService {
  rpc GetUser(GetUserReq) returns (User) {
    option (google.api.http) = { get: "/users/{id}" };
  }
  rpc CreateUser(CreateUserReq) returns (User) {
    option (google.api.http) = { post: "/users" body: "*" };
  }
}
```

### Generate

```bash
protoc --go_out=. --go-httpx_out=. your.proto
```

### Generated Code

```go
type UserServiceHTTPServer interface {
    GetUser(context.Context, *GetUserReq) (*User, error)
    CreateUser(context.Context, *CreateUserReq) (*User, error)
}

func RegisterUserServiceRoutes(router *httpx.Router, service UserServiceHTTPServer)
```

### Usage

```go
type userService struct{}

func (s *userService) GetUser(ctx context.Context, req *GetUserReq) (*User, error) { ... }
func (s *userService) CreateUser(ctx context.Context, req *CreateUserReq) (*User, error) { ... }

func main() {
    srv := httpx.NewServer()
    r := srv.Route()
    RegisterUserServiceRoutes(r, &userService{})
    srv.Start()
}
```

## Common Patterns

### Full CRUD Example

```go
package main

import (
    "context"
    "net/http"

    "github.com/plsenp/httpx"
    "github.com/plsenp/httpx/middleware"
    "github.com/plsenp/httpx/openapi"
)

type User struct {
    ID    int64  `json:"id"`
    Name  string `json:"name" validate:"required" description:"User name" example:"Alice"`
    Email string `json:"email" validate:"required,email" description:"Email" example:"alice@example.com"`
}

type CreateUserReq struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
}

type ListUsersReq struct {
    Page int `query:"page" description:"Page number"`
    Size int `query:"size" description:"Page size"`
}

func main() {
    srv := httpx.NewServer(
        httpx.WithAddr(":8080"),
        httpx.WithMiddleware(middleware.Logger, middleware.Recovery),
        httpx.WithOpenAPISpec(openapi.Config{
            Title:   "User API",
            Version: "1.0.0",
        }),
    )
    r := srv.Route()

    httpx.GET(r, "/users", func(ctx context.Context, req *ListUsersReq) ([]*User, error) {
        return []*User{}, nil
    }).Summary("List users").Tags("users").Register()

    httpx.POST(r, "/users", func(ctx context.Context, req *CreateUserReq) (*User, error) {
        return &User{ID: 1, Name: req.Name, Email: req.Email}, nil
    }).Summary("Create user").Tags("users").Register()

    r.GET("/users/{id}", func(ctx *httpx.Ctx) error {
        id := ctx.Param("id")
        return ctx.JSON(http.StatusOK, map[string]string{"id": id})
    })

    srv.Start()
}
```

### Middleware Chain

```go
r := srv.Route()
r.Use(middleware.Logger)
r.Use(middleware.Recovery)

api := r.Group("/api")
api.Use(AuthMiddleware)

admin := api.Group("/admin")
admin.Use(AdminOnlyMiddleware)
admin.GET("/dashboard", handler)
```

### File Upload

```go
r.POST("/upload", func(ctx *httpx.Ctx) error {
    type UploadReq struct {
        Name   string `form:"name"`
        Avatar string `file:"avatar"`
    }
    var req UploadReq
    if err := ctx.BindMultipart(&req); err != nil {
        return err
    }
    return ctx.JSON(200, map[string]string{"name": req.Name})
})
```

## Important Notes

1. **Always call `.Register()`** on type-safe endpoint builders, otherwise routes are not registered.
2. **Handler return values**: Ctx-style handlers return `error`; type-safe handlers return `(*Resp, error)`.
3. **Path syntax**: Use `{id}` not `:id` or `{id:.+}`. This is Go 1.22+ ServeMux syntax.
4. **Default listen address**: `"0.0.0.0:8080"`. Override with `httpx.WithAddr()`.
5. **ResponseWriter wraps**: If you need the original `http.ResponseWriter`, use `ctx.RawWriter()`.
6. **Blank imports required**: Content codecs are registered via blank imports in `codec.go`. Do not remove them.
7. **Validation is automatic** in type-safe endpoints but **manual** in Ctx-style handlers (call `ctx.Validate()`).
8. **Error handler**: The default writes `{"code":400,"message":"..."}` JSON. Errors from handlers flow through `srv.errHandler`.
