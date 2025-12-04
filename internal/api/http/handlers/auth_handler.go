package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/example/offgridflow/internal/api/http/responders"
	"github.com/example/offgridflow/internal/auth"
)

// -----------------------------------------------------------------------------
// API Key Handler Types
// -----------------------------------------------------------------------------

// APIKeyResponse represents an API key in API responses (without sensitive data).
type APIKeyResponse struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	KeyPrefix  string     `json:"keyPrefix"`           // First few chars for identification
	Scopes     []string   `json:"scopes"`              // Granted permissions
	ExpiresAt  *time.Time `json:"expiresAt,omitempty"` // Expiration time (nil = never)
	LastUsedAt *time.Time `json:"lastUsedAt,omitempty"`
	IsActive   bool       `json:"isActive"`
	CreatedAt  time.Time  `json:"createdAt"`
}

// CreateAPIKeyRequest represents a request to create an API key.
type CreateAPIKeyRequest struct {
	Name   string   `json:"name"`             // Display name for the key
	Scopes []string `json:"scopes,omitempty"` // Requested permissions
	// ExpiresIn is an optional duration string (e.g. "720h" for 30 days).
	ExpiresIn string `json:"expiresIn,omitempty"`
}

// CreateAPIKeyResponse includes the raw key (only shown once).
type CreateAPIKeyResponse struct {
	// Key is the full API key. It is only returned once and never stored in plaintext.
	Key     string         `json:"key"`
	Details APIKeyResponse `json:"details"`
}

// -----------------------------------------------------------------------------
// API Key Handler
// -----------------------------------------------------------------------------

// APIKeyHandler manages API keys for a tenant.
type APIKeyHandler struct {
	authService *auth.Service
	logger      *slog.Logger
}

// NewAPIKeyHandlerWithService creates an API key handler with the given auth service.
func NewAPIKeyHandlerWithService(authService *auth.Service, logger *slog.Logger) *APIKeyHandler {
	if logger == nil {
		logger = slog.Default().With("component", "api-key-handler")
	}
	return &APIKeyHandler{
		authService: authService,
		logger:      logger,
	}
}

// NewAPIKeyHandler creates a handler for API key management (legacy interface).
//
//   - GET  /api/auth/keys : List API keys for the authenticated tenant
//   - POST /api/auth/keys : Create a new API key
func NewAPIKeyHandler(authService *auth.Service) http.HandlerFunc {
	handler := NewAPIKeyHandlerWithService(authService, nil)
	return handler.ServeHTTP
}

// ServeHTTP handles API key requests.
func (h *APIKeyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tenant, ok := auth.TenantFromContext(r.Context())
	if !ok || tenant == nil {
		responders.Unauthorized(w, "unauthorized", "authentication required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.listAPIKeys(w, r, tenant.ID)
	case http.MethodPost:
		h.createAPIKey(w, r, tenant.ID)
	case http.MethodDelete:
		h.revokeAPIKey(w, r, tenant.ID)
	default:
		responders.MethodNotAllowed(w, http.MethodGet, http.MethodPost, http.MethodDelete)
	}
}

// listAPIKeys returns all API keys for the tenant.
func (h *APIKeyHandler) listAPIKeys(w http.ResponseWriter, r *http.Request, tenantID string) {
	keys, err := h.authService.GetStore().ListAPIKeysByTenant(r.Context(), tenantID)
	if err != nil {
		h.logger.Error("failed to list API keys",
			"tenantId", tenantID,
			"error", err.Error(),
		)
		responders.InternalError(w, "failed to list API keys")
		return
	}

	response := make([]APIKeyResponse, 0, len(keys))
	for _, k := range keys {
		response = append(response, apiKeyToResponse(k))
	}

	responders.JSON(w, http.StatusOK, response)
}

// createAPIKey creates a new API key for the tenant.
func (h *APIKeyHandler) createAPIKey(w http.ResponseWriter, r *http.Request, tenantID string) {
	defer r.Body.Close()

	var req CreateAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responders.BadRequest(w, "invalid_request", "invalid request body")
		return
	}

	// Validate request
	if errs := validateAPIKeyRequest(req); len(errs) > 0 {
		responders.ValidationErrors(w, errs)
		return
	}

	// Parse expiration if provided
	var expiresIn *time.Duration
	if req.ExpiresIn != "" {
		d, err := time.ParseDuration(req.ExpiresIn)
		if err != nil {
			responders.BadRequest(w, "invalid_request", "invalid expiresIn format - use Go duration (e.g., 720h)")
			return
		}
		if d < time.Hour {
			responders.BadRequest(w, "invalid_request", "expiresIn must be at least 1 hour")
			return
		}
		if d > 365*24*time.Hour {
			responders.BadRequest(w, "invalid_request", "expiresIn must be at most 1 year")
			return
		}
		expiresIn = &d
	}

	// Default scopes if not provided
	scopes := req.Scopes
	if len(scopes) == 0 {
		scopes = []string{"read:emissions", "read:compliance"}
	}

	// Validate scopes
	if errs := validateAPIKeyScopes(scopes); len(errs) > 0 {
		responders.ValidationErrors(w, errs)
		return
	}

	// Get user ID from context if available
	var userID string
	if user, ok := auth.UserFromContext(r.Context()); ok && user != nil {
		userID = user.ID
	}

	rawKey, apiKey, err := h.authService.CreateAPIKey(r.Context(), tenantID, userID, req.Name, scopes, expiresIn)
	if err != nil {
		h.logger.Error("failed to create API key",
			"tenantId", tenantID,
			"error", err.Error(),
		)
		responders.InternalError(w, "failed to create API key")
		return
	}

	h.logger.Info("API key created",
		"tenantId", tenantID,
		"keyId", apiKey.ID,
		"scopes", scopes,
	)

	responders.Created(w, CreateAPIKeyResponse{
		Key:     rawKey,
		Details: apiKeyToResponse(apiKey),
	})
}

