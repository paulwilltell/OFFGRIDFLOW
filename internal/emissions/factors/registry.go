// Package factors provides emission factor registry implementations.
//
// The registry stores and retrieves emission factors used in carbon
// accounting calculations. It supports multiple scopes, regions, and
// factor sources while handling versioning and time-validity.
package factors

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/example/offgridflow/internal/emissions"
)

// =============================================================================
// Registry Configuration
// =============================================================================

// RegistryConfig configures the in-memory factor registry.
type RegistryConfig struct {
	// Logger for registry operations.
	Logger *slog.Logger

	// PreloadDefaults seeds the registry with default factors.
	PreloadDefaults bool

	// ValidateOnRegister checks factor validity on registration.
	ValidateOnRegister bool
}

// DefaultRegistryConfig returns sensible defaults.
func DefaultRegistryConfig() RegistryConfig {
	return RegistryConfig{
		PreloadDefaults:    true,
		ValidateOnRegister: true,
	}
}

// =============================================================================
// In-Memory Registry Implementation
// =============================================================================

// InMemoryRegistry stores emission factors in memory.
//
// The registry supports:
//   - Scope-based organization (Scope 1, 2, 3)
//   - Region-specific factors
//   - Time-validity periods
//   - Concurrent access with read-write locking
//
// For production use, consider a database-backed implementation that
// extends this interface.
type InMemoryRegistry struct {
	factors map[string]emissions.EmissionFactor
	logger  *slog.Logger
	config  RegistryConfig
	mu      sync.RWMutex
}

// NewInMemoryRegistry creates a new in-memory factor registry.
func NewInMemoryRegistry(cfg RegistryConfig) *InMemoryRegistry {
	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}

	r := &InMemoryRegistry{
		factors: make(map[string]emissions.EmissionFactor),
		logger:  logger,
		config:  cfg,
	}

	if cfg.PreloadDefaults {
		r.seedDefaults()
	}

	return r
}

// GetFactor retrieves a factor by its unique ID.
func (r *InMemoryRegistry) GetFactor(ctx context.Context, id string) (emissions.EmissionFactor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	factor, ok := r.factors[id]
	if !ok {
		return emissions.EmissionFactor{}, fmt.Errorf(
			"factor %q: %w", id, emissions.ErrFactorNotFound,
		)
	}

	return factor, nil
}

// FindFactor looks up the best matching factor for an activity.
func (r *InMemoryRegistry) FindFactor(ctx context.Context, query emissions.FactorQuery) (emissions.EmissionFactor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var matches []emissions.EmissionFactor

	for _, factor := range r.factors {
		if query.Matches(factor) {
			matches = append(matches, factor)
		}
	}

	if len(matches) == 0 {
		return emissions.EmissionFactor{}, fmt.Errorf(
			"no factor matching scope=%s region=%q source=%q unit=%q: %w",
			query.Scope, query.Region, query.Source, query.Unit,
			emissions.ErrFactorNotFound,
		)
	}

	// Sort by specificity (more specific matches first)
	sort.Slice(matches, func(i, j int) bool {
		return r.specificity(matches[i], query) > r.specificity(matches[j], query)
	})

	best := matches[0]

	r.logger.Debug("found emission factor",
		"factor_id", best.ID,
		"scope", best.Scope.String(),
		"region", best.Region,
		"value", best.ValueKgCO2ePerUnit,
	)

	return best, nil
}

// ListFactors returns all factors matching the given criteria.
func (r *InMemoryRegistry) ListFactors(ctx context.Context, query emissions.FactorQuery) ([]emissions.EmissionFactor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var matches []emissions.EmissionFactor

	for _, factor := range r.factors {
		if query.Matches(factor) {
			matches = append(matches, factor)
		}
	}

	// Sort by ID for consistent ordering
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].ID < matches[j].ID
	})

	return matches, nil
}

// RegisterFactor adds or updates a factor in the registry.
func (r *InMemoryRegistry) RegisterFactor(ctx context.Context, factor emissions.EmissionFactor) error {
	if r.config.ValidateOnRegister && !factor.IsValid() {
		return fmt.Errorf("invalid emission factor %q: missing required fields", factor.ID)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.factors[factor.ID] = factor

	r.logger.Info("registered emission factor",
		"factor_id", factor.ID,
		"scope", factor.Scope.String(),
		"region", factor.Region,
		"value", factor.ValueKgCO2ePerUnit,
	)

	return nil
}

// Count returns the number of factors in the registry.
func (r *InMemoryRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.factors)
}

