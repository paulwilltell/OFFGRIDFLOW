# Runbook: OffGridFlow API Down

**Alert Name**: `OffGridFlowAPIDown`  
**Severity**: Critical  
**Team**: Platform  

## Symptoms

- Prometheus alert fired: `up{job="offgridflow-api"} == 0`
- API endpoints returning connection refused or timeout errors
- Health check endpoint `/health/live` not responding
- Users reporting "Service Unavailable" errors

## Impact

- **User Impact**: HIGH - All API functionality unavailable
- **Business Impact**: CRITICAL - No carbon data can be accessed or recorded
- **SLA Impact**: Production SLA breach after 5 minutes

## Triage

### 1. Confirm the Issue (2 minutes)

```bash
# Check if API is actually down
curl -f http://api.offgridflow.com/health/live

# Check Kubernetes pod status
kubectl get pods -n offgridflow -l app=offgridflow-api

# Check recent pod events
kubectl describe pod -n offgridflow -l app=offgridflow-api

# Check pod logs
kubectl logs -n offgridflow -l app=offgridflow-api --tail=100
```

### 2. Identify Root Cause (5 minutes)

Common causes and checks:

#### A. Pod Crash Loop
```bash
# Check restart count
kubectl get pods -n offgridflow -l app=offgridflow-api

# If restart count > 0, check crash reason
kubectl logs -n offgridflow -l app=offgridflow-api --previous
```

**Likely causes**: 
- Database connection failure
- Configuration error
- OOM kill

#### B. Database Connection Failure
```bash
# Check database connectivity
kubectl exec -n offgridflow deployment/offgridflow-api -- \
  pg_isready -h postgres-service -p 5432

# Check database logs
kubectl logs -n offgridflow statefulset/postgres --tail=50
```

#### C. Resource Exhaustion
```bash
# Check resource usage
kubectl top pods -n offgridflow -l app=offgridflow-api

# Check if hitting limits
kubectl describe pod -n offgridflow -l app=offgridflow-api | grep -A 5 "Limits"
```

#### D. Network/Ingress Issues
```bash
# Check ingress status
kubectl get ingress -n offgridflow

# Check service endpoints
kubectl get endpoints -n offgridflow offgridflow-api-service
```

## Resolution Steps

### Quick Fix (Immediate - 2 minutes)

If pods are in CrashLoopBackOff or Error state:

```bash
# Restart deployment
kubectl rollout restart deployment/offgridflow-api -n offgridflow

# Watch rollout status
kubectl rollout status deployment/offgridflow-api -n offgridflow
```

### Fix 1: Database Connection Issues

```bash
# Verify database is running
kubectl get pods -n offgridflow -l app=postgres

# Check database secrets
kubectl get secret offgridflow-secrets -n offgridflow -o yaml

# If database is down, restart it
kubectl rollout restart statefulset/postgres -n offgridflow

# Wait for database to be ready
kubectl wait --for=condition=ready pod -l app=postgres -n offgridflow --timeout=300s

# Restart API after database is ready
kubectl rollout restart deployment/offgridflow-api -n offgridflow
```

### Fix 2: OOM (Out of Memory) Issues

```bash
# Confirm OOM kill in events
kubectl get events -n offgridflow --sort-by='.lastTimestamp' | grep OOMKilled

# Temporary fix: Increase memory limits
kubectl patch deployment offgridflow-api -n offgridflow -p '
{
  "spec": {
    "template": {
      "spec": {
        "containers": [{
          "name": "api",
          "resources": {
            "limits": {"memory": "2Gi"},
            "requests": {"memory": "1Gi"}
          }
        }]
      }
    }
  }
}'

# Monitor memory usage
watch kubectl top pods -n offgridflow -l app=offgridflow-api
```

### Fix 3: Configuration Errors

```bash
# Check ConfigMap
kubectl get configmap offgridflow-config -n offgridflow -o yaml

# Check for recent changes
kubectl rollout history deployment/offgridflow-api -n offgridflow

# Rollback to previous version if needed
kubectl rollout undo deployment/offgridflow-api -n offgridflow

# Watch rollout
kubectl rollout status deployment/offgridflow-api -n offgridflow
```

### Fix 4: Image Pull Errors

```bash
# Check if image pull failed
kubectl describe pod -n offgridflow -l app=offgridflow-api | grep "Failed to pull image"

# Verify image exists
docker pull ghcr.io/your-org/offgridflow-api:latest

# If image doesn't exist, rollback or fix image tag
kubectl set image deployment/offgridflow-api -n offgridflow \
  api=ghcr.io/your-org/offgridflow-api:v1.2.3
```

## Verification (3 minutes)

```bash
# 1. Check pod health
kubectl get pods -n offgridflow -l app=offgridflow-api
# Expected: All pods Running with 0 restarts

# 2. Check health endpoint
curl http://api.offgridflow.com/health/ready
# Expected: {"status":"healthy"...}

# 3. Test API functionality
curl -H "Authorization: Bearer $TOKEN" \
  http://api.offgridflow.com/v1/activities?limit=1
# Expected: 200 OK with activity data

# 4. Check metrics
curl http://api.offgridflow.com:8081/metrics | grep "http_requests_total"
# Expected: Metrics being collected

# 5. Check error rate in Grafana
# Navigate to: https://grafana.offgridflow.com/d/api-overview
# Verify: Error rate < 1%
```

## Post-Incident Steps

1. **Document the incident** in incident tracker with:
   - Root cause
   - Time to detect
   - Time to resolve
   - Impact duration

2. **Create follow-up tasks**:
   ```bash
   # If memory issue: Tune memory limits permanently
   # If database issue: Review connection pooling
   # If config issue: Improve validation in CI/CD
   ```

3. **Update runbook** if new issue or solution discovered

4. **Schedule post-mortem** if incident duration > 30 minutes

## Escalation

- **Primary**: Platform team (#platform-oncall)
- **Secondary**: Backend team (#backend-engineering)
- **Manager**: Director of Engineering
- **After hours**: PagerDuty escalation policy

## Prevention

- [ ] Implement pre-deployment health checks in CI/CD
- [ ] Add database connection retry logic with exponential backoff
- [ ] Set up proper resource limits and HPA
- [ ] Enable pod disruption budgets
- [ ] Add automated rollback on failed deployments

## Related Alerts

- `PostgreSQLDown` - Database connectivity issues
- `PodCrashLooping` - Pod restart issues
- `OffGridFlowAPIHighErrorRate` - Partial API degradation

## References

- [Kubernetes Troubleshooting Guide](https://kubernetes.io/docs/tasks/debug/)
- [API Architecture Diagram](../architecture.md)
- [Database Connection Pooling Config](../database.md)
- [Deployment Procedure](../deployment.md)

---
**Last Updated**: 2025-12-27  
**Owner**: Platform Team  
**Review Cycle**: Quarterly
