package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/plsenp/httpx"
)

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		status := http.StatusOK
		if rw, ok := w.(*httpx.ResponseWriter); ok {
			status = rw.Status()
		}

		slog.Info("Request processed", "method", r.Method, "path", r.URL.Path, "status", status, "duration", time.Since(start))
	})
}
