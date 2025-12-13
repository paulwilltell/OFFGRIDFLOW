# AI Enhancement Specification

> **OffGridFlow AI Copilot Layer**  
> **Status**: Planned | **Priority**: High | **Target**: Phase 2

---

## Overview

This specification defines the AI copilot capabilities to be integrated into OffGridFlow, enhancing the platform's ability to provide intelligent insights, automate analysis, and generate narrative content for compliance reports.

---

## 1. AI Capabilities Matrix

| Capability | Description | Priority | Complexity |
|------------|-------------|----------|------------|
| **Anomaly Detection** | Identify unusual patterns in emissions data | High | Medium |
| **Methodology Suggestions** | Recommend calculation approaches | High | Low |
| **Narrative Generation** | Auto-generate report narratives | High | Medium |
| **Data Quality Assessment** | Score and improve data quality | Medium | Low |
| **Natural Language Queries** | Query emissions data conversationally | Medium | Medium |
| **Trend Forecasting** | Predict future emissions trends | Low | High |

---

## 2. Anomaly Detection Engine

### 2.1 Detection Types

```go
// internal/ai/anomaly/types.go

type AnomalyType string

const (
    AnomalyTypeSpike       AnomalyType = "spike"         // Sudden increase
    AnomalyTypeDrop        AnomalyType = "drop"          // Sudden decrease
    AnomalyTypePattern     AnomalyType = "pattern"       // Unusual pattern
    AnomalyTypeMissing     AnomalyType = "missing"       // Expected data missing
    AnomalyTypeOutlier     AnomalyType = "outlier"       // Statistical outlier
    AnomalyTypeDrift       AnomalyType = "drift"         // Gradual shift
)

type Anomaly struct {
    ID          uuid.UUID     `json:"id"`
    TenantID    string        `json:"tenant_id"`
    Type        AnomalyType   `json:"type"`
    Severity    string        `json:"severity"`    // low, medium, high, critical
    ActivityID  *uuid.UUID    `json:"activity_id,omitempty"`
    Scope       string        `json:"scope"`
    Description string        `json:"description"`
    Confidence  float64       `json:"confidence"`  // 0.0 - 1.0
    DetectedAt  time.Time     `json:"detected_at"`
    Metadata    map[string]any `json:"metadata"`
}
```

### 2.2 Detection Algorithms

