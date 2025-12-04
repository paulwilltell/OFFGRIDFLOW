// Package allocation provides the emission allocation service.
package allocation

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/example/offgridflow/internal/emissions"
)

// =============================================================================
// Service Configuration
// =============================================================================

// ServiceConfig configures the allocation service.
type ServiceConfig struct {
	// Rules is the initial set of allocation rules.
	Rules []Rule

	// Logger for allocation operations.
	Logger *slog.Logger

	// StrictMode fails on any allocation error rather than skipping.
	StrictMode bool

	// ValidateRules validates rules on service creation.
	ValidateRules bool

	// EnableMetrics collects allocation metrics.
	EnableMetrics bool
}

// DefaultServiceConfig returns sensible defaults.
func DefaultServiceConfig() ServiceConfig {
	return ServiceConfig{
		ValidateRules: true,
		EnableMetrics: true,
	}
}

// =============================================================================
// Sentinel Errors
// =============================================================================

var (
	// ErrNilService is returned when methods are called on a nil *Service.
	ErrNilService = errors.New("allocation: nil Service receiver")

	// ErrNoApplicableRule is returned when no rule applies to an emission.
	ErrNoApplicableRule = errors.New("allocation: no applicable rule found")

	// ErrAllocationFailed is returned when allocation calculation fails.
	ErrAllocationFailed = errors.New("allocation: calculation failed")
)

// =============================================================================
// Allocation Result Types
// =============================================================================

// AllocationResult represents emissions after allocation.
type AllocationResult struct {
	// OriginalID is the ID of the source emission record.
	OriginalID string `json:"original_id"`

	// AllocatedEmissions contains the distributed emission records.
	AllocatedEmissions []AllocatedEmission `json:"allocated_emissions"`

	// RuleID is the rule that was applied.
	RuleID string `json:"rule_id"`

	// Dimension is the allocation dimension used.
	Dimension Dimension `json:"dimension"`

	// AllocatedAt is when allocation was performed.
	AllocatedAt time.Time `json:"allocated_at"`
}

// AllocatedEmission represents a single allocated emission portion.
type AllocatedEmission struct {
	// TargetID identifies the allocation target.
	TargetID string `json:"target_id"`

	// TargetName is the human-readable target name.
	TargetName string `json:"target_name,omitempty"`

	// Percentage is the allocation percentage applied.
	Percentage float64 `json:"percentage"`

	// EmissionsKgCO2e is the allocated emission amount.
	EmissionsKgCO2e float64 `json:"emissions_kg_co2e"`

	// SourceRecord contains the original emission data.
	SourceRecord emissions.EmissionRecord `json:"source_record,omitempty"`
}

// BatchAllocationResult summarizes batch allocation results.
type BatchAllocationResult struct {
	// Results contains individual allocation results.
	Results []AllocationResult `json:"results"`

	// TotalInputEmissions is the sum before allocation.
	TotalInputEmissions float64 `json:"total_input_emissions_kg_co2e"`

	// TotalOutputEmissions is the sum after allocation (should equal input).
	TotalOutputEmissions float64 `json:"total_output_emissions_kg_co2e"`

	// SuccessCount is how many records were successfully allocated.
	SuccessCount int `json:"success_count"`

	// ErrorCount is how many records failed allocation.
	ErrorCount int `json:"error_count"`

	// Errors contains error details for failed allocations.
	Errors []AllocationError `json:"errors,omitempty"`

	// ProcessedAt is when the batch was processed.
	ProcessedAt time.Time `json:"processed_at"`
}

// AllocationError records a single allocation failure.
type AllocationError struct {
	// RecordID is the emission record that failed.
	RecordID string `json:"record_id"`

	// Error describes what went wrong.
	Error string `json:"error"`
}

// =============================================================================
// Allocation Metrics
// =============================================================================

