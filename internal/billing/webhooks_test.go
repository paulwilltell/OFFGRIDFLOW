// Package billing provides webhook event tests.
package billing

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	"github.com/stripe/stripe-go/v82"
)

// TestWebhookSubscriptionCreated verifies subscription creation webhook handling.
func TestWebhookSubscriptionCreated(t *testing.T) {
	store := NewInMemoryStore()
	client := &StripeClient{
		priceFree:       "price_free",
		priceBasic:      "price_basic",
		pricePro:        "price_pro",
		priceEnterprise: "price_enterprise",
	}
	service := NewService(client, store)
	handler := NewWebhookHandler(client, service, slog.Default())

	ctx := context.Background()

	// Create a subscription event
	event := &stripe.Event{
		Type: "customer.subscription.created",
		Data: &stripe.EventData{
			Raw: mustMarshal(stripe.Subscription{
				ID:     "sub_123",
				Status: stripe.SubscriptionStatusActive,
				Customer: &stripe.Customer{
					ID: "cust_123",
					Metadata: map[string]string{
						"tenant_id": "tenant-123",
					},
				},
				Items: &stripe.SubscriptionItemList{
					Data: []*stripe.SubscriptionItem{
						{
							Price: &stripe.Price{
								ID: "price_basic",
							},
						},
					},
				},
				Metadata: map[string]string{
					"tenant_id": "tenant-123",
				},
			}),
		},
	}

	// Ensure tenant subscription exists first
	err := store.Upsert(ctx, &Subscription{
		ID:               "sub_old",
		TenantID:         "tenant-123",
		StripeCustomerID: "cust_123",
		Status:           StatusTrialing,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	})
	if err != nil {
		t.Fatalf("setup subscription failed: %v", err)
	}

	// Handle the webhook
	err = handler.handleSubscriptionCreated(ctx, event)
	if err != nil {
		t.Fatalf("webhook handler failed: %v", err)
	}

	// Verify subscription state updated
	sub, err := store.GetByTenantID(ctx, "tenant-123")
	if err != nil {
		t.Fatalf("get subscription failed: %v", err)
	}
	if sub == nil {
		t.Fatalf("subscription not found")
	}
	if sub.StripeSubscriptionID != "sub_123" {
		t.Fatalf("expected sub_123, got %s", sub.StripeSubscriptionID)
	}
	if sub.Status != StatusActive {
		t.Fatalf("expected active status, got %s", sub.Status)
	}
	if sub.Plan != "basic" {
		t.Fatalf("expected basic plan, got %s", sub.Plan)
	}

	t.Log("PASS: Subscription creation webhook handled correctly")
}

// TestWebhookSubscriptionUpdated verifies subscription update webhook handling.
func TestWebhookSubscriptionUpdated(t *testing.T) {
	store := NewInMemoryStore()
	client := &StripeClient{
		priceFree:       "price_free",
		priceBasic:      "price_basic",
		pricePro:        "price_pro",
		priceEnterprise: "price_enterprise",
	}
	service := NewService(client, store)
	handler := NewWebhookHandler(client, service, slog.Default())

	ctx := context.Background()

	// Create initial subscription
	err := store.Upsert(ctx, &Subscription{
		ID:                   "sub_1",
		TenantID:             "tenant-456",
		StripeCustomerID:     "cust_456",
		StripeSubscriptionID: "sub_456",
		Status:               StatusActive,
		Plan:                 "basic",
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	})
	if err != nil {
		t.Fatalf("setup subscription failed: %v", err)
	}

	// Create subscription update event (status change)
	periodEnd := time.Now().AddDate(0, 1, 0).Unix()
	event := &stripe.Event{
		Type: "customer.subscription.updated",
		Data: &stripe.EventData{
			Raw: mustMarshal(stripe.Subscription{
				ID:     "sub_456",
				Status: stripe.SubscriptionStatusPastDue,
				Customer: &stripe.Customer{
					ID: "cust_456",
					Metadata: map[string]string{
						"tenant_id": "tenant-456",
					},
				},
				Items: &stripe.SubscriptionItemList{
					Data: []*stripe.SubscriptionItem{
						{
							CurrentPeriodEnd: periodEnd,
							Price: &stripe.Price{
								ID: "price_basic",
							},
						},
					},
				},
				Metadata: map[string]string{
					"tenant_id": "tenant-456",
				},
			}),
		},
	}

	// Handle the webhook
	err = handler.handleSubscriptionUpdated(ctx, event)
	if err != nil {
		t.Fatalf("webhook handler failed: %v", err)
	}

	// Verify subscription state updated
	sub, err := store.GetByTenantID(ctx, "tenant-456")
	if err != nil {
		t.Fatalf("get subscription failed: %v", err)
	}
	if sub == nil {
		t.Fatalf("subscription not found")
	}
	if sub.Status != StatusPastDue {
		t.Fatalf("expected past_due status, got %s", sub.Status)
	}
	if sub.CurrentPeriodEnd == nil {
		t.Fatalf("current period end should be set")
	}

	t.Log("PASS: Subscription update webhook handled correctly")
}

