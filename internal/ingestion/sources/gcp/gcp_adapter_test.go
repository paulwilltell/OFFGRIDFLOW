package gcp

import (
	"testing"
	"time"
)

func TestConvertToActivities_MapsRegionAndScopes(t *testing.T) {
	adapter := &Adapter{config: Config{OrgID: "org-1"}}
	records := []CarbonRecord{
		{
			BillingAccountID: "bill-1",
			Project: Project{ID: "proj-1", Name: "Demo"},
			Service: Service{ID: "6F81-5844-456A", Description: "Compute"},
			Location: Location{Location: "us-central1", Country: "US"},
			UsageMonth:           "202401",
			CarbonFootprintKgCO2: 1000,
			Scope2Emissions:      250,
			CFEScore:             90,
		},
	}

	activities := adapter.convertToActivities(records)
	if len(activities) != 2 {
		t.Fatalf("expected total and scope2 activities, got %d", len(activities))
	}

	total := activities[0]
	if total.Unit != "tonne" {
		t.Errorf("expected unit tonne, got %s", total.Unit)
	}
	if total.Location != "US-CENTRAL" {
		t.Errorf("expected mapped region US-CENTRAL, got %s", total.Location)
	}
	if total.Metadata["gcp_project_id"] != "proj-1" {
		t.Errorf("project metadata missing")
	}

	scope2 := activities[1]
	if scope2.Category != "cloud_compute_scope2" {
		t.Errorf("expected scope2 category, got %s", scope2.Category)
	}
	if scope2.Quantity <= 0 {
		t.Errorf("expected positive quantity")
	}
}

func TestToStringAndToFloatHelpers(t *testing.T) {
	if toString(nil) != "" {
		t.Errorf("expected empty string for nil")
	}
	if toString(123) != "123" {
		t.Errorf("expected numeric to string conversion")
	}
	if toFloat("12.5") != 12.5 {
		t.Errorf("expected 12.5 from string")
	}
	if toFloat(int64(5)) != 5 {
		t.Errorf("expected conversion from int64")
	}
}

func TestMapGCPRegionDefault(t *testing.T) {
	if mapGCPRegion("unknown-region") != "GLOBAL" {
		t.Errorf("expected fallback GLOBAL")
	}
	if mapGCPRegion("europe-west1") != "EU-WEST" {
		t.Errorf("expected mapping to EU-WEST")
	}
}

func TestParseUsageMonth(t *testing.T) {
	adapter := &Adapter{config: Config{OrgID: "org"}}
	records := []CarbonRecord{{UsageMonth: "202401", CarbonFootprintKgCO2: 100, Project: Project{ID: "p"}, Service: Service{ID: "svc"}, Location: Location{Location: "us-east1"}}}

	acts := adapter.convertToActivities(records)
	if len(acts) == 0 {
		t.Fatalf("expected activities to be generated")
	}
	if acts[0].PeriodStart != time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC) {
		t.Errorf("unexpected period start: %v", acts[0].PeriodStart)
	}
}
