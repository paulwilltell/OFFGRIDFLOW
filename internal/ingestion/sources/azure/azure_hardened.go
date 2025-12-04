//go:build hardened
// +build hardened

// Package azure provides a hardened ingestion adapter for Azure cloud emissions data.
// This version includes rate limiting, pagination, error classification, and observability.
package azure

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/example/offgridflow/internal/ingestion"
	"github.com/google/uuid"
)

// =============================================================================
// Hardened Azure Adapter Configuration
// =============================================================================

// HardenedConfig extends Config with additional hardening options.
type HardenedConfig struct {
	// Base configuration
	Config

	// RateLimitCapacity is the token bucket capacity (default: 60)
	RateLimitCapacity float64

	// RateLimitPerSec is the refill rate in requests/second (default: 3.0)
	// Azure has stricter limits: ~1800 calls/min = 30/sec per docs, but we're conservative
	RateLimitPerSec float64

	// MaxRetries is the maximum number of retry attempts (default: 3)
	MaxRetries int

	// RequestTimeout is the timeout for individual API calls (default: 45s)
	// Azure can be slower than AWS
	RequestTimeout time.Duration

	// TokenRefreshThreshold is how early to refresh tokens (default: 5 min)
	TokenRefreshThreshold time.Duration

	// Logger for structured logging (optional)
	Logger *slog.Logger

	// Observability configuration (optional)
	Observability *ingestion.ObservabilityConfig

	// MaxPages limits pagination to prevent runaway requests (default: 1000)
	MaxPages int

	// MaxPageSize is the max items per page (default: 1000)
	MaxPageSize int

	// FetchEmissionsAPI enables fetching from Emissions Impact Dashboard (default: true)
	FetchEmissionsAPI bool

	// FetchCostManagement enables fetching from Cost Management API (default: false)
	FetchCostManagement bool
}

// NewHardenedConfig creates a new hardened configuration with defaults.
func NewHardenedConfig(cfg Config) *HardenedConfig {
	return &HardenedConfig{
		Config:                cfg,
		RateLimitCapacity:     60,
		RateLimitPerSec:       3.0, // Conservative: Azure allows ~30/sec but we're safe
		MaxRetries:            3,
		RequestTimeout:        45 * time.Second, // Azure is slower
		TokenRefreshThreshold: 5 * time.Minute,
		MaxPages:              1000,
		MaxPageSize:           1000,
		FetchEmissionsAPI:     true,
		FetchCostManagement:   false,
		Logger:                slog.Default(),
		Observability:         ingestion.NewObservabilityConfig("azure"),
	}
}

// Validate checks hardened configuration.
func (hc *HardenedConfig) Validate() error {
	if err := hc.Config.Validate(); err != nil {
		return err
	}

	if hc.RateLimitCapacity <= 0 {
		return fmt.Errorf("azure: rate_limit_capacity must be positive")
	}

	if hc.RateLimitPerSec <= 0 {
		return fmt.Errorf("azure: rate_limit_per_sec must be positive")
	}

	if hc.MaxRetries < 1 {
		return fmt.Errorf("azure: max_retries must be >= 1")
	}

	if hc.RequestTimeout < 1*time.Second {
		return fmt.Errorf("azure: request_timeout must be >= 1s")
	}

	if hc.MaxPageSize <= 0 || hc.MaxPageSize > 10000 {
		return fmt.Errorf("azure: max_page_size must be between 1 and 10000")
	}

	return nil
}

// =============================================================================
// Token Provider (Azure OAuth)
// =============================================================================

// TokenProvider manages Azure OAuth tokens with automatic refresh.
type TokenProvider struct {
	clientID       string
	clientSecret   string
	tenantID       string
	token          string
	tokenExpiresAt time.Time
	refreshBefore  time.Duration
	httpClient     *http.Client
	logger         *slog.Logger
}