// ServiceMetrics tracks allocation service performance.
type ServiceMetrics struct {
	// TotalAllocations is the count of allocation operations.
	TotalAllocations int64 `json:"total_allocations"`

	// SuccessfulAllocations is the count of successful allocations.
	SuccessfulAllocations int64 `json:"successful_allocations"`

	// FailedAllocations is the count of failed allocations.
	FailedAllocations int64 `json:"failed_allocations"`

	// ByDimension tracks allocations per dimension.
	ByDimension map[Dimension]int64 `json:"by_dimension"`

	// ByRule tracks allocations per rule.
	ByRule map[string]int64 `json:"by_rule"`

	// LastAllocationAt is when the last allocation occurred.
	LastAllocationAt time.Time `json:"last_allocation_at"`

	mu sync.RWMutex
}

// NewServiceMetrics creates a new metrics instance.
func NewServiceMetrics() *ServiceMetrics {
	return &ServiceMetrics{
		ByDimension: make(map[Dimension]int64),
		ByRule:      make(map[string]int64),
	}
}

// RecordAllocation records an allocation operation.
func (m *ServiceMetrics) RecordAllocation(dim Dimension, ruleID string, success bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalAllocations++
	if success {
		m.SuccessfulAllocations++
	} else {
		m.FailedAllocations++
	}

	m.ByDimension[dim]++
	m.ByRule[ruleID]++
	m.LastAllocationAt = time.Now()
}

// Clone creates a copy of the metrics for reporting.
func (m *ServiceMetrics) Clone() *ServiceMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Create new instance without copying the mutex
	// The returned value will have a zero-value mutex which is fine for a snapshot
	clone := &ServiceMetrics{
		TotalAllocations:      m.TotalAllocations,
		SuccessfulAllocations: m.SuccessfulAllocations,
		FailedAllocations:     m.FailedAllocations,
		ByDimension:           make(map[Dimension]int64, len(m.ByDimension)),
		ByRule:                make(map[string]int64, len(m.ByRule)),
		LastAllocationAt:      m.LastAllocationAt,
		// mu is intentionally not copied - zero value is appropriate
	}
	for k, v := range m.ByDimension {
		clone.ByDimension[k] = v
	}
	for k, v := range m.ByRule {
		clone.ByRule[k] = v
	}

	return clone
}

// =============================================================================
// Service Implementation
// =============================================================================

// Service applies allocation rules to emissions.
//
// The service manages allocation rules and executes them against emission
// records to distribute emissions across organizational dimensions.
//
// Example usage:
//
//	svc, err := allocation.NewService(allocation.ServiceConfig{
//	    Rules: rules,
//	    Logger: logger,
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	results, err := svc.AllocateBatch(ctx, emissionRecords, DimensionBusinessUnit)
type Service struct {
	ruleSet    *RuleSet
	logger     *slog.Logger
	config     ServiceConfig
	metrics    *ServiceMetrics
	driverData map[string]map[string]float64 // dimension -> target -> driver value
	mu         sync.RWMutex
}

// NewService constructs a Service with the provided configuration.
func NewService(cfg ServiceConfig) (*Service, error) {
	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}

	ruleSet := NewRuleSet()

	for i, r := range cfg.Rules {
		if cfg.ValidateRules {
			if err := r.Validate(); err != nil {
				return nil, fmt.Errorf("invalid rule at index %d (id=%q): %w", i, r.ID, err)
			}
		}
		if err := ruleSet.Add(r); err != nil {
			return nil, fmt.Errorf("failed to add rule %q: %w", r.ID, err)
		}
	}

	var metrics *ServiceMetrics
	if cfg.EnableMetrics {
		metrics = NewServiceMetrics()
	}

	return &Service{
		ruleSet:    ruleSet,
		logger:     logger,
		config:     cfg,
		metrics:    metrics,
		driverData: make(map[string]map[string]float64),
	}, nil
}

