package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLogger(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap with logger middleware
	loggedHandler := Logger(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	loggedHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	// Logger writes to stdout, we can't easily capture it
	// But we can verify the handler still works
	if rec.Body.String() != "OK" {
		t.Errorf("Expected body 'OK', got '%s'", rec.Body.String())
	}
}

func TestRecovery(t *testing.T) {
	tests := []struct {
		name     string
		handler  http.Handler
		wantCode int
		wantBody string
	}{
		{
			name: "normal request",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			}),
			wantCode: http.StatusOK,
			wantBody: "OK",
		},
		{
			name: "panic recovery",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic("something went wrong")
			}),
			wantCode: http.StatusInternalServerError,
			wantBody: "internal_server_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recoveryHandler := Recovery(tt.handler)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()

			recoveryHandler.ServeHTTP(rec, req)

			if rec.Code != tt.wantCode {
				t.Errorf("Expected status %d, got %d", tt.wantCode, rec.Code)
			}

			if !strings.Contains(rec.Body.String(), tt.wantBody) {
				t.Errorf("Expected body to contain '%s', got '%s'", tt.wantBody, rec.Body.String())
			}
		})
	}
}

func TestRecoveryWithCustomHandler(t *testing.T) {
	customHandler := func(w http.ResponseWriter, r *http.Request, err any) {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"error":"custom_error"}`))
	}

	config := RecoveryConfig{
		LogStack: false,
		Handler:  customHandler,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	recoveryHandler := RecoveryWithConfig(config)(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	recoveryHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", rec.Code)
	}

	if !strings.Contains(rec.Body.String(), "custom_error") {
		t.Errorf("Expected custom error, got '%s'", rec.Body.String())
	}
}

func TestCORS(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	tests := []struct {
		name       string
		config     CORSConfig
		origin     string
		method     string
		wantOrigin string
		wantCode   int
	}{
		{
			name:       "allow all origins",
			config:     DefaultCORSConfig,
			origin:     "http://example.com",
			method:     http.MethodGet,
			wantOrigin: "http://example.com",
			wantCode:   http.StatusOK,
		},
		{
			name: "specific origin allowed",
			config: CORSConfig{
				AllowOrigins: []string{"http://allowed.com"},
				AllowMethods: []string{"GET", "POST"},
			},
			origin:     "http://allowed.com",
			method:     http.MethodGet,
			wantOrigin: "http://allowed.com",
			wantCode:   http.StatusOK,
		},
		{
			name: "specific origin not allowed",
			config: CORSConfig{
				AllowOrigins: []string{"http://allowed.com"},
				AllowMethods: []string{"GET", "POST"},
			},
			origin:     "http://notallowed.com",
			method:     http.MethodGet,
			wantOrigin: "",
			wantCode:   http.StatusOK,
		},
		{
			name:       "preflight request",
			config:     DefaultCORSConfig,
			origin:     "http://example.com",
			method:     http.MethodOptions,
			wantOrigin: "http://example.com",
			wantCode:   http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			corsHandler := CORS(tt.config)(handler)

			req := httptest.NewRequest(tt.method, "/test", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			rec := httptest.NewRecorder()

			corsHandler.ServeHTTP(rec, req)

			if rec.Code != tt.wantCode {
				t.Errorf("Expected status %d, got %d", tt.wantCode, rec.Code)
			}

			gotOrigin := rec.Header().Get("Access-Control-Allow-Origin")
			if gotOrigin != tt.wantOrigin {
				t.Errorf("Expected Access-Control-Allow-Origin '%s', got '%s'", tt.wantOrigin, gotOrigin)
			}
		})
	}
}

func TestCORSHeaders(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	config := CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		ExposeHeaders:    []string{"X-Custom-Header"},
		MaxAge:           3600,
	}

	corsHandler := CORS(config)(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	rec := httptest.NewRecorder()

	corsHandler.ServeHTTP(rec, req)

	// Check all CORS headers
	tests := []struct {
		header   string
		expected string
	}{
		{"Access-Control-Allow-Origin", "http://example.com"},
		{"Access-Control-Allow-Methods", "GET, POST, PUT, DELETE"},
		{"Access-Control-Allow-Headers", "Content-Type, Authorization"},
		{"Access-Control-Allow-Credentials", "true"},
		{"Access-Control-Expose-Headers", "X-Custom-Header"},
	}

	for _, tt := range tests {
		got := rec.Header().Get(tt.header)
		if got != tt.expected {
			t.Errorf("Expected %s '%s', got '%s'", tt.header, tt.expected, got)
		}
	}
}
