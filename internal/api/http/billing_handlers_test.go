package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/example/offgridflow/internal/auth"
	"github.com/example/offgridflow/internal/billing"
)

func TestBillingHandlersGetPlans(t *testing.T) {
	h := NewBillingHandlersWithConfig(BillingHandlersConfig{})

	req := httptest.NewRequest(http.MethodGet, "/api/billing/plans", nil)
	rec := httptest.NewRecorder()

	h.GetPlans(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var payload BillingPlansResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(payload.Plans) < 3 {
		t.Fatalf("expected at least 3 plans, got %d", len(payload.Plans))
	}
}

func TestBillingHandlersGetStatus_UnpaidSubscriptionNotSubscribed(t *testing.T) {
	store := billing.NewInMemoryStore()
	tenantID := uuid.NewString()
	err := store.Upsert(context.Background(), &billing.Subscription{
		ID:        uuid.NewString(),
		TenantID:  tenantID,
		Plan:      "basic",
		Status:    billing.StatusUnpaid,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("setup subscription: %v", err)
	}

	h := NewBillingHandlersWithConfig(BillingHandlersConfig{
		Service: billing.NewService(nil, store),
	})

	req := httptest.NewRequest(http.MethodGet, "/api/billing/status", nil)
	req = req.WithContext(auth.WithTenant(req.Context(), &auth.Tenant{ID: tenantID, Name: "ACME"}))
	rec := httptest.NewRecorder()

	h.GetStatus(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if subscribed, _ := payload["subscribed"].(bool); subscribed {
		t.Fatalf("expected subscribed=false, got true")
	}
	if status, _ := payload["status"].(string); status != "unpaid" {
		t.Fatalf("expected status=unpaid, got %q", status)
	}
}

func TestBillingHandlersCreateCheckoutSession_ServiceDisabled(t *testing.T) {
	h := NewBillingHandlersWithConfig(BillingHandlersConfig{})

	req := httptest.NewRequest(http.MethodPost, "/api/billing/checkout", strings.NewReader(`{"plan":"basic"}`))
	req = req.WithContext(auth.WithTenant(req.Context(), &auth.Tenant{ID: "tenant-1", Name: "ACME"}))
	req = req.WithContext(auth.WithUser(req.Context(), &auth.User{ID: "user-1", Email: "test@example.com"}))
	rec := httptest.NewRecorder()

	h.CreateCheckoutSession(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}