```go
// internal/ai/anomaly/detector.go

type Detector struct {
    logger     *slog.Logger
    store      Store
    thresholds Thresholds
}

type Thresholds struct {
    SpikePercentage    float64 // Default: 50% increase
    DropPercentage     float64 // Default: 50% decrease
    OutlierStdDev      float64 // Default: 3 standard deviations
    MinConfidence      float64 // Default: 0.7
}

// DetectAnomalies analyzes emissions data for anomalies
func (d *Detector) DetectAnomalies(ctx context.Context, tenantID string, opts DetectOptions) ([]Anomaly, error) {
    var anomalies []Anomaly
    
    // 1. Time series analysis for spikes/drops
    timeSeriesAnomalies, err := d.analyzeTimeSeries(ctx, tenantID, opts)
    if err != nil {
        return nil, err
    }
    anomalies = append(anomalies, timeSeriesAnomalies...)
    
    // 2. Statistical outlier detection
    outliers, err := d.detectOutliers(ctx, tenantID, opts)
    if err != nil {
        return nil, err
    }
    anomalies = append(anomalies, outliers...)
    
    // 3. Pattern analysis (seasonality deviations)
    patternAnomalies, err := d.analyzePatterns(ctx, tenantID, opts)
    if err != nil {
        return nil, err
    }
    anomalies = append(anomalies, patternAnomalies...)
    
    // 4. Missing data detection
    missingData, err := d.detectMissingData(ctx, tenantID, opts)
    if err != nil {
        return nil, err
    }
    anomalies = append(anomalies, missingData...)
    
    return d.filterByConfidence(anomalies), nil
}

// Statistical methods
func (d *Detector) analyzeTimeSeries(ctx context.Context, tenantID string, opts DetectOptions) ([]Anomaly, error) {
    // Fetch historical emissions grouped by period
    history, err := d.store.GetEmissionsTimeSeries(ctx, tenantID, opts.StartDate, opts.EndDate)
    if err != nil {
        return nil, err
    }
    
    var anomalies []Anomaly
    for i := 1; i < len(history); i++ {
        prev := history[i-1]
        curr := history[i]
        
        // Calculate percentage change
        if prev.Value > 0 {
            change := (curr.Value - prev.Value) / prev.Value * 100
            
            if change > d.thresholds.SpikePercentage {
                anomalies = append(anomalies, Anomaly{
                    ID:          uuid.New(),
                    TenantID:    tenantID,
                    Type:        AnomalyTypeSpike,
                    Severity:    d.calculateSeverity(change),
                    Description: fmt.Sprintf("%.1f%% increase in emissions from %s to %s", 
                        change, prev.Period, curr.Period),
                    Confidence:  0.85,
                    DetectedAt:  time.Now(),
                })
            }
            
            if change < -d.thresholds.DropPercentage {
                anomalies = append(anomalies, Anomaly{
                    ID:          uuid.New(),
                    TenantID:    tenantID,
                    Type:        AnomalyTypeDrop,
                    Severity:    d.calculateSeverity(-change),
                    Description: fmt.Sprintf("%.1f%% decrease in emissions from %s to %s", 
                        -change, prev.Period, curr.Period),
                    Confidence:  0.85,
                    DetectedAt:  time.Now(),
                })
            }
        }
    }
    
    return anomalies, nil
}

func (d *Detector) detectOutliers(ctx context.Context, tenantID string, opts DetectOptions) ([]Anomaly, error) {
    // Z-score based outlier detection
    activities, err := d.store.GetActivitiesWithEmissions(ctx, tenantID)
    if err != nil {
        return nil, err
    }
    
    // Group by category and calculate statistics
    categoryStats := make(map[string]*stats)
    for _, a := range activities {
        if categoryStats[a.Category] == nil {
            categoryStats[a.Category] = &stats{}
        }
        categoryStats[a.Category].Add(a.Emissions)
    }
    
    var anomalies []Anomaly
    for _, a := range activities {
        s := categoryStats[a.Category]
        zScore := (a.Emissions - s.Mean()) / s.StdDev()
        
        if math.Abs(zScore) > d.thresholds.OutlierStdDev {
            anomalies = append(anomalies, Anomaly{
                ID:          uuid.New(),
                TenantID:    tenantID,
                Type:        AnomalyTypeOutlier,
                ActivityID:  &a.ID,
                Severity:    d.calculateOutlierSeverity(zScore),
                Description: fmt.Sprintf("Activity '%s' emissions (%.2f) is %.1f standard deviations from category mean",
                    a.Name, a.Emissions, zScore),
                Confidence:  0.9,
                DetectedAt:  time.Now(),
            })
        }
    }
    
    return anomalies, nil
}
```

### 2.3 API Endpoints

```go
// GET /api/ai/anomalies
// Returns detected anomalies for the tenant

// POST /api/ai/anomalies/detect
// Triggers anomaly detection with custom parameters
type DetectRequest struct {
    StartDate   time.Time `json:"start_date"`
    EndDate     time.Time `json:"end_date"`
    Scopes      []string  `json:"scopes,omitempty"`      // Filter by scope
    Categories  []string  `json:"categories,omitempty"`  // Filter by category
    MinSeverity string    `json:"min_severity,omitempty"` // low, medium, high
}

// Response
type DetectResponse struct {
    Anomalies []Anomaly `json:"anomalies"`
    Summary   Summary   `json:"summary"`
}

type Summary struct {
    Total    int            `json:"total"`
    BySeverity map[string]int `json:"by_severity"`
    ByType     map[string]int `json:"by_type"`
}
```

---

## 3. Narrative Generation Engine

### 3.1 Architecture

```go
// internal/ai/narrative/generator.go

type Generator struct {
    client    AIClient       // OpenAI or local LLM
    templates TemplateStore
    logger    *slog.Logger
}

type AIClient interface {
    Complete(ctx context.Context, prompt string, opts CompletionOpts) (string, error)
}

// Generate creates narrative content for compliance reports
func (g *Generator) Generate(ctx context.Context, req NarrativeRequest) (*Narrative, error) {
    // 1. Prepare context from emissions data
    dataContext, err := g.prepareContext(ctx, req)
    if err != nil {
        return nil, err
    }
    
    // 2. Select appropriate template
    template, err := g.templates.Get(req.ReportType, req.Section)
    if err != nil {
        return nil, err
    }
    
    // 3. Build prompt
    prompt := g.buildPrompt(template, dataContext)
    
    // 4. Generate narrative
    response, err := g.client.Complete(ctx, prompt, CompletionOpts{
        MaxTokens:   2000,
        Temperature: 0.3, // Lower for factual accuracy
    })
    if err != nil {
        return nil, err
    }
    
    // 5. Post-process and validate
    narrative, err := g.postProcess(response, req)
    if err != nil {
        return nil, err
    }
    
    return narrative, nil
}
```

