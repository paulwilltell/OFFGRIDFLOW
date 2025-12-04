// Package azure provides an ingestion adapter for Azure Emissions Impact Dashboard data.
// It fetches carbon emissions data from Azure Cost Management and the
// Azure Emissions Impact Dashboard API.
package azure

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/example/offgridflow/internal/ingestion"
	"github.com/google/uuid"
)

// =============================================================================
// Configuration
// =============================================================================

// Config holds Azure adapter configuration.
type Config struct {
	// TenantID is the Azure AD tenant ID.
	TenantID string `json:"tenant_id"`

	// ClientID is the Azure AD application (client) ID.
	ClientID string `json:"client_id"`

	// ClientSecret is the Azure AD client secret.
	ClientSecret string `json:"-"` // Excluded from JSON

	// SubscriptionID is the Azure subscription ID.
	SubscriptionID string `json:"subscription_id"`

	// OrgID is the OffGridFlow organization ID to associate activities with.
	OrgID string `json:"org_id"`

	// StartDate is the beginning of the date range to fetch (inclusive).
	StartDate time.Time `json:"start_date"`

	// EndDate is the end of the date range to fetch (exclusive).
	EndDate time.Time `json:"end_date"`

	// HTTPClient allows injecting a custom HTTP client for testing.
	HTTPClient *http.Client `json:"-"`
}

// Validate checks that the configuration is valid.
func (c *Config) Validate() error {
	if c.TenantID == "" {
		return fmt.Errorf("azure: tenant_id is required")
	}
	if c.ClientID == "" {
		return fmt.Errorf("azure: client_id is required")
	}
	if c.ClientSecret == "" {
		return fmt.Errorf("azure: client_secret is required")
	}
	if c.SubscriptionID == "" {
		return fmt.Errorf("azure: subscription_id is required")
	}
	if c.OrgID == "" {
		return fmt.Errorf("azure: org_id is required")
	}
	return nil
}

// =============================================================================
// Azure API Response Types
// =============================================================================

// EmissionsResponse represents the Azure Emissions Impact Dashboard API response.
type EmissionsResponse struct {
	Value    []EmissionRecord `json:"value"`
	NextLink string           `json:"nextLink,omitempty"`
}

// EmissionRecord represents a single emissions record from Azure.
type EmissionRecord struct {
	ID                     string    `json:"id"`
	SubscriptionID         string    `json:"subscriptionId"`
	ResourceGroup          string    `json:"resourceGroup"`
	ResourceType           string    `json:"resourceType"`
	ResourceName           string    `json:"resourceName"`
	Region                 string    `json:"region"`
	ServiceName            string    `json:"serviceName"`
	MeterCategory          string    `json:"meterCategory"`
	MeterSubcategory       string    `json:"meterSubcategory"`
	Date                   time.Time `json:"date"`
	Scope1CO2e             float64   `json:"scope1CO2e"`
	Scope2CO2e             float64   `json:"scope2CO2e"`
	Scope3CO2e             float64   `json:"scope3CO2e"`
	TotalCO2e              float64   `json:"totalCO2e"`
	EnergyConsumptionKWh   float64   `json:"energyConsumptionKWh"`
	CarbonIntensity        float64   `json:"carbonIntensity"` // gCO2e/kWh
	RenewableEnergyPercent float64   `json:"renewableEnergyPercent"`
	Currency               string    `json:"currency"`
	Cost                   float64   `json:"cost"`
}

// CostManagementQuery represents a cost query request.
type CostManagementQuery struct {
	Type       string       `json:"type"`
	Timeframe  string       `json:"timeframe"`
	TimePeriod *TimePeriod  `json:"timePeriod,omitempty"`
	Dataset    QueryDataset `json:"dataset"`
}

// TimePeriod represents a custom time period for queries.
type TimePeriod struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// QueryDataset defines what data to retrieve.
type QueryDataset struct {
	Granularity string       `json:"granularity"`
	Aggregation QueryAggr    `json:"aggregation"`
	Grouping    []QueryGroup `json:"grouping,omitempty"`
	Filter      *QueryFilter `json:"filter,omitempty"`
}

