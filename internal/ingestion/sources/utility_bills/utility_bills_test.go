package utility_bills

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/example/offgridflow/internal/ingestion"
)

// =============================================================================
// Test Data
// =============================================================================

const sampleCSV = `meter_id,location,period_start,period_end,kwh,category,provider
METER-001,US-WEST,2025-01-01,2025-01-31,12500.5,electricity,PG&E
METER-002,EU-CENTRAL,2025-01-01,2025-01-31,8750.25,electricity,EDF
METER-003,US-EAST,2025-02-01,2025-02-28,9200.0,electricity,ConEd
`

const sampleJSON = `{
  "bills": [
    {
      "meter_id": "METER-001",
      "location": "US-WEST",
      "period_start": "2025-01-01",
      "period_end": "2025-01-31",
      "quantity": 12500.5,
      "unit": "kWh",
      "category": "electricity",
      "provider": "PG&E"
    },
    {
      "meter_id": "METER-002",
      "location": "EU-CENTRAL",
      "period_start": "2025-01-01",
      "period_end": "2025-01-31",
      "quantity": 8750.25,
      "unit": "kWh",
      "category": "electricity",
      "provider": "EDF"
    }
  ]
}`

const invalidCSV = `meter_id,location,period_start,period_end,kwh
METER-001,US-WEST,invalid-date,2025-01-31,12500.5
METER-002,EU-CENTRAL,2025-01-01,2025-01-31,not-a-number
`

// =============================================================================
// Adapter Tests
// =============================================================================

func TestNewAdapter(t *testing.T) {
	config := DefaultConfig("test-org")
	adapter := NewAdapter(config)

	if adapter == nil {
		t.Fatal("NewAdapter returned nil")
	}

	if adapter.config.DefaultOrgID != "test-org" {
		t.Errorf("expected org ID 'test-org', got %s", adapter.config.DefaultOrgID)
	}

	if adapter.logger == nil {
		t.Error("logger should not be nil")
	}

	if adapter.parser == nil {
		t.Error("parser should not be nil")
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig("test-org")

	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"DefaultOrgID", config.DefaultOrgID, "test-org"},
		{"DefaultLocation", config.DefaultLocation, "US"},
		{"EnableDeduplication", config.EnableDeduplication, true},
		{"StrictValidation", config.StrictValidation, false},
		{"MaxConcurrentParsing", config.MaxConcurrentParsing, 4},
		{"MaxFileSize", config.MaxFileSize, int64(50 * 1024 * 1024)},
		{"AutoEnrichLocation", config.AutoEnrichLocation, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s: expected %v, got %v", tt.name, tt.expected, tt.got)
			}
		})
	}
}

func TestIngest(t *testing.T) {
	config := DefaultConfig("test-org")
	adapter := NewAdapter(config)

	ctx := context.Background()
	activities, err := adapter.Ingest(ctx)

	if err != nil {
		t.Errorf("Ingest should not return error: %v", err)
	}

	if len(activities) != 0 {
		t.Errorf("Ingest should return empty slice, got %d activities", len(activities))
	}
}

// =============================================================================
// CSV Ingestion Tests
// =============================================================================

