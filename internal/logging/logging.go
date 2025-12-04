// Package logging provides structured logging for OffGridFlow using Go's
// standard library slog package. It supports multiple output formats,
// log levels, and integrates with the application configuration.
//
// Features:
//   - Structured JSON logging for production
//   - Human-readable text logging for development
//   - Contextual logging with request IDs and user IDs
//   - Log level configuration via environment
//   - Sensitive data redaction
//
// Usage:
//
//	logger := logging.New(logging.Config{
//	    Level:  slog.LevelInfo,
//	    Format: logging.FormatJSON,
//	})
//
//	logger.Info("server starting", slog.Int("port", 8090))
//
//	// With context
//	ctx := logging.WithRequestID(ctx, "req-123")
//	logging.FromContext(ctx).Info("handling request")
package logging

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

// =============================================================================
// Log Format Constants
// =============================================================================

// Format specifies the log output format.
type Format string

const (
	// FormatJSON outputs structured JSON logs, ideal for production and log aggregation.
	FormatJSON Format = "json"

	// FormatText outputs human-readable text logs, ideal for development.
	FormatText Format = "text"
)

// =============================================================================
// Context Keys
// =============================================================================

type contextKey string

const (
	// loggerKey is the context key for storing the logger.
	loggerKey contextKey = "offgridflow_logger"

	// requestIDKey is the context key for request correlation IDs.
	requestIDKey contextKey = "offgridflow_request_id"

	// userIDKey is the context key for authenticated user IDs.
	userIDKey contextKey = "offgridflow_user_id"

	// traceIDKey is the context key for distributed trace IDs.
	traceIDKey contextKey = "offgridflow_trace_id"
)

// =============================================================================
// Configuration
// =============================================================================

// Config holds logger configuration.
type Config struct {
	// Level is the minimum log level to output.
	// Defaults to slog.LevelInfo if zero.
	Level slog.Level

	// Format specifies the output format (json or text).
	// Defaults to FormatJSON if empty.
	Format Format

	// Output is the destination for log output.
	// Defaults to os.Stdout if nil.
	Output io.Writer

	// AddSource includes source file and line number in log output.
	// Recommended for development, may add overhead in production.
	AddSource bool

	// TimeFormat specifies the time format for text output.
	// Defaults to time.RFC3339 if empty. Ignored for JSON format.
	TimeFormat string

	// AppName is included in every log entry for multi-service environments.
	AppName string

	// Environment is included in every log entry (development, production, etc.).
	Environment string
}

// applyDefaults fills in default values for unset fields.
func (c *Config) applyDefaults() {
	if c.Format == "" {
		c.Format = FormatJSON
	}
	if c.Output == nil {
		c.Output = os.Stdout
	}
	if c.TimeFormat == "" {
		c.TimeFormat = time.RFC3339
	}
	if c.AppName == "" {
		c.AppName = "offgridflow"
	}
}

// =============================================================================
// Logger Construction
// =============================================================================

// New creates a new structured logger with the given configuration.
func New(cfg Config) *slog.Logger {
	cfg.applyDefaults()

	opts := &slog.HandlerOptions{
		Level:     cfg.Level,
		AddSource: cfg.AddSource,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Redact sensitive fields
			if isSensitiveKey(a.Key) {
				return slog.String(a.Key, "[REDACTED]")
			}

			// Format time consistently for text output
			if a.Key == slog.TimeKey && cfg.Format == FormatText {
				if t, ok := a.Value.Any().(time.Time); ok {
					return slog.String(a.Key, t.Format(cfg.TimeFormat))
				}
			}

			return a
		},
	}

	var handler slog.Handler
	switch cfg.Format {
	case FormatText:
		handler = slog.NewTextHandler(cfg.Output, opts)
	default:
		handler = slog.NewJSONHandler(cfg.Output, opts)
	}

	// Wrap with default attributes
	if cfg.AppName != "" || cfg.Environment != "" {
		attrs := make([]slog.Attr, 0, 2)
		if cfg.AppName != "" {
			attrs = append(attrs, slog.String("app", cfg.AppName))
		}
		if cfg.Environment != "" {
			attrs = append(attrs, slog.String("env", cfg.Environment))
		}
		handler = handler.WithAttrs(attrs)
	}

	return slog.New(handler)
}