// QueryAggr defines aggregation settings.
type QueryAggr struct {
	TotalCost struct {
		Name     string `json:"name"`
		Function string `json:"function"`
	} `json:"totalCost"`
}

// QueryGroup defines grouping settings.
type QueryGroup struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

// QueryFilter defines filter settings.
type QueryFilter struct {
	Dimensions *FilterDimension `json:"dimensions,omitempty"`
}

// FilterDimension defines dimension filters.
type FilterDimension struct {
	Name     string   `json:"name"`
	Operator string   `json:"operator"`
	Values   []string `json:"values"`
}

// OAuthTokenResponse represents the Azure AD token response.
type OAuthTokenResponse struct {
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	AccessToken string `json:"access_token"`
}

// =============================================================================
// Adapter Implementation
// =============================================================================

// Adapter ingests carbon emissions data from Azure.
type Adapter struct {
	config      Config
	client      *http.Client
	accessToken string
	tokenExpiry time.Time
	cred        *azidentity.ClientSecretCredential
}

// NewAdapter creates a new Azure ingestion adapter.
func NewAdapter(cfg Config) (*Adapter, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	cred, err := azidentity.NewClientSecretCredential(cfg.TenantID, cfg.ClientID, cfg.ClientSecret, nil)
	if err != nil {
		return nil, fmt.Errorf("azure: failed to create credential: %w", err)
	}

	return &Adapter{
		config: cfg,
		client: client,
		cred:   cred,
	}, nil
}

// Ingest fetches carbon emissions data from Azure and returns activities.
func (a *Adapter) Ingest(ctx context.Context) ([]ingestion.Activity, error) {
	var records []EmissionRecord
	err := ingestion.WithRetry(ctx, 3, 2*time.Second, func() error {
		if err := a.ensureAuthenticated(ctx); err != nil {
			return err
		}
		var fetchErr error
		records, fetchErr = a.fetchEmissionsData(ctx)
		return fetchErr
	})
	if err != nil {
		return nil, fmt.Errorf("azure: failed to fetch emissions: %w", err)
	}
	return a.convertToActivities(records), nil
}

// ensureAuthenticated obtains or refreshes the Azure AD access token.
func (a *Adapter) ensureAuthenticated(ctx context.Context) error {
	// Check if we have a valid token
	if a.accessToken != "" && time.Now().Before(a.tokenExpiry) {
		return nil
	}

	if a.cred == nil {
		return fmt.Errorf("azure: credential not configured")
	}

	token, err := a.cred.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{"https://management.azure.com/.default"},
	})
	if err != nil {
		return fmt.Errorf("azure: failed to acquire token: %w", err)
	}

	a.accessToken = token.Token
	a.tokenExpiry = token.ExpiresOn

	return nil
}

// fetchEmissionsData retrieves emissions data from Azure Emissions Impact Dashboard.
func (a *Adapter) fetchEmissionsData(ctx context.Context) ([]EmissionRecord, error) {
	// Note: The Emissions Impact Dashboard API endpoint
	// In production, this would be the actual Microsoft Sustainability API
	base := fmt.Sprintf(
		"https://management.azure.com/subscriptions/%s/providers/Microsoft.Carbon/carbonEmissions?api-version=2023-04-01",
		a.config.SubscriptionID,
	)

	startDate := a.config.StartDate.Format("2006-01-02")
	endDate := a.config.EndDate.Format("2006-01-02")
	firstURL := fmt.Sprintf("%s&startDate=%s&endDate=%s", base, startDate, endDate)

	allRecords := make([]EmissionRecord, 0)
	seen := make(map[string]bool)
	nextURL := firstURL
	maxPages := 100

	for page := 0; page < maxPages && nextURL != ""; page++ {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if seen[nextURL] {
			return nil, fmt.Errorf("azure: detected pagination loop on %s", nextURL)
		}
		seen[nextURL] = true

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, nextURL, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+a.accessToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := a.client.Do(req)
		if err != nil {
			return nil, err
		}
		func() {
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				err = fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
				return
			}

			var result EmissionsResponse
			if decodeErr := json.NewDecoder(resp.Body).Decode(&result); decodeErr != nil {
				err = fmt.Errorf("failed to decode response: %w", decodeErr)
				return
			}

			allRecords = append(allRecords, result.Value...)
			nextURL = strings.TrimSpace(result.NextLink)
		}()
		if err != nil {
			return nil, err
		}
	}

	if nextURL != "" {
		return nil, fmt.Errorf("azure: pagination exceeded %d pages", maxPages)
	}

	return allRecords, nil
}

