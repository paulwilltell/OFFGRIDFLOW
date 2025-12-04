// Package middleware provides HTTP middleware for authentication, authorization,
// and request processing for the OffGridFlow API.
package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/example/offgridflow/internal/api/http/responders"
	"github.com/example/offgridflow/internal/auth"
)

// -----------------------------------------------------------------------------
// Constants
// -----------------------------------------------------------------------------

const (
	// sessionCookieName is the name of the session cookie.
	sessionCookieName = "offgrid_session"

	// apiKeyHeader is the header for API key authentication.
	apiKeyHeader = "X-API-Key"

	// authorizationHeader is the standard Authorization header.
	authorizationHeader = "Authorization"

	// bearerPrefix is the prefix for Bearer token authentication.
	bearerPrefix = "Bearer "

	// tenantIDHeader allows lightweight tenant scoping without full auth.
	tenantIDHeader = "X-Tenant-ID"
	orgIDHeader    = "X-Org-ID"
)

// -----------------------------------------------------------------------------
// Authentication Middleware
// -----------------------------------------------------------------------------

// AuthMiddleware provides API key and session authentication for HTTP handlers.
// It supports multiple authentication methods:
//   - Session cookies (for browser-based clients)
//   - Bearer tokens in Authorization header (JWT sessions or API keys)
//   - X-API-Key header (for programmatic access)
type AuthMiddleware struct {
	authStore       auth.Store
	sessionManager  *auth.SessionManager
	logger          *slog.Logger
	requireAuth     bool
	allowedPaths    map[string]bool
	allowedPrefixes []string
	mu              sync.RWMutex
}

// AuthMiddlewareConfig holds configuration for auth middleware.
type AuthMiddlewareConfig struct {
	// AuthStore is the backing store for users, tenants, and API keys.
	AuthStore auth.Store

	// SessionManager handles JWT session token creation and validation.
	// If nil, JWT session authentication is disabled.
	SessionManager *auth.SessionManager

	// Logger is used for authentication-related logging.
	// If nil, a default logger is created.
	Logger *slog.Logger

	// RequireAuth determines if authentication is mandatory.
	// If false, unauthenticated requests are allowed through.
	RequireAuth bool

	// AllowedPaths are paths that bypass authentication even when required.
	AllowedPaths []string
}

// NewAuthMiddleware creates a new authentication middleware with the given configuration.
func NewAuthMiddleware(cfg AuthMiddlewareConfig) *AuthMiddleware {
	allowed := make(map[string]bool, len(cfg.AllowedPaths)+2)
	prefixes := make([]string, 0, len(cfg.AllowedPaths))
	for _, p := range cfg.AllowedPaths {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if strings.HasSuffix(p, "*") {
			prefixes = append(prefixes, strings.TrimSuffix(p, "*"))
			continue
		}
		if strings.HasSuffix(p, "/") {
			prefixes = append(prefixes, p)
			continue
		}
		allowed[p] = true
	}
	// Always allow health check endpoints
	allowed["/health"] = true
	allowed["/healthz"] = true
	allowed["/readyz"] = true

	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default().With("component", "auth-middleware")
	}

	return &AuthMiddleware{
		authStore:       cfg.AuthStore,
		sessionManager:  cfg.SessionManager,
		logger:          logger,
		requireAuth:     cfg.RequireAuth,
		allowedPaths:    allowed,
		allowedPrefixes: prefixes,
	}
}

// AddAllowedPath adds a path to the allowed paths list (thread-safe).
func (m *AuthMiddleware) AddAllowedPath(path string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if strings.HasSuffix(path, "*") {
		m.allowedPrefixes = append(m.allowedPrefixes, strings.TrimSuffix(path, "*"))
		return
	}
	if strings.HasSuffix(path, "/") {
		m.allowedPrefixes = append(m.allowedPrefixes, path)
		return
	}
	m.allowedPaths[path] = true
}

