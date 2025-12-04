// Package billing provides integration tests for the billing system.
package billing

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v82"
)

// TestFullBillingWorkflow verifies the complete subscription lifecycle.
//
// Flow: Customer Creation → Subscription Creation → Invoice Paid → Subscription Updated → Cancellation
func TestFullBillingWorkflow(t *testing.T) {
	store := NewInMemoryStore()
	client := &StripeClient{
		priceFree:       "price_free",
		priceBasic:      "price_basic",
		pricePro:        "price_pro",
		priceEnterprise: "price_enterprise",
	}
	service := NewService(client, store)
	handler := NewWebhookHandler(client, service, nil)

	ctx := context.Background()
	tenantID := "tenant-" + uuid.NewString()[:8]
	customerID := "cust_" + uuid.NewString()[:8]
	subscriptionID := "sub_" + uuid.NewString()[:8]

	// STEP 1: Customer created
	customerEvent := &stripe.Event{
		Type: "customer.created",
		Data: &stripe.EventData{
			Raw: mustMarshal(stripe.Customer{
				ID:    customerID,
				Email: "user@example.com",
				Name:  "Test User",
				Metadata: map[string]string{
					"tenant_id": tenantID,
				},
			}),
		},
	}

	err := handler.handleCustomerCreated(ctx, customerEvent)
	if err != nil {
		t.Fatalf("Step 1 (customer created): %v", err)
	}

	// Verify customer subscription created
	sub, _ := store.GetByTenantID(ctx, tenantID)
	if sub == nil {
		t.Fatalf("Step 1: subscription not created")
	}
	if sub.StripeCustomerID != customerID {
		t.Fatalf("Step 1: customer ID mismatch")
	}

	// STEP 2: Subscription created
	subscriptionEvent := &stripe.Event{
		Type: "customer.subscription.created",
		Data: &stripe.EventData{
			Raw: mustMarshal(stripe.Subscription{
				ID:     subscriptionID,
				Status: stripe.SubscriptionStatusActive,
				Customer: &stripe.Customer{
					ID: customerID,
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
					"tenant_id": tenantID,
				},
			}),
		},
	}

	err = handler.handleSubscriptionCreated(ctx, subscriptionEvent)
	if err != nil {
		t.Fatalf("Step 2 (subscription created): %v", err)
	}

	// Verify subscription active
	sub, _ = store.GetByTenantID(ctx, tenantID)
	if sub.Status != StatusActive {
		t.Fatalf("Step 2: expected active status, got %s", sub.Status)
	}
	if sub.Plan != "basic" {
		t.Fatalf("Step 2: expected basic plan, got %s", sub.Plan)
	}

	// STEP 3: Invoice paid (renewal)
	invoiceEvent := &stripe.Event{
		Type: "invoice.paid",
		Data: &stripe.EventData{
			Raw: mustMarshal(stripe.Invoice{
				ID:            "inv_" + uuid.NewString()[:8],
				Customer:      &stripe.Customer{ID: customerID},
				CustomerEmail: "user@example.com",
				AmountPaid:    9900,
				AmountDue:     9900,
				Currency:      "usd",
			}),
		},
	}

	err = handler.handleInvoicePaid(ctx, invoiceEvent)
	if err != nil {
		t.Fatalf("Step 3 (invoice paid): %v", err)
	}

	// STEP 4: Subscription updated with new period end
	periodEnd := time.Now().AddDate(0, 1, 0).Unix()
	updateEvent := &stripe.Event{
		Type: "customer.subscription.updated",
		Data: &stripe.EventData{
			Raw: mustMarshal(stripe.Subscription{
				ID:     subscriptionID,
				Status: stripe.SubscriptionStatusActive,
				Customer: &stripe.Customer{
					ID: customerID,
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
					"tenant_id": tenantID,
				},
			}),
		},
	}

	err = handler.handleSubscriptionUpdated(ctx, updateEvent)
	if err != nil {
		t.Fatalf("Step 4 (subscription updated): %v", err)
	}

	// Verify period end set
	sub, _ = store.GetByTenantID(ctx, tenantID)
	if sub.CurrentPeriodEnd == nil {
		t.Fatalf("Step 4: current period end not set")
	}

	// STEP 5: Subscription canceled
	cancelEvent := &stripe.Event{
		Type: "customer.subscription.deleted",
		Data: &stripe.EventData{
			Raw: mustMarshal(stripe.Subscription{
				ID:     subscriptionID,
				Status: stripe.SubscriptionStatusCanceled,
				Customer: &stripe.Customer{
					ID: customerID,
				},
				Metadata: map[string]string{
					"tenant_id": tenantID,
				},
			}),
		},
	}

	err = handler.handleSubscriptionDeleted(ctx, cancelEvent)
	if err != nil {
		t.Fatalf("Step 5 (subscription deleted): %v", err)
	}

	// Verify subscription downgraded to free
	sub, _ = store.GetByTenantID(ctx, tenantID)
	if sub.Plan != "free" {
		t.Fatalf("Step 5: expected free plan, got %s", sub.Plan)
	}
	if sub.Status != StatusCanceled {
		t.Fatalf("Step 5: expected canceled status, got %s", sub.Status)
	}

	t.Logf("✓ Full billing workflow completed successfully (%s)", tenantID)
}

