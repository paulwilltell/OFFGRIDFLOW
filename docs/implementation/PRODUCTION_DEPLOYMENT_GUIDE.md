# OffGridFlow Production Deployment Guide

## Prerequisites Checklist

### 1. System Requirements
- [ ] Go 1.21+ installed
- [ ] Docker and Docker Compose installed
- [ ] PostgreSQL client tools (psql) installed
- [ ] kubectl installed (for Kubernetes deployments)
- [ ] Git installed

### 2. Infrastructure Ready
- [ ] PostgreSQL 15+ database provisioned
- [ ] Redis 7+ instance provisioned
- [ ] S3 bucket created (for AWS ingestion)
- [ ] Domain configured with SSL certificates
- [ ] Load balancer configured
- [ ] OpenTelemetry Collector deployed

### 3. Third-Party Services Configured
- [ ] Stripe account created (live mode)
- [ ] Stripe webhook endpoint configured
- [ ] SendGrid account created
- [ ] Sentry project created (optional)
- [ ] Datadog account created (optional)

### 4. Secrets Generated
- [ ] JWT secret (48+ characters) - Use: `openssl rand -base64 48`
- [ ] Session secret (48+ characters) - Use: `openssl rand -base64 48`
- [ ] Encryption key (32 characters) - Use: `openssl rand -base64 32`
- [ ] Database password (strong, unique)
- [ ] Redis password (strong, unique)

## Deployment Steps

### Step 1: Configure Environment

```powershell
# Copy production template
cp .env.production.template .env.production

# Edit and fill in all values
# CRITICAL: Replace ALL placeholder values
notepad .env.production
```

**Required values to update:**
- `DATABASE_URL` - Your production PostgreSQL connection string
- `DB_PASSWORD` - Strong database password
- `JWT_SECRET` - Generate with: `openssl rand -base64 48`
- `SESSION_SECRET` - Generate with: `openssl rand -base64 48`
- `ENCRYPTION_KEY` - Generate with: `openssl rand -base64 32`
- `STRIPE_SECRET_KEY` - Your live Stripe secret key (starts with `sk_live_`)
- `STRIPE_PUBLISHABLE_KEY` - Your live Stripe publishable key
- `STRIPE_WEBHOOK_SECRET` - From Stripe webhook configuration
- `AWS_ACCESS_KEY_ID` - If using AWS ingestion
- `AWS_SECRET_ACCESS_KEY` - If using AWS ingestion
- `SMTP_PASSWORD` - Your SendGrid API key
- `REDIS_URL` - Your Redis connection string
- `FRONTEND_URL` - Your production frontend URL

### Step 2: Validate Configuration

```powershell
# Run the deployment checklist
.\scripts\deployment-checklist.ps1 -EnvFile .env.production
```

Fix any errors before proceeding.

### Step 3: Install Dependencies

```powershell
# Install Go dependencies
go mod tidy
go mod verify

# Vendor dependencies (optional, for offline builds)
go mod vendor
```

### Step 4: Run Database Migrations

```powershell
# Run migrations
.\scripts\migrate.ps1

# Verify tables
$env:PGPASSWORD="your_password"
psql -h your-db-host -U offgridflow -d offgridflow -c "\dt"
```

Expected tables:
- tenants
- users
- api_keys
- subscriptions
- activities
- emissions
- emission_factors
- jobs
- audit_logs
- compliance_reports
- data_sources

### Step 5: Build Application

```powershell
# Build API
go build -o offgridflow-api.exe ./cmd/api

# Build Worker
go build -o offgridflow-worker.exe ./cmd/worker

# Test binaries
.\offgridflow-api.exe --version
```

### Step 6: Test Locally with Docker Compose

```powershell
# Start all services
docker-compose up -d

# Check health
curl http://localhost:8080/health

# View logs
docker-compose logs -f api

# Run integration tests
.\scripts\test-integration.ps1
```

### Step 7: Deploy to Staging

```powershell
# Deploy to staging environment
.\scripts\deploy-staging.ps1

# Monitor deployment
kubectl get pods -n offgridflow -w

# Check health endpoints
kubectl port-forward svc/offgridflow-api 8080:8080 -n offgridflow
curl http://localhost:8080/health
```

### Step 8: Run Smoke Tests

```powershell
# Test authentication
$response = Invoke-WebRequest -Uri "https://staging.offgridflow.com/api/v1/auth/login" -Method Post -Body (@{email="test@example.com";password="test123"} | ConvertTo-Json) -ContentType "application/json"

# Test emissions calculation
# Test job queue
# Test webhooks
# Test rate limiting
```

### Step 9: Deploy to Production

```powershell
# Set production environment
$env:CLUSTER_NAME = "offgridflow-production"
$env:NAMESPACE = "offgridflow-prod"
$env:ENV = "production"

# Deploy
.\scripts\deploy-staging.ps1

# Monitor rollout
kubectl rollout status deployment/offgridflow-api -n offgridflow-prod
kubectl rollout status deployment/offgridflow-worker -n offgridflow-prod
```

### Step 10: Post-Deployment Verification