// Clear removes all factors from the registry.
func (r *InMemoryRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factors = make(map[string]emissions.EmissionFactor)
}

// specificity calculates how closely a factor matches the query.
// Higher values indicate a more specific match.
func (r *InMemoryRegistry) specificity(factor emissions.EmissionFactor, query emissions.FactorQuery) int {
	score := 0

	// Exact region match is most valuable
	if query.Region != "" && strings.EqualFold(factor.Region, query.Region) {
		score += 100
	}

	// Category match adds precision
	if query.Category != "" && strings.EqualFold(factor.Category, query.Category) {
		score += 50
	}

	// Exact source match
	if query.Source != "" && strings.EqualFold(factor.Source, query.Source) {
		score += 25
	}

	// Time validity
	if !query.ValidAt.IsZero() && factor.IsCurrentlyValid(query.ValidAt) {
		score += 10
	}

	// Prefer factors with data sources
	if factor.DataSource != "" {
		score += 5
	}

	return score
}

// seedDefaults populates the registry with standard emission factors.
func (r *InMemoryRegistry) seedDefaults() {
	now := time.Now().UTC()

	// Scope 2: Electricity Grid Emission Factors
	// Source: IEA, EPA eGRID 2023, EEA 2023
	scope2Factors := []emissions.EmissionFactor{
		{
			ID:                 "grid-uk",
			Scope:              emissions.Scope2,
			Region:             "UK",
			Source:             "electricity",
			Unit:               "kWh",
			ValueKgCO2ePerUnit: 0.193,
			Method:             emissions.MethodLocationBased,
			DataSource:         "UK DEFRA 2024",
			CreatedAt:          now,
		},
		{
			ID:                 "grid-germany",
			Scope:              emissions.Scope2,
			Region:             "DE",
			Source:             "electricity",
			Unit:               "kWh",
			ValueKgCO2ePerUnit: 0.366,
			Method:             emissions.MethodLocationBased,
			DataSource:         "EEA 2023 (Germany)",
			CreatedAt:          now,
		},
		{
			ID:                 "grid-france",
			Scope:              emissions.Scope2,
			Region:             "FR",
			Source:             "electricity",
			Unit:               "kWh",
			ValueKgCO2ePerUnit: 0.056,
			Method:             emissions.MethodLocationBased,
			DataSource:         "EEA 2023 (France nuclear heavy mix)",
			CreatedAt:          now,
		},
		{
			ID:                 "grid-india",
			Scope:              emissions.Scope2,
			Region:             "IN",
			Source:             "electricity",
			Unit:               "kWh",
			ValueKgCO2ePerUnit: 0.708,
			Method:             emissions.MethodLocationBased,
			DataSource:         "IEA 2023 (India)",
			CreatedAt:          now,
		},
		{
			ID:                 "grid-china",
			Scope:              emissions.Scope2,
			Region:             "CN",
			Source:             "electricity",
			Unit:               "kWh",
			ValueKgCO2ePerUnit: 0.681,
			Method:             emissions.MethodLocationBased,
			DataSource:         "IEA 2023 (China)",
			CreatedAt:          now,
		},
		// United States regions
		{
			ID:                 "grid-us-west",
			Scope:              emissions.Scope2,
			Region:             "US-WEST",
			Source:             "electricity",
			Unit:               "kWh",
			ValueKgCO2ePerUnit: 0.298,
			Method:             emissions.MethodLocationBased,
			DataSource:         "EPA eGRID 2023 (WECC)",
			CreatedAt:          now,
		},
		{
			ID:                 "grid-us-east",
			Scope:              emissions.Scope2,
			Region:             "US-EAST",
			Source:             "electricity",
			Unit:               "kWh",
			ValueKgCO2ePerUnit: 0.388,
			Method:             emissions.MethodLocationBased,
			DataSource:         "EPA eGRID 2023 (NPCC/RFC/SERC)",
			CreatedAt:          now,
		},
		{
			ID:                 "grid-us-texas",
			Scope:              emissions.Scope2,
			Region:             "US-TEXAS",
			Source:             "electricity",
			Unit:               "kWh",
			ValueKgCO2ePerUnit: 0.395,
			Method:             emissions.MethodLocationBased,
			DataSource:         "EPA eGRID 2023 (ERCOT)",
			CreatedAt:          now,
		},
		{
			ID:                 "grid-us-midwest",
			Scope:              emissions.Scope2,
			Region:             "US-MIDWEST",
			Source:             "electricity",
			Unit:               "kWh",
			ValueKgCO2ePerUnit: 0.452,
			Method:             emissions.MethodLocationBased,
			DataSource:         "EPA eGRID 2023 (MRO)",
			CreatedAt:          now,
		},
		{
			ID:                 "grid-us-average",
			Scope:              emissions.Scope2,
			Region:             "US-AVERAGE",
			Source:             "electricity",
			Unit:               "kWh",
			ValueKgCO2ePerUnit: 0.386,
			Method:             emissions.MethodLocationBased,
			DataSource:         "EPA eGRID 2023 (US Average)",
			CreatedAt:          now,
		},

		// European regions
		{
			ID:                 "grid-eu-central",
			Scope:              emissions.Scope2,
			Region:             "EU-CENTRAL",
			Source:             "electricity",
			Unit:               "kWh",
			ValueKgCO2ePerUnit: 0.350,
			Method:             emissions.MethodLocationBased,
			DataSource:         "EEA 2023 (DE/PL)",
			CreatedAt:          now,
		},
		{
			ID:                 "grid-eu-north",
			Scope:              emissions.Scope2,
			Region:             "EU-NORTH",
			Source:             "electricity",
			Unit:               "kWh",
			ValueKgCO2ePerUnit: 0.150,
			Method:             emissions.MethodLocationBased,
			DataSource:         "EEA 2023 (Nordic)",
			CreatedAt:          now,
		},
		{
			ID:                 "grid-eu-west",
			Scope:              emissions.Scope2,
			Region:             "EU-WEST",
			Source:             "electricity",
			Unit:               "kWh",
			ValueKgCO2ePerUnit: 0.185,
			Method:             emissions.MethodLocationBased,
			DataSource:         "EEA 2023 (FR/BE)",
			CreatedAt:          now,
		},
		{
			ID:                 "grid-eu-south",
			Scope:              emissions.Scope2,
			Region:             "EU-SOUTH",
			Source:             "electricity",
			Unit:               "kWh",
			ValueKgCO2ePerUnit: 0.295,
			Method:             emissions.MethodLocationBased,
			DataSource:         "EEA 2023 (IT/ES)",
			CreatedAt:          now,
		},
		{
			ID:                 "grid-eu-average",
			Scope:              emissions.Scope2,
			Region:             "EU-AVERAGE",
			Source:             "electricity",
			Unit:               "kWh",
			ValueKgCO2ePerUnit: 0.276,
			Method:             emissions.MethodLocationBased,
			DataSource:         "EEA 2023 (EU Average)",
			CreatedAt:          now,
		},

		// Asia-Pacific regions
		{
			ID:                 "grid-asia-pacific",
			Scope:              emissions.Scope2,
			Region:             "ASIA-PACIFIC",
			Source:             "electricity",
			Unit:               "kWh",
			ValueKgCO2ePerUnit: 0.550,
			Method:             emissions.MethodLocationBased,
			DataSource:         "IEA 2023 (APAC Average)",
			CreatedAt:          now,
		},
		{
			ID:                 "grid-asia-japan",
			Scope:              emissions.Scope2,
			Region:             "ASIA-JAPAN",
			Source:             "electricity",
			Unit:               "kWh",
			ValueKgCO2ePerUnit: 0.470,
			Method:             emissions.MethodLocationBased,
			DataSource:         "IEA 2023 (Japan)",
			CreatedAt:          now,
		},
		{
			ID:                 "grid-asia-australia",
			Scope:              emissions.Scope2,
			Region:             "ASIA-AUSTRALIA",
			Source:             "electricity",
			Unit:               "kWh",
			ValueKgCO2ePerUnit: 0.656,
			Method:             emissions.MethodLocationBased,
			DataSource:         "IEA 2023 (Australia)",
			CreatedAt:          now,
		},

		// Other major regions
		{
			ID:                 "grid-canada",
			Scope:              emissions.Scope2,
			Region:             "CANADA",
			Source:             "electricity",
			Unit:               "kWh",
			ValueKgCO2ePerUnit: 0.130,
			Method:             emissions.MethodLocationBased,
			DataSource:         "IEA 2023 (Canada)",
			CreatedAt:          now,
		},
		{
			ID:                 "grid-latam-brazil",
			Scope:              emissions.Scope2,
			Region:             "LATAM-BRAZIL",
			Source:             "electricity",
			Unit:               "kWh",
			ValueKgCO2ePerUnit: 0.075,
			Method:             emissions.MethodLocationBased,
			DataSource:         "IEA 2023 (Brazil - mostly hydro)",
			CreatedAt:          now,
		},
	}

	// Scope 1: Fuel Combustion Factors
	// Source: EPA GHG Emission Factors Hub
	scope1Factors := []emissions.EmissionFactor{
		{
			ID:                 "fuel-diesel",
			Scope:              emissions.Scope1,
			Region:             "GLOBAL",
			Source:             "fleet",
			Category:           "diesel",
			Unit:               "L",
			ValueKgCO2ePerUnit: 2.68,
			Method:             emissions.MethodActivityBased,
			DataSource:         "EPA GHG Emission Factors Hub",
			CreatedAt:          now,
		},
		{
			ID:                 "fuel-fuel-oil-2",
			Scope:              emissions.Scope1,
			Region:             "GLOBAL",
			Source:             "stationary_combustion",
			Category:           "fuel_oil_2",
			Unit:               "L",
			ValueKgCO2ePerUnit: 2.96,
			Method:             emissions.MethodActivityBased,
			DataSource:         "EPA GHG Emission Factors Hub",
			CreatedAt:          now,
		},
		{
			ID:                 "fuel-fuel-oil-6",
			Scope:              emissions.Scope1,
			Region:             "GLOBAL",
			Source:             "stationary_combustion",
			Category:           "fuel_oil_6",
			Unit:               "L",
			ValueKgCO2ePerUnit: 3.25,
			Method:             emissions.MethodActivityBased,
			DataSource:         "EPA GHG Emission Factors Hub",
			CreatedAt:          now,
		},
		{
			ID:                 "fuel-gasoline",
			Scope:              emissions.Scope1,
			Region:             "GLOBAL",
			Source:             "fleet",
			Category:           "gasoline",
			Unit:               "L",
			ValueKgCO2ePerUnit: 2.31,
			Method:             emissions.MethodActivityBased,
			DataSource:         "EPA GHG Emission Factors Hub",
			CreatedAt:          now,
		},
		{
			ID:                 "fuel-natural-gas",
			Scope:              emissions.Scope1,
			Region:             "GLOBAL",
			Source:             "stationary_combustion",
			Category:           "natural_gas",
			Unit:               "m3",
			ValueKgCO2ePerUnit: 1.93,
			Method:             emissions.MethodActivityBased,
			DataSource:         "EPA GHG Emission Factors Hub",
			CreatedAt:          now,
		},
		{
			ID:                 "fuel-propane",
			Scope:              emissions.Scope1,
			Region:             "GLOBAL",
			Source:             "stationary_combustion",
			Category:           "propane",
			Unit:               "L",
			ValueKgCO2ePerUnit: 1.51,
			Method:             emissions.MethodActivityBased,
			DataSource:         "EPA GHG Emission Factors Hub",
			CreatedAt:          now,
		},
	}

	// Register all factors
	for _, factor := range scope2Factors {
		r.factors[factor.ID] = factor
	}

	for _, factor := range scope1Factors {
		r.factors[factor.ID] = factor
	}

	r.logger.Info("seeded default emission factors",
		"scope1_count", len(scope1Factors),
		"scope2_count", len(scope2Factors),
		"total_count", len(r.factors),
	)
}

