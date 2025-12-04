package middleware

import (
	"fmt"
	"net/http"

	"github.com/example/offgridflow/internal/auth"
	"github.com/example/offgridflow/internal/ratelimit"
)

// RateLimitMiddleware applies rate limiting based on tenant tier
func RateLimitMiddleware(limiter *ratelimit.MultiTierLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Get tenant from context
			tenant, ok := auth.TenantFromContext(ctx)
			if !ok {
				// No tenant context, apply strictest limit
				key := ratelimit.IPKeyFunc(getClientIP(r))
				if !limiter.Allow(ctx, "free", key) {
					writeRateLimitError(w, "free")
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			// Determine tier from tenant plan
			tier := tenant.Plan
			if tier == "" {
				tier = "free"
			}

			// Use tenant ID as rate limit key
			key := ratelimit.DefaultKeyFunc(tenant.ID)

			if !limiter.Allow(ctx, tier, key) {
				writeRateLimitError(w, tier)
				return
			}

			// Add rate limit headers
			remaining := limiter.Remaining(ctx, tier, key)
			w.Header().Set("X-RateLimit-Limit", getTierLimit(tier))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitByAPIKey applies rate limiting using API key as the identifier
func RateLimitByAPIKey(limiter *ratelimit.RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Get API key from context
			apiKey, ok := auth.APIKeyFromContext(ctx)
			if !ok {
				// No API key, use IP-based limiting
				key := ratelimit.IPKeyFunc(getClientIP(r))
				if !limiter.Allow(ctx, key) {
					http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			key := ratelimit.APIKeyFunc(apiKey.ID)
			if !limiter.Allow(ctx, key) {
				w.Header().Set("X-RateLimit-Limit", "0")
				w.Header().Set("X-RateLimit-Remaining", "0")
				http.Error(w, "Rate limit exceeded for this API key", http.StatusTooManyRequests)
				return
			}

			remaining := limiter.Remaining(ctx, key)
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

			next.ServeHTTP(w, r)
		})
	}
}

func writeRateLimitError(w http.ResponseWriter, tier string) {
	w.Header().Set("X-RateLimit-Limit", getTierLimit(tier))
	w.Header().Set("X-RateLimit-Remaining", "0")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusTooManyRequests)

	response := fmt.Sprintf(`{"error":"Rate limit exceeded","tier":"%s","message":"Too many requests. Please upgrade your plan for higher limits."}`, tier)
	w.Write([]byte(response))
}

func getTierLimit(tier string) string {
	limits := map[string]string{
		"free":       "10",
		"pro":        "100",
		"enterprise": "1000",
	}

	limit, ok := limits[tier]
	if !ok {
		return "10"
	}
	return limit
}

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fallback to RemoteAddr
	return r.RemoteAddr
}
