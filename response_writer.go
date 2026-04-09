package httpx

import (
	"bufio"
	"net"
	"net/http"
)

type ResponseWriter struct {
	http.ResponseWriter
	status  int
	size    int
	written bool
}

func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,
		status:         http.StatusOK,
	}
}

func (w *ResponseWriter) WriteHeader(code int) {
	if w.written {
		return
	}
	w.status = code
}

func (w *ResponseWriter) Write(data []byte) (int, error) {
	if !w.written {
		w.flush()
	}
	n, err := w.ResponseWriter.Write(data)
	w.size += n
	return n, err
}

func (w *ResponseWriter) flush() {
	if w.written {
		return
	}
	w.ResponseWriter.WriteHeader(w.status)
	w.written = true
}

func (w *ResponseWriter) Status() int {
	return w.status
}

func (w *ResponseWriter) Size() int {
	return w.size
}

func (w *ResponseWriter) Written() bool {
	return w.written
}

func (w *ResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (w *ResponseWriter) Flush() {
	if !w.written {
		w.flush()
	}
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (w *ResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := w.ResponseWriter.(http.Hijacker); ok {
		return h.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}

func (w *ResponseWriter) Push(target string, opts *http.PushOptions) error {
	if p, ok := w.ResponseWriter.(http.Pusher); ok {
		return p.Push(target, opts)
	}
	return http.ErrNotSupported
}
