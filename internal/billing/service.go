package billing

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v82"
)

// Store abstracts subscription persistence.
type Store interface {
	GetByTenantID(ctx context.Context, tenantID string) (*Subscription, error)
	GetByStripeCustomer(ctx context.Context, customerID string) (*Subscription, error)
	Upsert(ctx context.Context, sub *Subscription) error
}

// Service provides high-level billing operations.
type Service struct {
	stripe *StripeClient
	store  Store
}

// NewService constructs a billing service.
func NewService(stripeClient *StripeClient, store Store) *Service {
	return &Service{stripe: stripeClient, store: store}
}

// StartSubscription initiates checkout and returns the Stripe-hosted URL.
func (s *Service) StartSubscription(ctx context.Context, tenantID, tenantName, email, plan, successURL, cancelURL string) (string, error) {
	// Check if subscription already exists (and hence customer)
	sub, _ := s.store.GetByTenantID(ctx, tenantID)
	var customerID string
	if sub != nil && sub.StripeCustomerID != "" {
		customerID = sub.StripeCustomerID
	} else {
		var err error
		customerID, err = s.stripe.CreateCustomer(email, tenantName, tenantID)
		if err != nil {
			return "", err
		}
		// Persist customer link
		newSub := &Subscription{
			ID:               uuid.NewString(),
			TenantID:         tenantID,
			StripeCustomerID: customerID,
			Status:           StatusTrialing,
			Plan:             plan,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		if err := s.store.Upsert(ctx, newSub); err != nil {
			return "", err
		}
	}

	return s.stripe.CreateCheckoutSession(customerID, plan, successURL, cancelURL)
}

// HandleWebhookEvent processes Stripe webhook events.
func (s *Service) HandleWebhookEvent(ctx context.Context, event *stripe.Event) error {
	switch event.Type {
	case "checkout.session.completed":
		var session stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
			return err
		}
		return s.activateSubscription(ctx, session.Customer.ID, session.Subscription.ID)

	case "customer.subscription.updated", "customer.subscription.deleted":
		var sub stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
			return err
		}
		return s.syncSubscription(ctx, sub.Customer.ID, &sub)
	case "invoice.payment_failed":
		var inv stripe.Invoice
		if err := json.Unmarshal(event.Data.Raw, &inv); err != nil {
			return err
		}
		if inv.Customer == nil {
			return errors.New("billing: invoice missing customer")
		}
		sub, err := s.store.GetByStripeCustomer(ctx, inv.Customer.ID)
		if err != nil {
			return err
		}
		sub.Status = StatusPastDue
		return s.store.Upsert(ctx, sub)
	}
	return nil
}

// GetSubscription retrieves the current subscription state for a tenant.
func (s *Service) GetSubscription(ctx context.Context, tenantID string) (*Subscription, error) {
	return s.store.GetByTenantID(ctx, tenantID)
}

// ParseWebhook parses and validates a Stripe webhook request.
func (s *Service) ParseWebhook(r *http.Request) (*stripe.Event, error) {
	return s.stripe.ParseWebhook(r)
}

// CreateBillingPortalSession creates a Stripe billing portal session for managing subscriptions.
func (s *Service) CreateBillingPortalSession(ctx context.Context, tenantID, returnURL string) (string, error) {
	sub, err := s.store.GetByTenantID(ctx, tenantID)
	if err != nil {
		return "", err
	}
	if sub == nil || sub.StripeCustomerID == "" {
		return "", errors.New("billing: no active subscription found")
	}
	return s.stripe.CreateBillingPortalSession(sub.StripeCustomerID, returnURL)
}

// HasActiveSubscription checks if a tenant has an active subscription.
func (s *Service) HasActiveSubscription(ctx context.Context, tenantID string) (bool, error) {
	sub, err := s.store.GetByTenantID(ctx, tenantID)
	if err != nil {
		return false, err
	}
	if sub == nil {
		return false, nil
	}
	return sub.IsActive(), nil
}

func (s *Service) activateSubscription(ctx context.Context, customerID, stripeSubID string) error {
	sub, err := s.store.GetByStripeCustomer(ctx, customerID)
	if err != nil {
		return err
	}
	sub.StripeSubscriptionID = stripeSubID
	sub.Status = StatusActive
	sub.UpdatedAt = time.Now()
	return s.store.Upsert(ctx, sub)
}

func (s *Service) syncSubscription(ctx context.Context, customerID string, stripeSub *stripe.Subscription) error {
	sub, err := s.store.GetByStripeCustomer(ctx, customerID)
	if err != nil {
		return err
	}
	if stripeSub.Status != "" {
		sub.Status = SubscriptionStatus(stripeSub.Status)
	}
	// Stripe SDK v82 uses Items.Data[0].CurrentPeriodEnd or similar
	// For now, use the subscription's overall period if available
	if stripeSub.Items != nil && len(stripeSub.Items.Data) > 0 {
		item := stripeSub.Items.Data[0]
		if item.CurrentPeriodEnd != 0 {
			t := time.Unix(item.CurrentPeriodEnd, 0)
			sub.CurrentPeriodEnd = &t
		}
	}
	sub.UpdatedAt = time.Now()
	return s.store.Upsert(ctx, sub)
}

