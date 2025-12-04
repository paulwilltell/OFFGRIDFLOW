# OffGridFlow REST API Reference

This document provides complete REST API documentation for the OffGridFlow carbon accounting platform.

## Base URL

```
Production: https://api.offgridflow.com/v1
Development: http://localhost:8080/api
```

## Authentication

All API requests require authentication via JWT Bearer token.

```http
Authorization: Bearer <access_token>
```

### Obtain Access Token

```http
POST /api/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123",
  "tenant_id": "tenant-uuid"
}
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_in": 3600,
  "user": {
    "id": "user-uuid",
    "email": "user@example.com",
    "name": "John Doe",
    "role": "admin"
  }
}
```

### Refresh Token

```http
POST /api/auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

---

## Emissions

### Get Emissions Summary

```http
GET /api/emissions/summary
```

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| start_date | string | Start date (YYYY-MM-DD) |
| end_date | string | End date (YYYY-MM-DD) |

**Response:**
```json
{
  "total_emissions_tco2e": 1250.75,
  "scope1": {
    "total": 450.25,
    "categories": {
      "stationary_combustion": 200.5,
      "mobile_combustion": 150.25,
      "fugitive_emissions": 99.5
    }
  },
  "scope2": {
    "total": 300.50,
    "location_based": 320.00,
    "market_based": 280.00
  },
  "scope3": {
    "total": 500.00,
    "categories": {
      "purchased_goods": 200.00,
      "business_travel": 75.00,
      "employee_commuting": 50.00,
      "upstream_transport": 175.00
    }
  },
  "period": {
    "start": "2024-01-01",
    "end": "2024-12-31"
  }
}
```

### List Scope 2 Emissions

```http
GET /api/emissions/scope2
```

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| page | integer | Page number (default: 1) |
| per_page | integer | Items per page (default: 20, max: 100) |
| region | string | Filter by region |
| start_date | string | Filter by start date |
| end_date | string | Filter by end date |

**Response:**
```json
{
  "data": [
    {
      "id": "emission-uuid",
      "tenant_id": "tenant-uuid",
      "region": "us-west-2",
      "source_type": "electricity",
      "quantity_kwh": 10000,
      "emissions_tco2e": 4.5,
      "calculation_method": "location-based",
      "period_start": "2024-01-01",
      "period_end": "2024-01-31",
      "created_at": "2024-02-01T10:30:00Z"
    }
  ],
  "page_info": {
    "page": 1,
    "per_page": 20,
    "total_items": 156,
    "total_pages": 8
  }
}
```

### Create Scope 2 Emission

```http
POST /api/emissions/scope2
Content-Type: application/json

