package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/example/offgridflow/internal/auth"
	"github.com/example/offgridflow/internal/ratelimit"
)

// RateLimitMiddleware applies rate limiting based on tenant tier
func RateLimitMiddleware(limiter *ratelimit.MultiTierLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip rate limiting only for preflight requests
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}
			// All other methods (GET, POST, PUT, DELETE) are rate limited

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
		"free":       "5",
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
	// Only trust the leftmost (client) IP from X-Forwarded-For
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		if idx := strings.IndexByte(forwarded, ','); idx != -1 {
			return strings.TrimSpace(forwarded[:idx])
		}
		return strings.TrimSpace(forwarded)
	}

	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return strings.TrimSpace(realIP)
	}

	return r.RemoteAddr
}

// IPRateLimiter provides simple per-IP rate limiting for public endpoints.
type IPRateLimiter struct {
	mu      sync.Mutex
	counts  map[string]int
	resets  map[string]time.Time
	limit   int
	window  time.Duration
}

// NewIPRateLimiter creates a rate limiter that allows limit requests per IP per window.
func NewIPRateLimiter(limit int, window time.Duration) *IPRateLimiter {
	return &IPRateLimiter{
		counts: make(map[string]int),
		resets: make(map[string]time.Time),
		limit:  limit,
		window: window,
	}
}

// Wrap wraps an http.HandlerFunc with IP-based rate limiting.
func (rl *IPRateLimiter) Wrap(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)
		rl.mu.Lock()
		resetAt, exists := rl.resets[ip]
		if !exists || time.Now().After(resetAt) {
			rl.counts[ip] = 0
			rl.resets[ip] = time.Now().Add(rl.window)
		}
		rl.counts[ip]++
		count := rl.counts[ip]
		rl.mu.Unlock()

		if count > rl.limit {
			w.Header().Set("Retry-After", fmt.Sprintf("%d", int(rl.window.Seconds())))
			http.Error(w, `{"error":"rate_limit_exceeded","message":"too many requests"}`, http.StatusTooManyRequests)
			return
		}
		next(w, r)
	}
}
