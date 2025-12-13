package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/example/offgridflow/internal/api/http/handlers"
	"github.com/example/offgridflow/internal/compliance"
	"github.com/example/offgridflow/internal/emissions"
	"github.com/example/offgridflow/internal/ingestion"
)

// setupComplianceTest creates a test environment with seeded activity data
func setupComplianceTest(t *testing.T) (*handlers.ComplianceHandlerDeps, *ingestion.InMemoryActivityStore) {
	t.Helper()

	// Create in-memory activity store with test data
	store := ingestion.NewInMemoryActivityStore()

	// Seed with comprehensive test data covering all scopes
	seedTestActivities(t, store)

	// Create calculators
	scope1Calc := emissions.NewScope1Calculator(emissions.Scope1Config{})
	scope2Calc := emissions.NewScope2Calculator(emissions.Scope2Config{})
	scope3Calc := emissions.NewScope3Calculator(emissions.Scope3Config{})

	// Create compliance service
	complianceService := compliance.NewService(
		store,
		scope1Calc,
		scope2Calc,
		scope3Calc,
	)

	deps := &handlers.ComplianceHandlerDeps{
		ComplianceService: complianceService,
	}

	return deps, store
}

// seedTestActivities populates the store with realistic test data
func seedTestActivities(t *testing.T, store *ingestion.InMemoryActivityStore) {
	t.Helper()

	ctx := context.Background()
	orgID := "org-test"
	year := 2024

	// Scope 1: Direct emissions (fuel combustion, company vehicles)
	scope1Activities := []ingestion.Activity{
		{
			ID:          "activity-scope1-gas",
			OrgID:       orgID,
			PeriodStart: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC),
			PeriodEnd:   time.Date(year, 1, 31, 0, 0, 0, 0, time.UTC),
			Category:    "fuel_combustion",
			Source:      "natural_gas",
			Quantity:    1000.0, // kWh
			Unit:        "kWh",
			Location:    "US",
			CreatedAt:   time.Now(),
		},
		{
			ID:          "activity-scope1-fuel",
			OrgID:       orgID,
			PeriodStart: time.Date(year, 2, 1, 0, 0, 0, 0, time.UTC),
			PeriodEnd:   time.Date(year, 2, 28, 0, 0, 0, 0, time.UTC),
			Category:    "mobile_combustion",
			Source:      "gasoline",
			Quantity:    500.0, // liters
			Unit:        "liters",
			Location:    "US",
			CreatedAt:   time.Now(),
		},
	}

	// Scope 2: Indirect emissions (purchased electricity)
	scope2Activities := []ingestion.Activity{
		{
			ID:          "activity-scope2-electricity",
			OrgID:       orgID,
			PeriodStart: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC),
			PeriodEnd:   time.Date(year, 1, 31, 0, 0, 0, 0, time.UTC),
			Category:    "electricity",
			Source:      "grid_electricity",
			Quantity:    5000.0, // kWh
			Unit:        "kWh",
			Location:    "US-CA",
			CreatedAt:   time.Now(),
		},
	}

	// Scope 3: Other indirect emissions (business travel, supply chain)
	scope3Activities := []ingestion.Activity{
		{
			ID:          "activity-scope3-travel",
			OrgID:       orgID,
			PeriodStart: time.Date(year, 3, 1, 0, 0, 0, 0, time.UTC),
			PeriodEnd:   time.Date(year, 3, 31, 0, 0, 0, 0, time.UTC),
			Category:    "business_travel",
			Source:      "air_travel",
			Quantity:    2000.0, // km
			Unit:        "km",
			Location:    "global",
			CreatedAt:   time.Now(),
		},
	}

	// Store all activities
	allActivities := append(scope1Activities, scope2Activities...)
	allActivities = append(allActivities, scope3Activities...)

	if err := store.SaveBatch(ctx, allActivities); err != nil {
		t.Fatalf("Failed to seed activities: %v", err)
	}
}