func TestIngestFile_CSV_Success(t *testing.T) {
	store := ingestion.NewInMemoryActivityStore()
	config := DefaultConfig("test-org")
	config.Store = store
	config.EnableDeduplication = false

	adapter := NewAdapter(config)
	ctx := context.Background()

	reader := strings.NewReader(sampleCSV)
	activities, errors, err := adapter.IngestFile(ctx, "test.csv", reader)

	if err != nil {
		t.Fatalf("IngestFile failed: %v", err)
	}

	if len(errors) > 0 {
		t.Errorf("expected no import errors, got %d", len(errors))
		for _, e := range errors {
			t.Logf("  row %d: %s", e.Row, e.Message)
		}
	}

	if len(activities) != 3 {
		t.Fatalf("expected 3 activities, got %d", len(activities))
	}

	// Verify first activity
	act := activities[0]
	if act.MeterID != "METER-001" {
		t.Errorf("expected meter METER-001, got %s", act.MeterID)
	}
	if act.Location != "US-WEST" {
		t.Errorf("expected location US-WEST, got %s", act.Location)
	}
	if act.Quantity != 12500.5 {
		t.Errorf("expected quantity 12500.5, got %f", act.Quantity)
	}
	if act.Unit != "kWh" {
		t.Errorf("expected unit kWh, got %s", act.Unit)
	}
	if act.Category != "electricity" {
		t.Errorf("expected category electricity, got %s", act.Category)
	}
	if act.OrgID != "test-org" {
		t.Errorf("expected org test-org, got %s", act.OrgID)
	}
	if act.Source != string(ingestion.SourceUtilityBill) {
		t.Errorf("expected source utility_bill, got %s", act.Source)
	}

	// Verify provider metadata
	if provider, ok := act.Metadata["provider"]; !ok || provider != "PG&E" {
		t.Errorf("expected provider PG&E in metadata, got %s", provider)
	}

	// Verify storage
	stored, err := store.List(ctx)
	if err != nil {
		t.Fatalf("failed to list stored activities: %v", err)
	}
	if len(stored) != 3 {
		t.Errorf("expected 3 stored activities, got %d", len(stored))
	}
}

func TestIngestFile_CSV_WithErrors(t *testing.T) {
	config := DefaultConfig("test-org")
	config.StrictValidation = false // Allow partial success

	adapter := NewAdapter(config)
	ctx := context.Background()

	reader := strings.NewReader(invalidCSV)
	activities, errors, err := adapter.IngestFile(ctx, "invalid.csv", reader)

	// Should not return fatal error in non-strict mode
	if err != nil {
		t.Fatalf("IngestFile should not fail in non-strict mode: %v", err)
	}

	if len(errors) != 2 {
		t.Errorf("expected 2 import errors, got %d", len(errors))
	}

	if len(activities) != 0 {
		t.Errorf("expected 0 valid activities, got %d", len(activities))
	}
}

func TestIngestFile_CSV_StrictMode(t *testing.T) {
	config := DefaultConfig("test-org")
	config.StrictValidation = true

	adapter := NewAdapter(config)
	ctx := context.Background()

	reader := strings.NewReader(invalidCSV)
	_, _, err := adapter.IngestFile(ctx, "invalid.csv", reader)

	// Should return error in strict mode
	if err == nil {
		t.Error("IngestFile should fail in strict mode with invalid data")
	}
}

// =============================================================================
// JSON Ingestion Tests
// =============================================================================

func TestIngestFile_JSON_Success(t *testing.T) {
	config := DefaultConfig("test-org")
	config.EnableDeduplication = false

	adapter := NewAdapter(config)
	ctx := context.Background()

	reader := strings.NewReader(sampleJSON)
	activities, errors, err := adapter.IngestFile(ctx, "test.json", reader)

	if err != nil {
		t.Fatalf("IngestFile failed: %v", err)
	}

	if len(errors) > 0 {
		t.Errorf("expected no import errors, got %d", len(errors))
	}

	if len(activities) != 2 {
		t.Fatalf("expected 2 activities, got %d", len(activities))
	}

	// Verify first activity
	act := activities[0]
	if act.MeterID != "METER-001" {
		t.Errorf("expected meter METER-001, got %s", act.MeterID)
	}
	if act.Quantity != 12500.5 {
		t.Errorf("expected quantity 12500.5, got %f", act.Quantity)
	}
}

