# Grafana Dashboards

Pre-configured Grafana dashboards for monitoring OffGridFlow performance and infrastructure.

---

## Available Dashboards

### 1. API Performance Dashboard
**File**: `api-performance-dashboard.json`

**Metrics**:
- Request rate (RPS) per endpoint
- Response time (p95, p99)
- Error rate with thresholds
- Active goroutines
- Memory usage (heap in-use, heap alloc)
- CPU usage
- Top endpoints by request count
- Top endpoints by latency

**Alerts**:
- API p95 latency > 200ms for 5 minutes
- Error rate > 1% (warning), > 5% (critical)

**Use Cases**:
- Monitor API health in real-time
- Identify slow endpoints
- Track error rates
- Capacity planning

---

### 2. Database Performance Dashboard
**File**: `database-performance-dashboard.json`

**Metrics**:
- Query rate (commits/rollbacks per second)
- Active connections vs. pool limit
- Cache hit ratio (target: > 90%)
- Slow queries (> 100ms)
- Deadlock detection
- Database size growth
- Top tables by size
- Replication lag
- Connection pool stats (open, in-use, idle)

**Alerts**:
- Replication lag > 30 seconds
- Connections > 90% of pool
- Cache hit ratio < 90%

**Use Cases**:
- Identify slow queries
- Monitor connection pool health
- Track database growth
- Detect replication issues

---

### 3. Worker Performance Dashboard
**File**: `worker-performance-dashboard.json`

**Metrics**:
- Batch processing rate
- Batch duration (p95, p99)
- Queue depth (pending jobs)
- Batch success rate
- Failed batches count
- Batch size distribution
- Throughput (records/sec)
- Worker memory usage
- Active workers count

**Thresholds**:
- Queue depth: Warning at 100, Critical at 500
- Success rate: Target > 99%

**Use Cases**:
- Monitor batch job health
- Identify bottlenecks
- Track throughput
- Queue management

---

### 4. System Resources Dashboard
**File**: `system-resources-dashboard.json`

**Metrics**:
- Pod CPU usage (per pod)
- Pod memory usage (per pod)
- HPA replica count (current vs. desired)
- Network I/O (RX/TX per pod)
- Disk I/O (reads/writes)
- Pod restarts (last 1 hour)
- Pods not ready
- Resource quotas
- Node resource allocation %

**Alerts**:
- Pod restarts > 5 in 1 hour
- Any pods not ready
- Resource allocation > 85%

**Use Cases**:
- Monitor Kubernetes health
- Track auto-scaling behavior
- Identify resource constraints
- Capacity planning

---

## Installation

### Method 1: Grafana UI (Manual Import)

1. Open Grafana: `http://localhost:3000` (or your Grafana URL)
2. Navigate to: **Dashboards → Import**
3. Click **Upload JSON file**
4. Select one of the JSON files from this directory
5. Choose Prometheus data source
6. Click **Import**

Repeat for each dashboard.

---

### Method 2: Provisioning (Automated)

Create a provisioning configuration:

**File**: `grafana-provisioning.yaml`
```yaml
apiVersion: 1

providers:
  - name: 'OffGridFlow Dashboards'
    orgId: 1
    folder: 'OffGridFlow'
    type: file
    disableDeletion: false
    updateIntervalSeconds: 10
    allowUiUpdates: true
    options:
      path: /var/lib/grafana/dashboards/offgridflow
```

Mount the dashboards directory:

```yaml
# docker-compose.yml or kubernetes volume
volumes:
  - ./infra/grafana:/var/lib/grafana/dashboards/offgridflow:ro
  - ./grafana-provisioning.yaml:/etc/grafana/provisioning/dashboards/offgridflow.yaml:ro
```

Restart Grafana:
```bash
docker-compose restart grafana
# or
kubectl rollout restart deployment/grafana -n offgridflow
```

---

## Data Sources

All dashboards require **Prometheus** as the data source.

**Prometheus Configuration**:
- See: `infra/prometheus.yml`
- Scrape interval: 15s
- Evaluation interval: 15s

**Metrics Exporters**:
- OffGridFlow API: `/metrics` endpoint on port 8080
- PostgreSQL: postgres_exporter
- Kubernetes: kube-state-metrics