// NewFromEnv creates a logger configured from environment variables.
//
// Environment variables:
//   - OFFGRIDFLOW_LOG_LEVEL: debug, info, warn, error (default: info)
//   - OFFGRIDFLOW_LOG_FORMAT: json, text (default: json)
//   - OFFGRIDFLOW_LOG_SOURCE: true, false (default: false)
func NewFromEnv() *slog.Logger {
	return New(Config{
		Level:       parseLogLevel(os.Getenv("OFFGRIDFLOW_LOG_LEVEL")),
		Format:      parseLogFormat(os.Getenv("OFFGRIDFLOW_LOG_FORMAT")),
		AddSource:   parseBool(os.Getenv("OFFGRIDFLOW_LOG_SOURCE")),
		Environment: os.Getenv("OFFGRIDFLOW_APP_ENV"),
	})
}

// Default returns the default logger for the application.
// This creates a production-ready JSON logger.
func Default() *slog.Logger {
	return New(Config{
		Level:  slog.LevelInfo,
		Format: FormatJSON,
	})
}

// Development returns a development-friendly logger with text output and debug level.
func Development() *slog.Logger {
	return New(Config{
		Level:     slog.LevelDebug,
		Format:    FormatText,
		AddSource: true,
	})
}

// =============================================================================
// Context Integration
// =============================================================================

// NewContext returns a new context with the logger attached.
func NewContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// FromContext retrieves the logger from context.
// Returns the default logger if none is found.
func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerKey).(*slog.Logger); ok && logger != nil {
		return logger
	}
	return slog.Default()
}

// WithRequestID adds a request ID to the context and returns a logger with it attached.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	ctx = context.WithValue(ctx, requestIDKey, requestID)

	logger := FromContext(ctx).With(slog.String("request_id", requestID))
	return NewContext(ctx, logger)
}

// WithUserID adds a user ID to the context and returns a logger with it attached.
func WithUserID(ctx context.Context, userID string) context.Context {
	ctx = context.WithValue(ctx, userIDKey, userID)

	logger := FromContext(ctx).With(slog.String("user_id", userID))
	return NewContext(ctx, logger)
}

// WithTraceID adds a trace ID to the context for distributed tracing.
func WithTraceID(ctx context.Context, traceID string) context.Context {
	ctx = context.WithValue(ctx, traceIDKey, traceID)

	logger := FromContext(ctx).With(slog.String("trace_id", traceID))
	return NewContext(ctx, logger)
}

// RequestIDFromContext retrieves the request ID from context.
func RequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

// UserIDFromContext retrieves the user ID from context.
func UserIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(userIDKey).(string); ok {
		return id
	}
	return ""
}

// =============================================================================
// Error Logging Helpers
// =============================================================================

// Error logs an error with stack context.
// It includes the file and line number where the error occurred.
func Error(logger *slog.Logger, msg string, err error, attrs ...slog.Attr) {
	if logger == nil {
		logger = slog.Default()
	}

	// Get caller information
	_, file, line, ok := runtime.Caller(1)
	if ok {
		attrs = append(attrs,
			slog.String("error", err.Error()),
			slog.String("error_file", file),
			slog.Int("error_line", line),
		)
	} else {
		attrs = append(attrs, slog.String("error", err.Error()))
	}

	args := make([]any, 0, len(attrs))
	for _, attr := range attrs {
		args = append(args, attr)
	}

	logger.Error(msg, args...)
}

// ErrorContext logs an error using the logger from context.
func ErrorContext(ctx context.Context, msg string, err error, attrs ...slog.Attr) {
	Error(FromContext(ctx), msg, err, attrs...)
}

// =============================================================================
// Sensitive Data Handling
// =============================================================================

// sensitiveKeys lists field names that should be redacted.
var sensitiveKeys = map[string]bool{
	"password":       true,
	"passwd":         true,
	"secret":         true,
	"token":          true,
	"api_key":        true,
	"apikey":         true,
	"authorization":  true,
	"auth":           true,
	"credential":     true,
	"private_key":    true,
	"access_token":   true,
	"refresh_token":  true,
	"jwt":            true,
	"session":        true,
	"cookie":         true,
	"credit_card":    true,
	"card_number":    true,
	"cvv":            true,
	"ssn":            true,
	"stripe_key":     true,
	"webhook_secret": true,
}