// =============================================================================
// Composite Registry
// =============================================================================

// CompositeRegistry combines multiple registries with fallback lookup.
//
// This is useful for layering:
//   - Custom organization factors (highest priority)
//   - Database-backed factors
//   - Default in-memory factors (fallback)
type CompositeRegistry struct {
	registries []emissions.FactorRegistry
	logger     *slog.Logger
}

// NewCompositeRegistry creates a registry that searches multiple sources.
// Registries are searched in order; first match wins.
func NewCompositeRegistry(logger *slog.Logger, registries ...emissions.FactorRegistry) *CompositeRegistry {
	if logger == nil {
		logger = slog.Default()
	}

	return &CompositeRegistry{
		registries: registries,
		logger:     logger,
	}
}

// GetFactor searches registries in order for the factor.
func (c *CompositeRegistry) GetFactor(ctx context.Context, id string) (emissions.EmissionFactor, error) {
	for i, r := range c.registries {
		factor, err := r.GetFactor(ctx, id)
		if err == nil {
			c.logger.Debug("found factor in registry",
				"factor_id", id,
				"registry_index", i,
			)
			return factor, nil
		}
	}

	return emissions.EmissionFactor{}, fmt.Errorf(
		"factor %q not found in any registry: %w",
		id, emissions.ErrFactorNotFound,
	)
}

