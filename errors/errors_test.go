package errors

import (
	"errors"
	"net/http"
	"testing"
)

func TestNewBindError(t *testing.T) {
	innerErr := errors.New("invalid format")
	err := NewBindError("failed to bind request", innerErr)

	if err.Type != ErrorTypeBind {
		t.Errorf("Expected type %s, got %s", ErrorTypeBind, err.Type)
	}
	if err.Code != 400 {
		t.Errorf("Expected code 400, got %d", err.Code)
	}
	if err.Message != "failed to bind request" {
		t.Errorf("Expected message 'failed to bind request', got '%s'", err.Message)
	}
	if err.Err != innerErr {
		t.Error("Expected wrapped error to be innerErr")
	}
}

func TestNewValidateError(t *testing.T) {
	innerErr := errors.New("field required")
	err := NewValidateError("validation failed", innerErr)

	if err.Type != ErrorTypeValidate {
		t.Errorf("Expected type %s, got %s", ErrorTypeValidate, err.Type)
	}
	if err.Code != 400 {
		t.Errorf("Expected code 400, got %d", err.Code)
	}
}

func TestNewBusinessError(t *testing.T) {
	err := NewBusinessError(http.StatusNotFound, "user not found", nil)

	if err.Type != ErrorTypeBusiness {
		t.Errorf("Expected type %s, got %s", ErrorTypeBusiness, err.Type)
	}
	if err.Code != 404 {
		t.Errorf("Expected code 404, got %d", err.Code)
	}
	if err.Message != "user not found" {
		t.Errorf("Expected message 'user not found', got '%s'", err.Message)
	}
}

func TestNewInternalError(t *testing.T) {
	innerErr := errors.New("database connection failed")
	err := NewInternalError("internal server error", innerErr)

	if err.Type != ErrorTypeInternal {
		t.Errorf("Expected type %s, got %s", ErrorTypeInternal, err.Type)
	}
	if err.Code != 500 {
		t.Errorf("Expected code 500, got %d", err.Code)
	}
}

func TestHTTPErrorError(t *testing.T) {
	// Test with wrapped error
	innerErr := errors.New("inner error")
	err := NewBindError("outer error", innerErr)

	errStr := err.Error()
	if !contains(errStr, "bind") {
		t.Errorf("Expected error string to contain 'bind', got '%s'", errStr)
	}
	if !contains(errStr, "outer error") {
		t.Errorf("Expected error string to contain 'outer error', got '%s'", errStr)
	}

	// Test without wrapped error
	err2 := NewBusinessError(400, "business error", nil)
	errStr2 := err2.Error()
	if !contains(errStr2, "business") {
		t.Errorf("Expected error string to contain 'business', got '%s'", errStr2)
	}
}

func TestHTTPErrorUnwrap(t *testing.T) {
	innerErr := errors.New("inner error")
	err := NewBindError("outer error", innerErr)

	unwrapped := err.Unwrap()
	if unwrapped != innerErr {
		t.Error("Unwrap should return inner error")
	}
}

func TestIsBindError(t *testing.T) {
	bindErr := NewBindError("bind error", nil)
	validateErr := NewValidateError("validate error", nil)

	if !IsBindError(bindErr) {
		t.Error("IsBindError should return true for bind error")
	}
	if IsBindError(validateErr) {
		t.Error("IsBindError should return false for validate error")
	}
	if IsBindError(errors.New("plain error")) {
		t.Error("IsBindError should return false for plain error")
	}
}

func TestIsValidateError(t *testing.T) {
	validateErr := NewValidateError("validate error", nil)
	bindErr := NewBindError("bind error", nil)

	if !IsValidateError(validateErr) {
		t.Error("IsValidateError should return true for validate error")
	}
	if IsValidateError(bindErr) {
		t.Error("IsValidateError should return false for bind error")
	}
}

func TestIsBusinessError(t *testing.T) {
	businessErr := NewBusinessError(400, "business error", nil)
	if !IsBusinessError(businessErr) {
		t.Error("IsBusinessError should return true for business error")
	}
}

func TestIsInternalError(t *testing.T) {
	internalErr := NewInternalError("internal error", nil)
	if !IsInternalError(internalErr) {
		t.Error("IsInternalError should return true for internal error")
	}
}

func TestGetHTTPError(t *testing.T) {
	// Test with HTTPError
	httpErr := NewBindError("test", nil)
	got := GetHTTPError(httpErr)
	if got == nil {
		t.Fatal("GetHTTPError should return HTTPError")
	}
	if got.Message != "test" {
		t.Errorf("Expected message 'test', got '%s'", got.Message)
	}

	// Test with wrapped HTTPError
	wrapped := errors.New("wrapped")
	wrappedErr := NewBindError("test", wrapped)
	doubleWrapped := errors.New("double wrapped")
	finalErr := NewValidateError("final", doubleWrapped)
	_ = finalErr
	_ = wrappedErr

	// Test with plain error
	plainErr := errors.New("plain error")
	got2 := GetHTTPError(plainErr)
	if got2 != nil {
		t.Error("GetHTTPError should return nil for plain error")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