// AddRule adds a new allocation rule.
func (s *Service) AddRule(rule Rule) error {
	if s == nil {
		return ErrNilService
	}

	return s.ruleSet.Add(rule)
}

// RemoveRule removes an allocation rule.
func (s *Service) RemoveRule(ruleID string) {
	if s == nil {
		return
	}

	s.ruleSet.Remove(ruleID)
}

// GetRule retrieves a rule by ID.
func (s *Service) GetRule(ruleID string) (Rule, bool) {
	if s == nil {
		return Rule{}, false
	}

	return s.ruleSet.Get(ruleID)
}

// ListRules returns all rules in priority order.
func (s *Service) ListRules() []Rule {
	if s == nil {
		return nil
	}

	return s.ruleSet.All()
}

// SetDriverData sets driver data for driver-based allocations.
// Driver data maps targets to their driver values (e.g., revenue, headcount).
func (s *Service) SetDriverData(dimension Dimension, targetID string, value float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	dimKey := string(dimension)
	if s.driverData[dimKey] == nil {
		s.driverData[dimKey] = make(map[string]float64)
	}
	s.driverData[dimKey][targetID] = value
}

// Allocate allocates a single emission record using the specified dimension.
func (s *Service) Allocate(
	ctx context.Context,
	record emissions.EmissionRecord,
	dimension Dimension,
) (AllocationResult, error) {
	if s == nil {
		return AllocationResult{}, ErrNilService
	}

	if err := ctx.Err(); err != nil {
		return AllocationResult{}, err
	}

	// Find applicable rule
	rule, err := s.findApplicableRule(record, dimension)
	if err != nil {
		if s.metrics != nil {
			s.metrics.RecordAllocation(dimension, "", false)
		}
		return AllocationResult{}, err
	}

	// Perform allocation based on method
	result, err := s.executeAllocation(record, rule)
	if err != nil {
		if s.metrics != nil {
			s.metrics.RecordAllocation(dimension, rule.ID, false)
		}
		return AllocationResult{}, fmt.Errorf("execute allocation: %w", err)
	}

	if s.metrics != nil {
		s.metrics.RecordAllocation(dimension, rule.ID, true)
	}

	s.logger.Debug("allocated emission",
		"record_id", record.ID,
		"rule_id", rule.ID,
		"dimension", dimension,
		"emissions_kg_co2e", record.EmissionsKgCO2e,
		"targets", len(result.AllocatedEmissions),
	)

	return result, nil
}

// AllocateBatch allocates multiple emission records.
func (s *Service) AllocateBatch(
	ctx context.Context,
	records []emissions.EmissionRecord,
	dimension Dimension,
) (BatchAllocationResult, error) {
	if s == nil {
		return BatchAllocationResult{}, ErrNilService
	}

	if err := ctx.Err(); err != nil {
		return BatchAllocationResult{}, err
	}

	result := BatchAllocationResult{
		Results:     make([]AllocationResult, 0, len(records)),
		Errors:      make([]AllocationError, 0),
		ProcessedAt: time.Now().UTC(),
	}

	for _, record := range records {
		result.TotalInputEmissions += record.EmissionsKgCO2e

		alloc, err := s.Allocate(ctx, record, dimension)
		if err != nil {
			result.ErrorCount++
			result.Errors = append(result.Errors, AllocationError{
				RecordID: record.ID,
				Error:    err.Error(),
			})

			if s.config.StrictMode {
				return result, fmt.Errorf("strict mode: allocation failed for %s: %w", record.ID, err)
			}
			continue
		}

		result.Results = append(result.Results, alloc)
		result.SuccessCount++

		for _, ae := range alloc.AllocatedEmissions {
			result.TotalOutputEmissions += ae.EmissionsKgCO2e
		}
	}

	s.logger.Info("batch allocation complete",
		"dimension", dimension,
		"input_count", len(records),
		"success_count", result.SuccessCount,
		"error_count", result.ErrorCount,
		"total_emissions_kg_co2e", result.TotalInputEmissions,
	)

	return result, nil
}

