// Package emissions/calculator provides the core emissions calculation engine.
//
// The Engine orchestrates calculations across all emission scopes, delegating
// to scope-specific calculators while providing unified access and aggregation.
package emissions

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
)

// =============================================================================
// Engine Configuration
// =============================================================================

// EngineConfig configures the emissions calculation engine.
type EngineConfig struct {
	// Registry provides emission factor lookup.
	Registry FactorRegistry

	// Logger for calculation events.
	Logger *slog.Logger

	// DefaultRegion is used when activity location is empty.
	DefaultRegion string

	// EnableParallelBatch enables concurrent batch processing.
	EnableParallelBatch bool

	// MaxBatchConcurrency limits concurrent calculations.
	MaxBatchConcurrency int

	// StrictMode fails on any calculation error rather than skipping.
	StrictMode bool
}

// DefaultEngineConfig returns sensible defaults for the engine.
func DefaultEngineConfig() EngineConfig {
	return EngineConfig{
		DefaultRegion:       "US-AVERAGE",
		EnableParallelBatch: true,
		MaxBatchConcurrency: 10,
		StrictMode:          false,
	}
}

// =============================================================================
// Engine Implementation
// =============================================================================

// Engine is the main emissions calculation orchestrator.
//
// It manages scope-specific calculators and provides a unified API for
// calculating emissions from activities. The engine handles factor lookup,
// calculator routing, and result aggregation.
//
// Example usage:
//
//	engine := emissions.NewEngine(config)
//	engine.RegisterCalculator(scope2.NewCalculator(registry))
//
//	result, err := engine.Calculate(ctx, activity)
//	if err != nil {
//	    log.Error("calculation failed", "error", err)
//	}
type Engine struct {
	config      EngineConfig
	registry    FactorRegistry
	calculators map[Scope]Calculator
	logger      *slog.Logger
	mu          sync.RWMutex
}

// NewEngine creates a new emissions calculation engine.
func NewEngine(cfg EngineConfig) *Engine {
	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}

	return &Engine{
		config:      cfg,
		registry:    cfg.Registry,
		calculators: make(map[Scope]Calculator),
		logger:      logger,
	}
}

// RegisterCalculator adds a scope-specific calculator.
func (e *Engine) RegisterCalculator(scope Scope, calc Calculator) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.calculators[scope] = calc
	e.logger.Info("registered calculator",
		"scope", scope.String(),
	)
}

// GetCalculator returns the calculator for the given scope.
func (e *Engine) GetCalculator(scope Scope) (Calculator, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	calc, ok := e.calculators[scope]
	return calc, ok
}

// Calculate computes emissions for a single activity.
//
// The engine automatically routes to the appropriate scope-specific calculator
// based on the activity source and available calculators.
func (e *Engine) Calculate(ctx context.Context, activity Activity) (EmissionRecord, error) {
	if activity == nil {
		return EmissionRecord{}, ErrNilActivity
	}

	start := time.Now()

	// Determine appropriate scope
	scope := e.determineScope(activity)

	e.logger.Debug("calculating emissions",
		"activity_id", activity.GetID(),
		"source", activity.GetSource(),
		"scope", scope.String(),
		"quantity", activity.GetQuantity(),
		"unit", activity.GetUnit(),
	)

	// Get the appropriate calculator
	calc, ok := e.GetCalculator(scope)
	if !ok {
		return EmissionRecord{}, fmt.Errorf(
			"no calculator registered for %s: %w",
			scope.String(), ErrCalculationFailed,
		)
	}

	// Check if calculator supports this activity
	if !calc.Supports(activity) {
		return EmissionRecord{}, fmt.Errorf(
			"calculator does not support activity source=%s unit=%s: %w",
			activity.GetSource(), activity.GetUnit(), ErrUnsupportedUnit,
		)
	}

	// Perform calculation
	record, err := calc.Calculate(ctx, activity)
	if err != nil {
		e.logger.Error("calculation failed",
			"activity_id", activity.GetID(),
			"scope", scope.String(),
			"error", err,
		)
		return EmissionRecord{}, fmt.Errorf("calculate %s: %w", activity.GetID(), err)
	}

	duration := time.Since(start)
	e.logger.Info("calculated emissions",
		"activity_id", activity.GetID(),
		"scope", scope.String(),
		"emissions_kg_co2e", record.EmissionsKgCO2e,
		"duration_ms", duration.Milliseconds(),
	)

	return record, nil
}

// CalculateBatch processes multiple activities and returns aggregated results.
func (e *Engine) CalculateBatch(ctx context.Context, activities []Activity) (BatchResult, error) {
	start := time.Now()
	result := BatchResult{
		Records: make([]EmissionRecord, 0, len(activities)),
		Results: make([]CalculationResult, 0, len(activities)),
	}

	if len(activities) == 0 {
		result.ProcessedAt = time.Now()
		return result, nil
	}

	e.logger.Info("starting batch calculation",
		"activity_count", len(activities),
		"parallel", e.config.EnableParallelBatch,
	)

	processActivity := func(a Activity) CalculationResult {
		record, err := e.Calculate(ctx, a)
		if err != nil {
			return CalculationResult{
				ActivityID: a.GetID(),
				Error:      err,
			}
		}
		return CalculationResult{
			ActivityID: a.GetID(),
			Record:     &record,
		}
	}

	if e.config.EnableParallelBatch && len(activities) > 1 {
		result.Results = e.calculateParallel(ctx, activities, processActivity)
	} else {
		for _, a := range activities {
			result.Results = append(result.Results, processActivity(a))
		}
	}

	// Aggregate results
	for _, r := range result.Results {
		if r.IsSuccess() {
			result.Records = append(result.Records, *r.Record)
			result.TotalEmissionsKgCO2e += r.Record.EmissionsKgCO2e
			result.SuccessCount++
		} else {
			result.ErrorCount++
			if e.config.StrictMode {
				return result, fmt.Errorf(
					"strict mode: failed to calculate activity %s: %w",
					r.ActivityID, r.Error,
				)
			}
		}
	}

	result.ProcessedAt = time.Now()

	e.logger.Info("batch calculation complete",
		"total_activities", len(activities),
		"success_count", result.SuccessCount,
		"error_count", result.ErrorCount,
		"total_emissions_kg_co2e", result.TotalEmissionsKgCO2e,
		"duration_ms", time.Since(start).Milliseconds(),
	)

	return result, nil
}

