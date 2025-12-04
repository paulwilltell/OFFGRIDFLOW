package connectors

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/costmanagement/armcostmanagement"
)

// AzureConnector handles Azure emissions and cost data ingestion
type AzureConnector struct {
	client     *armcostmanagement.QueryClient
	httpClient *http.Client
	credential azcore.TokenCredential
	config     AzureConfig
}

// AzureConfig holds Azure connector configuration
type AzureConfig struct {
	SubscriptionID  string
	TenantID        string
	ClientID        string
	ClientSecret    string
	EmissionFactors map[string]float64 // region -> kgCO2e/kWh
}

// AzureUsageRecord represents parsed Azure usage data
type AzureUsageRecord struct {
	Date           time.Time
	SubscriptionID string
	ResourceGroup  string
	ServiceName    string
	Region         string
	MeterCategory  string
	UsageQuantity  float64
	Cost           float64
	EmissionsKgCO2 float64
}

// AzureEmissionsData represents data from Emissions Impact Dashboard API
type AzureEmissionsData struct {
	Date              time.Time
	Region            string
	ServiceName       string
	EmissionsMtCO2e   float64 // Microsoft reports in metric tons
	EnergyConsumedMWh float64
}

// NewAzureConnector creates a new Azure connector
func NewAzureConnector(ctx context.Context, cfg AzureConfig) (*AzureConnector, error) {
	if cfg.EmissionFactors == nil {
		cfg.EmissionFactors = defaultAzureEmissionFactors()
	}

	// Create credential
	cred, err := azidentity.NewClientSecretCredential(
		cfg.TenantID,
		cfg.ClientID,
		cfg.ClientSecret,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure credential: %w", err)
	}

	// Create cost management client
	client, err := armcostmanagement.NewQueryClient(cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cost management client: %w", err)
	}

	return &AzureConnector{
		client:     client,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		credential: cred,
		config:     cfg,
	}, nil
}

// FetchCostAndUsage retrieves cost and usage data from Azure Cost Management API
func (c *AzureConnector) FetchCostAndUsage(ctx context.Context, startDate, endDate time.Time) ([]AzureUsageRecord, error) {
	scope := fmt.Sprintf("/subscriptions/%s", c.config.SubscriptionID)

	// Build query parameters
	query := armcostmanagement.QueryDefinition{
		Type:      ptr(armcostmanagement.ExportTypeUsage),
		Timeframe: ptr(armcostmanagement.TimeframeTypeCustom),
		TimePeriod: &armcostmanagement.QueryTimePeriod{
			From: &startDate,
			To:   &endDate,
		},
		Dataset: &armcostmanagement.QueryDataset{
			Granularity: ptr(armcostmanagement.GranularityTypeDaily),
			Aggregation: map[string]*armcostmanagement.QueryAggregation{
				"totalCost": {
					Name:     ptr("PreTaxCost"),
					Function: ptr(armcostmanagement.FunctionTypeSum),
				},
				"usage": {
					Name:     ptr("UsageQuantity"),
					Function: ptr(armcostmanagement.FunctionTypeSum),
				},
			},
			Grouping: []*armcostmanagement.QueryGrouping{
				{
					Type: ptr(armcostmanagement.QueryColumnTypeDimension),
					Name: ptr("ServiceName"),
				},
				{
					Type: ptr(armcostmanagement.QueryColumnTypeDimension),
					Name: ptr("ResourceLocation"),
				},
				{
					Type: ptr(armcostmanagement.QueryColumnTypeDimension),
					Name: ptr("MeterCategory"),
				},
			},
		},
	}

	result, err := c.client.Usage(ctx, scope, query, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch cost and usage: %w", err)
	}

	return c.parseQueryResult(result.QueryResult), nil
}

