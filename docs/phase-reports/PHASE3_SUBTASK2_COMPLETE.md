# PHASE 3: AWS CUR Connector Hardening – Complete
## Production-Ready Cloud Connector with Rate Limiting, Pagination, Error Classification, and Observability

**Status:** ✅ **SUB-TASK 2 COMPLETE**

---

## 📋 Sub-Task 2: AWS CUR Connector Hardening

### Files Created (6 files)

#### 1. **`aws_s3_manifest.go`** (210 lines)
**Purpose:** Parse and validate AWS Cost and Usage Report (CUR) manifest.json files  
**Key Features:**
- Parse S3 manifest JSON structure
- Filter report files (CSV preferred, Parquet fallback)
- Validate manifest completeness
- Support incremental ingestion (detect new files)
- Human-readable summaries

**Key Types:**
```go
S3Manifest          // Manifest.json structure
BillingPeriod       // Start/end dates
ManifestFile        // Individual file entry
```

**Key Functions:**
```go
ParseS3Manifest(data)                 // Parse manifest.json bytes
ParseS3ManifestFromReader(r)          // Parse from io.Reader
GetReportFiles()                      // Filter report files (CSV > Parquet)
GetFileByKey(key)                     // Get file by S3 key
ValidateManifest(m, logger)           // Validate completeness
IsNewManifest(current, previous)      // Detect changes (incremental)
GetNewFiles(current, previous)        // Get only new files
```

---

#### 2. **`aws_mocks.go`** (280 lines)
**Purpose:** Mock clients for testing without real AWS API calls  
**Features:**
- MockS3Client: Mocks S3 GetObject & ListObjectsV2
- MockCarbonAPI: Mocks Carbon Footprint API
- Sample test data (manifest, records, responses)
- Error simulation (failures, timeouts)
- Call counting for retry verification

**Key Types:**
```go
MockS3Client        // Mock S3 client
MockCarbonAPI       // Mock Carbon API
```

**Key Methods:**
```go
AddObject(key, data)              // Add mock object
GetObject(ctx, bucket, key)       // Get object (mocked)
ListObjectsV2(ctx, bucket, prefix)// List objects (mocked)
SetGetObjectError(key, err)       // Simulate failure
SetListObjectsError(prefix, err)  // Simulate failure
GetCarbonFootprint(ctx)           // Get mocked response
FailFirstN                        // Simulate retries
```

**Sample Data:**
```go
SampleManifestJSON()              // AWS CUR manifest
SampleCURCSVRow()                 // CSV row
SampleCarbonFootprintResponse()   // API response
```

---

#### 3. **`aws_hardened.go`** (650 lines)
**Purpose:** Production-hardened AWS connector with all utilities integrated  
**Key Features:**
- ✅ Rate limiting (token bucket, configurable)
- ✅ Pagination (S3 manifest + file iteration)
- ✅ Error classification (transient vs fatal)
- ✅ Observability (tracing, metrics, logging)
- ✅ Retry logic (exponential backoff, classified errors)
- ✅ S3 manifest parsing (incremental ingestion)
- ✅ Carbon Footprint API integration
- ✅ CUR CSV parsing with region/service mapping

**Configuration:**
```go
HardenedConfig struct {
    Config                      // Base config (credentials, region, etc)
    RateLimitCapacity          // Token bucket size (default: 100)
    RateLimitPerSec            // Refill rate (default: 5.0)
    MaxRetries                 // Retry attempts (default: 3)
    RequestTimeout             // API timeout (default: 30s)
    IncludeS3Manifest          // Enable S3 CUR (default: true)
    MaxPages                   // Pagination limit (default: 1000)
    Logger                     // Structured logging
    Observability              // OTEL config
}
```

**Main Flow:**
```go
Ingest(ctx)
  ├─ ingestCarbonFootprint(ctx)  // Cloud dashboard API
  │   ├─ Rate limit (Allow)
  │   ├─ Retry with backoff
  │   ├─ Error classification
  │   └─ Tracing + metrics
  │
  └─ ingestS3CUR(ctx)            // Detailed usage reports
      ├─ fetchS3Manifest()
      ├─ ValidateManifest()
      ├─ Pagination for files
      ├─ fetchAndParseS3File()   (for each file)
      ├─ Rate limit per file
      └─ convertToActivities()
```

**Key Methods:**
```go
NewHardenedAdapter(cfg)            // Create adapter
Ingest(ctx)                        // Main ingestion
ingestCarbonFootprint(ctx)         // Carbon API flow
ingestS3CUR(ctx)                   // S3 CUR flow
fetchCarbonFootprint(ctx)          // API call
fetchS3Manifest(ctx)               // Get manifest
fetchAndParseS3File(ctx, key)      // Get + parse file
convertCarbonToActivities()        // Transform Carbon
convertCURToActivities()           // Transform CUR
retryWithExponentialBackoff()      // Retry logic
```