{
  "region": "us-west-2",
  "source_type": "electricity",
  "quantity_kwh": 10000,
  "calculation_method": "location-based",
  "period_start": "2024-01-01",
  "period_end": "2024-01-31"
}
```

**Response:** `201 Created`
```json
{
  "id": "emission-uuid",
  "emissions_tco2e": 4.5,
  "emission_factor_used": 0.00045,
  "created_at": "2024-02-01T10:30:00Z"
}
```

### Get Scope 2 Summary

```http
GET /api/emissions/scope2/summary
```

**Response:**
```json
{
  "total_emissions_tco2e": 300.50,
  "total_kwh": 667778,
  "average_emission_factor": 0.00045,
  "activity_count": 45,
  "location_based_total": 320.00,
  "market_based_total": 280.00
}
```

### List Scope 3 Emissions

```http
GET /api/emissions/scope3
```

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| category | string | Scope 3 category (1-15) |
| supplier_id | string | Filter by supplier |

**Response:**
```json
{
  "data": [
    {
      "id": "emission-uuid",
      "category": 1,
      "category_name": "Purchased Goods and Services",
      "supplier_id": "supplier-uuid",
      "supplier_name": "Acme Corp",
      "emissions_tco2e": 150.25,
      "data_quality_score": 0.85,
      "period_start": "2024-01-01",
      "period_end": "2024-12-31"
    }
  ],
  "page_info": {
    "page": 1,
    "per_page": 20,
    "total_items": 89
  }
}
```

---

## Compliance

### Get Compliance Summary

```http
GET /api/compliance/summary
```

**Response:**
```json
{
  "totals": {
    "scope1_tons": 450.25,
    "scope2_tons": 300.50,
    "scope3_tons": 500.00
  },
  "frameworks": [
    {
      "name": "CSRD",
      "status": "partial",
      "completion": 75,
      "deadline": "2025-01-01",
      "missing_requirements": [
        "Double materiality assessment",
        "Biodiversity disclosures"
      ]
    },
    {
      "name": "SEC Climate",
      "status": "not_started",
      "completion": 0,
      "deadline": "2026-01-01"
    }
  ]
}
```

### Get Framework Status

```http
GET /api/compliance/{framework}
```

**Path Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| framework | string | Framework code (csrd, sec, cbam, ifrs) |

**Response:**
```json
{
  "framework": "csrd",
  "name": "Corporate Sustainability Reporting Directive",
  "status": "partial",
  "completion_percentage": 75,
  "requirements": [
    {
      "id": "req-1",
      "name": "E1: Climate Change",
      "status": "complete",
      "evidence_count": 5
    },
    {
      "id": "req-2",
      "name": "E2: Pollution",
      "status": "in_progress",
      "evidence_count": 2
    }
  ],
  "deadline": "2025-01-01",
  "last_updated": "2024-06-15T14:30:00Z"
}
```

### Generate Compliance Report

```http
POST /api/compliance/{framework}/report
Content-Type: application/json

{
  "period_start": "2024-01-01",
  "period_end": "2024-12-31",
  "format": "pdf",
  "include_narratives": true
}
```

**Response:** `202 Accepted`
```json
{
  "report_id": "report-uuid",
  "status": "generating",
  "estimated_completion": "2024-06-15T15:00:00Z"
}
```

---

## Activities

### List Activities

```http
GET /api/activities
```

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| type | string | Activity type filter |
| scope | integer | Scope filter (1, 2, or 3) |
| page | integer | Page number |
| per_page | integer | Items per page |

**Response:**
```json
{
  "data": [
    {
      "id": "activity-uuid",
      "type": "electricity_consumption",
      "scope": 2,
      "description": "Main office electricity",
      "quantity": 10000,
      "unit": "kWh",
      "emissions_tco2e": 4.5,
      "period_start": "2024-01-01",
      "period_end": "2024-01-31",
      "created_at": "2024-02-01T10:30:00Z"
    }
  ],
  "page_info": {
    "page": 1,
    "per_page": 20,
    "total_items": 245
  }
}
```

### Create Activity

```http
POST /api/activities
Content-Type: application/json

{
  "type": "business_travel",
  "scope": 3,
  "description": "Q1 flight emissions",
  "quantity": 5000,
  "unit": "km",
  "transport_mode": "air",
  "period_start": "2024-01-01",
  "period_end": "2024-03-31"
}
```

---

## AI Chat

### Send Chat Message

```http
POST /api/ai/chat
Content-Type: application/json

{
  "prompt": "Summarize my Scope 2 emissions for Q1 2024",
  "context": {
    "include_charts": true,
    "detail_level": "summary"
  }
}
```

**Response:**
```json
{
  "response": "Your Scope 2 emissions for Q1 2024 totaled 75.25 tCO2e...",
  "data_references": [
    {
      "type": "scope2_summary",
      "period": "Q1 2024",
      "value": 75.25
    }
  ],
  "suggested_actions": [
    "Consider renewable energy certificates for market-based reduction",
    "Benchmark against industry peers"
  ]
}
```

---

## OffGrid Mode

### Get Current Mode

```http
GET /api/offgrid/mode
```

**Response:**
```json
{
  "mode": "normal",
  "connectivity_score": 100,
  "last_sync": "2024-06-15T14:30:00Z",
  "pending_sync_items": 0
}
```

### Switch Mode

```http
POST /api/offgrid/mode
Content-Type: application/json

