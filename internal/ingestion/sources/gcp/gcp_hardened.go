// Package gcp provides a hardened ingestion adapter for GCP cloud emissions data.
// This version includes rate limiting, pagination, error classification, and observability.
package gcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/example/offgridflow/internal/ingestion"
	"github.com/google/uuid"
	"google.golang.org/api/option"
)

// =============================================================================
// Hardened GCP Adapter Configuration
// =============================================================================

// HardenedConfig extends Config with additional hardening options.
type HardenedConfig struct {
	// Base configuration
	Config

	// RateLimitCapacity is the token bucket capacity (default: 200)
	RateLimitCapacity float64

	// RateLimitPerSec is the refill rate in requests/second (default: 20.0)
	// GCP is generous: allows high concurrency, we're more conservative but higher than Azure
	RateLimitPerSec float64

	// MaxRetries is the maximum number of retry attempts (default: 3)
	MaxRetries int

	// RequestTimeout is the timeout for individual API calls (default: 60s)
	// BigQuery can be slower due to query execution
	RequestTimeout time.Duration

	// Logger for structured logging (optional)
	Logger *slog.Logger

	// Observability configuration (optional)
	Observability *ingestion.ObservabilityConfig

	// MaxPages limits pagination to prevent runaway requests (default: 1000)
	MaxPages int

	// MaxPageSize is the max rows per BigQuery page (default: 1000)
	MaxPageSize int64

	// FetchBigQueryData enables fetching from BigQuery (default: true)
	FetchBigQueryData bool

	// FetchBillingAPI enables fetching from Cloud Billing API (default: false)
	FetchBillingAPI bool

	// BigQueryProjectID overrides the project ID for BigQuery (optional)
	BigQueryProjectID string
}

// NewHardenedConfig creates a new hardened configuration with defaults.
func NewHardenedConfig(cfg Config) *HardenedConfig {
	return &HardenedConfig{
		Config:            cfg,
		RateLimitCapacity: 200,
		RateLimitPerSec:   20.0, // GCP allows high rate
		MaxRetries:        3,
		RequestTimeout:    60 * time.Second, // BigQuery can be slower
		MaxPages:          1000,
		MaxPageSize:       1000,
		FetchBigQueryData: true,
		FetchBillingAPI:   false,
		Logger:            slog.Default(),
		Observability:     ingestion.NewObservabilityConfig("gcp"),
	}
}

// Validate checks hardened configuration.
func (hc *HardenedConfig) Validate() error {
	if err := hc.Config.Validate(); err != nil {
		return err
	}

	if hc.RateLimitCapacity <= 0 {
		return fmt.Errorf("gcp: rate_limit_capacity must be positive")
	}

	if hc.RateLimitPerSec <= 0 {
		return fmt.Errorf("gcp: rate_limit_per_sec must be positive")
	}

	if hc.MaxRetries < 1 {
		return fmt.Errorf("gcp: max_retries must be >= 1")
	}

	if hc.RequestTimeout < 1*time.Second {
		return fmt.Errorf("gcp: request_timeout must be >= 1s")
	}

	if hc.MaxPageSize <= 0 || hc.MaxPageSize > 100000 {
		return fmt.Errorf("gcp: max_page_size must be between 1 and 100000")
	}

	return nil
}

// =============================================================================
// Service Account Authenticator
// =============================================================================

// ServiceAccountAuth handles GCP service account authentication.
type ServiceAccountAuth struct {
	keyJSON  string
	jsonData map[string]interface{}
	logger   *slog.Logger
}

// NewServiceAccountAuth creates a new service account authenticator.
func NewServiceAccountAuth(keyJSON string, logger *slog.Logger) (*ServiceAccountAuth, error) {
	if logger == nil {
		logger = slog.Default()
	}

	// Validate JSON
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(keyJSON), &data); err != nil {
		return nil, ingestion.NewClassifiedError(
			ingestion.ErrorClassBadRequest,
			"invalid service account JSON",
			err,
		)
	}

	// Verify required fields
	requiredFields := []string{"type", "project_id", "private_key", "client_email"}
	for _, field := range requiredFields {
		if _, ok := data[field]; !ok {
			return nil, ingestion.NewClassifiedError(
				ingestion.ErrorClassBadRequest,
				fmt.Sprintf("service account JSON missing %s", field),
				nil,
			)
		}
	}

	return &ServiceAccountAuth{
		keyJSON:  keyJSON,
		jsonData: data,
		logger:   logger,
	}, nil
}