// TestPaymentFailureRecovery verifies recovery from payment failures.
func TestPaymentFailureRecovery(t *testing.T) {
	store := NewInMemoryStore()
	client := &StripeClient{}
	service := NewService(client, store)
	handler := NewWebhookHandler(client, service, nil)

	ctx := context.Background()
	tenantID := "tenant-recovery"
	customerID := "cust_recovery"

	// Create active subscription
	err := store.Upsert(ctx, &Subscription{
		ID:                   uuid.NewString(),
		TenantID:             tenantID,
		StripeCustomerID:     customerID,
		StripeSubscriptionID: "sub_recovery",
		Status:               StatusActive,
		Plan:                 "pro",
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	})
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	// STEP 1: Payment fails
	failEvent := &stripe.Event{
		Type: "invoice.payment_failed",
		Data: &stripe.EventData{
			Raw: mustMarshal(stripe.Invoice{
				ID:        "inv_fail",
				Customer:  &stripe.Customer{ID: customerID},
				AmountDue: 9900,
			}),
		},
	}

	err = handler.handleInvoicePaymentFailed(ctx, failEvent)
	if err != nil {
		t.Fatalf("payment failed handler: %v", err)
	}

	// Verify marked as past due
	sub, _ := store.GetByTenantID(ctx, tenantID)
	if sub.Status != StatusPastDue {
		t.Fatalf("Step 1: expected past_due, got %s", sub.Status)
	}

	// STEP 2: Customer retries and payment succeeds
	successEvent := &stripe.Event{
		Type: "invoice.paid",
		Data: &stripe.EventData{
			Raw: mustMarshal(stripe.Invoice{
				ID:         "inv_success",
				Customer:   &stripe.Customer{ID: customerID},
				AmountPaid: 9900,
				Currency:   "usd",
			}),
		},
	}

	err = handler.handleInvoicePaid(ctx, successEvent)
	if err != nil {
		t.Fatalf("invoice paid handler: %v", err)
	}

	// Status should remain or be updated to active (via subscription.updated)
	sub, _ = store.GetByTenantID(ctx, tenantID)
	if sub == nil {
		t.Fatalf("Step 2: subscription lost")
	}

	t.Logf("✓ Payment failure recovery workflow completed (%s)", tenantID)
}

