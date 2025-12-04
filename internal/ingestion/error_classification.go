// Package ingestion provides error classification and handling for cloud ingestion.
package ingestion

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// ErrorClass categorizes errors for retry decisions.
type ErrorClass string

const (
	// ErrorClassTransient indicates a temporary error that can be retried.
	// Examples: rate limiting (429), timeout, temporary service outage
	ErrorClassTransient ErrorClass = "transient"

	// ErrorClassAuth indicates authentication/authorization failure (won't recover without intervention).
	// Examples: invalid credentials, expired token, permission denied
	ErrorClassAuth ErrorClass = "auth"

	// ErrorClassBadRequest indicates invalid request (won't recover with same request).
	// Examples: malformed JSON, invalid parameters, schema mismatch
	ErrorClassBadRequest ErrorClass = "bad_request"

	// ErrorClassNotFound indicates resource doesn't exist (retrying won't help).
	// Examples: bucket not found, file doesn't exist
	ErrorClassNotFound ErrorClass = "not_found"

	// ErrorClassFatal indicates an unrecoverable error.
	// Examples: disk full, internal server error (5xx)
	ErrorClassFatal ErrorClass = "fatal"

	// ErrorClassUnknown indicates we couldn't determine the error class.
	ErrorClassUnknown ErrorClass = "unknown"
)

// ClassifiedError wraps an error with its classification for retry decisions.
type ClassifiedError struct {
	Class   ErrorClass
	Message string
	Wrapped error
}

// NewClassifiedError creates a classified error.
func NewClassifiedError(class ErrorClass, message string, wrapped error) *ClassifiedError {
	return &ClassifiedError{
		Class:   class,
		Message: message,
		Wrapped: wrapped,
	}
}

// Error implements the error interface.
func (ce *ClassifiedError) Error() string {
	if ce.Wrapped != nil {
		return fmt.Sprintf("%s: %s (%v)", ce.Class, ce.Message, ce.Wrapped)
	}
	return fmt.Sprintf("%s: %s", ce.Class, ce.Message)
}

// Unwrap implements error unwrapping for error chains.
func (ce *ClassifiedError) Unwrap() error {
	return ce.Wrapped
}

// IsRetryable returns true if the error should trigger a retry.
func (ce *ClassifiedError) IsRetryable() bool {
	return ce.Class == ErrorClassTransient
}

// ClassifyError categorizes an error based on type and HTTP status codes.
// This is the main entry point for error classification.
func ClassifyError(err error) *ClassifiedError {
	if err == nil {
		return nil
	}

	// Check if already classified
	var ce *ClassifiedError
	if errors.As(err, &ce) {
		return ce
	}

	// Classify by error type or message
	errStr := err.Error()

	// Authentication/Authorization errors
	if strings.Contains(errStr, "unauthorized") ||
		strings.Contains(errStr, "403") ||
		strings.Contains(errStr, "invalid credentials") ||
		strings.Contains(errStr, "access denied") ||
		strings.Contains(errStr, "permission denied") {
		return NewClassifiedError(ErrorClassAuth, "authentication or authorization failed", err)
	}

	// Not found errors
	if strings.Contains(errStr, "404") ||
		strings.Contains(errStr, "not found") ||
		strings.Contains(errStr, "does not exist") ||
		strings.Contains(errStr, "NoSuchBucket") ||
		strings.Contains(errStr, "NoSuchKey") {
		return NewClassifiedError(ErrorClassNotFound, "resource not found", err)
	}

	// Bad request errors
	if strings.Contains(errStr, "400") ||
		strings.Contains(errStr, "bad request") ||
		strings.Contains(errStr, "invalid") ||
		strings.Contains(errStr, "malformed") ||
		strings.Contains(errStr, "validation error") {
		return NewClassifiedError(ErrorClassBadRequest, "bad request or invalid data", err)
	}

	// Transient errors (rate limiting, timeout, temporary service outage)
	if strings.Contains(errStr, "429") ||
		strings.Contains(errStr, "rate limit") ||
		strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "i/o timeout") ||
		strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "temporary failure") ||
		strings.Contains(errStr, "service unavailable") ||
		strings.Contains(errStr, "503") ||
		strings.Contains(errStr, "context deadline") {
		return NewClassifiedError(ErrorClassTransient, "temporary error, should retry", err)
	}

	// Fatal errors
	if strings.Contains(errStr, "500") ||
		strings.Contains(errStr, "disk full") ||
		strings.Contains(errStr, "out of memory") ||
		strings.Contains(errStr, "fatal") {
		return NewClassifiedError(ErrorClassFatal, "fatal error, cannot recover", err)
	}

	// Unknown classification
	return NewClassifiedError(ErrorClassUnknown, "unknown error", err)
}

// ClassifyHTTPError categorizes HTTP errors by status code.
func ClassifyHTTPError(statusCode int, body string) *ClassifiedError {
	switch statusCode {
	case http.StatusBadRequest:
		return NewClassifiedError(ErrorClassBadRequest, fmt.Sprintf("bad request (400): %s", body), nil)
	case http.StatusUnauthorized:
		return NewClassifiedError(ErrorClassAuth, "unauthorized (401)", nil)
	case http.StatusForbidden:
		return NewClassifiedError(ErrorClassAuth, "forbidden (403)", nil)
	case http.StatusNotFound:
		return NewClassifiedError(ErrorClassNotFound, "not found (404)", nil)
	case http.StatusConflict:
		return NewClassifiedError(ErrorClassBadRequest, "conflict (409)", nil)
	case http.StatusTooManyRequests:
		return NewClassifiedError(ErrorClassTransient, "rate limited (429)", nil)
	case http.StatusInternalServerError:
		return NewClassifiedError(ErrorClassFatal, "internal server error (500)", nil)
	case http.StatusServiceUnavailable:
		return NewClassifiedError(ErrorClassTransient, "service unavailable (503)", nil)
	case http.StatusGatewayTimeout:
		return NewClassifiedError(ErrorClassTransient, "gateway timeout (504)", nil)
	default:
		if statusCode >= 500 {
			return NewClassifiedError(ErrorClassFatal, fmt.Sprintf("server error (%d)", statusCode), nil)
		}
		if statusCode >= 400 {
			return NewClassifiedError(ErrorClassBadRequest, fmt.Sprintf("client error (%d)", statusCode), nil)
		}
		if statusCode >= 300 {
			return NewClassifiedError(ErrorClassUnknown, fmt.Sprintf("redirect (%d)", statusCode), nil)
		}
		return NewClassifiedError(ErrorClassUnknown, fmt.Sprintf("unexpected status (%d)", statusCode), nil)
	}
}

// ShouldRetry returns true if error should be retried.
func ShouldRetry(err error) bool {
	ce := ClassifyError(err)
	if ce == nil {
		return false
	}
	return ce.IsRetryable()
}