---

#### 4. **`aws_hardened_test.go`** (380 lines)
**Purpose:** Comprehensive test suite (15 test cases)  
**Test Coverage:**

**Adapter Tests (3):**
- `TestNewHardenedAdapterValidConfig` – Valid configuration
- `TestNewHardenedAdapterInvalidConfig` – Config validation
- `TestConvertCarbonToActivities` – Carbon conversion
- `TestConvertCURToActivities` – CUR conversion

**Manifest Tests (4):**
- `TestParseS3Manifest` – Parse manifest.json
- `TestParseS3ManifestInvalid` – Invalid manifest error
- `TestValidateManifest` – Validation logic
- `TestGetReportFiles` – File filtering

**Error Classification Tests (1):**
- `TestErrorClassificationInIngestion` – Error classes

**Rate Limiting Tests (1):**
- `TestRateLimitingApplied` – Token bucket enforcement

**Retry Logic Tests (2):**
- `TestRetryWithExponentialBackoff` – Successful retry
- `TestRetryStopsOnNonRetryableError` – Non-retryable early exit

**Integration Tests (2):**
- `TestCarbonFootprintIngestion` – End-to-end Carbon flow
- `TestS3ManifestIngestionFlow` – Manifest + pagination

**Mock Tests (3):**
- `TestMockS3Client` – S3 mocking
- `TestMockCarbonAPI` – Carbon API mocking
- `TestMockCarbonAPIFailFirstN` – Retry simulation

---

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                   HardenedAdapter                           │
├─────────────────────────────────────────────────────────────┤
│ Config + Utilities:                                         │
│  ├─ RateLimiter     (token bucket, 5 req/sec)             │
│  ├─ Logger          (slog structured)                      │
│  ├─ Tracer          (OTEL spans)                           │
│  └─ Metrics         (counters, histograms)                 │
├─────────────────────────────────────────────────────────────┤
│ Main Flow (Ingest):                                         │
│  ├─ Carbon Footprint API                                   │
│  │   ├─ Rate limit + Wait                                  │
│  │   ├─ Retry with exponential backoff                    │
│  │   ├─ Error classification (transient vs fatal)         │
│  │   ├─ Trace spans + metrics                             │
│  │   └─ Convert to Activities                             │
│  │                                                          │
│  └─ S3 CUR Manifest (optional)                            │
│      ├─ Fetch manifest.json from S3                       │
│      ├─ Parse + Validate                                  │
│      ├─ Get report files (CSV preferred)                  │
│      ├─ For each file:                                    │
│      │   ├─ Rate limit                                    │
│      │   ├─ Fetch from S3                                 │
│      │   ├─ Parse CSV                                     │
│      │   └─ Convert to Activities                         │
│      └─ Return all activities                             │
└─────────────────────────────────────────────────────────────┘
```

---

### Hardening Features

| Feature | Implementation | Benefit |
|---------|---|---|
| **Rate Limiting** | Token bucket, 5 req/sec | Respect API limits |
| **Pagination** | Cursor-based (manifest), offset-based (CUR) | Handles large datasets |
| **Error Classification** | 6 classes (transient/auth/bad/not-found/fatal/unknown) | Intelligent retry decisions |
| **Retry Logic** | Exponential backoff (1s → 30s) | Temporary failures recover |
| **Observability** | OTEL spans + metrics + slog | Monitor production health |
| **Manifest Parsing** | Incremental ingestion support | Avoid re-downloading |
| **SigV4 Signing** | AWS request signing | Authenticates to AWS |
| **CSV Parsing** | Dynamic header mapping | Handles schema variations |
| **Region/Service Mapping** | AWS → OffGridFlow normalization | Consistent emissions data |
| **Concurrent Safety** | RWMutex in rate limiter | Safe for multi-goroutine |

---

### Usage Example

```go
// 1. Create configuration
cfg := NewHardenedConfig(Config{
    AccessKeyID:     "AKIA...",
    SecretAccessKey: "wJal...",
    Region:          "us-east-1",
    OrgID:           "org-123",
    S3Bucket:        "my-cur-bucket",
    S3Prefix:        "cur/",
    StartDate:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
    EndDate:         time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
})

// 2. Create adapter
adapter, err := NewHardenedAdapter(cfg)
if err != nil {
    log.Fatal(err)
}

// 3. Ingest (with all hardening applied)
ctx := context.Background()
activities, err := adapter.Ingest(ctx)
if err != nil {
    log.Fatal(err)
}