// TestPlanUpgrade verifies subscription upgrade flow.
func TestPlanUpgrade(t *testing.T) {
	store := NewInMemoryStore()
	client := &StripeClient{
		priceBasic: "price_basic",
		pricePro:   "price_pro",
	}
	service := NewService(client, store)
	handler := NewWebhookHandler(client, service, nil)

	ctx := context.Background()
	tenantID := "tenant-upgrade"
	customerID := "cust_upgrade"
	subscriptionID := "sub_upgrade"

	// Create subscription on basic plan
	err := store.Upsert(ctx, &Subscription{
		ID:                   uuid.NewString(),
		TenantID:             tenantID,
		StripeCustomerID:     customerID,
		StripeSubscriptionID: subscriptionID,
		Status:               StatusActive,
		Plan:                 "basic",
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	})
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	// Subscription updated with pro plan price
	upgradeEvent := &stripe.Event{
		Type: "customer.subscription.updated",
		Data: &stripe.EventData{
			Raw: mustMarshal(stripe.Subscription{
				ID:     subscriptionID,
				Status: stripe.SubscriptionStatusActive,
				Customer: &stripe.Customer{
					ID: customerID,
				},
				Items: &stripe.SubscriptionItemList{
					Data: []*stripe.SubscriptionItem{
						{
							Price: &stripe.Price{
								ID: "price_pro",
							},
						},
					},
				},
				Metadata: map[string]string{
					"tenant_id": tenantID,
				},
			}),
		},
	}

	err = handler.handleSubscriptionUpdated(ctx, upgradeEvent)
	if err != nil {
		t.Fatalf("upgrade handler: %v", err)
	}

	// Verify plan upgraded
	sub, _ := store.GetByTenantID(ctx, tenantID)
	if sub.Plan != "pro" {
		t.Fatalf("expected pro plan, got %s", sub.Plan)
	}
	if sub.Status != StatusActive {
		t.Fatalf("expected active status, got %s", sub.Status)
	}

	t.Logf("✓ Plan upgrade workflow completed (%s)", tenantID)
}

