package billing

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
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
	stripe      *StripeClient
	store       Store
	reportStore ReportPurchaseStore
}

// NewService constructs a billing service.
func NewService(stripeClient *StripeClient, store Store) *Service {
	return &Service{stripe: stripeClient, store: store}
}

// SetReportStore wires the pay-per-report entitlement store. Optional; when
// unset, report-payment checks fail closed (treated as unpaid).
func (s *Service) SetReportStore(rs ReportPurchaseStore) {
	if s != nil {
		s.reportStore = rs
	}
}

// StartReportCheckout creates a one-time $149 checkout to unlock report exports.
// Reuses the tenant's Stripe customer (creating one if needed).
func (s *Service) StartReportCheckout(ctx context.Context, tenantID, tenantName, email, successURL, cancelURL string) (string, error) {
	if s == nil || s.stripe == nil || s.store == nil {
		return "", errors.New("billing: service not configured")
	}

	sub, err := s.store.GetByTenantID(ctx, tenantID)
	if err != nil {
		return "", err
	}

	var customerID string
	if sub != nil && sub.StripeCustomerID != "" {
		customerID = sub.StripeCustomerID
	} else {
		customerID, err = s.stripe.CreateCustomer(email, tenantName, tenantID)
		if err != nil {
			return "", err
		}
		now := time.Now()
		if sub == nil {
			sub = &Subscription{ID: uuid.NewString(), TenantID: tenantID, CreatedAt: now}
		}
		sub.StripeCustomerID = customerID
		sub.UpdatedAt = now
		if err := s.store.Upsert(ctx, sub); err != nil {
			return "", err
		}
	}

	return s.stripe.CreateReportCheckoutSession(customerID, tenantID, successURL, cancelURL)
}

// HasPaidForReport reports whether the tenant has unlocked report exports.
// Fails closed: any error or missing store returns false.
func (s *Service) HasPaidForReport(ctx context.Context, tenantID string) bool {
	if s == nil || s.reportStore == nil || tenantID == "" {
		return false
	}
	paid, err := s.reportStore.HasPaid(ctx, tenantID)
	if err != nil {
		return false
	}
	return paid
}

// RecordReportPurchase grants the report entitlement for a tenant.
func (s *Service) RecordReportPurchase(ctx context.Context, tenantID, stripeSessionID string) error {
	if s == nil || s.reportStore == nil {
		return errors.New("billing: report store not configured")
	}
	return s.reportStore.Record(ctx, &ReportPurchase{
		TenantID:        tenantID,
		StripeSessionID: stripeSessionID,
		AmountCents:     ReportPriceCents,
		Currency:        "usd",
		Status:          "paid",
	})
}

// StartSubscription initiates checkout and returns the Stripe-hosted URL.
func (s *Service) StartSubscription(ctx context.Context, tenantID, tenantName, email, plan, successURL, cancelURL string) (string, error) {
	if s == nil || s.stripe == nil || s.store == nil {
		return "", errors.New("billing: service not configured")
	}

	// Check if subscription already exists (and hence customer).
	sub, err := s.store.GetByTenantID(ctx, tenantID)
	if err != nil {
		return "", err
	}

	var customerID string
	if sub != nil && sub.StripeCustomerID != "" {
		customerID = sub.StripeCustomerID
	} else {
		var createErr error
		customerID, createErr = s.stripe.CreateCustomer(email, tenantName, tenantID)
		if createErr != nil {
			return "", createErr
		}
		// Persist customer link. A checkout session is not an active subscription yet.
		now := time.Now()
		if sub == nil {
			sub = &Subscription{
				ID:        uuid.NewString(),
				TenantID:  tenantID,
				CreatedAt: now,
			}
		}
		sub.StripeCustomerID = customerID
		sub.Status = StatusUnpaid
		sub.Plan = plan
		sub.UpdatedAt = now
		if err := s.store.Upsert(ctx, sub); err != nil {
			return "", err
		}
	}

	// Keep selected plan persisted for non-active subscriptions so webhook activation
	// completes with the customer's chosen tier.
	if sub != nil && !sub.IsActive() && sub.Plan != plan {
		sub.Plan = plan
		sub.Status = StatusUnpaid
		sub.UpdatedAt = time.Now()
		if err := s.store.Upsert(ctx, sub); err != nil {
			return "", err
		}
	}

	return s.stripe.CreateCheckoutSession(customerID, plan, successURL, cancelURL)
}

