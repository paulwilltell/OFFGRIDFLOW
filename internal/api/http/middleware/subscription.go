package middleware

import (
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/example/offgridflow/internal/api/http/responders"
	"github.com/example/offgridflow/internal/auth"
	"github.com/example/offgridflow/internal/billing"
)

// -----------------------------------------------------------------------------
// Subscription Middleware
// -----------------------------------------------------------------------------

// SubscriptionMiddleware enforces active subscription requirements for premium features.
// Routes can be marked as free-tier accessible via the FreePaths configuration.
type SubscriptionMiddleware struct {
	billingSvc *billing.Service
	logger     *slog.Logger
	freePaths  map[string]bool
	freePrefix []string
	mu         sync.RWMutex
}

// SubscriptionMiddlewareConfig holds configuration for subscription middleware.
type SubscriptionMiddlewareConfig struct {
	// BillingService is the billing service to check subscriptions.
	BillingService *billing.Service

	// Logger for subscription-related logging.
	// If nil, a default logger is created.
	Logger *slog.Logger

	// FreePaths are paths that don't require an active subscription.
	// These are accessible on the free tier.
	FreePaths []string
}

// NewSubscriptionMiddleware creates new subscription enforcement middleware.
func NewSubscriptionMiddleware(cfg SubscriptionMiddlewareConfig) *SubscriptionMiddleware {
	freePaths := make(map[string]bool, len(cfg.FreePaths)+6)
	freePrefix := make([]string, 0, len(cfg.FreePaths))
	for _, p := range cfg.FreePaths {
		if strings.HasSuffix(p, "/") {
			freePrefix = append(freePrefix, p)
		} else {
			freePaths[p] = true
		}
	}

	// Always allow core billing endpoints on free tier
	freePaths["/api/billing/checkout"] = true
	freePaths["/api/billing/status"] = true
	freePaths["/api/billing/portal"] = true
	freePaths["/api/billing/webhook"] = true
	// Allow basic emissions endpoints by default
	freePrefix = append(freePrefix, "/api/emissions/")
	freePaths["/api/compliance/summary"] = true

	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default().With("component", "subscription-middleware")
	}

	return &SubscriptionMiddleware{
		billingSvc: cfg.BillingService,
		logger:     logger,
		freePaths:  freePaths,
		freePrefix: freePrefix,
	}
}

// AddFreePath adds a path to the free tier list (thread-safe).
func (m *SubscriptionMiddleware) AddFreePath(path string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.freePaths[path] = true
}

// IsFreePath checks if a path is accessible on free tier (thread-safe).
func (m *SubscriptionMiddleware) IsFreePath(path string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.freePaths[path] {
		return true
	}
	for _, prefix := range m.freePrefix {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

// Wrap returns a handler that enforces an active subscription for premium features.
// The tenant context must be set by AuthMiddleware before this middleware runs.
func (m *SubscriptionMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if path is accessible on free tier
		if m.IsFreePath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Get tenant from context (set by auth middleware)
		tenant, ok := auth.TenantFromContext(r.Context())
		if !ok || tenant == nil {
			responders.Unauthorized(w, "unauthorized", "authentication required")
			return
		}

		// Check subscription status
		sub, err := m.billingSvc.GetSubscription(r.Context(), tenant.ID)
		if err != nil {
			m.logger.Error("failed to check subscription",
				"tenantId", tenant.ID,
				"error", err.Error(),
			)
			// On billing service error, allow request through (fail open)
			// This prevents billing issues from blocking all API access
			next.ServeHTTP(w, r)
			return
		}

		if sub == nil || !sub.IsActive() {
			m.logger.Debug("subscription required but not active",
				"tenantId", tenant.ID,
				"path", r.URL.Path,
			)

			responders.PaymentRequired(w,
				"an active subscription is required for this feature - upgrade at /settings/billing")
			return
		}

		// Subscription is active, continue
		next.ServeHTTP(w, r)
	})
}

// RequireActiveSubscription wraps a single handler with subscription enforcement.
func (m *SubscriptionMiddleware) RequireActiveSubscription(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m.Wrap(http.HandlerFunc(next)).ServeHTTP(w, r)
	}
}

// -----------------------------------------------------------------------------
// Plan-based Feature Gating
// -----------------------------------------------------------------------------

// RequirePlan returns middleware that checks for a minimum subscription plan.
// Plan hierarchy: free < basic < pro < enterprise
func (m *SubscriptionMiddleware) RequirePlan(minPlan string) func(http.Handler) http.Handler {
	planHierarchy := map[string]int{
		"free":       0,
		"basic":      1,
		"pro":        2,
		"enterprise": 3,
	}

	minLevel, ok := planHierarchy[minPlan]
	if !ok {
		minLevel = 0
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenant, ok := auth.TenantFromContext(r.Context())
			if !ok || tenant == nil {
				responders.Unauthorized(w, "unauthorized", "authentication required")
				return
			}

			sub, err := m.billingSvc.GetSubscription(r.Context(), tenant.ID)
			if err != nil {
				// Fail open on billing errors
				next.ServeHTTP(w, r)
				return
			}

			if sub == nil || !sub.IsActive() {
				responders.PaymentRequired(w,
					"subscription required - upgrade at /settings/billing")
				return
			}

			userLevel := planHierarchy[sub.Plan]
			if userLevel < minLevel {
				responders.PaymentRequired(w,
					"this feature requires the "+minPlan+" plan or higher")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireFeature returns middleware that checks for a specific feature flag.
// Feature flags can be defined per-plan in the billing configuration.
func (m *SubscriptionMiddleware) RequireFeature(feature string) func(http.Handler) http.Handler {
	// Feature definitions per plan
	planFeatures := map[string]map[string]bool{
		"free": {
			"basic_emissions":   true,
			"single_user":       true,
			"manual_data_entry": true,
		},
		"basic": {
			"basic_emissions":     true,
			"single_user":         true,
			"manual_data_entry":   true,
			"scope2_calculations": true,
			"csv_import":          true,
			"basic_reports":       true,
		},
		"pro": {
			"basic_emissions":     true,
			"single_user":         false, // Multi-user
			"manual_data_entry":   true,
			"scope2_calculations": true,
			"scope1_calculations": true,
			"scope3_calculations": true,
			"csv_import":          true,
			"api_import":          true,
			"basic_reports":       true,
			"advanced_reports":    true,
			"compliance_reports":  true,
			"ai_chat":             true,
		},
		"enterprise": {
			// All features enabled
			"basic_emissions":       true,
			"manual_data_entry":     true,
			"scope2_calculations":   true,
			"scope1_calculations":   true,
			"scope3_calculations":   true,
			"csv_import":            true,
			"api_import":            true,
			"basic_reports":         true,
			"advanced_reports":      true,
			"compliance_reports":    true,
			"ai_chat":               true,
			"sso":                   true,
			"audit_logs":            true,
			"custom_integrations":   true,
			"dedicated_support":     true,
			"data_retention_custom": true,
		},
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenant, ok := auth.TenantFromContext(r.Context())
			if !ok || tenant == nil {
				responders.Unauthorized(w, "unauthorized", "authentication required")
				return
			}

			sub, err := m.billingSvc.GetSubscription(r.Context(), tenant.ID)
			if err != nil {
				// Fail open on billing errors
				next.ServeHTTP(w, r)
				return
			}

			plan := "free"
			if sub != nil && sub.IsActive() {
				plan = sub.Plan
			}

			features := planFeatures[plan]
			if features == nil {
				features = planFeatures["free"]
			}

			if !features[feature] {
				responders.PaymentRequired(w,
					"the '"+feature+"' feature requires a plan upgrade")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
