// Package http provides billing-related HTTP handlers for subscription management.
//
// Endpoints:
//   - POST /api/billing/checkout  - Create a Stripe checkout session
//   - POST /api/billing/webhook   - Handle Stripe webhook events
//   - GET  /api/billing/status    - Get current subscription status
//   - POST /api/billing/portal    - Create Stripe customer portal session
package http

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/example/offgridflow/internal/api/http/responders"
	"github.com/example/offgridflow/internal/auth"
	"github.com/example/offgridflow/internal/billing"
)

// -----------------------------------------------------------------------------
// Request/Response Types
// -----------------------------------------------------------------------------

// CheckoutRequest represents a request to create a checkout session.
type CheckoutRequest struct {
	Plan string `json:"plan"` // Subscription plan: "basic", "professional", "enterprise"
}

// CheckoutResponse represents the response from checkout session creation.
type CheckoutResponse struct {
	URL       string `json:"url"`                 // Stripe checkout URL
	SessionID string `json:"sessionId,omitempty"` // Stripe session ID
}

// SubscriptionStatusResponse represents the current subscription status.
type SubscriptionStatusResponse struct {
	Subscribed       bool    `json:"subscribed"`
	Plan             *string `json:"plan,omitempty"`
	Status           *string `json:"status,omitempty"`
	CurrentPeriodEnd *string `json:"currentPeriodEnd,omitempty"`
}

// PortalResponse represents the response from portal session creation.
type PortalResponse struct {
	URL string `json:"url"` // Stripe customer portal URL
}

// -----------------------------------------------------------------------------
// Handler Configuration
// -----------------------------------------------------------------------------

// BillingHandlersConfig holds configuration for billing handlers.
type BillingHandlersConfig struct {
	Service    *billing.Service
	SuccessURL string // Redirect after successful checkout
	CancelURL  string // Redirect if checkout canceled
	PortalURL  string // Redirect after billing portal
	Logger     *slog.Logger
}

// BillingHandlers bundles billing-related endpoints.
type BillingHandlers struct {
	service    *billing.Service
	successURL string
	cancelURL  string
	portalURL  string
	logger     *slog.Logger
}

// Supported subscription plans.
var validPlans = map[string]bool{
	"basic":        true,
	"professional": true,
	"enterprise":   true,
}

// -----------------------------------------------------------------------------
// Constructor Functions
// -----------------------------------------------------------------------------

// NewBillingHandlers creates a handler set for billing routes with defaults.
func NewBillingHandlers(svc *billing.Service) *BillingHandlers {
	return NewBillingHandlersWithConfig(BillingHandlersConfig{
		Service:    svc,
		SuccessURL: "/settings/billing?success=true",
		CancelURL:  "/settings/billing?canceled=true",
		PortalURL:  "/settings/billing",
	})
}

// NewBillingHandlersWithConfig creates a handler set with custom configuration.
func NewBillingHandlersWithConfig(cfg BillingHandlersConfig) *BillingHandlers {
	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default().With("component", "billing-handlers")
	}

	return &BillingHandlers{
		service:    cfg.Service,
		successURL: cfg.SuccessURL,
		cancelURL:  cfg.CancelURL,
		portalURL:  cfg.PortalURL,
		logger:     logger,
	}
}

// -----------------------------------------------------------------------------
// HTTP Handlers
// -----------------------------------------------------------------------------

// CreateCheckoutSession handles POST /api/billing/checkout.
// Creates a Stripe checkout session for subscription purchase.
func (h *BillingHandlers) CreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		responders.MethodNotAllowed(w, http.MethodPost)
		return
	}

	// Extract authenticated user and tenant
	user, ok := auth.UserFromContext(r.Context())
	if !ok || user == nil {
		responders.Unauthorized(w, "unauthorized", "authentication required")
		return
	}

	tenant, ok := auth.TenantFromContext(r.Context())
	if !ok || tenant == nil {
		responders.Unauthorized(w, "unauthorized", "tenant context missing")
		return
	}

	// Parse request body
	var req CheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responders.BadRequest(w, "invalid_request", "invalid JSON payload")
		return
	}

	// Validate and normalize plan
	plan := strings.ToLower(strings.TrimSpace(req.Plan))
	if plan == "" {
		plan = "basic"
	}

	if !validPlans[plan] {
		responders.BadRequest(w, "invalid_plan", "plan must be one of: basic, professional, enterprise")
		return
	}

	// Create Stripe checkout session
	ctx := r.Context()
	url, err := h.service.StartSubscription(ctx, tenant.ID, tenant.Name, user.Email, plan, h.successURL, h.cancelURL)
	if err != nil {
		h.logger.Error("failed to create checkout session",
			"tenantId", tenant.ID,
			"plan", plan,
			"error", err.Error(),
		)
		responders.InternalError(w, "failed to create checkout session")
		return
	}

	h.logger.Info("checkout session created",
		"tenantId", tenant.ID,
		"userId", user.ID,
		"plan", plan,
	)

	responders.JSON(w, http.StatusOK, CheckoutResponse{URL: url})
}

