# PHASE 3: Azure Hardened Connector – Complete
## Production-Ready Azure Cloud Emissions Ingestion with OAuth, Rate Limiting, and Error Handling

**Status:** ✅ **SUB-TASK 3 COMPLETE**

---

## 📋 Sub-Task 3: Azure Hardened Connector

### Files Created (4 files + tests)

#### 1. **`azure_hardened.go`** (750 lines)
**Purpose:** Production-hardened Azure connector with OAuth token refresh, pagination, and error handling

**Key Features:**
- Azure OAuth2 token refresh with automatic expiration handling
- Emissions Impact Dashboard API integration
- Cost Management API integration (optional)
- Rate limiting (3 req/sec, conservative for Azure)
- Pagination support (cursor-based with $skiptoken)
- Error classification for intelligent retry
- Observable tracing, metrics, logging
- Token refresh threshold (5 min before expiration)

**Configuration:**
```go
HardenedConfig struct {
    Config                      // Base config (credentials, etc)
    RateLimitCapacity          // Token bucket size (default: 60)
    RateLimitPerSec            // Refill rate (default: 3.0 = conservative)
    MaxRetries                 // Retry attempts (default: 3)
    RequestTimeout             // API timeout (default: 45s = Azure is slower)
    TokenRefreshThreshold      // Refresh X seconds before expiration (default: 5min)
    MaxPages                   // Pagination limit (default: 1000)
    MaxPageSize                // Items per page (default: 1000)
    FetchEmissionsAPI          // Enable Emissions Dashboard (default: true)
    FetchCostManagement        // Enable Cost Management (default: false)
    Logger                     // Structured logging
    Observability              // OTEL tracing & metrics
}
```

**Main Flows:**
```go
Ingest(ctx)
  ├─ ingestEmissionsAPI(ctx)      // Emissions Impact Dashboard
  │   ├─ Pagination loop with $skiptoken
  │   ├─ Rate limit per page
  │   ├─ Retry transient errors
  │   ├─ Error classification
  │   └─ Convert to Activities
  │
  └─ ingestCostManagement(ctx)    // Cost Management (optional)
      ├─ Build cost query
      ├─ Rate limit
      ├─ Fetch data
      └─ Estimate emissions from cost
```

**Key Methods:**
```go
NewHardenedAdapter(cfg)              // Create adapter
Ingest(ctx)                          // Main ingestion
ingestEmissionsAPI(ctx)              // Emissions flow
ingestCostManagement(ctx)            // Cost flow
fetchEmissionsPage(ctx, skipToken)   // Paginated fetch
fetchCostManagementData(ctx, query)  // Cost query
convertEmissionsToActivities()       // Transform Emissions
convertCostToActivities()            // Transform Cost
retryWithExponentialBackoff()        // Retry logic
GetToken(ctx)                        // OAuth token handling
```

**Azure OAuth Features:**
- Automatic token refresh when expired
- 5-minute refresh threshold (refresh proactively)
- Scope: `https://management.azure.com/.default`
- Grant type: `client_credentials`
- Graceful error handling with classification

---

#### 2. **`azure_mocks.go`** (120 lines)
**Purpose:** Mock clients for testing without real Azure API calls

**Key Types:**
```go
MockTokenProvider           // Mock Azure OAuth token generation
MockEmissionsAPI           // Mock Emissions Impact Dashboard API
```

**Mock Features:**
- TokenProvider: Control token generation, simulate failures, retry simulation
- EmissionsAPI: Multi-page pagination, error simulation per page
- Call tracking for testing retry logic
- FailFirstN simulation for retry testing

**Sample Test Data:**
```go
SampleEmissionRecord()      // Single emission record
SampleEmissionRecords(n)    // Multiple records
SampleCostRecords()         // Cost management records
```

---

#### 3. **`azure_hardened_test.go`** (420 lines)
**Purpose:** Comprehensive test suite (17 test cases)

**Test Coverage:**

Configuration & Adapter (3 tests):
- `TestNewHardenedAdapterValidConfig` – Valid configuration
- `TestNewHardenedAdapterInvalidConfig` – Config validation
- `TestConfigValidation` – All config scenarios