// convertToActivities transforms Azure emissions data into OffGridFlow activities.
func (a *Adapter) convertToActivities(records []EmissionRecord) []ingestion.Activity {
	activities := make([]ingestion.Activity, 0)
	now := time.Now().UTC()

	for _, record := range records {
		// Create activities for each scope
		periodStart := record.Date
		periodEnd := record.Date.AddDate(0, 0, 1) // Daily granularity

		// Map Azure region to OffGridFlow location code
		location := mapAzureRegion(record.Region)

		// Scope 1 emissions (direct)
		if record.Scope1CO2e > 0 {
			activities = append(activities, ingestion.Activity{
				ID:          uuid.NewString(),
				Source:      "azure_emissions",
				Category:    "cloud_compute_scope1",
				Location:    location,
				PeriodStart: periodStart,
				PeriodEnd:   periodEnd,
				Quantity:    record.Scope1CO2e / 1000, // Convert kg to tonnes
				Unit:        "tonne",
				OrgID:       a.config.OrgID,
				Metadata: map[string]string{
					"azure_subscription_id": record.SubscriptionID,
					"azure_resource_group":  record.ResourceGroup,
					"azure_resource_type":   record.ResourceType,
					"azure_service_name":    record.ServiceName,
					"azure_region":          record.Region,
					"emission_scope":        "scope1",
					"renewable_energy_pct":  fmt.Sprintf("%.1f", record.RenewableEnergyPercent),
					"data_source":           "azure_emissions_dashboard",
				},
				CreatedAt:   now,
				DataQuality: "measured",
				ExternalID:  fmt.Sprintf("azure_%s_scope1", record.ID),
			})
		}

		// Scope 2 emissions (indirect - electricity)
		if record.Scope2CO2e > 0 {
			activities = append(activities, ingestion.Activity{
				ID:          uuid.NewString(),
				Source:      "azure_emissions",
				Category:    "cloud_compute_scope2",
				Location:    location,
				PeriodStart: periodStart,
				PeriodEnd:   periodEnd,
				Quantity:    record.Scope2CO2e / 1000, // Convert kg to tonnes
				Unit:        "tonne",
				OrgID:       a.config.OrgID,
				Metadata: map[string]string{
					"azure_subscription_id":  record.SubscriptionID,
					"azure_resource_group":   record.ResourceGroup,
					"azure_resource_type":    record.ResourceType,
					"azure_service_name":     record.ServiceName,
					"azure_region":           record.Region,
					"emission_scope":         "scope2",
					"energy_consumption_kwh": fmt.Sprintf("%.2f", record.EnergyConsumptionKWh),
					"carbon_intensity":       fmt.Sprintf("%.2f", record.CarbonIntensity),
					"renewable_energy_pct":   fmt.Sprintf("%.1f", record.RenewableEnergyPercent),
					"data_source":            "azure_emissions_dashboard",
				},
				CreatedAt:   now,
				DataQuality: "measured",
				ExternalID:  fmt.Sprintf("azure_%s_scope2", record.ID),
			})
		}

		// Scope 3 emissions (value chain)
		if record.Scope3CO2e > 0 {
			activities = append(activities, ingestion.Activity{
				ID:          uuid.NewString(),
				Source:      "azure_emissions",
				Category:    "cloud_compute_scope3",
				Location:    location,
				PeriodStart: periodStart,
				PeriodEnd:   periodEnd,
				Quantity:    record.Scope3CO2e / 1000, // Convert kg to tonnes
				Unit:        "tonne",
				OrgID:       a.config.OrgID,
				Metadata: map[string]string{
					"azure_subscription_id": record.SubscriptionID,
					"azure_resource_group":  record.ResourceGroup,
					"azure_resource_type":   record.ResourceType,
					"azure_service_name":    record.ServiceName,
					"azure_region":          record.Region,
					"emission_scope":        "scope3",
					"data_source":           "azure_emissions_dashboard",
				},
				CreatedAt:   now,
				DataQuality: "measured",
				ExternalID:  fmt.Sprintf("azure_%s_scope3", record.ID),
			})
		}
	}

	return activities
}

