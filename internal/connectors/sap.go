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

// SAPConnectorConfig holds configuration for SAP connector
type SAPConnectorConfig struct {
	BaseURL      string `json:"base_url"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Company      string `json:"company"`
	Plant        string `json:"plant,omitempty"`
	OrgID        string `json:"org_id"`
}

// SAPConnector integrates with SAP ERP systems for energy and emissions data
type SAPConnector struct {
	config      SAPConnectorConfig
	httpClient  *http.Client
	token       string
	tokenExpiry time.Time
	store       ingestion.ActivityStore
}

// NewSAPConnector creates a new SAP connector instance
func NewSAPConnector(config SAPConnectorConfig) (*SAPConnector, error) {
	if config.BaseURL == "" {
		return nil, fmt.Errorf("SAP base URL is required")
	}
	if config.ClientID == "" || config.ClientSecret == "" {
		return nil, fmt.Errorf("SAP credentials are required")
	}
	if config.OrgID == "" {
		return nil, fmt.Errorf("SAP org_id is required")
	}

	return &SAPConnector{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// SetStore wires an ActivityStore for persistence.
func (c *SAPConnector) SetStore(store ingestion.ActivityStore) {
	c.store = store
}

// Connect establishes connection and authenticates with SAP
func (c *SAPConnector) Connect(ctx context.Context) error {
	// OAuth2 token request
	tokenURL := c.config.BaseURL + "/oauth/token"

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create token request: %w", err)
	}

	req.SetBasicAuth(c.config.ClientID, c.config.ClientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to authenticate with SAP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("SAP authentication failed: %s", string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("failed to decode token response: %w", err)
	}

	c.token = tokenResp.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return nil
}

// FetchEnergyData retrieves energy consumption data from SAP
func (c *SAPConnector) FetchEnergyData(ctx context.Context, startDate, endDate time.Time) ([]EnergyRecord, error) {
	if time.Now().After(c.tokenExpiry) {
		if err := c.Connect(ctx); err != nil {
			return nil, err
		}
	}

	endpoint := fmt.Sprintf("%s/api/energy/consumption?company=%s&from=%s&to=%s",
		c.config.BaseURL,
		c.config.Company,
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"),
	)

	if c.config.Plant != "" {
		endpoint += "&plant=" + c.config.Plant
	}

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch energy data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("SAP API error: %s", string(body))
	}

	var result struct {
		Data []EnergyRecord `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Data, nil
}

// FetchEmissionsData retrieves emissions data from SAP Sustainability module
func (c *SAPConnector) FetchEmissionsData(ctx context.Context, startDate, endDate time.Time) ([]EmissionRecord, error) {
	if time.Now().After(c.tokenExpiry) {
		if err := c.Connect(ctx); err != nil {
			return nil, err
		}
	}

	endpoint := fmt.Sprintf("%s/api/sustainability/emissions?company=%s&from=%s&to=%s",
		c.config.BaseURL,
		c.config.Company,
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"),
	)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch emissions data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("SAP API error: %s", string(body))
	}

	var result struct {
		Data []EmissionRecord `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Data, nil
}

// EnergyRecord represents energy consumption data from SAP
type EnergyRecord struct {
	Date       time.Time `json:"date"`
	Plant      string    `json:"plant"`
	Meter      string    `json:"meter"`
	EnergyType string    `json:"energy_type"`
	Quantity   float64   `json:"quantity"`
	Unit       string    `json:"unit"`
	CostCenter string    `json:"cost_center"`
}

// EmissionRecord represents emissions data from SAP
type EmissionRecord struct {
	Date         time.Time `json:"date"`
	Source       string    `json:"source"`
	EmissionType string    `json:"emission_type"`
	KgCO2e       float64   `json:"kg_co2e"`
	Scope        string    `json:"scope"`
	Plant        string    `json:"plant"`
}

// Sync performs a full data synchronization from SAP
func (c *SAPConnector) Sync(ctx context.Context, startDate, endDate time.Time) error {
	if c.store == nil {
		return fmt.Errorf("sap connector: activity store not configured")
	}

	if err := c.Connect(ctx); err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}

	// Fetch energy data
	energyData, err := c.FetchEnergyData(ctx, startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to fetch energy data: %w", err)
	}

	// Fetch emissions data
	emissionsData, err := c.FetchEmissionsData(ctx, startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to fetch emissions data: %w", err)
	}

	location := c.config.Plant
	if location == "" {
		location = "GLOBAL"
	}

	activities := make([]ingestion.Activity, 0, len(energyData)+len(emissionsData))

	for _, e := range energyData {
		builder := ingestion.NewActivityBuilder().
			WithID(uuid.NewString()).
			WithSource(ingestion.SourceUtilityBill.String()).
			WithCategory(e.EnergyType).
			WithMeterID(e.Meter).
			WithLocation(location).
			WithPeriod(e.Date, e.Date.AddDate(0, 0, 1)).
			WithQuantity(e.Quantity, e.Unit).
			WithOrgID(c.config.OrgID).
			WithMetadata("company", c.config.Company).
			WithMetadata("cost_center", e.CostCenter)

		act, buildErr := builder.Build()
		if buildErr != nil {
			return fmt.Errorf("build sap energy activity: %w", buildErr)
		}
		activities = append(activities, act)
	}

	for _, em := range emissionsData {
		builder := ingestion.NewActivityBuilder().
			WithID(uuid.NewString()).
			WithSource("sap_emissions").
			WithCategory(em.Scope).
			WithLocation(location).
			WithPeriod(em.Date, em.Date.AddDate(0, 0, 1)).
			WithQuantity(em.KgCO2e, ingestion.UnitKg.String()).
			WithOrgID(c.config.OrgID).
			WithMetadata("company", c.config.Company).
			WithMetadata("plant", em.Plant).
			WithMetadata("emission_type", em.EmissionType).
			WithMetadata("source", em.Source)

		act, buildErr := builder.Build()
		if buildErr != nil {
			return fmt.Errorf("build sap emission activity: %w", buildErr)
		}
		activities = append(activities, act)
	}

	if len(activities) > 0 {
		if err := c.store.SaveBatch(ctx, activities); err != nil {
			return fmt.Errorf("persist sap activities: %w", err)
		}
	}

	return nil
}

// Close releases resources held by the connector
func (c *SAPConnector) Close() error {
	c.httpClient.CloseIdleConnections()
	return nil
}
