package billing

import (
	"context"
	"sync"
	"time"
)

// UsageLimits defines the limits for each subscription plan.
type UsageLimits struct {
	// MaxEmissionRecords is the maximum number of emission records allowed.
	MaxEmissionRecords int
	// MaxUsers is the maximum number of users per tenant.
	MaxUsers int
	// MaxAPICallsPerMonth is the maximum API calls allowed per month.
	MaxAPICallsPerMonth int
	// MaxDataSourcesConnected is the maximum number of data source integrations.
	MaxDataSourcesConnected int
	// AIQueriesPerMonth is the number of AI chat queries allowed per month.
	AIQueriesPerMonth int
	// ReportsPerMonth is the number of compliance reports that can be generated.
	ReportsPerMonth int
	// DataRetentionMonths is how long historical data is retained.
	DataRetentionMonths int
}

// PlanLimits defines usage limits for each plan tier.
var PlanLimits = map[string]UsageLimits{
	"free": {
		MaxEmissionRecords:      100,
		MaxUsers:                1,
		MaxAPICallsPerMonth:     1000,
		MaxDataSourcesConnected: 0,
		AIQueriesPerMonth:       0,
		ReportsPerMonth:         1,
		DataRetentionMonths:     3,
	},
	"basic": {
		MaxEmissionRecords:      1000,
		MaxUsers:                1,
		MaxAPICallsPerMonth:     10000,
		MaxDataSourcesConnected: 1,
		AIQueriesPerMonth:       10,
		ReportsPerMonth:         5,
		DataRetentionMonths:     12,
	},
	"pro": {
		MaxEmissionRecords:      -1, // Unlimited
		MaxUsers:                -1, // Unlimited
		MaxAPICallsPerMonth:     100000,
		MaxDataSourcesConnected: 5,
		AIQueriesPerMonth:       100,
		ReportsPerMonth:         -1, // Unlimited
		DataRetentionMonths:     36,
	},
	"enterprise": {
		MaxEmissionRecords:      -1, // Unlimited
		MaxUsers:                -1, // Unlimited
		MaxAPICallsPerMonth:     -1, // Unlimited
		MaxDataSourcesConnected: -1, // Unlimited
		AIQueriesPerMonth:       -1, // Unlimited
		ReportsPerMonth:         -1, // Unlimited
		DataRetentionMonths:     -1, // Unlimited
	},
}

// UsageType represents a type of usage to track.
type UsageType string

const (
	UsageAPICall         UsageType = "api_call"
	UsageAIQuery         UsageType = "ai_query"
	UsageReportGenerated UsageType = "report"
	UsageEmissionRecord  UsageType = "emission_record"
)

// UsageRecord tracks usage of a specific type for a tenant.
type UsageRecord struct {
	TenantID  string
	Type      UsageType
	Count     int
	Period    string // Format: "2024-01" for monthly tracking
	UpdatedAt time.Time
}

// UsageTracker tracks and enforces usage limits.
type UsageTracker struct {
	store   UsageStore
	billing *Service
	cache   map[string]map[UsageType]*UsageRecord // tenantID -> type -> record
}

// UsageStore persists usage records.
type UsageStore interface {
	GetUsage(ctx context.Context, tenantID string, usageType UsageType, period string) (*UsageRecord, error)
	IncrementUsage(ctx context.Context, tenantID string, usageType UsageType, period string, delta int) (*UsageRecord, error)
	GetTotalCount(ctx context.Context, tenantID string, usageType UsageType) (int, error)
}

// NewUsageTracker creates a new usage tracker.
func NewUsageTracker(store UsageStore, billing *Service) *UsageTracker {
	return &UsageTracker{
		store:   store,
		billing: billing,
		cache:   make(map[string]map[UsageType]*UsageRecord),
	}
}

// CheckLimit checks if a tenant has exceeded a specific usage limit.
// Returns true if the usage is allowed, false if limit exceeded.
func (t *UsageTracker) CheckLimit(ctx context.Context, tenantID string, usageType UsageType) (bool, error) {
	// Get subscription plan
	sub, err := t.billing.GetSubscription(ctx, tenantID)
	if err != nil {
		return true, err // Fail open
	}

	plan := "free"
	if sub != nil && sub.IsActive() {
		plan = sub.Plan
	}

	limits := PlanLimits[plan]
	if limits.MaxEmissionRecords == 0 && limits.MaxUsers == 0 {
		limits = PlanLimits["free"]
	}

	// Get current usage
	period := currentPeriod()
	var currentCount int
	var limit int

	switch usageType {
	case UsageAPICall:
		record, err := t.store.GetUsage(ctx, tenantID, usageType, period)
		if err != nil {
			return true, err
		}
		if record != nil {
			currentCount = record.Count
		}
		limit = limits.MaxAPICallsPerMonth

	case UsageAIQuery:
		record, err := t.store.GetUsage(ctx, tenantID, usageType, period)
		if err != nil {
			return true, err
		}
		if record != nil {
			currentCount = record.Count
		}
		limit = limits.AIQueriesPerMonth

	case UsageReportGenerated:
		record, err := t.store.GetUsage(ctx, tenantID, usageType, period)
		if err != nil {
			return true, err
		}
		if record != nil {
			currentCount = record.Count
		}
		limit = limits.ReportsPerMonth

	case UsageEmissionRecord:
		totalCount, err := t.store.GetTotalCount(ctx, tenantID, usageType)
		if err != nil {
			return true, err
		}
		currentCount = totalCount
		limit = limits.MaxEmissionRecords
	}

	// -1 means unlimited
	if limit < 0 {
		return true, nil
	}

	return currentCount < limit, nil
}

