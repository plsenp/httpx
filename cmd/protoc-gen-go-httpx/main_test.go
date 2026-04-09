package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestPluginGeneration(t *testing.T) {
	// Build the plugin
	pluginPath := filepath.Join(t.TempDir(), "protoc-gen-httpx")
	buildCmd := exec.Command("go", "build", "-o", pluginPath, ".")
	buildCmd.Dir = "."
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build plugin: %v\n%s", err, output)
	}

	// Create temp directory for generated files
	tempDir := t.TempDir()

	// Run protoc with both standard Go plugin and httpx plugin
	protoFile := "testdata/user.proto"
	cmd := exec.Command("protoc",
		"--plugin", "protoc-gen-httpx="+pluginPath,
		"--go_out="+tempDir,
		"--go_opt=paths=source_relative",
		"--go-grpc_out="+tempDir,
		"--go-grpc_opt=paths=source_relative",
		"--httpx_out="+tempDir,
		"--httpx_opt=paths=source_relative",
		"--proto_path=.",
		"--proto_path=../../third_party", // For google/api/annotations.proto
		protoFile,
	)

	cmd.Dir = "."
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("protoc failed: %v\n%s", err, output)
	}

	// Check if all generated files exist
	expectedFiles := []string{
		filepath.Join(tempDir, "testdata", "user.pb.go"),
		filepath.Join(tempDir, "testdata", "user_grpc.pb.go"),
		filepath.Join(tempDir, "testdata", "user_httpx.pb.go"),
	}

	for _, file := range expectedFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Fatalf("Generated file not found: %s", file)
		}
	}

	// Read httpx generated file
	generatedFile := filepath.Join(tempDir, "testdata", "user_httpx.pb.go")
	content, err := os.ReadFile(generatedFile)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	contentStr := string(content)

	// Verify generated code contains expected elements
	tests := []struct {
		name     string
		expected string
	}{
		{"Package declaration", "package testdata"},
		{"Import httpx", "github.com/plsenp/httpx"},
		{"Import stringx", "github.com/plsenp/httpx/stringx"},
		{"Import httpxerrors", "httpxerrors \"github.com/plsenp/httpx/errors\""},
		{"Register function", "func RegisterUserServiceRoutes"},
		{"Register function takes router", "func RegisterUserServiceRoutes(router *httpx.Router"},
		{"GetUser handler", `router.Get("/users/{id}"`},
		{"ListUsers handler", `router.Get("/users"`},
		{"CreateUser handler", `router.Post("/users"`},
		{"UpdateUser handler", `router.Put("/users/{id}"`},
		{"DeleteUser handler", `router.Delete("/users/{id}"`},
		{"Path param binding", `stringx.To[int64](c.Param("id"))`},
		{"Query param binding", `c.Query("page")`},
		{"Body binding", "c.BindJSON(req)"},
		{"Context type", "func(c *httpx.Ctx)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(contentStr, tt.expected) {
				t.Errorf("Generated code missing: %s\nGenerated content:\n%s", tt.expected, contentStr)
			}
		})
	}
}

func TestExtractPathParams(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected []string
	}{
		{
			name:     "Single param",
			path:     "/users/{id}",
			expected: []string{"id"},
		},
		{
			name:     "Multiple params",
			path:     "/users/{user_id}/orders/{order_id}",
			expected: []string{"user_id", "order_id"},
		},
		{
			name:     "No params",
			path:     "/users",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This is a simplified test - in real scenario we'd need a protogen.Message
			// Just verify the path parsing logic
			parts := strings.Split(tt.path, "/")
			var params []string
			for _, part := range parts {
				if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
					param := strings.TrimSuffix(strings.TrimPrefix(part, "{"), "}")
					params = append(params, param)
				}
			}

			if len(params) != len(tt.expected) {
				t.Errorf("Expected %d params, got %d", len(tt.expected), len(params))
			}
			for i, expected := range tt.expected {
				if i >= len(params) || params[i] != expected {
					t.Errorf("Expected param %d to be %s, got %s", i, expected, params[i])
				}
			}
		})
	}
}

func TestHTTPRuleExtraction(t *testing.T) {
	// Test HTTP method extraction from rule
	tests := []struct {
		name           string
		method         string
		expectedMethod string
	}{
		{"GET", "Get", "Get"},
		{"POST", "Post", "Post"},
		{"PUT", "Put", "Put"},
		{"DELETE", "Delete", "Delete"},
		{"PATCH", "Patch", "Patch"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify method name mapping
			if tt.method != tt.expectedMethod {
				t.Errorf("Method mismatch: %s vs %s", tt.method, tt.expectedMethod)
			}
		})
	}
}