// IsAllowedPath checks if a path is in the allowed paths list (thread-safe).
func (m *AuthMiddleware) IsAllowedPath(path string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.allowedPaths[path] {
		return true
	}
	for _, prefix := range m.allowedPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

// Wrap wraps an http.Handler with authentication.
// Authentication flow:
//  1. Check if path is in allowed list (bypass auth)
//  2. Try session cookie (for browser clients)
//  3. Try Authorization header (Bearer token - JWT or API key)
//  4. Try X-API-Key header (explicit API key)
//  5. If requireAuth and no valid auth found, return 401
func (m *AuthMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Attempt to hydrate tenant context even if authentication is optional.
		r = m.hydrateTenantContext(r)

		// Check if path bypasses authentication
		if m.IsAllowedPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		ctx := r.Context()

		// Try session cookie first (browser clients)
		if authenticated, newCtx := m.trySessionCookie(r); authenticated {
			next.ServeHTTP(w, r.WithContext(newCtx))
			return
		}

		// Try Authorization header (Bearer token)
		if authenticated, newCtx := m.tryBearerToken(r); authenticated {
			next.ServeHTTP(w, r.WithContext(newCtx))
			return
		}

		// Try X-API-Key header
		apiKey := r.Header.Get(apiKeyHeader)
		if apiKey != "" {
			if authenticated, newCtx := m.validateAPIKey(ctx, apiKey); authenticated {
				next.ServeHTTP(w, r.WithContext(newCtx))
				return
			}
			// Invalid API key - return error regardless of requireAuth
			m.writeUnauthorized(w, "invalid_api_key", "invalid API key")
			return
		}

		// No valid authentication found
		if m.requireAuth {
			m.writeUnauthorized(w, "missing_authentication",
				"authentication required - provide X-API-Key header or Authorization: Bearer <token>")
			return
		}

		// Auth not required, continue without context
		next.ServeHTTP(w, r)
	})
}

// hydrateTenantContext injects a minimal tenant context when a tenant/org ID is provided
// via headers or query parameters. This supports lightweight multi-tenant scoping even
// when full authentication is not enforced (e.g., demo or single-tenant modes).
func (m *AuthMiddleware) hydrateTenantContext(r *http.Request) *http.Request {
	ctx := r.Context()

	// Do not override an existing tenant context.
	if tenant, ok := auth.TenantFromContext(ctx); ok && tenant != nil && tenant.ID != "" {
		return r
	}

	tenantID := strings.TrimSpace(r.Header.Get(tenantIDHeader))
	if tenantID == "" {
		tenantID = strings.TrimSpace(r.Header.Get(orgIDHeader))
	}
	if tenantID == "" {
		tenantID = strings.TrimSpace(r.URL.Query().Get("tenant_id"))
	}
	if tenantID == "" {
		tenantID = strings.TrimSpace(r.URL.Query().Get("org_id"))
	}
	if tenantID == "" {
		return r
	}

	// Try to resolve full tenant from the store; otherwise fall back to ID-only context.
	if m.authStore != nil {
		if tenant, err := m.authStore.GetTenant(ctx, tenantID); err == nil && tenant != nil && tenant.IsActive {
			return r.WithContext(auth.WithTenant(ctx, tenant))
		}
	}

	return r.WithContext(auth.WithTenantID(ctx, tenantID))
}

// trySessionCookie attempts to authenticate via session cookie.
func (m *AuthMiddleware) trySessionCookie(r *http.Request) (bool, context.Context) {
	if m.sessionManager == nil {
		return false, nil
	}

	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		return false, nil
	}

	claims, err := m.sessionManager.ParseToken(cookie.Value)
	if err != nil {
		m.logger.Debug("invalid session cookie",
			"error", err.Error(),
			"path", r.URL.Path,
		)
		return false, nil
	}

	ctx := m.contextFromClaims(r.Context(), claims)
	if ctx == r.Context() {
		// Context unchanged means user/tenant validation failed
		return false, nil
	}

	return true, ctx
}

// tryBearerToken attempts to authenticate via Authorization: Bearer header.
func (m *AuthMiddleware) tryBearerToken(r *http.Request) (bool, context.Context) {
	authHeader := r.Header.Get(authorizationHeader)
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return false, nil
	}

	token := strings.TrimPrefix(authHeader, bearerPrefix)
	if token == "" {
		return false, nil
	}

	// Try as JWT session token first
	if m.sessionManager != nil {
		if claims, err := m.sessionManager.ParseToken(token); err == nil {
			ctx := m.contextFromClaims(r.Context(), claims)
			if ctx != r.Context() {
				return true, ctx
			}
		}
	}

	// Try as API key
	return m.validateAPIKey(r.Context(), token)
}

