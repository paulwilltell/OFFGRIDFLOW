package gcp

import (
	"testing"
	"time"
)

func TestGCPAdapter_ConvertToActivities(t *testing.T) {
	cfg := Config{
		ProjectID: "test-project",
		OrgID:     "org-123",
	}

	adapter, err := NewAdapter(cfg)
	if err != nil {
		t.Fatalf("NewAdapter failed: %v", err)
	}

	records := []CarbonRecord{
		{
			BillingAccountID: "billing-123",
			Project: Project{
				ID:     "project-456",
				Name:   "Test Project",
				Number: "123456789",
			},
			Service: Service{
				ID:          "6F81-5844-456A",
				Description: "Compute Engine",
			},
			Location: Location{
				Location: "us-central1",
				Country:  "US",
			},
			UsageMonth:           "202401",
			CarbonFootprintKgCO2: 1500.0,
			Scope1Emissions:      0.0,
			Scope2Emissions:      1200.0,
			Scope3Emissions:      300.0,
			ElectricityKWh:       3000.0,
			CFEScore:             75.5,
			CarbonModelVersion:   "v1",
		},
	}

	activities := adapter.convertToActivities(records)

	if len(activities) == 0 {
		t.Fatal("expected activities, got none")
	}

	// Should have total + scope2 activity
	if len(activities) != 2 {
		t.Errorf("expected 2 activities, got %d", len(activities))
	}

	// Check total activity
	totalActivity := activities[0]
	if totalActivity.Source != "gcp_carbon_footprint" {
		t.Errorf("expected source gcp_carbon_footprint, got %s", totalActivity.Source)
	}
	if totalActivity.OrgID != "org-123" {
		t.Errorf("expected orgID org-123, got %s", totalActivity.OrgID)
	}
	if totalActivity.Quantity != 1.5 { // 1500 kg -> 1.5 tonnes
		t.Errorf("expected quantity 1.5, got %f", totalActivity.Quantity)
	}
	if totalActivity.Unit != "tonne" {
		t.Errorf("expected unit tonne, got %s", totalActivity.Unit)
	}
	if totalActivity.Location != "US-CENTRAL" {
		t.Errorf("expected location US-CENTRAL, got %s", totalActivity.Location)
	}

	// Check metadata
	if totalActivity.Metadata["gcp_project_id"] != "project-456" {
		t.Errorf("expected project_id project-456, got %s", totalActivity.Metadata["gcp_project_id"])
	}
	if totalActivity.Metadata["gcp_service_id"] != "6F81-5844-456A" {
		t.Errorf("expected service_id 6F81-5844-456A, got %s", totalActivity.Metadata["gcp_service_id"])
	}
	if totalActivity.Metadata["cfe_score"] != "75.5" {
		t.Errorf("expected cfe_score 75.5, got %s", totalActivity.Metadata["cfe_score"])
	}

	// Check scope2 activity
	scope2Activity := activities[1]
	if scope2Activity.Category != "cloud_compute_scope2" {
		t.Errorf("expected category cloud_compute_scope2, got %s", scope2Activity.Category)
	}
	if scope2Activity.Quantity != 1.2 { // 1200 kg -> 1.2 tonnes
		t.Errorf("expected quantity 1.2, got %f", scope2Activity.Quantity)
	}
}

func TestGCPAdapter_MapRegion(t *testing.T) {
	tests := []struct {
		gcpRegion string
		expected  string
	}{
		{"us-central1", "US-CENTRAL"},
		{"us-east1", "US-EAST"},
		{"us-west1", "US-WEST"},
		{"europe-west1", "EU-WEST"},
		{"europe-central2", "EU-CENTRAL"},
		{"asia-east1", "ASIA-PACIFIC"},
		{"unknown-region", "GLOBAL"},
	}

	for _, tt := range tests {
		t.Run(tt.gcpRegion, func(t *testing.T) {
			result := mapGCPRegion(tt.gcpRegion)
			if result != tt.expected {
				t.Errorf("mapGCPRegion(%s) = %s, expected %s", tt.gcpRegion, result, tt.expected)
			}
		})
	}
}

func TestGCPAdapter_CategorizeService(t *testing.T) {
	tests := []struct {
		serviceID string
		expected  string
	}{
		{"6F81-5844-456A", "cloud_compute"},
		{"95FF-2EF5-5EA1", "cloud_database"},
		{"152E-C115-5142", "cloud_storage"},
		{"29E7-DA93-CA13", "cloud_compute_serverless"},
		{"UNKNOWN-SERVICE", "cloud_other"},
	}

	for _, tt := range tests {
		t.Run(tt.serviceID, func(t *testing.T) {
			result := categorizeGCPService(tt.serviceID)
			if result != tt.expected {
				t.Errorf("categorizeGCPService(%s) = %s, expected %s", tt.serviceID, result, tt.expected)
			}
		})
	}
}

func TestGCPAdapter_ValidateConfig(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		cfg := Config{
			ProjectID: "test-project",
			OrgID:     "org-123",
		}
		err := cfg.Validate()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("missing project ID", func(t *testing.T) {
		cfg := Config{
			OrgID: "org-123",
		}
		err := cfg.Validate()
		if err == nil {
			t.Error("expected error for missing project ID")
		}
	})

	t.Run("missing org ID", func(t *testing.T) {
		cfg := Config{
			ProjectID: "test-project",
		}
		err := cfg.Validate()
		if err == nil {
			t.Error("expected error for missing org ID")
		}
	})
}

func TestGCPAdapter_PeriodParsing(t *testing.T) {
	cfg := Config{
		ProjectID: "test-project",
		OrgID:     "org-123",
	}

	adapter, _ := NewAdapter(cfg)

	records := []CarbonRecord{
		{
			BillingAccountID: "billing-123",
			Project: Project{
				ID:   "project-456",
				Name: "Test Project",
			},
			Service: Service{
				ID:          "service-789",
				Description: "Test Service",
			},
			Location: Location{
				Location: "us-central1",
			},
			UsageMonth:           "202401",
			CarbonFootprintKgCO2: 100.0,
		},
	}

	activities := adapter.convertToActivities(records)

	if len(activities) == 0 {
		t.Fatal("expected activities, got none")
	}

	activity := activities[0]
	expectedStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	expectedEnd := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)

	if !activity.PeriodStart.Equal(expectedStart) {
		t.Errorf("expected period start %v, got %v", expectedStart, activity.PeriodStart)
	}
	if !activity.PeriodEnd.Equal(expectedEnd) {
		t.Errorf("expected period end %v, got %v", expectedEnd, activity.PeriodEnd)
	}
}