---

## Alert Configuration

### Alertmanager Setup

Alerts are defined in dashboards but require Alertmanager for notifications.

**Example Alertmanager Config**:
```yaml
# alertmanager.yml
route:
  group_by: ['alertname', 'cluster']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 12h
  receiver: 'slack-notifications'

receivers:
  - name: 'slack-notifications'
    slack_configs:
      - api_url: 'YOUR_SLACK_WEBHOOK'
        channel: '#offgridflow-alerts'
        title: '{{ .GroupLabels.alertname }}'
        text: '{{ range .Alerts }}{{ .Annotations.description }}{{ end }}'
```

---

## Customization

### Adding Custom Panels

1. Open dashboard in Grafana
2. Click **Add panel** or **Edit**
3. Configure query, visualization, thresholds
4. Click **Save**
5. Export updated JSON: **Dashboard settings → JSON Model**
6. Save to this directory

### Modifying Thresholds

Edit JSON file:
```json
"thresholds": [
  {"value": 200, "colorMode": "warning"},
  {"value": 500, "colorMode": "critical"}
]
```

### Changing Refresh Rate

Edit JSON file:
```json
"refresh": "30s"  // Options: 5s, 10s, 30s, 1m, 5m
```

---

## Dashboard Screenshots

*(Add screenshots after deployment)*

### API Performance
![API Dashboard](screenshots/api-dashboard.png)

### Database Performance
![Database Dashboard](screenshots/database-dashboard.png)

### Worker Performance
![Worker Dashboard](screenshots/worker-dashboard.png)

### System Resources
![System Dashboard](screenshots/system-dashboard.png)

---

## Troubleshooting

### No Data Showing

**Check Prometheus targets**:
```bash
# Open Prometheus UI
http://localhost:9090/targets

# Verify targets are UP
# Expected: offgridflow-api, postgres, otel-collector
```

**Check Grafana data source**:
1. Navigate to: **Configuration → Data Sources**
2. Click **Prometheus**
3. Click **Test** button
4. Should show: "Data source is working"

### Metrics Missing

**Verify metric names**:
```bash
# Query Prometheus directly
curl http://localhost:9090/api/v1/query?query=http_requests_total

# Check if metric exists
```

**Check scrape configs**: See `infra/prometheus.yml`

### Dashboard Not Loading

**Check Grafana logs**:
```bash
docker-compose logs grafana
# or
kubectl logs deployment/grafana -n offgridflow
```

**Common issues**:
- Invalid JSON format
- Missing data source
- Permissions issue

---

## Performance Baseline

**Expected Metrics** (under normal load):

| Metric | Expected Range |
|--------|----------------|
| API p95 latency | 50-150ms |
| Database connections | 5-15 (out of 25) |
| Cache hit ratio | > 90% |
| Error rate | < 0.5% |
| CPU usage | 20-60% |
| Memory usage | 30-70% |
| Queue depth | < 50 |
| Batch success rate | > 99% |

---

## Integration with CI/CD

Monitor dashboard health in CI:

```bash
# scripts/check-grafana-health.sh
#!/bin/bash

# Check if dashboards exist
EXPECTED_DASHBOARDS=("api-performance" "database-performance" "worker-performance" "system-resources")

for dashboard in "${EXPECTED_DASHBOARDS[@]}"; do
    response=$(curl -s "http://grafana:3000/api/search?query=$dashboard")
    if [[ $response == "[]" ]]; then
        echo "❌ Dashboard $dashboard not found"
        exit 1
    fi
done

echo "✅ All dashboards loaded"
```

---

## Next Steps

1. **Deploy Grafana**: `docker-compose up grafana` or `kubectl apply -f infra/k8s/grafana.yaml`
2. **Import Dashboards**: Use one of the methods above
3. **Configure Alerts**: Set up Alertmanager
4. **Customize**: Adjust thresholds for your environment
5. **Monitor**: Access dashboards at `http://localhost:3000`

---

**Last Updated**: December 5, 2025  
**Version**: 1.0.0  
**Maintained By**: OffGridFlow Platform Team