### 3.2 Report Sections

```go
type NarrativeSection string

const (
    // CSRD/ESRS sections
    SectionExecutiveSummary    NarrativeSection = "executive_summary"
    SectionClimateStrategy     NarrativeSection = "climate_strategy"
    SectionEmissionsOverview   NarrativeSection = "emissions_overview"
    SectionScope1Analysis      NarrativeSection = "scope1_analysis"
    SectionScope2Analysis      NarrativeSection = "scope2_analysis"
    SectionScope3Analysis      NarrativeSection = "scope3_analysis"
    SectionReductionTargets    NarrativeSection = "reduction_targets"
    SectionRisksOpportunities  NarrativeSection = "risks_opportunities"
    SectionMethodology         NarrativeSection = "methodology"
    SectionDataQuality         NarrativeSection = "data_quality"
)

type NarrativeRequest struct {
    TenantID     string           `json:"tenant_id"`
    ReportType   string           `json:"report_type"`   // CSRD, SEC, etc.
    ReportingYear int             `json:"reporting_year"`
    Section      NarrativeSection `json:"section"`
    EmissionsData EmissionsContext `json:"emissions_data"`
    Preferences  Preferences      `json:"preferences"`
}

type Preferences struct {
    Tone       string `json:"tone"`       // formal, technical, accessible
    Length     string `json:"length"`     // brief, standard, detailed
    Language   string `json:"language"`   // en, de, fr, etc.
    IncludeCharts bool `json:"include_charts"`
}
```

### 3.3 Prompt Templates

```yaml
# internal/ai/narrative/templates/csrd_executive_summary.yaml
id: csrd_executive_summary
report_type: CSRD
section: executive_summary
version: 1.0

system_prompt: |
  You are an expert sustainability reporting analyst. Generate professional,
  accurate narrative content for CSRD/ESRS compliance reports. Use precise
  language, cite specific figures, and maintain a formal business tone.
  
  IMPORTANT:
  - All figures must come from the provided data context
  - Do not hallucinate or estimate numbers
  - Use metric units (tonnes CO2e, not tons)
  - Reference the reporting period explicitly

user_template: |
  Generate an executive summary for a CSRD sustainability report.
  
  ## Company Context
  - Organization: {{.CompanyName}}
  - Industry: {{.Industry}}
  - Reporting Period: {{.PeriodStart}} to {{.PeriodEnd}}
  
  ## Emissions Data
  - Scope 1: {{.Scope1Total}} tonnes CO2e
  - Scope 2: {{.Scope2Total}} tonnes CO2e ({{.Scope2Method}})
  - Scope 3: {{.Scope3Total}} tonnes CO2e
  - Total: {{.TotalEmissions}} tonnes CO2e
  
  ## Year-over-Year Change
  - Total change: {{.YoYChange}}%
  - Scope 1 change: {{.Scope1YoY}}%
  - Scope 2 change: {{.Scope2YoY}}%
  - Scope 3 change: {{.Scope3YoY}}%
  
  ## Key Categories
  {{range .TopCategories}}
  - {{.Name}}: {{.Emissions}} tonnes CO2e ({{.Percentage}}%)
  {{end}}
  
  Generate a 3-4 paragraph executive summary that:
  1. Opens with total emissions and context
  2. Highlights key changes and drivers
  3. Notes significant achievements or challenges
  4. Concludes with forward-looking statement

output_format: markdown
max_tokens: 1500
temperature: 0.3
```

### 3.4 API Endpoints

```go
// POST /api/ai/narrative/generate
type GenerateNarrativeRequest struct {
    ReportType    string           `json:"report_type"`
    Section       NarrativeSection `json:"section"`
    ReportingYear int              `json:"reporting_year"`
    Preferences   Preferences      `json:"preferences,omitempty"`
}

type GenerateNarrativeResponse struct {
    Content     string    `json:"content"`
    Section     string    `json:"section"`
    GeneratedAt time.Time `json:"generated_at"`
    Metadata    struct {
        TokensUsed int    `json:"tokens_used"`
        Model      string `json:"model"`
        Version    string `json:"template_version"`
    } `json:"metadata"`
}

// GET /api/ai/narrative/templates
// Returns available narrative templates

// POST /api/ai/narrative/batch
// Generate multiple sections at once
type BatchGenerateRequest struct {
    ReportType    string             `json:"report_type"`
    ReportingYear int                `json:"reporting_year"`
    Sections      []NarrativeSection `json:"sections"`
}
```

---

## 4. Methodology Suggestion Engine

