// Package aws provides mocks for testing AWS connectors.
package aws

import (
	"context"
	"fmt"
	"io"
	"strings"
)

// =============================================================================
// Mock S3 Client
// =============================================================================

// MockS3Client mocks AWS S3 GetObject and ListObjectsV2 operations.
type MockS3Client struct {
	// Objects maps S3 keys to file contents
	Objects map[string][]byte

	// ListResults maps prefix to list of keys and sizes
	ListResults map[string][]struct {
		Key  string
		Size int64
	}

	// GetObjectErrors maps keys to errors (simulates failures)
	GetObjectErrors map[string]error

	// ListObjectsErrors simulates ListObjectsV2 failures
	ListObjectsErrors map[string]error

	// CallCounts tracks number of calls (for testing retry logic)
	GetObjectCalls    int
	ListObjectsCalls  int
}

// NewMockS3Client creates a new mock S3 client.
func NewMockS3Client() *MockS3Client {
	return &MockS3Client{
		Objects:           make(map[string][]byte),
		ListResults:       make(map[string][]struct{ Key string; Size int64 }),
		GetObjectErrors:   make(map[string]error),
		ListObjectsErrors: make(map[string]error),
	}
}

// MockGetObjectInput represents GetObject parameters.
type MockGetObjectInput struct {
	Bucket string
	Key    string
}

// MockGetObjectOutput represents GetObject response.
type MockGetObjectOutput struct {
	Body io.ReadCloser
}

// GetObject mocks S3 GetObject operation.
func (m *MockS3Client) GetObject(ctx context.Context, bucket, key string) ([]byte, error) {
	m.GetObjectCalls++

	// Check for simulated error
	if err, exists := m.GetObjectErrors[key]; exists {
		return nil, err
	}

	// Check for object
	if data, exists := m.Objects[key]; exists {
		return data, nil
	}

	return nil, fmt.Errorf("mock: NoSuchKey: The specified key does not exist")
}

// ListObjectsV2 mocks S3 ListObjectsV2 operation.
func (m *MockS3Client) ListObjectsV2(ctx context.Context, bucket, prefix string) ([]struct{ Key string; Size int64 }, error) {
	m.ListObjectsCalls++

	// Check for simulated error
	if err, exists := m.ListObjectsErrors[prefix]; exists {
		return nil, err
	}

	// Return list results for prefix
	if results, exists := m.ListResults[prefix]; exists {
		return results, nil
	}

	// Build list from objects map (matches prefix)
	var results []struct {
		Key  string
		Size int64
	}
	for key, data := range m.Objects {
		if strings.HasPrefix(key, prefix) {
			results = append(results, struct{ Key string; Size int64 }{
				Key:  key,
				Size: int64(len(data)),
			})
		}
	}

	return results, nil
}

// AddObject adds an object to the mock S3 client.
func (m *MockS3Client) AddObject(key string, data []byte) {
	m.Objects[key] = data
}

// AddListResult adds a list operation result.
func (m *MockS3Client) AddListResult(prefix string, results []struct{ Key string; Size int64 }) {
	m.ListResults[prefix] = results
}

// SetGetObjectError sets an error for a specific key.
func (m *MockS3Client) SetGetObjectError(key string, err error) {
	m.GetObjectErrors[key] = err
}

// SetListObjectsError sets an error for a specific prefix.
func (m *MockS3Client) SetListObjectsError(prefix string, err error) {
	m.ListObjectsErrors[prefix] = err
}

// Reset clears all call counts.
func (m *MockS3Client) Reset() {
	m.GetObjectCalls = 0
	m.ListObjectsCalls = 0
}

// =============================================================================
// Mock Carbon Footprint API
// =============================================================================

// MockCarbonAPI mocks AWS Carbon Footprint API responses.
type MockCarbonAPI struct {
	// Response is the API response to return
	Response *CarbonFootprintSummary

	// Error is the error to return
	Error error

	// CallCount tracks number of calls
	CallCount int

	// FailFirstN fails the first N calls, then succeeds
	FailFirstN int
	callsMade  int
}