// UpdateTenantStripeCustomer updates the Stripe customer ID for a tenant
func (s *Service) UpdateTenantStripeCustomer(ctx context.Context, tenantID, customerID string) error {
	sub, err := s.store.GetByTenantID(ctx, tenantID)
	if err != nil {
		return err
	}
	if sub == nil {
		sub = &Subscription{
			ID:        uuid.NewString(),
			TenantID:  tenantID,
			Status:    StatusTrialing,
			CreatedAt: time.Now(),
		}
	}
	sub.StripeCustomerID = customerID
	sub.UpdatedAt = time.Now()
	return s.store.Upsert(ctx, sub)
}

// UpdateTenantSubscription updates the subscription details for a tenant
func (s *Service) UpdateTenantSubscription(ctx context.Context, tenantID, subscriptionID, plan, status string) error {
	return s.UpdateTenantSubscriptionWithPeriod(ctx, tenantID, subscriptionID, plan, status, nil)
}

// UpdateTenantSubscriptionWithPeriod updates the subscription details for a tenant including the period end
func (s *Service) UpdateTenantSubscriptionWithPeriod(ctx context.Context, tenantID, subscriptionID, plan, status string, periodEnd *time.Time) error {
	sub, err := s.store.GetByTenantID(ctx, tenantID)
	if err != nil {
		return err
	}
	if sub == nil {
		sub = &Subscription{
			ID:        uuid.NewString(),
			TenantID:  tenantID,
			CreatedAt: time.Now(),
		}
	}
	sub.StripeSubscriptionID = subscriptionID
	sub.Plan = plan
	sub.Status = SubscriptionStatus(status)
	if periodEnd != nil {
		sub.CurrentPeriodEnd = periodEnd
	}
	sub.UpdatedAt = time.Now()
	return s.store.Upsert(ctx, sub)
}

// Ready reports whether the billing service has the minimum dependencies configured.
func (s *Service) Ready() error {
	if s == nil {
		return errors.New("billing: service is nil")
	}
	if s.stripe == nil {
		return errors.New("billing: stripe client missing")
	}
	if s.store == nil {
		return errors.New("billing: store not configured")
	}
	return nil
}

// StripeClientReady reports whether Stripe connectivity is configured.
func (s *Service) StripeClientReady() bool {
	return s != nil && s.stripe != nil
}

// ============================================
// Postgres store implementation
// ============================================

// PostgresStore implements Store using PostgreSQL.
type PostgresStore struct {
	db *sql.DB
}

// NewPostgresStore creates a new subscription store.
func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

// GetByTenantID returns the subscription for a tenant if exists.
func (s *PostgresStore) GetByTenantID(ctx context.Context, tenantID string) (*Subscription, error) {
	sub := &Subscription{}
	var periodEnd sql.NullTime
	err := s.db.QueryRowContext(ctx, `
        SELECT id, tenant_id, stripe_customer_id, stripe_subscription_id, status, plan, current_period_end, created_at, updated_at
        FROM subscriptions WHERE tenant_id = $1
    `, tenantID).Scan(&sub.ID, &sub.TenantID, &sub.StripeCustomerID, &sub.StripeSubscriptionID, &sub.Status, &sub.Plan, &periodEnd, &sub.CreatedAt, &sub.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if periodEnd.Valid {
		sub.CurrentPeriodEnd = &periodEnd.Time
	}
	return sub, nil
}

// GetByStripeCustomer returns the subscription with the given Stripe customer ID.
func (s *PostgresStore) GetByStripeCustomer(ctx context.Context, customerID string) (*Subscription, error) {
	sub := &Subscription{}
	var periodEnd sql.NullTime
	err := s.db.QueryRowContext(ctx, `
        SELECT id, tenant_id, stripe_customer_id, stripe_subscription_id, status, plan, current_period_end, created_at, updated_at
        FROM subscriptions WHERE stripe_customer_id = $1
    `, customerID).Scan(&sub.ID, &sub.TenantID, &sub.StripeCustomerID, &sub.StripeSubscriptionID, &sub.Status, &sub.Plan, &periodEnd, &sub.CreatedAt, &sub.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("billing: subscription not found")
	}
	if err != nil {
		return nil, err
	}
	if periodEnd.Valid {
		sub.CurrentPeriodEnd = &periodEnd.Time
	}
	return sub, nil
}

// Upsert inserts or updates a subscription row.
func (s *PostgresStore) Upsert(ctx context.Context, sub *Subscription) error {
	if sub.ID == "" {
		sub.ID = uuid.NewString()
	}
	_, err := s.db.ExecContext(ctx, `
        INSERT INTO subscriptions (id, tenant_id, stripe_customer_id, stripe_subscription_id, status, plan, current_period_end, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        ON CONFLICT (tenant_id) DO UPDATE SET
            stripe_customer_id = EXCLUDED.stripe_customer_id,
            stripe_subscription_id = EXCLUDED.stripe_subscription_id,
            status = EXCLUDED.status,
            plan = EXCLUDED.plan,
            current_period_end = EXCLUDED.current_period_end,
            updated_at = EXCLUDED.updated_at
    `, sub.ID, sub.TenantID, sub.StripeCustomerID, sub.StripeSubscriptionID, sub.Status, sub.Plan, sub.CurrentPeriodEnd, sub.CreatedAt, sub.UpdatedAt)
	return err
}