// calculateParallel processes activities concurrently with bounded parallelism.
func (e *Engine) calculateParallel(
	ctx context.Context,
	activities []Activity,
	process func(Activity) CalculationResult,
) []CalculationResult {
	results := make([]CalculationResult, len(activities))
	semaphore := make(chan struct{}, e.config.MaxBatchConcurrency)

	var wg sync.WaitGroup

	for i, activity := range activities {
		wg.Add(1)

		go func(idx int, a Activity) {
			defer wg.Done()

			// Acquire semaphore slot
			select {
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
			case <-ctx.Done():
				results[idx] = CalculationResult{
					ActivityID: a.GetID(),
					Error:      ctx.Err(),
				}
				return
			}

			results[idx] = process(a)
		}(i, activity)
	}

	wg.Wait()
	return results
}

// determineScope infers the emission scope from activity characteristics.
func (e *Engine) determineScope(activity Activity) Scope {
	source := activity.GetSource()

	switch source {
	case "fleet", "on-site", "refrigerants", "stationary_combustion":
		return Scope1
	case "utility_bill", "electricity", "steam", "heating", "cooling":
		return Scope2
	case "travel", "commuting", "purchases", "waste", "upstream", "downstream":
		return Scope3
	default:
		// Default to Scope 2 for utility-style data
		return Scope2
	}
}

// Registry returns the engine's factor registry.
func (e *Engine) Registry() FactorRegistry {
	return e.registry
}

// =============================================================================
// ID Generation
// =============================================================================

// GenerateRecordID creates a unique emission record identifier.
func GenerateRecordID() string {
	return fmt.Sprintf("em_%s", uuid.New().String()[:12])
}

// =============================================================================
// Adapter Functions
// =============================================================================

// ActivityAdapter wraps an ingestion.Activity to implement the Activity interface.
// This allows the emissions package to work with activity data from ingestion.
type ActivityAdapter struct {
	ID          string
	Source      string
	Category    string
	Location    string
	Quantity    float64
	Unit        string
	PeriodStart time.Time
	PeriodEnd   time.Time
	OrgID       string
	WorkspaceID string
}

// GetID implements Activity.
func (a ActivityAdapter) GetID() string { return a.ID }

// GetSource implements Activity.
func (a ActivityAdapter) GetSource() string { return a.Source }

// GetCategory implements Activity.
func (a ActivityAdapter) GetCategory() string { return a.Category }

// GetLocation implements Activity.
func (a ActivityAdapter) GetLocation() string { return a.Location }

// GetQuantity implements Activity.
func (a ActivityAdapter) GetQuantity() float64 { return a.Quantity }

// GetUnit implements Activity.
func (a ActivityAdapter) GetUnit() string { return a.Unit }

// GetPeriodStart implements Activity.
func (a ActivityAdapter) GetPeriodStart() time.Time { return a.PeriodStart }

// GetPeriodEnd implements Activity.
func (a ActivityAdapter) GetPeriodEnd() time.Time { return a.PeriodEnd }

// GetOrgID implements Activity.
func (a ActivityAdapter) GetOrgID() string { return a.OrgID }

// GetWorkspaceID implements Activity.
func (a ActivityAdapter) GetWorkspaceID() string { return a.WorkspaceID }

// =============================================================================
// Calculation Utilities
// =============================================================================

// SumEmissions calculates total emissions from a slice of records.
func SumEmissions(records []EmissionRecord) float64 {
	var total float64
	for _, r := range records {
		total += r.EmissionsKgCO2e
	}
	return total
}

// SumEmissionsByScope groups and sums emissions by scope.
func SumEmissionsByScope(records []EmissionRecord) map[Scope]float64 {
	totals := make(map[Scope]float64)
	for _, r := range records {
		totals[r.Scope] += r.EmissionsKgCO2e
	}
	return totals
}

// FilterByScope returns records matching the given scope.
func FilterByScope(records []EmissionRecord, scope Scope) []EmissionRecord {
	var filtered []EmissionRecord
	for _, r := range records {
		if r.Scope == scope {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

// FilterByOrg returns records for the given organization.
func FilterByOrg(records []EmissionRecord, orgID string) []EmissionRecord {
	var filtered []EmissionRecord
	for _, r := range records {
		if r.OrgID == orgID {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

// KgToTonnes converts kilograms to metric tonnes.
func KgToTonnes(kg float64) float64 {
	return kg / 1000.0
}

// TonnesToKg converts metric tonnes to kilograms.
func TonnesToKg(tonnes float64) float64 {
	return tonnes * 1000.0
}

// =============================================================================
// Error Helpers
// =============================================================================

// IsNotFoundError returns true if the error indicates a missing factor.
func IsNotFoundError(err error) bool {
	return errors.Is(err, ErrFactorNotFound)
}

// IsUnsupportedError returns true if the error indicates an unsupported activity.
func IsUnsupportedError(err error) bool {
	return errors.Is(err, ErrUnsupportedUnit)
}
