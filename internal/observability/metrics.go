package observability

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// MetricsProvider manages OpenTelemetry metrics
type MetricsProvider struct {
	provider *sdkmetric.MeterProvider
	config   TracerConfig
}

// NewMetricsProvider creates and configures an OTLP metrics provider
func NewMetricsProvider(ctx context.Context, config TracerConfig) (*MetricsProvider, error) {
	// Create OTLP exporter
	exporter, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithEndpoint(config.OTLPEndpoint),
		otlpmetrichttp.WithInsecure(), // Use WithTLSCredentials in production
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP metrics exporter: %w", err)
	}

	// Create resource
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

	// Create meter provider
	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(30*time.Second))),
		sdkmetric.WithResource(res),
	)

	// Set global meter provider
	otel.SetMeterProvider(provider)

	return &MetricsProvider{
		provider: provider,
		config:   config,
	}, nil
}

// Shutdown gracefully shuts down the metrics provider
func (mp *MetricsProvider) Shutdown(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return mp.provider.Shutdown(ctx)
}

// Meter returns a named meter
func (mp *MetricsProvider) Meter(name string) metric.Meter {
	return mp.provider.Meter(name)
}

// Metrics holds all application metrics
type Metrics struct {
	meter metric.Meter

	// HTTP metrics
	HTTPRequestDuration   metric.Float64Histogram
	HTTPRequestCount      metric.Int64Counter
	HTTPRequestsInFlight  metric.Int64UpDownCounter

	// Database metrics
	DBQueryDuration       metric.Float64Histogram
	DBQueryCount          metric.Int64Counter
	DBConnectionsActive   metric.Int64UpDownCounter

	// Emissions metrics
	EmissionsCalculated   metric.Int64Counter
	EmissionsKgCO2e       metric.Float64Counter
	EmissionsRecordCount  metric.Int64Counter

	// Connector metrics
	ConnectorSyncDuration metric.Float64Histogram
	ConnectorSyncCount    metric.Int64Counter
	ConnectorRecordsFetched metric.Int64Counter
	ConnectorErrors       metric.Int64Counter

	// Report metrics
	ReportGenerationDuration metric.Float64Histogram
	ReportGenerationCount    metric.Int64Counter
	ReportSizeBytes          metric.Int64Histogram

	// Job metrics
	JobExecutionDuration  metric.Float64Histogram
	JobExecutionCount     metric.Int64Counter
	JobSuccessCount       metric.Int64Counter
	JobFailureCount       metric.Int64Counter
	JobRetryCount         metric.Int64Counter
	JobQueueDepth         metric.Int64UpDownCounter

	// Billing metrics
	BillingOperationCount metric.Int64Counter
	BillingAmountCents    metric.Int64Counter
	ActiveSubscriptions   metric.Int64UpDownCounter

	// Cache metrics
	CacheHits             metric.Int64Counter
	CacheMisses           metric.Int64Counter
	CacheEvictions        metric.Int64Counter

	// Auth metrics
	AuthAttempts          metric.Int64Counter
	AuthSuccesses         metric.Int64Counter
	AuthFailures          metric.Int64Counter
	ActiveSessions        metric.Int64UpDownCounter

	// Rate limit metrics
	RateLimitExceeded     metric.Int64Counter
}