// isSensitiveKey checks if a key name should have its value redacted.
func isSensitiveKey(key string) bool {
	return sensitiveKeys[strings.ToLower(key)]
}

// AddSensitiveKey adds a key to the list of sensitive keys that will be redacted.
func AddSensitiveKey(key string) {
	sensitiveKeys[strings.ToLower(key)] = true
}

// =============================================================================
// Helper Functions
// =============================================================================

// parseLogLevel parses a log level string to slog.Level.
func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// parseLogFormat parses a format string to Format.
func parseLogFormat(format string) Format {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "text", "console":
		return FormatText
	default:
		return FormatJSON
	}
}

// parseBool parses a boolean string.
func parseBool(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true", "1", "yes", "on":
		return true
	default:
		return false
	}
}

// =============================================================================
// HTTP Middleware Helper Types
// =============================================================================

// HTTPLogEntry represents a structured HTTP request/response log entry.
type HTTPLogEntry struct {
	Method       string        `json:"method"`
	Path         string        `json:"path"`
	StatusCode   int           `json:"status_code"`
	Duration     time.Duration `json:"duration_ns"`
	DurationMS   float64       `json:"duration_ms"`
	RequestID    string        `json:"request_id,omitempty"`
	UserID       string        `json:"user_id,omitempty"`
	RemoteAddr   string        `json:"remote_addr,omitempty"`
	UserAgent    string        `json:"user_agent,omitempty"`
	BytesWritten int64         `json:"bytes_written,omitempty"`
	Error        string        `json:"error,omitempty"`
}

// LogValue implements slog.LogValuer for structured logging.
func (e HTTPLogEntry) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("method", e.Method),
		slog.String("path", e.Path),
		slog.Int("status", e.StatusCode),
		slog.Float64("duration_ms", e.DurationMS),
	}

	if e.RequestID != "" {
		attrs = append(attrs, slog.String("request_id", e.RequestID))
	}
	if e.UserID != "" {
		attrs = append(attrs, slog.String("user_id", e.UserID))
	}
	if e.RemoteAddr != "" {
		attrs = append(attrs, slog.String("remote_addr", e.RemoteAddr))
	}
	if e.Error != "" {
		attrs = append(attrs, slog.String("error", e.Error))
	}

	return slog.GroupValue(attrs...)
}

// HTTPMiddleware returns an HTTP middleware that logs requests/responses.
func HTTPMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			lrw := &loggingResponseWriter{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(lrw, r)

			duration := time.Since(start)
			entry := HTTPLogEntry{
				Method:       r.Method,
				Path:         r.URL.Path,
				StatusCode:   lrw.status,
				Duration:     duration,
				DurationMS:   float64(duration) / float64(time.Millisecond),
				RequestID:    RequestIDFromContext(r.Context()),
				UserID:       UserIDFromContext(r.Context()),
				RemoteAddr:   r.RemoteAddr,
				UserAgent:    r.UserAgent(),
				BytesWritten: lrw.bytes,
			}

			logger.Info("http_request", slog.Any("http", entry))
		})
	}
}

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
	bytes  int64
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.status = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	n, err := lrw.ResponseWriter.Write(b)
	lrw.bytes += int64(n)
	return n, err
}

// =============================================================================
// Compatibility Layer
// =============================================================================

// LegacyLogger wraps slog.Logger with the standard log.Logger interface.
// This provides backward compatibility with code using the standard log package.
type LegacyLogger struct {
	logger *slog.Logger
}

// NewLegacy creates a LegacyLogger wrapping the given slog.Logger.
func NewLegacy(logger *slog.Logger) *LegacyLogger {
	return &LegacyLogger{logger: logger}
}

// Printf implements log.Logger.Printf.
func (l *LegacyLogger) Printf(format string, v ...any) {
	l.logger.Info(strings.TrimSpace(format), v...)
}

// Println implements log.Logger.Println.
func (l *LegacyLogger) Println(v ...any) {
	if len(v) > 0 {
		if msg, ok := v[0].(string); ok {
			l.logger.Info(msg, v[1:]...)
			return
		}
	}
	l.logger.Info("log", slog.Any("args", v))
}

// Fatalf logs an error and would normally exit (but doesn't to match slog behavior).
func (l *LegacyLogger) Fatalf(format string, v ...any) {
	l.logger.Error(strings.TrimSpace(format), v...)
}
