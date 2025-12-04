// Package gcp provides mocks for testing GCP connectors.
package gcp

import (
	"context"
	"fmt"
)

// =============================================================================
// Mock Service Account Auth
// =============================================================================

// MockServiceAccountAuth mocks GCP service account authentication.
type MockServiceAccountAuth struct {
	ProjectID       string
	GetCallCount    int
	Error           error
	ValidateOnlyKey bool
}

// NewMockServiceAccountAuth creates a mock service account authenticator.
func NewMockServiceAccountAuth(projectID string) *MockServiceAccountAuth {
	return &MockServiceAccountAuth{
		ProjectID: projectID,
	}
}

// GetProjectID returns the mocked project ID.
func (msaa *MockServiceAccountAuth) GetProjectID() string {
	return msaa.ProjectID
}

// SetError sets an error to return.
func (msaa *MockServiceAccountAuth) SetError(err error) {
	msaa.Error = err
}

// Reset resets call counts.
func (msaa *MockServiceAccountAuth) Reset() {
	msaa.GetCallCount = 0
}

// =============================================================================
// Mock BigQuery Results
// =============================================================================

// MockBigQueryResults mocks BigQuery query results with pagination.
type MockBigQueryResults struct {
	Records      []CarbonRecord
	GetCallCount int
	Error        error
	FailFirstN   int
	callCount    int
}

// NewMockBigQueryResults creates mock BigQuery results.
func NewMockBigQueryResults() *MockBigQueryResults {
	return &MockBigQueryResults{
		Records: make([]CarbonRecord, 0),
	}
}

// AddRecord adds a mock carbon record.
func (mbr *MockBigQueryResults) AddRecord(rec CarbonRecord) {
	mbr.Records = append(mbr.Records, rec)
}

// GetRecords returns all mocked records.
func (mbr *MockBigQueryResults) GetRecords(ctx context.Context) ([]CarbonRecord, error) {
	mbr.GetCallCount++
	mbr.callCount++

	if mbr.FailFirstN > 0 && mbr.callCount <= mbr.FailFirstN {
		return nil, fmt.Errorf("mock: temporary query error")
	}

	if mbr.Error != nil {
		return nil, mbr.Error
	}

	return mbr.Records, nil
}

// SetError sets an error to return.
func (mbr *MockBigQueryResults) SetError(err error) {
	mbr.Error = err
}

// Reset resets state.
func (mbr *MockBigQueryResults) Reset() {
	mbr.GetCallCount = 0
	mbr.callCount = 0
}

// =============================================================================
// Sample Test Data
// =============================================================================

// SampleCarbonRecord returns a sample GCP carbon record.
func SampleCarbonRecord() CarbonRecord {
	return CarbonRecord{
		BillingAccountID: "012345-678901-ABCDEF",
		Project: Project{
			ID:     "my-project-prod",
			Number: "123456789012",
			Name:   "My Project (Production)",
		},
		Service: Service{
			ID:          "6F81-5844-456A",
			Description: "Compute Engine",
		},
		Location: Location{
			Location: "us-central1",
			Country:  "United States",
			Region:   "Iowa",
		},
		UsageMonth:           "202401",
		CarbonFootprintKgCO2: 125.50,
		CarbonModelVersion:   "3.0",
		Scope1Emissions:      0.0,
		Scope2Emissions:      125.50,
		Scope3Emissions:      0.0,
		ElectricityKWh:       500.0,
		CFEScore:             45.0,
	}
}

// SampleCarbonRecords returns multiple sample records.
func SampleCarbonRecords(count int) []CarbonRecord {
	records := make([]CarbonRecord, count)
	for i := 0; i < count; i++ {
		rec := SampleCarbonRecord()
		rec.Project.ID = fmt.Sprintf("project-%d", i)
		rec.CarbonFootprintKgCO2 = float64(i+1) * 50.0
		rec.UsageMonth = fmt.Sprintf("202401%02d", (i%28)+1)
		records[i] = rec
	}
	return records
}

// SampleServiceAccountKey returns a sample GCP service account JSON (fake).
func SampleServiceAccountKey(projectID string) string {
	return fmt.Sprintf(`{
  "type": "service_account",
  "project_id": "%s",
  "private_key_id": "key123abc",
  "private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC7W...\n-----END PRIVATE KEY-----\n",
  "client_email": "service-account@%s.iam.gserviceaccount.com",
  "client_id": "123456789012345678901",
  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
  "token_uri": "https://oauth2.googleapis.com/token",
  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
  "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/service-account%%40%s.iam.gserviceaccount.com"
}`, projectID, projectID, projectID)
}
