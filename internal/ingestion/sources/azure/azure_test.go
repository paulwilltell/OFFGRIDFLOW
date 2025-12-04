package azure

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

// MockHTTPClient for testing
type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

func TestAzureAdapter_ConvertToActivities(t *testing.T) {
	cfg := Config{
		TenantID:       "tenant-123",
		ClientID:       "client-456",
		ClientSecret:   "secret-789",
		SubscriptionID: "sub-abc",
		OrgID:          "org-xyz",
	}

	adapter := &Adapter{
		config: cfg,
		cred:   nil, // Not needed for conversion test
	}

	records := []EmissionRecord{
		{
			ID:             "record-1",
			SubscriptionID: "sub-abc",
			ResourceGroup:  "rg-test",
			ResourceType:   "Microsoft.Compute/virtualMachines",
			ServiceName:    "Virtual Machines",
			Region:         "eastus",
			Date:           time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			Scope1CO2e:     0.0,
			Scope2CO2e:     100.5,
			Scope3CO2e:     20.3,
			TotalCO2e:      120.8,
			EnergyConsumptionKWh:   250.0,
			CarbonIntensity:        402.0,
			RenewableEnergyPercent: 35.5,
			Cost:                   150.75,
			Currency:               "USD",
		},
	}

	activities := adapter.convertToActivities(records)

	// Should create 2 activities: scope2 and scope3 (scope1 is 0)
	if len(activities) != 2 {
		t.Fatalf("expected 2 activities, got %d", len(activities))
	}

	// Check scope2 activity
	scope2 := activities[0]
	if scope2.Source != "azure_emissions" {
		t.Errorf("expected source azure_emissions, got %s", scope2.Source)
	}
	if scope2.Category != "cloud_compute_scope2" {
		t.Errorf("expected category cloud_compute_scope2, got %s", scope2.Category)
	}
	if scope2.Quantity != 0.1005 { // 100.5 kg -> 0.1005 tonnes
		t.Errorf("expected quantity 0.1005, got %f", scope2.Quantity)
	}
	if scope2.Unit != "tonne" {
		t.Errorf("expected unit tonne, got %s", scope2.Unit)
	}
	if scope2.Location != "US-EAST" {
		t.Errorf("expected location US-EAST, got %s", scope2.Location)
	}
	if scope2.OrgID != "org-xyz" {
		t.Errorf("expected orgID org-xyz, got %s", scope2.OrgID)
	}

	// Check metadata
	if scope2.Metadata["azure_subscription_id"] != "sub-abc" {
		t.Errorf("expected subscription sub-abc, got %s", scope2.Metadata["azure_subscription_id"])
	}
	if scope2.Metadata["azure_resource_group"] != "rg-test" {
		t.Errorf("expected resource group rg-test, got %s", scope2.Metadata["azure_resource_group"])
	}
	if scope2.Metadata["emission_scope"] != "scope2" {
		t.Errorf("expected emission_scope scope2, got %s", scope2.Metadata["emission_scope"])
	}
	if scope2.Metadata["renewable_energy_pct"] != "35.5" {
		t.Errorf("expected renewable_energy_pct 35.5, got %s", scope2.Metadata["renewable_energy_pct"])
	}

	// Check scope3 activity
	scope3 := activities[1]
	if scope3.Category != "cloud_compute_scope3" {
		t.Errorf("expected category cloud_compute_scope3, got %s", scope3.Category)
	}
	// Allow for floating point precision
	if diff := scope3.Quantity - 0.0203; diff > 0.0001 || diff < -0.0001 {
		t.Errorf("expected quantity ~0.0203, got %f", scope3.Quantity)
	}
}

func TestAzureAdapter_MapRegion(t *testing.T) {
	tests := []struct {
		azureRegion string
		expected    string
	}{
		{"eastus", "US-EAST"},
		{"eastus2", "US-EAST"},
		{"westus", "US-WEST"},
		{"centralus", "US-CENTRAL"},
		{"westeurope", "EU-WEST"},
		{"northeurope", "EU-WEST"},
		{"uksouth", "EU-WEST"},
		{"germanywestcentral", "EU-CENTRAL"},
		{"australiaeast", "ASIA-PACIFIC"},
		{"brazilsouth", "LATAM"},
		{"unknown", "GLOBAL"},
	}

	for _, tt := range tests {
		t.Run(tt.azureRegion, func(t *testing.T) {
			result := mapAzureRegion(tt.azureRegion)
			if result != tt.expected {
				t.Errorf("mapAzureRegion(%s) = %s, expected %s", tt.azureRegion, result, tt.expected)
			}
		})
	}
}