// FindFactor searches registries in order for a matching factor.
func (c *CompositeRegistry) FindFactor(ctx context.Context, query emissions.FactorQuery) (emissions.EmissionFactor, error) {
	for i, r := range c.registries {
		factor, err := r.FindFactor(ctx, query)
		if err == nil {
			c.logger.Debug("found matching factor in registry",
				"factor_id", factor.ID,
				"registry_index", i,
			)
			return factor, nil
		}
	}

	return emissions.EmissionFactor{}, fmt.Errorf(
		"no matching factor found in any registry: %w",
		emissions.ErrFactorNotFound,
	)
}

// ListFactors aggregates factors from all registries.
func (c *CompositeRegistry) ListFactors(ctx context.Context, query emissions.FactorQuery) ([]emissions.EmissionFactor, error) {
	seen := make(map[string]bool)
	var all []emissions.EmissionFactor

	for _, r := range c.registries {
		factors, err := r.ListFactors(ctx, query)
		if err != nil {
			continue
		}

		for _, f := range factors {
			if !seen[f.ID] {
				seen[f.ID] = true
				all = append(all, f)
			}
		}
	}

	return all, nil
}

// RegisterFactor registers to the first registry.
func (c *CompositeRegistry) RegisterFactor(ctx context.Context, factor emissions.EmissionFactor) error {
	if len(c.registries) == 0 {
		return emissions.ErrRegistryNotAvailable
	}

	return c.registries[0].RegisterFactor(ctx, factor)
}

