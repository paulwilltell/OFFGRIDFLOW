// Package ingestion provides observability instrumentation for ingestion flows.
// Uses OpenTelemetry (OTEL) for metrics and tracing.
package ingestion

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// ObservabilityConfig holds configuration for ingestion observability.
type ObservabilityConfig struct {
	// Tracer for distributed tracing
	Tracer trace.Tracer

	// Meter for metrics
	Meter metric.Meter

	// Logger for structured logging
	Logger *slog.Logger

	// EnableTracing enables trace spans
	EnableTracing bool

	// EnableMetrics enables metric recording
	EnableMetrics bool

	// ServiceName for OTEL attributes
	ServiceName string
}

// NewObservabilityConfig creates a default observability configuration.
func NewObservabilityConfig(serviceName string) *ObservabilityConfig {
	return &ObservabilityConfig{
		Tracer:        otel.Tracer(serviceName),
		Meter:         otel.Meter(serviceName),
		Logger:        slog.Default().With("service", serviceName),
		EnableTracing: true,
		EnableMetrics: true,
		ServiceName:   serviceName,
	}
}

// IngestionMetrics tracks metrics for ingestion operations.
type IngestionMetrics struct {
	// Counter for successful ingestions
	SuccessCount metric.Int64Counter

	// Counter for failed ingestions
	FailureCount metric.Int64Counter

	// Counter for total items ingested
	ItemCount metric.Int64Counter

	// Histogram for ingestion latency
	LatencyHistogram metric.Float64Histogram

	// Gauge for items per batch
	BatchSizeGauge metric.Int64Gauge

	// Meter for recording metrics
	meter metric.Meter
}

// NewIngestionMetrics creates metrics for ingestion tracking.
func NewIngestionMetrics(meter metric.Meter) (*IngestionMetrics, error) {
	if meter == nil {
		return nil, fmt.Errorf("observability: meter is required")
	}

	successCount, err := meter.Int64Counter("ingestion_success_total",
		metric.WithDescription("Total successful ingestion operations"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create success counter: %w", err)
	}

	failureCount, err := meter.Int64Counter("ingestion_failure_total",
		metric.WithDescription("Total failed ingestion operations"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create failure counter: %w", err)
	}

	itemCount, err := meter.Int64Counter("ingestion_items_total",
		metric.WithDescription("Total items ingested"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create item counter: %w", err)
	}

	latencyHist, err := meter.Float64Histogram("ingestion_latency_seconds",
		metric.WithDescription("Ingestion operation latency in seconds"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create latency histogram: %w", err)
	}

	// Note: Gauge creation may not be available in all OTEL versions
	// We'll create it if available, otherwise skip
	var batchGauge metric.Int64Gauge
	if g, err := meter.Int64Gauge("ingestion_batch_size"); err == nil {
		batchGauge = g
	}

	return &IngestionMetrics{
		SuccessCount:     successCount,
		FailureCount:     failureCount,
		ItemCount:        itemCount,
		LatencyHistogram: latencyHist,
		BatchSizeGauge:   batchGauge,
		meter:            meter,
	}, nil
}

// RecordSuccess records a successful ingestion.
func (m *IngestionMetrics) RecordSuccess(ctx context.Context, itemCount int, latency time.Duration, connectorType string) {
	if m == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("connector", connectorType),
	}

	m.SuccessCount.Add(ctx, 1, metric.WithAttributes(attrs...))
	m.ItemCount.Add(ctx, int64(itemCount), metric.WithAttributes(attrs...))
	m.LatencyHistogram.Record(ctx, latency.Seconds(), metric.WithAttributes(attrs...))
}

// RecordFailure records a failed ingestion.
func (m *IngestionMetrics) RecordFailure(ctx context.Context, connectorType, errorClass string) {
	if m == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("connector", connectorType),
		attribute.String("error_class", errorClass),
	}

	m.FailureCount.Add(ctx, 1, metric.WithAttributes(attrs...))
}

// InvocationTracer wraps a function with tracing and logging.
type InvocationTracer struct {
	tracer trace.Tracer
	logger *slog.Logger
	meter  metric.Meter
}

// NewInvocationTracer creates a tracer for function invocations.
func NewInvocationTracer(config *ObservabilityConfig) *InvocationTracer {
	if config == nil {
		config = NewObservabilityConfig("ingestion")
	}

	return &InvocationTracer{
		tracer: config.Tracer,
		logger: config.Logger,
		meter:  config.Meter,
	}
}

// TraceInvocation wraps a function call with tracing and logging.
//
// Example:
//
//	result, err := tracer.TraceInvocation(ctx, "aws.fetch_data", func(ctx context.Context) (any, error) {
//	    return fetchAWSData(ctx, config)
//	})
func (t *InvocationTracer) TraceInvocation(ctx context.Context, operationName string, fn func(context.Context) error, attrs ...attribute.KeyValue) error {
	ctx, span := t.tracer.Start(ctx, operationName)
	defer span.End()

	// Add attributes to span
	for _, attr := range attrs {
		span.SetAttributes(attr)
	}

	start := time.Now()

	// Execute function
	err := fn(ctx)

	// Record metrics
	latency := time.Since(start)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		t.logger.Error(operationName, "error", err, "latency_ms", latency.Milliseconds())
		return err
	}

	span.SetStatus(codes.Ok, "success")
	t.logger.Info(operationName, "latency_ms", latency.Milliseconds())
	return nil
}

// TraceInvocationWithResult wraps a function with tracing, logging, and result return.
//
// Example:
//
//	data, err := tracer.TraceInvocationWithResult(ctx, "aws.parse_data", func(ctx context.Context) (any, error) {
//	    return parseCSV(ctx, response)
//	})
func (t *InvocationTracer) TraceInvocationWithResult(ctx context.Context, operationName string, fn func(context.Context) (any, error), attrs ...attribute.KeyValue) (any, error) {
	ctx, span := t.tracer.Start(ctx, operationName)
	defer span.End()

	for _, attr := range attrs {
		span.SetAttributes(attr)
	}

	start := time.Now()

	result, err := fn(ctx)

	latency := time.Since(start)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		t.logger.Error(operationName, "error", err, "latency_ms", latency.Milliseconds())
		return nil, err
	}

	span.SetStatus(codes.Ok, "success")
	t.logger.Info(operationName, "latency_ms", latency.Milliseconds())
	return result, nil
}

// LogIngestionEvent logs a structured ingestion event.
func (t *InvocationTracer) LogIngestionEvent(ctx context.Context, eventType string, attrs ...slog.Attr) {
	t.logger.InfoContext(ctx, eventType, attrsToAny(attrs)...)
}

// LogIngestionError logs an ingestion error with classification.
func (t *InvocationTracer) LogIngestionError(ctx context.Context, err error, connectorType string, attrs ...slog.Attr) {
	ce := ClassifyError(err)
	errorClass := "unknown"
	if ce != nil {
		errorClass = string(ce.Class)
	}

	allAttrs := append(
		[]slog.Attr{
			slog.String("connector", connectorType),
			slog.String("error_class", errorClass),
			slog.String("error", err.Error()),
		},
		attrs...,
	)

	t.logger.ErrorContext(ctx, "ingestion error", attrsToAny(allAttrs)...)
}

func attrsToAny(attrs []slog.Attr) []any {
	if len(attrs) == 0 {
		return nil
	}
	args := make([]any, len(attrs))
	for i, attr := range attrs {
		args[i] = attr
	}
	return args
}