func TestAzureAdapter_ValidateConfig(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		cfg := Config{
			TenantID:       "tenant-123",
			ClientID:       "client-456",
			ClientSecret:   "secret-789",
			SubscriptionID: "sub-abc",
			OrgID:          "org-xyz",
		}
		err := cfg.Validate()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("missing tenant ID", func(t *testing.T) {
		cfg := Config{
			ClientID:       "client-456",
			ClientSecret:   "secret-789",
			SubscriptionID: "sub-abc",
			OrgID:          "org-xyz",
		}
		err := cfg.Validate()
		if err == nil {
			t.Error("expected error for missing tenant ID")
		}
	})

	t.Run("missing client secret", func(t *testing.T) {
		cfg := Config{
			TenantID:       "tenant-123",
			ClientID:       "client-456",
			SubscriptionID: "sub-abc",
			OrgID:          "org-xyz",
		}
		err := cfg.Validate()
		if err == nil {
			t.Error("expected error for missing client secret")
		}
	})

	t.Run("missing org ID", func(t *testing.T) {
		cfg := Config{
			TenantID:       "tenant-123",
			ClientID:       "client-456",
			ClientSecret:   "secret-789",
			SubscriptionID: "sub-abc",
		}
		err := cfg.Validate()
		if err == nil {
			t.Error("expected error for missing org ID")
		}
	})
}

func TestAzureAdapter_FetchEmissionsWithPagination(t *testing.T) {
	cfg := Config{
		TenantID:       "tenant-123",
		ClientID:       "client-456",
		ClientSecret:   "secret-789",
		SubscriptionID: "sub-abc",
		OrgID:          "org-xyz",
		StartDate:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:        time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
	}

	callCount := 0
	mockClient := &http.Client{
		Transport: &mockTransport{
			roundTrip: func(req *http.Request) (*http.Response, error) {
				callCount++
				
				var responseBody string
				if callCount == 1 {
					// First page
					responseBody = `{
						"value": [
							{
								"id": "record-1",
								"subscriptionId": "sub-abc",
								"resourceGroup": "rg-test",
								"serviceName": "Virtual Machines",
								"region": "eastus",
								"date": "2024-01-15T00:00:00Z",
								"scope2CO2e": 100.0,
								"totalCO2e": 100.0
							}
						],
						"nextLink": "https://management.azure.com/page2"
					}`
				} else {
					// Second page
					responseBody = `{
						"value": [
							{
								"id": "record-2",
								"subscriptionId": "sub-abc",
								"resourceGroup": "rg-test",
								"serviceName": "Storage",
								"region": "westus",
								"date": "2024-01-16T00:00:00Z",
								"scope2CO2e": 50.0,
								"totalCO2e": 50.0
							}
						]
					}`
				}

				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(responseBody)),
					Header:     make(http.Header),
				}, nil
			},
		},
	}

	cfg.HTTPClient = mockClient
	adapter := &Adapter{
		config:      cfg,
		client:      mockClient,
		accessToken: "test-token",
		tokenExpiry: time.Now().Add(1 * time.Hour),
	}

	records, err := adapter.fetchEmissionsData(context.Background())
	if err != nil {
		t.Fatalf("fetchEmissionsData failed: %v", err)
	}

	if len(records) != 2 {
		t.Errorf("expected 2 records from pagination, got %d", len(records))
	}

	if callCount != 2 {
		t.Errorf("expected 2 API calls for pagination, got %d", callCount)
	}
}

type mockTransport struct {
	roundTrip func(req *http.Request) (*http.Response, error)
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTrip(req)
}

func TestAzureAdapter_EnsureAuthenticated(t *testing.T) {
	t.Run("uses cached token", func(t *testing.T) {
		cfg := Config{
			TenantID:       "tenant-123",
			ClientID:       "client-456",
			ClientSecret:   "secret-789",
			SubscriptionID: "sub-abc",
			OrgID:          "org-xyz",
		}

		adapter := &Adapter{
			config:      cfg,
			accessToken: "cached-token",
			tokenExpiry: time.Now().Add(30 * time.Minute),
		}

		err := adapter.ensureAuthenticated(context.Background())
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if adapter.accessToken != "cached-token" {
			t.Error("token should not have changed")
		}
	})
}
