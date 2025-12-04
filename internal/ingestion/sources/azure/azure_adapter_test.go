package azure

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestIngest_WithMockClientProducesScopeActivities(t *testing.T) {
	mockResp := `{
		"value": [{
			"id": "rec-1",
			"subscriptionId": "sub-123",
			"resourceGroup": "rg-1",
			"resourceType": "vm",
			"resourceName": "vm-1",
			"region": "eastus",
			"serviceName": "Compute",
			"meterCategory": "Compute",
			"meterSubcategory": "VM",
			"date": "2024-01-15T00:00:00Z",
			"scope1CO2e": 10.0,
			"scope2CO2e": 20.0,
			"scope3CO2e": 5.0,
			"totalCO2e": 35.0,
			"energyConsumptionKWh": 50.0,
			"carbonIntensity": 15.5,
			"renewableEnergyPercent": 60.0,
			"currency": "USD",
			"cost": 12.34
		}]
	}`

	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Header.Get("Authorization") == "" {
			t.Fatalf("expected bearer token on request")
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(bytes.NewBufferString(mockResp)),
		}, nil
	})}

	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)
	adapter, err := NewAdapter(Config{
		TenantID:       "tenant",
		ClientID:       "client",
		ClientSecret:   "super-secret-value",
		SubscriptionID: "sub-123",
		OrgID:          "org-abc",
		StartDate:      start,
		EndDate:        end,
		HTTPClient:     client,
	})
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}

	adapter.accessToken = "token"
	adapter.tokenExpiry = time.Now().Add(time.Hour)

	activities, err := adapter.Ingest(context.Background())
	if err != nil {
		t.Fatalf("ingest failed: %v", err)
	}

	if len(activities) != 3 {
		t.Fatalf("expected three scope activities, got %d", len(activities))
	}

	scope2 := activities[1]
	if scope2.Category != "cloud_compute_scope2" {
		t.Errorf("expected scope2 category, got %s", scope2.Category)
	}
	if scope2.Location != "US-EAST" {
		t.Errorf("expected region mapping to US-EAST, got %s", scope2.Location)
	}
	if scope2.OrgID != "org-abc" {
		t.Errorf("expected org id propagation")
	}
	if scope2.Quantity <= 0 {
		t.Errorf("expected positive quantity, got %f", scope2.Quantity)
	}
}

func TestConvertToActivities_BuildsMetadata(t *testing.T) {
	adapter := &Adapter{config: Config{OrgID: "org-meta"}}
	records := []EmissionRecord{
		{
			ID:             "rec-1",
			SubscriptionID: "sub-1",
			ResourceGroup:  "rg",
			ResourceType:   "db",
			ResourceName:   "db-1",
			Region:         "westeurope",
			ServiceName:    "Postgres",
			Scope1CO2e:     1.0,
			Scope2CO2e:     2.0,
			Scope3CO2e:     0,
			Date:           time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	acts := adapter.convertToActivities(records)
	if len(acts) != 2 {
		t.Fatalf("expected two activities (scope1 & scope2), got %d", len(acts))
	}
	if acts[0].Metadata["azure_subscription_id"] != "sub-1" {
		t.Errorf("missing subscription metadata")
	}
	if acts[0].Location != "EU-WEST" {
		t.Errorf("expected mapped region EU-WEST, got %s", acts[0].Location)
	}
}
