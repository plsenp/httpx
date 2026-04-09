# httpx

A lightweight, type-safe Go web framework built on the standard library.

## Features

- **Standard Library First** — Built on `net/http.ServeMux` (Go 1.22+ pattern syntax), fully compatible with `http.Handler`
- **Type-Safe Endpoints** — Generic route builders with `func(ctx context.Context, req *Req) (*Resp, error)` signature
- **Auto OpenAPI Generation** — Generate OpenAPI 3.0 specs from Go struct tags, with built-in Swagger UI
- **Flexible Binding** — Automatic binding from query, path, header, body (JSON/XML/YAML/Form/Proto)
- **Built-in Middleware** — Logger, Recovery, CORS out of the box
- **Protobuf Code Generation** — `protoc-gen-go-httpx` plugin for gRPC-Gateway style HTTP APIs

## Installation

```bash
go get github.com/plsenp/httpx
```

Requires Go 1.24+.

## Quick Start

```go
package main

import (
    "net/http"

    "github.com/plsenp/httpx"
)

func main() {
    srv := httpx.NewServer()
    r := srv.Route()

    r.GET("/hello", func(ctx *httpx.Ctx) error {
        return ctx.String(http.StatusOK, "Hello, World!")
    })

    srv.Start()
}
```

## Routing

### Basic Routes

```go
r := srv.Route()

r.GET("/users", listUsers)
r.POST("/users", createUser)
r.PUT("/users/{id}", updateUser)
r.DELETE("/users/{id}", deleteUser)
```

### Route Groups

```go
api := r.Group("/api")
api.GET("/users", listUsers)

v1 := api.Group("/v1")
v1.GET("/status", getStatus)
```

### Mount http.Handler

```go
r.Mount("/assets", fileServer)
```

## Type-Safe Endpoints

Use generic builders for compile-time type safety and automatic OpenAPI documentation:

```go
type CreateUserReq struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
}

type UserResp struct {
    ID    int64  `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

httpx.POST(r, "/users", func(ctx context.Context, req *CreateUserReq) (*UserResp, error) {
    return &UserResp{ID: 1, Name: req.Name, Email: req.Email}, nil
}).
    Summary("Create a user").
    Description("Creates a new user with the given name and email").
    Tags("users").
    Register()
```

The builder automatically:
- Binds request parameters to `Req`
- Validates using `go-playground/validator` tags
- Renders response as JSON (or other content types via `Accept` header)
- Registers OpenAPI documentation

### Supported Methods

`httpx.GET`, `httpx.POST`, `httpx.PUT`, `httpx.DELETE`, `httpx.PATCH`, `httpx.HEAD`, `httpx.OPTIONS`

## Request Binding

### Automatic Binding

`ctx.Bind(&v)` automatically selects the binding source based on HTTP method:

| Method                     | Binding Sources |
| -------------------------- | --------------- |
| GET, HEAD, OPTIONS, DELETE | Query + Path    |
| POST, PUT, PATCH           | Body + Path     |

### Manual Binding

```go
ctx.BindQuery(&v)     // Query parameters only
ctx.BindPath(&v)      // Path parameters only
ctx.BindHeader(&v)    // Headers only
ctx.BindBody(&v)      // Request body only
ctx.BindJSON(&v)      // JSON body only
ctx.BindForm(&v)      // URL-encoded form only
ctx.BindMultipart(&v) // Multipart form (including files)
```

### Struct Tags

```go
type SearchReq struct {
    ID     string `path:"id"`                          // Path parameter
    Q      string `query:"q"`                          // Query parameter
    Token  string `header:"X-Token"`                   // Header
    Name   string `json:"name"`                        // JSON body
    Avatar string `file:"avatar"`                      // Uploaded file (*multipart.FileHeader)
    Photos []string `file:"photos"`                    // Multiple files ([]*multipart.FileHeader)
}
```

## Response Rendering

```go
ctx.JSON(code, data)                  // JSON response
ctx.String(code, "Hello %s", name)    // Plain text
ctx.Data(code, "image/png", bytes)    // Raw bytes
ctx.Render(code, data)                // Auto-negotiate by Accept header
```

## Validation

Uses [go-playground/validator](https://github.com/go-playground/validator) tags:

```go
type Req struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"gte=0,lte=150"`
}
```

Call `ctx.Validate(&req)` or use type-safe endpoints (auto-validated).

## Middleware

### Built-in Middleware

```go
import "github.com/plsenp/httpx/middleware"

srv := httpx.NewServer(
    httpx.WithMiddleware(
        middleware.Logger,
        middleware.Recovery,
        middleware.CORS(middleware.DefaultCORSConfig),
    ),
)
```

