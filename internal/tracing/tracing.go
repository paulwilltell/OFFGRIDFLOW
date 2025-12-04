// Package tracing provides OpenTelemetry tracing instrumentation for OffGridFlow.
//
// This package sets up distributed tracing to track requests across the application,
// including AI operations, database queries, and HTTP handlers.
package tracing

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

// Config holds configuration for tracing setup.
type Config struct {
	// ServiceName identifies the application in traces
	ServiceName string

	// ServiceVersion is the application version
	ServiceVersion string

	// Environment (development, staging, production)
	Environment string

	// OTLPEndpoint is the OTLP collector endpoint
	// Defaults to http://localhost:4318 (AI Toolkit trace collector)
	OTLPEndpoint string

	// SamplingRate controls trace sampling (0.0 to 1.0)
	// 1.0 = sample all traces, 0.5 = sample 50%, etc.
	// Defaults to 1.0
	SamplingRate float64

	// Enabled controls whether tracing is active
	Enabled bool

	// Logger for tracing operations
	Logger *slog.Logger
}

// Provider wraps the OpenTelemetry trace provider with shutdown capability.
type Provider struct {
	provider *sdktrace.TracerProvider
	logger   *slog.Logger
}

// Shutdown gracefully shuts down the trace provider, flushing any pending spans.
func (p *Provider) Shutdown(ctx context.Context) error {
	if p.provider == nil {
		return nil
	}

	p.logger.Info("shutting down trace provider")
	
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := p.provider.Shutdown(shutdownCtx); err != nil {
		p.logger.Error("failed to shutdown trace provider", "error", err)
		return fmt.Errorf("tracing: shutdown failed: %w", err)
	}

	p.logger.Info("trace provider shutdown complete")
	return nil
}

// Setup initializes OpenTelemetry tracing with the provided configuration.
//
// It configures:
//   - OTLP HTTP exporter for sending traces to AI Toolkit
//   - Resource attributes identifying the service
//   - Sampling strategy
//   - Global trace provider and propagators
//
// Returns a Provider that must be shut down when the application exits.
func Setup(cfg Config) (*Provider, error) {
	if !cfg.Enabled {
		return &Provider{logger: cfg.Logger}, nil
	}

	// Apply defaults
	if cfg.ServiceName == "" {
		cfg.ServiceName = "offgridflow"
	}
	if cfg.ServiceVersion == "" {
		cfg.ServiceVersion = "dev"
	}
	if cfg.Environment == "" {
		cfg.Environment = "development"
	}
	if cfg.OTLPEndpoint == "" {
		cfg.OTLPEndpoint = "http://localhost:4318" // AI Toolkit default
	}
	if cfg.SamplingRate <= 0 || cfg.SamplingRate > 1.0 {
		cfg.SamplingRate = 1.0
	}
	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}

	logger.Info("initializing tracing",
		"service", cfg.ServiceName,
		"version", cfg.ServiceVersion,
		"environment", cfg.Environment,
		"endpoint", cfg.OTLPEndpoint,
		"sampling_rate", cfg.SamplingRate,
	)

	// Create OTLP HTTP exporter
	exporter, err := otlptrace.New(
		context.Background(),
		otlptracehttp.NewClient(
			otlptracehttp.WithEndpoint(stripScheme(cfg.OTLPEndpoint)),
			otlptracehttp.WithInsecure(), // Local development
		),
	)
	if err != nil {
		return nil, fmt.Errorf("tracing: failed to create OTLP exporter: %w", err)
	}

	// Create resource with service information
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			semconv.DeploymentEnvironment(cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("tracing: failed to create resource: %w", err)
	}

	// Create trace provider with sampling
	var sampler sdktrace.Sampler
	if cfg.SamplingRate >= 1.0 {
		sampler = sdktrace.AlwaysSample()
	} else if cfg.SamplingRate <= 0 {
		sampler = sdktrace.NeverSample()
	} else {
		sampler = sdktrace.TraceIDRatioBased(cfg.SamplingRate)
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	// Set global providers
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	logger.Info("tracing initialized successfully")

	return &Provider{
		provider: provider,
		logger:   logger,
	}, nil
}

// stripScheme removes http:// or https:// prefix from endpoint URL
func stripScheme(endpoint string) string {
	if len(endpoint) > 7 && endpoint[:7] == "http://" {
		return endpoint[7:]
	}
	if len(endpoint) > 8 && endpoint[:8] == "https://" {
		return endpoint[8:]
	}
	return endpoint
}

// StartSpan is a convenience function to start a new span.
//
// Example:
//
//	ctx, span := tracing.StartSpan(ctx, "operation-name")
//	defer span.End()
func StartSpan(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	tracer := otel.Tracer("offgridflow")
	return tracer.Start(ctx, spanName, opts...)
}

// RecordError records an error on the span and sets its status.
//
// Example:
//
//	if err != nil {
//	    tracing.RecordError(span, err, "failed to process request")
//	}
func RecordError(span trace.Span, err error, description string) {
	if span == nil || err == nil {
		return
	}
	span.RecordError(err)
	span.SetStatus(codes.Error, description)
}

// SetAttributes is a convenience function to set multiple attributes on a span.
//
// Example:
//
//	tracing.SetAttributes(span, map[string]interface{}{
//	    "user.id": userID,
//	    "tenant.id": tenantID,
//	})
func SetAttributes(span trace.Span, attrs map[string]interface{}) {
	if span == nil {
		return
	}

	kvs := make([]attribute.KeyValue, 0, len(attrs))
	for k, v := range attrs {
		switch val := v.(type) {
		case string:
			kvs = append(kvs, attribute.String(k, val))
		case int:
			kvs = append(kvs, attribute.Int(k, val))
		case int64:
			kvs = append(kvs, attribute.Int64(k, val))
		case float64:
			kvs = append(kvs, attribute.Float64(k, val))
		case bool:
			kvs = append(kvs, attribute.Bool(k, val))
		default:
			kvs = append(kvs, attribute.String(k, fmt.Sprintf("%v", val)))
		}
	}
	span.SetAttributes(kvs...)
}

// AddEvent adds an event to the span with optional attributes.
//
// Example:
//
//	tracing.AddEvent(span, "cache.hit", map[string]interface{}{
//	    "cache.key": key,
//	})
func AddEvent(span trace.Span, name string, attrs map[string]interface{}) {
	if span == nil {
		return
	}

	kvs := make([]attribute.KeyValue, 0, len(attrs))
	for k, v := range attrs {
		switch val := v.(type) {
		case string:
			kvs = append(kvs, attribute.String(k, val))
		case int:
			kvs = append(kvs, attribute.Int(k, val))
		case int64:
			kvs = append(kvs, attribute.Int64(k, val))
		case float64:
			kvs = append(kvs, attribute.Float64(k, val))
		case bool:
			kvs = append(kvs, attribute.Bool(k, val))
		default:
			kvs = append(kvs, attribute.String(k, fmt.Sprintf("%v", val)))
		}
	}

	span.AddEvent(name, trace.WithAttributes(kvs...))
}
