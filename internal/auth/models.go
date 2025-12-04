// Package auth provides authentication, authorization, and multi-tenancy
// for OffGridFlow. It supports both API key authentication (for programmatic
// access) and JWT session authentication (for web applications).
//
// **Cleanly matching frontend↔backend auth flows:**
// - Next.js sessions and API tokens share the same JWT claims structure
// - Login, refresh, and logout flows enforce consistent state across layers
// - Role-based access control (RBAC) contracts are identical in web and API
//
// The package implements a multi-tenant model where:
//   - Tenants are isolated organizations with separate data and configurations
//   - Users belong to tenants and have role-based access control
//   - API keys provide scoped programmatic access with optional expiration
//
// # Context Helpers
//
// Authentication data is propagated through context.Context using type-safe
// helper functions:
//
//ctx = auth.WithTenant(ctx, tenant)
//ctx = auth.WithUser(ctx, user)
//tenant, ok := auth.TenantFromContext(ctx)
//
// # API Key Format
//
// API keys follow the format: ogf_{env}_{random_hex}
// where env is 'live', 'test', or 'dev', and random_hex is 64 characters.
// Example: ogf_live_a1b2c3d4e5f6...
package auth

import (
"context"
"crypto/rand"
"crypto/sha256"
"encoding/hex"
"errors"
"fmt"
"strings"
"time"
)

// -----------------------------------------------------------------------------
// Context Keys
// -----------------------------------------------------------------------------

// contextKey is an unexported type for context keys to prevent collisions
// with keys defined in other packages.
type contextKey string

const (
// ContextKeyTenant is the context key for the authenticated Tenant.
ContextKeyTenant contextKey = "tenant"
// ContextKeyUser is the context key for the authenticated User.
ContextKeyUser contextKey = "user"
// ContextKeyAPIKey is the context key for the validated APIKey.
ContextKeyAPIKey contextKey = "apiKey"
)

// -----------------------------------------------------------------------------
// Sentinel Errors
// -----------------------------------------------------------------------------

var (
// ErrInvalidAPIKey indicates the API key is missing, malformed, revoked, or expired.
ErrInvalidAPIKey = errors.New("auth: invalid or expired API key")

// ErrMissingAPIKey indicates no API key was provided in the request.
ErrMissingAPIKey = errors.New("auth: missing API key")

// ErrTenantNotFound indicates the requested tenant does not exist.
ErrTenantNotFound = errors.New("auth: tenant not found")

// ErrUserNotFound indicates the requested user does not exist.
ErrUserNotFound = errors.New("auth: user not found")

// ErrUnauthorized indicates the request lacks valid authentication.
ErrUnauthorized = errors.New("auth: unauthorized")

// ErrForbidden indicates the authenticated user lacks permission.
ErrForbidden = errors.New("auth: forbidden")

// ErrInvalidKeyScope indicates an API key lacks the required scope.
ErrInvalidKeyScope = errors.New("auth: API key does not have required scope")

// ErrInvalidPassword indicates password verification failed.
ErrInvalidPassword = errors.New("auth: invalid password")

// ErrKeyGenerationFailed indicates cryptographic key generation failed.
ErrKeyGenerationFailed = errors.New("auth: failed to generate secure key")
)

// -----------------------------------------------------------------------------
// Domain Models
// -----------------------------------------------------------------------------

// Tenant represents an OffGridFlow organization (multi-tenant isolation unit).
// All users, API keys, and data are scoped to a single tenant.
type Tenant struct {
ID        string    `json:"id"`
Name      string    `json:"name"`
Slug      string    `json:"slug"`
Plan      string    `json:"plan"` // free, pro, enterprise
IsActive  bool      `json:"is_active"`
CreatedAt time.Time `json:"created_at"`
UpdatedAt time.Time `json:"updated_at"`
}

// Validate checks if the tenant has valid required fields.
func (t *Tenant) Validate() error {
if t.Name == "" {
return errors.New("auth: tenant name is required")
}
if t.Plan == "" {
t.Plan = "free" // Default plan
}
validPlans := map[string]bool{"free": true, "pro": true, "enterprise": true}
if !validPlans[t.Plan] {
return fmt.Errorf("auth: invalid plan %q", t.Plan)
}
return nil
}