### 4.1 Suggestion Types

```go
// internal/ai/methodology/suggester.go

type SuggestionType string

const (
    SuggestionCalculationMethod SuggestionType = "calculation_method"
    SuggestionEmissionFactor    SuggestionType = "emission_factor"
    SuggestionDataSource        SuggestionType = "data_source"
    SuggestionCategoryMapping   SuggestionType = "category_mapping"
    SuggestionBestPractice      SuggestionType = "best_practice"
)

type Suggestion struct {
    ID          uuid.UUID      `json:"id"`
    Type        SuggestionType `json:"type"`
    Title       string         `json:"title"`
    Description string         `json:"description"`
    Rationale   string         `json:"rationale"`
    Impact      string         `json:"impact"`      // low, medium, high
    Effort      string         `json:"effort"`      // low, medium, high
    ActivityID  *uuid.UUID     `json:"activity_id,omitempty"`
    Actions     []Action       `json:"actions"`
}

type Action struct {
    Label       string `json:"label"`
    Description string `json:"description"`
    AutoApply   bool   `json:"auto_apply"` // Can be applied automatically
}
```

### 4.2 Suggestion Engine

```go
type Suggester struct {
    factorDB  EmissionFactorStore
    rules     RuleEngine
    logger    *slog.Logger
}

func (s *Suggester) GetSuggestions(ctx context.Context, tenantID string) ([]Suggestion, error) {
    var suggestions []Suggestion
    
    // 1. Check for better emission factors
    factorSuggestions, err := s.suggestBetterFactors(ctx, tenantID)
    if err != nil {
        s.logger.Warn("factor suggestions failed", "error", err)
    } else {
        suggestions = append(suggestions, factorSuggestions...)
    }
    
    // 2. Check for methodology improvements
    methodSuggestions, err := s.suggestMethodologyImprovements(ctx, tenantID)
    if err != nil {
        s.logger.Warn("methodology suggestions failed", "error", err)
    } else {
        suggestions = append(suggestions, methodSuggestions...)
    }
    
    // 3. Check for data quality improvements
    dataSuggestions, err := s.suggestDataImprovements(ctx, tenantID)
    if err != nil {
        s.logger.Warn("data suggestions failed", "error", err)
    } else {
        suggestions = append(suggestions, dataSuggestions...)
    }
    
    return s.prioritize(suggestions), nil
}

func (s *Suggester) suggestBetterFactors(ctx context.Context, tenantID string) ([]Suggestion, error) {
    // Find activities using default/global factors where regional factors exist
    activities, err := s.getActivitiesWithDefaultFactors(ctx, tenantID)
    if err != nil {
        return nil, err
    }
    
    var suggestions []Suggestion
    for _, a := range activities {
        // Check if a more specific factor exists
        betterFactor, found := s.factorDB.FindMoreSpecific(ctx, a.CurrentFactor, a.Location)
        if found {
            suggestions = append(suggestions, Suggestion{
                ID:          uuid.New(),
                Type:        SuggestionEmissionFactor,
                Title:       fmt.Sprintf("More accurate emission factor available for %s", a.Name),
                Description: fmt.Sprintf("A region-specific emission factor (%s) is available for %s instead of the global default.",
                    betterFactor.ID, a.Location),
                Rationale:   "Region-specific factors improve accuracy and are preferred by compliance frameworks.",
                Impact:      "medium",
                Effort:      "low",
                ActivityID:  &a.ID,
                Actions: []Action{
                    {Label: "Apply suggested factor", AutoApply: true},
                    {Label: "Review factor details", AutoApply: false},
                },
            })
        }
    }
    
    return suggestions, nil
}
```

---

## 5. Natural Language Query Interface

### 5.1 Query Types

```go
type QueryIntent string

const (
    IntentEmissionsTotal     QueryIntent = "emissions_total"
    IntentEmissionsTrend     QueryIntent = "emissions_trend"
    IntentEmissionsBreakdown QueryIntent = "emissions_breakdown"
    IntentCompare            QueryIntent = "compare"
    IntentTopContributors    QueryIntent = "top_contributors"
    IntentReductionProgress  QueryIntent = "reduction_progress"
)

type ParsedQuery struct {
    Intent     QueryIntent      `json:"intent"`
    Entities   map[string]any   `json:"entities"`
    TimeRange  *TimeRange       `json:"time_range,omitempty"`
    Scope      []string         `json:"scope,omitempty"`
    Categories []string         `json:"categories,omitempty"`
    Confidence float64          `json:"confidence"`
}
```

### 5.2 Example Queries

