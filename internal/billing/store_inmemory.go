// Package billing provides the in-memory store implementation for testing.
package billing

import (
	"context"
	"errors"
	"sync"
	"time"
)

// InMemoryStore provides a thread-safe in-memory implementation of Store.
// Used for testing and development without a database.
type InMemoryStore struct {
	mu            sync.RWMutex
	byTenantID    map[string]*Subscription
	byCustomerID  map[string]*Subscription
}

// NewInMemoryStore creates a new in-memory subscription store.
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		byTenantID:   make(map[string]*Subscription),
		byCustomerID: make(map[string]*Subscription),
	}
}

// GetByTenantID retrieves a subscription by tenant ID.
func (s *InMemoryStore) GetByTenantID(ctx context.Context, tenantID string) (*Subscription, error) {
	if tenantID == "" {
		return nil, errors.New("billing: tenant ID required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	sub, ok := s.byTenantID[tenantID]
	if !ok {
		return nil, nil
	}

	// Return a copy to prevent external mutation
	copy := *sub
	return &copy, nil
}

// GetByStripeCustomer retrieves a subscription by Stripe customer ID.
func (s *InMemoryStore) GetByStripeCustomer(ctx context.Context, customerID string) (*Subscription, error) {
	if customerID == "" {
		return nil, errors.New("billing: customer ID required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	sub, ok := s.byCustomerID[customerID]
	if !ok {
		return nil, nil
	}

	// Return a copy to prevent external mutation
	copy := *sub
	return &copy, nil
}

// Upsert creates or updates a subscription.
func (s *InMemoryStore) Upsert(ctx context.Context, sub *Subscription) error {
	if sub == nil {
		return errors.New("billing: subscription required")
	}
	if sub.ID == "" {
		return errors.New("billing: subscription ID required")
	}
	if sub.TenantID == "" {
		return errors.New("billing: tenant ID required")
	}

	// Validate and normalize status
	switch sub.Status {
	case StatusTrialing, StatusActive, StatusPastDue, StatusCanceled, StatusUnpaid:
		// Valid status
	default:
		return errors.New("billing: invalid subscription status: " + string(sub.Status))
	}

	// Update timestamps
	if sub.CreatedAt.IsZero() {
		sub.CreatedAt = time.Now()
	}
	sub.UpdatedAt = time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	// Store by tenant ID
	oldSub := s.byTenantID[sub.TenantID]
	s.byTenantID[sub.TenantID] = sub

	// Update customer ID mapping
	if oldSub != nil && oldSub.StripeCustomerID != "" {
		delete(s.byCustomerID, oldSub.StripeCustomerID)
	}
	if sub.StripeCustomerID != "" {
		s.byCustomerID[sub.StripeCustomerID] = sub
	}

	return nil
}

// List returns all subscriptions (useful for testing and metrics).
func (s *InMemoryStore) List(ctx context.Context) ([]*Subscription, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	subs := make([]*Subscription, 0, len(s.byTenantID))
	for _, sub := range s.byTenantID {
		copy := *sub
		subs = append(subs, &copy)
	}
	return subs, nil
}

// Clear removes all subscriptions (useful for testing).
func (s *InMemoryStore) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.byTenantID = make(map[string]*Subscription)
	s.byCustomerID = make(map[string]*Subscription)
}

// Count returns the number of subscriptions.
func (s *InMemoryStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.byTenantID)
}