// TestWebhookSubscriptionDeleted verifies subscription deletion webhook handling.
func TestWebhookSubscriptionDeleted(t *testing.T) {
	store := NewInMemoryStore()
	client := &StripeClient{}
	service := NewService(client, store)
	handler := NewWebhookHandler(client, service, slog.Default())

	ctx := context.Background()

	// Create initial subscription
	err := store.Upsert(ctx, &Subscription{
		ID:                   "sub_2",
		TenantID:             "tenant-789",
		StripeCustomerID:     "cust_789",
		StripeSubscriptionID: "sub_789",
		Status:               StatusActive,
		Plan:                 "pro",
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	})
	if err != nil {
		t.Fatalf("setup subscription failed: %v", err)
	}

	// Create subscription deleted event
	event := &stripe.Event{
		Type: "customer.subscription.deleted",
		Data: &stripe.EventData{
			Raw: mustMarshal(stripe.Subscription{
				ID:     "sub_789",
				Status: stripe.SubscriptionStatusCanceled,
				Customer: &stripe.Customer{
					ID: "cust_789",
					Metadata: map[string]string{
						"tenant_id": "tenant-789",
					},
				},
				Metadata: map[string]string{
					"tenant_id": "tenant-789",
				},
			}),
		},
	}

	// Handle the webhook
	err = handler.handleSubscriptionDeleted(ctx, event)
	if err != nil {
		t.Fatalf("webhook handler failed: %v", err)
	}

	// Verify subscription downgraded to free
	sub, err := store.GetByTenantID(ctx, "tenant-789")
	if err != nil {
		t.Fatalf("get subscription failed: %v", err)
	}
	if sub == nil {
		t.Fatalf("subscription not found")
	}
	if sub.Status != StatusCanceled {
		t.Fatalf("expected canceled status, got %s", sub.Status)
	}
	if sub.Plan != "free" {
		t.Fatalf("expected free plan, got %s", sub.Plan)
	}

	t.Log("PASS: Subscription deletion webhook handled correctly")
}

// TestWebhookInvoicePaid verifies invoice payment webhook handling.
func TestWebhookInvoicePaid(t *testing.T) {
	store := NewInMemoryStore()
	client := &StripeClient{}
	service := NewService(client, store)
	handler := NewWebhookHandler(client, service, slog.Default())

	ctx := context.Background()

	// Create initial subscription
	err := store.Upsert(ctx, &Subscription{
		ID:               "sub_3",
		TenantID:         "tenant-paid",
		StripeCustomerID: "cust_paid",
		Status:           StatusPastDue,
		Plan:             "basic",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	})
	if err != nil {
		t.Fatalf("setup subscription failed: %v", err)
	}

	// Create invoice paid event
	event := &stripe.Event{
		Type: "invoice.paid",
		Data: &stripe.EventData{
			Raw: mustMarshal(stripe.Invoice{
				ID:            "inv_123",
				Customer:      &stripe.Customer{ID: "cust_paid"},
				CustomerEmail: "user@example.com",
				AmountPaid:    9900, // $99.00
				AmountDue:     9900,
				Currency:      "usd",
			}),
		},
	}

	// Handle the webhook (should not error)
	err = handler.handleInvoicePaid(ctx, event)
	if err != nil {
		t.Fatalf("webhook handler failed: %v", err)
	}

	t.Log("PASS: Invoice paid webhook handled correctly")
}

// TestWebhookInvoicePaymentFailed verifies payment failure webhook handling.
func TestWebhookInvoicePaymentFailed(t *testing.T) {
	store := NewInMemoryStore()
	client := &StripeClient{}
	service := NewService(client, store)
	handler := NewWebhookHandler(client, service, slog.Default())

	ctx := context.Background()

	// Create initial subscription
	err := store.Upsert(ctx, &Subscription{
		ID:               "sub_4",
		TenantID:         "tenant-failed",
		StripeCustomerID: "cust_failed",
		Status:           StatusActive,
		Plan:             "pro",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	})
	if err != nil {
		t.Fatalf("setup subscription failed: %v", err)
	}

	// Create invoice payment failed event
	event := &stripe.Event{
		Type: "invoice.payment_failed",
		Data: &stripe.EventData{
			Raw: mustMarshal(stripe.Invoice{
				ID:              "inv_fail",
				Customer:        &stripe.Customer{ID: "cust_failed"},
				CustomerEmail:   "user@example.com",
				AmountDue:       9900,
				AmountRemaining: 9900,
				Currency:        "usd",
			}),
		},
	}

	// Handle the webhook
	err = handler.handleInvoicePaymentFailed(ctx, event)
	if err != nil {
		t.Fatalf("webhook handler failed: %v", err)
	}

	// Verify subscription marked as past due
	sub, err := store.GetByTenantID(ctx, "tenant-failed")
	if err != nil {
		t.Fatalf("get subscription failed: %v", err)
	}
	if sub == nil {
		t.Fatalf("subscription not found")
	}
	if sub.Status != StatusPastDue {
		t.Fatalf("expected past_due status, got %s", sub.Status)
	}

	t.Log("PASS: Invoice payment failed webhook handled correctly")
}

