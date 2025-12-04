// Package auth - session.go provides JWT session management.
package auth

import (
"errors"
"time"

"github.com/golang-jwt/jwt/v5"
)

// -----------------------------------------------------------------------------
// Session Configuration
// -----------------------------------------------------------------------------

// Session configuration constants.
const (
// DefaultSessionTTL is the default token lifetime.
DefaultSessionTTL = 24 * time.Hour

// MinSessionTTL is the minimum allowed token lifetime.
MinSessionTTL = 5 * time.Minute

// MaxSessionTTL is the maximum allowed token lifetime.
MaxSessionTTL = 30 * 24 * time.Hour // 30 days

// RefreshThreshold is the time before expiry when a token should be refreshed.
RefreshThreshold = 1 * time.Hour
)

// Session errors.
var (
// ErrSessionSecretRequired indicates a JWT secret was not provided.
ErrSessionSecretRequired = errors.New("auth: JWT secret is required")

// ErrSessionSecretTooShort indicates the JWT secret is too short.
ErrSessionSecretTooShort = errors.New("auth: JWT secret must be at least 32 characters")

// ErrInvalidToken indicates the token is malformed or signature is invalid.
ErrInvalidToken = errors.New("auth: invalid session token")

// ErrTokenExpired indicates the token has expired.
ErrTokenExpired = errors.New("auth: session token has expired")

// ErrMissingClaims indicates required claims are missing from the token.
ErrMissingClaims = errors.New("auth: missing required claims in token")

// ErrUserTenantRequired indicates user and tenant are required for token creation.
ErrUserTenantRequired = errors.New("auth: user and tenant required for session token")
)

// -----------------------------------------------------------------------------
// Session Claims
// -----------------------------------------------------------------------------

// SessionClaims represents the JWT payload for authenticated users.
// Embeds jwt.RegisteredClaims for standard JWT fields (sub, exp, iat, etc.).
type SessionClaims struct {
// UserID is the authenticated user's ID (same as Subject).
UserID string `json:"user_id"`

// TenantID is the user's organization.
TenantID string `json:"tenant_id"`

// Email is the user's email address.
Email string `json:"email"`

// Role is the user's primary role.
Role string `json:"role"`

// Roles contains all user roles including primary.
Roles []string `json:"roles,omitempty"`

// Standard JWT claims (subject, issuer, expiry, etc.)
jwt.RegisteredClaims
}

// Validate checks if the claims contain all required fields.
func (c *SessionClaims) Validate() error {
if c.UserID == "" {
return ErrMissingClaims
}
if c.TenantID == "" {
return ErrMissingClaims
}
return nil
}

// ShouldRefresh returns true if the token is close to expiry.
func (c *SessionClaims) ShouldRefresh() bool {
if c.ExpiresAt == nil {
return false
}
return time.Until(c.ExpiresAt.Time) < RefreshThreshold
}

// TimeUntilExpiry returns the duration until the token expires.
func (c *SessionClaims) TimeUntilExpiry() time.Duration {
if c.ExpiresAt == nil {
return 0
}
return time.Until(c.ExpiresAt.Time)
}

// -----------------------------------------------------------------------------
// Session Manager
// -----------------------------------------------------------------------------

// SessionManager issues and validates signed JWT session tokens.
// It is safe for concurrent use.
type SessionManager struct {
secret []byte
ttl    time.Duration
issuer string
}

// SessionManagerOption configures a SessionManager.
type SessionManagerOption func(*SessionManager)

// WithTTL sets the token lifetime.
func WithTTL(ttl time.Duration) SessionManagerOption {
return func(m *SessionManager) {
if ttl >= MinSessionTTL && ttl <= MaxSessionTTL {
m.ttl = ttl
}
}
}

// WithIssuer sets the token issuer claim.
func WithIssuer(issuer string) SessionManagerOption {
return func(m *SessionManager) {
m.issuer = issuer
}
}

// NewSessionManager creates a session manager with the provided secret.
// The secret must be at least 32 characters for security.
func NewSessionManager(secret string, opts ...SessionManagerOption) (*SessionManager, error) {
if secret == "" {
return nil, ErrSessionSecretRequired
}
if len(secret) < 32 {
return nil, ErrSessionSecretTooShort
}

m := &SessionManager{
secret: []byte(secret),
ttl:    DefaultSessionTTL,
issuer: "offgridflow",
}

for _, opt := range opts {
opt(m)
}

return m, nil
}

// SetTTL updates the token lifetime for new tokens.
func (m *SessionManager) SetTTL(ttl time.Duration) {
if ttl >= MinSessionTTL && ttl <= MaxSessionTTL {
m.ttl = ttl
}
}

// GetTTL returns the current token lifetime.
func (m *SessionManager) GetTTL() time.Duration {
return m.ttl
}

// CreateToken issues a signed JWT for the provided user and tenant.
func (m *SessionManager) CreateToken(user *User, tenant *Tenant) (string, error) {
if user == nil || tenant == nil {
return "", ErrUserTenantRequired
}

now := time.Now()
claims := SessionClaims{
UserID:   user.ID,
TenantID: tenant.ID,
Email:    user.Email,
Role:     user.Role,
Roles:    user.AllRoles(),
RegisteredClaims: jwt.RegisteredClaims{
Subject:   user.ID,
Issuer:    m.issuer,
IssuedAt:  jwt.NewNumericDate(now),
ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
NotBefore: jwt.NewNumericDate(now),
},
}

token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
signed, err := token.SignedString(m.secret)
if err != nil {
return "", err
}

return signed, nil
}

// CreateTokenWithClaims issues a token with custom claims.
func (m *SessionManager) CreateTokenWithClaims(claims SessionClaims) (string, error) {
if err := claims.Validate(); err != nil {
return "", err
}

now := time.Now()
if claims.IssuedAt == nil {
claims.IssuedAt = jwt.NewNumericDate(now)
}
if claims.ExpiresAt == nil {
claims.ExpiresAt = jwt.NewNumericDate(now.Add(m.ttl))
}
if claims.Issuer == "" {
claims.Issuer = m.issuer
}
if claims.Subject == "" {
claims.Subject = claims.UserID
}

token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
return token.SignedString(m.secret)
}

// ParseToken validates a JWT string and returns the session claims.
func (m *SessionManager) ParseToken(tokenString string) (*SessionClaims, error) {
if tokenString == "" {
return nil, ErrInvalidToken
}

parsed, err := jwt.ParseWithClaims(tokenString, &SessionClaims{}, func(token *jwt.Token) (interface{}, error) {
if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
return nil, ErrInvalidToken
}
return m.secret, nil
})

if err != nil {
if errors.Is(err, jwt.ErrTokenExpired) {
return nil, ErrTokenExpired
}
return nil, ErrInvalidToken
}

claims, ok := parsed.Claims.(*SessionClaims)
if !ok || !parsed.Valid {
return nil, ErrInvalidToken
}

if err := claims.Validate(); err != nil {
return nil, err
}

return claims, nil
}

// RefreshToken issues a new token with extended expiry.
func (m *SessionManager) RefreshToken(tokenString string) (string, error) {
claims, err := m.ParseToken(tokenString)
if err != nil {
return "", err
}

now := time.Now()
claims.IssuedAt = jwt.NewNumericDate(now)
claims.ExpiresAt = jwt.NewNumericDate(now.Add(m.ttl))

token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
return token.SignedString(m.secret)
}

// ValidateToken checks if a token is valid without returning claims.
func (m *SessionManager) ValidateToken(tokenString string) error {
_, err := m.ParseToken(tokenString)
return err
}