// User represents a member of a tenant with role-based permissions.
type User struct {
ID           string     `json:"id"`
Email        string     `json:"email"`
Name         string     `json:"name"`
PasswordHash string     `json:"-"` // Never serialize
TenantID     string     `json:"tenant_id"`
Role         string     `json:"role"`  // Primary: admin, editor, viewer
Roles        []string   `json:"roles"` // Additional roles for fine-grained access
IsActive     bool       `json:"is_active"`
LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
CreatedAt    time.Time  `json:"created_at"`
UpdatedAt    time.Time  `json:"updated_at"`
}

// AllRoles returns the primary role plus any additional roles.
func (u *User) AllRoles() []string {
roles := make([]string, 0, 1+len(u.Roles))
if u.Role != "" {
roles = append(roles, u.Role)
}
roles = append(roles, u.Roles...)
return roles
}

// HasRole checks if the user has a specific role.
func (u *User) HasRole(role string) bool {
if u.Role == role {
return true
}
for _, r := range u.Roles {
if r == role {
return true
}
}
return false
}

// IsAdmin returns true if the user has admin privileges.
func (u *User) IsAdmin() bool {
return u.HasRole("admin")
}

// Validate checks if the user has valid required fields.
func (u *User) Validate() error {
if u.Email == "" {
return errors.New("auth: user email is required")
}
if u.TenantID == "" {
return errors.New("auth: user must belong to a tenant")
}
if u.Role == "" {
u.Role = "viewer" // Default role
}
return nil
}

// APIKey represents an API key used for programmatic access.
// The actual key is only shown once at creation time.
type APIKey struct {
ID         string     `json:"id"`
KeyHash    string     `json:"-"`                        // Never serialize - internal use only
KeyPrefix  string     `json:"key_prefix"`               // First 12 chars for identification
Label      string     `json:"label"`
TenantID   string     `json:"tenant_id"`
UserID     string     `json:"user_id,omitempty"`
Scopes     []string   `json:"scopes"`
ExpiresAt  *time.Time `json:"expires_at,omitempty"`
LastUsedAt *time.Time `json:"last_used_at,omitempty"`
IsActive   bool       `json:"is_active"`
CreatedAt  time.Time  `json:"created_at"`
}

// HasScope checks whether the API key authorizes the given scope.
// Wildcard "*" and "admin" scopes grant access to everything.
func (k *APIKey) HasScope(scope string) bool {
for _, s := range k.Scopes {
if s == "*" || s == "admin" {
return true
}
// Exact match
if s == scope {
return true
}
// Prefix match: "emissions" matches "emissions:read"
if strings.HasPrefix(scope, s+":") {
return true
}
}
return false
}

// IsExpired reports if the API key has passed its expiration date.
func (k *APIKey) IsExpired() bool {
if k.ExpiresAt == nil {
return false
}
return time.Now().After(*k.ExpiresAt)
}

// IsValid returns true if the key is active and not expired.
func (k *APIKey) IsValid() bool {
return k.IsActive && !k.IsExpired()
}

// DaysUntilExpiry returns the number of days until expiration, or -1 if never.
func (k *APIKey) DaysUntilExpiry() int {
if k.ExpiresAt == nil {
return -1
}
d := time.Until(*k.ExpiresAt)
if d < 0 {
return 0
}
return int(d.Hours() / 24)
}

// Validate checks if the API key has valid required fields.
func (k *APIKey) Validate() error {
if k.TenantID == "" {
return errors.New("auth: API key must belong to a tenant")
}
if k.Label == "" {
return errors.New("auth: API key label is required")
}
if len(k.Scopes) == 0 {
return errors.New("auth: API key must have at least one scope")
}
return nil
}

// -----------------------------------------------------------------------------
// Key Generation
// -----------------------------------------------------------------------------

// APIKeyLength is the number of random bytes used in key generation (32 = 256 bits).
const APIKeyLength = 32

// ValidEnvironments are the allowed API key environment prefixes.
var ValidEnvironments = map[string]bool{
"live": true,
"test": true,
"dev":  true,
}