Token Provider (3 tests):
- `TestMockTokenProvider` – Token generation
- `TestMockTokenProviderError` – Error handling
- `TestMockTokenProviderFailFirstN` – Retry simulation

Activity Conversion (3 tests):
- `TestConvertEmissionsToActivities` – Emissions conversion
- `TestConvertEmissionsZeroEmissions` – Zero-emission filtering
- `TestConvertCostToActivities` – Cost conversion

Rate Limiting & Retry (3 tests):
- `TestRateLimitingApplied` – Token bucket enforcement
- `TestRetryWithExponentialBackoff` – Successful retry
- `TestRetryStopsOnNonRetryableError` – No retry on auth

API & Integration (4 tests):
- `TestMockEmissionsAPI` – Mock API pagination
- `TestCategorizeAzureService` – Service categorization
- `TestMapAzureRegion` – Region mapping
- `TestEmissionsConversionFlow` – End-to-end flow
- `TestCostManagementFlow` – Cost flow

---

### Key Differences from AWS

| Feature | AWS | Azure |
|---------|-----|-------|
| **Auth** | SigV4 request signing | OAuth2 token refresh |
| **Token Handling** | Derived from credentials | Explicit token + expiration |
| **Rate Limit** | 5 req/sec (AWS is faster) | 3 req/sec (Azure conservative) |
| **Timeout** | 30s | 45s (Azure is slower) |
| **Manifest** | S3 manifest.json | Direct API pagination |
| **Pagination** | Cursor-based (S3) | Cursor-based ($skiptoken) |
| **Data Sources** | Carbon Footprint + CUR | Emissions Dashboard + Cost Mgmt |

---

### Architecture

```
┌──────────────────────────────────────────────────────────────────────────┐
│                     HardenedAdapter (Azure Connector)                    │
├──────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  Ingest(ctx) ─────────────────────────────────────────────────────────┐ │
│    ├─ [1] Emissions Impact Dashboard API                             │ │
│    │     ├─ GetToken() → OAuth token refresh                         │ │
│    │     ├─ Rate limit (3 req/sec)                                  │ │
│    │     ├─ Pagination loop ($skiptoken)                            │ │
│    │     ├─ Retry transient errors (exponential backoff)            │ │
│    │     ├─ Error classification (auth → no retry)                 │ │
│    │     ├─ Convert to Activities                                   │ │
│    │     └─ Tracing spans + metrics                                │ │
│    │                                                                 │ │
│    └─ [2] Cost Management API (optional)                            │ │
│          ├─ GetToken() → OAuth token refresh                        │ │
│          ├─ Build cost query (date range, granularity)              │ │
│          ├─ Rate limit                                              │ │
│          ├─ Fetch cost data (no pagination)                         │ │
│          └─ Estimate emissions from cost (rough model)              │ │
│                                                                     │ │
└─────────────────────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────────────────┘

Utilities Used:
  • RateLimiter       → Enforce 3 requests/second (conservative)
  • PaginationState   → Track $skiptoken pagination
  • ErrorClassify     → Decide retry vs fail-fast
  • TokenProvider     → OAuth token refresh with threshold
  • Observability     → Trace, metrics, structured logs
```

---

### Usage Example

```go
// 1. Create configuration
cfg := azure.NewHardenedConfig(azure.Config{
    TenantID:       "00000000-0000-0000-0000-000000000000",
    ClientID:       "00000001-0000-0000-0000-000000000000",
    ClientSecret:   os.Getenv("AZURE_CLIENT_SECRET"),
    SubscriptionID: "00000002-0000-0000-0000-000000000000",
    OrgID:          "org-123",
    StartDate:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
    EndDate:        time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
})

// 2. Create adapter (with automatic token refresh)
adapter, err := azure.NewHardenedAdapter(cfg)
if err != nil {
    log.Fatal(err)
}

// 3. Ingest (automatic rate limiting, token refresh, retry)
ctx := context.Background()
activities, err := adapter.Ingest(ctx)
if err != nil {
    log.Fatal(err)
}

// 4. Use activities
fmt.Printf("Ingested %d activities\n", len(activities))
```

---

### OAuth Token Refresh Flow