```powershell
# Check all pods are running
kubectl get pods -n offgridflow-prod

# Check logs for errors
kubectl logs -f deployment/offgridflow-api -n offgridflow-prod

# Test health endpoints
curl https://app.offgridflow.com/health

# Test critical paths
# - User registration
# - Login
# - Emissions calculation
# - Data ingestion
# - Report generation
```

## Monitoring Setup

### Grafana Dashboards
1. Access Grafana at `http://localhost:3001` (or your Grafana URL)
2. Import dashboards from `infra/grafana/dashboards/`
3. Configure alerting rules

### Jaeger Tracing
1. Access Jaeger UI at `http://localhost:16686`
2. Search for traces by service name: `offgridflow-api`

### Logs
```powershell
# View API logs
kubectl logs -f deployment/offgridflow-api -n offgridflow-prod

# View worker logs
kubectl logs -f deployment/offgridflow-worker -n offgridflow-prod

# Search logs
kubectl logs deployment/offgridflow-api -n offgridflow-prod | Select-String "ERROR"
```

## Rollback Procedure

If issues are detected:

```powershell
# Rollback to previous version
kubectl rollout undo deployment/offgridflow-api -n offgridflow-prod
kubectl rollout undo deployment/offgridflow-worker -n offgridflow-prod

# Verify rollback
kubectl rollout status deployment/offgridflow-api -n offgridflow-prod
```

## Backup & Recovery

### Database Backup
```powershell
# Backup database
pg_dump -h your-db-host -U offgridflow -d offgridflow > backup-$(Get-Date -Format 'yyyyMMdd-HHmmss').sql

# Restore database
psql -h your-db-host -U offgridflow -d offgridflow < backup-20241201-120000.sql
```

### Application Data Backup
```powershell
# Backup S3 data
aws s3 sync s3://offgridflow-production-data ./backup-s3/

# Backup Redis (if persistent)
redis-cli --rdb dump.rdb
```

## Scaling

### Horizontal Scaling
```powershell
# Scale API
kubectl scale deployment offgridflow-api --replicas=5 -n offgridflow-prod

# Scale Workers
kubectl scale deployment offgridflow-worker --replicas=10 -n offgridflow-prod
```

### Autoscaling
```yaml
# Already configured in k8s/api-deployment.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: offgridflow-api
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: offgridflow-api
  minReplicas: 2
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
```

## Troubleshooting

### API Not Starting
```powershell
# Check logs
kubectl logs deployment/offgridflow-api -n offgridflow-prod --tail=100

# Check env vars
kubectl describe pod <pod-name> -n offgridflow-prod

# Check database connection
kubectl exec -it <pod-name> -n offgridflow-prod -- psql $DATABASE_URL -c "SELECT 1"
```

### High Memory Usage
```powershell
# Check memory usage
kubectl top pods -n offgridflow-prod

# Increase memory limits
kubectl edit deployment offgridflow-api -n offgridflow-prod
```

### Database Connection Pool Exhausted
```powershell
# Check active connections
psql -h your-db-host -U offgridflow -d offgridflow -c "SELECT count(*) FROM pg_stat_activity WHERE datname='offgridflow'"

# Increase max connections in .env
DB_MAX_CONNECTIONS=200
```

## Security Checklist

- [ ] All secrets stored in Kubernetes secrets (not in code)
- [ ] SSL/TLS enabled everywhere
- [ ] Database uses SSL connections
- [ ] CORS origins restricted to production domains
- [ ] Rate limiting enabled
- [ ] Audit logging enabled
- [ ] 2FA enabled for admin users
- [ ] Regular security updates applied
- [ ] Vulnerability scanning enabled
- [ ] Backup encryption enabled

## Performance Optimization

### Database Indexes
Already created in schema:
- Users email index
- Activities period_start/end indexes
- Emissions period indexes
- Jobs status and tenant_id indexes

### Redis Caching
Enable caching for:
- Emission factors (24h TTL)
- User sessions (configurable)
- Rate limiting counters (1min TTL)

### CDN Configuration
- Static assets should be served via CDN
- API responses cached where appropriate
- GZIP compression enabled

## Maintenance

### Regular Tasks
- [ ] Weekly: Review error logs
- [ ] Weekly: Check disk usage
- [ ] Weekly: Review performance metrics
- [ ] Monthly: Update dependencies
- [ ] Monthly: Review and optimize slow queries
- [ ] Quarterly: Security audit
- [ ] Quarterly: Backup restoration test

### Updating OffGridFlow
```powershell
# Pull latest code
git pull origin main

# Run tests
go test ./...

# Build new images
docker build -t offgridflow-api:new-version .

# Deploy with zero-downtime
kubectl set image deployment/offgridflow-api api=offgridflow-api:new-version -n offgridflow-prod
kubectl rollout status deployment/offgridflow-api -n offgridflow-prod
```

## Support Contacts

- **Database Issues**: DBA team
- **Infrastructure Issues**: DevOps team
- **Application Bugs**: Development team
- **Security Issues**: Security team

## Production URLs

- **API**: https://api.offgridflow.com
- **Web App**: https://app.offgridflow.com
- **Admin Panel**: https://admin.offgridflow.com
- **Grafana**: https://grafana.offgridflow.com
- **Jaeger**: https://jaeger.offgridflow.com (internal only)

---

**Last Updated**: 2024-12-01
**Version**: 1.0.0
