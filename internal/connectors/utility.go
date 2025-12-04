package connectors

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/example/offgridflow/internal/ingestion"
	"github.com/google/uuid"
)

// UtilityConnectorConfig holds configuration for utility provider connector
type UtilityConnectorConfig struct {
	Provider     string   `json:"provider"`
	APIKey       string   `json:"api_key"`
	AccountID    string   `json:"account_id"`
	BaseURL      string   `json:"base_url"`
	MeterNumbers []string `json:"meter_numbers,omitempty"`
	OrgID        string   `json:"org_id"`
}

// UtilityConnector integrates with utility providers for energy consumption data
type UtilityConnector struct {
	config     UtilityConnectorConfig
	httpClient *http.Client
	store      ingestion.ActivityStore
}

// NewUtilityConnector creates a new utility connector instance
func NewUtilityConnector(config UtilityConnectorConfig) (*UtilityConnector, error) {
	if config.Provider == "" {
		return nil, fmt.Errorf("utility provider is required")
	}
	if config.APIKey == "" {
		return nil, fmt.Errorf("utility API key is required")
	}
	if config.AccountID == "" {
		return nil, fmt.Errorf("utility account ID is required")
	}
	if config.OrgID == "" {
		return nil, fmt.Errorf("utility org_id is required")
	}

	// Set default base URL based on provider
	if config.BaseURL == "" {
		config.BaseURL = getDefaultBaseURL(config.Provider)
	}

	return &UtilityConnector{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// SetStore wires an ActivityStore for persistence.
func (c *UtilityConnector) SetStore(store ingestion.ActivityStore) {
	c.store = store
}

// getDefaultBaseURL returns the default API endpoint for known providers
func getDefaultBaseURL(provider string) string {
	defaultURLs := map[string]string{
		"pge":           "https://api.pge.com/v1",
		"sce":           "https://api.sce.com/v1",
		"sdge":          "https://api.sdge.com/v1",
		"edf":           "https://api.edf.fr/v1",
		"eon":           "https://api.eon.com/v1",
		"national_grid": "https://api.nationalgrid.com/v1",
	}

	if url, ok := defaultURLs[provider]; ok {
		return url
	}
	return ""
}

// Connect validates connectivity with the utility provider
func (c *UtilityConnector) Connect(ctx context.Context) error {
	endpoint := c.config.BaseURL + "/auth/validate"

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to utility provider: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("utility authentication failed: %s", string(body))
	}

	return nil
}

// FetchUsageData retrieves electricity usage data from the utility provider
func (c *UtilityConnector) FetchUsageData(ctx context.Context, startDate, endDate time.Time) ([]UsageRecord, error) {
	endpoint := fmt.Sprintf("%s/accounts/%s/usage?start=%s&end=%s&interval=daily",
		c.config.BaseURL,
		c.config.AccountID,
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"),
	)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch usage data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("utility API error: %s", string(body))
	}

	var result struct {
		Usage []UsageRecord `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Usage, nil
}

// FetchBillingData retrieves billing information including demand charges
func (c *UtilityConnector) FetchBillingData(ctx context.Context, startDate, endDate time.Time) ([]BillingRecord, error) {
	endpoint := fmt.Sprintf("%s/accounts/%s/billing?start=%s&end=%s",
		c.config.BaseURL,
		c.config.AccountID,
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"),
	)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch billing data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("utility API error: %s", string(body))
	}

	var result struct {
		Bills []BillingRecord `json:"bills"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Bills, nil
}

// FetchIntervalData retrieves high-resolution interval data (15-min or hourly)
func (c *UtilityConnector) FetchIntervalData(ctx context.Context, meterNumber string, startDate, endDate time.Time) ([]IntervalRecord, error) {
	endpoint := fmt.Sprintf("%s/meters/%s/intervals?start=%s&end=%s",
		c.config.BaseURL,
		meterNumber,
		startDate.Format(time.RFC3339),
		endDate.Format(time.RFC3339),
	)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch interval data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("utility API error: %s", string(body))
	}

	var result struct {
		Intervals []IntervalRecord `json:"intervals"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Intervals, nil
}

// UsageRecord represents daily energy usage data
type UsageRecord struct {
	Date         time.Time `json:"date"`
	MeterNumber  string    `json:"meter_number"`
	UsageKWh     float64   `json:"usage_kwh"`
	PeakDemandKW float64   `json:"peak_demand_kw,omitempty"`
	ServiceType  string    `json:"service_type"`
	RateSchedule string    `json:"rate_schedule"`
}

// BillingRecord represents utility billing information
type BillingRecord struct {
	BillingDate  time.Time `json:"billing_date"`
	PeriodStart  time.Time `json:"period_start"`
	PeriodEnd    time.Time `json:"period_end"`
	TotalKWh     float64   `json:"total_kwh"`
	TotalCost    float64   `json:"total_cost"`
	DemandCharge float64   `json:"demand_charge"`
	EnergyCharge float64   `json:"energy_charge"`
	Currency     string    `json:"currency"`
}

// IntervalRecord represents high-resolution meter data
type IntervalRecord struct {
	Timestamp   time.Time `json:"timestamp"`
	MeterNumber string    `json:"meter_number"`
	KWh         float64   `json:"kwh"`
	KW          float64   `json:"kw,omitempty"`
	Quality     string    `json:"quality"` // "actual", "estimated", "missing"
}

// Sync performs a full data synchronization from the utility provider
func (c *UtilityConnector) Sync(ctx context.Context, startDate, endDate time.Time) error {
	if c.store == nil {
		return fmt.Errorf("utility connector: activity store not configured")
	}

	// Fetch usage data
	usageData, err := c.FetchUsageData(ctx, startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to fetch usage data: %w", err)
	}

	// Fetch billing data
	billingData, err := c.FetchBillingData(ctx, startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to fetch billing data: %w", err)
	}

	// Fetch interval data for each meter
	for _, meterNum := range c.config.MeterNumbers {
		intervalData, err := c.FetchIntervalData(ctx, meterNum, startDate, endDate)
		if err != nil {
			// Log error but continue with other meters
			fmt.Printf("Warning: failed to fetch interval data for meter %s: %v\n", meterNum, err)
			continue
		}
		_ = intervalData
	}

	activities := make([]ingestion.Activity, 0, len(usageData)+len(billingData))
	defaultLocation := "GLOBAL"

	for _, u := range usageData {
		builder := ingestion.NewActivityBuilder().
			WithID(uuid.NewString()).
			WithSource(ingestion.SourceUtilityBill.String()).
			WithCategory(u.ServiceType).
			WithMeterID(u.MeterNumber).
			WithLocation(defaultLocation).
			WithPeriod(u.Date, u.Date.AddDate(0, 0, 1)).
			WithQuantity(u.UsageKWh, ingestion.UnitKWh.String()).
			WithOrgID(c.config.OrgID).
			WithMetadata("rate_schedule", u.RateSchedule).
			WithMetadata("provider", c.config.Provider).
			WithMetadata("account_id", c.config.AccountID)

		act, buildErr := builder.Build()
		if buildErr != nil {
			return fmt.Errorf("build usage activity: %w", buildErr)
		}
		activities = append(activities, act)
	}

	for _, b := range billingData {
		builder := ingestion.NewActivityBuilder().
			WithID(uuid.NewString()).
			WithSource(ingestion.SourceUtilityBill.String()).
			WithCategory("billing").
			WithLocation(defaultLocation).
			WithPeriod(b.PeriodStart, b.PeriodEnd).
			WithQuantity(b.TotalCost, ingestion.UnitUSD.String()).
			WithOrgID(c.config.OrgID).
			WithMetadata("currency", b.Currency).
			WithMetadata("demand_charge", fmt.Sprintf("%.2f", b.DemandCharge)).
			WithMetadata("energy_charge", fmt.Sprintf("%.2f", b.EnergyCharge)).
			WithMetadata("provider", c.config.Provider).
			WithMetadata("account_id", c.config.AccountID)

		act, buildErr := builder.Build()
		if buildErr != nil {
			return fmt.Errorf("build billing activity: %w", buildErr)
		}
		activities = append(activities, act)
	}

	if len(activities) > 0 {
		if err := c.store.SaveBatch(ctx, activities); err != nil {
			return fmt.Errorf("persist utility activities: %w", err)
		}
	}

	return nil
}

// Close releases resources held by the connector
func (c *UtilityConnector) Close() error {
	c.httpClient.CloseIdleConnections()
	return nil
}