func TestIngestFile_JSON_SingleBill(t *testing.T) {
	config := DefaultConfig("test-org")
	adapter := NewAdapter(config)
	ctx := context.Background()

	singleBillJSON := `{
  "meter_id": "METER-001",
  "location": "US-WEST",
  "period_start": "2025-01-01",
  "period_end": "2025-01-31",
  "quantity": 12500.5,
  "unit": "kWh",
  "category": "electricity"
}`

	reader := strings.NewReader(singleBillJSON)
	activities, errors, err := adapter.IngestFile(ctx, "single.json", reader)

	if err != nil {
		t.Fatalf("IngestFile failed: %v", err)
	}

	if len(errors) > 0 {
		t.Errorf("expected no import errors, got %d", len(errors))
	}

	if len(activities) != 1 {
		t.Fatalf("expected 1 activity, got %d", len(activities))
	}
}

// =============================================================================
// Deduplication Tests
// =============================================================================

func TestDeduplication(t *testing.T) {
	config := DefaultConfig("test-org")
	config.EnableDeduplication = true
	config.DeduplicationWindow = 24 * time.Hour

	adapter := NewAdapter(config)
	ctx := context.Background()

	// First ingestion
	reader1 := strings.NewReader(sampleCSV)
	activities1, _, err := adapter.IngestFile(ctx, "test1.csv", reader1)
	if err != nil {
		t.Fatalf("first ingestion failed: %v", err)
	}

	// Second ingestion with same data
	reader2 := strings.NewReader(sampleCSV)
	activities2, _, err := adapter.IngestFile(ctx, "test2.csv", reader2)
	if err != nil {
		t.Fatalf("second ingestion failed: %v", err)
	}

	// Second ingestion should be deduplicated
	if len(activities2) != 0 {
		t.Errorf("expected 0 activities after deduplication, got %d", len(activities2))
	}

	if len(activities1) != 3 {
		t.Errorf("expected 3 activities in first ingestion, got %d", len(activities1))
	}
}

func TestDeduplication_Disabled(t *testing.T) {
	config := DefaultConfig("test-org")
	config.EnableDeduplication = false

	adapter := NewAdapter(config)
	ctx := context.Background()

	// First ingestion
	reader1 := strings.NewReader(sampleCSV)
	activities1, _, err := adapter.IngestFile(ctx, "test1.csv", reader1)
	if err != nil {
		t.Fatalf("first ingestion failed: %v", err)
	}

	// Second ingestion with same data
	reader2 := strings.NewReader(sampleCSV)
	activities2, _, err := adapter.IngestFile(ctx, "test2.csv", reader2)
	if err != nil {
		t.Fatalf("second ingestion failed: %v", err)
	}

	// Both should have same number of activities
	if len(activities1) != len(activities2) {
		t.Errorf("expected same number of activities, got %d and %d", len(activities1), len(activities2))
	}
}

// =============================================================================
// Enrichment Tests
// =============================================================================

func TestEnrichment_ProviderMapping(t *testing.T) {
	config := DefaultConfig("test-org")
	config.DefaultLocation = "UNKNOWN"
	config.AutoEnrichLocation = true
	config.UtilityProviderMappings = map[string]string{
		"PG&E": "US-CA-PGAE",
		"EDF":  "EU-FR",
	}

	adapter := NewAdapter(config)
	ctx := context.Background()

	reader := strings.NewReader(sampleCSV)
	activities, _, err := adapter.IngestFile(ctx, "test.csv", reader)
	if err != nil {
		t.Fatalf("IngestFile failed: %v", err)
	}

	// First activity should have original location (not default)
	if activities[0].Location != "US-WEST" {
		t.Errorf("expected location US-WEST, got %s", activities[0].Location)
	}
}

