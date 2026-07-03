package billing

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"time"
)

// ReportPurchase records a one-time payment that unlocks report exports for a tenant.
type ReportPurchase struct {
	ID              string
	TenantID        string
	StripeSessionID string
	AmountCents     int64
	Currency        string
	Status          string
	CreatedAt       time.Time
}

// ReportPurchaseStore persists report entitlements (pay-per-report model).
type ReportPurchaseStore interface {
	// Record grants a tenant the report entitlement. Idempotent by stripe_session_id.
	Record(ctx context.Context, p *ReportPurchase) error
	// HasPaid reports whether the tenant has at least one paid report purchase.
	HasPaid(ctx context.Context, tenantID string) (bool, error)
}

// -----------------------------------------------------------------------------
// Postgres implementation
// -----------------------------------------------------------------------------

// PostgresReportStore persists report purchases in Postgres.
type PostgresReportStore struct {
	db *sql.DB
}

// NewPostgresReportStore constructs a Postgres-backed report purchase store.
func NewPostgresReportStore(db *sql.DB) *PostgresReportStore {
	return &PostgresReportStore{db: db}
}

// Record inserts a report purchase, ignoring duplicates by stripe_session_id.
func (s *PostgresReportStore) Record(ctx context.Context, p *ReportPurchase) error {
	if p == nil || p.TenantID == "" || p.StripeSessionID == "" {
		return errors.New("billing: tenant ID and stripe session ID required")
	}
	amount := p.AmountCents
	if amount == 0 {
		amount = ReportPriceCents
	}
	currency := p.Currency
	if currency == "" {
		currency = "usd"
	}
	status := p.Status
	if status == "" {
		status = "paid"
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO report_purchases (tenant_id, stripe_session_id, amount_cents, currency, status)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (stripe_session_id) DO NOTHING
	`, p.TenantID, p.StripeSessionID, amount, currency, status)
	return err
}

// HasPaid returns true if the tenant has any paid report purchase.
func (s *PostgresReportStore) HasPaid(ctx context.Context, tenantID string) (bool, error) {
	if tenantID == "" {
		return false, errors.New("billing: tenant ID required")
	}
	var exists bool
	err := s.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM report_purchases
			WHERE tenant_id = $1 AND status = 'paid'
		)
	`, tenantID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// -----------------------------------------------------------------------------
// In-memory implementation
// -----------------------------------------------------------------------------

// InMemoryReportStore is a thread-safe report store for tests/dev.
type InMemoryReportStore struct {
	mu       sync.RWMutex
	sessions map[string]bool   // stripe_session_id -> recorded
	tenants  map[string]bool   // tenant_id -> has paid
}

// NewInMemoryReportStore creates an in-memory report purchase store.
func NewInMemoryReportStore() *InMemoryReportStore {
	return &InMemoryReportStore{
		sessions: make(map[string]bool),
		tenants:  make(map[string]bool),
	}
}

// Record grants the tenant entitlement (idempotent by session ID).
func (s *InMemoryReportStore) Record(ctx context.Context, p *ReportPurchase) error {
	if p == nil || p.TenantID == "" || p.StripeSessionID == "" {
		return errors.New("billing: tenant ID and stripe session ID required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.sessions[p.StripeSessionID] {
		return nil
	}
	s.sessions[p.StripeSessionID] = true
	s.tenants[p.TenantID] = true
	return nil
}

// HasPaid reports whether the tenant has paid.
func (s *InMemoryReportStore) HasPaid(ctx context.Context, tenantID string) (bool, error) {
	if tenantID == "" {
		return false, errors.New("billing: tenant ID required")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.tenants[tenantID], nil
}
