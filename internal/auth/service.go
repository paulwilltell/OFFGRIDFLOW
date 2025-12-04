// Package auth - service.go provides high-level authentication operations.
package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// -----------------------------------------------------------------------------
// Service Configuration
// -----------------------------------------------------------------------------

// ServiceConfig holds configuration for the auth service.
type ServiceConfig struct {
	// DefaultKeyExpiry is the default API key expiration (nil = no expiry).
	DefaultKeyExpiry *time.Duration

	// MaxKeysPerTenant limits the number of API keys per tenant.
	MaxKeysPerTenant int

	// DefaultPlan is the plan assigned to new tenants.
	DefaultPlan string

	// RequireEmailVerification requires email verification for new users.
	RequireEmailVerification bool

	// Logger for auth operations.
	Logger *slog.Logger
}

// DefaultServiceConfig returns a configuration with sensible defaults.
func DefaultServiceConfig() ServiceConfig {
	thirtyDays := 30 * 24 * time.Hour
	return ServiceConfig{
		DefaultKeyExpiry: &thirtyDays,
		MaxKeysPerTenant: 10,
		DefaultPlan:      "free",
		Logger:           slog.Default(),
	}
}

// -----------------------------------------------------------------------------
// Service
// -----------------------------------------------------------------------------

// Service provides high-level authentication and authorization operations.
// It coordinates between the Store (persistence), Authorizer (permissions),
// and other auth components to provide a unified API.
type Service struct {
	store      Store
	authorizer Authorizer
	config     ServiceConfig
	logger     *slog.Logger

	resetTokens map[string]passwordResetToken
	resetMu     sync.Mutex
}

// NewService creates a new auth service with default configuration.
// If authorizer is nil, a default RBAC authorizer is used.
func NewService(store Store, authorizer Authorizer) *Service {
	return NewServiceWithConfig(store, authorizer, DefaultServiceConfig())
}

// NewServiceWithConfig creates a new auth service with custom configuration.
func NewServiceWithConfig(store Store, authorizer Authorizer, config ServiceConfig) *Service {
	if authorizer == nil {
		authorizer = NewRBACAuthorizer()
	}
	if config.Logger == nil {
		config.Logger = slog.Default()
	}

	return &Service{
		store:       store,
		authorizer:  authorizer,
		config:      config,
		logger:      config.Logger.With("component", "auth.Service"),
		resetTokens: make(map[string]passwordResetToken),
	}
}

// -----------------------------------------------------------------------------
// API Key Validation
// -----------------------------------------------------------------------------

