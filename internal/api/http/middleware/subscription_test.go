package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/offgridflow/internal/auth"
	"github.com/example/offgridflow/internal/billing"
)

func TestSubscriptionMiddlewareBlocksPaidPaths(t *testing.T) {
	store := billing.NewInMemoryStore()
	_ = store.Upsert(nil, &billing.Subscription{
		ID:               "sub1",
		TenantID:         "tenant-1",
		StripeCustomerID: "cust1",
		Status:           billing.StatusCanceled,
	})
	mw := NewSubscriptionMiddleware(SubscriptionMiddlewareConfig{
		BillingService: billing.NewService(nil, store),
	})

	protected := mw.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/ai/chat", nil)
	req = req.WithContext(auth.WithTenant(req.Context(), &auth.Tenant{ID: "tenant-1"}))
	rr := httptest.NewRecorder()

	protected.ServeHTTP(rr, req)

	if rr.Code != http.StatusPaymentRequired {
		t.Fatalf("expected 402, got %d", rr.Code)
	}
}

func TestSubscriptionMiddlewareAllowsFreePaths(t *testing.T) {
	mw := NewSubscriptionMiddleware(SubscriptionMiddlewareConfig{
		BillingService: billing.NewService(nil, billing.NewInMemoryStore()),
	})

	protected := mw.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/emissions/scope2", nil)
	req = req.WithContext(auth.WithTenant(req.Context(), &auth.Tenant{ID: "tenant-1"}))
	rr := httptest.NewRecorder()

	protected.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}
