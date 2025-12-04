package observability

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
)

// Context key for request ID
type requestIDKey struct{}

// LoggingMiddleware provides structured logging with request IDs for HTTP handlers
type LoggingMiddleware struct {
	logger *slog.Logger
}

// NewLoggingMiddleware creates a new logging middleware
func NewLoggingMiddleware(logger *slog.Logger) *LoggingMiddleware {
	if logger == nil {
		logger = slog.Default()
	}
	return &LoggingMiddleware{
		logger: logger,
	}
}

// Handler wraps an HTTP handler with logging
func (m *LoggingMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate or extract request ID
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Add request ID to response headers
		w.Header().Set("X-Request-ID", requestID)

		// Create context with request ID
		ctx := context.WithValue(r.Context(), requestIDKey{}, requestID)

		// Extract tenant/user information if available
		tenantID := GetTenantID(ctx)
		userID := GetUserID(ctx)

		// Create logger with request context
		logger := m.logger.With(
			slog.String("request_id", requestID),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("remote_addr", r.RemoteAddr),
			slog.String("user_agent", r.UserAgent()),
		)

		if tenantID != "" {
			logger = logger.With(slog.String("tenant_id", tenantID))
		}
		if userID != "" {
			logger = logger.With(slog.String("user_id", userID))
		}

		// Log request start
		logger.Info("request started")

		// Wrap response writer to capture status code
		rw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Record start time
		start := time.Now()

		// Call next handler
		next.ServeHTTP(rw, r.WithContext(ctx))

		// Record duration
		duration := time.Since(start)

		// Log request completion
		logLevel := slog.LevelInfo
		if rw.statusCode >= 500 {
			logLevel = slog.LevelError
		} else if rw.statusCode >= 400 {
			logLevel = slog.LevelWarn
		}

		logger.Log(r.Context(), logLevel, "request completed",
			slog.Int("status_code", rw.statusCode),
			slog.Int64("bytes_written", rw.bytesWritten),
			slog.Duration("latency", duration),
			slog.Int64("latency_ms", duration.Milliseconds()),
		)
	})
}

// loggingResponseWriter wraps http.ResponseWriter to capture status code and bytes written
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
}

func (rw *loggingResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *loggingResponseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += int64(n)
	return n, err
}

// GetRequestID retrieves the request ID from context
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(requestIDKey{}).(string); ok {
		return requestID
	}
	return ""
}

// LoggerFromContext returns a logger with request context
func LoggerFromContext(ctx context.Context, base *slog.Logger) *slog.Logger {
	if base == nil {
		base = slog.Default()
	}

	logger := base

	// Add request ID if available
	if requestID := GetRequestID(ctx); requestID != "" {
		logger = logger.With(slog.String("request_id", requestID))
	}

	// Add tenant/user information if available
	if tenantID := GetTenantID(ctx); tenantID != "" {
		logger = logger.With(slog.String("tenant_id", tenantID))
	}
	if userID := GetUserID(ctx); userID != "" {
		logger = logger.With(slog.String("user_id", userID))
	}

	return logger
}

// NewStructuredLogger creates a structured logger with JSON output
func NewStructuredLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Rename timestamp field for better compatibility
			if a.Key == slog.TimeKey {
				a.Key = "timestamp"
			}
			// Rename level field
			if a.Key == slog.LevelKey {
				a.Key = "level"
			}
			// Rename message field
			if a.Key == slog.MessageKey {
				a.Key = "message"
			}
			return a
		},
	}))
}