// ValidateAPIKey validates an API key and returns the associated tenant, user, and key.
// This is the primary authentication method for API requests.
//
// Returns:
//   - Tenant: The organization the key belongs to
//   - User: The user who created the key (may be nil for tenant-level keys)
//   - APIKey: The validated key metadata
//   - error: ErrInvalidAPIKey if the key is invalid, expired, or revoked
func (s *Service) ValidateAPIKey(ctx context.Context, rawKey string) (*Tenant, *User, *APIKey, error) {
	if rawKey == "" {
		s.logger.DebugContext(ctx, "API key validation failed: empty key")
		return nil, nil, nil, ErrMissingAPIKey
	}

	// Hash the key for lookup
	keyHash := HashAPIKey(rawKey)

	// Look up key by hash
	key, err := s.store.GetAPIKeyByHash(ctx, keyHash)
	if err != nil {
		if errors.Is(err, ErrInvalidAPIKey) {
			s.logger.DebugContext(ctx, "API key validation failed: key not found",
				"key_prefix", safeKeyPrefix(rawKey))
		}
		return nil, nil, nil, ErrInvalidAPIKey
	}

	// Check key status
	if !key.IsActive {
		s.logger.DebugContext(ctx, "API key validation failed: key revoked",
			"key_id", key.ID)
		return nil, nil, nil, ErrInvalidAPIKey
	}

	if key.IsExpired() {
		s.logger.DebugContext(ctx, "API key validation failed: key expired",
			"key_id", key.ID,
			"expired_at", key.ExpiresAt)
		return nil, nil, nil, ErrInvalidAPIKey
	}

	// Get tenant
	tenant, err := s.store.GetTenant(ctx, key.TenantID)
	if err != nil {
		s.logger.ErrorContext(ctx, "API key validation failed: tenant not found",
			"key_id", key.ID,
			"tenant_id", key.TenantID,
			"error", err)
		return nil, nil, nil, err
	}

	if !tenant.IsActive {
		s.logger.WarnContext(ctx, "API key validation failed: tenant inactive",
			"tenant_id", tenant.ID)
		return nil, nil, nil, ErrUnauthorized
	}

	// Get user if associated
	var user *User
	if key.UserID != "" {
		user, err = s.store.GetUser(ctx, key.UserID)
		if err != nil {
			s.logger.WarnContext(ctx, "API key validation: user not found",
				"key_id", key.ID,
				"user_id", key.UserID,
				"error", err)
			// Continue without user - key is still valid
		} else if !user.IsActive {
			s.logger.WarnContext(ctx, "API key validation failed: user inactive",
				"user_id", user.ID)
			return nil, nil, nil, ErrUnauthorized
		}
	}

	// Update last used timestamp (fire and forget)
	go func() {
		bgCtx := context.Background()
		if err := s.store.UpdateAPIKeyLastUsed(bgCtx, key.ID); err != nil {
			s.logger.WarnContext(bgCtx, "Failed to update API key last used",
				"key_id", key.ID,
				"error", err)
		}
	}()

	s.logger.DebugContext(ctx, "API key validated successfully",
		"key_id", key.ID,
		"tenant_id", tenant.ID)

	return tenant, user, key, nil
}

// safeKeyPrefix returns the first 12 characters of a key for logging.
func safeKeyPrefix(key string) string {
	if len(key) > 12 {
		return key[:12] + "..."
	}
	return "[short key]"
}

// -----------------------------------------------------------------------------
// Tenant Management
// -----------------------------------------------------------------------------

// CreateTenant creates a new tenant with an admin user and API key.
// Returns the raw API key which should only be shown once to the user.
//
// Example:
//
// tenant, user, rawKey, err := service.CreateTenant(ctx, "Acme Corp", "admin@acme.com", "Admin User")
//
//	if err != nil {
//	   return err
//	}
//
// fmt.Printf("Your API key: %s (save this, it won't be shown again)\n", rawKey)
func (s *Service) CreateTenant(ctx context.Context, name, adminEmail, adminName string) (*Tenant, *User, string, error) {
	// Validate inputs
	if name == "" {
		return nil, nil, "", errors.New("auth: tenant name is required")
	}
	if adminEmail == "" {
		return nil, nil, "", errors.New("auth: admin email is required")
	}

	s.logger.InfoContext(ctx, "Creating new tenant",
		"name", name,
		"admin_email", adminEmail)

	// Create tenant
	tenant := &Tenant{
		Name:     name,
		Plan:     s.config.DefaultPlan,
		IsActive: true,
	}
	if err := s.store.CreateTenant(ctx, tenant); err != nil {
		s.logger.ErrorContext(ctx, "Failed to create tenant",
			"name", name,
			"error", err)
		return nil, nil, "", fmt.Errorf("auth: failed to create tenant: %w", err)
	}

	// Create admin user
	user := &User{
		Email:    adminEmail,
		Name:     adminName,
		TenantID: tenant.ID,
		Role:     "admin",
		Roles:    []string{"admin"},
		IsActive: true,
	}
	if err := s.store.CreateUser(ctx, user); err != nil {
		s.logger.ErrorContext(ctx, "Failed to create admin user",
			"tenant_id", tenant.ID,
			"email", adminEmail,
			"error", err)
		return nil, nil, "", fmt.Errorf("auth: failed to create admin user: %w", err)
	}

	// Generate API key
	rawKey, apiKey, err := func() (string, *APIKey, error) {
		var expiresAt *time.Time
		if s.config.DefaultKeyExpiry != nil {
			t := time.Now().Add(*s.config.DefaultKeyExpiry)
			expiresAt = &t
		}
		return GenerateAPIKey("live", tenant.ID, user.ID, "Default API Key", []string{"*"}, expiresAt)
	}()
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to generate API key",
			"tenant_id", tenant.ID,
			"error", err)
		return nil, nil, "", fmt.Errorf("auth: failed to generate API key: %w", err)
	}

	if err := s.store.CreateAPIKey(ctx, apiKey); err != nil {
		s.logger.ErrorContext(ctx, "Failed to store API key",
			"tenant_id", tenant.ID,
			"error", err)
		return nil, nil, "", fmt.Errorf("auth: failed to store API key: %w", err)
	}

	s.logger.InfoContext(ctx, "Tenant created successfully",
		"tenant_id", tenant.ID,
		"user_id", user.ID,
		"key_id", apiKey.ID)

	return tenant, user, rawKey, nil
}