// findApplicableRule finds the best matching rule for the record.
func (s *Service) findApplicableRule(record emissions.EmissionRecord, dimension Dimension) (Rule, error) {
	rules := s.ruleSet.GetByDimension(dimension)

	// Sort by priority (should already be sorted, but ensure)
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Priority > rules[j].Priority
	})

	for _, rule := range rules {
		if s.ruleMatchesRecord(rule, record) {
			return rule, nil
		}
	}

	return Rule{}, fmt.Errorf(
		"no rule for dimension %s matching record %s: %w",
		dimension, record.ID, ErrNoApplicableRule,
	)
}

// ruleMatchesRecord checks if a rule's filters match the record.
func (s *Service) ruleMatchesRecord(rule Rule, record emissions.EmissionRecord) bool {
	for _, filter := range rule.Filters {
		if !s.filterMatches(filter, record) {
			return false
		}
	}
	return true
}

// filterMatches evaluates a single filter against a record.
func (s *Service) filterMatches(filter RuleFilter, record emissions.EmissionRecord) bool {
	var fieldValue string

	switch filter.Field {
	case "scope":
		fieldValue = record.Scope.String()
	case "region":
		fieldValue = record.Region
	case "method":
		fieldValue = string(record.Method)
	case "org_id":
		fieldValue = record.OrgID
	default:
		// Unknown field - filter doesn't match
		return false
	}

	switch filter.Operator {
	case "eq", "=", "==":
		return fieldValue == filter.Value
	case "neq", "!=", "<>":
		return fieldValue != filter.Value
	case "contains":
		return contains(fieldValue, filter.Value)
	default:
		return false
	}
}

// executeAllocation performs the actual allocation calculation.
func (s *Service) executeAllocation(record emissions.EmissionRecord, rule Rule) (AllocationResult, error) {
	result := AllocationResult{
		OriginalID:  record.ID,
		RuleID:      rule.ID,
		Dimension:   rule.Dimension,
		AllocatedAt: time.Now().UTC(),
	}

	switch rule.Method {
	case MethodFixed:
		return s.executeFixedAllocation(record, rule, result)

	case MethodRevenue, MethodHeadcount, MethodArea, MethodEnergy, MethodProduction:
		return s.executeDriverAllocation(record, rule, result)

	case MethodExpression:
		return s.executeExpressionAllocation(record, rule, result)

	default:
		return result, fmt.Errorf("unsupported allocation method: %s", rule.Method)
	}
}

// executeFixedAllocation applies fixed percentage allocations.
func (s *Service) executeFixedAllocation(
	record emissions.EmissionRecord,
	rule Rule,
	result AllocationResult,
) (AllocationResult, error) {
	for _, target := range rule.Allocations {
		allocated := AllocatedEmission{
			TargetID:        target.TargetID,
			TargetName:      target.TargetName,
			Percentage:      target.Percentage,
			EmissionsKgCO2e: record.EmissionsKgCO2e * (target.Percentage / 100.0),
			SourceRecord:    record,
		}
		result.AllocatedEmissions = append(result.AllocatedEmissions, allocated)
	}

	return result, nil
}

// executeDriverAllocation calculates allocation based on driver data.
func (s *Service) executeDriverAllocation(
	record emissions.EmissionRecord,
	rule Rule,
	result AllocationResult,
) (AllocationResult, error) {
	s.mu.RLock()
	driverValues := s.driverData[string(rule.Dimension)]
	s.mu.RUnlock()

	if len(driverValues) == 0 {
		return result, fmt.Errorf("no driver data for dimension %s", rule.Dimension)
	}

	// Calculate total driver value
	var total float64
	for _, target := range rule.Allocations {
		if val, ok := driverValues[target.TargetID]; ok {
			total += val
		}
	}

	if total == 0 {
		return result, fmt.Errorf("total driver value is zero for dimension %s", rule.Dimension)
	}

	// Calculate proportional allocations
	for _, target := range rule.Allocations {
		driverVal := driverValues[target.TargetID]
		percentage := (driverVal / total) * 100.0

		allocated := AllocatedEmission{
			TargetID:        target.TargetID,
			TargetName:      target.TargetName,
			Percentage:      percentage,
			EmissionsKgCO2e: record.EmissionsKgCO2e * (driverVal / total),
			SourceRecord:    record,
		}
		result.AllocatedEmissions = append(result.AllocatedEmissions, allocated)
	}

	return result, nil
}