// GetProjectID extracts the GCP project ID from the service account key.
func (saa *ServiceAccountAuth) GetProjectID() string {
	if projectID, ok := saa.jsonData["project_id"].(string); ok {
		return projectID
	}
	return ""
}

// GetOption returns the option.ClientOption for GCP client initialization.
func (saa *ServiceAccountAuth) GetOption() option.ClientOption {
	return option.WithCredentialsJSON([]byte(saa.keyJSON))
}

// =============================================================================
// Hardened Adapter
// =============================================================================

// HardenedAdapter ingests GCP emissions data with production hardening.
type HardenedAdapter struct {
	config   *HardenedConfig
	bqClient *bigquery.Client
	auth     *ServiceAccountAuth
	limiter  *ingestion.RateLimiter
	tracer   *ingestion.InvocationTracer
	metrics  *ingestion.IngestionMetrics
	logger   *slog.Logger
}

// NewHardenedAdapter creates a new hardened GCP adapter.
func NewHardenedAdapter(ctx context.Context, cfg *HardenedConfig) (*HardenedAdapter, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// Setup service account authentication
	var auth *ServiceAccountAuth
	var err error

	if cfg.ServiceAccountKey != "" {
		auth, err = NewServiceAccountAuth(cfg.ServiceAccountKey, cfg.Logger)
		if err != nil {
			return nil, err
		}
	} else {
		// Use application default credentials
		auth = &ServiceAccountAuth{logger: cfg.Logger}
	}

	// Get project ID (for BigQuery)
	projectID := cfg.ProjectID
	if auth.GetProjectID() != "" {
		projectID = auth.GetProjectID()
	}
	if cfg.BigQueryProjectID != "" {
		projectID = cfg.BigQueryProjectID
	}

	// Initialize BigQuery client
	var bqClient *bigquery.Client
	if cfg.FetchBigQueryData {
		var opt option.ClientOption
		if auth.keyJSON != "" {
			opt = auth.GetOption()
		}

		bqClient, err = bigquery.NewClient(ctx, projectID, opt)
		if err != nil {
			return nil, ingestion.NewClassifiedError(
				ingestion.ErrorClassAuth,
				"failed to create BigQuery client",
				err,
			)
		}
	}

	// Create rate limiter
	limiter := ingestion.NewRateLimiter(
		cfg.RateLimitCapacity,
		cfg.RateLimitPerSec,
		50*time.Millisecond,
	)

	// Setup observability
	tracer := ingestion.NewInvocationTracer(cfg.Observability)
	metrics, err := ingestion.NewIngestionMetrics(cfg.Observability.Meter)
	if err != nil {
		cfg.Logger.Warn("failed to setup ingestion metrics", "error", err)
	}

	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}

	return &HardenedAdapter{
		config:   cfg,
		bqClient: bqClient,
		auth:     auth,
		limiter:  limiter,
		tracer:   tracer,
		metrics:  metrics,
		logger:   logger,
	}, nil
}

// Close closes the BigQuery client.
func (ha *HardenedAdapter) Close() error {
	if ha.bqClient != nil {
		return ha.bqClient.Close()
	}
	return nil
}

// =============================================================================
// Ingest - Main Entry Point
// =============================================================================

// Ingest fetches GCP emissions data with full hardening.
func (ha *HardenedAdapter) Ingest(ctx context.Context) ([]ingestion.Activity, error) {
	start := time.Now()
	var activities []ingestion.Activity

	// Ingest from BigQuery
	if ha.config.FetchBigQueryData && ha.bqClient != nil {
		bqActivities, err := ha.ingestBigQuery(ctx)
		if err != nil {
			ha.recordMetricsFailure("bigquery", err)
			return nil, err
		}
		activities = append(activities, bqActivities...)
		ha.logger.Info("ingested from bigquery", "count", len(bqActivities))
	}

	// Optionally ingest from Cloud Billing API
	if ha.config.FetchBillingAPI {
		billingActivities, err := ha.ingestBillingAPI(ctx)
		if err != nil {
			// Log but don't fail if billing is optional
			ha.logger.Warn("failed to ingest billing api", "error", err)
		} else {
			activities = append(activities, billingActivities...)
			ha.logger.Info("ingested from billing api", "count", len(billingActivities))
		}
	}

	// Record success metrics
	latency := time.Since(start)
	ha.recordMetricsSuccess(len(activities), latency)
	ha.logger.Info("gcp ingestion complete", "total_activities", len(activities), "latency_ms", latency.Milliseconds())

	return activities, nil
}