// GetTenant retrieves a tenant by ID.
func (s *Service) GetTenant(ctx context.Context, id string) (*Tenant, error) {
	return s.store.GetTenant(ctx, id)
}

// UpdateTenant updates a tenant's information.
func (s *Service) UpdateTenant(ctx context.Context, tenant *Tenant) error {
	if tenant == nil {
		return errors.New("auth: tenant is required")
	}
	return s.store.UpdateTenant(ctx, tenant)
}

// -----------------------------------------------------------------------------
// API Key Management
// -----------------------------------------------------------------------------

// CreateAPIKey creates a new API key for a tenant.
// Returns the raw key (show once) and the API key metadata.
func (s *Service) CreateAPIKey(ctx context.Context, tenantID, userID, name string, scopes []string, expiresIn *time.Duration) (string, *APIKey, error) {
	// Verify tenant exists
	tenant, err := s.store.GetTenant(ctx, tenantID)
	if err != nil {
		return "", nil, err
	}
	if !tenant.IsActive {
		return "", nil, ErrUnauthorized
	}

	// Check key limit
	existingKeys, err := s.store.ListAPIKeysByTenant(ctx, tenantID)
	if err != nil {
		return "", nil, fmt.Errorf("auth: failed to list existing keys: %w", err)
	}
	activeCount := 0
	for _, k := range existingKeys {
		if k.IsActive {
			activeCount++
		}
	}
	if s.config.MaxKeysPerTenant > 0 && activeCount >= s.config.MaxKeysPerTenant {
		return "", nil, fmt.Errorf("auth: tenant has reached maximum API key limit (%d)", s.config.MaxKeysPerTenant)
	}

	// Set expiration
	var expiresAt *time.Time
	if expiresIn != nil {
		t := time.Now().Add(*expiresIn)
		expiresAt = &t
	}

	// Default scopes
	if len(scopes) == 0 {
		scopes = []string{"*"}
	}

	// Generate key
	rawKey, apiKey, err := GenerateAPIKey("live", tenantID, userID, name, scopes, expiresAt)
	if err != nil {
		return "", nil, err
	}

	if err := s.store.CreateAPIKey(ctx, apiKey); err != nil {
		return "", nil, fmt.Errorf("auth: failed to store API key: %w", err)
	}

	s.logger.InfoContext(ctx, "API key created",
		"key_id", apiKey.ID,
		"tenant_id", tenantID,
		"label", name)

	return rawKey, apiKey, nil
}

// RevokeAPIKey revokes an API key, making it permanently unusable.
func (s *Service) RevokeAPIKey(ctx context.Context, keyID string) error {
	if err := s.store.RevokeAPIKey(ctx, keyID); err != nil {
		return fmt.Errorf("auth: failed to revoke API key: %w", err)
	}

	s.logger.InfoContext(ctx, "API key revoked", "key_id", keyID)
	return nil
}

// ListAPIKeys returns all API keys for a tenant.
// Note: Key hashes are not included for security.
func (s *Service) ListAPIKeys(ctx context.Context, tenantID string) ([]*APIKey, error) {
	return s.store.ListAPIKeysByTenant(ctx, tenantID)
}