// NewTokenProvider creates a token provider for Azure OAuth.
func NewTokenProvider(clientID, clientSecret, tenantID string, refreshBefore time.Duration, logger *slog.Logger) *TokenProvider {
	if logger == nil {
		logger = slog.Default()
	}
	return &TokenProvider{
		clientID:      clientID,
		clientSecret:  clientSecret,
		tenantID:      tenantID,
		refreshBefore: refreshBefore,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// GetToken returns a valid OAuth token, refreshing if necessary.
func (tp *TokenProvider) GetToken(ctx context.Context) (string, error) {
	// Check if token is still valid
	if tp.token != "" && time.Now().Before(tp.tokenExpiresAt.Add(-tp.refreshBefore)) {
		return tp.token, nil
	}

	// Token expired or not yet obtained, refresh it
	token, expiresAt, err := tp.refreshToken(ctx)
	if err != nil {
		return "", ingestion.NewClassifiedError(
			ingestion.ErrorClassAuth,
			"failed to refresh Azure token",
			err,
		)
	}

	tp.token = token
	tp.tokenExpiresAt = expiresAt
	tp.logger.Debug("azure token refreshed", "expires_at", expiresAt)

	return token, nil
}

// refreshToken fetches a new token from Azure AD.
func (tp *TokenProvider) refreshToken(ctx context.Context) (string, time.Time, error) {
	tokenURL := fmt.Sprintf(
		"https://login.microsoftonline.com/%s/oauth2/v2.0/token",
		tp.tenantID,
	)

	data := url.Values{
		"client_id":     {tp.clientID},
		"client_secret": {tp.clientSecret},
		"scope":         {"https://management.azure.com/.default"},
		"grant_type":    {"client_credentials"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", time.Time{}, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := tp.httpClient.Do(req)
	if err != nil {
		return "", time.Time{}, ingestion.ClassifyError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		ce := ingestion.ClassifyHTTPError(resp.StatusCode, string(body))
		return "", time.Time{}, ce
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", time.Time{}, ingestion.NewClassifiedError(
			ingestion.ErrorClassBadRequest,
			"failed to decode token response",
			err,
		)
	}

	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	return tokenResp.AccessToken, expiresAt, nil
}

// =============================================================================
// Hardened Adapter
// =============================================================================

// HardenedAdapter ingests Azure emissions data with production hardening.
type HardenedAdapter struct {
	config        *HardenedConfig
	client        *http.Client
	tokenProvider *TokenProvider
	limiter       *ingestion.RateLimiter
	tracer        *ingestion.InvocationTracer
	metrics       *ingestion.IngestionMetrics
	logger        *slog.Logger
}

// NewHardenedAdapter creates a new hardened Azure adapter.
func NewHardenedAdapter(cfg *HardenedConfig) (*HardenedAdapter, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// Create HTTP client with timeout
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{
			Timeout: cfg.RequestTimeout,
		}
	}

	// Create rate limiter (Azure is stricter than AWS, so lower defaults)
	limiter := ingestion.NewRateLimiter(
		cfg.RateLimitCapacity,
		cfg.RateLimitPerSec,
		100*time.Millisecond,
	)

	// Setup token provider
	tokenProvider := NewTokenProvider(
		cfg.ClientID,
		cfg.ClientSecret,
		cfg.TenantID,
		cfg.TokenRefreshThreshold,
		cfg.Logger,
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
		config:        cfg,
		client:        client,
		tokenProvider: tokenProvider,
		limiter:       limiter,
		tracer:        tracer,
		metrics:       metrics,
		logger:        logger,
	}, nil
}

// =============================================================================
// Ingest - Main Entry Point
// =============================================================================

// Ingest fetches Azure emissions data with full hardening.
func (ha *HardenedAdapter) Ingest(ctx context.Context) ([]ingestion.Activity, error) {
	start := time.Now()
	var activities []ingestion.Activity

	// Ingest from Emissions Impact Dashboard API
	if ha.config.FetchEmissionsAPI {
		emissionsActivities, err := ha.ingestEmissionsAPI(ctx)
		if err != nil {
			ha.recordMetricsFailure("emissions_api", err)
			return nil, err
		}
		activities = append(activities, emissionsActivities...)
		ha.logger.Info("ingested from emissions api", "count", len(emissionsActivities))
	}

	// Optionally ingest from Cost Management API
	if ha.config.FetchCostManagement {
		costActivities, err := ha.ingestCostManagement(ctx)
		if err != nil {
			// Log but don't fail if cost management is optional
			ha.logger.Warn("failed to ingest cost management", "error", err)
		} else {
			activities = append(activities, costActivities...)
			ha.logger.Info("ingested from cost management", "count", len(costActivities))
		}
	}

	// Record success metrics
	latency := time.Since(start)
	ha.recordMetricsSuccess(len(activities), latency)
	ha.logger.Info("azure ingestion complete", "total_activities", len(activities), "latency_ms", latency.Milliseconds())

	return activities, nil
}

// =============================================================================
// Emissions Impact Dashboard API
// =============================================================================

// ingestEmissionsAPI fetches from Emissions Impact Dashboard with hardening.
func (ha *HardenedAdapter) ingestEmissionsAPI(ctx context.Context) ([]ingestion.Activity, error) {
	var activities []ingestion.Activity

	err := ha.tracer.TraceInvocation(ctx, "azure.ingest_emissions_api", func(ctx context.Context) error {
		// Pagination for emissions data
		pagination := ingestion.NewPaginationState(ha.config.MaxPageSize)
		pagination.MaxPages = ha.config.MaxPages

		for !pagination.IsDone() {
			// Rate limit
			if _, err := ha.limiter.Allow(ctx); err != nil {
				return ingestion.NewClassifiedError(ingestion.ErrorClassTransient, "rate limiter cancelled", err)
			}

			// Fetch emissions page with retry
			var records []EmissionRecord
			var nextLink string

			err := ha.retryWithExponentialBackoff(ctx, func() error {
				var err error
				records, nextLink, err = ha.fetchEmissionsPage(ctx, pagination.Cursor)
				return err
			})

			if err != nil {
				ha.tracer.LogIngestionError(ctx, err, "azure-emissions-api")
				return err
			}

			if len(records) == 0 {
				break
			}

			// Convert to activities
			pageActivities := ha.convertEmissionsToActivities(records)
			activities = append(activities, pageActivities...)

			// Update pagination
			if nextLink == "" {
				break
			}
			pagination.SetCursorBased(nextLink)
			_ = pagination.AdvancePage()

			ha.logger.Debug("emissions page processed", "records", len(records), "next_link", nextLink != "")
		}

		ha.logger.Info("emissions api complete", "total_records", len(activities), "pages", pagination.CurrentPage)
		return nil
	})

	return activities, err
}

// fetchEmissionsPage fetches one page of emissions data.
func (ha *HardenedAdapter) fetchEmissionsPage(ctx context.Context, skipToken string) ([]EmissionRecord, string, error) {
	endpoint := fmt.Sprintf(
		"https://management.azure.com/subscriptions/%s/providers/Microsoft.Sustainability/emissionsData?api-version=2021-09-09-preview",
		ha.config.SubscriptionID,
	)

	// Add pagination token if present
	if skipToken != "" {
		endpoint += "&$skiptoken=" + url.QueryEscape(skipToken)
	}

	// Add filters for date range
	endpoint += fmt.Sprintf(
		"&$filter=date%% ge %s and date%% lt %s",
		url.QueryEscape(ha.config.StartDate.Format("2006-01-02")),
		url.QueryEscape(ha.config.EndDate.Format("2006-01-02")),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, "", ingestion.ClassifyError(err)
	}

	// Add authorization header
	token, err := ha.tokenProvider.GetToken(ctx)
	if err != nil {
		return nil, "", err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := ha.client.Do(req)
	if err != nil {
		return nil, "", ingestion.ClassifyError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		ce := ingestion.ClassifyHTTPError(resp.StatusCode, string(body))
		return nil, "", ce
	}

	var result struct {
		Value    []EmissionRecord `json:"value"`
		NextLink string           `json:"nextLink,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, "", ingestion.NewClassifiedError(
			ingestion.ErrorClassBadRequest,
			"failed to parse emissions response",
			err,
		)
	}

	return result.Value, result.NextLink, nil
}

// =============================================================================
// Cost Management API (Optional)
// =============================================================================

// ingestCostManagement fetches from Cost Management API.
func (ha *HardenedAdapter) ingestCostManagement(ctx context.Context) ([]ingestion.Activity, error) {
	var activities []ingestion.Activity

	err := ha.tracer.TraceInvocation(ctx, "azure.ingest_cost_management", func(ctx context.Context) error {
		// Rate limit
		if _, err := ha.limiter.Allow(ctx); err != nil {
			return ingestion.NewClassifiedError(ingestion.ErrorClassTransient, "rate limiter cancelled", err)
		}

		// Build cost query
		query := CostManagementQuery{
			Type:      "Usage",
			Timeframe: "Custom",
			TimePeriod: &TimePeriod{
				From: ha.config.StartDate.Format("2006-01-02T00:00:00Z"),
				To:   ha.config.EndDate.Format("2006-01-02T00:00:00Z"),
			},
			Dataset: QueryDataset{
				Granularity: "Monthly",
				Aggregation: QueryAggr{
					TotalCost: struct {
						Name     string `json:"name"`
						Function string `json:"function"`
					}{
						Name:     "totalCost",
						Function: "Sum",
					},
				},
			},
		}

		// Fetch with retry
		records, err := ha.fetchCostManagementData(ctx, query)
		if err != nil {
			ha.tracer.LogIngestionError(ctx, err, "azure-cost-management")
			return err
		}

		// Convert to activities (use cost-based estimation)
		activities = ha.convertCostToActivities(records)
		ha.logger.Info("cost management api complete", "records", len(records), "activities", len(activities))

		return nil
	})

	return activities, err
}

// fetchCostManagementData queries the Cost Management API.
func (ha *HardenedAdapter) fetchCostManagementData(ctx context.Context, query CostManagementQuery) ([]struct {
	Date time.Time
	Cost float64
}, error) {
	endpoint := fmt.Sprintf(
		"https://management.azure.com/subscriptions/%s/providers/Microsoft.CostManagement/query?api-version=2021-10-01",
		ha.config.SubscriptionID,
	)

	bodyBytes, _ := json.Marshal(query)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, ingestion.ClassifyError(err)
	}

	// Add authorization
	token, err := ha.tokenProvider.GetToken(ctx)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := ha.client.Do(req)
	if err != nil {
		return nil, ingestion.ClassifyError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		ce := ingestion.ClassifyHTTPError(resp.StatusCode, string(body))
		return nil, ce
	}

	var result struct {
		Properties struct {
			Rows [][]interface{} `json:"rows"`
		} `json:"properties"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, ingestion.NewClassifiedError(
			ingestion.ErrorClassBadRequest,
			"failed to parse cost management response",
			err,
		)
	}

	// Parse rows into structured data
	var records []struct {
		Date time.Time
		Cost float64
	}

	for _, row := range result.Properties.Rows {
		if len(row) >= 2 {
			dateStr := fmt.Sprintf("%v", row[0])
			costVal := 0.0
			if costF, ok := row[1].(float64); ok {
				costVal = costF
			}

			date, _ := time.Parse("2006-01-02", dateStr)
			records = append(records, struct {
				Date time.Time
				Cost float64
			}{date, costVal})
		}
	}

	return records, nil
}

// =============================================================================
// Activity Conversion
// =============================================================================

// convertEmissionsToActivities transforms Azure emissions records to activities.
func (ha *HardenedAdapter) convertEmissionsToActivities(records []EmissionRecord) []ingestion.Activity {
	activities := make([]ingestion.Activity, 0, len(records))
	now := time.Now().UTC()

	for _, record := range records {
		// Use total CO2e or calculate from scope 1+2+3
		emissions := record.TotalCO2e
		if emissions <= 0 {
			emissions = record.Scope1CO2e + record.Scope2CO2e + record.Scope3CO2e
		}

		if emissions <= 0 {
			continue // Skip zero-emission records
		}

		activity := ingestion.Activity{
			ID:          uuid.NewString(),
			Source:      "azure_emissions",
			Category:    categorizeAzureService(record.ServiceName, record.ResourceType),
			Location:    mapAzureRegion(record.Region),
			PeriodStart: record.Date,
			PeriodEnd:   record.Date.AddDate(0, 0, 1),
			Quantity:    emissions,
			Unit:        "kg", // Azure reports in kg CO2e
			OrgID:       ha.config.OrgID,
			Metadata: map[string]string{
				"azure_service":           record.ServiceName,
				"azure_meter_category":    record.MeterCategory,
				"azure_meter_subcategory": record.MeterSubcategory,
				"azure_resource_type":     record.ResourceType,
				"azure_resource_name":     record.ResourceName,
				"azure_region":            record.Region,
				"energy_kwh":              fmt.Sprintf("%.2f", record.EnergyConsumptionKWh),
				"carbon_intensity":        fmt.Sprintf("%.2f gCO2e/kWh", record.CarbonIntensity),
				"renewable_percent":       fmt.Sprintf("%.1f%%", record.RenewableEnergyPercent),
				"data_source":             "azure_emissions_api",
			},
			CreatedAt:   now,
			DataQuality: "measured",
			ExternalID:  record.ID,
		}

		activities = append(activities, activity)
	}

	return activities
}

// convertCostToActivities converts cost data to emissions activities (estimated).
func (ha *HardenedAdapter) convertCostToActivities(records []struct {
	Date time.Time
	Cost float64
}) []ingestion.Activity {
	activities := make([]ingestion.Activity, 0, len(records))
	now := time.Now().UTC()

	for _, record := range records {
		if record.Cost <= 0 {
			continue
		}

		// Estimate CO2e from cost (rough: ~100g CO2e per $1 of cloud compute)
		// This is a placeholder; real implementations would use more sophisticated models
		estimatedEmissions := record.Cost * 0.1 // kg CO2e

		activity := ingestion.Activity{
			ID:          uuid.NewString(),
			Source:      "azure_cost_management",
			Category:    "cloud_compute_estimated",
			Location:    "GLOBAL",
			PeriodStart: record.Date,
			PeriodEnd:   record.Date.AddDate(0, 1, 0),
			Quantity:    estimatedEmissions,
			Unit:        "kg",
			OrgID:       ha.config.OrgID,
			Metadata: map[string]string{
				"cost_usd":    fmt.Sprintf("%.2f", record.Cost),
				"data_source": "azure_cost_management",
				"note":        "estimated emissions from cost data",
			},
			CreatedAt:   now,
			DataQuality: "estimated",
			ExternalID:  fmt.Sprintf("azure_cost_%s_%f", record.Date.Format("2006-01-02"), record.Cost),
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

	return fmt.Errorf("azure: max retries exceeded")
}

// =============================================================================
// Metrics
// =============================================================================

func (ha *HardenedAdapter) recordMetricsSuccess(itemCount int, latency time.Duration) {
	if ha.metrics == nil {
		return
	}
	ha.metrics.RecordSuccess(context.Background(), itemCount, latency, "azure")
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
	ha.metrics.RecordFailure(context.Background(), "azure-"+source, errorClass)
}

// =============================================================================
// Helper Functions
// =============================================================================

func categorizeAzureService(serviceName, resourceType string) string {
	categoryMap := map[string]string{
		"Virtual Machines":         "cloud_compute",
		"App Service":              "cloud_compute_serverless",
		"Azure SQL Database":       "cloud_database",
		"Cosmos DB":                "cloud_database",
		"Storage Account":          "cloud_storage",
		"CDN":                      "cloud_cdn",
		"Azure Kubernetes Service": "cloud_compute_kubernetes",
		"Container Instances":      "cloud_compute_containers",
		"Data Factory":             "data_processing",
		"Synapse Analytics":        "data_warehouse",
	}

	if category, ok := categoryMap[serviceName]; ok {
		return category
	}

	// Try resource type fallback
	if strings.Contains(strings.ToLower(resourceType), "virtualMachine") {
		return "cloud_compute"
	}

	return "cloud_other"
}

func mapAzureRegion(azureRegion string) string {
	regionMap := map[string]string{
		"eastus":             "US-EAST",
		"eastus2":            "US-EAST",
		"westus":             "US-WEST",
		"westus2":            "US-WEST",
		"westus3":            "US-WEST",
		"northeurope":        "EU-NORTH",
		"westeurope":         "EU-WEST",
		"uksouth":            "EU-WEST",
		"ukwest":             "EU-WEST",
		"centraleurope":      "EU-CENTRAL",
		"switzerlandnorth":   "EU-CENTRAL",
		"switzerlandwest":    "EU-CENTRAL",
		"japaneast":          "ASIA-PACIFIC",
		"japanwest":          "ASIA-PACIFIC",
		"australiaeast":      "ASIA-PACIFIC",
		"australiasoutheast": "ASIA-PACIFIC",
		"southeastasia":      "ASIA-PACIFIC",
		"eastasia":           "ASIA-PACIFIC",
		"koreacentral":       "ASIA-PACIFIC",
		"koreasoouth":        "ASIA-PACIFIC",
		"southindia":         "ASIA-PACIFIC",
		"eastindia":          "ASIA-PACIFIC",
		"brasilsouth":        "LATAM",
		"uaenorth":           "MENA",
		"southafricanorth":   "AFRICA",
	}

	if location, ok := regionMap[strings.ToLower(azureRegion)]; ok {
		return location
	}

	return "GLOBAL"
}