// NewMetrics creates and registers all application metrics
func NewMetrics(meterName string) (*Metrics, error) {
	meter := otel.Meter(meterName)

	// HTTP metrics
	httpRequestDuration, err := meter.Float64Histogram(
		"http.request.duration",
		metric.WithDescription("HTTP request duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	httpRequestCount, err := meter.Int64Counter(
		"http.request.count",
		metric.WithDescription("Total HTTP requests"),
	)
	if err != nil {
		return nil, err
	}

	httpRequestsInFlight, err := meter.Int64UpDownCounter(
		"http.requests.inflight",
		metric.WithDescription("Current number of in-flight HTTP requests"),
	)
	if err != nil {
		return nil, err
	}

	// Database metrics
	dbQueryDuration, err := meter.Float64Histogram(
		"db.query.duration",
		metric.WithDescription("Database query duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	dbQueryCount, err := meter.Int64Counter(
		"db.query.count",
		metric.WithDescription("Total database queries"),
	)
	if err != nil {
		return nil, err
	}

	dbConnectionsActive, err := meter.Int64UpDownCounter(
		"db.connections.active",
		metric.WithDescription("Active database connections"),
	)
	if err != nil {
		return nil, err
	}

	// Emissions metrics
	emissionsCalculated, err := meter.Int64Counter(
		"emissions.calculated.count",
		metric.WithDescription("Total emissions calculations performed"),
	)
	if err != nil {
		return nil, err
	}

	emissionsKgCO2e, err := meter.Float64Counter(
		"emissions.kg_co2e.total",
		metric.WithDescription("Total emissions calculated in kg CO2e"),
		metric.WithUnit("kg"),
	)
	if err != nil {
		return nil, err
	}

	emissionsRecordCount, err := meter.Int64Counter(
		"emissions.records.processed",
		metric.WithDescription("Total emissions records processed"),
	)
	if err != nil {
		return nil, err
	}

	// Connector metrics
	connectorSyncDuration, err := meter.Float64Histogram(
		"connector.sync.duration",
		metric.WithDescription("Connector sync duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	connectorSyncCount, err := meter.Int64Counter(
		"connector.sync.count",
		metric.WithDescription("Total connector syncs"),
	)
	if err != nil {
		return nil, err
	}

	connectorRecordsFetched, err := meter.Int64Counter(
		"connector.records.fetched",
		metric.WithDescription("Total records fetched by connectors"),
	)
	if err != nil {
		return nil, err
	}

	connectorErrors, err := meter.Int64Counter(
		"connector.errors.count",
		metric.WithDescription("Total connector errors"),
	)
	if err != nil {
		return nil, err
	}

	// Report metrics
	reportGenerationDuration, err := meter.Float64Histogram(
		"report.generation.duration",
		metric.WithDescription("Report generation duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	reportGenerationCount, err := meter.Int64Counter(
		"report.generation.count",
		metric.WithDescription("Total reports generated"),
	)
	if err != nil {
		return nil, err
	}

	reportSizeBytes, err := meter.Int64Histogram(
		"report.size.bytes",
		metric.WithDescription("Report size in bytes"),
		metric.WithUnit("bytes"),
	)
	if err != nil {
		return nil, err
	}

	// Job metrics
	jobExecutionDuration, err := meter.Float64Histogram(
		"job.execution.duration",
		metric.WithDescription("Job execution duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	jobExecutionCount, err := meter.Int64Counter(
		"job.execution.count",
		metric.WithDescription("Total job executions"),
	)
	if err != nil {
		return nil, err
	}

	jobSuccessCount, err := meter.Int64Counter(
		"job.success.count",
		metric.WithDescription("Total successful jobs"),
	)
	if err != nil {
		return nil, err
	}

	jobFailureCount, err := meter.Int64Counter(
		"job.failure.count",
		metric.WithDescription("Total failed jobs"),
	)
	if err != nil {
		return nil, err
	}

	jobRetryCount, err := meter.Int64Counter(
		"job.retry.count",
		metric.WithDescription("Total job retries"),
	)
	if err != nil {
		return nil, err
	}

	jobQueueDepth, err := meter.Int64UpDownCounter(
		"job.queue.depth",
		metric.WithDescription("Current job queue depth"),
	)
	if err != nil {
		return nil, err
	}

	// Billing metrics
	billingOperationCount, err := meter.Int64Counter(
		"billing.operation.count",
		metric.WithDescription("Total billing operations"),
	)
	if err != nil {
		return nil, err
	}

	billingAmountCents, err := meter.Int64Counter(
		"billing.amount.cents",
		metric.WithDescription("Total billing amount in cents"),
		metric.WithUnit("cents"),
	)
	if err != nil {
		return nil, err
	}

	activeSubscriptions, err := meter.Int64UpDownCounter(
		"billing.subscriptions.active",
		metric.WithDescription("Current active subscriptions"),
	)
	if err != nil {
		return nil, err
	}

	// Cache metrics
	cacheHits, err := meter.Int64Counter(
		"cache.hits.count",
		metric.WithDescription("Total cache hits"),
	)
	if err != nil {
		return nil, err
	}

	cacheMisses, err := meter.Int64Counter(
		"cache.misses.count",
		metric.WithDescription("Total cache misses"),
	)
	if err != nil {
		return nil, err
	}

	cacheEvictions, err := meter.Int64Counter(
		"cache.evictions.count",
		metric.WithDescription("Total cache evictions"),
	)
	if err != nil {
		return nil, err
	}

	// Auth metrics
	authAttempts, err := meter.Int64Counter(
		"auth.attempts.count",
		metric.WithDescription("Total authentication attempts"),
	)
	if err != nil {
		return nil, err
	}

	authSuccesses, err := meter.Int64Counter(
		"auth.success.count",
		metric.WithDescription("Total successful authentications"),
	)
	if err != nil {
		return nil, err
	}

	authFailures, err := meter.Int64Counter(
		"auth.failure.count",
		metric.WithDescription("Total failed authentications"),
	)
	if err != nil {
		return nil, err
	}

	activeSessions, err := meter.Int64UpDownCounter(
		"auth.sessions.active",
		metric.WithDescription("Current active sessions"),
	)
	if err != nil {
		return nil, err
	}

	// Rate limit metrics
	rateLimitExceeded, err := meter.Int64Counter(
		"ratelimit.exceeded.count",
		metric.WithDescription("Total rate limit exceeded events"),
	)
	if err != nil {
		return nil, err
	}

	return &Metrics{
		meter:                     meter,
		HTTPRequestDuration:       httpRequestDuration,
		HTTPRequestCount:          httpRequestCount,
		HTTPRequestsInFlight:      httpRequestsInFlight,
		DBQueryDuration:           dbQueryDuration,
		DBQueryCount:              dbQueryCount,
		DBConnectionsActive:       dbConnectionsActive,
		EmissionsCalculated:       emissionsCalculated,
		EmissionsKgCO2e:           emissionsKgCO2e,
		EmissionsRecordCount:      emissionsRecordCount,
		ConnectorSyncDuration:     connectorSyncDuration,
		ConnectorSyncCount:        connectorSyncCount,
		ConnectorRecordsFetched:   connectorRecordsFetched,
		ConnectorErrors:           connectorErrors,
		ReportGenerationDuration:  reportGenerationDuration,
		ReportGenerationCount:     reportGenerationCount,
		ReportSizeBytes:           reportSizeBytes,
		JobExecutionDuration:      jobExecutionDuration,
		JobExecutionCount:         jobExecutionCount,
		JobSuccessCount:           jobSuccessCount,
		JobFailureCount:           jobFailureCount,
		JobRetryCount:             jobRetryCount,
		JobQueueDepth:             jobQueueDepth,
		BillingOperationCount:     billingOperationCount,
		BillingAmountCents:        billingAmountCents,
		ActiveSubscriptions:       activeSubscriptions,
		CacheHits:                 cacheHits,
		CacheMisses:               cacheMisses,
		CacheEvictions:            cacheEvictions,
		AuthAttempts:              authAttempts,
		AuthSuccesses:             authSuccesses,
		AuthFailures:              authFailures,
		ActiveSessions:            activeSessions,
		RateLimitExceeded:         rateLimitExceeded,
	}, nil
}

// RecordHTTPRequest records HTTP request metrics
func (m *Metrics) RecordHTTPRequest(ctx context.Context, method, route string, statusCode int, duration time.Duration) {
	attrs := []attribute.KeyValue{
		attribute.String("http.method", method),
		attribute.String("http.route", route),
		attribute.Int("http.status_code", statusCode),
	}

	m.HTTPRequestDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
	m.HTTPRequestCount.Add(ctx, 1, metric.WithAttributes(attrs...))
}

// RecordDBQuery records database query metrics
func (m *Metrics) RecordDBQuery(ctx context.Context, operation, table string, duration time.Duration) {
	attrs := []attribute.KeyValue{
		attribute.String("db.operation", operation),
		attribute.String("db.table", table),
	}

	m.DBQueryDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
	m.DBQueryCount.Add(ctx, 1, metric.WithAttributes(attrs...))
}

// RecordEmissionsCalculation records emissions calculation metrics
func (m *Metrics) RecordEmissionsCalculation(ctx context.Context, scope, methodology string, kgCO2e float64, recordCount int64) {
	attrs := []attribute.KeyValue{
		attribute.String("emissions.scope", scope),
		attribute.String("emissions.methodology", methodology),
	}

	m.EmissionsCalculated.Add(ctx, 1, metric.WithAttributes(attrs...))
	m.EmissionsKgCO2e.Add(ctx, kgCO2e, metric.WithAttributes(attrs...))
	m.EmissionsRecordCount.Add(ctx, recordCount, metric.WithAttributes(attrs...))
}

// RecordConnectorSync records connector sync metrics
func (m *Metrics) RecordConnectorSync(ctx context.Context, connectorType string, recordsFetched int64, duration time.Duration, success bool) {
	attrs := []attribute.KeyValue{
		attribute.String("connector.type", connectorType),
	}

	m.ConnectorSyncDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
	m.ConnectorSyncCount.Add(ctx, 1, metric.WithAttributes(attrs...))
	m.ConnectorRecordsFetched.Add(ctx, recordsFetched, metric.WithAttributes(attrs...))

	if !success {
		m.ConnectorErrors.Add(ctx, 1, metric.WithAttributes(attrs...))
	}
}

// RecordReportGeneration records report generation metrics
func (m *Metrics) RecordReportGeneration(ctx context.Context, reportType, format string, sizeBytes int64, duration time.Duration) {
	attrs := []attribute.KeyValue{
		attribute.String("report.type", reportType),
		attribute.String("report.format", format),
	}

	m.ReportGenerationDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
	m.ReportGenerationCount.Add(ctx, 1, metric.WithAttributes(attrs...))
	m.ReportSizeBytes.Record(ctx, sizeBytes, metric.WithAttributes(attrs...))
}

// RecordJobExecution records job execution metrics
func (m *Metrics) RecordJobExecution(ctx context.Context, jobType, status string, duration time.Duration) {
	attrs := []attribute.KeyValue{
		attribute.String("job.type", jobType),
		attribute.String("job.status", status),
	}

	m.JobExecutionDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
	m.JobExecutionCount.Add(ctx, 1, metric.WithAttributes(attrs...))

	switch status {
	case "completed":
		m.JobSuccessCount.Add(ctx, 1, metric.WithAttributes(attrs...))
	case "failed":
		m.JobFailureCount.Add(ctx, 1, metric.WithAttributes(attrs...))
	case "retrying":
		m.JobRetryCount.Add(ctx, 1, metric.WithAttributes(attrs...))
	}
}
