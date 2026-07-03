// Package http provides billing-related HTTP handlers for subscription management.
//
// Endpoints:
//   - POST /api/billing/checkout  - Create a Stripe checkout session
//   - POST /api/billing/webhook   - Handle Stripe webhook events
//   - GET  /api/billing/status    - Get current subscription status
//   - GET  /api/billing/plans     - List available subscription plans
//   - POST /api/billing/portal    - Create Stripe customer portal session
package http

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
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
	Plan       string `json:"plan"`                  // Canonical plan field
	PlanID     string `json:"plan_id"`               // Backward-compatible alias
	SuccessURL string `json:"success_url,omitempty"` // Absolute URL after successful checkout
	CancelURL  string `json:"cancel_url,omitempty"`  // Absolute URL when checkout is canceled
}

// CheckoutResponse represents the response from checkout session creation.
type CheckoutResponse struct {
	URL         string `json:"url"`                    // Canonical checkout URL
	CheckoutURL string `json:"checkout_url,omitempty"` // Backward-compatible alias
	SessionID   string `json:"sessionId,omitempty"`    // Stripe session ID (if available)
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
	URL       string `json:"url"`                  // Canonical portal URL
	PortalURL string `json:"portal_url,omitempty"` // Backward-compatible alias
}

// PortalRequest contains an optional return URL for billing portal.
type PortalRequest struct {
	ReturnURL string `json:"return_url,omitempty"`
}

// BillingPlan represents a purchasable subscription plan.
type BillingPlan struct {
	ID          string   `json:"id"`
	PriceID     string   `json:"price_id"`
	Name        string   `json:"name"`
	AmountCents int64    `json:"amount_cents"`
	Interval    string   `json:"interval"`
	Features    []string `json:"features"`
}

// BillingPlansResponse returns the available plan catalog.
type BillingPlansResponse struct {
	Plans []BillingPlan `json:"plans"`
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
	"pro":          true,
	"professional": true, // accepted alias, normalized to "pro"
	"enterprise":   true,
}

var defaultPlanCatalog = []BillingPlan{
	// ── Audit Prep — $6,500/yr ───────────────────────────────────────────────
	{
		ID:          "basic",
		PriceID:     "basic_annual",
		Name:        "Audit Prep",
		AmountCents: 650000,
		Interval:    "year",
		Features: []string{
			"Scope 1 & 2 emissions tracking",
			"CSV & utility bill import",
			"Single compliance framework (CSRD or SB 253)",
			"PDF compliance reports",
			"EPA eGRID emission factors",
			"Up to 5 users",
			"Email support",
		},
	},
	// ── Compliance Pro — $10,800/yr ──────────────────────────────────────────
	{
		ID:          "pro",
		PriceID:     "pro_annual",
		Name:        "Compliance Pro",
		AmountCents: 1080000,
		Interval:    "year",
		Features: []string{
			"Scope 1, 2 & basic Scope 3 tracking",
			"CSRD + SEC compliance frameworks",
			"Cloud connectors (AWS, Azure, GCP)",
			"PDF + XBRL exports",
			"EPA eGRID + DEFRA + IEA factors",
			"Up to 15 users",
			"Priority email support",
		},
	},
	// ── Enterprise — $15,000/yr ──────────────────────────────────────────────
	{
		ID:          "enterprise",
		PriceID:     "enterprise_annual",
		Name:        "Enterprise",
		AmountCents: 1500000,
		Interval:    "year",
		Features: []string{
			"Full Scope 1, 2 & 3 tracking",
			"All 5 compliance frameworks (CSRD, SEC, SB 253, CBAM, IFRS S2)",
			"Cloud connectors + SAP integration",
			"PDF + XBRL/iXBRL exports",
			"Advanced analytics & forecasting",
			"Up to 25 users",
			"Dedicated account manager",
		},
	},
	// ── Global — Custom pricing ──────────────────────────────────────────────
	{
		ID:          "global",
		PriceID:     "",
		Name:        "Global",
		AmountCents: 0,
		Interval:    "year",
		Features: []string{
			"Everything in Enterprise",
			"All global frameworks including GRI and CDP",
			"Multi-region compliance (EU, UK, CA & more)",
			"Custom calculation methodologies",
			"On-site implementation support",
			"White-label branding & SSO",
			"Executive dashboard & board reporting",
			"Dedicated customer success manager",
			"99.9% SLA guarantee",
			"Custom pricing — contact us",
		},
	},
}

// -----------------------------------------------------------------------------
// Constructor Functions
// -----------------------------------------------------------------------------

