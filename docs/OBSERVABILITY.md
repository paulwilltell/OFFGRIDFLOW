# OffGridFlow Observability

Complete observability stack with logging, metrics, and distributed tracing.

## Architecture

```
┌─────────────┐      ┌──────────────────┐      ┌─────────────┐
│OffGridFlow │─────▶│ OTEL Collector   │─────▶│  Jaeger     │
│   API       │      │                  │      │  (Traces)   │
└─────────────┘      └──────────────────┘      └─────────────┘
                              │
                              │
                              ▼
                     ┌──────────────────┐
                     │   Prometheus     │
                     │   (Metrics)      │
                     └──────────────────┘
                              │
                              ▼
                     ┌──────────────────┐
                     │    Grafana       │
                     │  (Visualization) │
                     └──────────────────┘
```

## Components

### 1. **Logging**
- Structured JSON logging using Go's `log/slog`
- Request IDs for request correlation
- Context-aware logging with tenant/user information
- Log levels: DEBUG, INFO, WARN, ERROR

### 2. **Metrics** (Prometheus/OpenTelemetry)
- **HTTP Metrics**:
  - `http_request_count` - Total HTTP requests by method, route, status
  - `http_request_duration` - Request latency histogram
  - `http_requests_inflight` - Current active requests

- **Database Metrics**:
  - `db_query_count` - Total database queries
  - `db_query_duration` - Query latency
  - `db_connections_active` - Active connections

- **Emissions Metrics**:
  - `emissions_calculated_count` - Emissions calculations performed
  - `emissions_kg_co2e_total` - Total emissions in kg CO2e
  - `emissions_records_processed` - Records processed

- **Connector/Ingestion Metrics**:
  - `connector_sync_count` - Connector syncs by provider
  - `connector_sync_duration` - Sync duration
  - `connector_records_fetched` - Records fetched
  - `connector_errors_count` - Connector errors

- **Job Metrics**:
  - `job_execution_count` - Job executions by type
  - `job_execution_duration` - Job execution time
  - `job_success_count` / `job_failure_count` - Success/failure rates
  - `job_queue_depth` - Current queue depth

### 3. **Tracing** (OpenTelemetry + Jaeger)
- Distributed traces across all services
- Automatic HTTP request tracing
- Database query tracing
- Custom spans for business operations:
  - Emissions calculations
  - Connector syncs
  - Report generation
  - Billing operations

## Quick Start

### 1. Start the Observability Stack

```bash
docker-compose up -d
```

This starts:
- **Prometheus** at http://localhost:9090
- **Grafana** at http://localhost:3001 (admin/admin)
- **Jaeger UI** at http://localhost:16686
- **OTEL Collector** at localhost:4318 (HTTP) / 4317 (gRPC)

### 2. Access Dashboards

**Grafana**: http://localhost:3001
- Username: `admin`
- Password: `admin`
- Pre-configured dashboards in "OffGridFlow" folder

**Jaeger**: http://localhost:16686
- Search for traces by service: `offgridflow-api`, `offgridflow-worker`
- Filter by operation, tags, duration

**Prometheus**: http://localhost:9090
- Query metrics directly
- View targets and alerts

### 3. Metrics Endpoint

The API exposes Prometheus metrics at:
```
http://localhost:8080/metrics
```

## Configuration

### Environment Variables

```bash
# Service identification
OTEL_SERVICE_NAME=offgridflow-api
OTEL_SERVICE_VERSION=1.0.0
ENVIRONMENT=production

# OTEL Collector endpoint
OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4318

# Sampling (0.0 to 1.0, where 1.0 = 100%)
OTEL_TRACE_SAMPLE_RATE=1.0

# Enable/disable observability features
ENABLE_METRICS=true
ENABLE_TRACING=true
ENABLE_LOGGING=true
```

### Code Integration

#### Initialize Observability

```go
import "github.com/example/offgridflow/internal/observability"

func main() {
    ctx := context.Background()
    
    // Create observability provider
    config := observability.DefaultConfig()
    provider, err := observability.NewProvider(ctx, config)
    if err != nil {
        log.Fatal(err)
    }
    defer provider.Shutdown(ctx)
    
    // Get logger and metrics
    logger := provider.Logger()
    metrics := provider.Metrics()
    
    // Wrap HTTP handlers
    handler := provider.HTTPMiddleware()(yourHandler)
}
```

#### Add Custom Metrics

```go
// Record a custom metric
metrics.RecordHTTPRequest(ctx, "GET", "/api/emissions", 200, duration)

// Record connector sync
metrics.RecordConnectorSync(ctx, "aws", 1500, duration, true)
```

#### Add Custom Traces

