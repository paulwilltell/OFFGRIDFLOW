// Package responders provides standardized HTTP response utilities for the API layer.
//
// This package ensures consistent JSON response formatting across all API handlers,
// including structured error responses, pagination support, and cache control headers.
package responders

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// -----------------------------------------------------------------------------
// Error Types and Structures
// -----------------------------------------------------------------------------

// ErrorDetail standardizes error payloads returned by the API.
// All API errors MUST use this structure for client consistency.
type ErrorDetail struct {
	Code      string            `json:"code"`                // Machine-readable error code (e.g., "validation_error")
	Message   string            `json:"message"`             // Human-readable error message
	Detail    string            `json:"detail,omitempty"`    // Optional additional context
	Status    int               `json:"status"`              // HTTP status code
	RequestID string            `json:"requestId,omitempty"` // Correlation ID for debugging
	Timestamp string            `json:"timestamp,omitempty"` // ISO 8601 timestamp
	Fields    map[string]string `json:"fields,omitempty"`    // Field-level validation errors
}

// ErrorResponse is the envelope for JSON errors.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ValidationError represents a field-level validation failure.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// -----------------------------------------------------------------------------
// Pagination Support
// -----------------------------------------------------------------------------

// PageInfo contains pagination metadata for list responses.
type PageInfo struct {
	Page       int  `json:"page"`       // Current page (1-indexed)
	PerPage    int  `json:"perPage"`    // Items per page
	Total      int  `json:"total"`      // Total number of items
	TotalPages int  `json:"totalPages"` // Total number of pages
	HasNext    bool `json:"hasNext"`    // Whether there are more pages
	HasPrev    bool `json:"hasPrev"`    // Whether there are previous pages
}

// PagedResponse wraps a list response with pagination info.
type PagedResponse[T any] struct {
	Data     []T      `json:"data"`
	PageInfo PageInfo `json:"pageInfo"`
}

// NewPageInfo creates pagination metadata from parameters.
func NewPageInfo(page, perPage, total int) PageInfo {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + perPage - 1) / perPage
	}

	return PageInfo{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

// -----------------------------------------------------------------------------
// Response Writers
// -----------------------------------------------------------------------------

// JSON writes a JSON response with the provided status code.
// Content-Type header is automatically set to application/json.
func JSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if payload == nil {
		return
	}

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		// Log encoding error - response headers already sent, so we can't change status
		// In production, this should use structured logging
		_ = err
	}
}

