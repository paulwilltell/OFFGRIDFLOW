# OffGridFlow - Quick Start Guide

## 🚀 Get Started in 5 Minutes

### Prerequisites
- Docker Desktop installed
- Git installed
- 8GB RAM minimum

### Step 1: Clone and Setup (1 min)

```powershell
# Already in the directory
cd C:\Users\pault\OffGridFlow

# Copy environment file
Copy-Item .env.production.template .env
```

### Step 2: Start Services (2 min)

```powershell
# Start all services
docker-compose up -d

# Wait for services to be healthy (30 seconds)
Start-Sleep -Seconds 30

# Check status
docker-compose ps
```

### Step 3: Run Migrations (1 min)

```powershell
# Apply database schema
.\scripts\migrate.ps1
```

### Step 4: Verify (1 min)

```powershell
# Check API health
Invoke-WebRequest http://localhost:8080/health

# Check web app
Start-Process http://localhost:3000
```

## 📍 Access Points

| Service | URL | Credentials |
|---------|-----|-------------|
| **Web App** | http://localhost:3000 | Sign up |
| **API** | http://localhost:8080 | See docs |
| **API Docs** | http://localhost:8080/api/v1/docs | - |
| **Grafana** | http://localhost:3001 | admin/admin |
| **Jaeger** | http://localhost:16686 | - |
| **Prometheus** | http://localhost:9090 | - |

## 🎯 Try These Features

### 1. Calculate Emissions

```powershell
$body = @{
    scope = "scope1"
    category = "stationary_combustion"
    fuel_type = "natural_gas"
    quantity = 1000
    unit = "therms"
} | ConvertTo-Json

Invoke-WebRequest -Uri http://localhost:8080/api/v1/emissions/calculate `
    -Method Post `
    -Body $body `
    -ContentType "application/json" `
    -Headers @{"X-Tenant-ID"="default"}
```

### 2. View Jobs

```powershell
# List all jobs
Invoke-WebRequest http://localhost:8080/api/v1/jobs?tenant_id=default
```

### 3. Export Reports

```powershell
# Generate XBRL report
Invoke-WebRequest -Uri http://localhost:8080/api/v1/export/xbrl `
    -Method Post `
    -Headers @{"X-Tenant-ID"="default"}
```

## 🛠️ Useful Commands

### View Logs
```powershell
# API logs
docker-compose logs -f api

# Worker logs
docker-compose logs -f worker

# All logs
docker-compose logs -f
```

### Restart Services
```powershell
# Restart API
docker-compose restart api

# Restart all
docker-compose restart
```

### Stop Everything
```powershell
docker-compose down
```

### Run Tests
```powershell
# All tests
.\scripts\test-all.ps1

# Unit tests only
go test ./...

# With coverage
.\scripts\test-all.ps1 -Coverage
```

## 🔧 Configuration

### Environment Variables
Edit `.env` to configure:

```env
# Database
DATABASE_URL=postgresql://...

# APIs
STRIPE_SECRET_KEY=sk_test_...
AWS_ACCESS_KEY_ID=...

# Features
ENABLE_2FA=true
RATE_LIMIT_REQUESTS_PER_MINUTE=100
```

### Add Cloud Account

1. Open http://localhost:3000
2. Sign in or create account
3. Go to Settings → Connectors
4. Add AWS/Azure/GCP credentials
5. Sync data

## 📊 Monitoring

### Grafana Dashboards
1. Go to http://localhost:3001
2. Login: admin/admin
3. Browse dashboards:
   - OffGridFlow Overview
   - API Performance
   - Emissions Metrics
   - Job Queue Status

### Jaeger Traces
1. Go to http://localhost:16686
2. Select service: offgridflow-api
3. View distributed traces

## 🐛 Troubleshooting

### API Not Starting
```powershell
# Check logs
docker-compose logs api

# Restart
docker-compose restart api
```

### Database Connection Error
```powershell
# Check if database is running
docker-compose ps postgres

# Restart database
docker-compose restart postgres

# Check migrations
.\scripts\migrate.ps1
```

### Port Already in Use
```powershell
# Change port in docker-compose.yml
# Edit ports section for conflicting service
```

### Clean Start
```powershell
# Stop and remove everything
docker-compose down -v

# Remove old images
docker-compose rm -f

# Start fresh
docker-compose up -d
.\scripts\migrate.ps1
```

## 📚 Learn More

- [Production Deployment Guide](./PRODUCTION_DEPLOYMENT_GUIDE.md)
- [API Documentation](http://localhost:8080/api/v1/docs)
- [Architecture Overview](./docs/architecture/)
- [Contributing Guide](./CONTRIBUTING.md)

## 🆘 Get Help

- Check logs: `docker-compose logs -f`
- Run health checks: `.\scripts\deployment-checklist.ps1`
- View metrics in Grafana
- Search issues on GitHub

## 🎉 What's Next?

1. ✅ System running locally
2. 📊 Explore the dashboards
3. 🔌 Connect your cloud accounts
4. 📈 Calculate emissions
5. 📋 Generate compliance reports
6. 🚀 Deploy to production

---

**Need help?** Check the [Full Documentation](./docs/) or [Deployment Guide](./PRODUCTION_DEPLOYMENT_GUIDE.md)

**Ready for production?** Run `.\scripts\deploy-complete.ps1 -Environment production`