// =============================================================================
// BigQuery Ingestion
// =============================================================================

// ingestBigQuery fetches carbon data from BigQuery with hardening.
func (ha *HardenedAdapter) ingestBigQuery(ctx context.Context) ([]ingestion.Activity, error) {
	var activities []ingestion.Activity

	err := ha.tracer.TraceInvocation(ctx, "gcp.ingest_bigquery", func(ctx context.Context) error {
		// Build query
		query := ha.buildCarbonFootprintQuery()

		// Pagination state
		pagination := ingestion.NewPaginationState(int(ha.config.MaxPageSize))
		pagination.MaxPages = ha.config.MaxPages

		// Execute query with retry
		var records []CarbonRecord

		err := ha.retryWithExponentialBackoff(ctx, func() error {
			var err error
			records, err = ha.executeBigQueryQuery(ctx, query)
			return err
		})

		if err != nil {
			ha.tracer.LogIngestionError(ctx, err, "gcp-bigquery")
			return err
		}

		if len(records) == 0 {
			ha.logger.Warn("bigquery returned no records")
			return nil
		}

		// Convert to activities
		pageActivities := ha.convertCarbonRecordsToActivities(records)
		activities = append(activities, pageActivities...)
		pagination.TotalFetched = len(records)

		ha.logger.Info("bigquery query complete", "records", len(records), "activities", len(pageActivities))
		return nil
	})

	return activities, err
}

// buildCarbonFootprintQuery builds a BigQuery SQL query for carbon data.
func (ha *HardenedConfig) buildCarbonFootprintQuery() string {
	dataset := ha.BigQueryDataset
	table := ha.BigQueryTable

	if dataset == "" {
		dataset = "carbon_footprint"
	}
	if table == "" {
		table = "carbon_footprint_export"
	}

	// Build date filter
	startMonth := ha.StartDate.Format("200601")
	endMonth := ha.EndDate.Format("200601")

	return fmt.Sprintf(`
SELECT
  billing_account_id,
  project.id as project_id,
  project.name as project_name,
  service.id as service_id,
  service.description as service_description,
  location.location as location,
  location.country as country,
  location.region as region,
  usage_month,
  carbon_footprint_kg_co2,
  carbon_model_version,
  CAST(COALESCE(scope_1_emissions_kg_co2, 0) as FLOAT64) as scope_1_emissions_kg_co2,
  CAST(COALESCE(scope_2_emissions_kg_co2, 0) as FLOAT64) as scope_2_emissions_kg_co2,
  CAST(COALESCE(scope_3_emissions_kg_co2, 0) as FLOAT64) as scope_3_emissions_kg_co2,
  CAST(COALESCE(electricity_consumption_kwh, 0) as FLOAT64) as electricity_consumption_kwh,
  CAST(COALESCE(carbon_free_energy_score, 0) as FLOAT64) as carbon_free_energy_score
FROM %s.%s
WHERE usage_month >= '%s' AND usage_month < '%s'
ORDER BY usage_month DESC, project.id
`, dataset, table, startMonth, endMonth)
}

// executeBigQueryQuery executes a BigQuery query with rate limiting.
func (ha *HardenedAdapter) executeBigQueryQuery(ctx context.Context, query string) ([]CarbonRecord, error) {
	// Rate limit before querying
	if _, err := ha.limiter.Allow(ctx); err != nil {
		return nil, ingestion.NewClassifiedError(ingestion.ErrorClassTransient, "rate limiter cancelled", err)
	}

	// Execute query
	q := ha.bqClient.Query(query)
	q.Location = "US" // Carbon footprint data is typically in US multi-region

	// Create job config with timeout
	queryCtx, cancel := context.WithTimeout(ctx, ha.config.RequestTimeout)
	defer cancel()

	it, err := q.Read(queryCtx)
	if err != nil {
		ce := ingestion.ClassifyError(err)
		return nil, ce
	}

	var records []CarbonRecord

	// Iterate through rows
	for {
		var values []bigquery.Value
		err := it.Next(&values)
		if err != nil {
			break // End of iterator or error
		}

		var rec CarbonRecord

		if len(values) < 15 {
			continue // Skip malformed rows
		}

		// Parse each column
		rec.BillingAccountID, _ = values[0].(string)
		rec.Project.ID, _ = values[1].(string)
		rec.Project.Name, _ = values[2].(string)
		rec.Service.ID, _ = values[3].(string)
		rec.Service.Description, _ = values[4].(string)
		rec.Location.Location, _ = values[5].(string)
		rec.Location.Country, _ = values[6].(string)
		rec.Location.Region, _ = values[7].(string)
		rec.UsageMonth, _ = values[8].(string)

		if f, ok := values[9].(float64); ok {
			rec.CarbonFootprintKgCO2 = f
		}

		rec.CarbonModelVersion, _ = values[10].(string)

		if f, ok := values[11].(float64); ok {
			rec.Scope1Emissions = f
		}
		if f, ok := values[12].(float64); ok {
			rec.Scope2Emissions = f
		}
		if f, ok := values[13].(float64); ok {
			rec.Scope3Emissions = f
		}
		if f, ok := values[14].(float64); ok {
			rec.ElectricityKWh = f
		}

		records = append(records, rec)
	}

	return records, nil
}

