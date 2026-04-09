package httpx

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/plsenp/httpx/encoding"
	"github.com/plsenp/httpx/encoding/form"
	"github.com/plsenp/httpx/errors"
)

// Ctx encapsulates the HTTP request and response, providing convenient
// methods for binding, validation, and rendering. It is passed to all handlers
// and middleware in the framework.
type Ctx struct {
	req    *http.Request
	writer *ResponseWriter
	server *Server
}

func (c *Ctx) Get(key any) any {
	return c.req.Context().Value(key)
}

func (c *Ctx) Set(key, value any) {
	c.req = c.req.WithContext(context.WithValue(c.req.Context(), key, value))
}

// Request returns the http.Request.
func (c *Ctx) Request() *http.Request {
	return c.req
}

// Writer returns the ResponseWriter.
func (c *Ctx) Writer() *ResponseWriter {
	return c.writer
}

// RawWriter returns the underlying http.ResponseWriter.
func (c *Ctx) RawWriter() http.ResponseWriter {
	return c.writer.ResponseWriter
}

// Param gets path parameter by key.
func (c *Ctx) Param(key string) string {
	return c.req.PathValue(key)
}

// Query gets query parameter by key.
func (c *Ctx) Query(key string) string {
	return c.req.URL.Query().Get(key)
}

// Header gets header by key.
func (c *Ctx) Header(key string) string {
	return c.req.Header.Get(key)
}

// ========================= param binding =========================

// Bind binds request parameters to the struct.
// It supports query, path, and body binding.
// If the request method is GET, HEAD, OPTIONS, or DELETE, it binds query and path parameters.
// If the request method is POST, PUT or PATCH, it binds body and path parameters.
func (c *Ctx) Bind(v any) error {
	switch c.req.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodDelete:
		if err := c.BindQuery(v); err != nil {
			return err
		}
		return c.BindPath(v)
	default:
		if err := c.BindBody(v); err != nil {
			return err
		}
		return c.BindPath(v)
	}
}

// BindJSON binds JSON body to the struct.
func (c *Ctx) BindJSON(v any) error {
	codec := encoding.GetCodec("json")
	data, err := io.ReadAll(c.req.Body)
	if err != nil {
		return errors.NewBindError("decode json", err)
	}
	return codec.Unmarshal(data, v)
}

func (c *Ctx) BindForm(v any) error {
	codec := encoding.GetCodec(form.Name)
	if err := c.req.ParseForm(); err != nil {
		return errors.NewBindError("parse form", err)
	}
	return codec.Unmarshal([]byte(c.req.Form.Encode()), v)
}

func (c *Ctx) BindMultipart(v any) error {
	codec := encoding.GetCodec(form.Name)
	if err := c.req.ParseMultipartForm(1024 * 1024 * 10); err != nil {
		return errors.NewBindError("parse multipart form", err)
	}
	if err := codec.Unmarshal([]byte(c.req.Form.Encode()), v); err != nil {
		return errors.NewBindError("decode form", err)
	}
	if err := bindFiles(v, c.req.MultipartForm.File); err != nil {
		return errors.NewBindError("bind multipart files", err)
	}
	return nil
}

// BindQuery only binds query parameters.
func (c *Ctx) BindQuery(v any) error {
	return c.server.queryBind(c.req, v)
}

// BindPath only binds path parameters.
func (c *Ctx) BindPath(v any) error {
	return c.server.pathBind(c.req, v)
}

// BindHeader only binds headers.
func (c *Ctx) BindHeader(v any) error {
	return c.server.headerBind(c.req, v)
}

// BindBody only binds request body.
func (c *Ctx) BindBody(v any) error {
	return c.server.bodyBind(c.req, v)
}

// ========================= response rendering =========================

// Render writes response using the configured renderer.
func (c *Ctx) Render(code int, v any) error {
	return c.server.renderFunc(c.writer, c.req, code, v)
}

// JSON writes JSON response (alias for Render).
func (c *Ctx) JSON(code int, v any) error {
	c.writer.Header().Set("Content-Type", "application/json")
	c.writer.WriteHeader(code)
	if err := json.NewEncoder(c.writer).Encode(v); err != nil {
		return err
	}
	return nil
}

// Data writes raw bytes response with custom content type.
func (c *Ctx) Data(code int, contentType string, data []byte) error {
	c.writer.Header().Set("Content-Type", contentType)
	c.writer.WriteHeader(code)
	_, err := c.writer.Write(data)
	return err
}

// String writes string response.
func (c *Ctx) String(code int, format string, args ...any) error {
	return c.Data(code, "text/plain; charset=utf-8", []byte(fmt.Sprintf(format, args...)))
}

// Validate validates the struct.
func (c *Ctx) Validate(v any) error {
	return c.server.validate.Validate(v)
}