// HandleWebhook handles POST /api/billing/webhook.
// Processes Stripe webhook events for subscription lifecycle management.
func (h *BillingHandlers) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		responders.MethodNotAllowed(w, http.MethodPost)
		return
	}

	// Parse and validate webhook signature
	event, err := h.service.ParseWebhook(r)
	if err != nil {
		h.logger.Warn("webhook signature validation failed",
			"error", err.Error(),
		)
		responders.Error(w, http.StatusBadRequest, "webhook_invalid", "invalid signature")
		return
	}

	// Process the webhook event
	ctx := r.Context()
	if err := h.service.HandleWebhookEvent(ctx, event); err != nil {
		h.logger.Error("webhook event processing failed",
			"eventType", event.Type,
			"error", err.Error(),
		)
		responders.InternalError(w, "webhook processing failed")
		return
	}

	h.logger.Info("webhook event processed",
		"eventType", event.Type,
		"eventId", event.ID,
	)

	// Stripe expects a 200 response
	w.WriteHeader(http.StatusOK)
}

// GetStatus handles GET /api/billing/status.
// Returns the current subscription status for the authenticated tenant.
func (h *BillingHandlers) GetStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responders.MethodNotAllowed(w, http.MethodGet)
		return
	}

	tenant, ok := auth.TenantFromContext(r.Context())
	if !ok || tenant == nil {
		responders.Unauthorized(w, "unauthorized", "authentication required")
		return
	}

	// Fetch subscription from billing service
	ctx := r.Context()
	sub, err := h.service.GetSubscription(ctx, tenant.ID)
	if err != nil {
		h.logger.Error("failed to fetch subscription",
			"tenantId", tenant.ID,
			"error", err.Error(),
		)
		responders.InternalError(w, "failed to fetch subscription status")
		return
	}

	// Return unsubscribed status if no subscription found
	if sub == nil {
		responders.JSON(w, http.StatusOK, SubscriptionStatusResponse{
			Subscribed: false,
		})
		return
	}

	// Format timestamps
	var periodEnd string
	if sub.CurrentPeriodEnd != nil {
		periodEnd = sub.CurrentPeriodEnd.Format(time.RFC3339)
	}

	// Convert status to string
	status := string(sub.Status)

	response := SubscriptionStatusResponse{
		Subscribed:       true,
		Plan:             &sub.Plan,
		Status:           &status,
		CurrentPeriodEnd: &periodEnd,
	}

	// Cache subscription status briefly
	responders.SetCacheControl(w, 30*time.Second, true) // private, 30 seconds
	responders.JSON(w, http.StatusOK, response)
}

// CreatePortalSession handles POST /api/billing/portal.
// Creates a Stripe customer portal session for subscription management.
func (h *BillingHandlers) CreatePortalSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		responders.MethodNotAllowed(w, http.MethodPost)
		return
	}

	tenant, ok := auth.TenantFromContext(r.Context())
	if !ok || tenant == nil {
		responders.Unauthorized(w, "unauthorized", "authentication required")
		return
	}

	// Create portal session
	ctx := r.Context()
	url, err := h.service.CreateBillingPortalSession(ctx, tenant.ID, h.portalURL)
	if err != nil {
		h.logger.Error("failed to create portal session",
			"tenantId", tenant.ID,
			"error", err.Error(),
		)
		responders.InternalError(w, "failed to create portal session")
		return
	}

	h.logger.Info("portal session created",
		"tenantId", tenant.ID,
	)

	responders.JSON(w, http.StatusOK, PortalResponse{URL: url})
}
