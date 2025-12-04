// Package sap provides an ingestion adapter for SAP ERP systems.
// It fetches energy consumption and emissions data from SAP ERP/S4HANA
// including the SAP Sustainability module for comprehensive carbon accounting.
package sap

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/example/offgridflow/internal/ingestion"
	"github.com/google/uuid"
)

// =============================================================================
// Configuration
// =============================================================================

// Config holds SAP adapter configuration.
type Config struct {
	// BaseURL is the SAP API base URL (e.g., https://api.sap.company.com)
	BaseURL string `json:"base_url"`

	// ClientID is the OAuth2 client ID for SAP API authentication
	ClientID string `json:"client_id"`

	// ClientSecret is the OAuth2 client secret
	ClientSecret string `json:"-"` // Excluded from JSON

	// Company is the SAP company code
	Company string `json:"company"`

	// Plant is the optional SAP plant/facility code for filtering
	Plant string `json:"plant,omitempty"`

	// OrgID is the OffGridFlow organization ID to associate activities with
	OrgID string `json:"org_id"`

	// StartDate is the beginning of the date range to fetch (inclusive)
	StartDate time.Time `json:"start_date"`

	// EndDate is the end of the date range to fetch (exclusive)
	EndDate time.Time `json:"end_date"`

	// HTTPClient allows injecting a custom HTTP client for testing
	HTTPClient *http.Client `json:"-"`
}

// Validate checks that the configuration is valid.
func (c *Config) Validate() error {
	if c.BaseURL == "" {
		return fmt.Errorf("sap: base_url is required")
	}
	if c.ClientID == "" {
		return fmt.Errorf("sap: client_id is required")
	}
	if c.ClientSecret == "" {
		return fmt.Errorf("sap: client_secret is required")
	}
	if c.Company == "" {
		return fmt.Errorf("sap: company is required")
	}
	if c.OrgID == "" {
		return fmt.Errorf("sap: org_id is required")
	}
	return nil
}

// =============================================================================
// SAP API Response Types
// =============================================================================

// TokenResponse represents the OAuth2 token response from SAP
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope,omitempty"`
}

// EnergyDataResponse represents the SAP energy consumption API response
type EnergyDataResponse struct {
	Data []EnergyRecord `json:"d"`
}

// EnergyRecord represents energy consumption data from SAP
type EnergyRecord struct {
	RecordID    string  `json:"RecordID"`
	Plant       string  `json:"Plant"`
	Meter       string  `json:"Meter"`
	Date        string  `json:"Date"`
	EnergyType  string  `json:"EnergyType"`
	Quantity    float64 `json:"Quantity"`
	Unit        string  `json:"Unit"`
	CostCenter  string  `json:"CostCenter"`
	Description string  `json:"Description,omitempty"`
}

// EmissionsDataResponse represents the SAP Sustainability emissions API response
type EmissionsDataResponse struct {
	Data []EmissionRecord `json:"d"`
}

// EmissionRecord represents emissions data from SAP Sustainability module
type EmissionRecord struct {
	RecordID     string  `json:"RecordID"`
	Date         string  `json:"Date"`
	Source       string  `json:"Source"`
	EmissionType string  `json:"EmissionType"`
	KgCO2e       float64 `json:"KgCO2e"`
	Scope        string  `json:"Scope"`
	Plant        string  `json:"Plant,omitempty"`
	Description  string  `json:"Description,omitempty"`
}

// =============================================================================
// Adapter Implementation
// =============================================================================

// Adapter ingests energy and emissions data from SAP ERP systems.
type Adapter struct {
	config      Config
	client      *http.Client
	token       string
	tokenExpiry time.Time
}

// NewAdapter creates a new SAP ingestion adapter.
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

	return &Adapter{
		config: cfg,
		client: client,
	}, nil
}

// Ingest fetches energy and emissions data from SAP and returns activities.
func (a *Adapter) Ingest(ctx context.Context) ([]ingestion.Activity, error) {
	// Authenticate first
	if err := a.authenticate(ctx); err != nil {
		return nil, fmt.Errorf("sap: authentication failed: %w", err)
	}

	var allActivities []ingestion.Activity

	// Fetch energy consumption data
	energyActivities, err := a.fetchEnergyData(ctx)
	if err != nil {
		return nil, fmt.Errorf("sap: failed to fetch energy data: %w", err)
	}
	allActivities = append(allActivities, energyActivities...)

	// Fetch emissions data from Sustainability module
	emissionsActivities, err := a.fetchEmissionsData(ctx)
	if err != nil {
		// Log warning but don't fail if emissions module is unavailable
		// Some SAP installations may not have the Sustainability module
		fmt.Printf("sap: warning - failed to fetch emissions data: %v\n", err)
	} else {
		allActivities = append(allActivities, emissionsActivities...)
	}

	return allActivities, nil
}