### Custom Middleware

Middleware follows the standard `func(http.Handler) http.Handler` pattern:

```go
func Auth(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if token == "" {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

### Route-Level Middleware

```go
r.GET("/admin", handler, Auth)
r.Use(Auth) // applies to all routes in this group
```

### CORS Configuration

```go
corsConfig := middleware.CORSConfig{
    AllowOrigins:     []string{"https://example.com"},
    AllowMethods:     []string{"GET", "POST"},
    AllowHeaders:     []string{"Content-Type", "Authorization"},
    AllowCredentials: true,
    ExposeHeaders:    []string{"X-Custom-Header"},
    MaxAge:           3600,
}
middleware.CORS(corsConfig)
```

## OpenAPI Documentation

### Enable OpenAPI

```go
import "github.com/plsenp/httpx/openapi"

srv := httpx.NewServer(httpx.WithOpenAPISpec(openapi.Config{
    Title:       "My API",
    Version:     "1.0.0",
    Description: "My awesome API",
}))
```

### Endpoints

| Path                 | Description           |
| -------------------- | --------------------- |
| `/docs/openapi.json` | OpenAPI 3.0 JSON spec |
| `/docs`              | Swagger UI            |

### Struct Tags for OpenAPI

```go
type User struct {
    Name  string `json:"name" validate:"required" description:"User's full name" example:"John Doe"`
    Email string `json:"email" validate:"required,email" description:"Email address" example:"john@example.com"`
    Age   int    `json:"age" description:"Age in years" example:"30"`
}
```

Type-safe endpoint builders automatically register OpenAPI operations with summary, description, tags, and schema.

## Error Handling

```go
import httpxerrors "github.com/plsenp/httpx/errors"

// Error types
httpxerrors.NewBindError("message", err)       // 400
httpxerrors.NewValidateError("message", err)   // 400
httpxerrors.NewBusinessError(409, "conflict", err) // Custom code
httpxerrors.NewInternalError("message", err)   // 500

// Type checking
httpxerrors.IsBindError(err)
httpxerrors.IsValidateError(err)
httpxerrors.IsBusinessError(err)
httpxerrors.IsInternalError(err)
```

## Server Configuration

```go
srv := httpx.NewServer(
    httpx.WithAddr(":9090"),
    httpx.WithReadTimeout(30*time.Second),
    httpx.WithReadHeaderTimeout(10*time.Second),
    httpx.WithWriteTimeout(30*time.Second),
    httpx.WithIdleTimeout(60*time.Second),
    httpx.WithMiddleware(middleware.Logger, middleware.Recovery),
    httpx.WithOpenAPISpec(openapi.Config{
        Title:   "My API",
        Version: "1.0.0",
    }),
)
```

## Protobuf Code Generation

Use `protoc-gen-go-httpx` to generate HTTP routes from protobuf definitions with `google.api.http` annotations:

```bash
protoc --go_out=. --go-httpx_out=. your.proto
```

Generated code provides:

```go
type UserServiceHTTPServer interface {
    GetUser(context.Context, *GetUserReq) (*User, error)
    CreateUser(context.Context, *CreateUserReq) (*User, error)
}

func RegisterUserServiceRoutes(router *httpx.Router, service UserServiceHTTPServer)
```

## Context API

| Method                              | Description                      |
| ----------------------------------- | -------------------------------- |
| `ctx.Bind(&v)`                      | Auto-bind request parameters     |
| `ctx.Validate(&v)`                  | Validate struct                  |
| `ctx.JSON(code, v)`                 | JSON response                    |
| `ctx.String(code, format, args...)` | Text response                    |
| `ctx.Data(code, contentType, data)` | Raw data response                |
| `ctx.Render(code, v)`               | Content-negotiated response      |
| `ctx.Param(key)`                    | Path parameter                   |
| `ctx.Query(key)`                    | Query parameter                  |
| `ctx.Header(key)`                   | Request header                   |
| `ctx.Request()`                     | Underlying `*http.Request`       |
| `ctx.Writer()`                      | `*ResponseWriter`                |
| `ctx.RawWriter()`                   | Underlying `http.ResponseWriter` |
| `ctx.Get(key)`                      | Get context value                |
| `ctx.Set(key, value)`               | Set context value                |

## Content Negotiation

The framework supports multiple content types via `Content-Type` and `Accept` headers:

- `application/json` (default)
- `application/xml`
- `application/yaml`
- `application/x-www-form-urlencoded`
- `multipart/form-data`
- `application/protobuf`

## License

MIT