// -----------------------------------------------------------------------------
// User Management
// -----------------------------------------------------------------------------

// GetUser retrieves a user by ID.
func (s *Service) GetUser(ctx context.Context, id string) (*User, error) {
	return s.store.GetUser(ctx, id)
}

// GetUserByEmail retrieves a user by email address.
func (s *Service) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	return s.store.GetUserByEmail(ctx, email)
}

// CreateUser creates a new user within a tenant.
func (s *Service) CreateUser(ctx context.Context, user *User) error {
	if user == nil {
		return errors.New("auth: user is required")
	}
	if err := user.Validate(); err != nil {
		return err
	}
	return s.store.CreateUser(ctx, user)
}

// UpdateUser updates a user's information.
func (s *Service) UpdateUser(ctx context.Context, user *User) error {
	if user == nil {
		return errors.New("auth: user is required")
	}
	return s.store.UpdateUser(ctx, user)
}

// ListUsersByTenant returns all users for a tenant.
func (s *Service) ListUsersByTenant(ctx context.Context, tenantID string) ([]*User, error) {
	return s.store.ListUsersByTenant(ctx, tenantID)
}

// -----------------------------------------------------------------------------
// Authorization
// -----------------------------------------------------------------------------

// Authorize checks if a user has permission for an action on a resource.
func (s *Service) Authorize(ctx context.Context, user *User, action, resource string) (bool, error) {
	if user == nil {
		return false, ErrUnauthorized
	}
	return s.authorizer.Authorize(ctx, *user, action, resource)
}

// RequirePermission checks authorization and returns ErrForbidden if denied.
// Use this in handlers to enforce access control.
func (s *Service) RequirePermission(ctx context.Context, user *User, action, resource string) error {
	allowed, err := s.Authorize(ctx, user, action, resource)
	if err != nil {
		return err
	}
	if !allowed {
		s.logger.WarnContext(ctx, "Permission denied",
			"user_id", user.ID,
			"action", action,
			"resource", resource)
		return ErrForbidden
	}
	return nil
}

// -----------------------------------------------------------------------------
// Accessors
// -----------------------------------------------------------------------------

// GetStore returns the underlying store for advanced operations.
func (s *Service) GetStore() Store {
	return s.store
}

// GetAuthorizer returns the underlying authorizer.
func (s *Service) GetAuthorizer() Authorizer {
	return s.authorizer
}

// -----------------------------------------------------------------------------
// Password Reset
// -----------------------------------------------------------------------------

type passwordResetToken struct {
	UserID    string
	ExpiresAt time.Time
}

// CreatePasswordResetToken generates a one-time password reset token for a user.
// The token is stored in-memory and expires after one hour.
func (s *Service) CreatePasswordResetToken(ctx context.Context, email string) (string, *User, error) {
	user, err := s.store.GetUserByEmail(ctx, email)
	if err != nil {
		return "", nil, err
	}
	if !user.IsActive {
		return "", nil, ErrUnauthorized
	}

	token := generateResetToken()
	s.resetMu.Lock()
	s.resetTokens[token] = passwordResetToken{
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	s.resetMu.Unlock()
	return token, user, nil
}

// ResetPassword validates a reset token and updates the user's password.
func (s *Service) ResetPassword(ctx context.Context, token, newPassword string) error {
	if token == "" {
		return errors.New("auth: reset token required")
	}
	s.resetMu.Lock()
	entry, ok := s.resetTokens[token]
	if ok && time.Now().After(entry.ExpiresAt) {
		ok = false
		delete(s.resetTokens, token)
	}
	if ok {
		delete(s.resetTokens, token)
	}
	s.resetMu.Unlock()
	if !ok {
		return errors.New("auth: invalid or expired reset token")
	}

	hash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}
	return s.store.UpdateUserPassword(ctx, entry.UserID, hash)
}

func generateResetToken() string {
	raw := make([]byte, 24)
	if _, err := rand.Read(raw); err != nil {
		now := time.Now().UnixNano()
		raw = []byte(fmt.Sprintf("%d", now))
	}
	return hex.EncodeToString(raw)
}
