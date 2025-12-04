// Package emissions provides test helpers and utilities for emissions testing.
package emissions

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// InMemoryRegistry is a simple in-memory implementation of FactorRegistry
// for testing purposes. It stores emission factors in memory and provides
// basic lookup functionality.
type InMemoryRegistry struct {
	factors map[string]EmissionFactor
	mu      sync.RWMutex
}

// NewInMemoryRegistry creates a new in-memory factor registry for testing.
func NewInMemoryRegistry() *InMemoryRegistry {
	return &InMemoryRegistry{
		factors: make(map[string]EmissionFactor),
	}
}

// GetFactor retrieves a factor by its unique ID.
func (r *InMemoryRegistry) GetFactor(ctx context.Context, id string) (EmissionFactor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	factor, ok := r.factors[id]
	if !ok {
		return EmissionFactor{}, fmt.Errorf("%w: id=%s", ErrFactorNotFound, id)
	}

	return factor, nil
}

// FindFactor looks up the best matching factor for an activity.
func (r *InMemoryRegistry) FindFactor(ctx context.Context, query FactorQuery) (EmissionFactor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Iterate through factors to find best match
	var bestMatch *EmissionFactor
	var bestScore int

	for _, factor := range r.factors {
		if !query.Matches(factor) {
			continue
		}

		// Calculate match score (higher is better)
		score := 0
		if query.Scope != ScopeUnspecified && factor.Scope == query.Scope {
			score += 10
		}
		if query.Region != "" && factor.Region == query.Region {
			score += 5
		}
		if query.Source != "" && factor.Source == query.Source {
			score += 5
		}
		if query.Category != "" && factor.Category == query.Category {
			score += 3
		}
		if query.Unit != "" && factor.Unit == query.Unit {
			score += 2
		}

		// Prefer currently valid factors
		if query.ValidAt.IsZero() || factor.IsCurrentlyValid(query.ValidAt) {
			score += 1
		}

		if bestMatch == nil || score > bestScore {
			bestScore = score
			factorCopy := factor
			bestMatch = &factorCopy
		}
	}

	if bestMatch == nil {
		return EmissionFactor{}, fmt.Errorf(
			"%w: scope=%s region=%s source=%s category=%s unit=%s",
			ErrFactorNotFound,
			query.Scope,
			query.Region,
			query.Source,
			query.Category,
			query.Unit,
		)
	}

	return *bestMatch, nil
}

// ListFactors returns all factors matching the given criteria.
func (r *InMemoryRegistry) ListFactors(ctx context.Context, query FactorQuery) ([]EmissionFactor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var matches []EmissionFactor

	for _, factor := range r.factors {
		if query.Matches(factor) {
			matches = append(matches, factor)
		}
	}

	return matches, nil
}

// RegisterFactor adds or updates a factor in the registry.
func (r *InMemoryRegistry) RegisterFactor(ctx context.Context, factor EmissionFactor) error {
	if !factor.IsValid() {
		return errors.New("invalid emission factor: missing required fields")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.factors[factor.ID] = factor
	return nil
}

// Clear removes all factors from the registry (useful for test cleanup).
func (r *InMemoryRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.factors = make(map[string]EmissionFactor)
}

// Count returns the number of factors in the registry.
func (r *InMemoryRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.factors)
}