// authenticate obtains an OAuth2 access token from SAP
func (a *Adapter) authenticate(ctx context.Context) error {
	// Check if token is still valid
	if a.token != "" && time.Now().Before(a.tokenExpiry) {
		return nil
	}

	tokenURL := a.config.BaseURL + "/oauth/token"

	// Prepare OAuth2 request
	data := strings.NewReader("grant_type=client_credentials")
	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, data)
	if err != nil {
		return fmt.Errorf("failed to create token request: %w", err)
	}

	req.SetBasicAuth(a.config.ClientID, a.config.ClientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("authentication failed (status %d): %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("failed to decode token response: %w", err)
	}

	a.token = tokenResp.AccessToken
	// Set expiry with 5 minute buffer
	a.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn-300) * time.Second)

	return nil
}

// fetchEnergyData retrieves energy consumption data from SAP
func (a *Adapter) fetchEnergyData(ctx context.Context) ([]ingestion.Activity, error) {
	endpoint := fmt.Sprintf("%s/api/energy/consumption?company=%s&from=%s&to=%s",
		a.config.BaseURL,
		a.config.Company,
		a.config.StartDate.Format("2006-01-02"),
		a.config.EndDate.Format("2006-01-02"),
	)

	if a.config.Plant != "" {
		endpoint += "&plant=" + a.config.Plant
	}

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+a.token)
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result EnergyDataResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert to activities
	return a.convertEnergyToActivities(result.Data), nil
}

// fetchEmissionsData retrieves emissions data from SAP Sustainability module
func (a *Adapter) fetchEmissionsData(ctx context.Context) ([]ingestion.Activity, error) {
	endpoint := fmt.Sprintf("%s/api/sustainability/emissions?company=%s&from=%s&to=%s",
		a.config.BaseURL,
		a.config.Company,
		a.config.StartDate.Format("2006-01-02"),
		a.config.EndDate.Format("2006-01-02"),
	)

	if a.config.Plant != "" {
		endpoint += "&plant=" + a.config.Plant
	}

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+a.token)
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result EmissionsDataResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert to activities
	return a.convertEmissionsToActivities(result.Data), nil
}

// convertEnergyToActivities transforms SAP energy records into OffGridFlow activities
func (a *Adapter) convertEnergyToActivities(records []EnergyRecord) []ingestion.Activity {
	activities := make([]ingestion.Activity, 0, len(records))
	now := time.Now().UTC()

	for _, record := range records {
		// Parse date
		date, err := time.Parse("2006-01-02", record.Date)
		if err != nil {
			fmt.Printf("sap: warning - invalid date format for record %s: %v\n", record.RecordID, err)
			continue
		}

		// Map SAP energy type to category
		category := mapEnergyType(record.EnergyType)

		// Map SAP unit to OffGridFlow unit
		unit := mapUnit(record.Unit)

		// Determine location from plant code (simplified mapping)
		location := mapPlantToLocation(record.Plant)

		activity := ingestion.Activity{
			ID:          uuid.NewString(),
			Source:      "sap_erp",
			Category:    category,
			MeterID:     record.Meter,
			Location:    location,
			PeriodStart: date,
			PeriodEnd:   date.AddDate(0, 0, 1), // Daily granularity
			Quantity:    record.Quantity,
			Unit:        unit,
			OrgID:       a.config.OrgID,
			Metadata: map[string]string{
				"sap_record_id":   record.RecordID,
				"sap_plant":       record.Plant,
				"sap_meter":       record.Meter,
				"sap_energy_type": record.EnergyType,
				"sap_cost_center": record.CostCenter,
				"sap_company":     a.config.Company,
				"data_source":     "sap_erp_energy",
				"description":     record.Description,
			},
			CreatedAt:     now,
			UpdatedAt:     now,
			ExternalID:    record.RecordID,
			DataQuality:   "measured", // SAP data is typically measured
		}

		activities = append(activities, activity)
	}

	return activities
}

