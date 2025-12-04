package observability

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

// TracerConfig holds tracer configuration
type TracerConfig struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	OTLPEndpoint   string
	SampleRate     float64
}

// TracerProvider manages OpenTelemetry tracing
type TracerProvider struct {
	provider *sdktrace.TracerProvider
	config   TracerConfig
}

// NewTracerProvider creates and configures an OTLP tracer provider
func NewTracerProvider(ctx context.Context, config TracerConfig) (*TracerProvider, error) {
	// Create OTLP exporter
	client := otlptracehttp.NewClient(
		otlptracehttp.WithEndpoint(config.OTLPEndpoint),
		otlptracehttp.WithInsecure(), // Use WithTLSCredentials in production
	)

	exporter, err := otlptrace.New(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceVersion(config.ServiceVersion),
			semconv.DeploymentEnvironment(config.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Configure sampler based on sample rate
	var sampler sdktrace.Sampler
	if config.SampleRate >= 1.0 {
		sampler = sdktrace.AlwaysSample()
	} else if config.SampleRate <= 0.0 {
		sampler = sdktrace.NeverSample()
	} else {
		sampler = sdktrace.TraceIDRatioBased(config.SampleRate)
	}

	// Create tracer provider
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	// Set global tracer provider
	otel.SetTracerProvider(provider)

	// Set global propagator for context propagation
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return &TracerProvider{
		provider: provider,
		config:   config,
	}, nil
}

// Shutdown gracefully shuts down the tracer provider
func (tp *TracerProvider) Shutdown(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return tp.provider.Shutdown(ctx)
}

// Tracer returns a named tracer
func (tp *TracerProvider) Tracer(name string) trace.Tracer {
	return tp.provider.Tracer(name)
}

// SpanHelper provides helper methods for common tracing patterns
type SpanHelper struct {
	tracer trace.Tracer
}

// NewSpanHelper creates a new span helper
func NewSpanHelper(tracerName string) *SpanHelper {
	return &SpanHelper{
		tracer: otel.Tracer(tracerName),
	}
}

// StartSpan starts a new span with common attributes
func (sh *SpanHelper) StartSpan(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	return sh.tracer.Start(ctx, name, trace.WithAttributes(attrs...))
}

// RecordError records an error on the current span
func (sh *SpanHelper) RecordError(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}

// RecordSuccess sets span status to OK
func (sh *SpanHelper) RecordSuccess(span trace.Span) {
	span.SetStatus(codes.Ok, "")
}

// AddEvent adds an event to the span
func (sh *SpanHelper) AddEvent(span trace.Span, name string, attrs ...attribute.KeyValue) {
	span.AddEvent(name, trace.WithAttributes(attrs...))
}

// TraceHTTPRequest adds HTTP request attributes to span
func TraceHTTPRequest(span trace.Span, method, url, userAgent string, statusCode int) {
	span.SetAttributes(
		semconv.HTTPMethod(method),
		semconv.HTTPURL(url),
		semconv.HTTPUserAgent(userAgent),
		semconv.HTTPStatusCode(statusCode),
	)
}

// TraceDBQuery adds database query attributes to span
func TraceDBQuery(span trace.Span, operation, table, query string, rowsAffected int64) {
	span.SetAttributes(
		semconv.DBSystemKey.String("postgresql"),
		semconv.DBOperationKey.String(operation),
		semconv.DBSQLTableKey.String(table),
		attribute.String("db.query", query),
		attribute.Int64("db.rows_affected", rowsAffected),
	)
}

// TraceTenantOperation adds tenant context to span
func TraceTenantOperation(span trace.Span, tenantID, userID, operation string) {
	span.SetAttributes(
		attribute.String("tenant.id", tenantID),
		attribute.String("user.id", userID),
		attribute.String("operation", operation),
	)
}

// TraceEmissionsCalculation adds emissions calculation attributes
func TraceEmissionsCalculation(span trace.Span, scope, methodology string, emissionsKg float64, recordCount int) {
	span.SetAttributes(
		attribute.String("emissions.scope", scope),
		attribute.String("emissions.methodology", methodology),
		attribute.Float64("emissions.kg_co2e", emissionsKg),
		attribute.Int("emissions.record_count", recordCount),
	)
}

// TraceConnectorSync adds connector sync attributes
func TraceConnectorSync(span trace.Span, connectorType, connectorID string, recordsFetched, recordsProcessed int) {
	span.SetAttributes(
		attribute.String("connector.type", connectorType),
		attribute.String("connector.id", connectorID),
		attribute.Int("connector.records_fetched", recordsFetched),
		attribute.Int("connector.records_processed", recordsProcessed),
	)
}

// TraceReportGeneration adds report generation attributes
func TraceReportGeneration(span trace.Span, reportType, format string, sizeBytes int64, duration time.Duration) {
	span.SetAttributes(
		attribute.String("report.type", reportType),
		attribute.String("report.format", format),
		attribute.Int64("report.size_bytes", sizeBytes),
		attribute.Int64("report.generation_ms", duration.Milliseconds()),
	)
}

// TraceBillingOperation adds billing operation attributes
func TraceBillingOperation(span trace.Span, operation, customerID, subscriptionID string, amountCents int64) {
	span.SetAttributes(
		attribute.String("billing.operation", operation),
		attribute.String("billing.customer_id", customerID),
		attribute.String("billing.subscription_id", subscriptionID),
		attribute.Int64("billing.amount_cents", amountCents),
	)
}

// TraceJobExecution adds background job attributes
func TraceJobExecution(span trace.Span, jobType, jobID, status string, attempts int, duration time.Duration) {
	span.SetAttributes(
		attribute.String("job.type", jobType),
		attribute.String("job.id", jobID),
		attribute.String("job.status", status),
		attribute.Int("job.attempts", attempts),
		attribute.Int64("job.duration_ms", duration.Milliseconds()),
	)
}

// Common attribute keys
var (
	AttrTenantID       = attribute.Key("tenant.id")
	AttrUserID         = attribute.Key("user.id")
	AttrOperation      = attribute.Key("operation")
	AttrEmissionsScope = attribute.Key("emissions.scope")
	AttrConnectorType  = attribute.Key("connector.type")
	AttrReportType     = attribute.Key("report.type")
	AttrJobType        = attribute.Key("job.type")
	AttrCacheHit       = attribute.Key("cache.hit")
	AttrErrorType      = attribute.Key("error.type")
	AttrResponseTimeMs = attribute.Key("response_time_ms")
)