// validateAPIKey validates an API key and returns an authenticated context.
func (m *AuthMiddleware) validateAPIKey(ctx context.Context, rawKey string) (bool, context.Context) {
	if m.authStore == nil {
		return false, nil
	}

	keyHash := auth.HashAPIKey(rawKey)
	key, err := m.authStore.GetAPIKeyByHash(ctx, keyHash)
	if err != nil || key == nil {
		return false, nil
	}

	// Validate key state
	if !key.IsActive {
		m.logger.Warn("attempted use of revoked API key",
			"keyId", key.ID,
			"tenantId", key.TenantID,
		)
		return false, nil
	}

	if key.IsExpired() {
		m.logger.Debug("attempted use of expired API key",
			"keyId", key.ID,
			"tenantId", key.TenantID,
		)
		return false, nil
	}

	// Get and validate tenant
	tenant, err := m.authStore.GetTenant(ctx, key.TenantID)
	if err != nil {
		m.logger.Error("failed to get tenant for API key",
			"keyId", key.ID,
			"tenantId", key.TenantID,
			"error", err.Error(),
		)
		return false, nil
	}

	if !tenant.IsActive {
		m.logger.Warn("attempted use of API key for inactive tenant",
			"keyId", key.ID,
			"tenantId", key.TenantID,
		)
		return false, nil
	}

	// Get user if key is associated with one
	var user *auth.User
	if key.UserID != "" {
		user, err = m.authStore.GetUser(ctx, key.UserID)
		if err != nil {
			m.logger.Warn("failed to get user for API key",
				"keyId", key.ID,
				"userId", key.UserID,
				"error", err.Error(),
			)
			// Continue without user context
		} else if !user.IsActive {
			m.logger.Warn("attempted use of API key for inactive user",
				"keyId", key.ID,
				"userId", key.UserID,
			)
			return false, nil
		}
	}

	// Update last used timestamp (non-blocking)
	go m.updateAPIKeyLastUsed(key.ID)

	// Build authenticated context
	ctx = auth.WithTenant(ctx, tenant)
	ctx = auth.WithAPIKey(ctx, key)
	if user != nil {
		ctx = auth.WithUser(ctx, user)
	}

	return true, ctx
}

// contextFromClaims builds a request context from JWT session claims.
func (m *AuthMiddleware) contextFromClaims(ctx context.Context, claims *auth.SessionClaims) context.Context {
	if m.authStore == nil {
		return ctx
	}

	// Get and validate tenant
	tenant, err := m.authStore.GetTenant(ctx, claims.TenantID)
	if err != nil {
		m.logger.Error("failed to get tenant from claims",
			"tenantId", claims.TenantID,
			"error", err.Error(),
		)
		return ctx
	}

	if !tenant.IsActive {
		m.logger.Debug("tenant from claims is inactive",
			"tenantId", claims.TenantID,
		)
		return ctx
	}

	// Get and validate user
	user, err := m.authStore.GetUser(ctx, claims.UserID)
	if err != nil {
		m.logger.Error("failed to get user from claims",
			"userId", claims.UserID,
			"error", err.Error(),
		)
		return ctx
	}

	if !user.IsActive {
		m.logger.Debug("user from claims is inactive",
			"userId", claims.UserID,
		)
		return ctx
	}

	// Build authenticated context
	ctx = auth.WithTenant(ctx, tenant)
	ctx = auth.WithUser(ctx, user)
	return ctx
}

// updateAPIKeyLastUsed updates the last used timestamp asynchronously.
func (m *AuthMiddleware) updateAPIKeyLastUsed(keyID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := m.authStore.UpdateAPIKeyLastUsed(ctx, keyID); err != nil {
		m.logger.Warn("failed to update API key last used",
			"keyId", keyID,
			"error", err.Error(),
		)
	}
}