// TestWebhookCustomerCreated verifies customer creation webhook handling.
func TestWebhookCustomerCreated(t *testing.T) {
	store := NewInMemoryStore()
	client := &StripeClient{}
	service := NewService(client, store)
	handler := NewWebhookHandler(client, service, slog.Default())

	ctx := context.Background()

	// Create customer created event
	event := &stripe.Event{
		Type: "customer.created",
		Data: &stripe.EventData{
			Raw: mustMarshal(stripe.Customer{
				ID:    "cust_new",
				Email: "new@example.com",
				Name:  "New Customer",
				Metadata: map[string]string{
					"tenant_id": "tenant-new",
				},
			}),
		},
	}

	// Handle the webhook
	err := handler.handleCustomerCreated(ctx, event)
	if err != nil {
		t.Fatalf("webhook handler failed: %v", err)
	}

	// Verify subscription created
	sub, err := store.GetByTenantID(ctx, "tenant-new")
	if err != nil {
		t.Fatalf("get subscription failed: %v", err)
	}
	if sub == nil {
		t.Fatalf("subscription not found")
	}
	if sub.StripeCustomerID != "cust_new" {
		t.Fatalf("expected cust_new, got %s", sub.StripeCustomerID)
	}

	t.Log("PASS: Customer creation webhook handled correctly")
}

// TestWebhookErrorHandling verifies error handling for malformed events.
func TestWebhookErrorHandling(t *testing.T) {
	store := NewInMemoryStore()
	client := &StripeClient{}
	service := NewService(client, store)
	handler := NewWebhookHandler(client, service, slog.Default())

	ctx := context.Background()

	// Create event with invalid data
	event := &stripe.Event{
		Type: "customer.subscription.created",
		Data: &stripe.EventData{
			Raw: []byte(`{"invalid": "data"}`),
		},
	}

	// Handle the webhook (should error on missing fields)
	err := handler.handleSubscriptionCreated(ctx, event)
	if err == nil {
		t.Fatalf("expected error for malformed event, got nil")
	}

	t.Log("PASS: Error handling verified")
}

// TestSubscriptionStateTransitions verifies correct state transitions.
func TestSubscriptionStateTransitions(t *testing.T) {
	store := NewInMemoryStore()

	ctx := context.Background()

	sub := &Subscription{
		ID:        "sub_trans",
		TenantID:  "tenant-trans",
		Status:    StatusTrialing,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Create subscription
	err := store.Upsert(ctx, sub)
	if err != nil {
		t.Fatalf("upsert failed: %v", err)
	}

	// Transition to active
	sub.Status = StatusActive
	err = store.Upsert(ctx, sub)
	if err != nil {
		t.Fatalf("upsert failed: %v", err)
	}

	retrieved, _ := store.GetByTenantID(ctx, "tenant-trans")
	if retrieved.Status != StatusActive {
		t.Fatalf("expected active, got %s", retrieved.Status)
	}

	// Transition to past due
	sub.Status = StatusPastDue
	err = store.Upsert(ctx, sub)
	if err != nil {
		t.Fatalf("upsert failed: %v", err)
	}

	retrieved, _ = store.GetByTenantID(ctx, "tenant-trans")
	if retrieved.Status != StatusPastDue {
		t.Fatalf("expected past_due, got %s", retrieved.Status)
	}

	// Transition to canceled
	sub.Status = StatusCanceled
	err = store.Upsert(ctx, sub)
	if err != nil {
		t.Fatalf("upsert failed: %v", err)
	}

	retrieved, _ = store.GetByTenantID(ctx, "tenant-trans")
	if retrieved.Status != StatusCanceled {
		t.Fatalf("expected canceled, got %s", retrieved.Status)
	}

	t.Log("PASS: State transitions verified")
}

// Helper function to marshal Stripe objects
func mustMarshal(obj interface{}) []byte {
	data, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}
	return data
}
