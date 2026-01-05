# OffGridFlow Production Deployment Checklist

> **Purpose**: This checklist ensures all critical systems are verified before deploying to production.  
> **Owner**: Platform Team  
> **Review**: Before every major release

---

## Pre-Deployment Checklist

### 1. Code Quality & Testing ✅

- [ ] All CI/CD pipelines passing (GitHub Actions)
  - [ ] Backend tests: `go test ./... -race -cover`
  - [ ] Frontend tests: `npm run test:coverage`
  - [ ] Integration tests passed
  - [ ] Security scans passed (Trivy, gosec, CodeQL)

- [ ] Code coverage meets minimum threshold (60%+)
  ```bash
  # Backend
  go test -coverprofile=coverage.out ./...
  go tool cover -func=coverage.out | grep total
  
  # Frontend
  cd web && npm run test:coverage
  ```

- [ ] No known critical or high-severity vulnerabilities
  ```bash
  # Check dependency vulnerabilities
  go list -json -m all | nancy sleuth
  cd web && npm audit --production
  ```

- [ ] Code review completed and approved (minimum 2 approvers)

- [ ] All TODOs and FIXMEs addressed or tracked in backlog

### 2. Configuration & Secrets 🔐

- [ ] Production secrets rotated and secured
  ```bash
  # Generate new secrets
  ./scripts/rotate-secrets.sh offgridflow-secrets production
  
  # Verify sealed secrets
  kubectl get sealedsecrets -n offgridflow
  ```

- [ ] Environment variables validated
  ```bash
  ./scripts/validate_env_production.sh
  ```

- [ ] Database connection strings verified
  ```bash
  kubectl exec -n offgridflow deployment/offgridflow-api -- \
    pg_isready -h $POSTGRES_HOST -p 5432
  ```

- [ ] API keys for third-party services valid
  - [ ] Stripe keys (production)
  - [ ] OpenAI API key (if enabled)
  - [ ] AWS credentials (for cloud ingestion)
  - [ ] Azure credentials (for cloud ingestion)
  - [ ] GCP service account (for cloud ingestion)

- [ ] TLS certificates valid and not expiring < 30 days
  ```bash
  echo | openssl s_client -connect api.offgridflow.com:443 2>/dev/null | \
    openssl x509 -noout -dates
  ```

### 3. Database 🗄️

- [ ] Database migrations tested in staging
  ```bash
  # Dry run migrations
  go run cmd/migrate/main.go -command version
  go run cmd/migrate/main.go -command up
  ```

- [ ] Backup system operational
  ```bash
  # Verify last backup
  aws s3 ls s3://offgridflow-backups/backups/postgres/ | tail -5
  
  # Test restore process
  ./scripts/restore-backup.sh [latest-backup] offgridflow_test
  ```

- [ ] Database connection pool sized correctly
  - [ ] `max_connections` in PostgreSQL ≥ (app_replicas × max_open_conns) + 20
  - [ ] Connection pool config matches expected load

- [ ] Database performance indexes created
  ```sql
  -- Verify critical indexes exist
  SELECT schemaname, tablename, indexname 
  FROM pg_indexes 
  WHERE schemaname = 'public';
  ```

- [ ] Database disk space > 50% free
  ```bash
  kubectl exec -n offgridflow statefulset/postgres -- \
    df -h /var/lib/postgresql/data
  ```

### 4. Infrastructure & Resources 🏗️

- [ ] Kubernetes cluster health verified
  ```bash
  kubectl get nodes
  kubectl top nodes
  kubectl get pods --all-namespaces | grep -v Running
  ```

- [ ] Resource limits and requests configured
  ```bash
  kubectl describe deployment offgridflow-api -n offgridflow | grep -A 4 "Limits"
  ```

- [ ] Horizontal Pod Autoscaler (HPA) configured
  ```bash
  kubectl get hpa -n offgridflow
  kubectl describe hpa offgridflow-api -n offgridflow
  ```

- [ ] Pod Disruption Budgets (PDB) in place
  ```bash
  kubectl get pdb -n offgridflow
  ```

- [ ] Network policies applied
  ```bash
  kubectl get networkpolicies -n offgridflow
  ```

- [ ] Ingress controller configured with rate limiting
  ```bash
  kubectl describe ingress offgridflow-ingress -n offgridflow
  ```

### 5. Observability & Monitoring 📊

