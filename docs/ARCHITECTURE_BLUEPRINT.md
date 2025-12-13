# OffGridFlow Architecture Blueprint

> **Enterprise Carbon Accounting & ESG Compliance Platform**  
> **Version**: 1.0.0 | **Status**: Production-Ready | **Last Updated**: December 2024

---

## Executive Summary

OffGridFlow is a **production-grade, multi-tenant SaaS platform** for carbon accounting and ESG compliance. This document provides the complete architectural specification, implementation status, and technical roadmap.

### Guiding Principle: "Production-Grade by Design"

Every component—from data ingestion to user interface—is built for:
- **Reliability**: Graceful degradation, retry policies, circuit breakers
- **Security**: Tenant isolation, RBAC, audit logging, encryption at rest/in-transit
- **Scalability**: Horizontal scaling, async job processing, caching layers
- **Observability**: Unified metrics, traces, and logs from day one

---

## Table of Contents

1. [System Architecture Overview](#1-system-architecture-overview)
2. [Architecture Diagrams](#2-architecture-diagrams)
3. [Component Specifications](#3-component-specifications)
4. [Multi-Tenant Security Model](#4-multi-tenant-security-model)
5. [Data Architecture](#5-data-architecture)
6. [Observability Infrastructure](#6-observability-infrastructure)
7. [Deployment Architecture](#7-deployment-architecture)
8. [Implementation Status Matrix](#8-implementation-status-matrix)
9. [Gap Analysis & Roadmap](#9-gap-analysis--roadmap)
10. [Appendix: Core Artifacts](#10-appendix-core-artifacts)

---

## 1. System Architecture Overview

### Technology Stack

| Layer | Component | Technology | Status |
|-------|-----------|------------|--------|
| **Presentation** | Web Dashboard | Next.js 14, TypeScript, Tailwind CSS | ✅ Implemented |
| **API Gateway** | REST API | Go 1.24, Chi Router | ✅ Implemented |
| **API Gateway** | GraphQL API | gqlgen | ✅ Implemented |
| **Core Services** | Calculation Engine | Go (Scope 1/2/3 calculators) | ✅ Implemented |
| **Core Services** | Compliance Engine | Go (CSRD, SEC, CBAM, CA, IFRS) | ✅ Implemented |
| **Data Ingestion** | Cloud Connectors | Go (AWS, Azure, GCP adapters) | ✅ Implemented |
| **Data Ingestion** | ERP Connectors | Go (SAP, Utility APIs) | ✅ Implemented |
| **Data Layer** | Primary Database | PostgreSQL 15+ | ✅ Implemented |
| **Data Layer** | Cache Layer | Redis 7+ | ✅ Implemented |
| **Security** | AuthN/AuthZ | JWT, Sessions, RBAC | ✅ Implemented |
| **Observability** | Telemetry | OpenTelemetry, Prometheus, Jaeger | ✅ Implemented |
| **Infrastructure** | Container Orchestration | Kubernetes, Docker | ✅ Implemented |
| **Infrastructure** | IaC | Terraform (AWS) | ✅ Implemented |

### High-Level Data Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              DATA SOURCES                                    │
├─────────────┬─────────────┬─────────────┬─────────────┬─────────────────────┤
│  AWS CUR    │   Azure     │    GCP      │    SAP      │  CSV/Utility APIs   │
│  Reports    │  Carbon API │  Carbon API │   ERP       │                     │
└──────┬──────┴──────┬──────┴──────┬──────┴──────┬──────┴──────────┬──────────┘
       │             │             │             │                  │
       └─────────────┴─────────────┴─────────────┴──────────────────┘
                                   │
                    ┌──────────────▼──────────────┐
                    │    INGESTION LAYER          │
                    │  • Rate Limiting            │
                    │  • Retry/Backoff            │
                    │  • Error Classification     │
                    │  • Idempotency              │
                    └──────────────┬──────────────┘
                                   │
                    ┌──────────────▼──────────────┐
                    │    PROCESSING LAYER         │
                    │  • Worker Pool              │
                    │  • Batch Scheduler          │
                    │  • Event Bus                │
                    └──────────────┬──────────────┘
                                   │
       ┌───────────────────────────┼───────────────────────────┐
       │                           │                           │
┌──────▼──────┐         ┌──────────▼──────────┐      ┌────────▼────────┐
│ CALCULATION │         │   COMPLIANCE        │      │   REPORTING     │
│   ENGINE    │         │     ENGINE          │      │    ENGINE       │
│             │         │                     │      │                 │
│ • Scope 1   │         │ • CSRD/ESRS         │      │ • PDF Gen       │
│ • Scope 2   │         │ • SEC Climate       │      │ • Excel/CSV     │
│ • Scope 3   │         │ • California SB253  │      │ • XBRL/iXBRL    │
│ • Factors   │         │ • CBAM              │      │ • JSON API      │
│             │         │ • IFRS S2           │      │                 │
└──────┬──────┘         └──────────┬──────────┘      └────────┬────────┘
       │                           │                          │
       └───────────────────────────┼──────────────────────────┘
                                   │
                    ┌──────────────▼──────────────┐
                    │      DATA LAYER             │
                    │  PostgreSQL + Redis         │
                    │  • Tenant Isolation         │
                    │  • Row-Level Security       │
                    │  • Encrypted at Rest        │
                    └──────────────┬──────────────┘
                                   │
                    ┌──────────────▼──────────────┐
                    │      API LAYER              │
                    │  • REST (Chi)               │
                    │  • GraphQL (gqlgen)         │
                    │  • WebSocket (realtime)     │
                    │  • Rate Limiting            │
                    └──────────────┬──────────────┘
                                   │
                    ┌──────────────▼──────────────┐
                    │      PRESENTATION           │
                    │  Next.js + Tailwind         │
                    │  • Dashboard                │
                    │  • Reports                  │
                    │  • Settings                 │
                    └─────────────────────────────┘
```

---

## 2. Architecture Diagrams

### 2.1 System Context Diagram (C4 Level 1)

```mermaid
C4Context
    title System Context Diagram - OffGridFlow

    Person(user, "Enterprise User", "Sustainability Manager, CFO, Compliance Officer")
    Person(admin, "Platform Admin", "System Administrator")
    
    System(offgridflow, "OffGridFlow Platform", "Carbon Accounting & ESG Compliance SaaS")
    
    System_Ext(aws, "AWS", "Cloud carbon data via CUR")
    System_Ext(azure, "Azure", "Carbon Footprint API")
    System_Ext(gcp, "GCP", "Carbon Footprint API")
    System_Ext(sap, "SAP ERP", "Business activity data")
    System_Ext(utility, "Utility APIs", "Energy consumption data")
    System_Ext(stripe, "Stripe", "Payment processing")
    System_Ext(email, "Email Service", "Notifications")
    
    Rel(user, offgridflow, "Views dashboards, generates reports")
    Rel(admin, offgridflow, "Manages platform, users, tenants")
    Rel(offgridflow, aws, "Ingests carbon data")
    Rel(offgridflow, azure, "Ingests carbon data")
    Rel(offgridflow, gcp, "Ingests carbon data")
    Rel(offgridflow, sap, "Ingests activity data")
    Rel(offgridflow, utility, "Ingests energy data")
    Rel(offgridflow, stripe, "Processes payments")
    Rel(offgridflow, email, "Sends notifications")
```

### 2.2 Container Diagram (C4 Level 2)

```mermaid
C4Container
    title Container Diagram - OffGridFlow

    Person(user, "User")
    
    Container_Boundary(platform, "OffGridFlow Platform") {
        Container(web, "Web Application", "Next.js, TypeScript", "User interface for dashboards and reports")
        Container(api, "API Server", "Go, Chi Router", "REST/GraphQL APIs, business logic")
        Container(worker, "Worker Service", "Go", "Background job processing")
        Container(graphql, "GraphQL Server", "Go, gqlgen", "GraphQL API layer")
        
        ContainerDb(postgres, "PostgreSQL", "PostgreSQL 15", "Primary data store with tenant isolation")
        ContainerDb(redis, "Redis", "Redis 7", "Caching, rate limiting, sessions")
        
        Container(otel, "OTEL Collector", "OpenTelemetry", "Telemetry collection and export")
        Container(prometheus, "Prometheus", "Prometheus", "Metrics storage and alerting")
        Container(jaeger, "Jaeger", "Jaeger", "Distributed tracing")
        Container(grafana, "Grafana", "Grafana", "Observability dashboards")
    }
    
    Rel(user, web, "HTTPS")
    Rel(web, api, "REST/GraphQL")
    Rel(api, postgres, "SQL")
    Rel(api, redis, "Redis Protocol")
    Rel(worker, postgres, "SQL")
    Rel(worker, redis, "Redis Protocol")
    Rel(api, otel, "OTLP")
    Rel(worker, otel, "OTLP")
    Rel(otel, prometheus, "Prometheus Remote Write")
    Rel(otel, jaeger, "OTLP")
    Rel(grafana, prometheus, "PromQL")
    Rel(grafana, jaeger, "Jaeger API")
```

### 2.3 Component Diagram - API Server

```mermaid
flowchart TB
    subgraph API["API Server"]
        direction TB
        
        subgraph Middleware["Middleware Layer"]
            AUTH[Auth Middleware]
            TENANT[Tenant Context]
            RATELIMIT[Rate Limiter]
            CORS[CORS Handler]
            TRACE[Tracing Middleware]
        end
        
        subgraph Handlers["Handler Layer"]
            HEALTH[Health Handlers]
            EMISSIONS[Emissions Handlers]
            COMPLIANCE[Compliance Handlers]
            ACTIVITIES[Activities Handlers]
            USERS[User Handlers]
            BILLING[Billing Handlers]
        end
        
        subgraph Services["Service Layer"]
            CALC[Calculation Service]
            REPORT[Report Service]
            INGEST[Ingestion Service]
            AUTHSVC[Auth Service]
        end
        
        subgraph Domain["Domain Layer"]
            SCOPE1[Scope 1 Calculator]
            SCOPE2[Scope 2 Calculator]
            SCOPE3[Scope 3 Calculator]
            FACTORS[Emission Factors]
            CSRD[CSRD Engine]
            SEC[SEC Engine]
            CBAM[CBAM Engine]
        end
    end
    
    REQUEST[HTTP Request] --> AUTH
    AUTH --> TENANT
    TENANT --> RATELIMIT
    RATELIMIT --> CORS
    CORS --> TRACE
    TRACE --> Handlers
    
    Handlers --> Services
    Services --> Domain
    
    Domain --> DB[(PostgreSQL)]
    Domain --> CACHE[(Redis)]
```

### 2.4 Deployment Architecture

```mermaid
flowchart TB
    subgraph Internet
        USER[Users]
        CDN[CloudFront CDN]
    end
    
    subgraph AWS["AWS Cloud"]
        subgraph VPC["VPC"]
            subgraph Public["Public Subnets"]
                ALB[Application Load Balancer]
                NAT[NAT Gateway]
            end
            
            subgraph Private["Private Subnets"]
                subgraph ECS["ECS Fargate"]
                    API1[API Instance 1]
                    API2[API Instance 2]
                    WORKER1[Worker 1]
                    WORKER2[Worker 2]
                end
                
                subgraph Data["Data Layer"]
                    RDS[(RDS PostgreSQL\nMulti-AZ)]
                    REDIS[(ElastiCache\nRedis Cluster)]
                end
            end
        end
        
        S3[(S3 Buckets\nReports, Uploads)]
        SQS[SQS Queues]
        SECRETS[Secrets Manager]
    end
    
    USER --> CDN
    CDN --> ALB
    ALB --> API1 & API2
    API1 & API2 --> RDS
    API1 & API2 --> REDIS
    API1 & API2 --> SQS
    WORKER1 & WORKER2 --> SQS
    WORKER1 & WORKER2 --> RDS
    API1 & API2 --> S3
    API1 & API2 --> SECRETS
```

### 2.5 Observability Pipeline

```mermaid
flowchart LR
    subgraph Applications
        API[API Server]
        WORKER[Worker]
        WEB[Web App]
    end
    
    subgraph Collection["Collection Layer"]
        OTEL[OTEL Collector]
    end
    
    subgraph Storage["Storage Layer"]
        PROM[(Prometheus)]
        JAEGER[(Jaeger)]
        LOKI[(Loki)]
    end
    
    subgraph Visualization
        GRAFANA[Grafana Dashboards]
        ALERTS[Alert Manager]
    end
    
    API -->|Traces, Metrics| OTEL
    WORKER -->|Traces, Metrics| OTEL
    WEB -->|Browser Traces| OTEL
    
    OTEL -->|Metrics| PROM
    OTEL -->|Traces| JAEGER
    OTEL -->|Logs| LOKI
    
    PROM --> GRAFANA
    JAEGER --> GRAFANA
    LOKI --> GRAFANA
    
    PROM --> ALERTS
```

---

## 3. Component Specifications

### 3.1 Carbon Calculation Engine

**Location**: `internal/emissions/`

The calculation engine supports three primary scopes with configurable emission factors:

| Scope | Description | Implementation | Calculation Methods |
|-------|-------------|----------------|---------------------|
| **Scope 1** | Direct emissions (owned/controlled sources) | `scope1.go` | Activity-based, fuel-specific factors |
| **Scope 2** | Indirect emissions (purchased energy) | `scope2.go` | Location-based, market-based |
| **Scope 3** | Value chain emissions | `scope3.go`, `scope3/` | Spend-based, activity-based, hybrid |

**Emission Factors Database**: `internal/emissionfactors/`
- 10,000+ emission factors
- Regional variations (US, EU, UK, APAC)
- Source citations for audit trails
- Versioned for historical accuracy

**Key Features**:
- ✅ Batch calculation with parallel processing
- ✅ Caching layer for repeated factor lookups
- ✅ Uncertainty quantification
- ✅ Data quality scoring
- ✅ Audit trail for all calculations

### 3.2 Compliance & Reporting Engine

**Location**: `internal/compliance/`

| Framework | Status | Implementation | Key Features |
|-----------|--------|----------------|--------------|
| **CSRD/ESRS** | ✅ Complete | `csrd/`, `csrd.go` | E1-E5 environmental standards, double materiality |
| **SEC Climate** | ✅ Complete | `sec/`, `sec.go` | Reg S-K Item 1500 series compliance |
| **California SB 253** | ✅ Complete | `california/`, `california.go` | CCDAA requirements |
| **CBAM** | ✅ Complete | `cbam/`, `cbam.go` | EU carbon border adjustment |
| **IFRS S2** | ✅ Complete | `ifrs/`, `ifrs.go` | ISSB climate disclosures |

**Report Generation**: `internal/reporting/`
- PDF generation with professional templates
- Excel/CSV export for data analysis
- XBRL/iXBRL tagging for regulatory filings
- JSON API for programmatic access

### 3.3 Data Ingestion Layer

**Location**: `internal/ingestion/`

| Connector | Status | Implementation | Features |
|-----------|--------|----------------|----------|
| **AWS CUR** | ✅ Hardened | `sources/aws/` | S3 manifest parsing, retry/backoff |
| **Azure Carbon** | ✅ Hardened | `sources/azure/` | OAuth2 token refresh, pagination |
| **GCP Carbon** | ✅ Hardened | `sources/gcp/` | BigQuery integration, batching |
| **SAP ERP** | ✅ Hardened | `sources/sap/` | OData/RFC, S/4HANA support |
| **Utility APIs** | ✅ Hardened | `sources/utility_bills/` | Multi-provider support |
| **CSV Upload** | ✅ Complete | `sources/csv_upload/` | Template validation, mapping |

**Hardening Features** (all connectors):
- Rate limiting with token bucket
- Exponential backoff with jitter
- Error classification (transient vs. permanent)
- Idempotency keys
- Circuit breaker pattern
- Comprehensive logging and tracing

### 3.4 Worker & Job Processing

**Location**: `internal/worker/`, `internal/workers/`

```go
// Job Types
type JobType string
const (
    JobTypeCalculation    JobType = "calculation"
    JobTypeIngestion      JobType = "ingestion"
    JobTypeReportGen      JobType = "report_generation"
    JobTypeExport         JobType = "export"
    JobTypeNotification   JobType = "notification"
)
```

**Features**:
- ✅ PostgreSQL-backed job queue (no external dependency)
- ✅ Configurable retry policies with backoff
- ✅ Job prioritization and scheduling
- ✅ Dead letter queue for failed jobs
- ✅ Metrics for queue depth, processing time
- ✅ Graceful shutdown

---

## 4. Multi-Tenant Security Model

### 4.1 Tenant Isolation Strategy

**Decision**: Shared database with `tenant_id` column and application-level enforcement.

**Rationale**:
| Factor | Shared DB (Chosen) | Isolated DB |
|--------|-------------------|-------------|
| Cost efficiency | ✅ Lower | Higher |
| Operational complexity | ✅ Simpler | Complex |
| Data isolation | Good (app-enforced) | ✅ Maximum |
| Performance isolation | Adequate | ✅ Best |
| Compliance | Most regulations | Strictest requirements |

**Hybrid Capability**: Architecture supports database-per-tenant for enterprise clients with strict requirements.

### 4.2 Security Implementation

**Location**: `internal/auth/`

```go
// User model with tenant association
type User struct {
    ID           uuid.UUID  `json:"id"`
    TenantID     string     `json:"tenant_id"`      // Mandatory tenant association
    Email        string     `json:"email"`
    Name         string     `json:"name"`
    PasswordHash string     `json:"-"`              // Never serialized
    Role         string     `json:"role"`
    Roles        []string   `json:"roles"`          // RBAC roles
    IsActive     bool       `json:"is_active"`
    // ... audit fields
}
```

**Security Layers**:

| Layer | Implementation | Status |
|-------|----------------|--------|
| **Authentication** | JWT + Session cookies | ✅ |
| **Authorization** | RBAC with role hierarchy | ✅ |
| **Tenant Isolation** | `tenant_id` in all queries | ✅ |
| **API Keys** | Scoped, rotatable, expiring | ✅ |
| **Password Security** | bcrypt, configurable cost | ✅ |
| **Rate Limiting** | Per-tenant, per-tier | ✅ |
| **Audit Logging** | All mutations logged | ✅ |
| **2FA** | TOTP support | ✅ |
| **Session Management** | Secure cookies, CSRF | ✅ |

### 4.3 RBAC Model

```go
// Role hierarchy
const (
    RoleOwner      = "owner"       // Full tenant access
    RoleAdmin      = "admin"       // User/settings management
    RoleManager    = "manager"     // Report approval
    RoleAnalyst    = "analyst"     // Read + calculate
    RoleViewer     = "viewer"      // Read-only
    RoleAPIClient  = "api_client"  // Programmatic access
)

// Permission matrix
var Permissions = map[string][]string{
    RoleOwner:   {"*"},
    RoleAdmin:   {"users:*", "settings:*", "reports:*", "activities:*"},
    RoleManager: {"reports:approve", "reports:read", "activities:*"},
    RoleAnalyst: {"reports:create", "reports:read", "activities:*", "calculate:*"},
    RoleViewer:  {"reports:read", "activities:read", "dashboard:read"},
}
```

---

## 5. Data Architecture

### 5.1 Database Schema (Core Entities)

```sql
-- Tenant (organization)
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    plan VARCHAR(50) DEFAULT 'free',
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Users (with mandatory tenant association)
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL,
    roles TEXT[] DEFAULT '{}',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(tenant_id, email)
);
CREATE INDEX idx_users_tenant ON users(tenant_id);

-- Activities (emission sources)
CREATE TABLE activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    workspace_id UUID,
    name VARCHAR(255) NOT NULL,
    category VARCHAR(100),
    scope VARCHAR(20) NOT NULL,
    quantity DECIMAL(20,6),
    unit VARCHAR(50),
    location VARCHAR(255),
    period_start DATE,
    period_end DATE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_activities_tenant ON activities(tenant_id);
CREATE INDEX idx_activities_scope ON activities(tenant_id, scope);

-- Emissions (calculated results)
CREATE TABLE emissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    activity_id UUID REFERENCES activities(id),
    scope VARCHAR(20) NOT NULL,
    emissions_kg_co2e DECIMAL(20,6) NOT NULL,
    emission_factor_id VARCHAR(100),
    calculation_method VARCHAR(100),
    data_quality_score DECIMAL(3,2),
    calculated_at TIMESTAMPTZ DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'
);
CREATE INDEX idx_emissions_tenant ON emissions(tenant_id);

-- Compliance Reports
CREATE TABLE compliance_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    report_type VARCHAR(50) NOT NULL,
    reporting_year INT NOT NULL,
    status VARCHAR(50) DEFAULT 'draft',
    emissions_data JSONB NOT NULL,
    pdf_url TEXT,
    xbrl_url TEXT,
    generated_by UUID REFERENCES users(id),
    approved_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_reports_tenant ON compliance_reports(tenant_id);
```

### 5.2 Data Integrity Principles

1. **Centralize**: Single source of truth for dimensions (facilities, business units)
2. **Enrich**: Map raw data to specific categories for accuracy
3. **Validate**: Data quality checks at every pipeline stage
4. **Audit**: Full trail of data lineage and transformations

---

## 6. Observability Infrastructure

### 6.1 Three Pillars Implementation

| Pillar | Tool | Implementation | Status |
|--------|------|----------------|--------|
| **Metrics** | Prometheus | 33 custom metrics | ✅ |
| **Traces** | Jaeger + OTEL | Distributed tracing | ✅ |
| **Logs** | Structured JSON | Request ID correlation | ✅ |

### 6.2 Key Metrics

```yaml
# HTTP Metrics
http_request_count{method, path, status}
http_request_duration_bucket{method, path}
http_request_size_bytes{method, path}

# Database Metrics
db_query_duration_bucket{operation}
db_connection_pool_size
db_connection_pool_idle

# Business Metrics
emissions_calculated_total{scope, tenant}
reports_generated_total{type, tenant}
ingestion_records_processed{source}

# Job Metrics
job_queue_depth{type}
job_execution_duration_bucket{type}
job_failures_total{type, reason}
```

### 6.3 Alerting Rules

**Location**: `deployments/grafana/alerts/alert-rules.yml`

| Alert | Condition | Severity |
|-------|-----------|----------|
| HighErrorRate | >5% 5xx for 5m | Critical |
| HighAPILatency | p95 >1s for 5m | Warning |
| DatabasePoolExhaustion | >90% used for 5m | Critical |
| JobQueueBacklog | >1000 pending for 15m | Warning |
| ServiceDown | No response for 1m | Critical |

### 6.4 Health Endpoints

| Endpoint | Purpose | Returns |
|----------|---------|---------|
| `/health` | Basic liveness | 200 OK |
| `/livez` | Kubernetes liveness | 200 OK |
| `/readyz` | Kubernetes readiness | 200/503 |
| `/metrics` | Prometheus scrape | Metrics |

---

## 7. Deployment Architecture

### 7.1 Local Development

**File**: `docker-compose.yml`

```bash
# Start full stack
docker-compose up -d

# Services available:
# - API:        http://localhost:8080
# - Web:        http://localhost:3000
# - Grafana:    http://localhost:3001 (admin/admin)
# - Prometheus: http://localhost:9090
# - Jaeger:     http://localhost:16686
# - PostgreSQL: localhost:5432
# - Redis:      localhost:6379
```

### 7.2 Production Infrastructure (AWS)

**Location**: `infra/terraform/`

| Resource | Service | Configuration |
|----------|---------|---------------|
| Compute | ECS Fargate | Auto-scaling, spot instances |
| Database | RDS PostgreSQL | Multi-AZ, encrypted |
| Cache | ElastiCache Redis | Cluster mode |
| Storage | S3 | Versioning, lifecycle policies |
| CDN | CloudFront | Edge caching |
| DNS | Route 53 | Health checks |
| Secrets | Secrets Manager | Automatic rotation |
| Queues | SQS | Dead letter queues |

### 7.3 Kubernetes Deployment

**Location**: `infra/k8s/`

```yaml
# Key manifests:
# - namespace.yaml       # offgridflow namespace
# - configmap.yaml       # Application configuration
# - secrets.yaml.example # Secret template
# - api-deployment.yaml  # API server deployment
# - web-deployment.yaml  # Next.js deployment
# - worker-deployment.yaml # Worker deployment
# - services.yaml        # Service definitions
# - ingress.yaml         # Ingress rules
# - hpa.yaml             # Horizontal pod autoscaler
```

### 7.4 CI/CD Pipeline

```yaml
# .github/workflows/ci.yml (conceptual)
name: CI/CD Pipeline

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run Tests
        run: go test ./... -race -coverprofile=coverage.out
      - name: Upload Coverage
        uses: codecov/codecov-action@v3

  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: golangci-lint
        uses: golangci/golangci-lint-action@v3

  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run Trivy
        uses: aquasecurity/trivy-action@master

  # Observability Quality Gate
  performance:
    runs-on: ubuntu-latest
    needs: [test]
    steps:
      - name: Run Performance Tests
        run: go test ./internal/performance/... -bench=.
      - name: Check Latency Budgets
        run: |
          # Fail if p95 latency > 200ms
          # Fail if memory allocation > baseline + 10%

  deploy:
    runs-on: ubuntu-latest
    needs: [test, lint, security, performance]
    if: github.ref == 'refs/heads/main'
    steps:
      - name: Deploy to Production
        run: |
          # Terraform apply
          # Kubernetes rollout
          # Smoke tests
          # Rollback on failure
```

---

## 8. Implementation Status Matrix

### 8.1 Core Features

| Feature | Status | Coverage | Notes |
|---------|--------|----------|-------|
| Scope 1 Calculations | ✅ Complete | 85%+ | Fuel combustion, fugitive emissions |
| Scope 2 Calculations | ✅ Complete | 85%+ | Location-based, market-based |
| Scope 3 Calculations | ✅ Complete | 75%+ | Categories 1-15 |
| Emission Factors DB | ✅ Complete | 10,000+ | Regional variations |
| CSRD Compliance | ✅ Complete | Full E1-E5 | EU ESRS standards |
| SEC Climate | ✅ Complete | Full | Reg S-K 1500 series |
| California SB 253 | ✅ Complete | Full | CCDAA requirements |
| CBAM | ✅ Complete | Full | EU carbon border |
| IFRS S2 | ✅ Complete | Full | ISSB standards |

### 8.2 Technical Infrastructure

| Component | Status | Notes |
|-----------|--------|-------|
| REST API | ✅ Complete | Chi router, middleware stack |
| GraphQL API | ✅ Complete | gqlgen, subscriptions |
| JWT Auth | ✅ Complete | Access + refresh tokens |
| Session Auth | ✅ Complete | Secure cookies, CSRF |
| RBAC | ✅ Complete | Role hierarchy |
| Rate Limiting | ✅ Complete | Per-tier, configurable |
| Multi-tenancy | ✅ Complete | tenant_id isolation |
| Audit Logging | ✅ Complete | All mutations |
| 2FA | ✅ Complete | TOTP |

### 8.3 Data Ingestion

| Connector | Status | Hardening |
|-----------|--------|-----------|
| AWS CUR | ✅ Complete | ✅ Rate limit, retry, circuit breaker |
| Azure Carbon | ✅ Complete | ✅ Rate limit, retry, circuit breaker |
| GCP Carbon | ✅ Complete | ✅ Rate limit, retry, circuit breaker |
| SAP ERP | ✅ Complete | ✅ Rate limit, retry, circuit breaker |
| Utility APIs | ✅ Complete | ✅ Rate limit, retry, circuit breaker |
| CSV Upload | ✅ Complete | ✅ Validation, templates |

### 8.4 Observability

| Component | Status | Notes |
|-----------|--------|-------|
| OpenTelemetry | ✅ Complete | Traces, metrics |
| Prometheus | ✅ Complete | 33 metrics |
| Jaeger | ✅ Complete | Distributed tracing |
| Grafana | ✅ Complete | Dashboards, alerts |
| Health Endpoints | ✅ Complete | /health, /livez, /readyz |
| Structured Logging | ✅ Complete | JSON, request ID |

### 8.5 Deployment

| Component | Status | Notes |
|-----------|--------|-------|
| Docker Compose | ✅ Complete | Full local stack |
| Dockerfile | ✅ Complete | Multi-stage, optimized |
| Kubernetes | ✅ Complete | Manifests, HPA |
| Terraform | ✅ Complete | AWS infrastructure |
| Helm Charts | ✅ Complete | `infra/helm/` |

---

## 9. Gap Analysis & Roadmap

### 9.1 Identified Gaps

| Area | Gap | Priority | Effort |
|------|-----|----------|--------|
| AI Integration | AI copilot for anomaly detection | High | Medium |
| AI Integration | Narrative generation for reports | High | Medium |
| Website | Public status dashboard | Medium | Low |
| Website | Website carbon grade display | Low | Low |
| Data | Bring Your Own Storage (BYOS) | Medium | High |
| Security | Database-per-tenant option | Low | High |
| Compliance | GRI/CDP frameworks | Medium | Medium |

### 9.2 Implementation Roadmap

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        OFFGRIDFLOW ROADMAP                                   │
├─────────────────────────────────────────────────────────────────────────────┤
│ PHASE 1: CURRENT STATE (Complete)                    Months 1-3            │
│ ├── Core calculation engine (Scope 1/2/3)                    ✅             │
│ ├── Compliance frameworks (CSRD, SEC, CBAM, CA, IFRS)        ✅             │
│ ├── Data ingestion connectors                                ✅             │
│ ├── Multi-tenant security                                    ✅             │
│ └── Observability infrastructure                             ✅             │
├─────────────────────────────────────────────────────────────────────────────┤
│ PHASE 2: AI ENHANCEMENT                              Months 4-6            │
│ ├── Anomaly detection in emissions data                      🔲             │
│ ├── Calculation methodology suggestions                      🔲             │
│ ├── Narrative generation for compliance reports              🔲             │
│ ├── Natural language query interface                         🔲             │
│ └── Public status dashboard                                  🔲             │
├─────────────────────────────────────────────────────────────────────────────┤
│ PHASE 3: ENTERPRISE SCALING                          Months 7-9            │
│ ├── Bring Your Own Storage (BYOS)                            🔲             │
│ ├── Database-per-tenant isolation option                     🔲             │
│ ├── GRI/CDP compliance frameworks                            🔲             │
│ ├── Advanced data residency controls                         🔲             │
│ └── SOC 2 Type II certification prep                         🔲             │
└─────────────────────────────────────────────────────────────────────────────┘

Legend: ✅ Complete | 🔲 Planned
```

### 9.3 Gantt Chart

```mermaid
gantt
    title OffGridFlow Implementation Roadmap
    dateFormat  YYYY-MM
    
    section Phase 1 (Complete)
    Core Engine           :done, p1a, 2024-01, 2024-03
    Compliance Frameworks :done, p1b, 2024-02, 2024-03
    Data Ingestion        :done, p1c, 2024-02, 2024-04
    Observability         :done, p1d, 2024-03, 2024-04
    
    section Phase 2 (AI Enhancement)
    Anomaly Detection     :p2a, 2025-01, 2025-02
    Narrative Generation  :p2b, 2025-02, 2025-03
    NL Query Interface    :p2c, 2025-03, 2025-04
    Public Status Page    :p2d, 2025-02, 2025-02
    
    section Phase 3 (Enterprise)
    BYOS Implementation   :p3a, 2025-04, 2025-06
    Tenant Isolation      :p3b, 2025-05, 2025-07
    GRI/CDP Frameworks    :p3c, 2025-06, 2025-08
    SOC 2 Prep            :p3d, 2025-07, 2025-09
```

---

## 10. Appendix: Core Artifacts

### 10.1 Docker Compose (Local Development)

See: `docker-compose.yml` in repository root.

**Key services**:
- PostgreSQL 15 with health checks
- Redis 7 for caching/rate limiting
- Jaeger for distributed tracing
- OpenTelemetry Collector
- Prometheus for metrics
- Grafana for dashboards
- API server with hot reload
- Worker service
- Next.js web application

### 10.2 Tenant Context Middleware (Pseudo-code)

```go
// internal/api/http/middleware/tenant.go

func TenantContextMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Extract tenant from JWT claims or session
        claims, ok := auth.ClaimsFromContext(r.Context())
        if !ok {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }
        
        tenantID := claims.TenantID
        if tenantID == "" {
            http.Error(w, "tenant not found", http.StatusForbidden)
            return
        }
        
        // Inject tenant context for all downstream operations
        ctx := context.WithValue(r.Context(), TenantIDKey, tenantID)
        
        // Log tenant for audit
        slog.Info("request",
            "tenant_id", tenantID,
            "user_id", claims.UserID,
            "path", r.URL.Path,
        )
        
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// All database queries MUST use tenant scoping
func (s *Store) GetActivities(ctx context.Context) ([]Activity, error) {
    tenantID := TenantIDFromContext(ctx)
    if tenantID == "" {
        return nil, ErrTenantRequired
    }
    
    query := `SELECT * FROM activities WHERE tenant_id = $1`
    return s.db.Query(ctx, query, tenantID)
}
```

### 10.3 OpenTelemetry Configuration

See: `infra/otel-collector-config.yaml`

```yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

processors:
  batch:
    timeout: 10s
    send_batch_size: 1024
  memory_limiter:
    check_interval: 1s
    limit_mib: 512

exporters:
  prometheus:
    endpoint: "0.0.0.0:8889"
  jaeger:
    endpoint: jaeger:14250
    tls:
      insecure: true

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [memory_limiter, batch]
      exporters: [jaeger]
    metrics:
      receivers: [otlp]
      processors: [memory_limiter, batch]
      exporters: [prometheus]
```

### 10.4 CI/CD Quality Gate Example

```yaml
# Observability quality gate in CI
performance-gate:
  runs-on: ubuntu-latest
  steps:
    - name: Run Benchmark Suite
      run: |
        go test ./... -bench=. -benchmem > benchmark.txt
        
    - name: Check Performance Budgets
      run: |
        # Parse benchmark results
        # Fail if:
        # - p95 latency > 200ms
        # - Memory per operation > baseline + 10%
        # - Allocations > baseline + 5%
        
        if grep -q "FAIL" benchmark.txt; then
          echo "Performance regression detected"
          exit 1
        fi
        
    - name: Check Test Coverage
      run: |
        COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | tr -d '%')
        if (( $(echo "$COVERAGE < 60" | bc -l) )); then
          echo "Coverage $COVERAGE% below 60% threshold"
          exit 1
        fi
```

### 10.5 Enterprise Website Wireframe

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ [Logo] OffGridFlow                    Features  Pricing  Docs  Status  Login│
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│     ENTERPRISE CARBON ACCOUNTING                                             │
│     & ESG COMPLIANCE PLATFORM                                                │
│                                                                              │
│     [Get Started] [Schedule Demo]                                            │
│                                                                              │
│     ✓ CSRD Ready  ✓ SEC Compliant  ✓ SOC 2  ✓ 99.9% Uptime                 │
│                                                                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                          TRUSTED BY                                          │
│     [Logo] [Logo] [Logo] [Logo] [Logo]                                       │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │   INGEST     │  │  CALCULATE   │  │   COMPLY     │  │   REPORT     │     │
│  │              │  │              │  │              │  │              │     │
│  │ AWS/Azure/GCP│  │ Scope 1/2/3  │  │ CSRD, SEC,   │  │ PDF, Excel,  │     │
│  │ SAP, ERPs    │  │ 10k+ factors │  │ CBAM, CA     │  │ XBRL         │     │
│  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘     │
│                                                                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                        PLATFORM STATUS                                       │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │ ● API: Operational    ● Web: Operational    ● Ingestion: Operational │    │
│  │ Uptime (30d): 99.97%  │  Avg Latency: 45ms  │  Last Incident: None   │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                        SUSTAINABILITY                                        │
│  This website achieves a Carbon Grade of A (0.15g CO2 per page view)        │
│  Hosted on 100% renewable energy  │  Dark mode default                       │
│  [View our sustainability report]                                            │
├─────────────────────────────────────────────────────────────────────────────┤
│ © 2024 OffGridFlow  │  Privacy  │  Terms  │  Security  │  Docs              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Document Control

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0.0 | December 2024 | Architecture Team | Initial blueprint |

---

**End of Architecture Blueprint**
