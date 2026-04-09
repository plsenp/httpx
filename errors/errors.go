// Package errors provides standardized error types for httpx
package errors

import (
	"errors"
	"fmt"
)

// ErrorType represents the type of error
type ErrorType string

const (
	// ErrorTypeBind represents binding errors (path, query, body, header)
	ErrorTypeBind ErrorType = "bind"
	// ErrorTypeValidate represents validation errors
	ErrorTypeValidate ErrorType = "validate"
	// ErrorTypeBusiness represents business logic errors from service layer
	ErrorTypeBusiness ErrorType = "business"
	// ErrorTypeInternal represents internal server errors
	ErrorTypeInternal ErrorType = "internal"
)

// HTTPError is a standardized HTTP error with type information
type HTTPError struct {
	Type    ErrorType `json:"-"`
	Code    int       `json:"code"`
	Message string    `json:"message"`
	Details string    `json:"details,omitempty"`
	Err     error     `json:"-"`
}

// Error implements the error interface
func (e *HTTPError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

// Unwrap returns the wrapped error
func (e *HTTPError) Unwrap() error {
	return e.Err
}

// NewBindError creates a new binding error
func NewBindError(message string, err error) *HTTPError {
	return &HTTPError{
		Type:    ErrorTypeBind,
		Code:    400,
		Message: message,
		Err:     err,
	}
}

// NewValidateError creates a new validation error
func NewValidateError(message string, err error) *HTTPError {
	return &HTTPError{
		Type:    ErrorTypeValidate,
		Code:    400,
		Message: message,
		Err:     err,
	}
}

// NewBusinessError creates a new business logic error
func NewBusinessError(code int, message string, err error) *HTTPError {
	return &HTTPError{
		Type:    ErrorTypeBusiness,
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// NewInternalError creates a new internal server error
func NewInternalError(message string, err error) *HTTPError {
	return &HTTPError{
		Type:    ErrorTypeInternal,
		Code:    500,
		Message: message,
		Err:     err,
	}
}

// IsBindError checks if the error is a binding error
func IsBindError(err error) bool {
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.Type == ErrorTypeBind
	}
	return false
}

// IsValidateError checks if the error is a validation error
func IsValidateError(err error) bool {
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.Type == ErrorTypeValidate
	}
	return false
}

// IsBusinessError checks if the error is a business logic error
func IsBusinessError(err error) bool {
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.Type == ErrorTypeBusiness
	}
	return false
}

// IsInternalError checks if the error is an internal error
func IsInternalError(err error) bool {
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.Type == ErrorTypeInternal
	}
	return false
}

// GetHTTPError extracts HTTPError from error if present
func GetHTTPError(err error) *HTTPError {
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		return httpErr
	}
	return nil
}
