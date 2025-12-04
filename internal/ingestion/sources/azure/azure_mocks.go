// Package azure provides mocks for testing Azure connectors.
package azure

import (
	"context"
	"fmt"
	"time"
)

// =============================================================================
// Mock Token Provider
// =============================================================================

// MockTokenProvider mocks Azure OAuth token generation.
type MockTokenProvider struct {
	Token             string
	ExpiresAt         time.Time
	GetTokenCalls     int
	RefreshTokenCalls int
	Error             error
	FailFirstN        int
	callCount         int
}

// NewMockTokenProvider creates a mock token provider.
func NewMockTokenProvider() *MockTokenProvider {
	return &MockTokenProvider{
		Token:     "mock_access_token_xyz",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
}

// GetToken returns a mocked token.
func (mtp *MockTokenProvider) GetToken(ctx context.Context) (string, error) {
	mtp.GetTokenCalls++
	mtp.callCount++

	if mtp.FailFirstN > 0 && mtp.callCount <= mtp.FailFirstN {
		return "", fmt.Errorf("mock: temporary auth failure")
	}

	if mtp.Error != nil {
		return "", mtp.Error
	}

	return mtp.Token, nil
}

// SetToken sets the token to return.
func (mtp *MockTokenProvider) SetToken(token string) {
	mtp.Token = token
}

// SetError sets an error to return.
func (mtp *MockTokenProvider) SetError(err error) {
	mtp.Error = err
}

// Reset resets call counts.
func (mtp *MockTokenProvider) Reset() {
	mtp.GetTokenCalls = 0
	mtp.RefreshTokenCalls = 0
	mtp.callCount = 0
}

// =============================================================================
// Mock Emissions API
// =============================================================================

// MockEmissionsAPI mocks the Emissions Impact Dashboard API.
type MockEmissionsAPI struct {
	Records           [][]EmissionRecord // Paginated results
	NextLinks         []string            // NextLink for each page
	CallCount         int
	CurrentPageIndex  int
	GetPageErrors     map[int]error  // Errors for specific pages
	FailFirstN        int            // Fail first N calls
	callCount         int
}

// NewMockEmissionsAPI creates a mock emissions API.
func NewMockEmissionsAPI() *MockEmissionsAPI {
	return &MockEmissionsAPI{
		Records:       make([][]EmissionRecord, 0),
		NextLinks:     make([]string, 0),
		GetPageErrors: make(map[int]error),
	}
}

// AddPage adds a page of emissions records.
func (mea *MockEmissionsAPI) AddPage(records []EmissionRecord, nextLink string) {
	mea.Records = append(mea.Records, records)
	mea.NextLinks = append(mea.NextLinks, nextLink)
}

// GetPage returns the next page of records.
func (mea *MockEmissionsAPI) GetPage(ctx context.Context, skipToken string) ([]EmissionRecord, string, error) {
	mea.CallCount++
	mea.callCount++

	if mea.FailFirstN > 0 && mea.callCount <= mea.FailFirstN {
		return nil, "", fmt.Errorf("mock: temporary API error")
	}

	// Check for page-specific error
	if err, exists := mea.GetPageErrors[mea.CurrentPageIndex]; exists {
		return nil, "", err
	}

	// Return current page
	if mea.CurrentPageIndex < len(mea.Records) {
		records := mea.Records[mea.CurrentPageIndex]
		nextLink := ""
		if mea.CurrentPageIndex < len(mea.NextLinks) {
			nextLink = mea.NextLinks[mea.CurrentPageIndex]
		}

		mea.CurrentPageIndex++
		return records, nextLink, nil
	}

	return nil, "", nil
}

// Reset resets the API state.
func (mea *MockEmissionsAPI) Reset() {
	mea.CallCount = 0
	mea.CurrentPageIndex = 0
	mea.callCount = 0
}

// =============================================================================
// Sample Test Data
// =============================================================================

// SampleEmissionRecord returns a sample Azure emission record.
func SampleEmissionRecord() EmissionRecord {
	return EmissionRecord{
		ID:                     "emission-001",
		SubscriptionID:         "sub-123",
		ResourceGroup:          "rg-prod",
		ResourceType:           "Microsoft.Compute/virtualMachines",
		ResourceName:           "vm-prod-01",
		Region:                 "eastus",
		ServiceName:            "Virtual Machines",
		MeterCategory:          "Compute",
		MeterSubcategory:       "VM",
		Date:                   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Scope1CO2e:             0.0,
		Scope2CO2e:             12.5,  // Grid electricity
		Scope3CO2e:             0.0,
		TotalCO2e:              12.5,
		EnergyConsumptionKWh:   50.0,
		CarbonIntensity:        250.0,  // gCO2e/kWh
		RenewableEnergyPercent: 25.0,
		Currency:               "USD",
		Cost:                   15.00,
	}
}

// SampleEmissionRecords returns multiple sample records.
func SampleEmissionRecords(count int) []EmissionRecord {
	records := make([]EmissionRecord, count)
	for i := 0; i < count; i++ {
		rec := SampleEmissionRecord()
		rec.ID = fmt.Sprintf("emission-%03d", i)
		rec.ResourceName = fmt.Sprintf("resource-%d", i)
		rec.TotalCO2e = float64(i+1) * 10.0
		records[i] = rec
	}
	return records
}

// SampleCostRecords returns sample cost management records.
func SampleCostRecords() []struct {
	Date time.Time
	Cost float64
} {
	return []struct {
		Date time.Time
		Cost float64
	}{
		{time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), 1500.00},
		{time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC), 1650.00},
		{time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), 1450.00},
		{time.Date(2024, 1, 22, 0, 0, 0, 0, time.UTC), 1700.00},
	}
}