| Natural Language | Intent | Response |
|------------------|--------|----------|
| "What were our total emissions last year?" | emissions_total | "Your total emissions for 2024 were 12,450 tonnes CO2e..." |
| "Show me Scope 3 breakdown by category" | emissions_breakdown | Chart + "Scope 3 emissions breakdown: Business Travel 35%..." |
| "Compare Q1 vs Q2 emissions" | compare | "Q1: 3,200 tonnes, Q2: 3,450 tonnes (+7.8%)..." |
| "What are our top 5 emission sources?" | top_contributors | "Top 5 sources: 1. Fleet vehicles (2,100 tonnes)..." |
| "How are we tracking against our 2030 target?" | reduction_progress | "You've reduced emissions by 15% toward your 30% target..." |

### 5.3 API Endpoint

```go
// POST /api/ai/query
type QueryRequest struct {
    Query string `json:"query"`
}

type QueryResponse struct {
    Answer      string       `json:"answer"`
    Data        any          `json:"data,omitempty"`        // Structured data for charts
    Confidence  float64      `json:"confidence"`
    Sources     []DataSource `json:"sources"`               // Data sources used
    Suggestions []string     `json:"suggestions,omitempty"` // Follow-up queries
}
```

---

## 6. Integration Architecture

```mermaid
flowchart TB
    subgraph Frontend
        UI[Dashboard UI]
        CHAT[Chat Interface]
    end
    
    subgraph API["API Layer"]
        AIAPI[/api/ai/*]
    end
    
    subgraph AIServices["AI Services"]
        ANOMALY[Anomaly Detector]
        NARRATIVE[Narrative Generator]
        SUGGEST[Methodology Suggester]
        NLQ[NL Query Engine]
    end
    
    subgraph AIProviders["AI Providers"]
        OPENAI[OpenAI API]
        LOCAL[Local LLM\nOllama]
    end
    
    subgraph DataLayer["Data Layer"]
        EMISSIONS[(Emissions DB)]
        FACTORS[(Factors DB)]
        TEMPLATES[(Templates)]
    end
    
    UI --> AIAPI
    CHAT --> AIAPI
    
    AIAPI --> ANOMALY
    AIAPI --> NARRATIVE
    AIAPI --> SUGGEST
    AIAPI --> NLQ
    
    ANOMALY --> EMISSIONS
    NARRATIVE --> EMISSIONS
    NARRATIVE --> OPENAI
    NARRATIVE --> LOCAL
    SUGGEST --> FACTORS
    NLQ --> OPENAI
    NLQ --> LOCAL
    NLQ --> EMISSIONS
```

---

## 7. Configuration

```yaml
# config/ai.yaml
ai:
  # Provider settings
  provider: "openai"  # openai, local, hybrid
  
  openai:
    api_key: "${OPENAI_API_KEY}"
    model: "gpt-4o-mini"
    max_tokens: 4000
    temperature: 0.3
  
  local:
    url: "http://localhost:11434"
    model: "llama3.2:3b"
    
  # Feature toggles
  features:
    anomaly_detection: true
    narrative_generation: true
    methodology_suggestions: true
    nl_queries: true
    
  # Anomaly detection thresholds
  anomaly:
    spike_threshold_percent: 50
    drop_threshold_percent: 50
    outlier_std_dev: 3.0
    min_confidence: 0.7
    
  # Rate limits (per tenant)
  rate_limits:
    narrative_per_hour: 20
    queries_per_hour: 100
    anomaly_scans_per_day: 10
```

---

## 8. Implementation Timeline

| Phase | Task | Duration | Dependencies |
|-------|------|----------|--------------|
| 1 | Anomaly detection core | 2 weeks | Emissions data |
| 1 | Anomaly API + UI | 1 week | Core detection |
| 2 | Narrative templates | 1 week | - |
| 2 | Narrative generator | 2 weeks | Templates, AI provider |
| 2 | Narrative API + UI | 1 week | Generator |
| 3 | Methodology suggester | 2 weeks | Factors DB |
| 3 | NL query parser | 2 weeks | AI provider |
| 4 | Integration testing | 1 week | All components |
| 4 | Documentation | 1 week | All components |

**Total: ~12 weeks (Phase 2 timeline)**

---

## 9. Success Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| Anomaly detection accuracy | >90% | Manual validation sample |
| Narrative quality score | >4.0/5.0 | User feedback |
| Query understanding accuracy | >85% | Intent classification |
| User adoption | >60% | Feature usage analytics |
| Time saved per report | >2 hours | User surveys |

---

**End of AI Enhancement Specification**