// TestConcurrentWebhookProcessing verifies thread-safety under concurrent webhook events.
func TestConcurrentWebhookProcessing(t *testing.T) {
	store := NewInMemoryStore()
	client := &StripeClient{}
	handler := NewWebhookHandler(client, NewService(client, store), nil)

	ctx := context.Background()

	// Process 10 parallel webhook events
	numWorkers := 10
	done := make(chan error, numWorkers)

	for i := 0; i < numWorkers; i++ {
		go func(index int) {
			tenantID := "tenant-concurrent-" + string(rune(index))
			customerID := "cust_" + string(rune(index))

			// Customer created event
			event := &stripe.Event{
				Type: "customer.created",
				Data: &stripe.EventData{
					Raw: mustMarshal(stripe.Customer{
						ID: customerID,
						Metadata: map[string]string{
							"tenant_id": tenantID,
						},
					}),
				},
			}

			done <- handler.handleCustomerCreated(ctx, event)
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numWorkers; i++ {
		if err := <-done; err != nil {
			t.Fatalf("concurrent event %d: %v", i, err)
		}
	}

	// Verify all subscriptions were created
	count := store.Count()
	if count != numWorkers {
		t.Fatalf("expected %d subscriptions, got %d", numWorkers, count)
	}

	t.Logf("✓ Concurrent webhook processing verified (%d events)", numWorkers)
}

// TestStoreIsolation verifies in-memory store isolation between subscriptions.
func TestStoreIsolation(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	// Create 3 different subscriptions
	subs := make([]*Subscription, 3)
	for i := 0; i < 3; i++ {
		subs[i] = &Subscription{
			ID:               uuid.NewString(),
			TenantID:         "tenant-" + string(rune(48+i)),
			StripeCustomerID: "cust_" + string(rune(48+i)),
			Status:           StatusActive,
			Plan:             "basic",
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		err := store.Upsert(ctx, subs[i])
		if err != nil {
			t.Fatalf("upsert %d: %v", i, err)
		}
	}

	// Verify each retrieval is isolated
	for i := 0; i < 3; i++ {
		retrieved, err := store.GetByTenantID(ctx, subs[i].TenantID)
		if err != nil {
			t.Fatalf("get %d: %v", i, err)
		}
		if retrieved == nil {
			t.Fatalf("subscription %d not found", i)
		}
		if retrieved.StripeCustomerID != subs[i].StripeCustomerID {
			t.Fatalf("subscription %d corrupted", i)
		}

		// Verify by customer ID as well
		byCustomer, err := store.GetByStripeCustomer(ctx, subs[i].StripeCustomerID)
		if err != nil {
			t.Fatalf("get by customer %d: %v", i, err)
		}
		if byCustomer.TenantID != subs[i].TenantID {
			t.Fatalf("customer lookup %d corrupted", i)
		}
	}

	if store.Count() != 3 {
		t.Fatalf("expected 3 subscriptions, got %d", store.Count())
	}

	t.Logf("✓ Store isolation verified (3 subscriptions)")
}

// TestServiceReady verifies the Ready() health check.
func TestServiceReady(t *testing.T) {
	store := NewInMemoryStore()
	client := &StripeClient{}
	service := NewService(client, store)

	// Service should be ready
	if err := service.Ready(); err != nil {
		t.Fatalf("service should be ready: %v", err)
	}

	// Service with nil stripe should fail
	nilStripeService := &Service{store: store}
	if err := nilStripeService.Ready(); err == nil {
		t.Fatalf("service with nil stripe should not be ready")
	}

	// Service with nil store should fail
	nilStoreService := &Service{stripe: client}
	if err := nilStoreService.Ready(); err == nil {
		t.Fatalf("service with nil store should not be ready")
	}

	t.Log("✓ Service ready checks verified")
}

// TestSubscriptionStatusLifecycle verifies all status transitions are valid.
func TestSubscriptionStatusLifecycle(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	sub := &Subscription{
		ID:        uuid.NewString(),
		TenantID:  "tenant-lifecycle",
		Status:    StatusTrialing,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Valid transitions to test
	transitions := []SubscriptionStatus{
		StatusTrialing,
		StatusActive,
		StatusPastDue,
		StatusActive,
		StatusCanceled,
	}

	for i, newStatus := range transitions {
		sub.Status = newStatus
		sub.UpdatedAt = time.Now()
		if err := store.Upsert(ctx, sub); err != nil {
			t.Fatalf("transition %d (%s): %v", i, newStatus, err)
		}

		retrieved, _ := store.GetByTenantID(ctx, sub.TenantID)
		if retrieved.Status != newStatus {
			t.Fatalf("transition %d: status mismatch, expected %s, got %s", i, newStatus, retrieved.Status)
		}
	}

	t.Logf("✓ Subscription lifecycle verified (%d transitions)", len(transitions))
}

// BenchmarkWebhookProcessing measures webhook throughput.
func BenchmarkWebhookProcessing(b *testing.B) {
	store := NewInMemoryStore()
	client := &StripeClient{pricePro: "price_pro"}
	service := NewService(client, store)
	handler := NewWebhookHandler(client, service, nil)

	ctx := context.Background()

	// Pre-create subscription
	store.Upsert(ctx, &Subscription{
		ID:               uuid.NewString(),
		TenantID:         "bench-tenant",
		StripeCustomerID: "cust_bench",
		Status:           StatusActive,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		event := &stripe.Event{
			Type: "customer.subscription.updated",
			Data: &stripe.EventData{
				Raw: mustMarshal(stripe.Subscription{
					ID:     "sub_bench",
					Status: stripe.SubscriptionStatusActive,
					Customer: &stripe.Customer{
						ID: "cust_bench",
					},
					Items: &stripe.SubscriptionItemList{
						Data: []*stripe.SubscriptionItem{
							{
								Price: &stripe.Price{ID: "price_pro"},
							},
						},
					},
					Metadata: map[string]string{"tenant_id": "bench-tenant"},
				}),
			},
		}
		_ = handler.handleSubscriptionUpdated(ctx, event)
	}
}
