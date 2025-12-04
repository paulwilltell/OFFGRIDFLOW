package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/example/offgridflow/internal/api/http/responders"
	"github.com/example/offgridflow/internal/auth"
	"github.com/example/offgridflow/internal/billing"
)

// UsageMiddleware enforces usage limits and tracks API usage.
type UsageMiddleware struct {
	tracker *billing.UsageTracker
	logger  *slog.Logger
}

// UsageMiddlewareConfig holds configuration for usage middleware.
type UsageMiddlewareConfig struct {
	Tracker *billing.UsageTracker
	Logger  *slog.Logger
}

// NewUsageMiddleware creates new usage enforcement middleware.
func NewUsageMiddleware(cfg UsageMiddlewareConfig) *UsageMiddleware {
	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default().With("component", "usage-middleware")
	}

	return &UsageMiddleware{
		tracker: cfg.Tracker,
		logger:  logger,
	}
}

// Wrap returns a handler that tracks API usage and enforces limits.
func (m *UsageMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenant, ok := auth.TenantFromContext(r.Context())
		if !ok || tenant == nil {
			// No tenant context, skip usage tracking
			next.ServeHTTP(w, r)
			return
		}

		// Track API call usage
		if err := m.tracker.RecordUsage(r.Context(), tenant.ID, billing.UsageAPICall, 1); err != nil {
			m.logger.Error("failed to record API usage",
				"tenantId", tenant.ID,
				"error", err.Error(),
			)
		}

		// Check API call limit
		allowed, err := m.tracker.CheckLimit(r.Context(), tenant.ID, billing.UsageAPICall)
		if err != nil {
			m.logger.Error("failed to check API limit",
				"tenantId", tenant.ID,
				"error", err.Error(),
			)
			// Fail open
			next.ServeHTTP(w, r)
			return
		}

		if !allowed {
			responders.TooManyRequests(w, "API call limit exceeded for your plan - upgrade at /settings/billing")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequireAIQuota checks AI query quota before allowing AI chat requests.
func (m *UsageMiddleware) RequireAIQuota(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/ai/") {
			next.ServeHTTP(w, r)
			return
		}

		tenant, ok := auth.TenantFromContext(r.Context())
		if !ok || tenant == nil {
			responders.Unauthorized(w, "unauthorized", "authentication required")
			return
		}

		// Check AI query limit
		allowed, err := m.tracker.CheckLimit(r.Context(), tenant.ID, billing.UsageAIQuery)
		if err != nil {
			m.logger.Error("failed to check AI limit",
				"tenantId", tenant.ID,
				"error", err.Error(),
			)
			next.ServeHTTP(w, r)
			return
		}

		if !allowed {
			responders.PaymentRequired(w, "AI query limit exceeded for your plan - upgrade to Pro for more queries")
			return
		}

		// Record AI query usage
		if err := m.tracker.RecordUsage(r.Context(), tenant.ID, billing.UsageAIQuery, 1); err != nil {
			m.logger.Error("failed to record AI usage",
				"tenantId", tenant.ID,
				"error", err.Error(),
			)
		}

		next.ServeHTTP(w, r)
	})
}

// RequireReportQuota checks report generation quota.
func (m *UsageMiddleware) RequireReportQuota(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/compliance/") || r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}

		tenant, ok := auth.TenantFromContext(r.Context())
		if !ok || tenant == nil {
			responders.Unauthorized(w, "unauthorized", "authentication required")
			return
		}

		// Check report generation limit
		allowed, err := m.tracker.CheckLimit(r.Context(), tenant.ID, billing.UsageReportGenerated)
		if err != nil {
			m.logger.Error("failed to check report limit",
				"tenantId", tenant.ID,
				"error", err.Error(),
			)
			next.ServeHTTP(w, r)
			return
		}

		if !allowed {
			responders.PaymentRequired(w, "Report generation limit exceeded for your plan - upgrade for more reports")
			return
		}

		// Record report generation
		if err := m.tracker.RecordUsage(r.Context(), tenant.ID, billing.UsageReportGenerated, 1); err != nil {
			m.logger.Error("failed to record report usage",
				"tenantId", tenant.ID,
				"error", err.Error(),
			)
		}

		next.ServeHTTP(w, r)
	})
}

// RequireEmissionRecordQuota checks emission record storage quota.
func (m *UsageMiddleware) RequireEmissionRecordQuota(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/emissions/") || r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}

		tenant, ok := auth.TenantFromContext(r.Context())
		if !ok || tenant == nil {
			responders.Unauthorized(w, "unauthorized", "authentication required")
			return
		}

		// Check emission record limit
		allowed, err := m.tracker.CheckLimit(r.Context(), tenant.ID, billing.UsageEmissionRecord)
		if err != nil {
			m.logger.Error("failed to check emission record limit",
				"tenantId", tenant.ID,
				"error", err.Error(),
			)
			next.ServeHTTP(w, r)
			return
		}

		if !allowed {
			responders.PaymentRequired(w, "Emission record limit exceeded for your plan - upgrade for unlimited records")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// GetUsageHandler returns a handler for GET /api/billing/usage.
func (m *UsageMiddleware) GetUsageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			responders.MethodNotAllowed(w, http.MethodGet)
			return
		}

		tenant, ok := auth.TenantFromContext(r.Context())
		if !ok || tenant == nil {
			responders.Unauthorized(w, "unauthorized", "authentication required")
			return
		}

		summary, err := m.tracker.GetUsageSummary(r.Context(), tenant.ID)
		if err != nil {
			responders.InternalError(w, "failed to get usage summary")
			return
		}

		responders.JSON(w, http.StatusOK, summary)
	}
}