// JSONPretty writes a pretty-printed JSON response (useful for debugging endpoints).
func JSONPretty(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if payload == nil {
		return
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(payload)
}

// NoContent writes a 204 No Content response.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Created writes a 201 Created response with the created resource.
func Created(w http.ResponseWriter, payload any) {
	JSON(w, http.StatusCreated, payload)
}

// CreatedWithLocation writes a 201 Created response with a Location header.
func CreatedWithLocation(w http.ResponseWriter, location string, payload any) {
	w.Header().Set("Location", location)
	JSON(w, http.StatusCreated, payload)
}

// Accepted writes a 202 Accepted response for async operations.
func Accepted(w http.ResponseWriter, payload any) {
	JSON(w, http.StatusAccepted, payload)
}

// -----------------------------------------------------------------------------
// Paginated Response Writer
// -----------------------------------------------------------------------------

// Paged writes a paginated JSON response with metadata.
func Paged[T any](w http.ResponseWriter, status int, data []T, pageInfo PageInfo) {
	response := PagedResponse[T]{
		Data:     data,
		PageInfo: pageInfo,
	}
	JSON(w, status, response)
}

// -----------------------------------------------------------------------------
// Error Response Writers
// -----------------------------------------------------------------------------

// Error writes a standardized JSON error response.
// Optional detail parameter provides additional context for debugging.
func Error(w http.ResponseWriter, status int, code, message string, detail ...string) {
	resp := ErrorResponse{
		Error: ErrorDetail{
			Code:      code,
			Message:   message,
			Status:    status,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	}

	if len(detail) > 0 && detail[0] != "" {
		resp.Error.Detail = detail[0]
	}

	JSON(w, status, resp)
}

// ErrorWithRequestID writes an error response including a request ID for correlation.
func ErrorWithRequestID(w http.ResponseWriter, status int, code, message, requestID string) {
	resp := ErrorResponse{
		Error: ErrorDetail{
			Code:      code,
			Message:   message,
			Status:    status,
			RequestID: requestID,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	}
	JSON(w, status, resp)
}

// ValidationErrors writes a 400 Bad Request with field-level validation errors.
func ValidationErrors(w http.ResponseWriter, errors []ValidationError) {
	fields := make(map[string]string, len(errors))
	for _, e := range errors {
		fields[e.Field] = e.Message
	}

	resp := ErrorResponse{
		Error: ErrorDetail{
			Code:      "validation_error",
			Message:   "one or more fields failed validation",
			Status:    http.StatusBadRequest,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Fields:    fields,
		},
	}
	JSON(w, http.StatusBadRequest, resp)
}

// ValidationErrorMap writes a 400 Bad Request with a map of field errors.
func ValidationErrorMap(w http.ResponseWriter, fields map[string]string) {
	resp := ErrorResponse{
		Error: ErrorDetail{
			Code:      "validation_error",
			Message:   "one or more fields failed validation",
			Status:    http.StatusBadRequest,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Fields:    fields,
		},
	}
	JSON(w, http.StatusBadRequest, resp)
}

// -----------------------------------------------------------------------------
// Common Error Shortcuts
// -----------------------------------------------------------------------------

// BadRequest writes a 400 Bad Request error.
func BadRequest(w http.ResponseWriter, code, message string) {
	Error(w, http.StatusBadRequest, code, message)
}

// Unauthorized writes a 401 Unauthorized error.
func Unauthorized(w http.ResponseWriter, code, message string) {
	Error(w, http.StatusUnauthorized, code, message)
}

// Forbidden writes a 403 Forbidden error.
func Forbidden(w http.ResponseWriter, code, message string) {
	Error(w, http.StatusForbidden, code, message)
}

// NotFound writes a 404 Not Found error.
func NotFound(w http.ResponseWriter, resource string) {
	Error(w, http.StatusNotFound, "not_found", fmt.Sprintf("%s not found", resource))
}

// MethodNotAllowed writes a 405 Method Not Allowed error.
func MethodNotAllowed(w http.ResponseWriter, allowed ...string) {
	if len(allowed) > 0 {
		w.Header().Set("Allow", joinStrings(allowed, ", "))
	}
	Error(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
}

// Conflict writes a 409 Conflict error.
func Conflict(w http.ResponseWriter, code, message string) {
	Error(w, http.StatusConflict, code, message)
}

// PaymentRequired writes a 402 Payment Required error.
func PaymentRequired(w http.ResponseWriter, message string) {
	Error(w, http.StatusPaymentRequired, "payment_required", message)
}

// RateLimited writes a 429 Too Many Requests error with Retry-After header.
func RateLimited(w http.ResponseWriter, retryAfterSeconds int) {
	w.Header().Set("Retry-After", strconv.Itoa(retryAfterSeconds))
	Error(w, http.StatusTooManyRequests, "rate_limited", "too many requests, please slow down")
}

// TooManyRequests writes a 429 Too Many Requests error with a custom message.
func TooManyRequests(w http.ResponseWriter, message string) {
	Error(w, http.StatusTooManyRequests, "limit_exceeded", message)
}

// InternalError writes a 500 Internal Server Error.
// The detail message is hidden from the client for security.
func InternalError(w http.ResponseWriter, internalDetail string) {
	// Log the internal detail but don't expose to client
	_ = internalDetail
	Error(w, http.StatusInternalServerError, "internal_error", "an unexpected error occurred")
}

// ServiceUnavailable writes a 503 Service Unavailable error.
func ServiceUnavailable(w http.ResponseWriter, message string, retryAfterSeconds int) {
	if retryAfterSeconds > 0 {
		w.Header().Set("Retry-After", strconv.Itoa(retryAfterSeconds))
	}
	Error(w, http.StatusServiceUnavailable, "service_unavailable", message)
}

// -----------------------------------------------------------------------------
// Cache Control Headers
// -----------------------------------------------------------------------------

// SetCacheControl sets cache control headers on the response.
func SetCacheControl(w http.ResponseWriter, maxAge time.Duration, public bool) {
	directive := "private"
	if public {
		directive = "public"
	}
	w.Header().Set("Cache-Control", fmt.Sprintf("%s, max-age=%d", directive, int(maxAge.Seconds())))
}

// SetNoCache sets headers to prevent caching.
func SetNoCache(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
}

// SetETag sets an ETag header for conditional requests.
func SetETag(w http.ResponseWriter, etag string) {
	w.Header().Set("ETag", fmt.Sprintf(`"%s"`, etag))
}

// -----------------------------------------------------------------------------
// CORS Headers (for preflight support)
// -----------------------------------------------------------------------------

// SetCORSHeaders sets CORS headers for cross-origin requests.
func SetCORSHeaders(w http.ResponseWriter, origin string, methods []string) {
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", joinStrings(methods, ", "))
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key, X-Request-ID")
	w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours
}

// -----------------------------------------------------------------------------
// Request ID Helpers
// -----------------------------------------------------------------------------

// GetRequestID extracts or generates a request ID from the request.
func GetRequestID(r *http.Request) string {
	if id := r.Header.Get("X-Request-ID"); id != "" {
		return id
	}
	if id := r.Header.Get("X-Correlation-ID"); id != "" {
		return id
	}
	return ""
}

// SetRequestID sets the request ID on the response for client correlation.
func SetRequestID(w http.ResponseWriter, requestID string) {
	if requestID != "" {
		w.Header().Set("X-Request-ID", requestID)
	}
}

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

// joinStrings joins a slice of strings with a separator.
func joinStrings(items []string, sep string) string {
	if len(items) == 0 {
		return ""
	}
	if len(items) == 1 {
		return items[0]
	}

	result := items[0]
	for _, item := range items[1:] {
		result += sep + item
	}
	return result
}

// ParsePagination extracts page and perPage from query parameters.
func ParsePagination(r *http.Request, defaultPerPage int) (page, perPage int) {
	page = 1
	perPage = defaultPerPage
	if perPage <= 0 {
		perPage = 20
	}

	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if pp := r.URL.Query().Get("per_page"); pp != "" {
		if parsed, err := strconv.Atoi(pp); err == nil && parsed > 0 && parsed <= 100 {
			perPage = parsed
		}
	}

	return page, perPage
}
