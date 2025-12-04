package billing

import (
	"context"
	"testing"
	"time"

	"github.com/stripe/stripe-go/v82"
)

func TestServiceReadyValidation(t *testing.T) {
	svc := &Service{}
	if err := svc.Ready(); err == nil {
		t.Fatalf("expected error when dependencies missing")
	}

	svc = &Service{stripe: &StripeClient{}}
	if err := svc.Ready(); err == nil {
		t.Fatalf("expected error when store missing")
	}

	svc.store = NewInMemoryStore()
	if err := svc.Ready(); err != nil {
		t.Fatalf("expected ready once deps set, got %v", err)
	}
}

func TestHasActiveSubscription(t *testing.T) {
	store := NewInMemoryStore()
	svc := NewService(&StripeClient{}, store)

	active, err := svc.HasActiveSubscription(context.Background(), "tenant-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if active {
		t.Fatalf("expected no subscription to be inactive")
	}

	now := time.Now()
	_ = store.Upsert(context.Background(), &Subscription{
		ID:               "sub-1",
		TenantID:         "tenant-1",
		StripeCustomerID: "cust-1",
		Status:           StatusActive,
		CurrentPeriodEnd: &now,
	})

	active, err = svc.HasActiveSubscription(context.Background(), "tenant-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !active {
		t.Fatalf("expected subscription to be active")
	}
}

func TestHandleWebhookEvent_NoOpOnUnknown(t *testing.T) {
	svc := NewService(&StripeClient{}, NewInMemoryStore())
	event := &stripe.Event{Type: "unhandled.type"}
	if err := svc.HandleWebhookEvent(context.Background(), event); err != nil {
		t.Fatalf("expected no error for unknown events, got %v", err)
	}
}