// NewBillingHandlers creates a handler set for billing routes with defaults.
func NewBillingHandlers(svc *billing.Service) *BillingHandlers {
	return NewBillingHandlersWithConfig(BillingHandlersConfig{
		Service:    svc,
		SuccessURL: "http://localhost:3000/settings/billing?success=true",
		CancelURL:  "http://localhost:3000/settings/billing?canceled=true",
		PortalURL:  "http://localhost:3000/settings/billing",
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
		successURL: normalizeDefaultURL(cfg.SuccessURL, "http://localhost:3000/settings/billing?success=true"),
		cancelURL:  normalizeDefaultURL(cfg.CancelURL, "http://localhost:3000/settings/billing?canceled=true"),
		portalURL:  normalizeDefaultURL(cfg.PortalURL, "http://localhost:3000/settings/billing"),
		logger:     logger,
	}
}

// -----------------------------------------------------------------------------
// HTTP Handlers
// -----------------------------------------------------------------------------

// CreateCheckoutSession handles POST /api/billing/checkout.
// Creates a Stripe checkout session for subscription purchase.
func (h *BillingHandlers) CreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		responders.Error(w, http.StatusServiceUnavailable, "billing_disabled", "billing is not configured")
		return
	}

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
	if err := decodeJSONBody(r, &req); err != nil {
		responders.BadRequest(w, "invalid_request", "invalid JSON payload")
		return
	}

	// Validate and normalize plan
	plan := strings.TrimSpace(req.Plan)
	if plan == "" {
		plan = strings.TrimSpace(req.PlanID)
	}
	if plan == "" {
		plan = "basic"
	}
	plan = strings.ToLower(plan)

	successURL := resolveRedirectURL(req.SuccessURL, h.successURL)
	cancelURL := resolveRedirectURL(req.CancelURL, h.cancelURL)
	ctx := r.Context()

	// Pay-per-report: a one-time $149 checkout to unlock report exports.
	if plan == "report_export" || plan == "report" {
		url, err := h.service.StartReportCheckout(ctx, tenant.ID, tenant.Name, user.Email, successURL, cancelURL)
		if err != nil {
			h.logger.Error("failed to create report checkout session", "tenantId", tenant.ID, "error", err.Error())
			responders.InternalError(w, "failed to create checkout session")
			return
		}
		h.logger.Info("report checkout session created", "tenantId", tenant.ID, "userId", user.ID)
		responders.JSON(w, http.StatusOK, CheckoutResponse{URL: url, CheckoutURL: url})
		return
	}

	if !validPlans[plan] {
		responders.BadRequest(w, "invalid_plan", "plan must be one of: basic, pro, enterprise")
		return
	}
	if plan == "professional" {
		plan = "pro"
	}

	// Create Stripe checkout session
	url, err := h.service.StartSubscription(ctx, tenant.ID, tenant.Name, user.Email, plan, successURL, cancelURL)
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
		"successUrl", successURL,
		"cancelUrl", cancelURL,
	)

	responders.JSON(w, http.StatusOK, CheckoutResponse{
		URL:         url,
		CheckoutURL: url,
	})
}

// HandleWebhook handles POST /api/billing/webhook.
// Processes Stripe webhook events for subscription lifecycle management.
func (h *BillingHandlers) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		responders.Error(w, http.StatusServiceUnavailable, "billing_disabled", "billing is not configured")
		return
	}

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

// GetPlans handles GET /api/billing/plans.
// Returns the available subscription plans for checkout.
func (h *BillingHandlers) GetPlans(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responders.MethodNotAllowed(w, http.MethodGet)
		return
	}

	plans := make([]BillingPlan, 0, len(defaultPlanCatalog))
	plans = append(plans, defaultPlanCatalog...)

	responders.SetCacheControl(w, 5*time.Minute, true)
	responders.JSON(w, http.StatusOK, BillingPlansResponse{Plans: plans})
}

// GetStatus handles GET /api/billing/status.
// Returns the current subscription status for the authenticated tenant.
func (h *BillingHandlers) GetStatus(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		responders.Error(w, http.StatusServiceUnavailable, "billing_disabled", "billing is not configured")
		return
	}

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

	// Format timestamp only when available.
	var periodEnd *string
	if sub.CurrentPeriodEnd != nil {
		formatted := sub.CurrentPeriodEnd.Format(time.RFC3339)
		periodEnd = &formatted
	}

	// Convert status to string
	status := string(sub.Status)
	var planPtr *string
	if sub.Plan != "" {
		planPtr = &sub.Plan
	}

	response := SubscriptionStatusResponse{
		Subscribed:       sub.IsActive(),
		Plan:             planPtr,
		Status:           &status,
		CurrentPeriodEnd: periodEnd,
	}

	// Cache subscription status briefly
	responders.SetCacheControl(w, 30*time.Second, true) // private, 30 seconds
	responders.JSON(w, http.StatusOK, response)
}

// CreatePortalSession handles POST /api/billing/portal.
// Creates a Stripe customer portal session for subscription management.
func (h *BillingHandlers) CreatePortalSession(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		responders.Error(w, http.StatusServiceUnavailable, "billing_disabled", "billing is not configured")
		return
	}

	if r.Method != http.MethodPost {
		responders.MethodNotAllowed(w, http.MethodPost)
		return
	}

	tenant, ok := auth.TenantFromContext(r.Context())
	if !ok || tenant == nil {
		responders.Unauthorized(w, "unauthorized", "authentication required")
		return
	}

	var req PortalRequest
	if err := decodeJSONBody(r, &req); err != nil {
		responders.BadRequest(w, "invalid_request", "invalid JSON payload")
		return
	}
	returnURL := resolveRedirectURL(req.ReturnURL, h.portalURL)

	// Create portal session.
	ctx := r.Context()
	url, err := h.service.CreateBillingPortalSession(ctx, tenant.ID, returnURL)
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
		"returnUrl", returnURL,
	)

	responders.JSON(w, http.StatusOK, PortalResponse{
		URL:       url,
		PortalURL: url,
	})
}

func decodeJSONBody(r *http.Request, out interface{}) error {
	if r == nil || r.Body == nil {
		return nil
	}
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(out); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}
	return nil
}

func normalizeDefaultURL(raw, fallback string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		value = fallback
	}
	if parsed, err := url.Parse(value); err == nil && parsed.IsAbs() && isHTTPURL(parsed) {
		return parsed.String()
	}

	// If an invalid or relative URL is provided in config, fall back safely.
	return fallback
}

func resolveRedirectURL(candidate, fallback string) string {
	value := strings.TrimSpace(candidate)
	if value == "" {
		return fallback
	}
	parsed, err := url.Parse(value)
	if err != nil || !parsed.IsAbs() || !isHTTPURL(parsed) {
		return fallback
	}
	return parsed.String()
}

func isHTTPURL(u *url.URL) bool {
	if u == nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}
