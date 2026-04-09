package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/plsenp/httpx"
)

type RecoveryConfig struct {
	Logger   *slog.Logger
	Handler  func(w http.ResponseWriter, r *http.Request, err any)
	LogStack bool
}

func DefaultRecoveryConfig() RecoveryConfig {
	return RecoveryConfig{
		Logger:   slog.Default(),
		LogStack: true,
	}
}

func Recovery(next http.Handler) http.Handler {
	return RecoveryWithConfig(DefaultRecoveryConfig())(next)
}

func RecoveryWithConfig(config RecoveryConfig) func(http.Handler) http.Handler {
	logger := config.Logger
	if logger == nil {
		logger = slog.Default()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					attrs := []slog.Attr{
						slog.String("error", fmt.Sprintf("%v", err)),
						slog.String("method", r.Method),
						slog.String("path", r.URL.Path),
						slog.String("remote_addr", r.RemoteAddr),
					}
					if config.LogStack {
						attrs = append(attrs, slog.String("stack", string(debug.Stack())))
					}
					logger.LogAttrs(r.Context(), slog.LevelError, "panic recovered", attrs...)

					if config.Handler != nil {
						config.Handler(w, r, err)
					} else {
						defaultPanicHandler(w, r, err)
					}
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func defaultPanicHandler(w http.ResponseWriter, r *http.Request, _ any) {
	rw, ok := w.(*httpx.ResponseWriter)
	if ok {
		if !rw.Written() {
			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
	}
	fmt.Fprintf(w, `{"error":"%s","message":"%s"}`, "internal_server_error", "Internal server error")
}
