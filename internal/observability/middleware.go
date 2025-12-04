package observability

import (
	"context"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

// Context key types to avoid collisions
type contextKey string

const (
	tenantIDKey contextKey = "tenant_id"
	userIDKey   contextKey = "user_id"
)

// HTTPMiddleware provides OpenTelemetry tracing and metrics for HTTP handlers
type HTTPMiddleware struct {
	tracer  trace.Tracer
	metrics *Metrics
}

// NewHTTPMiddleware creates a new HTTP middleware with tracing and metrics
func NewHTTPMiddleware(tracerName string, metrics *Metrics) *HTTPMiddleware {
	return &HTTPMiddleware{
		tracer:  otel.Tracer(tracerName),
		metrics: metrics,
	}
}

// Handler wraps an HTTP handler with tracing and metrics
func (m *HTTPMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Extract trace context from headers
		ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(r.Header))

		// Start span
		spanName := r.Method + " " + r.URL.Path
		ctx, span := m.tracer.Start(ctx, spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				semconv.HTTPMethod(r.Method),
				semconv.HTTPURL(r.URL.String()),
				semconv.HTTPRoute(r.URL.Path),
				semconv.HTTPScheme(r.URL.Scheme),
				semconv.HTTPUserAgent(r.UserAgent()),
				semconv.HTTPClientIP(r.RemoteAddr),
			),
		)
		defer span.End()

		// Track in-flight requests
		if m.metrics != nil {
			m.metrics.HTTPRequestsInFlight.Add(ctx, 1)
			defer m.metrics.HTTPRequestsInFlight.Add(ctx, -1)
		}

		// Wrap response writer to capture status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Record start time
		start := time.Now()

		// Call next handler
		next.ServeHTTP(rw, r.WithContext(ctx))

		// Record duration
		duration := time.Since(start)

		// Set span attributes
		span.SetAttributes(
			semconv.HTTPStatusCode(rw.statusCode),
			attribute.Int64("http.response_size", rw.bytesWritten),
		)

		// Set span status based on HTTP status code
		if rw.statusCode >= 400 {
			span.SetStatus(codes.Error, http.StatusText(rw.statusCode))
		} else {
			span.SetStatus(codes.Ok, "")
		}

		// Record metrics
		if m.metrics != nil {
			m.metrics.RecordHTTPRequest(ctx, r.Method, r.URL.Path, rw.statusCode, duration)
		}
	})
}

// responseWriter wraps http.ResponseWriter to capture status code and bytes written
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += int64(n)
	return n, err
}

// DBMiddleware provides database query tracing
type DBMiddleware struct {
	tracer  trace.Tracer
	metrics *Metrics
}

// NewDBMiddleware creates a new database middleware
func NewDBMiddleware(tracerName string, metrics *Metrics) *DBMiddleware {
	return &DBMiddleware{
		tracer:  otel.Tracer(tracerName),
		metrics: metrics,
	}
}

// TraceQuery traces a database query
func (m *DBMiddleware) TraceQuery(ctx context.Context, operation, table, query string, fn func() (int64, error)) error {
	ctx, span := m.tracer.Start(ctx, "db.query",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			semconv.DBSystemKey.String("postgresql"),
			semconv.DBOperationKey.String(operation),
			semconv.DBSQLTableKey.String(table),
			attribute.String("db.query", query),
		),
	)
	defer span.End()

	start := time.Now()

	rowsAffected, err := fn()
	duration := time.Since(start)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
		span.SetAttributes(attribute.Int64("db.rows_affected", rowsAffected))
	}

	if m.metrics != nil {
		m.metrics.RecordDBQuery(ctx, operation, table, duration)
	}

	return err
}

// TenantContext adds tenant information to context and span
func TenantContext(ctx context.Context, tenantID, userID string) context.Context {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetAttributes(
			attribute.String("tenant.id", tenantID),
			attribute.String("user.id", userID),
		)
	}

	// Store in context for downstream use
	ctx = context.WithValue(ctx, tenantIDKey, tenantID)
	ctx = context.WithValue(ctx, userIDKey, userID)

	return ctx
}

// GetTenantID retrieves tenant ID from context
func GetTenantID(ctx context.Context) string {
	if tenantID, ok := ctx.Value(tenantIDKey).(string); ok {
		return tenantID
	}
	return ""
}

// GetUserID retrieves user ID from context
func GetUserID(ctx context.Context) string {
	if userID, ok := ctx.Value(userIDKey).(string); ok {
		return userID
	}
	return ""
}

// StartSpan is a convenience function to start a span with common attributes
func StartSpan(ctx context.Context, tracerName, operationName string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	tracer := otel.Tracer(tracerName)
	return tracer.Start(ctx, operationName, trace.WithAttributes(attrs...))
}

// RecordError records an error on the current span
func RecordError(ctx context.Context, err error) {
	if err == nil {
		return
	}

	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}

// AddEvent adds an event to the current span
func AddEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.AddEvent(name, trace.WithAttributes(attrs...))
	}
}

// SetSpanAttributes sets attributes on the current span
func SetSpanAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetAttributes(attrs...)
	}
}