// GenerateAPIKey creates a new API key with cryptographically secure randomness.
// Returns the plaintext key (show once to user) and the APIKey metadata for storage.
//
// Key format: ogf_{env}_{64_hex_chars}
// Example: ogf_live_a1b2c3d4e5f6789...
func GenerateAPIKey(env, tenantID, userID, label string, scopes []string, expiresAt *time.Time) (string, *APIKey, error) {
if !ValidEnvironments[env] {
return "", nil, fmt.Errorf("auth: invalid environment %q, must be live/test/dev", env)
}
if tenantID == "" {
return "", nil, errors.New("auth: tenant ID is required for API key")
}
if len(scopes) == 0 {
scopes = []string{"*"} // Default to full access
}

// Generate 32 bytes (256 bits) of randomness
randomBytes := make([]byte, APIKeyLength)
if _, err := rand.Read(randomBytes); err != nil {
return "", nil, fmt.Errorf("%w: %v", ErrKeyGenerationFailed, err)
}

// Build the full key: ogf_{env}_{hex}
prefix := "ogf_" + env + "_"
randomHex := hex.EncodeToString(randomBytes)
fullKey := prefix + randomHex

// Hash for storage (never store plaintext)
hash := sha256.Sum256([]byte(fullKey))
keyHash := hex.EncodeToString(hash[:])

apiKey := &APIKey{
KeyHash:   keyHash,
KeyPrefix: fullKey[:12], // "ogf_live_a1b2" for identification
Label:     label,
TenantID:  tenantID,
UserID:    userID,
Scopes:    scopes,
ExpiresAt: expiresAt,
IsActive:  true,
CreatedAt: time.Now(),
}

return fullKey, apiKey, nil
}

// HashAPIKey computes SHA-256 hash of a plaintext key for secure storage/lookup.
func HashAPIKey(rawKey string) string {
hash := sha256.Sum256([]byte(rawKey))
return hex.EncodeToString(hash[:])
}

// ParseAPIKeyEnv extracts the environment from an API key prefix.
// Returns empty string if the key format is invalid.
func ParseAPIKeyEnv(rawKey string) string {
if !strings.HasPrefix(rawKey, "ogf_") {
return ""
}
parts := strings.SplitN(rawKey, "_", 3)
if len(parts) < 3 {
return ""
}
if ValidEnvironments[parts[1]] {
return parts[1]
}
return ""
}

// -----------------------------------------------------------------------------
// Context Helpers
// -----------------------------------------------------------------------------

// WithTenant adds tenant info to context.
func WithTenant(ctx context.Context, tenant *Tenant) context.Context {
if tenant == nil {
return ctx
}
return context.WithValue(ctx, ContextKeyTenant, tenant)
}

// WithUser adds user info to context.
func WithUser(ctx context.Context, user *User) context.Context {
if user == nil {
return ctx
}
return context.WithValue(ctx, ContextKeyUser, user)
}

// WithAPIKey adds API key info to context.
func WithAPIKey(ctx context.Context, key *APIKey) context.Context {
if key == nil {
return ctx
}
return context.WithValue(ctx, ContextKeyAPIKey, key)
}

// TenantFromContext extracts tenant info from context.
func TenantFromContext(ctx context.Context) (*Tenant, bool) {
tenant, ok := ctx.Value(ContextKeyTenant).(*Tenant)
return tenant, ok && tenant != nil
}

// TenantIDFromContext returns the tenant ID if present, otherwise empty string.
// This is useful in lightweight multi-tenant scenarios where only the ID is needed.
func TenantIDFromContext(ctx context.Context) string {
tenant, ok := TenantFromContext(ctx)
if !ok || tenant == nil {
return ""
}
return tenant.ID
}

// WithTenantID injects a minimal tenant context when only the ID is known.
// It avoids overwriting an existing tenant already set on the context.
func WithTenantID(ctx context.Context, tenantID string) context.Context {
if tenantID == "" {
return ctx
}
if tenant, ok := TenantFromContext(ctx); ok && tenant != nil && tenant.ID != "" {
return ctx
}
return WithTenant(ctx, &Tenant{ID: tenantID})
}

// UserFromContext extracts user info from context.
func UserFromContext(ctx context.Context) (*User, bool) {
user, ok := ctx.Value(ContextKeyUser).(*User)
return user, ok && user != nil
}

// APIKeyFromContext extracts API key info from context.
func APIKeyFromContext(ctx context.Context) (*APIKey, bool) {
key, ok := ctx.Value(ContextKeyAPIKey).(*APIKey)
return key, ok && key != nil
}

// MustTenantFromContext extracts tenant or panics (for use when auth is guaranteed).
func MustTenantFromContext(ctx context.Context) *Tenant {
tenant, ok := TenantFromContext(ctx)
if !ok {
panic("auth: tenant not in context")
}
return tenant
}

// MustUserFromContext extracts user or panics (for use when auth is guaranteed).
func MustUserFromContext(ctx context.Context) *User {
user, ok := UserFromContext(ctx)
if !ok {
panic("auth: user not in context")
}
return user
}