- [ ] Prometheus scraping all targets
  ```bash
  # Check Prometheus targets
  curl http://prometheus.offgridflow.com/api/v1/targets | jq '.data.activeTargets[] | {job: .labels.job, health: .health}'
  ```

- [ ] Grafana dashboards operational
  - [ ] API Overview dashboard
  - [ ] Database Performance dashboard
  - [ ] Kubernetes Cluster dashboard
  - [ ] Business Metrics dashboard

- [ ] Alert rules loaded and firing (test mode)
  ```bash
  # Verify alert rules
  curl http://prometheus.offgridflow.com/api/v1/rules | jq '.data.groups[].rules[] | {alert: .name, state: .state}'
  ```

- [ ] Alert notification channels configured
  - [ ] PagerDuty integration
  - [ ] Slack #alerts channel
  - [ ] Email notifications for critical alerts

- [ ] Distributed tracing operational (Jaeger)
  ```bash
  curl http://jaeger.offgridflow.com/api/services | jq
  ```

- [ ] Log aggregation working (Loki/Elasticsearch)
  ```bash
  # Test log query
  kubectl logs -n offgridflow -l app=offgridflow-api --tail=10
  ```

### 6. Security 🔒

- [ ] Security scanning completed
  - [ ] Container images scanned (Trivy)
  - [ ] Code static analysis (gosec, CodeQL)
  - [ ] Dependency vulnerabilities checked

- [ ] RBAC policies reviewed and applied
  ```bash
  kubectl get rolebindings -n offgridflow
  kubectl get clusterrolebindings | grep offgridflow
  ```

- [ ] Service accounts follow least privilege principle
  ```bash
  kubectl get serviceaccounts -n offgridflow
  ```

- [ ] Secrets encrypted at rest (Sealed Secrets or Vault)
  ```bash
  kubectl get sealedsecrets -n offgridflow
  ```

- [ ] Network segmentation in place
  - [ ] Database not publicly accessible
  - [ ] Redis not publicly accessible
  - [ ] Internal services use ClusterIP

- [ ] Rate limiting enabled on API endpoints
  ```bash
  # Test rate limit
  for i in {1..100}; do curl -w "%{http_code}\n" -o /dev/null -s http://api.offgridflow.com/v1/health; done
  ```

- [ ] WAF rules configured (if using Cloudflare/AWS WAF)

- [ ] DDoS protection enabled

### 7. Performance & Load Testing 🚀

- [ ] Load tests passed with acceptable performance
  ```bash
  # Run k6 load test
  k6 run scripts/load-test.k6.js
  
  # Verify results
  # - P95 latency < 500ms
  # - P99 latency < 2000ms
  # - Error rate < 1%
  # - Throughput > 1000 req/s
  ```

- [ ] Database query performance acceptable
  ```sql
  -- Check slow queries
  SELECT query, mean_exec_time, calls
  FROM pg_stat_statements
  ORDER BY mean_exec_time DESC
  LIMIT 10;
  ```

- [ ] Frontend Lighthouse score > 90
  ```bash
  lighthouse https://app.offgridflow.com --view
  ```

- [ ] CDN caching configured for static assets

- [ ] Database connection pooling optimized
  ```sql
  -- Check connection usage
  SELECT count(*) as connections, state
  FROM pg_stat_activity
  GROUP BY state;
  ```

### 8. Disaster Recovery 💾

- [ ] Backup strategy documented and tested
  - [ ] Daily automated backups running
  - [ ] Backup retention: 30 days
  - [ ] Backups stored in separate region/account

- [ ] Restore procedure tested successfully
  ```bash
  # Test restore to staging
  ./scripts/restore-backup.sh offgridflow_backup_20250127.sql.gz staging_db
  ```

- [ ] RTO (Recovery Time Objective) defined: **4 hours**

- [ ] RPO (Recovery Point Objective) defined: **1 hour**

- [ ] Disaster recovery runbook completed and reviewed

- [ ] Off-site backup verification performed

### 9. Documentation 📚

- [ ] API documentation up to date (OpenAPI/Swagger)

- [ ] Architecture diagrams current

- [ ] Runbooks created for common incidents
  - [ ] API Down
  - [ ] Database connectivity issues
  - [ ] High error rate
  - [ ] Performance degradation
  - [ ] Security incidents

- [ ] Deployment procedure documented