// executeExpressionAllocation evaluates an expression-based allocation.
func (s *Service) executeExpressionAllocation(
	record emissions.EmissionRecord,
	rule Rule,
	result AllocationResult,
) (AllocationResult, error) {
	if len(rule.Allocations) == 0 {
		return result, fmt.Errorf("expression allocation requires targets")
	}

	weights := parseExpressionWeights(rule.Expression, rule.Allocations)
	totalWeight := 0.0
	for _, w := range weights {
		totalWeight += w
	}

	// If expression failed to produce weights, fall back to equal distribution.
	if totalWeight == 0 {
		totalWeight = float64(len(rule.Allocations))
		for i := range rule.Allocations {
			weights[i] = 1.0
		}
	}

	for idx, target := range rule.Allocations {
		w := weights[idx]
		percentage := (w / totalWeight) * 100.0

		allocated := AllocatedEmission{
			TargetID:        target.TargetID,
			TargetName:      target.TargetName,
			Percentage:      percentage,
			EmissionsKgCO2e: record.EmissionsKgCO2e * (percentage / 100.0),
			SourceRecord:    record,
		}
		result.AllocatedEmissions = append(result.AllocatedEmissions, allocated)
	}

	return result, nil
}

// parseExpressionWeights parses simple weight expressions of the form
// "targetA:60,targetB:40" or "targetA=2;targetB=1".
// Unknown or invalid expressions return zero weights to trigger fallback.
func parseExpressionWeights(expr string, targets []AllocationTarget) []float64 {
	weights := make([]float64, len(targets))
	if strings.TrimSpace(expr) == "" {
		return weights
	}

	parts := strings.FieldsFunc(expr, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n'
	})
	if len(parts) == 0 {
		return weights
	}

	indexByID := make(map[string]int, len(targets))
	for i, t := range targets {
		indexByID[strings.ToLower(t.TargetID)] = i
	}

	for _, part := range parts {
		p := strings.TrimSpace(part)
		if p == "" {
			continue
		}
		split := strings.FieldsFunc(p, func(r rune) bool { return r == ':' || r == '=' })
		if len(split) != 2 {
			continue
		}
		id := strings.ToLower(strings.TrimSpace(split[0]))
		valStr := strings.TrimSpace(split[1])
		if idx, ok := indexByID[id]; ok {
			if v, err := strconv.ParseFloat(valStr, 64); err == nil && v > 0 {
				weights[idx] = v
			}
		}
	}
	return weights
}

// Metrics returns a copy of the current service metrics.
func (s *Service) Metrics() *ServiceMetrics {
	if s == nil || s.metrics == nil {
		return nil
	}

	return s.metrics.Clone()
}

// =============================================================================
// Helper Functions
// =============================================================================

// contains checks if a string contains a substring (case-insensitive).
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || len(substr) == 0 ||
			(len(s) > 0 && containsIgnoreCase(s, substr)))
}

// containsIgnoreCase is a simple case-insensitive contains check.
func containsIgnoreCase(s, substr string) bool {
	s = toLower(s)
	substr = toLower(substr)
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// toLower is a simple lowercase conversion.
func toLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			b[i] = c + 32
		} else {
			b[i] = c
		}
	}
	return string(b)
}