// =============================================================================
// Helper Functions
// =============================================================================

// NewDefaultRegistry creates a pre-populated in-memory registry.
func NewDefaultRegistry() *InMemoryRegistry {
	return NewInMemoryRegistry(DefaultRegistryConfig())
}

// FactorSummary provides an overview of registry contents.
type FactorSummary struct {
	TotalCount int            `json:"total_count"`
	ByScope    map[string]int `json:"by_scope"`
	ByRegion   map[string]int `json:"by_region"`
	BySource   map[string]int `json:"by_source"`
	DateRange  [2]time.Time   `json:"date_range"`
}

// Summarize returns statistics about factors in the registry.
func (r *InMemoryRegistry) Summarize() FactorSummary {
	r.mu.RLock()
	defer r.mu.RUnlock()

	summary := FactorSummary{
		TotalCount: len(r.factors),
		ByScope:    make(map[string]int),
		ByRegion:   make(map[string]int),
		BySource:   make(map[string]int),
	}

	var minDate, maxDate time.Time

	for _, f := range r.factors {
		summary.ByScope[f.Scope.String()]++
		summary.ByRegion[f.Region]++
		summary.BySource[f.Source]++

		if !f.CreatedAt.IsZero() {
			if minDate.IsZero() || f.CreatedAt.Before(minDate) {
				minDate = f.CreatedAt
			}
			if maxDate.IsZero() || f.CreatedAt.After(maxDate) {
				maxDate = f.CreatedAt
			}
		}
	}

	summary.DateRange = [2]time.Time{minDate, maxDate}

	return summary
}