// RecordUsage records usage of a specific type.
func (t *UsageTracker) RecordUsage(ctx context.Context, tenantID string, usageType UsageType, count int) error {
	period := currentPeriod()
	_, err := t.store.IncrementUsage(ctx, tenantID, usageType, period, count)
	return err
}

// GetUsageSummary returns usage summary for a tenant.
func (t *UsageTracker) GetUsageSummary(ctx context.Context, tenantID string) (map[string]interface{}, error) {
	// Get subscription plan
	sub, err := t.billing.GetSubscription(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	plan := "free"
	if sub != nil && sub.IsActive() {
		plan = sub.Plan
	}

	limits := PlanLimits[plan]
	period := currentPeriod()

	// Get current usage for each type
	apiCalls, _ := t.store.GetUsage(ctx, tenantID, UsageAPICall, period)
	aiQueries, _ := t.store.GetUsage(ctx, tenantID, UsageAIQuery, period)
	reports, _ := t.store.GetUsage(ctx, tenantID, UsageReportGenerated, period)
	totalRecords, _ := t.store.GetTotalCount(ctx, tenantID, UsageEmissionRecord)

	return map[string]interface{}{
		"plan":   plan,
		"period": period,
		"usage": map[string]interface{}{
			"api_calls": map[string]interface{}{
				"current": getCount(apiCalls),
				"limit":   limits.MaxAPICallsPerMonth,
			},
			"ai_queries": map[string]interface{}{
				"current": getCount(aiQueries),
				"limit":   limits.AIQueriesPerMonth,
			},
			"reports": map[string]interface{}{
				"current": getCount(reports),
				"limit":   limits.ReportsPerMonth,
			},
			"emission_records": map[string]interface{}{
				"current": totalRecords,
				"limit":   limits.MaxEmissionRecords,
			},
		},
		"limits": limits,
	}, nil
}

// currentPeriod returns the current billing period (month).
func currentPeriod() string {
	return time.Now().Format("2006-01")
}

func getCount(record *UsageRecord) int {
	if record == nil {
		return 0
	}
	return record.Count
}

// InMemoryUsageStore provides an in-memory usage store for development.
type InMemoryUsageStore struct {
	mu      sync.RWMutex
	records map[string]map[UsageType]map[string]*UsageRecord // tenantID -> type -> period -> record
	totals  map[string]map[UsageType]int                     // tenantID -> type -> total count
}

// NewInMemoryUsageStore creates an in-memory usage store.
func NewInMemoryUsageStore() *InMemoryUsageStore {
	return &InMemoryUsageStore{
		records: make(map[string]map[UsageType]map[string]*UsageRecord),
		totals:  make(map[string]map[UsageType]int),
	}
}

// GetUsage returns usage for a specific tenant, type, and period.
func (s *InMemoryUsageStore) GetUsage(ctx context.Context, tenantID string, usageType UsageType, period string) (*UsageRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if tenantRecords, ok := s.records[tenantID]; ok {
		if typeRecords, ok := tenantRecords[usageType]; ok {
			if record, ok := typeRecords[period]; ok {
				return record, nil
			}
		}
	}
	return nil, nil
}

// IncrementUsage increments usage count.
func (s *InMemoryUsageStore) IncrementUsage(ctx context.Context, tenantID string, usageType UsageType, period string, delta int) (*UsageRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Initialize maps if needed
	if _, ok := s.records[tenantID]; !ok {
		s.records[tenantID] = make(map[UsageType]map[string]*UsageRecord)
	}
	if _, ok := s.records[tenantID][usageType]; !ok {
		s.records[tenantID][usageType] = make(map[string]*UsageRecord)
	}
	if _, ok := s.totals[tenantID]; !ok {
		s.totals[tenantID] = make(map[UsageType]int)
	}

	// Get or create record
	record, ok := s.records[tenantID][usageType][period]
	if !ok {
		record = &UsageRecord{
			TenantID:  tenantID,
			Type:      usageType,
			Period:    period,
			Count:     0,
			UpdatedAt: time.Now(),
		}
		s.records[tenantID][usageType][period] = record
	}

	record.Count += delta
	record.UpdatedAt = time.Now()

	// Update totals for non-periodic usage types (like emission records)
	if usageType == UsageEmissionRecord {
		s.totals[tenantID][usageType] += delta
	}

	return record, nil
}

// GetTotalCount returns total count for a usage type (across all periods).
func (s *InMemoryUsageStore) GetTotalCount(ctx context.Context, tenantID string, usageType UsageType) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if tenantTotals, ok := s.totals[tenantID]; ok {
		return tenantTotals[usageType], nil
	}
	return 0, nil
}