// mapAzureRegion converts Azure region codes to OffGridFlow location codes.
func mapAzureRegion(azureRegion string) string {
	regionMap := map[string]string{
		"eastus":             "US-EAST",
		"eastus2":            "US-EAST",
		"westus":             "US-WEST",
		"westus2":            "US-WEST",
		"westus3":            "US-WEST",
		"centralus":          "US-CENTRAL",
		"northcentralus":     "US-CENTRAL",
		"southcentralus":     "US-CENTRAL",
		"westeurope":         "EU-WEST",
		"northeurope":        "EU-WEST",
		"uksouth":            "EU-WEST",
		"ukwest":             "EU-WEST",
		"germanywestcentral": "EU-CENTRAL",
		"francecentral":      "EU-WEST",
		"switzerlandnorth":   "EU-CENTRAL",
		"norwayeast":         "EU-NORTH",
		"swedencentral":      "EU-NORTH",
		"australiaeast":      "ASIA-PACIFIC",
		"australiasoutheast": "ASIA-PACIFIC",
		"japaneast":          "ASIA-PACIFIC",
		"japanwest":          "ASIA-PACIFIC",
		"koreacentral":       "ASIA-PACIFIC",
		"southeastasia":      "ASIA-PACIFIC",
		"eastasia":           "ASIA-PACIFIC",
		"centralindia":       "ASIA-PACIFIC",
		"southindia":         "ASIA-PACIFIC",
		"brazilsouth":        "LATAM",
		"canadacentral":      "US-EAST",
		"canadaeast":         "US-EAST",
		"uaenorth":           "MENA",
		"southafricanorth":   "AFRICA",
	}

	if location, ok := regionMap[strings.ToLower(azureRegion)]; ok {
		return location
	}
	return "GLOBAL"
}

// =============================================================================
// Cost Management Integration
// =============================================================================

// FetchCostData retrieves cost data from Azure Cost Management.
// This can be used to estimate emissions based on spend.
func (a *Adapter) FetchCostData(ctx context.Context) ([]EmissionRecord, error) {
	if err := a.ensureAuthenticated(ctx); err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf(
		"https://management.azure.com/subscriptions/%s/providers/Microsoft.CostManagement/query?api-version=2023-03-01",
		a.config.SubscriptionID,
	)

	query := CostManagementQuery{
		Type:      "Usage",
		Timeframe: "Custom",
		TimePeriod: &TimePeriod{
			From: a.config.StartDate.Format("2006-01-02"),
			To:   a.config.EndDate.Format("2006-01-02"),
		},
		Dataset: QueryDataset{
			Granularity: "Daily",
			Aggregation: QueryAggr{
				TotalCost: struct {
					Name     string `json:"name"`
					Function string `json:"function"`
				}{
					Name:     "Cost",
					Function: "Sum",
				},
			},
			Grouping: []QueryGroup{
				{Type: "Dimension", Name: "ServiceName"},
				{Type: "Dimension", Name: "ResourceLocation"},
			},
		},
	}

	body, _ := json.Marshal(query)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+a.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("cost query failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse and convert cost data to emission estimates
	// Note: This would use spend-based emission factors
	return []EmissionRecord{}, nil
}