// NewMockCarbonAPI creates a new mock Carbon API.
func NewMockCarbonAPI() *MockCarbonAPI {
	return &MockCarbonAPI{
		Response: &CarbonFootprintSummary{
			TotalCO2e: 100.5,
			Unit:      "metric tons",
			EmissionsByService: []ServiceEmission{
				{
					ServiceCode: "AmazonEC2",
					ServiceName: "EC2",
					CO2e:        50.0,
					Region:      "us-east-1",
					Scope:       "Scope2",
				},
				{
					ServiceCode: "AmazonRDS",
					ServiceName: "RDS",
					CO2e:        25.5,
					Region:      "us-east-1",
					Scope:       "Scope2",
				},
			},
			Period: Period{},
		},
		Error: nil,
	}
}

// GetCarbonFootprint returns the mocked response.
func (m *MockCarbonAPI) GetCarbonFootprint(ctx context.Context) (*CarbonFootprintSummary, error) {
	m.CallCount++
	m.callsMade++

	// Simulate failures for first N calls
	if m.FailFirstN > 0 && m.callsMade <= m.FailFirstN {
		return nil, fmt.Errorf("mock: temporary API error")
	}

	if m.Error != nil {
		return nil, m.Error
	}

	return m.Response, nil
}

// SetResponse sets the response to return.
func (m *MockCarbonAPI) SetResponse(resp *CarbonFootprintSummary) {
	m.Response = resp
}

// SetError sets the error to return.
func (m *MockCarbonAPI) SetError(err error) {
	m.Error = err
}

// Reset resets call counts.
func (m *MockCarbonAPI) Reset() {
	m.CallCount = 0
	m.callsMade = 0
}

// =============================================================================
// Sample Test Data
// =============================================================================

// SampleManifestJSON returns a sample AWS CUR manifest.json for testing.
func SampleManifestJSON() []byte {
	manifest := `{
  "assemblyId": "12345abc-1234-5678-abcd-123456789012",
  "invoiceId": "",
  "billingPeriod": {
    "start": "2024-01-01T00:00:00.000Z",
    "end": "2024-02-01T00:00:00.000Z"
  },
  "files": [
    {
      "key": "cur/organization-id/2024/01/31/123456789012-cur-001.csv.gz",
      "size": 1048576,
      "reportKey": "2024-01-31/123456789012-cur"
    },
    {
      "key": "cur/organization-id/2024/01/31/123456789012-cur-002.csv.gz",
      "size": 2097152,
      "reportKey": "2024-01-31/123456789012-cur"
    }
  ],
  "charset": "UTF-8",
  "contentType": "text/csv",
  "reportKeys": [
    "2024-01-31/123456789012-cur"
  ],
  "reportName": "cur-example-report",
  "bucket": "example-cur-bucket",
  "columnHeaders": [
    "lineItem/UsageAccountId",
    "lineItem/UsageStartDate",
    "lineItem/UsageEndDate",
    "product/servicecode",
    "product/productname",
    "product/region",
    "lineItem/UsageAmount",
    "lineItem/UsageUnit",
    "lineItem/BlendedCost"
  ],
  "isTruncated": false
}`

	return []byte(manifest)
}

// SampleCURCSVRow returns a sample CUR CSV row for testing.
func SampleCURCSVRow() string {
	return `"123456789012","2024-01-01T00:00:00Z","2024-01-02T00:00:00Z","AmazonEC2","EC2 - Other","us-east-1","730","Hrs","100.00","100.00","i3.large"`
}

// SampleCarbonFootprintResponse returns a sample Carbon Footprint API response.
func SampleCarbonFootprintResponse() *CarbonFootprintSummary {
	return &CarbonFootprintSummary{
		TotalCO2e: 123.45,
		Unit:      "metric tons",
		EmissionsByService: []ServiceEmission{
			{
				ServiceCode: "AmazonEC2",
				ServiceName: "Amazon Elastic Compute Cloud",
				CO2e:        75.50,
				Region:      "us-east-1",
				Scope:       "Scope2",
			},
			{
				ServiceCode: "AmazonS3",
				ServiceName: "Amazon Simple Storage Service",
				CO2e:        25.25,
				Region:      "us-west-2",
				Scope:       "Scope2",
			},
			{
				ServiceCode: "AmazonRDS",
				ServiceName: "Amazon Relational Database Service",
				CO2e:        22.70,
				Region:      "eu-west-1",
				Scope:       "Scope2",
			},
		},
		Period: Period{
			Start: Period{}.Start,
			End:   Period{}.End,
		},
	}
}