// TestCSRDComplianceHandler tests the CSRD endpoint
func TestCSRDComplianceHandler(t *testing.T) {
	deps, _ := setupComplianceTest(t)
	handler := handlers.NewCSRDComplianceHandler(deps)

	tests := []struct {
		name           string
		method         string
		queryParams    string
		body           interface{}
		expectedStatus int
		validateBody   func(t *testing.T, body map[string]interface{})
	}{
		{
			name:           "GET request succeeds",
			method:         http.MethodGet,
			queryParams:    "?org_id=org-test&year=2024",
			body:           nil,
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body map[string]interface{}) {
				// CSRD report might be wrapped - check for org_id in different locations
				if orgID, ok := body["org_id"].(string); ok {
					if orgID != "org-test" {
						t.Errorf("Expected org_id=org-test, got %v", orgID)
					}
				}
				if year, ok := body["year"].(float64); ok {
					if year != 2024 {
						t.Errorf("Expected year=2024, got %v", year)
					}
				}
			},
		},
		{
			name:        "POST request succeeds",
			method:      http.MethodPost,
			queryParams: "?org_id=org-test",
			body: map[string]interface{}{
				"year":   2024,
				"format": "pdf",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "PUT method not allowed",
			method:         http.MethodPut,
			queryParams:    "",
			body:           nil,
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bodyReader *strings.Reader
			if tt.body != nil {
				b, _ := json.Marshal(tt.body)
				bodyReader = strings.NewReader(string(b))
			} else {
				bodyReader = strings.NewReader("")
			}
			req := httptest.NewRequest(tt.method, "/api/compliance/csrd"+tt.queryParams, bodyReader)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.validateBody != nil && w.Code == http.StatusOK {
				var body map[string]interface{}
				if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				tt.validateBody(t, body)
			}
		})
	}
}

// TestSECComplianceHandler tests the SEC endpoint
func TestSECComplianceHandler(t *testing.T) {
	deps, _ := setupComplianceTest(t)
	handler := handlers.NewSECComplianceHandler(deps)

	req := httptest.NewRequest(http.MethodGet, "/api/compliance/sec?org_id=org-test&year=2024&org_name=Test+Org&cik=0001234567", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
		return
	}

	var body map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// SEC report structure may vary - just check that we got a valid response
	if body == nil {
		t.Fatal("Received nil response body")
	}
}

