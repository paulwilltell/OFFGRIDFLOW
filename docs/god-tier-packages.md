# OffGridFlow API Reference - God-Tier Packages

This document provides comprehensive API documentation for OffGridFlow's advanced feature packages.

## Table of Contents

1. [Anomaly Detection](#anomaly-detection)
2. [Benchmarking Service](#benchmarking-service)
3. [Blockchain Audit Trail](#blockchain-audit-trail)
4. [Carbon Credit Marketplace](#carbon-credit-marketplace)
5. [Demo Mode](#demo-mode)
6. [Edge Sync](#edge-sync)
7. [Email Notifications](#email-notifications)
8. [Emission Factors Auto-Update](#emission-factors-auto-update)
9. [AI Narratives](#ai-narratives)
10. [Rate Limiting](#rate-limiting)
11. [Regulatory Monitoring](#regulatory-monitoring)
12. [Data Residency](#data-residency)
13. [Scenario Modeling](#scenario-modeling)
14. [SOC2 Compliance](#soc2-compliance)
15. [Supplier Engagement](#supplier-engagement)
16. [Supply Chain Graph](#supply-chain-graph)
17. [White-Label API](#white-label-api)

---

## Anomaly Detection

**Package:** `internal/anomaly`

Provides AI-powered anomaly detection for emissions data to identify reporting errors, data quality issues, and unusual patterns.

### Types

```go
type AnomalyType string

const (
    AnomalySpike     AnomalyType = "spike"     // Sudden increase
    AnomalyDrop      AnomalyType = "drop"      // Sudden decrease
    AnomalyPattern   AnomalyType = "pattern"   // Pattern deviation
    AnomalyMissing   AnomalyType = "missing"   // Missing data
)

type Severity string

const (
    SeverityLow      Severity = "low"
    SeverityMedium   Severity = "medium"
    SeverityHigh     Severity = "high"
    SeverityCritical Severity = "critical"
)

type Anomaly struct {
    ID          string      // Unique identifier
    TenantID    string      // Tenant identifier
    Type        AnomalyType // Type of anomaly
    Severity    Severity    // Severity level
    Description string      // Human-readable description
    DataPoint   DataPoint   // The anomalous data point
    Expected    float64     // Expected value
    Actual      float64     // Actual value
    Confidence  float64     // Detection confidence (0-1)
    DetectedAt  time.Time   // Detection timestamp
}
```

### Detector Interface

```go
type Detector interface {
    // Detect analyzes data points and returns detected anomalies
    Detect(ctx context.Context, tenantID string, data []DataPoint) ([]Anomaly, error)
    
    // DetectRealTime processes a single data point for real-time detection
    DetectRealTime(ctx context.Context, tenantID string, point DataPoint) (*Anomaly, error)
}
```

### Usage Example

```go
detector := anomaly.NewDetector(anomaly.DefaultConfig())
anomalies, err := detector.Detect(ctx, tenantID, emissionsData)
if err != nil {
    return fmt.Errorf("detection failed: %w", err)
}

for _, a := range anomalies {
    if a.Severity == anomaly.SeverityHigh {
        notifyStakeholders(a)
    }
}
```

---

## Benchmarking Service

**Package:** `internal/benchmarking`

Provides industry benchmarking capabilities to compare emissions performance against peers.

### Types

```go
type IndustryCode string

const (
    IndustryManufacturing IndustryCode = "MANUFACTURING"
    IndustryTechnology    IndustryCode = "TECHNOLOGY"
    IndustryRetail        IndustryCode = "RETAIL"
    IndustryTransport     IndustryCode = "TRANSPORT"
)

type BenchmarkResult struct {
    TenantID       string       // Tenant identifier
    Industry       IndustryCode // Industry classification
    Percentile     int          // Percentile ranking (1-100)
    PeerAverage    float64      // Industry average emissions
    TopQuartile    float64      // Top 25% threshold
    YourValue      float64      // Your emissions value
    Recommendation string       // AI-generated recommendation
    CalculatedAt   time.Time    // Calculation timestamp
}
```

### Service Interface

```go
type Service interface {
    // Compare compares tenant emissions against industry peers
    Compare(ctx context.Context, tenantID string, industry IndustryCode) (*BenchmarkResult, error)
    
    // GetIndustryStats retrieves aggregate industry statistics
    GetIndustryStats(ctx context.Context, industry IndustryCode) (*IndustryStats, error)
}
```

### Usage Example

```go
svc := benchmarking.NewService(db, config)
result, err := svc.Compare(ctx, tenantID, benchmarking.IndustryTechnology)
if err != nil {
    return err
}

fmt.Printf("You are in the %d percentile for your industry\n", result.Percentile)
```

---

## Blockchain Audit Trail

**Package:** `internal/blockchain`

Provides immutable blockchain-based audit trail for emissions data using cryptographic verification.

### Types

```go
type Block struct {
    Index        int64     // Block number
    Timestamp    time.Time // Block creation time
    DataHash     string    // Hash of data in this block
    PreviousHash string    // Hash of previous block
    Hash         string    // This block's hash
    Nonce        int64     // Proof of work nonce
}

type AuditEntry struct {
    ID         string    // Entry identifier
    TenantID   string    // Tenant identifier
    DataType   string    // Type of audited data
    DataID     string    // Reference to original data
    DataHash   string    // Hash of the data
    BlockHash  string    // Block containing this entry
    CreatedAt  time.Time // Entry creation time
}
```

### Chain Interface

```go
type Chain interface {
    // AddEntry adds a new audit entry to the blockchain
    AddEntry(ctx context.Context, entry AuditEntry) (*Block, error)
    
    // Verify checks if a data hash exists in the chain
    Verify(ctx context.Context, dataHash string) (bool, *AuditEntry, error)
    
    // GetChain retrieves the full blockchain
    GetChain(ctx context.Context) ([]Block, error)
}
```

### Usage Example

```go
chain := blockchain.NewChain(store, blockchain.DefaultConfig())

// Add audit entry for emissions report
entry := blockchain.AuditEntry{
    TenantID: tenantID,
    DataType: "emissions_report",
    DataID:   reportID,
    DataHash: calculateHash(reportData),
}

block, err := chain.AddEntry(ctx, entry)
if err != nil {
    return fmt.Errorf("failed to add audit entry: %w", err)
}

// Later: verify data integrity
verified, _, err := chain.Verify(ctx, dataHash)
```

---

## Carbon Credit Marketplace

**Package:** `internal/carboncredit`

Provides carbon credit trading and offset management capabilities.

### Types

```go
type CreditType string

const (
    CreditVCS       CreditType = "VCS"       // Verified Carbon Standard
    CreditGoldStd   CreditType = "GOLD_STD"  // Gold Standard
    CreditACR       CreditType = "ACR"       // American Carbon Registry
    CreditCAR       CreditType = "CAR"       // Climate Action Reserve
)

type CreditStatus string

const (
    StatusAvailable CreditStatus = "available"
    StatusReserved  CreditStatus = "reserved"
    StatusRetired   CreditStatus = "retired"
)

type CarbonCredit struct {
    ID          string       // Credit identifier
    ProjectID   string       // Source project
    Type        CreditType   // Credit standard
    VintageYear int          // Year credits generated
    Quantity    float64      // Tons CO2e
    PricePerTon float64      // Current price
    Status      CreditStatus // Availability status
}

type Order struct {
    ID        string    // Order identifier
    TenantID  string    // Buyer tenant
    CreditID  string    // Credit being purchased
    Quantity  float64   // Quantity ordered
    TotalCost float64   // Total cost
    Status    string    // Order status
    CreatedAt time.Time // Order timestamp
}
```

### Marketplace Interface

```go
type Marketplace interface {
    // ListCredits retrieves available carbon credits
    ListCredits(ctx context.Context, filters CreditFilters) ([]CarbonCredit, error)
    
    // Purchase creates an order for carbon credits
    Purchase(ctx context.Context, tenantID string, creditID string, quantity float64) (*Order, error)
    
    // Retire marks credits as used for offset
    Retire(ctx context.Context, tenantID string, creditID string, reason string) error
}
```

---

## Demo Mode

**Package:** `internal/demo`

Provides demonstration mode with synthetic data for trials and testing.

### Types

```go
type DemoConfig struct {
    TenantID         string        // Demo tenant identifier
    DataRealism      float64       // How realistic data should be (0-1)
    IndustryProfile  string        // Industry to simulate
    EmissionsScale   float64       // Scale factor for emissions
    Duration         time.Duration // Demo session duration
}
```

### Handler Interface

```go
type Handler interface {
    // StartDemo initializes a demo session
    StartDemo(ctx context.Context, config DemoConfig) (*Session, error)
    
    // GenerateData creates synthetic demo data
    GenerateData(ctx context.Context, sessionID string) error
    
    // EndDemo terminates a demo session and cleans up
    EndDemo(ctx context.Context, sessionID string) error
}
```

---

## Edge Sync

**Package:** `internal/edge`

Provides offline-first synchronization for edge deployments and field operations.

### Types

```go
type SyncState string

const (
    SyncPending     SyncState = "pending"
    SyncInProgress  SyncState = "in_progress"
    SyncComplete    SyncState = "complete"
    SyncConflict    SyncState = "conflict"
    SyncFailed      SyncState = "failed"
)

type SyncItem struct {
    ID         string    // Item identifier
    TenantID   string    // Tenant identifier
    EntityType string    // Type of entity
    EntityID   string    // Entity identifier
    Operation  string    // Operation type (create/update/delete)
    Data       []byte    // Serialized data
    State      SyncState // Sync state
    Attempts   int       // Sync attempts
    CreatedAt  time.Time // Queue time
    SyncedAt   *time.Time // Successful sync time
}
```

### Sync Interface

```go
type Sync interface {
    // Queue adds an item to the sync queue
    Queue(ctx context.Context, item SyncItem) error
    
    // Process syncs queued items to the cloud
    Process(ctx context.Context) (*SyncResult, error)
    
    // ResolveConflict resolves a sync conflict
    ResolveConflict(ctx context.Context, itemID string, resolution Resolution) error
}
```

---

## Email Notifications

**Package:** `internal/email`

Provides email notification capabilities for alerts, reports, and compliance deadlines.

### Types

```go
type EmailType string

const (
    EmailAlert      EmailType = "alert"
    EmailReport     EmailType = "report"
    EmailDigest     EmailType = "digest"
    EmailCompliance EmailType = "compliance"
)

type Email struct {
    To          []string          // Recipient addresses
    Subject     string            // Email subject
    Type        EmailType         // Email type
    TemplateID  string            // Template identifier
    Data        map[string]any    // Template data
    Attachments []Attachment      // File attachments
}
```

### Client Interface

```go
type Client interface {
    // Send sends an email
    Send(ctx context.Context, email Email) error
    
    // SendBatch sends multiple emails
    SendBatch(ctx context.Context, emails []Email) (*BatchResult, error)
    
    // ScheduleDigest schedules a recurring digest email
    ScheduleDigest(ctx context.Context, tenantID string, config DigestConfig) error
}
```

---

## Emission Factors Auto-Update

**Package:** `internal/emissionfactors`

Provides automatic emission factor updates from authoritative sources.

### Types

```go
type FactorSource string

const (
    SourceEPA     FactorSource = "EPA"     // US EPA
    SourceDefra   FactorSource = "DEFRA"   // UK DEFRA
    SourceGHG     FactorSource = "GHG"     // GHG Protocol
    SourceCustom  FactorSource = "CUSTOM"  // Custom factors
)

type EmissionFactor struct {
    ID           string       // Factor identifier
    Source       FactorSource // Data source
    Category     string       // Emissions category
    Region       string       // Geographic region
    Value        float64      // Factor value
    Unit         string       // Unit of measure
    ValidFrom    time.Time    // Validity start
    ValidTo      *time.Time   // Validity end
    LastUpdated  time.Time    // Last update time
}
```

### Updater Interface

```go
type Updater interface {
    // CheckForUpdates checks for new emission factor data
    CheckForUpdates(ctx context.Context) ([]FactorUpdate, error)
    
    // ApplyUpdates applies pending factor updates
    ApplyUpdates(ctx context.Context, updates []FactorUpdate) error
    
    // GetFactors retrieves current emission factors
    GetFactors(ctx context.Context, filters FactorFilters) ([]EmissionFactor, error)
}
```

---

## AI Narratives

**Package:** `internal/narratives`

Provides AI-generated narrative explanations for emissions data and reports.

### Types

```go
type NarrativeType string

const (
    NarrativeExecutiveSummary NarrativeType = "executive_summary"
    NarrativeTrendAnalysis    NarrativeType = "trend_analysis"
    NarrativeCompliance       NarrativeType = "compliance"
    NarrativeRecommendations  NarrativeType = "recommendations"
)

type Narrative struct {
    ID        string        // Narrative identifier
    TenantID  string        // Tenant identifier
    Type      NarrativeType // Narrative type
    Content   string        // Generated narrative text
    DataRange DateRange     // Data time range
    Confidence float64      // AI confidence score
    CreatedAt time.Time     // Generation timestamp
}
```

### Generator Interface

```go
type Generator interface {
    // Generate creates an AI narrative
    Generate(ctx context.Context, tenantID string, nType NarrativeType, data any) (*Narrative, error)
    
    // Regenerate recreates a narrative with feedback
    Regenerate(ctx context.Context, narrativeID string, feedback string) (*Narrative, error)
}
```

---

## Rate Limiting

**Package:** `internal/ratelimit`

Provides API rate limiting to protect service availability.

### Types

```go
type LimitType string

const (
    LimitPerSecond  LimitType = "per_second"
    LimitPerMinute  LimitType = "per_minute"
    LimitPerHour    LimitType = "per_hour"
    LimitPerDay     LimitType = "per_day"
)

type Limit struct {
    Key       string    // Rate limit key (tenant, API key, IP)
    Type      LimitType // Limit type
    Max       int       // Maximum requests
    Current   int       // Current count
    ResetsAt  time.Time // Reset timestamp
}
```

### Limiter Interface

```go
type Limiter interface {
    // Allow checks if a request is allowed
    Allow(ctx context.Context, key string) (bool, *Limit, error)
    
    // Reset resets the limit for a key
    Reset(ctx context.Context, key string) error
    
    // GetLimit retrieves current limit status
    GetLimit(ctx context.Context, key string) (*Limit, error)
}
```

---

## Regulatory Monitoring

**Package:** `internal/regulatory`

Provides monitoring of regulatory changes affecting emissions reporting.

### Types

```go
type Framework string

const (
    FrameworkCSRD      Framework = "CSRD"
    FrameworkSEC       Framework = "SEC"
    FrameworkCBAM      Framework = "CBAM"
    FrameworkIFRS      Framework = "IFRS_S2"
    FrameworkCalifornia Framework = "CALIFORNIA"
)

type RegulatoryUpdate struct {
    ID          string     // Update identifier
    Framework   Framework  // Affected framework
    Title       string     // Update title
    Summary     string     // Brief summary
    FullText    string     // Full regulation text
    EffectiveDate time.Time // When it takes effect
    Impact      string     // Impact assessment
    ActionRequired bool    // Requires user action
    PublishedAt time.Time  // Publication date
}
```

### Monitor Interface

```go
type Monitor interface {
    // CheckUpdates checks for regulatory changes
    CheckUpdates(ctx context.Context) ([]RegulatoryUpdate, error)
    
    // GetFrameworkStatus retrieves compliance status for a framework
    GetFrameworkStatus(ctx context.Context, tenantID string, framework Framework) (*FrameworkStatus, error)
    
    // Subscribe subscribes to regulatory updates
    Subscribe(ctx context.Context, tenantID string, frameworks []Framework) error
}
```

---

## Data Residency

**Package:** `internal/residency`

Provides multi-region data residency routing for compliance with data sovereignty requirements.

### Types

```go
type Region string

const (
    RegionUS     Region = "US"
    RegionEU     Region = "EU"
    RegionUK     Region = "UK"
    RegionAPAC   Region = "APAC"
)

type ResidencyPolicy struct {
    TenantID       string   // Tenant identifier
    PrimaryRegion  Region   // Primary data region
    AllowedRegions []Region // Allowed data regions
    CreatedAt      time.Time // Policy creation time
}
```

### Router Interface

```go
type Router interface {
    // Route determines the appropriate data region
    Route(ctx context.Context, tenantID string) (Region, error)
    
    // SetPolicy sets data residency policy
    SetPolicy(ctx context.Context, policy ResidencyPolicy) error
    
    // GetPolicy retrieves current policy
    GetPolicy(ctx context.Context, tenantID string) (*ResidencyPolicy, error)
}
```

---

## Scenario Modeling

**Package:** `internal/scenarios`

Provides what-if scenario modeling for emissions planning.

### Types

```go
type ScenarioType string

const (
    ScenarioReduction  ScenarioType = "reduction"
    ScenarioGrowth     ScenarioType = "growth"
    ScenarioTransition ScenarioType = "transition"
)

type Scenario struct {
    ID          string       // Scenario identifier
    TenantID    string       // Tenant identifier
    Name        string       // Scenario name
    Type        ScenarioType // Scenario type
    Assumptions []Assumption // Model assumptions
    Projections []Projection // Emissions projections
    CreatedAt   time.Time    // Creation timestamp
}

type Projection struct {
    Year           int     // Projection year
    Scope1         float64 // Projected Scope 1
    Scope2         float64 // Projected Scope 2
    Scope3         float64 // Projected Scope 3
    Total          float64 // Total projected
    VsBaseline     float64 // % change vs baseline
    Confidence     float64 // Projection confidence
}
```

### Engine Interface

```go
type Engine interface {
    // Create creates a new scenario
    Create(ctx context.Context, scenario Scenario) (*Scenario, error)
    
    // Calculate runs scenario calculations
    Calculate(ctx context.Context, scenarioID string) ([]Projection, error)
    
    // Compare compares multiple scenarios
    Compare(ctx context.Context, scenarioIDs []string) (*Comparison, error)
}
```

---

## SOC2 Compliance

**Package:** `internal/soc2`

Provides SOC2 compliance monitoring and evidence collection.

### Types

```go
type ControlCategory string

const (
    ControlSecurity     ControlCategory = "security"
    ControlAvailability ControlCategory = "availability"
    ControlProcessing   ControlCategory = "processing"
    ControlConfidentiality ControlCategory = "confidentiality"
    ControlPrivacy      ControlCategory = "privacy"
)

type Control struct {
    ID          string          // Control identifier
    Category    ControlCategory // Control category
    Name        string          // Control name
    Description string          // Control description
    Status      string          // Compliance status
    Evidence    []Evidence      // Supporting evidence
    LastReview  time.Time       // Last review date
}
```

### Compliance Interface

```go
type Compliance interface {
    // GetControls retrieves all SOC2 controls
    GetControls(ctx context.Context) ([]Control, error)
    
    // UpdateControl updates control status
    UpdateControl(ctx context.Context, controlID string, status string, evidence Evidence) error
    
    // GenerateReport generates SOC2 compliance report
    GenerateReport(ctx context.Context, startDate, endDate time.Time) (*Report, error)
}
```

---

## Supplier Engagement

**Package:** `internal/supplier`

Provides supplier engagement and Scope 3 data collection capabilities.

### Types

```go
type EngagementStatus string

const (
    EngagementInvited    EngagementStatus = "invited"
    EngagementActive     EngagementStatus = "active"
    EngagementResponded  EngagementStatus = "responded"
    EngagementDeclined   EngagementStatus = "declined"
)

type Supplier struct {
    ID               string           // Supplier identifier
    TenantID         string           // Parent tenant
    Name             string           // Supplier name
    ContactEmail     string           // Primary contact
    EngagementStatus EngagementStatus // Current status
    DataQuality      float64          // Data quality score
    LastContact      time.Time        // Last contact date
}

type DataRequest struct {
    ID          string    // Request identifier
    SupplierID  string    // Target supplier
    DataType    string    // Type of data requested
    Deadline    time.Time // Response deadline
    Reminders   int       // Reminders sent
}
```

### Engagement Interface

```go
type Engagement interface {
    // InviteSupplier sends engagement invitation
    InviteSupplier(ctx context.Context, supplier Supplier) error
    
    // RequestData sends a data request to supplier
    RequestData(ctx context.Context, request DataRequest) error
    
    // GetSupplierData retrieves submitted supplier data
    GetSupplierData(ctx context.Context, supplierID string) (*SupplierData, error)
}
```

---

## Supply Chain Graph

**Package:** `internal/supplychain`

Provides supply chain graph analysis for Scope 3 tracking.

### Types

```go
type NodeType string

const (
    NodeSupplier    NodeType = "supplier"
    NodeFacility    NodeType = "facility"
    NodeDistributor NodeType = "distributor"
    NodeCustomer    NodeType = "customer"
)

type Node struct {
    ID        string   // Node identifier
    Type      NodeType // Node type
    Name      string   // Node name
    Tier      int      // Supply chain tier
    Emissions float64  // Associated emissions
}

type Edge struct {
    From       string  // Source node
    To         string  // Target node
    Flow       float64 // Material/product flow
    Emissions  float64 // Transport emissions
}
```

### Graph Interface

```go
type Graph interface {
    // AddNode adds a supply chain node
    AddNode(ctx context.Context, node Node) error
    
    // AddEdge adds an edge between nodes
    AddEdge(ctx context.Context, edge Edge) error
    
    // CalculateScope3 calculates Scope 3 emissions from graph
    CalculateScope3(ctx context.Context, tenantID string) (*Scope3Result, error)
    
    // FindHotspots identifies emissions hotspots
    FindHotspots(ctx context.Context, tenantID string) ([]Hotspot, error)
}
```

---

## White-Label API

**Package:** `internal/whitelabel`

Provides multi-tenant white-label API capabilities.

### Types

```go
type BrandingConfig struct {
    TenantID   string            // Tenant identifier
    LogoURL    string            // Logo URL
    PrimaryColor string          // Primary brand color
    CompanyName string           // Display name
    Domain     string            // Custom domain
    Features   []string          // Enabled features
    CustomCSS  string            // Custom styling
}

type APIKey struct {
    ID        string    // Key identifier
    TenantID  string    // Owner tenant
    Key       string    // API key value
    Name      string    // Key name
    Scopes    []string  // Allowed scopes
    ExpiresAt *time.Time // Expiration time
    CreatedAt time.Time // Creation time
}
```

### API Interface

```go
type API interface {
    // SetBranding configures tenant branding
    SetBranding(ctx context.Context, config BrandingConfig) error
    
    // GetBranding retrieves tenant branding
    GetBranding(ctx context.Context, tenantID string) (*BrandingConfig, error)
    
    // CreateAPIKey creates a new API key
    CreateAPIKey(ctx context.Context, key APIKey) (*APIKey, error)
    
    // RevokeAPIKey revokes an API key
    RevokeAPIKey(ctx context.Context, keyID string) error
}
```

---

## Error Handling

All packages follow a consistent error handling pattern:

```go
var (
    ErrNotFound      = errors.New("resource not found")
    ErrUnauthorized  = errors.New("unauthorized")
    ErrInvalidInput  = errors.New("invalid input")
    ErrRateLimit     = errors.New("rate limit exceeded")
    ErrInternal      = errors.New("internal error")
)
```

## Context Usage

All methods accept a `context.Context` for:
- Cancellation propagation
- Timeout handling
- Request-scoped values (tenant ID, trace ID)

```go
ctx := context.WithTimeout(context.Background(), 30*time.Second)
result, err := service.Method(ctx, params)
```

## Best Practices

1. **Always check errors** - All methods return errors that should be handled
2. **Use contexts** - Pass appropriate contexts for timeout/cancellation
3. **Respect rate limits** - Check rate limit headers in responses
4. **Handle pagination** - Use cursor-based pagination for large datasets
5. **Cache when appropriate** - Cache emission factors and benchmarking data

---

*Last updated: 2024*