```go
import "github.com/example/offgridflow/internal/observability"

func processEmissions(ctx context.Context) error {
    ctx, span := observability.StartSpan(ctx, "emissions", "process_emissions")
    defer span.End()
    
    // Add attributes
    observability.SetSpanAttributes(ctx,
        attribute.String("org_id", orgID),
        attribute.Int("record_count", count),
    )
    
    // Record error if one occurs
    if err != nil {
        observability.RecordError(ctx, err)
        return err
    }
    
    return nil
}
```

## Dashboards

### API & Ingestion Dashboard

Located at: `deployments/grafana/dashboards/api-dashboard.json`

Panels:
1. **HTTP Request Rate** - Requests per second by endpoint
2. **API Latency (p95)** - 95th percentile response time
3. **Error Rate** - 4xx and 5xx error rates
4. **Connector Sync Rate** - Connector syncs per second
5. **Ingestion Job Duration** - p50 and p95 job durations
6. **Job Execution Counts** - Success/failure counts

### Creating Custom Dashboards

1. Open Grafana at http://localhost:3001
2. Click "+" → "Dashboard"
3. Add panels with PromQL queries
4. Export JSON and save to `deployments/grafana/dashboards/`

## Alerts

Alert rules are defined in: `deployments/grafana/alerts/alert-rules.yml`

### Configured Alerts

| Alert | Threshold | Duration | Severity |
|-------|-----------|----------|----------|
| HighErrorRate | 5xx rate > 5% | 5 min | Critical |
| HighClientErrorRate | 4xx rate > 20% | 10 min | Warning |
| HighAPILatency | p95 > 1s | 5 min | Warning |
| IngestionJobFailures | Failure rate > 0.1/s | 10 min | Warning |
| ConnectorSyncFailures | Error rate > 0.05/s | 15 min | Warning |
| DatabaseConnectionPoolExhaustion | Usage > 90% | 5 min | Critical |
| ServiceDown | Service unavailable | 1 min | Critical |

### Adding Custom Alerts

Edit `deployments/grafana/alerts/alert-rules.yml`:

```yaml
- alert: CustomAlert
  expr: your_metric > threshold
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "Alert description"
    description: "Detailed message"
```

## Troubleshooting

### No Metrics in Prometheus

1. Check OTEL Collector is running:
   ```bash
   docker logs offgridflow-otel-collector
   ```

2. Verify API is sending metrics:
   ```bash
   curl http://localhost:8080/metrics
   ```

3. Check Prometheus targets:
   http://localhost:9090/targets

### No Traces in Jaeger

1. Verify OTEL endpoint configuration:
   ```bash
   echo $OTEL_EXPORTER_OTLP_ENDPOINT
   ```

2. Check OTEL Collector logs for errors

3. Ensure sample rate is > 0:
   ```bash
   echo $OTEL_TRACE_SAMPLE_RATE
   ```

### High Cardinality Warnings

If you see warnings about high cardinality:
- Avoid using unbounded values (IDs, UUIDs) as label values
- Use labels for categorization, not identification
- Limit the number of unique label combinations

## Best Practices

### Logging

✅ **DO**:
- Use structured logging with key-value pairs
- Include request IDs for correlation
- Log at appropriate levels (DEBUG/INFO/WARN/ERROR)
- Include context (tenant_id, user_id, operation)

❌ **DON'T**:
- Log sensitive information (passwords, tokens)
- Log at DEBUG level in production by default
- Use string concatenation for messages

### Metrics

✅ **DO**:
- Use histograms for latency/duration
- Use counters for events (requests, errors)
- Use gauges for current values (connections, queue depth)
- Keep label cardinality low

❌ **DON'T**:
- Create metrics inside hot loops
- Use high-cardinality labels (user IDs, request IDs)
- Create too many metrics (start with key ones)

### Tracing

✅ **DO**:
- Trace all HTTP requests automatically
- Add spans for significant operations
- Include relevant attributes (org_id, record_count)
- Propagate context through async operations

❌ **DON'T**:
- Create spans for trivial operations
- Add too many attributes (adds overhead)
- Sample at 100% in high-traffic production

## Performance Impact

Observability has minimal performance impact when properly configured:

- **Logging**: < 1ms per request
- **Metrics**: < 0.5ms per metric update
- **Tracing**: 1-3ms per trace (with sampling)

Total overhead: < 5ms per request

## Production Considerations

1. **Sampling**: Set `OTEL_TRACE_SAMPLE_RATE=0.1` (10%) in production
2. **Retention**: Configure Prometheus retention (default: 15 days)
3. **Storage**: Monitor disk usage for Prometheus and Grafana
4. **Alerts**: Configure Alertmanager for notifications
5. **Security**: Use TLS for OTLP endpoints in production

## References

- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
- [Prometheus Best Practices](https://prometheus.io/docs/practices/)
- [Grafana Dashboards](https://grafana.com/docs/grafana/latest/dashboards/)
- [Jaeger Architecture](https://www.jaegertracing.io/docs/latest/architecture/)
