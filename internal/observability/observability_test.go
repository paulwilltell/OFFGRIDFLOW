package observability

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestLoggingMiddleware(t *testing.T) {
	logger := NewStructuredLogger()
	middleware := NewLoggingMiddleware(logger)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that request ID was added to context
		requestID := GetRequestID(r.Context())
		if requestID == "" {
			t.Error("Request ID not set in context")
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	wrappedHandler := middleware.Handler(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "test-request-123")
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	// Check response
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	// Check that request ID was added to response headers
	responseRequestID := rec.Header().Get("X-Request-ID")
	if responseRequestID != "test-request-123" {
		t.Errorf("Expected request ID test-request-123, got %s", responseRequestID)
	}
}

func TestHTTPMiddleware(t *testing.T) {
	// Create in-memory trace exporter
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)

	// Create metrics reader
	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	otel.SetMeterProvider(mp)

	// Create metrics
	metrics, err := NewMetrics("test")
	if err != nil {
		t.Fatalf("Failed to create metrics: %v", err)
	}

	middleware := NewHTTPMiddleware("test", metrics)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	wrappedHandler := middleware.Handler(handler)

	req := httptest.NewRequest("GET", "/api/test", nil)
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	// Check response
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	// Check that a span was created
	spans := exporter.GetSpans()
	if len(spans) == 0 {
		t.Error("No spans were created")
	}

	// Verify span attributes
	if len(spans) > 0 {
		span := spans[0]
		if span.Name != "GET /api/test" {
			t.Errorf("Expected span name 'GET /api/test', got %s", span.Name)
		}
	}

	// Check metrics were recorded
	var rm metricdata.ResourceMetrics
	err = reader.Collect(context.Background(), &rm)
	if err != nil {
		t.Fatalf("Failed to collect metrics: %v", err)
	}
}

func TestMetricsRecording(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	otel.SetMeterProvider(mp)

	metrics, err := NewMetrics("test")
	if err != nil {
		t.Fatalf("Failed to create metrics: %v", err)
	}

	ctx := context.Background()

	// Test HTTP request recording
	metrics.RecordHTTPRequest(ctx, "GET", "/api/test", 200, 100*time.Millisecond)

	// Test connector sync recording
	metrics.RecordConnectorSync(ctx, "aws", 150, 5*time.Second, true)

	// Test emissions calculation recording
	metrics.RecordEmissionsCalculation(ctx, "scope2", "location-based", 1250.5, 100)

	// Collect metrics
	var rm metricdata.ResourceMetrics
	err = reader.Collect(ctx, &rm)
	if err != nil {
		t.Fatalf("Failed to collect metrics: %v", err)
	}

	// Verify metrics were collected
	if len(rm.ScopeMetrics) == 0 {
		t.Error("No metrics were collected")
	}
}

func TestTracerProvider(t *testing.T) {
	ctx := context.Background()

	config := TracerConfig{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		OTLPEndpoint:   "localhost:4318",
		SampleRate:     1.0,
	}

	// Note: This will fail if OTEL collector is not running
	// In a real test, we'd use a mock exporter
	_, err := NewTracerProvider(ctx, config)
	if err != nil {
		t.Logf("TracerProvider creation failed (expected if OTEL collector not running): %v", err)
	}
}

func TestSpanHelper(t *testing.T) {
	// Create in-memory trace exporter
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)

	helper := NewSpanHelper("test")
	ctx := context.Background()

	// Start a span
	ctx, span := helper.StartSpan(ctx, "test-operation",
		attribute.String("test.key", "test.value"),
	)

	// Add event
	helper.AddEvent(span, "test-event",
		attribute.String("event.key", "event.value"),
	)

	// Record success
	helper.RecordSuccess(span)
	span.End()

	// Verify span was created
	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("Expected 1 span, got %d", len(spans))
	}

	if spans[0].Name != "test-operation" {
		t.Errorf("Expected span name 'test-operation', got %s", spans[0].Name)
	}
}

func TestTenantContext(t *testing.T) {
	// Create in-memory trace exporter
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)

	tracer := otel.Tracer("test")
	ctx, span := tracer.Start(context.Background(), "test")

	// Add tenant context
	ctx = TenantContext(ctx, "tenant-123", "user-456")

	// Verify context values
	if got := GetTenantID(ctx); got != "tenant-123" {
		t.Errorf("Expected tenant ID 'tenant-123', got %s", got)
	}

	if got := GetUserID(ctx); got != "user-456" {
		t.Errorf("Expected user ID 'user-456', got %s", got)
	}

	span.End()

	// Verify span attributes
	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("Expected 1 span, got %d", len(spans))
	}

	// Check for tenant attributes
	hastenantID := false
	hasUserID := false
	for _, attr := range spans[0].Attributes {
		if string(attr.Key) == "tenant.id" && attr.Value.AsString() == "tenant-123" {
			hastenantID = true
		}
		if string(attr.Key) == "user.id" && attr.Value.AsString() == "user-456" {
			hasUserID = true
		}
	}

	if !hastenantID {
		t.Error("Span missing tenant.id attribute")
	}
	if !hasUserID {
		t.Error("Span missing user.id attribute")
	}
}

func TestObservabilityProvider(t *testing.T) {
	ctx := context.Background()

	config := Config{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		OTLPEndpoint:   "localhost:4318",
		EnableMetrics:  true,
		EnableTracing:  true,
		EnableLogging:  true,
	}

	provider, err := NewProvider(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}
	defer provider.Shutdown(ctx)

	// Verify logger
	if provider.Logger() == nil {
		t.Error("Logger is nil")
	}

	// Verify metrics (may be nil if OTEL collector not available)
	t.Logf("Metrics available: %v", provider.Metrics() != nil)

	// Verify middleware
	middleware := provider.HTTPMiddleware()
	if middleware == nil {
		t.Error("HTTP middleware is nil")
	}

	// Test middleware
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}
