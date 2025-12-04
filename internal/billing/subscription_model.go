package billing

import (
	"time"
)

// SubscriptionStatus enumerates Stripe subscription states.
type SubscriptionStatus string

const (
	StatusTrialing SubscriptionStatus = "trialing"
	StatusActive   SubscriptionStatus = "active"
	StatusPastDue  SubscriptionStatus = "past_due"
	StatusCanceled SubscriptionStatus = "canceled"
	StatusUnpaid   SubscriptionStatus = "unpaid"
)

// Subscription represents a tenant's billing subscription.
type Subscription struct {
	ID                   string
	TenantID             string
	StripeCustomerID     string
	StripeSubscriptionID string
	Plan                 string // "basic" | "pro"
	Status               SubscriptionStatus
	CurrentPeriodEnd     *time.Time
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// IsActive checks if the subscription currently allows access to premium features.
func (s *Subscription) IsActive() bool {
	return s.Status == StatusTrialing || s.Status == StatusActive
}