{
  "mode": "offline"
}
```

**Response:**
```json
{
  "mode": "offline",
  "activated_at": "2024-06-15T14:35:00Z",
  "features_available": [
    "data_entry",
    "local_calculations",
    "cached_reports"
  ],
  "features_unavailable": [
    "ai_chat",
    "real_time_sync",
    "external_api_calls"
  ]
}
```

---

## Tenants

### List Tenants

```http
GET /api/tenants
```

**Response:**
```json
{
  "data": [
    {
      "id": "tenant-uuid",
      "name": "Acme Corporation",
      "industry": "TECHNOLOGY",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

### Switch Tenant

```http
POST /api/tenants/{tenant_id}/switch
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "tenant": {
    "id": "tenant-uuid",
    "name": "Acme Corporation"
  }
}
```

---

## Users

### Get Current User

```http
GET /api/users/me
```

**Response:**
```json
{
  "id": "user-uuid",
  "email": "user@example.com",
  "name": "John Doe",
  "role": "admin",
  "default_tenant_id": "tenant-uuid",
  "two_factor_enabled": true,
  "created_at": "2024-01-01T00:00:00Z"
}
```

### Update User

```http
PATCH /api/users/me
Content-Type: application/json

{
  "name": "John Smith",
  "default_tenant_id": "new-tenant-uuid"
}
```

---

## Error Responses

All errors follow a consistent format:

```json
{
  "code": "ERROR_CODE",
  "message": "Human-readable error message",
  "detail": "Additional details if available"
}
```

### Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `UNAUTHORIZED` | 401 | Invalid or expired token |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `NOT_FOUND` | 404 | Resource not found |
| `VALIDATION_ERROR` | 400 | Invalid request data |
| `RATE_LIMIT_EXCEEDED` | 429 | Too many requests |
| `INTERNAL_ERROR` | 500 | Server error |

---

## Rate Limiting

API requests are rate-limited per tenant:

| Plan | Requests/minute | Requests/day |
|------|-----------------|--------------|
| Free | 60 | 1,000 |
| Pro | 300 | 50,000 |
| Enterprise | 1,000 | Unlimited |

Rate limit headers are included in all responses:

```http
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1718456700
```

---

## Webhooks

Configure webhooks to receive real-time notifications.

### Webhook Events

| Event | Description |
|-------|-------------|
| `emission.created` | New emission record created |
| `emission.updated` | Emission record updated |
| `compliance.deadline` | Compliance deadline approaching |
| `anomaly.detected` | Anomaly detected in data |
| `report.generated` | Report generation completed |

### Webhook Payload

```json
{
  "event": "emission.created",
  "timestamp": "2024-06-15T14:30:00Z",
  "tenant_id": "tenant-uuid",
  "data": {
    "id": "emission-uuid",
    "type": "scope2",
    "emissions_tco2e": 4.5
  }
}
```

---

## SDK Examples

### JavaScript/TypeScript

```typescript
import { OffGridFlowClient } from '@offgridflow/sdk';

const client = new OffGridFlowClient({
  apiKey: 'your-api-key',
  tenantId: 'tenant-uuid'
});

const emissions = await client.emissions.getScope2Summary();
console.log(`Total emissions: ${emissions.total_emissions_tco2e} tCO2e`);
```

### Python

```python
from offgridflow import Client

client = Client(
    api_key="your-api-key",
    tenant_id="tenant-uuid"
)

emissions = client.emissions.get_scope2_summary()
print(f"Total emissions: {emissions['total_emissions_tco2e']} tCO2e")
```

### Go

```go
import "github.com/offgridflow/go-sdk"

client := offgridflow.NewClient(
    offgridflow.WithAPIKey("your-api-key"),
    offgridflow.WithTenantID("tenant-uuid"),
)

summary, err := client.Emissions.GetScope2Summary(ctx)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Total emissions: %.2f tCO2e\n", summary.TotalEmissionsTCO2e)
```

---

*API Version: v1*
*Last Updated: 2024*