// =============================================================================
// Cloud Billing API (Optional)
// =============================================================================

// ingestBillingAPI fetches from Cloud Billing API (optional).
func (ha *HardenedAdapter) ingestBillingAPI(ctx context.Context) ([]ingestion.Activity, error) {
	var activities []ingestion.Activity

	err := ha.tracer.TraceInvocation(ctx, "gcp.ingest_billing_api", func(ctx context.Context) error {
		// Rate limit
		if _, err := ha.limiter.Allow(ctx); err != nil {
			return ingestion.NewClassifiedError(ingestion.ErrorClassTransient, "rate limiter cancelled", err)
		}

		ha.logger.Debug("billing api fetch not yet implemented")
		return nil
	})

	return activities, err
}

// =============================================================================
// Activity Conversion
// =============================================================================

// convertCarbonRecordsToActivities transforms GCP carbon records to activities.
func (ha *HardenedAdapter) convertCarbonRecordsToActivities(records []CarbonRecord) []ingestion.Activity {
	activities := make([]ingestion.Activity, 0, len(records))
	now := time.Now().UTC()

	for _, record := range records {
		// Use total CO2e or sum of scopes
		emissions := record.CarbonFootprintKgCO2
		if emissions <= 0 {
			emissions = record.Scope1Emissions + record.Scope2Emissions + record.Scope3Emissions
		}

		if emissions <= 0 {
			continue // Skip zero-emission records
		}

		// Parse usage month (YYYYMM format)
		periodStart, _ := time.Parse("200601", record.UsageMonth)
		periodEnd := periodStart.AddDate(0, 1, 0)

		activity := ingestion.Activity{
			ID:          uuid.NewString(),
			Source:      "gcp_carbon_footprint",
			Category:    categorizeGCPServiceHardened(record.Service.ID, record.Service.Description),
			Location:    mapGCPRegionHardened(record.Location.Location),
			PeriodStart: periodStart,
			PeriodEnd:   periodEnd,
			Quantity:    emissions / 1000, // Convert kg to metric tonnes
			Unit:        "tonne",
			OrgID:       ha.config.OrgID,
			Metadata: map[string]string{
				"gcp_project_id":       record.Project.ID,
				"gcp_project_name":     record.Project.Name,
				"gcp_service_id":       record.Service.ID,
				"gcp_service_desc":     record.Service.Description,
				"gcp_location":         record.Location.Location,
				"gcp_country":          record.Location.Country,
				"gcp_region":           record.Location.Region,
				"carbon_model_version": record.CarbonModelVersion,
				"scope1_kg":            fmt.Sprintf("%.2f", record.Scope1Emissions),
				"scope2_kg":            fmt.Sprintf("%.2f", record.Scope2Emissions),
				"scope3_kg":            fmt.Sprintf("%.2f", record.Scope3Emissions),
				"electricity_kwh":      fmt.Sprintf("%.2f", record.ElectricityKWh),
				"carbon_free_energy":   fmt.Sprintf("%.1f%%", record.CFEScore),
				"data_source":          "gcp_carbon_footprint",
			},
			CreatedAt:   now,
			DataQuality: "measured",
			ExternalID:  fmt.Sprintf("gcp_%s_%s_%s", record.Project.ID, record.Service.ID, record.UsageMonth),
		}

		activities = append(activities, activity)
	}

	return activities
}

// =============================================================================
// Retry Logic
// =============================================================================