func (c *AzureConnector) parseQueryResult(result armcostmanagement.QueryResult) []AzureUsageRecord {
	records := make([]AzureUsageRecord, 0)

	if result.Properties == nil || result.Properties.Rows == nil {
		return records
	}

	for _, row := range result.Properties.Rows {
		if len(row) < 6 {
			continue
		}

		// Parse date
		dateStr, ok := row[0].(string)
		if !ok {
			continue
		}
		date, err := time.Parse("20060102", dateStr)
		if err != nil {
			continue
		}

		// Parse dimensions
		serviceName, _ := row[3].(string)
		region, _ := row[4].(string)
		meterCategory, _ := row[5].(string)

		// Parse metrics
		cost := parseFloat(row[1])
		usage := parseFloat(row[2])

		// Calculate emissions
		emissions := c.calculateEmissions(serviceName, region, usage, meterCategory)

		records = append(records, AzureUsageRecord{
			Date:           date,
			SubscriptionID: c.config.SubscriptionID,
			ServiceName:    serviceName,
			Region:         region,
			MeterCategory:  meterCategory,
			UsageQuantity:  usage,
			Cost:           cost,
			EmissionsKgCO2: emissions,
		})
	}

	return records
}

// FetchEmissionsImpact retrieves emissions data from Azure Emissions Impact Dashboard API
func (c *AzureConnector) FetchEmissionsImpact(ctx context.Context, startDate, endDate time.Time) ([]AzureEmissionsData, error) {
	// Azure Emissions Impact Dashboard API endpoint
	url := fmt.Sprintf(
		"https://management.azure.com/subscriptions/%s/providers/Microsoft.CostManagement/emissions?api-version=2023-03-01&startDate=%s&endDate=%s",
		c.config.SubscriptionID,
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"),
	)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Get access token
	token, err := c.credential.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{"https://management.azure.com/.default"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Value []struct {
			Date      string  `json:"date"`
			Location  string  `json:"location"`
			Service   string  `json:"service"`
			Emissions float64 `json:"emissionsMtCO2e"`
			Energy    float64 `json:"energyConsumedMWh"`
		} `json:"value"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	emissions := make([]AzureEmissionsData, 0, len(result.Value))
	for _, item := range result.Value {
		date, _ := time.Parse("2006-01-02", item.Date)
		emissions = append(emissions, AzureEmissionsData{
			Date:              date,
			Region:            item.Location,
			ServiceName:       item.Service,
			EmissionsMtCO2e:   item.Emissions,
			EnergyConsumedMWh: item.Energy,
		})
	}

	return emissions, nil
}

func (c *AzureConnector) calculateEmissions(service, region string, usage float64, meterCategory string) float64 {
	// For compute services
	if meterCategory == "Virtual Machines" || meterCategory == "Container Instances" {
		// Estimate: 1 vCore-hour ≈ 0.03 kWh
		estimatedKWh := usage * 0.03

		factor, ok := c.config.EmissionFactors[region]
		if !ok {
			factor = 0.4 // Default
		}

		return estimatedKWh * factor
	}

	// For storage
	if meterCategory == "Storage" {
		// Estimate: 1 GB-month ≈ 0.00001 kWh
		estimatedKWh := usage * 0.00001

		factor, ok := c.config.EmissionFactors[region]
		if !ok {
			factor = 0.4
		}

		return estimatedKWh * factor
	}

	return 0.0
}

func defaultAzureEmissionFactors() map[string]float64 {
	return map[string]float64{
		"eastus":        0.3854, // Virginia
		"eastus2":       0.3854, // Virginia
		"westus":        0.2451, // California
		"westus2":       0.2883, // Washington
		"centralus":     0.5821, // Iowa
		"northeurope":   0.2929, // Ireland
		"westeurope":    0.3380, // Netherlands
		"uksouth":       0.2331, // UK
		"southeastasia": 0.4990, // Singapore
		"japaneast":     0.4630, // Tokyo
		"australiaeast": 0.7900, // Australia
	}
}

func parseFloat(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case string:
		var f float64
		fmt.Sscanf(val, "%f", &f)
		return f
	default:
		return 0.0
	}
}

func ptr[T any](v T) *T {
	return &v
}

// Close cleans up resources
func (c *AzureConnector) Close() error {
	return nil
}
