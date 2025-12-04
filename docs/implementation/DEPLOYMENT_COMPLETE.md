# OffGridFlow Deployment Completion Guide

## ✅ Completed Tasks

### 1. ✅ Go Dependencies Installation
- Ran `go mod tidy` to install all required dependencies
- All Go modules are now up-to-date

### 2. ✅ Environment Configuration (.env)
Created comprehensive `.env` file with all required configuration:
- ✅ Database configuration (PostgreSQL)
- ✅ Server settings (port, environment, CORS)
- ✅ JWT authentication secrets
- ✅ Stripe API keys (need production values)
- ✅ AWS credentials (need production values)
- ✅ Azure credentials (need production values)
- ✅ GCP credentials (need production values)
- ✅ SAP connector settings (need production values)
- ✅ Utility API configuration (need production values)
- ✅ Email/SMTP settings (need production values)
- ✅ Redis configuration
- ✅ OpenTelemetry settings
- ✅ Worker/job queue configuration
- ✅ Rate limiting settings
- ✅ AI/ML configuration
- ✅ Security secrets
- ✅ Monitoring integration placeholders

### 3. ✅ Database Migrations
Created migration tools and updated schema:
- ✅ Added `jobs` table to `infra/db/schema.sql` for job queue
- ✅ Created `scripts/migrate.sh` (Bash)
- ✅ Created `scripts/migrate.ps1` (PowerShell)
- ✅ Both scripts support automated database setup and migration

### 4. ✅ OpenTelemetry Collector Deployment
Created complete observability infrastructure:
- ✅ `infra/otel-collector-config.yaml` - Local development config
- ✅ `infra/k8s/otel-collector.yaml` - Kubernetes deployment
- ✅ `infra/prometheus.yml` - Prometheus scrape configuration
- ✅ `infra/grafana/datasources/datasources.yml` - Grafana data sources
- ✅ `infra/grafana/dashboards/dashboards.yml` - Dashboard provisioning
- ✅ `docker-compose.yml` - Complete local stack with observability

### 5. ✅ Cloud Integration Testing
Created comprehensive test scripts:
- ✅ `scripts/test-integration.sh` (Bash)
- ✅ `scripts/test-integration.ps1` (PowerShell)
- Tests for:
  - AWS CUR connector
  - Azure emissions connector
  - GCP carbon connector
  - SAP connector
  - Utility bill connector
  - Job queue system
  - Emissions calculations

### 6. ✅ Staging Deployment Scripts
Created automated deployment tools:
- ✅ `scripts/deploy-staging.sh` (Bash)
- ✅ `scripts/deploy-staging.ps1` (PowerShell)
- Features:
  - Docker image building
  - Container registry push
  - Kubernetes namespace setup
  - Resource deployment
  - Health check verification
  - Status reporting

## 📋 Next Steps - Manual Configuration Required

### 1. Configure Production API Keys
Edit `.env` and replace placeholder values:

```bash
# Stripe (get from https://dashboard.stripe.com)
STRIPE_SECRET_KEY=sk_live_YOUR_LIVE_KEY
STRIPE_PUBLISHABLE_KEY=pk_live_YOUR_LIVE_KEY
STRIPE_WEBHOOK_SECRET=whsec_YOUR_WEBHOOK_SECRET
STRIPE_PRICE_ID_STARTER=price_YOUR_STARTER_ID
STRIPE_PRICE_ID_PROFESSIONAL=price_YOUR_PRO_ID
STRIPE_PRICE_ID_ENTERPRISE=price_YOUR_ENTERPRISE_ID

# AWS (IAM credentials with CUR read access)
AWS_ACCESS_KEY_ID=AKIA...
AWS_SECRET_ACCESS_KEY=...
AWS_S3_BUCKET=your-production-bucket

# Azure (Service Principal credentials)
AZURE_TENANT_ID=...
AZURE_CLIENT_ID=...
AZURE_CLIENT_SECRET=...
AZURE_SUBSCRIPTION_ID=...

# GCP (Service Account JSON path)
GOOGLE_APPLICATION_CREDENTIALS=/path/to/production-service-account.json
GCP_PROJECT_ID=your-production-project

# SAP
SAP_BASE_URL=https://your-sap-production.com
SAP_CLIENT_ID=...
SAP_CLIENT_SECRET=...

# Utility API
UTILITY_API_KEY=...

# Email (SendGrid or your SMTP provider)
SMTP_PASSWORD=YOUR_SENDGRID_API_KEY

# Security - Generate strong secrets
JWT_SECRET=$(openssl rand -base64 48)
SESSION_SECRET=$(openssl rand -base64 48)
ENCRYPTION_KEY=$(openssl rand -base64 32 | cut -c1-32)
```

### 2. Run Database Migrations

**Windows (PowerShell):**
```powershell
cd C:\Users\pault\OffGridFlow
.\scripts\migrate.ps1
```

**Linux/Mac (Bash):**
```bash
cd /path/to/OffGridFlow
chmod +x scripts/migrate.sh
./scripts/migrate.sh
```

### 3. Start Local Development Stack