// writeUnauthorized writes a 401 Unauthorized response.
func (m *AuthMiddleware) writeUnauthorized(w http.ResponseWriter, code, message string) {
	responders.Unauthorized(w, code, message)
}

// -----------------------------------------------------------------------------
// Scope Authorization Middleware
// -----------------------------------------------------------------------------

// RequireScope returns middleware that checks for a specific API key scope.
// This should be applied after AuthMiddleware.
func (m *AuthMiddleware) RequireScope(scope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key, ok := auth.APIKeyFromContext(r.Context())
			if !ok || key == nil {
				responders.Unauthorized(w, "unauthorized", "authentication required")
				return
			}

			if !key.HasScope(scope) {
				responders.Forbidden(w, "insufficient_scope",
					"required scope: "+scope)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyScope returns middleware that checks for any of the specified scopes.
func (m *AuthMiddleware) RequireAnyScope(scopes ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key, ok := auth.APIKeyFromContext(r.Context())
			if !ok || key == nil {
				responders.Unauthorized(w, "unauthorized", "authentication required")
				return
			}

			for _, scope := range scopes {
				if key.HasScope(scope) {
					next.ServeHTTP(w, r)
					return
				}
			}

			responders.Forbidden(w, "insufficient_scope",
				"required one of scopes: "+strings.Join(scopes, ", "))
		})
	}
}

// RequireAllScopes returns middleware that checks for all of the specified scopes.
func (m *AuthMiddleware) RequireAllScopes(scopes ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key, ok := auth.APIKeyFromContext(r.Context())
			if !ok || key == nil {
				responders.Unauthorized(w, "unauthorized", "authentication required")
				return
			}

			for _, scope := range scopes {
				if !key.HasScope(scope) {
					responders.Forbidden(w, "insufficient_scope",
						"required scope: "+scope)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// -----------------------------------------------------------------------------
// Role Authorization Middleware
// -----------------------------------------------------------------------------

// RequireRole returns middleware that checks for a specific user role.
func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := auth.UserFromContext(r.Context())
			if !ok || user == nil {
				responders.Unauthorized(w, "unauthorized", "authentication required")
				return
			}

			if user.Role != role {
				responders.Forbidden(w, "insufficient_role",
					"required role: "+role)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyRole returns middleware that checks for any of the specified roles.
func RequireAnyRole(roles ...string) func(http.Handler) http.Handler {
	roleSet := make(map[string]bool, len(roles))
	for _, r := range roles {
		roleSet[r] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := auth.UserFromContext(r.Context())
			if !ok || user == nil {
				responders.Unauthorized(w, "unauthorized", "authentication required")
				return
			}

			if !roleSet[user.Role] {
				responders.Forbidden(w, "insufficient_role",
					"required one of roles: "+strings.Join(roles, ", "))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// -----------------------------------------------------------------------------
// Context Helpers
// -----------------------------------------------------------------------------

// GetTenantID extracts tenant ID from request context.
// Returns empty string if no tenant in context.
func GetTenantID(r *http.Request) string {
	tenant, ok := auth.TenantFromContext(r.Context())
	if !ok || tenant == nil {
		return ""
	}
	return tenant.ID
}

// GetOrgID is an alias for GetTenantID (org_id == tenant_id in our model).
func GetOrgID(r *http.Request) string {
	return GetTenantID(r)
}

// GetUserID extracts user ID from request context.
// Returns empty string if no user in context.
func GetUserID(r *http.Request) string {
	user, ok := auth.UserFromContext(r.Context())
	if !ok || user == nil {
		return ""
	}
	return user.ID
}

// MustGetTenantID returns tenant ID or writes error if not authenticated.
func MustGetTenantID(w http.ResponseWriter, r *http.Request) (string, bool) {
	tenantID := GetTenantID(r)
	if tenantID == "" {
		responders.Unauthorized(w, "unauthorized", "authentication required")
		return "", false
	}
	return tenantID, true
}

// MustGetUserID returns user ID or writes error if not authenticated.
func MustGetUserID(w http.ResponseWriter, r *http.Request) (string, bool) {
	userID := GetUserID(r)
	if userID == "" {
		responders.Unauthorized(w, "unauthorized", "authentication required")
		return "", false
	}
	return userID, true
}