// HandleWebhookEvent processes Stripe webhook events.
func (s *Service) HandleWebhookEvent(ctx context.Context, event *stripe.Event) error {
	if s == nil || s.store == nil {
		return errors.New("billing: service not configured")
	}
	if event == nil {
		return errors.New("billing: webhook event is nil")
	}

	switch event.Type {
	case "checkout.session.completed":
		var session stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
			return err
		}
		if session.Customer == nil || session.Customer.ID == "" {
			return errors.New("billing: checkout session missing customer id")
		}
		subscriptionID := ""
		if session.Subscription != nil {
			subscriptionID = session.Subscription.ID
		}
		return s.activateSubscription(ctx, session.Customer.ID, subscriptionID)

	case "customer.subscription.updated", "customer.subscription.deleted":
		var sub stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
			return err
		}
		if sub.Customer == nil || sub.Customer.ID == "" {
			return errors.New("billing: subscription event missing customer id")
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
		if sub == nil {
			return fmt.Errorf("billing: no subscription found for customer %s", inv.Customer.ID)
		}
		sub.Status = StatusPastDue
		sub.UpdatedAt = time.Now()
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
	if s == nil || s.stripe == nil {
		return nil, errors.New("billing: stripe client not configured")
	}
	return s.stripe.ParseWebhook(r)
}

// CreateBillingPortalSession creates a Stripe billing portal session for managing subscriptions.
func (s *Service) CreateBillingPortalSession(ctx context.Context, tenantID, returnURL string) (string, error) {
	if s == nil || s.stripe == nil || s.store == nil {
		return "", errors.New("billing: service not configured")
	}

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
	if s == nil || s.store == nil {
		return false, errors.New("billing: service not configured")
	}

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
	if sub == nil {
		return fmt.Errorf("billing: no subscription found for customer %s", customerID)
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
	if sub == nil {
		return fmt.Errorf("billing: no subscription found for customer %s", customerID)
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
		if item.Price != nil && item.Price.ID != "" && s.stripe != nil {
			if plan, ok := s.stripe.PlanFromPriceID(item.Price.ID); ok && plan != "" {
				sub.Plan = string(plan)
			}
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
        ORDER BY updated_at DESC, created_at DESC
        LIMIT 1
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
// Returns nil, nil if no subscription is found (not an error).
func (s *PostgresStore) GetByStripeCustomer(ctx context.Context, customerID string) (*Subscription, error) {
	sub := &Subscription{}
	var periodEnd sql.NullTime
	err := s.db.QueryRowContext(ctx, `
        SELECT id, tenant_id, stripe_customer_id, stripe_subscription_id, status, plan, current_period_end, created_at, updated_at
        FROM subscriptions WHERE stripe_customer_id = $1
        ORDER BY updated_at DESC, created_at DESC
        LIMIT 1
    `, customerID).Scan(&sub.ID, &sub.TenantID, &sub.StripeCustomerID, &sub.StripeSubscriptionID, &sub.Status, &sub.Plan, &periodEnd, &sub.CreatedAt, &sub.UpdatedAt)
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

// Upsert inserts or updates a subscription row.
func (s *PostgresStore) Upsert(ctx context.Context, sub *Subscription) error {
	if sub.ID == "" {
		sub.ID = uuid.NewString()
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
        UPDATE subscriptions
        SET stripe_customer_id = $2,
            stripe_subscription_id = $3,
            status = $4,
            plan = $5,
            current_period_end = $6,
            updated_at = $7
        WHERE tenant_id = $1
    `, sub.TenantID, sub.StripeCustomerID, sub.StripeSubscriptionID, sub.Status, sub.Plan, sub.CurrentPeriodEnd, sub.UpdatedAt)
	if err != nil {
		return err
	}

	rowsUpdated, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsUpdated == 0 {
		_, err = tx.ExecContext(ctx, `
            INSERT INTO subscriptions (id, tenant_id, stripe_customer_id, stripe_subscription_id, status, plan, current_period_end, created_at, updated_at)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        `, sub.ID, sub.TenantID, sub.StripeCustomerID, sub.StripeSubscriptionID, sub.Status, sub.Plan, sub.CurrentPeriodEnd, sub.CreatedAt, sub.UpdatedAt)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