```bash
docker-compose up -d
```

This starts:
- PostgreSQL database
- Redis cache
- Jaeger (tracing)
- OpenTelemetry Collector
- Prometheus (metrics)
- Grafana (dashboards)
- OffGridFlow API
- OffGridFlow Worker
- OffGridFlow Web UI

### 4. Verify Services

```bash
# Check all services are running
docker-compose ps

# Check API health
curl http://localhost:8080/health

# Check Web UI
curl http://localhost:3000

# View Jaeger UI
open http://localhost:16686

# View Prometheus
open http://localhost:9090

# View Grafana (admin/admin)
open http://localhost:3001
```

### 5. Run Integration Tests

**Windows:**
```powershell
.\scripts\test-integration.ps1
```

**Linux/Mac:**
```bash
chmod +x scripts/test-integration.sh
./scripts/test-integration.sh
```

### 6. Deploy to Staging

**Prerequisites:**
- Kubernetes cluster configured
- Docker registry access
- kubectl configured

**Create Kubernetes secrets:**
```bash
kubectl create namespace offgridflow
kubectl create secret generic offgridflow-secrets \
  --from-env-file=.env.staging \
  -n offgridflow
```

**Deploy:**
```bash
# Windows
.\scripts\deploy-staging.ps1

# Linux/Mac
chmod +x scripts/deploy-staging.sh
./scripts/deploy-staging.sh
```

## 🔍 Monitoring & Observability

### Access Points (Local Development)

| Service | URL | Credentials |
|---------|-----|-------------|
| API | http://localhost:8080 | - |
| Web UI | http://localhost:3000 | - |
| Jaeger | http://localhost:16686 | - |
| Prometheus | http://localhost:9090 | - |
| Grafana | http://localhost:3001 | admin/admin |
| OTEL Collector | http://localhost:13133/health | - |

### Key Metrics to Monitor

1. **API Performance**
   - Request latency (p50, p95, p99)
   - Error rate
   - Requests per second

2. **Job Queue**
   - Jobs queued
   - Jobs processing
   - Jobs failed
   - Processing time

3. **Database**
   - Connection pool usage
   - Query performance
   - Active connections

4. **Ingestion**
   - Records processed
   - Success/failure rate
   - Processing duration

## 🚀 Production Readiness Checklist

### Security
- [ ] All API keys replaced with production values
- [ ] JWT secrets regenerated with strong random values
- [ ] HTTPS/TLS certificates configured
- [ ] Database credentials rotated
- [ ] Firewall rules configured
- [ ] Rate limiting enabled and tested
- [ ] CORS origins restricted to production domains

### Infrastructure
- [ ] Database backups configured
- [ ] Redis persistence configured
- [ ] Log aggregation set up (e.g., CloudWatch, Datadog)
- [ ] Metrics collection verified
- [ ] Alerting rules configured
- [ ] Auto-scaling policies set
- [ ] Disaster recovery plan documented

### Testing
- [ ] All integration tests passing
- [ ] Load testing completed
- [ ] Security audit performed
- [ ] Penetration testing completed
- [ ] Compliance validation run

### Monitoring
- [ ] All dashboards configured in Grafana
- [ ] Alert rules set in Prometheus
- [ ] On-call rotation established
- [ ] Runbooks documented
- [ ] SLO/SLA targets defined

### Documentation
- [ ] API documentation published
- [ ] Deployment runbook updated
- [ ] Architecture diagrams created
- [ ] Troubleshooting guide written
- [ ] User guides completed

## 🐛 Troubleshooting

### Database Connection Issues
```bash
# Check PostgreSQL is running
docker-compose ps postgres

# Check logs
docker-compose logs postgres

# Test connection manually
psql -h localhost -p 5432 -U offgridflow -d offgridflow
```

### API Not Starting
```bash
# Check logs
docker-compose logs api

# Verify environment variables
docker-compose exec api env | grep DATABASE_URL

# Check if port is available
netstat -an | grep 8080
```

### OpenTelemetry Not Receiving Traces
```bash
# Check OTEL Collector logs
docker-compose logs otel-collector

# Verify endpoint is accessible
curl http://localhost:13133

# Check API OTEL configuration
docker-compose exec api env | grep OTEL
```

## 📞 Support

For issues or questions:
1. Check logs: `docker-compose logs [service-name]`
2. Review configuration in `.env`
3. Verify all services are running: `docker-compose ps`
4. Check health endpoints
5. Review metrics in Grafana

## 🎯 Summary

All deployment infrastructure has been created and configured:
- ✅ Complete environment configuration
- ✅ Database migration scripts
- ✅ Full observability stack
- ✅ Integration test suite
- ✅ Automated staging deployment
- ✅ Docker Compose for local development
- ✅ Kubernetes manifests for production

**You now need to:**
1. Add production API keys to `.env`
2. Run migrations: `.\scripts\migrate.ps1`
3. Start services: `docker-compose up -d`
4. Run tests: `.\scripts\test-integration.ps1`
5. Deploy to staging: `.\scripts\deploy-staging.ps1`