- [ ] Rollback procedure documented and tested

- [ ] On-call rotation schedule published

- [ ] Incident response plan reviewed

### 10. Compliance & Legal ⚖️

- [ ] Data privacy policies implemented (GDPR, CCPA)

- [ ] Audit logging enabled for compliance
  ```bash
  # Verify audit logs
  kubectl logs -n offgridflow -l app=offgridflow-api | grep "audit"
  ```

- [ ] Data retention policies configured

- [ ] Terms of Service and Privacy Policy updated

- [ ] Security policy reviewed (if SOC 2 certified)

- [ ] Compliance frameworks configured (CSRD, SEC, CBAM)

---

## Deployment Execution

### Pre-Deployment (T-60 minutes)

```bash
# 1. Announce maintenance window (if required)
# Post in: #engineering, #customer-success

# 2. Enable read-only mode (optional for zero-downtime)
kubectl annotate deployment offgridflow-api -n offgridflow \
  maintenance.offgridflow.com/mode=read-only

# 3. Create database backup
kubectl exec -n offgridflow cronjob/postgres-backup -- \
  /backup-script.sh

# 4. Verify staging deployment successful
kubectl get pods -n offgridflow-staging
```

### Deployment (T-0)

```bash
# 1. Run database migrations
kubectl apply -f infra/k8s/migrate-job.yaml
kubectl wait --for=condition=complete job/migrate-job -n offgridflow --timeout=300s

# 2. Deploy backend
kubectl apply -f infra/k8s/api-deployment.yaml
kubectl rollout status deployment/offgridflow-api -n offgridflow

# 3. Deploy worker
kubectl apply -f infra/k8s/worker-deployment.yaml
kubectl rollout status deployment/offgridflow-worker -n offgridflow

# 4. Deploy frontend
kubectl apply -f infra/k8s/web-deployment.yaml
kubectl rollout status deployment/offgridflow-web -n offgridflow

# 5. Update ingress (if needed)
kubectl apply -f infra/k8s/ingress.yaml
```

### Post-Deployment Verification (T+15 minutes)

```bash
# 1. Health checks
curl https://api.offgridflow.com/health/ready
curl https://api.offgridflow.com/health/live

# 2. Smoke tests
./scripts/smoke-tests.sh production

# 3. Monitor error rates
# Check Grafana: https://grafana.offgridflow.com/d/api-overview

# 4. Check logs for errors
kubectl logs -n offgridflow -l app=offgridflow-api --tail=100 | grep ERROR

# 5. Verify metrics
curl https://api.offgridflow.com:8081/metrics | grep http_requests_total

# 6. Test critical user flows
# - User login
# - Create activity
# - Generate report
# - View dashboard
```

### Monitoring Period (T+2 hours)

- [ ] Monitor error rate < 1%
- [ ] Monitor P95 latency < 500ms
- [ ] Monitor database connections < 80% of max
- [ ] Monitor memory usage < 80% of limit
- [ ] Monitor CPU usage < 70% of limit
- [ ] No alerts firing in Prometheus
- [ ] Customer support reports no issues

---

## Rollback Procedure

If issues detected within 2 hours of deployment:

```bash
# 1. Immediate rollback
kubectl rollout undo deployment/offgridflow-api -n offgridflow
kubectl rollout undo deployment/offgridflow-worker -n offgridflow
kubectl rollout undo deployment/offgridflow-web -n offgridflow

# 2. Verify rollback
kubectl rollout status deployment/offgridflow-api -n offgridflow

# 3. Rollback database migrations (if needed)
go run cmd/migrate/main.go -command down -steps 1

# 4. Verify service health
curl https://api.offgridflow.com/health

# 5. Monitor for 30 minutes

# 6. Post incident review
```

---

## Sign-Off

| Role | Name | Signature | Date |
|------|------|-----------|------|
| **Engineering Lead** | | | |
| **DevOps Lead** | | | |
| **Security Lead** | | | |
| **QA Lead** | | | |
| **Product Manager** | | | |

---

## Post-Deployment

- [ ] Deployment retrospective scheduled (within 48 hours)
- [ ] Metrics reviewed and baseline updated
- [ ] Known issues documented
- [ ] Customer-facing changelog published
- [ ] Internal team notified of new features/changes

---

**Last Updated**: 2025-12-27  
**Template Version**: 2.0  
**Next Review**: 2026-01-27