// revokeAPIKey revokes an API key.
func (h *APIKeyHandler) revokeAPIKey(w http.ResponseWriter, r *http.Request, tenantID string) {
	keyID := r.URL.Query().Get("id")
	if keyID == "" {
		responders.BadRequest(w, "invalid_request", "id query parameter is required")
		return
	}

	// Get tenant's keys and verify ownership
	keys, err := h.authService.GetStore().ListAPIKeysByTenant(r.Context(), tenantID)
	if err != nil {
		h.logger.Error("failed to list API keys",
			"tenantId", tenantID,
			"error", err.Error(),
		)
		responders.InternalError(w, "failed to verify API key ownership")
		return
	}

	// Find the key and verify it belongs to this tenant
	var key *auth.APIKey
	for _, k := range keys {
		if k.ID == keyID {
			key = k
			break
		}
	}

	if key == nil {
		responders.NotFound(w, "API key")
		return
	}

	if err := h.authService.RevokeAPIKey(r.Context(), keyID); err != nil {
		h.logger.Error("failed to revoke API key",
			"keyId", keyID,
			"error", err.Error(),
		)
		responders.InternalError(w, "failed to revoke API key")
		return
	}

	h.logger.Info("API key revoked",
		"tenantId", tenantID,
		"keyId", keyID,
	)

	responders.NoContent(w)
}

// -----------------------------------------------------------------------------
// Me Handler (Extended)
// -----------------------------------------------------------------------------

// MeResponse represents the current authenticated context.
type MeResponse struct {
	Tenant *TenantInfo `json:"tenant"`
	User   *UserInfo   `json:"user,omitempty"`
	APIKey *KeyInfo    `json:"apiKey,omitempty"`
}

// TenantInfo is tenant info in API responses.
type TenantInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Plan string `json:"plan"`
}

// UserInfo is user info in API responses.
type UserInfo struct {
	ID    string   `json:"id"`
	Email string   `json:"email"`
	Name  string   `json:"name"`
	Roles []string `json:"roles"`
}

// KeyInfo is API key info in API responses.
type KeyInfo struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	KeyPrefix string   `json:"keyPrefix"`
	Scopes    []string `json:"scopes"`
}

// NewMeHandler creates a handler for GET /api/auth/me.
// This extended variant includes API key context when present.
func NewMeHandler() http.HandlerFunc {
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

		response := MeResponse{
			Tenant: &TenantInfo{
				ID:   tenant.ID,
				Name: tenant.Name,
				Plan: tenant.Plan,
			},
		}

		if user, ok := auth.UserFromContext(r.Context()); ok && user != nil {
			response.User = &UserInfo{
				ID:    user.ID,
				Email: user.Email,
				Name:  user.Name,
				Roles: user.Roles,
			}
		}

		if key, ok := auth.APIKeyFromContext(r.Context()); ok && key != nil {
			response.APIKey = &KeyInfo{
				ID:        key.ID,
				Name:      key.Label,
				KeyPrefix: key.KeyPrefix,
				Scopes:    key.Scopes,
			}
		}

		responders.JSON(w, http.StatusOK, response)
	}
}

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

// apiKeyToResponse converts an auth.APIKey to an APIKeyResponse.
func apiKeyToResponse(k *auth.APIKey) APIKeyResponse {
	return APIKeyResponse{
		ID:         k.ID,
		Name:       k.Label,
		KeyPrefix:  k.KeyPrefix,
		Scopes:     k.Scopes,
		ExpiresAt:  k.ExpiresAt,
		LastUsedAt: k.LastUsedAt,
		IsActive:   k.IsActive,
		CreatedAt:  k.CreatedAt,
	}
}

// validateAPIKeyRequest validates a CreateAPIKeyRequest.
func validateAPIKeyRequest(req CreateAPIKeyRequest) []responders.ValidationError {
	var errs []responders.ValidationError

	if req.Name == "" {
		errs = append(errs, responders.ValidationError{
			Field:   "name",
			Message: "name is required",
		})
	} else if len(req.Name) > 100 {
		errs = append(errs, responders.ValidationError{
			Field:   "name",
			Message: "name must be at most 100 characters",
		})
	}

	return errs
}

// allowedScopes defines valid API key scopes.
var allowedScopes = map[string]bool{
	// Emissions
	"read:emissions":  true,
	"write:emissions": true,

	// Compliance
	"read:compliance":  true,
	"write:compliance": true,

	// Activities
	"read:activities":  true,
	"write:activities": true,

	// Reports
	"read:reports":   true,
	"export:reports": true,

	// Organization management
	"manage:users":    true,
	"manage:settings": true,

	// Full access
	"admin": true,
}

// validateAPIKeyScopes validates requested scopes.
func validateAPIKeyScopes(scopes []string) []responders.ValidationError {
	var errs []responders.ValidationError

	for _, scope := range scopes {
		if !allowedScopes[scope] {
			errs = append(errs, responders.ValidationError{
				Field:   "scopes",
				Message: "invalid scope: " + scope,
			})
		}
	}

	return errs
}