// retryWithExponentialBackoff retries with classified error handling.
func (ha *HardenedAdapter) retryWithExponentialBackoff(ctx context.Context, fn func() error) error {
	backoff := 1 * time.Second

	for attempt := 1; attempt <= ha.config.MaxRetries; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		ce := ingestion.ClassifyError(err)
		if ce == nil {
			return err
		}

		if !ce.IsRetryable() {
			ha.logger.Info("non-retryable error, stopping retries", "class", ce.Class)
			return err
		}

		if attempt >= ha.config.MaxRetries {
			ha.logger.Warn("max retries exceeded", "attempts", attempt)
			return err
		}

		ha.logger.Info("retrying after error", "attempt", attempt, "backoff_ms", backoff.Milliseconds())
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
			backoff *= 2
			if backoff > 30*time.Second {
				backoff = 30 * time.Second
			}
		}
	}

	return fmt.Errorf("gcp: max retries exceeded")
}

// =============================================================================
// Metrics
// =============================================================================

func (ha *HardenedAdapter) recordMetricsSuccess(itemCount int, latency time.Duration) {
	if ha.metrics == nil {
		return
	}
	ha.metrics.RecordSuccess(context.Background(), itemCount, latency, "gcp")
}

func (ha *HardenedAdapter) recordMetricsFailure(source string, err error) {
	if ha.metrics == nil {
		return
	}
	ce := ingestion.ClassifyError(err)
	errorClass := "unknown"
	if ce != nil {
		errorClass = string(ce.Class)
	}
	ha.metrics.RecordFailure(context.Background(), "gcp-"+source, errorClass)
}

// =============================================================================
// Helper Functions
// =============================================================================

func categorizeGCPServiceHardened(serviceID, description string) string {
	categoryMap := map[string]string{
		"6F81-5844-456A": "cloud_compute",   // Compute Engine
		"95FF-2EF5-5EA1": "cloud_compute",   // GKE
		"A21B-3453-0BA2": "cloud_database",  // Cloud SQL
		"6002-A0B2-FB9F": "cloud_database",  // Firestore
		"24E6-581D-38E5": "cloud_storage",   // Cloud Storage
		"FEAB-9939-3218": "cloud_storage",   // Cloud Datastore
		"4ED8-1D4C-39A9": "cloud_network",   // Cloud CDN
		"8FC7-9B88-F69D": "data_processing", // Dataflow
		"A4C6-3B37-4BA7": "data_warehouse",  // BigQuery
	}

	if category, ok := categoryMap[serviceID]; ok {
		return category
	}

	// Try description-based matching
	lower := strings.ToLower(description)
	if strings.Contains(lower, "compute") {
		return "cloud_compute"
	}
	if strings.Contains(lower, "database") {
		return "cloud_database"
	}
	if strings.Contains(lower, "storage") {
		return "cloud_storage"
	}

	return "cloud_other"
}

func mapGCPRegionHardened(gcpLocation string) string {
	regionMap := map[string]string{
		"us-central1":             "US-CENTRAL",
		"us-east1":                "US-EAST",
		"us-east4":                "US-EAST",
		"us-west1":                "US-WEST",
		"us-west2":                "US-WEST",
		"us-west3":                "US-WEST",
		"us-west4":                "US-WEST",
		"us-south1":               "US-SOUTH",
		"europe-west1":            "EU-WEST",
		"europe-west2":            "EU-WEST",
		"europe-west3":            "EU-WEST",
		"europe-west6":            "EU-WEST",
		"europe-north1":           "EU-NORTH",
		"europe-southwest1":       "EU-WEST",
		"northamerica-northeast1": "US-EAST",
		"southamerica-east1":      "LATAM",
		"asia-east1":              "ASIA-PACIFIC",
		"asia-east2":              "ASIA-PACIFIC",
		"asia-northeast1":         "ASIA-PACIFIC",
		"asia-northeast2":         "ASIA-PACIFIC",
		"asia-northeast3":         "ASIA-PACIFIC",
		"asia-south1":             "ASIA-PACIFIC",
		"asia-south2":             "ASIA-PACIFIC",
		"asia-southeast1":         "ASIA-PACIFIC",
		"asia-southeast2":         "ASIA-PACIFIC",
		"australia-southeast1":    "ASIA-PACIFIC",
		"australia-southeast2":    "ASIA-PACIFIC",
		"me-west1":                "MENA",
		"africa-south1":           "AFRICA",
	}

	if location, ok := regionMap[gcpLocation]; ok {
		return location
	}

	return "GLOBAL"
}

// buildCarbonFootprintQuery is accessible from HardenedAdapter
func (ha *HardenedAdapter) buildCarbonFootprintQuery() string {
	return ha.config.buildCarbonFootprintQuery()
}