// convertEmissionsToActivities transforms SAP emissions records into OffGridFlow activities
func (a *Adapter) convertEmissionsToActivities(records []EmissionRecord) []ingestion.Activity {
	activities := make([]ingestion.Activity, 0, len(records))
	now := time.Now().UTC()

	for _, record := range records {
		// Parse date
		date, err := time.Parse("2006-01-02", record.Date)
		if err != nil {
			fmt.Printf("sap: warning - invalid date format for record %s: %v\n", record.RecordID, err)
			continue
		}

		// Map scope to category
		category := fmt.Sprintf("emissions_%s_%s", 
			strings.ToLower(record.Scope),
			normalizeEmissionType(record.EmissionType))

		// Determine location from plant code
		location := mapPlantToLocation(record.Plant)
		if location == "" {
			location = "UNKNOWN" // Default if no plant specified
		}

		activity := ingestion.Activity{
			ID:          uuid.NewString(),
			Source:      "sap_sustainability",
			Category:    category,
			Location:    location,
			PeriodStart: date,
			PeriodEnd:   date.AddDate(0, 0, 1), // Daily granularity
			Quantity:    record.KgCO2e,
			Unit:        "kg", // SAP reports in kg CO2e
			OrgID:       a.config.OrgID,
			Metadata: map[string]string{
				"sap_record_id":    record.RecordID,
				"sap_plant":        record.Plant,
				"sap_source":       record.Source,
				"sap_emission_type": record.EmissionType,
				"emission_scope":   record.Scope,
				"sap_company":      a.config.Company,
				"data_source":      "sap_sustainability",
				"description":      record.Description,
			},
			CreatedAt:     now,
			UpdatedAt:     now,
			ExternalID:    record.RecordID,
			DataQuality:   "measured",
		}

		activities = append(activities, activity)
	}

	return activities
}

// =============================================================================
// Helper Functions
// =============================================================================

// mapEnergyType maps SAP energy types to OffGridFlow categories
func mapEnergyType(sapType string) string {
	normalizedType := strings.ToLower(strings.TrimSpace(sapType))
	
	switch {
	case strings.Contains(normalizedType, "electric"):
		return "electricity"
	case strings.Contains(normalizedType, "natural_gas"), strings.Contains(normalizedType, "gas"):
		return "natural_gas"
	case strings.Contains(normalizedType, "diesel"):
		return "diesel"
	case strings.Contains(normalizedType, "fuel_oil"), strings.Contains(normalizedType, "heating_oil"):
		return "fuel_oil"
	case strings.Contains(normalizedType, "steam"):
		return "steam"
	case strings.Contains(normalizedType, "water"):
		return "water"
	default:
		return "energy_" + normalizedType
	}
}

// mapUnit maps SAP units to OffGridFlow units
func mapUnit(sapUnit string) string {
	normalizedUnit := strings.ToLower(strings.TrimSpace(sapUnit))
	
	switch normalizedUnit {
	case "kwh", "kw":
		return "kWh"
	case "mwh", "mw":
		return "MWh"
	case "gj":
		return "GJ"
	case "therm", "therms":
		return "therm"
	case "l", "liter", "litre":
		return "L"
	case "m3", "m³":
		return "m3"
	case "kg":
		return "kg"
	case "tonne", "ton", "mt":
		return "tonne"
	default:
		return sapUnit // Return as-is if no mapping found
	}
}

// mapPlantToLocation maps SAP plant codes to geographic regions
// In production, this would be configured via a mapping table
func mapPlantToLocation(plant string) string {
	// Simplified example mapping based on plant code patterns
	// This should be customized per organization
	
	if plant == "" {
		return "UNKNOWN"
	}
	
	// Example: Plant codes like "US-TX-001" -> "US-TEXAS"
	parts := strings.Split(plant, "-")
	if len(parts) >= 2 {
		countryCode := strings.ToUpper(parts[0])
		stateCode := strings.ToUpper(parts[1])
		return fmt.Sprintf("%s-%s", countryCode, stateCode)
	}
	
	// Example: Plant codes starting with country codes
	if strings.HasPrefix(plant, "US") {
		return "US-UNKNOWN"
	}
	if strings.HasPrefix(plant, "EU") || strings.HasPrefix(plant, "DE") {
		return "EU-CENTRAL"
	}
	if strings.HasPrefix(plant, "CN") {
		return "ASIA-CHINA"
	}
	
	return "GLOBAL" // Default fallback
}

// normalizeEmissionType normalizes emission type strings for categorization
func normalizeEmissionType(emissionType string) string {
	normalized := strings.ToLower(strings.TrimSpace(emissionType))
	normalized = strings.ReplaceAll(normalized, " ", "_")
	normalized = strings.ReplaceAll(normalized, "-", "_")
	return normalized
}
