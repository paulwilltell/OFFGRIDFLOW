# Observability Quick Start

Get full observability (logs, metrics, traces) running in under 2 minutes.

## Start the Stack

```bash
# Start all services including observability
docker-compose up -d

# Wait for services to be healthy (~30 seconds)
docker-compose ps
```

## Access the Tools

| Tool | URL | Credentials | Purpose |
|------|-----|-------------|---------|
| **Grafana** | http://localhost:3001 | admin/admin | Dashboards & visualizations |
| **Prometheus** | http://localhost:9090 | None | Metrics & alerts |
| **Jaeger** | http://localhost:16686 | None | Distributed tracing |
| **API Metrics** | http://localhost:8080/metrics | None | Prometheus endpoint |

## View Your First Dashboard

1. Open Grafana: http://localhost:3001
2. Login: `admin` / `admin`
3. Navigate to: **Dashboards** → **OffGridFlow** → **API & Ingestion Dashboard**
4. Generate some traffic:
   ```bash
   # Hit the API a few times
   for i in {1..20}; do
     curl http://localhost:8080/api/health
     curl http://localhost:8080/api/emissions
     sleep 1
   done
   ```
5. Watch the dashboard update in real-time

## View Your First Trace

1. Open Jaeger: http://localhost:16686
2. Select **Service**: `offgridflow-api`
3. Click **Find Traces**
4. Click on any trace to see the full request flow
5. Expand spans to see:
   - HTTP request details
   - Database queries
   - Emissions calculations
   - Response times

## Query Metrics Directly

Open Prometheus: http://localhost:9090/graph

Try these queries:

```promql
# Request rate per second
rate(http_request_count[5m])

# 95th percentile latency (in milliseconds)
histogram_quantile(0.95, rate(http_request_duration_bucket[5m])) * 1000

# Error rate percentage
(sum(rate(http_request_count{http_status_code=~"5.."}[5m])) / sum(rate(http_request_count[5m]))) * 100

# Active connector syncs
connector_sync_count

# Job execution duration p95
histogram_quantile(0.95, rate(job_execution_duration_bucket[5m]))
```

## Check Logs with Request IDs

```bash
# View structured logs
docker logs offgridflow-api --tail 50

# Filter by request ID
docker logs offgridflow-api 2>&1 | grep "request_id"

# Parse JSON logs with jq
docker logs offgridflow-api 2>&1 | jq 'select(.request_id != null)'

# Find slow requests (> 100ms)
docker logs offgridflow-api 2>&1 | jq 'select(.latency_ms > 100)'
```

## View Active Alerts

1. Open Prometheus: http://localhost:9090/alerts
2. You'll see 14 pre-configured alerts
3. Alerts will fire when thresholds are exceeded:
   - Error rate > 5% for 5 minutes
   - API latency p95 > 1 second
   - Connector failures
   - Job failures
   - And more...

## Trigger an Alert (Testing)

```bash
# Generate errors to trigger HighErrorRate alert
for i in {1..100}; do
  curl http://localhost:8080/api/nonexistent
done

# Wait 5 minutes, then check Prometheus alerts
# The "HighClientErrorRate" alert should fire
```

## Troubleshooting

### No metrics showing?

```bash
# Check if API is exposing metrics
curl http://localhost:8080/metrics

# Check Prometheus targets
curl http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | {job, health}'

# Check OTEL collector logs
docker logs offgridflow-otel-collector
```

### No traces in Jaeger?

```bash
# Check OTEL collector is receiving traces
docker logs offgridflow-otel-collector | grep -i trace

# Verify API is configured correctly
docker exec offgridflow-api env | grep OTEL
```

### Grafana dashboard not showing data?

1. Check data source: **Configuration** → **Data Sources** → **Prometheus**
2. Click **Test** to verify connection
3. Ensure time range is set to "Last 15 minutes"
4. Generate traffic if no data yet

## Next Steps

- **Customize dashboards**: Create panels for business-specific metrics
- **Configure alerts**: Add Slack/PagerDuty notifications
- **Explore traces**: Understand your application's flow
- **Optimize performance**: Use metrics to find bottlenecks

## Documentation

- Full guide: [docs/OBSERVABILITY.md](docs/OBSERVABILITY.md)
- Implementation details: [docs/OBSERVABILITY_SUMMARY.md](docs/OBSERVABILITY_SUMMARY.md)
- Checklist: [OBSERVABILITY_CHECKLIST.md](OBSERVABILITY_CHECKLIST.md)

## What You Get

✅ **Structured logging** with request IDs for correlation  
✅ **33 metrics** covering HTTP, DB, emissions, connectors, jobs  
✅ **Distributed tracing** for auth, emissions, ingestion flows  
✅ **Live dashboard** with API latency, error rates, job metrics  
✅ **14 alerts** for critical issues (error rates, latency, failures)  
✅ **< 5ms overhead** per request  

**When something breaks, you see where and why—no more guessing.**
