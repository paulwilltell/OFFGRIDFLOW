package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	DefaultCSRFTokenTTL     = 24 * time.Hour
	defaultCSRFCleanupEvery = 10 * time.Minute
	defaultTokenLength      = 32
)

const (
	DefaultCSRFHeaderName = "X-CSRF-Token"
	DefaultCSRFCookieName = "csrf_token"
)

// CSRFMiddlewareConfig configures CSRF protections.
type CSRFMiddlewareConfig struct {
	TokenTTL        time.Duration
	CleanupInterval time.Duration
	HeaderName      string
	CookieName      string
	ExemptPaths     []string
}

// CSRFMiddleware verifies CSRF tokens for mutating requests.
type CSRFMiddleware struct {
	tokens         map[string]time.Time
	headerName     string
	cookieName     string
	ttl            time.Duration
	exemptExact    map[string]struct{}
	exemptPrefixes []string
	mu             sync.RWMutex
}

// NewCSRFMiddleware creates a new middleware instance and starts cleanup.
func NewCSRFMiddleware(cfg CSRFMiddlewareConfig) *CSRFMiddleware {
	ttl := cfg.TokenTTL
	if ttl <= 0 {
		ttl = DefaultCSRFTokenTTL
	}

	cleanup := cfg.CleanupInterval
	if cleanup <= 0 {
		cleanup = defaultCSRFCleanupEvery
	}

	header := cfg.HeaderName
	if header == "" {
		header = DefaultCSRFHeaderName
	}

	cookie := cfg.CookieName
	if cookie == "" {
		cookie = DefaultCSRFCookieName
	}

	exact, prefixes := buildExemptPaths(cfg.ExemptPaths)

	m := &CSRFMiddleware{
		tokens:         make(map[string]time.Time),
		headerName:     header,
		cookieName:     cookie,
		ttl:            ttl,
		exemptExact:    exact,
		exemptPrefixes: prefixes,
	}

	go m.cleanup(cleanup)
	return m
}

// Wrap wraps an http.Handler to enforce CSRF protection on unsafe methods.
func (m *CSRFMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isSafeMethod(r.Method) || m.isExemptPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		cookie, err := r.Cookie(m.cookieName)
		if err != nil || cookie.Value == "" {
			http.Error(w, "CSRF token missing", http.StatusForbidden)
			return
		}

		headerToken := r.Header.Get(m.headerName)
		if !m.validateToken(cookie.Value, headerToken) {
			http.Error(w, "CSRF token invalid", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// GenerateToken creates a new CSRF token and stores its expiry.
func (m *CSRFMiddleware) GenerateToken() (string, error) {
	raw := make([]byte, defaultTokenLength)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}

	token := base64.URLEncoding.EncodeToString(raw)

	m.mu.Lock()
	m.tokens[token] = time.Now().Add(m.ttl)
	m.mu.Unlock()

	return token, nil
}

// CookieName returns the configured cookie name.
func (m *CSRFMiddleware) CookieName() string {
	return m.cookieName
}

// HeaderName returns the configured header name.
func (m *CSRFMiddleware) HeaderName() string {
	return m.headerName
}

// TokenTTL returns the configured token lifetime.
func (m *CSRFMiddleware) TokenTTL() time.Duration {
	return m.ttl
}

func (m *CSRFMiddleware) validateToken(cookieToken, headerToken string) bool {
	if cookieToken == "" || headerToken == "" {
		return false
	}

	if subtle.ConstantTimeCompare([]byte(cookieToken), []byte(headerToken)) != 1 {
		return false
	}

	m.mu.RLock()
	expiry, ok := m.tokens[cookieToken]
	m.mu.RUnlock()

	if !ok || time.Now().After(expiry) {
		return false
	}

	return true
}

func (m *CSRFMiddleware) cleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		m.mu.Lock()
		for token, expires := range m.tokens {
			if now.After(expires) {
				delete(m.tokens, token)
			}
		}
		m.mu.Unlock()
	}
}

func (m *CSRFMiddleware) isExemptPath(path string) bool {
	if m.exemptExact == nil && len(m.exemptPrefixes) == 0 {
		return false
	}

	if _, ok := m.exemptExact[path]; ok {
		return true
	}

	for _, prefix := range m.exemptPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	return false
}

func buildExemptPaths(paths []string) (map[string]struct{}, []string) {
	exact := make(map[string]struct{})
	prefixes := make([]string, 0, len(paths))

	for _, p := range paths {
		trimmed := strings.TrimSpace(p)
		if trimmed == "" {
			continue
		}
		if strings.HasSuffix(trimmed, "*") {
			prefixes = append(prefixes, strings.TrimSuffix(trimmed, "*"))
			continue
		}
		if strings.HasSuffix(trimmed, "/") {
			prefixes = append(prefixes, trimmed)
			continue
		}
		exact[trimmed] = struct{}{}
	}

	return exact, prefixes
}

func isSafeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	default:
		return false
	}
}