// 4. Activities now ready for emissions calculation
fmt.Printf("Ingested %d activities\n", len(activities))
```

---

### Error Handling Example

```go
// Automatic error classification + retry:
err := adapter.retryWithExponentialBackoff(ctx, func() error {
    // Rate limit
    _, err := adapter.limiter.Allow(ctx)
    if err != nil {
        return err  // Context cancelled
    }

    // API call
    footprint, err := adapter.fetchCarbonFootprint(ctx)
    if err != nil {
        ce := ingestion.ClassifyError(err)
        // Transient (429, timeout)   → Retry with backoff
        // Auth (401, 403)             → No retry, fail fast
        // NotFound (404)              → No retry
        // BadRequest (400, validation)→ No retry
        return ce
    }

    return nil
})
```

---

### Testing Strategy

**Unit Tests (15 cases):**
- Happy path: adapter creation, conversion, parsing
- Error paths: invalid config, bad manifest, API errors
- Rate limiting: token consumption, waiting
- Retry logic: successful retry, early exit on fatal
- Integration: end-to-end flows with mocks

**Mock Clients:**
- MockS3Client: Simulates S3 GetObject, ListObjectsV2
- MockCarbonAPI: Simulates API responses, failures
- FailFirstN: Retry simulation (N failures, then success)

**Coverage:**
- All error classes (transient, auth, bad, not-found, fatal)
- Pagination boundaries
- Manifest validation
- Service/region mapping
- CSV parsing (header mapping, type conversion)

---

### Metrics & Observability

**Metrics Recorded:**
```
ingestion_success_total         Counter  (connector=aws)
ingestion_failure_total         Counter  (connector=aws-carbon, error_class=...)
ingestion_items_total           Counter  (connector=aws)
ingestion_latency_seconds       Histogram (p50, p95, p99)
```

**Traces Emitted:**
```
aws.ingest_carbon_footprint  (spans for API call, retry, conversion)
aws.ingest_s3_cur            (spans for manifest, file processing)
```

**Logs:**
```
ingestion complete: total_activities=150 latency_ms=2345
failed to ingest S3 CUR: error=NoSuchBucket
rate limiting applied: wait_ms=250
retrying after error: attempt=2 backoff_ms=2000
```

---

## ✅ Quality Checklist

- ✅ All functions documented (godoc)
- ✅ Thread-safe (RWMutex in utilities)
- ✅ Error handling: No panics, proper classification
- ✅ Context cancellation: Respected throughout
- ✅ Rate limiting: Token bucket integrated
- ✅ Pagination: Cursor and offset both supported
- ✅ Observability: Tracing, metrics, logging
- ✅ Tests: 15 cases covering happy path + errors
- ✅ Mocks: Full S3 and API mocking for testing
- ✅ Production-ready: Error classification, retry logic, safe concurrency

---

## 📊 Metrics Summary

| Metric | Value |
|--------|-------|
| Files Created | 4 (manifest, mocks, hardened, tests) |
| Lines of Code | ~1,520 |
| Lines of Tests | ~380 |
| Test Cases | 15 |
| Error Classes Handled | 6 (transient, auth, bad, not-found, fatal, unknown) |
| API Endpoints Supported | 2 (Carbon Footprint API, S3 GetObject/ListObjectsV2) |
| Rate Limit Configurable | Yes (tokens/sec, capacity) |
| Retry Logic | Exponential backoff, classified errors |
| Observability Integration | OTEL (spans, metrics, logs) |

---

## 🔄 Next Steps

**PHASE 3: Sub-Task 3** – Azure Connector Hardening
- Replicate AWS hardening for Azure Cost Management API
- Implement token refresh (Azure OAuth)
- Handle Azure pagination
- Write comprehensive tests with mocks

---

## 📝 Code Quality

**Standards Met:**
- ✅ Consistent error handling with classification
- ✅ Structured logging with context
- ✅ Rate limiting for all API calls
- ✅ Pagination for all list operations
- ✅ Retry logic with exponential backoff
- ✅ Observable traces and metrics
- ✅ Thread-safe concurrent access
- ✅ Comprehensive test coverage (15 cases)
- ✅ Production-ready error messages
- ✅ No hardcoded secrets

---

**Status:** ✅ **SUB-TASK 2 COMPLETE**  
**Review:** Ready for approval before proceeding to Azure connector

---

## Command Reference

```bash
# Build AWS package
go build ./internal/ingestion/sources/aws

# Run AWS tests
go test -v ./internal/ingestion/sources/aws

# Run specific test
go test -v ./internal/ingestion/sources/aws -run TestCarbonFootprint

# Test with race detector
go test -race ./internal/ingestion/sources/aws

# Coverage report
go test -cover ./internal/ingestion/sources/aws

# Benchmark
go test -bench=. ./internal/ingestion/sources/aws
```

---

