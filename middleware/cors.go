package middleware

import (
	"net/http"
	"strconv"
	"strings"
)

// CORSConfig is the CORS configuration
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	AllowCredentials bool
	ExposeHeaders    []string
	MaxAge           int
}

// DefaultCORSConfig is the default CORS configuration
var DefaultCORSConfig = CORSConfig{
	AllowOrigins: []string{"*"},
	AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
	AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
	MaxAge:       86400,
}

// CORS returns a CORS middleware with custom config
func CORS(config CORSConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set CORS headers
			if len(config.AllowOrigins) > 0 {
				origin := r.Header.Get("Origin")
				allowed := false
				for _, o := range config.AllowOrigins {
					if o == "*" || o == origin {
						allowed = true
						break
					}
				}
				if allowed {
					if config.AllowCredentials {
						// When credentials are allowed, cannot use wildcard, echo the origin instead
						w.Header().Set("Access-Control-Allow-Origin", origin)
						w.Header().Set("Vary", "Origin")
					} else {
						// Always echo origin for consistency, even with wildcard
						w.Header().Set("Access-Control-Allow-Origin", origin)
					}
				}
			}

			if len(config.AllowMethods) > 0 {
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowMethods, ", "))
			}

			if len(config.AllowHeaders) > 0 {
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowHeaders, ", "))
			}

			if config.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if len(config.ExposeHeaders) > 0 {
				w.Header().Set("Access-Control-Expose-Headers", strings.Join(config.ExposeHeaders, ", "))
			}

			if config.MaxAge > 0 {
				w.Header().Set("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))
			}

			// Handle preflight request
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