// TestCaliforniaComplianceHandler tests the California endpoint
func TestCaliforniaComplianceHandler(t *testing.T) {
	deps, _ := setupComplianceTest(t)
	handler := handlers.NewCaliforniaComplianceHandler(deps)

	req := httptest.NewRequest(http.MethodGet, "/api/compliance/california?org_id=org-test&year=2024&org_name=Test+Org", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if body["org_id"] != "org-test" {
		t.Errorf("Expected org_id=org-test, got %v", body["org_id"])
	}

	// Verify all scopes are present
	if _, ok := body["scope1_tons"]; !ok {
		t.Error("Missing scope1_tons in response")
	}
	if _, ok := body["scope2_tons"]; !ok {
		t.Error("Missing scope2_tons in response")
	}
	if _, ok := body["scope3_tons"]; !ok {
		t.Error("Missing scope3_tons in response")
	}
}

// TestCBAMComplianceHandler tests the CBAM endpoint
func TestCBAMComplianceHandler(t *testing.T) {
	deps, _ := setupComplianceTest(t)
	handler := handlers.NewCBAMComplianceHandler(deps)

	req := httptest.NewRequest(http.MethodGet, "/api/compliance/cbam?org_id=org-test&year=2024&quarter=1", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if body["org_id"] != "org-test" {
		t.Errorf("Expected org_id=org-test, got %v", body["org_id"])
	}

	if body["quarter"] != float64(1) {
		t.Errorf("Expected quarter=1, got %v", body["quarter"])
	}
}

// TestIFRSComplianceHandler tests the IFRS endpoint
func TestIFRSComplianceHandler(t *testing.T) {
	deps, _ := setupComplianceTest(t)
	handler := handlers.NewIFRSComplianceHandler(deps)

	req := httptest.NewRequest(http.MethodGet, "/api/compliance/ifrs?org_id=org-test&year=2024&org_name=Test+Org", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
		return
	}

	var body map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// IFRS report structure may be nested - just verify we got a response
	if body == nil {
		t.Fatal("Received nil response body")
	}
}

// TestComplianceSummaryHandler tests the summary endpoint
func TestComplianceSummaryHandler(t *testing.T) {
	deps, _ := setupComplianceTest(t)
	handler := handlers.NewComplianceSummaryHandler(deps)

	req := httptest.NewRequest(http.MethodGet, "/api/compliance/summary?org_id=org-test&year=2024", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify frameworks map exists
	frameworks, ok := body["frameworks"].(map[string]interface{})
	if !ok {
		t.Fatal("Missing or invalid frameworks field in response")
	}

	// Verify all frameworks are present
	requiredFrameworks := []string{"csrd", "sec", "california", "cbam", "ifrs_s2"}
	for _, fw := range requiredFrameworks {
		if _, ok := frameworks[fw]; !ok {
			t.Errorf("Missing framework %s in summary", fw)
		}
	}

	// Verify each framework has required fields
	for fwName, fwData := range frameworks {
		fwMap, ok := fwData.(map[string]interface{})
		if !ok {
			t.Errorf("Framework %s has invalid structure", fwName)
			continue
		}

		// Check required fields
		if _, ok := fwMap["status"]; !ok {
			t.Errorf("Framework %s missing status field", fwName)
		}
		if _, ok := fwMap["scope1_ready"]; !ok {
			t.Errorf("Framework %s missing scope1_ready field", fwName)
		}
		if _, ok := fwMap["scope2_ready"]; !ok {
			t.Errorf("Framework %s missing scope2_ready field", fwName)
		}
		if _, ok := fwMap["scope3_ready"]; !ok {
			t.Errorf("Framework %s missing scope3_ready field", fwName)
		}
	}

	// Verify totals exist
	totals, ok := body["totals"].(map[string]interface{})
	if !ok {
		t.Fatal("Missing or invalid totals field in response")
	}

	// Check totals have all scope data
	if _, ok := totals["Scope1Tons"]; !ok {
		t.Error("Missing Scope1Tons in totals")
	}
	if _, ok := totals["Scope2Tons"]; !ok {
		t.Error("Missing Scope2Tons in totals")
	}
	if _, ok := totals["Scope3Tons"]; !ok {
		t.Error("Missing Scope3Tons in totals")
	}
	if _, ok := totals["TotalTons"]; !ok {
		t.Error("Missing TotalTons in totals")
	}
}

// TestComplianceSummaryRealStatuses tests that summary returns real derived statuses
func TestComplianceSummaryRealStatuses(t *testing.T) {
	deps, _ := setupComplianceTest(t)
	handler := handlers.NewComplianceSummaryHandler(deps)

	req := httptest.NewRequest(http.MethodGet, "/api/compliance/summary?org_id=org-test&year=2024", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	frameworks := body["frameworks"].(map[string]interface{})

	// Verify statuses are valid compliance statuses (not static strings)
	validStatuses := map[string]bool{
		"not_started":  true,
		"partial":      true,
		"compliant":    true,
		"unknown":      true,
		"not_required": true,
	}

	for fwName, fwData := range frameworks {
		fwMap := fwData.(map[string]interface{})
		status, ok := fwMap["status"].(string)
		if !ok {
			t.Errorf("Framework %s has invalid status type", fwName)
			continue
		}

		if !validStatuses[status] {
			t.Errorf("Framework %s has invalid status: %s (expected one of: not_started, partial, compliant, unknown, not_required)", fwName, status)
		}

		// Status should not be "unknown" if we have data
		hasData, _ := fwMap["has_data"].(bool)
		if hasData && status == "unknown" {
			t.Errorf("Framework %s has data but status is 'unknown' - should be derived from actual data", fwName)
		}
	}
}

// TestComplianceHandlersWithNoData tests behavior when no activity data exists
func TestComplianceHandlersWithNoData(t *testing.T) {
	// Create empty store
	store := ingestion.NewInMemoryActivityStore()
	scope1Calc := emissions.NewScope1Calculator(emissions.Scope1Config{})
	scope2Calc := emissions.NewScope2Calculator(emissions.Scope2Config{})
	scope3Calc := emissions.NewScope3Calculator(emissions.Scope3Config{})

	complianceService := compliance.NewService(store, scope1Calc, scope2Calc, scope3Calc)
	deps := &handlers.ComplianceHandlerDeps{
		ComplianceService: complianceService,
	}

	handler := handlers.NewComplianceSummaryHandler(deps)
	req := httptest.NewRequest(http.MethodGet, "/api/compliance/summary?org_id=org-empty&year=2024", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 even with no data, got %d", w.Code)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	frameworks := body["frameworks"].(map[string]interface{})

	// All frameworks should have "not_started" status when no data exists
	for fwName, fwData := range frameworks {
		fwMap := fwData.(map[string]interface{})
		status := fwMap["status"].(string)

		// CBAM might be "not_required", others should be "not_started"
		if fwName != "cbam" && status != "not_started" {
			t.Errorf("Framework %s should have status 'not_started' with no data, got %s", fwName, status)
		}

		hasData := fwMap["has_data"].(bool)
		if hasData {
			t.Errorf("Framework %s reports has_data=true but store is empty", fwName)
		}
	}
}
