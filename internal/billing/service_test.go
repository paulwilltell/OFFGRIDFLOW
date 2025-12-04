package billing

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stripe/stripe-go/v82"
)

func TestHandleInvoicePaymentFailedMarksPastDue(t *testing.T) {
	store := NewInMemoryStore()
	sub := &Subscription{
		ID:               "sub1",
		TenantID:         "tenant-1",
		StripeCustomerID: "cust1",
		Status:           StatusActive,
	}
	_ = store.Upsert(context.Background(), sub)
	svc := NewService(nil, store)

	inv := stripe.Invoice{
		Customer: &stripe.Customer{ID: "cust1"},
	}
	raw, _ := json.Marshal(inv)
	event := &stripe.Event{
		Type: "invoice.payment_failed",
		Data: &stripe.EventData{Raw: raw},
	}

	if err := svc.HandleWebhookEvent(context.Background(), event); err != nil {
		t.Fatalf("webhook handler error: %v", err)
	}

	updated, err := store.GetByTenantID(context.Background(), "tenant-1")
	if err != nil {
		t.Fatalf("get subscription: %v", err)
	}
	if updated.Status != StatusPastDue {
		t.Fatalf("expected past_due, got %s", updated.Status)
	}
}
