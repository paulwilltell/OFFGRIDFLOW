# Deployment Confidence: Push Button "Prod" Deploy

This document outlines the **confident infrastructure** approach for OffGridFlow, enabling repeatable, safe production deployments with a single command.

## 🎯 Goal

Run `scripts\deploy-complete.ps1` to execute:
1. Pre-flight safety checks
2. Database migrations
3. Docker image builds
4. Kubernetes rollouts
5. Post-deployment smoke tests

All in one repeatable, auditable flow.

## 🔒 Safety Rails

### Pre-Flight Checks
- ✅ Configuration validation (`deployment-checklist.ps1`)
- ✅ Environment variable verification
- ✅ Database connectivity test
- ✅ Redis connectivity test
- ✅ Secret presence validation (JWT, encryption keys)
- ✅ Container registry authentication
- ✅ Kubernetes cluster access

### Migration Safety
- ✅ Database backup before migrations
- ✅ Migration dry-run validation
- ✅ Rollback plan generation
- ✅ Schema version tracking
- ✅ Zero-downtime migration support

### Deployment Gates
- ✅ Staging validation required before prod
- ✅ Health check endpoints must pass
- ✅ Integration test suite must pass
- ✅ No critical security vulnerabilities
- ✅ Rate limiting configured
- ✅ Observability stack running

### Post-Deployment Validation
- ✅ Health endpoint checks (`/health`, `/ready`)
- ✅ Smoke tests for critical paths
- ✅ API authentication validation
- ✅ Database query performance
- ✅ Cache hit rate verification
- ✅ OpenTelemetry trace validation

## 📋 Deployment Flow

```powershell
# Single command production deployment
.\scripts\deploy-complete.ps1 -Environment production

# What happens:
# 1. Load .env.production
# 2. Run pre-flight checks
# 3. Backup database
# 4. Run migrations (with rollback plan)
# 5. Build Docker images (api, worker, web)
# 6. Push to container registry
# 7. Apply Kubernetes manifests
# 8. Wait for pod readiness
# 9. Run smoke tests
# 10. Verify metrics/logs/traces
# 11. Generate deployment report
```

## 🚦 Deployment Stages

### Staging First
```powershell
.\scripts\deploy-staging.ps1
```
- Uses `config/staging.yaml`
- Runs full test suite
- Validates compliance reports
- Stress tests ingestion pipelines
- Verifies cloud connectors (AWS/Azure/GCP)

### Production Rollout
```powershell
.\scripts\deploy-complete.ps1 -Environment production
```
- Requires staging success
- Uses `config/production.yaml`
- Blue/green deployment strategy
- Gradual pod replacement
- Automatic rollback on failure

## 🛠️ Infrastructure as Code

### Kubernetes Manifests
- `infra/k8s/api-deployment.yaml` - API server
- `infra/k8s/worker-deployment.yaml` - Background workers
- `infra/k8s/web-deployment.yaml` - Next.js frontend
- `infra/k8s/ingress.yaml` - Load balancer rules
- `infra/k8s/secrets.yaml` - Encrypted secrets

### Terraform Modules
- `infra/terraform/aws/` - AWS resources (RDS, S3, etc.)
- `infra/terraform/azure/` - Azure resources
- `infra/terraform/gcp/` - GCP resources
- `infra/terraform/kubernetes/` - K8s cluster setup

## 📊 Observability Integration

Every deployment automatically:
- Creates Grafana annotations
- Publishes deployment events to Prometheus
- Generates OpenTelemetry spans
- Updates status dashboard
- Sends notifications (Slack/email)

## 🔄 Rollback Strategy

### Automated Rollback Triggers
- Health checks fail for >2 minutes
- Error rate >5% within 10 minutes
- Database migration fails
- Critical pod crashes >3 times

### Manual Rollback
```powershell
.\scripts\rollback.ps1 -ToVersion v1.2.3
```

## 📝 Deployment Checklist

Created by `scripts\deployment-checklist.ps1`:

```yaml
environment: production
timestamp: 2024-12-01T10:11:30Z
checks:
  - name: database_connection
    status: passed
  - name: redis_connection
    status: passed
  - name: jwt_secret_present
    status: passed
  - name: stripe_keys_configured
    status: passed
  - name: aws_credentials_valid
    status: passed
  - name: azure_credentials_valid
    status: passed
  - name: gcp_credentials_valid
    status: passed
  - name: docker_registry_auth
    status: passed
  - name: kubernetes_cluster_access
    status: passed
```

## 🎓 Best Practices

1. **Never deploy on Friday** - Weekend incidents are costly
2. **Always test in staging first** - Catch issues early
3. **Deploy during low-traffic hours** - Minimize user impact
4. **Monitor for 30 minutes post-deploy** - Watch for anomalies
5. **Have rollback plan ready** - Know the escape hatch
6. **Document deployment changes** - Update CHANGELOG.md
7. **Run load tests** - Validate performance under load
8. **Verify compliance reports** - Ensure regulatory data flows

## 🔐 Security Considerations

- Secrets never in Git (use `.env`, vault, or K8s secrets)
- TLS certificates auto-renewed (Let's Encrypt)
- API rate limiting enabled by default
- Database credentials rotated regularly
- Container images scanned for vulnerabilities
- Access logs retained for audit trail

## 📞 Deployment Support

**On-Call Rotation**: See `docs/on-call.md`  
**Incident Response**: See `docs/incident-response.md`  
**Rollback Procedures**: See `docs/rollback.md`

---

**Confident Infrastructure = Repeatable + Safe + Observable Deployments**

Push `scripts\deploy-complete.ps1` with confidence. 🚀