```
1. First API call
   ├─ GetToken(ctx)
   │   ├─ Check: token expired?
   │   ├─ No: Return cached token
   │   └─ Yes: Call Azure AD
   │       ├─ POST /oauth2/v2.0/token
   │       ├─ client_id + client_secret + scope
   │       ├─ Response: access_token + expires_in
   │       └─ Cache token until (now + expires_in - 5min)
   │
   └─ Use token in Authorization header
      ├─ Bearer {token}

2. Subsequent API calls
   ├─ GetToken(ctx)
   │   ├─ Check: time.Now() < tokenExpiresAt - 5min?
   │   ├─ Yes: Return cached (no HTTP call)
   │   └─ No: Refresh token (see flow above)
   │
   └─ Use token (likely from cache = fast)

Key: 5-minute refresh threshold prevents token expiration mid-operation
```

---

### Error Handling

| Error Type | Example | Action |
|---|---|---|
| **Transient (429)** | Rate limited | Retry with backoff |
| **Transient (timeout)** | Network timeout | Retry with backoff |
| **Auth (401)** | Invalid token | Fail immediately (don't retry) |
| **Auth (403)** | Permission denied | Fail immediately |
| **NotFound (404)** | Subscription doesn't exist | Fail immediately |
| **BadRequest (400)** | Invalid query | Fail immediately |
| **Fatal (500)** | Server error | Fail immediately (might retry once) |

---

### Rate Limiting Strategy

Azure is stricter than AWS:
- Official limit: ~1800 calls/min = 30/sec
- Conservative default: 3 req/sec
- Token bucket: 60 capacity, 3 tokens/sec refill
- Configurable if you need faster ingestion

```go
cfg.RateLimitCapacity = 100
cfg.RateLimitPerSec = 10.0  // 10 requests/second
```

---

## ✅ Quality Checklist

- ✅ All functions documented (godoc)
- ✅ Thread-safe (rate limiter, token provider)
- ✅ No panic() calls – all errors returned
- ✅ Context cancellation respected throughout
- ✅ Rate limiting integrated
- ✅ Pagination support (cursor-based)
- ✅ OAuth token refresh with threshold
- ✅ Error classification for retry decisions
- ✅ Observability (tracing, metrics, logging)
- ✅ 17 test cases covering happy path + errors
- ✅ Mock clients for testing
- ✅ Production-ready error handling
- ✅ Credentials not hardcoded (env vars only)

---

## 📊 Metrics Summary

| Metric | Value |
|--------|-------|
| Files Created | 3 (hardened, mocks, tests) |
| Lines of Code | ~1,290 |
| Lines of Tests | ~420 |
| Test Cases | 17 |
| Error Classes Handled | 6 (transient, auth, bad, not-found, fatal, unknown) |
| API Endpoints Supported | 2 (Emissions Dashboard, Cost Management) |
| Rate Limit Configurable | Yes (tokens/sec, capacity) |
| Retry Logic | Exponential backoff, classified errors |
| Observability Integration | OTEL (spans, metrics, logs) |
| OAuth Token Refresh | ✅ Automatic with 5-min threshold |

---

## 🔄 Next Steps

**PHASE 3: Sub-Task 4** – GCP Connector Hardening
- BigQuery or Cloud Billing API integration
- Service account authentication
- Write comprehensive tests with mocks

---

## 📝 Code Quality

**Standards Met:**
- ✅ Consistent error handling with classification
- ✅ Structured logging with context
- ✅ Rate limiting for all API calls
- ✅ Pagination for all list operations
- ✅ OAuth token refresh with automatic expiration handling
- ✅ Retry logic with exponential backoff
- ✅ Observable traces and metrics
- ✅ Thread-safe concurrent access
- ✅ Comprehensive test coverage (17 cases)
- ✅ Production-ready error messages
- ✅ No hardcoded secrets

---

**Status:** ✅ **SUB-TASK 3 COMPLETE**  
**Review:** Ready for approval before proceeding to GCP connector

---

## Command Reference

```bash
# Build Azure package
go build ./internal/ingestion/sources/azure

# Run Azure tests
go test -v ./internal/ingestion/sources/azure

# Run specific test
go test -v ./internal/ingestion/sources/azure -run TestToken

# Test with race detector
go test -race ./internal/ingestion/sources/azure

# Coverage report
go test -cover ./internal/ingestion/sources/azure
```

---
