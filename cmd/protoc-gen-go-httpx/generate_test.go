package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestGenerateExample generates an example file for inspection
func TestGenerateExample(t *testing.T) {
	// Build the plugin
	pluginPath := filepath.Join(t.TempDir(), "protoc-gen-httpx")
	buildCmd := exec.Command("go", "build", "-o", pluginPath, ".")
	buildCmd.Dir = "."
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build plugin: %v\n%s", err, output)
	}

	// Use temp directory as output
	outputDir := t.TempDir()

	// Run protoc with both standard Go plugin and httpx plugin
	protoFile := "testdata/user.proto"
	cmd := exec.Command("protoc",
		"--plugin", "protoc-gen-httpx="+pluginPath,
		"--go_out="+outputDir,
		"--go_opt=paths=source_relative",
		"--go-grpc_out="+outputDir,
		"--go-grpc_opt=paths=source_relative",
		"--httpx_out="+outputDir,
		"--httpx_opt=paths=source_relative",
		"--proto_path=.",
		"--proto_path=../../third_party",
		protoFile,
	)

	cmd.Dir = "."
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("protoc failed: %v\n%s", err, output)
	}

	// Read and print generated httpx file
	generatedFile := filepath.Join(outputDir, "testdata", "user_httpx.pb.go")
	content, err := os.ReadFile(generatedFile)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	t.Logf("Generated file: %s\n%s", generatedFile, string(content))

	// Verify content
	contentStr := string(content)
	expectedElements := []string{
		"package testdata",
		"func RegisterUserServiceRoutes",
		`router.Get("/users/{id}"`,
		`stringx.To[int64](c.Param("id"))`,
		`c.Query("page")`,
		"c.BindJSON(req)",
	}

	for _, expected := range expectedElements {
		if !strings.Contains(contentStr, expected) {
			t.Errorf("Generated code missing: %s", expected)
		}
	}

	// Verify that standard pb.go files are also generated
	pbFile := filepath.Join(outputDir, "testdata", "user.pb.go")
	if _, err := os.Stat(pbFile); os.IsNotExist(err) {
		t.Errorf("Standard pb.go file not generated: %s", pbFile)
	}

	grpcFile := filepath.Join(outputDir, "testdata", "user_grpc.pb.go")
	if _, err := os.Stat(grpcFile); os.IsNotExist(err) {
		t.Errorf("gRPC pb.go file not generated: %s", grpcFile)
	}
}