func TestEnrichment_CategoryStandardization(t *testing.T) {
	config := DefaultConfig("test-org")
	config.AutoEnrichLocation = true

	tests := []struct {
		input    string
		expected string
	}{
		{"electric", "electricity"},
		{"power", "electricity"},
		{"elec", "electricity"},
		{"gas", "natural_gas"},
		{"natural_gas", "natural_gas"},
		{"water", "water"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := standardizeCategory(tt.input)
			if result != tt.expected {
				t.Errorf("standardizeCategory(%s) = %s, expected %s", tt.input, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// Batch Processing Tests
// =============================================================================

func TestIngestFiles_Multiple(t *testing.T) {
	config := DefaultConfig("test-org")
	config.EnableDeduplication = false
	config.MaxConcurrentParsing = 2

	adapter := NewAdapter(config)
	ctx := context.Background()

	files := map[string]io.Reader{
		"file1.csv":  strings.NewReader(sampleCSV),
		"file2.json": strings.NewReader(sampleJSON),
	}

	result, err := adapter.IngestFiles(ctx, files)
	if err != nil {
		t.Fatalf("IngestFiles failed: %v", err)
	}

	if result.TotalFiles != 2 {
		t.Errorf("expected 2 total files, got %d", result.TotalFiles)
	}

	if result.SuccessFiles != 2 {
		t.Errorf("expected 2 successful files, got %d", result.SuccessFiles)
	}

	if result.FailedFiles != 0 {
		t.Errorf("expected 0 failed files, got %d", result.FailedFiles)
	}

	// CSV has 3 activities, JSON has 2 = 5 total
	if result.TotalActivities != 5 {
		t.Errorf("expected 5 total activities, got %d", result.TotalActivities)
	}

	if result.HasErrors() {
		t.Error("expected no errors")
	}

	if result.SuccessRate() != 1.0 {
		t.Errorf("expected 100%% success rate, got %f", result.SuccessRate())
	}
}

func TestIngestFiles_Empty(t *testing.T) {
	config := DefaultConfig("test-org")
	adapter := NewAdapter(config)
	ctx := context.Background()

	files := map[string]io.Reader{}

	result, err := adapter.IngestFiles(ctx, files)
	if err != nil {
		t.Fatalf("IngestFiles should not fail with empty map: %v", err)
	}

	if result.TotalFiles != 0 {
		t.Errorf("expected 0 files, got %d", result.TotalFiles)
	}
}

// =============================================================================
// Validation Tests
// =============================================================================

func TestValidation_RequiredFields(t *testing.T) {
	config := DefaultConfig("test-org")
	adapter := NewAdapter(config)
	ctx := context.Background()

	// Missing required column
	missingColumn := `location,period_start,period_end,kwh
US-WEST,2025-01-01,2025-01-31,12500.5`

	reader := strings.NewReader(missingColumn)
	_, _, err := adapter.IngestFile(ctx, "missing.csv", reader)

	if err == nil {
		t.Error("expected error for missing required column")
	}
}

// =============================================================================
// Result Types Tests
// =============================================================================

func TestBatchResult_Methods(t *testing.T) {
	startTime := time.Now()
	endTime := startTime.Add(5 * time.Second)

	result := &BatchResult{
		TotalFiles:      10,
		SuccessFiles:    8,
		FailedFiles:     2,
		TotalActivities: 100,
		TotalErrors:     5,
		StartedAt:       startTime,
		CompletedAt:     endTime,
	}

	if !result.HasErrors() {
		t.Error("HasErrors should return true")
	}

	expectedRate := 0.8
	if result.SuccessRate() != expectedRate {
		t.Errorf("expected success rate %f, got %f", expectedRate, result.SuccessRate())
	}

	if result.Duration() != 5*time.Second {
		t.Errorf("expected duration 5s, got %v", result.Duration())
	}

	summary := result.Summary()
	if summary == "" {
		t.Error("Summary should not be empty")
	}
	if !contains(summary, "10 files") {
		t.Errorf("Summary should contain '10 files', got: %s", summary)
	}
}

func TestBatchResult_NoErrors(t *testing.T) {
	result := &BatchResult{
		TotalFiles:   5,
		SuccessFiles: 5,
		FailedFiles:  0,
		TotalErrors:  0,
	}

	if result.HasErrors() {
		t.Error("HasErrors should return false when no errors")
	}

	if result.SuccessRate() != 1.0 {
		t.Errorf("expected 100%% success rate, got %f", result.SuccessRate())
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}
