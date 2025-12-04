# OffGridFlow

**Enterprise Carbon Accounting & ESG Compliance Platform**

[![Production Ready](https://img.shields.io/badge/production-ready-green.svg)](./PRODUCTION_COMPLETE_FINAL.md)
[![Test Coverage](https://img.shields.io/badge/coverage-60%25+-green.svg)](./scripts/test-all.ps1)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](./LICENSE)

> **Status**: ✅ Production Ready - Version 1.0.0

## 🎯 What is OffGridFlow?

OffGridFlow is a complete, production-ready platform for:
- 📊 **Carbon Accounting** - Track Scope 1, 2, and 3 emissions
- 🌍 **ESG Compliance** - CSRD, SEC Climate, CBAM, California, and more
- ☁️ **Cloud Integration** - AWS, Azure, GCP carbon data ingestion
- 🏢 **Enterprise Connectors** - SAP, Utility APIs, CSV imports
- 📈 **Real-time Analytics** - Dashboards, trends, and insights
- 📋 **Regulatory Reporting** - XBRL, PDF, Excel exports
- 🔐 **Enterprise Security** - Multi-tenant, RBAC, 2FA, audit logs

## 💎 Why OffGridFlow?

**Reliable cloud ingestion for AWS/Azure/GCP.**  
Automated pipelines pull carbon data from AWS Cost & Usage Reports, Azure Carbon Footprint API, and GCP Carbon Footprint API—with built-in retry logic, idempotency, and observability.

**Fully wired compliance frameworks.**  
CSRD/ESRS, SEC Climate, CBAM, California SB 253, and IFRS S2 are embedded in the data model, validation rules, and reporting flows—no manual mapping required.

**Cleanly matching frontend↔backend auth flows.**  
Next.js sessions, API tokens, and RBAC share the same JWT claims and contracts. Login, refresh, and logout are enforced consistently across web and API layers.

**Confident infra (push button "prod" deploy).**  
Run `scripts\deploy-complete.ps1` to execute pre-flight checks, database migrations, Docker builds, and Kubernetes rollouts in one repeatable flow.

## 🚀 Quick Start (5 Minutes)

```powershell
# 1. Start services
docker-compose up -d

# 2. Run migrations
.\scripts\migrate.ps1

# 3. Open browser
Start-Process http://localhost:3000
```

**Full guide**: [QUICKSTART.md](./QUICKSTART.md)

## 📚 Documentation

| Document | Description |
|----------|-------------|
| [📖 Quick Start](./QUICKSTART.md) | Get running in 5 minutes |
| [🚀 Production Deployment](./PRODUCTION_DEPLOYMENT_GUIDE.md) | Complete deployment guide |
| [✅ Final Checklist](./FINAL_CHECKLIST.md) | All completed tasks |
| [🎉 Production Complete](./PRODUCTION_COMPLETE_FINAL.md) | Implementation summary |
| [📊 API Documentation](http://localhost:8080/api/v1/docs) | Interactive API docs |

## ✨ Key Features

### Carbon Accounting
- ✅ Scope 1, 2, 3 emissions calculations
- ✅ 10,000+ emission factors database
- ✅ Regional variations (US, EU, UK, etc.)
- ✅ Activity-based and spend-based methods
- ✅ Real-time calculation engine

### Compliance Frameworks
- ✅ **CSRD/ESRS** - EU Corporate Sustainability Reporting
- ✅ **SEC Climate** - US Securities regulations
- ✅ **CBAM** - Carbon Border Adjustment Mechanism
- ✅ **California Climate** - State-level reporting
- ✅ **IFRS S2** - Sustainability disclosure
- ✅ **GRI, CDP** - Voluntary frameworks

### Data Ingestion
- ✅ **AWS CUR** - Cost and Usage Reports
- ✅ **Azure** - Carbon Footprint API
- ✅ **GCP** - Carbon Footprint API
- ✅ **SAP** - ERP integration
- ✅ **Utility APIs** - Energy bills
- ✅ **CSV** - Bulk imports

### Exports & Reporting
- ✅ **XBRL/iXBRL** - Regulatory filings
- ✅ **PDF** - Professional reports
- ✅ **Excel/CSV** - Data exports
- ✅ **JSON API** - Programmatic access

## 🏗️ Architecture

```
┌─────────────────┐
│   Next.js Web   │ ← Users interact here
└────────┬────────┘
         │ HTTPS
┌────────▼────────┐
│   API Server    │ ← REST/JSON API (Go)
└────────┬────────┘
         │
    ┌────┴────┬─────────┬──────────┐
    │         │         │          │
┌───▼──┐  ┌──▼──┐  ┌───▼───┐  ┌──▼──┐
│ PG   │  │Redis│  │Workers│  │OTEL │
│ SQL  │  │Cache│  │ Jobs  │  │Trace│
└──────┘  └─────┘  └───────┘  └─────┘
```

## 🛠️ Technology Stack

- **Backend**: Go 1.24
- **Frontend**: Next.js 14, React, TypeScript
- **Database**: PostgreSQL 15+
- **Cache**: Redis 7+
- **Observability**: OpenTelemetry, Jaeger, Prometheus, Grafana
- **Infrastructure**: Docker, Kubernetes, Terraform
- **Cloud**: AWS, Azure, GCP

## 🏗️ Infrastructure & DevOps

**One-command local setup**:
```bash
make start
# or: ./scripts/dev-start.sh (Linux/macOS)
# or: .\scripts\dev-start.ps1 (Windows)
```

**Features**:
- ✅ **Local Development**: Docker Compose with full observability stack
- ✅ **Container Images**: Optimized multi-stage builds (API: 30MB, Worker: 25MB)
- ✅ **Kubernetes**: Production-ready manifests with auto-scaling and migrations
- ✅ **Terraform**: Complete AWS infrastructure (VPC, RDS, Redis, S3, SQS, ECS)
- ✅ **CI/CD**: Automated testing, building, and deployment via GitHub Actions

**Documentation**:
- [📖 Infrastructure Guide](./docs/INFRASTRUCTURE.md) - Comprehensive deployment guide
- [✅ Infrastructure Complete](./INFRASTRUCTURE_100_COMPLETE.md) - Detailed completion report
- [🔍 Verification Checklist](./INFRASTRUCTURE_VERIFICATION.md) - Testing checklist

**Deploy to Kubernetes**:
```bash
make k8s-deploy
```

**Provision AWS Infrastructure**:
```bash
cd infra/terraform
terraform apply
```

## 🧪 Testing

```powershell
# All tests with coverage
.\scripts\test-all.ps1 -Coverage

# Unit tests only
go test ./...

# Integration tests
.\scripts\test-integration.ps1
```

**Current Coverage**: 60%+ overall (Auth: 85%+, Emissions: 75%+, Handlers: 70%+)

## 📊 Monitoring

Access these dashboards after running `docker-compose up -d`:

- **Web App**: http://localhost:3000
- **API**: http://localhost:8080
- **Grafana**: http://localhost:3001 (admin/admin)
- **Jaeger**: http://localhost:16686
- **Prometheus**: http://localhost:9090

## 🚀 Deployment

```powershell
# Validate configuration
.\scripts\deployment-checklist.ps1 -EnvFile .env.production

# Deploy to production
.\scripts\deploy-complete.ps1 -Environment production
```

## 📝 License

MIT License

---

**Made with ❤️ for a sustainable future** 🌍
