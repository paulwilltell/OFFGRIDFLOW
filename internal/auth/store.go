// Package auth - store.go provides persistence for auth entities.
package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// -----------------------------------------------------------------------------
// Store Interface
// -----------------------------------------------------------------------------

// Store defines persistence operations for auth entities.
// Implementations must be safe for concurrent use.
type Store interface {
	// Tenant operations
	CreateTenant(ctx context.Context, tenant *Tenant) error
	GetTenant(ctx context.Context, id string) (*Tenant, error)
	GetTenantByName(ctx context.Context, name string) (*Tenant, error)
	UpdateTenant(ctx context.Context, tenant *Tenant) error
	DeleteTenant(ctx context.Context, id string) error
	ListTenants(ctx context.Context) ([]*Tenant, error)

	// User operations
	CreateUser(ctx context.Context, user *User) error
	GetUser(ctx context.Context, id string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByVerificationToken(ctx context.Context, token string) (*User, error)
	ListUsersByTenant(ctx context.Context, tenantID string) ([]*User, error)
	UpdateUser(ctx context.Context, user *User) error
	UpdateUserLastLogin(ctx context.Context, id string) error
	UpdateUserPassword(ctx context.Context, id, passwordHash string) error
	DeleteUser(ctx context.Context, id string) error // API Key operations
	CreateAPIKey(ctx context.Context, key *APIKey) error
	GetAPIKey(ctx context.Context, id string) (*APIKey, error)
	GetAPIKeyByHash(ctx context.Context, hash string) (*APIKey, error)
	ListAPIKeysByTenant(ctx context.Context, tenantID string) ([]*APIKey, error)
	UpdateAPIKey(ctx context.Context, key *APIKey) error
	UpdateAPIKeyLastUsed(ctx context.Context, id string) error
	RevokeAPIKey(ctx context.Context, id string) error
	DeleteAPIKey(ctx context.Context, id string) error
	CountAPIKeysByTenant(ctx context.Context, tenantID string) (int, error)
}

// -----------------------------------------------------------------------------
// PostgreSQL Store
// -----------------------------------------------------------------------------

// PostgresStore implements Store using PostgreSQL.
type PostgresStore struct {
	db *sql.DB
}

// NewPostgresStore creates a new PostgreSQL-backed auth store.
func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

// CreateTenant inserts a new tenant record.
func (s *PostgresStore) CreateTenant(ctx context.Context, tenant *Tenant) error {
	if err := tenant.Validate(); err != nil {
		return err
	}

	query := `
INSERT INTO tenants (id, name, slug, plan, is_active, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
`
	_, err := s.db.ExecContext(ctx, query,
		tenant.ID,
		tenant.Name,
		tenant.Slug,
		tenant.Plan,
		tenant.IsActive,
		tenant.CreatedAt,
		tenant.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("auth: failed to create tenant: %w", err)
	}
	return nil
}

// GetTenant retrieves a tenant by ID.
func (s *PostgresStore) GetTenant(ctx context.Context, id string) (*Tenant, error) {
	query := `
SELECT id, name, slug, plan, is_active, created_at, updated_at
FROM tenants WHERE id = $1
`
	var tenant Tenant
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.Slug,
		&tenant.Plan,
		&tenant.IsActive,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrTenantNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("auth: failed to get tenant: %w", err)
	}
	return &tenant, nil
}

// GetTenantByName retrieves a tenant by name.
func (s *PostgresStore) GetTenantByName(ctx context.Context, name string) (*Tenant, error) {
	query := `
SELECT id, name, slug, plan, is_active, created_at, updated_at
FROM tenants WHERE name = $1
`
	var tenant Tenant
	err := s.db.QueryRowContext(ctx, query, name).Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.Slug,
		&tenant.Plan,
		&tenant.IsActive,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrTenantNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("auth: failed to get tenant by name: %w", err)
	}
	return &tenant, nil
}

// UpdateTenant updates an existing tenant record.
func (s *PostgresStore) UpdateTenant(ctx context.Context, tenant *Tenant) error {
	if err := tenant.Validate(); err != nil {
		return err
	}

	query := `
UPDATE tenants SET name = $2, slug = $3, plan = $4, is_active = $5, updated_at = $6
WHERE id = $1
`
	result, err := s.db.ExecContext(ctx, query,
		tenant.ID,
		tenant.Name,
		tenant.Slug,
		tenant.Plan,
		tenant.IsActive,
		tenant.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("auth: failed to update tenant: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrTenantNotFound
	}
	return nil
}

// DeleteTenant removes a tenant by ID.
func (s *PostgresStore) DeleteTenant(ctx context.Context, id string) error {
	query := `DELETE FROM tenants WHERE id = $1`
	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("auth: failed to delete tenant: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrTenantNotFound
	}
	return nil
}

// ListTenants returns all tenants.
func (s *PostgresStore) ListTenants(ctx context.Context) ([]*Tenant, error) {
	query := `
SELECT id, name, slug, plan, is_active, created_at, updated_at
FROM tenants ORDER BY name
`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("auth: failed to list tenants: %w", err)
	}
	defer rows.Close()

	var tenants []*Tenant
	for rows.Next() {
		var t Tenant
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug, &t.Plan, &t.IsActive, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("auth: failed to scan tenant: %w", err)
		}
		tenants = append(tenants, &t)
	}
	return tenants, rows.Err()
}

// CreateUser inserts a new user record.
func (s *PostgresStore) CreateUser(ctx context.Context, user *User) error {
	if err := user.Validate(); err != nil {
		return err
	}

	query := `
INSERT INTO users (id, tenant_id, email, name, password_hash, role, roles, is_active, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, COALESCE($6, 'viewer'), COALESCE($7, 'viewer'), $8, $9, $10)
`
	_, err := s.db.ExecContext(ctx, query,
		user.ID,
		user.TenantID,
		user.Email,
		user.Name,
		user.PasswordHash,
		user.Role,
		user.Roles,
		user.IsActive,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("auth: failed to create user: %w", err)
	}
	return nil
}

// GetUser retrieves a user by ID.
func (s *PostgresStore) GetUser(ctx context.Context, id string) (*User, error) {
	query := `
SELECT id, tenant_id, email, name, password_hash, role, roles, is_active, last_login_at, created_at, updated_at
FROM users WHERE id = $1
`
	return s.scanUser(s.db.QueryRowContext(ctx, query, id))
}

// GetUserByEmail retrieves a user by email.
func (s *PostgresStore) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `
SELECT id, tenant_id, email, name, password_hash, role, roles, is_active, last_login_at, created_at, updated_at
FROM users WHERE email = $1
`
	return s.scanUser(s.db.QueryRowContext(ctx, query, email))
}

// GetUserByVerificationToken retrieves a user by their email verification token.
func (s *PostgresStore) GetUserByVerificationToken(ctx context.Context, token string) (*User, error) {
	query := `
SELECT id, tenant_id, email, name, password_hash, role, roles, is_active, last_login_at, created_at, updated_at
FROM users WHERE email_verification_token = $1
`
	return s.scanUser(s.db.QueryRowContext(ctx, query, token))
}

// ListUsersByTenant retrieves all users for a tenant.
func (s *PostgresStore) ListUsersByTenant(ctx context.Context, tenantID string) ([]*User, error) {
	query := `
SELECT id, tenant_id, email, name, password_hash, role, roles, is_active, last_login_at, created_at, updated_at
FROM users WHERE tenant_id = $1 ORDER BY email
`
	rows, err := s.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("auth: failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var u User
		var rolesStr sql.NullString
		if err := rows.Scan(
			&u.ID, &u.TenantID, &u.Email, &u.Name, &u.PasswordHash,
			&u.Role, &rolesStr, &u.IsActive, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("auth: failed to scan user: %w", err)
		}
		if rolesStr.Valid && rolesStr.String != "" {
			u.Roles = strings.Split(rolesStr.String, ",")
		} else {
			u.Roles = []string{}
		}
		users = append(users, &u)
	}
	return users, rows.Err()
}

// UpdateUser updates an existing user record.
func (s *PostgresStore) UpdateUser(ctx context.Context, user *User) error {
	query := `
UPDATE users SET email = $2, name = $3, password_hash = $4, role = $5, 
roles = $6, is_active = $7, last_login_at = $8, updated_at = $9
WHERE id = $1
`
	result, err := s.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.Name,
		user.PasswordHash,
		user.Role,
		user.Roles,
		user.IsActive,
		user.LastLoginAt,
		user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("auth: failed to update user: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrUserNotFound
	}
	return nil
}

// DeleteUser removes a user by ID.
func (s *PostgresStore) DeleteUser(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`
	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("auth: failed to delete user: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrUserNotFound
	}
	return nil
}

// UpdateUserLastLogin updates a user's last login timestamp.
func (s *PostgresStore) UpdateUserLastLogin(ctx context.Context, id string) error {
	query := `UPDATE users SET last_login_at = $2, updated_at = $3 WHERE id = $1`
	now := time.Now()
	result, err := s.db.ExecContext(ctx, query, id, now, now)
	if err != nil {
		return fmt.Errorf("auth: failed to update user last login: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrUserNotFound
	}
	return nil
}

// UpdateUserPassword updates a user's password hash.
func (s *PostgresStore) UpdateUserPassword(ctx context.Context, id, passwordHash string) error {
	query := `UPDATE users SET password_hash = $2, updated_at = $3 WHERE id = $1`
	result, err := s.db.ExecContext(ctx, query, id, passwordHash, time.Now())
	if err != nil {
		return fmt.Errorf("auth: failed to update user password: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (s *PostgresStore) scanUser(row *sql.Row) (*User, error) {
	var u User
	var rolesStr sql.NullString
	err := row.Scan(
		&u.ID, &u.TenantID, &u.Email, &u.Name, &u.PasswordHash,
		&u.Role, &rolesStr, &u.IsActive, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("auth: failed to get user: %w", err)
	}
	if rolesStr.Valid && rolesStr.String != "" {
		u.Roles = strings.Split(rolesStr.String, ",")
	} else {
		u.Roles = []string{}
	}
	return &u, nil
}

// CreateAPIKey inserts a new API key record.
func (s *PostgresStore) CreateAPIKey(ctx context.Context, key *APIKey) error {
	if err := key.Validate(); err != nil {
		return err
	}

	query := `
INSERT INTO api_keys (id, tenant_id, user_id, key_hash, key_prefix, label, scopes, is_active, expires_at, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
`
	_, err := s.db.ExecContext(ctx, query,
		key.ID,
		key.TenantID,
		key.UserID,
		key.KeyHash,
		key.KeyPrefix,
		key.Label,
		key.Scopes,
		key.IsActive,
		key.ExpiresAt,
		key.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("auth: failed to create API key: %w", err)
	}
	return nil
}

// GetAPIKey retrieves an API key by ID.
func (s *PostgresStore) GetAPIKey(ctx context.Context, id string) (*APIKey, error) {
	query := `
SELECT id, tenant_id, user_id, key_hash, key_prefix, label, scopes, is_active, expires_at, last_used_at, created_at
FROM api_keys WHERE id = $1
`
	return s.scanAPIKey(s.db.QueryRowContext(ctx, query, id))
}

// GetAPIKeyByHash retrieves an API key by its hash.
func (s *PostgresStore) GetAPIKeyByHash(ctx context.Context, hash string) (*APIKey, error) {
	query := `
SELECT id, tenant_id, user_id, key_hash, key_prefix, label, scopes, is_active, expires_at, last_used_at, created_at
FROM api_keys WHERE key_hash = $1
`
	return s.scanAPIKey(s.db.QueryRowContext(ctx, query, hash))
}

// ListAPIKeysByTenant retrieves all API keys for a tenant.
func (s *PostgresStore) ListAPIKeysByTenant(ctx context.Context, tenantID string) ([]*APIKey, error) {
	query := `
SELECT id, tenant_id, user_id, key_hash, key_prefix, label, scopes, is_active, expires_at, last_used_at, created_at
FROM api_keys WHERE tenant_id = $1 ORDER BY created_at DESC
`
	rows, err := s.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("auth: failed to list API keys: %w", err)
	}
	defer rows.Close()

	var keys []*APIKey
	for rows.Next() {
		var k APIKey
		if err := rows.Scan(
			&k.ID, &k.TenantID, &k.UserID, &k.KeyHash, &k.KeyPrefix,
			&k.Label, &k.Scopes, &k.IsActive, &k.ExpiresAt, &k.LastUsedAt, &k.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("auth: failed to scan API key: %w", err)
		}
		keys = append(keys, &k)
	}
	return keys, rows.Err()
}

// UpdateAPIKey updates an existing API key record.
func (s *PostgresStore) UpdateAPIKey(ctx context.Context, key *APIKey) error {
	query := `
UPDATE api_keys SET label = $2, scopes = $3, is_active = $4, expires_at = $5, last_used_at = $6
WHERE id = $1
`
	result, err := s.db.ExecContext(ctx, query,
		key.ID,
		key.Label,
		key.Scopes,
		key.IsActive,
		key.ExpiresAt,
		key.LastUsedAt,
	)
	if err != nil {
		return fmt.Errorf("auth: failed to update API key: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrInvalidAPIKey
	}
	return nil
}

// UpdateAPIKeyLastUsed updates the last_used_at timestamp for an API key.
func (s *PostgresStore) UpdateAPIKeyLastUsed(ctx context.Context, id string) error {
	query := `UPDATE api_keys SET last_used_at = $2 WHERE id = $1`
	result, err := s.db.ExecContext(ctx, query, id, time.Now())
	if err != nil {
		return fmt.Errorf("auth: failed to update API key last used: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrInvalidAPIKey
	}
	return nil
}

// RevokeAPIKey marks an API key as inactive.
func (s *PostgresStore) RevokeAPIKey(ctx context.Context, id string) error {
	query := `UPDATE api_keys SET is_active = false WHERE id = $1`
	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("auth: failed to revoke API key: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrInvalidAPIKey
	}
	return nil
}

// DeleteAPIKey removes an API key by ID.
func (s *PostgresStore) DeleteAPIKey(ctx context.Context, id string) error {
	query := `DELETE FROM api_keys WHERE id = $1`
	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("auth: failed to delete API key: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrInvalidAPIKey
	}
	return nil
}

// CountAPIKeysByTenant counts the number of API keys for a tenant.
func (s *PostgresStore) CountAPIKeysByTenant(ctx context.Context, tenantID string) (int, error) {
	query := `SELECT COUNT(*) FROM api_keys WHERE tenant_id = $1`
	var count int
	err := s.db.QueryRowContext(ctx, query, tenantID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("auth: failed to count API keys: %w", err)
	}
	return count, nil
}

func (s *PostgresStore) scanAPIKey(row *sql.Row) (*APIKey, error) {
	var k APIKey
	err := row.Scan(
		&k.ID, &k.TenantID, &k.UserID, &k.KeyHash, &k.KeyPrefix,
		&k.Label, &k.Scopes, &k.IsActive, &k.ExpiresAt, &k.LastUsedAt, &k.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrInvalidAPIKey
	}
	if err != nil {
		return nil, fmt.Errorf("auth: failed to get API key: %w", err)
	}
	return &k, nil
}

// -----------------------------------------------------------------------------
// In-Memory Store (for testing)
// -----------------------------------------------------------------------------

// InMemoryStore implements Store using in-memory maps.
// Useful for testing. Thread-safe via RWMutex.
type InMemoryStore struct {
	mu      sync.RWMutex
	tenants map[string]*Tenant
	users   map[string]*User
	apiKeys map[string]*APIKey
}

// NewInMemoryStore creates a new in-memory auth store.
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		tenants: make(map[string]*Tenant),
		users:   make(map[string]*User),
		apiKeys: make(map[string]*APIKey),
	}
}

func (s *InMemoryStore) CreateTenant(_ context.Context, tenant *Tenant) error {
	if err := tenant.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.tenants[tenant.ID]; exists {
		return errors.New("auth: tenant already exists")
	}
	s.tenants[tenant.ID] = copyTenant(tenant)
	return nil
}

func (s *InMemoryStore) GetTenant(_ context.Context, id string) (*Tenant, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tenants[id]
	if !ok {
		return nil, ErrTenantNotFound
	}
	return copyTenant(t), nil
}

func (s *InMemoryStore) GetTenantByName(_ context.Context, name string) (*Tenant, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, t := range s.tenants {
		if t.Name == name {
			return copyTenant(t), nil
		}
	}
	return nil, ErrTenantNotFound
}

func (s *InMemoryStore) UpdateTenant(_ context.Context, tenant *Tenant) error {
	if err := tenant.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.tenants[tenant.ID]; !exists {
		return ErrTenantNotFound
	}
	s.tenants[tenant.ID] = copyTenant(tenant)
	return nil
}

func (s *InMemoryStore) DeleteTenant(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.tenants[id]; !exists {
		return ErrTenantNotFound
	}
	delete(s.tenants, id)
	return nil
}

func (s *InMemoryStore) ListTenants(_ context.Context) ([]*Tenant, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*Tenant, 0, len(s.tenants))
	for _, t := range s.tenants {
		result = append(result, copyTenant(t))
	}
	return result, nil
}

func (s *InMemoryStore) CreateUser(_ context.Context, user *User) error {
	if err := user.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.users[user.ID]; exists {
		return errors.New("auth: user already exists")
	}
	s.users[user.ID] = copyUser(user)
	return nil
}

func (s *InMemoryStore) GetUser(_ context.Context, id string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	return copyUser(u), nil
}

func (s *InMemoryStore) GetUserByEmail(_ context.Context, email string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, u := range s.users {
		if u.Email == email {
			return copyUser(u), nil
		}
	}
	return nil, ErrUserNotFound
}

func (s *InMemoryStore) GetUserByVerificationToken(_ context.Context, token string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, u := range s.users {
		if u.EmailVerificationToken == token && token != "" {
			return copyUser(u), nil
		}
	}
	return nil, ErrUserNotFound
}

func (s *InMemoryStore) ListUsersByTenant(_ context.Context, tenantID string) ([]*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*User
	for _, u := range s.users {
		if u.TenantID == tenantID {
			result = append(result, copyUser(u))
		}
	}
	return result, nil
}

func (s *InMemoryStore) UpdateUser(_ context.Context, user *User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.users[user.ID]; !exists {
		return ErrUserNotFound
	}
	s.users[user.ID] = copyUser(user)
	return nil
}

func (s *InMemoryStore) DeleteUser(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.users[id]; !exists {
		return ErrUserNotFound
	}
	delete(s.users, id)
	return nil
}

func (s *InMemoryStore) UpdateUserLastLogin(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, exists := s.users[id]
	if !exists {
		return ErrUserNotFound
	}
	now := time.Now()
	u.LastLoginAt = &now
	u.UpdatedAt = now
	return nil
}

func (s *InMemoryStore) UpdateUserPassword(_ context.Context, id, passwordHash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, exists := s.users[id]
	if !exists {
		return ErrUserNotFound
	}
	u.PasswordHash = passwordHash
	u.UpdatedAt = time.Now()
	return nil
}

func (s *InMemoryStore) CreateAPIKey(_ context.Context, key *APIKey) error {
	if err := key.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.apiKeys[key.ID]; exists {
		return errors.New("auth: API key already exists")
	}
	s.apiKeys[key.ID] = copyAPIKey(key)
	return nil
}

func (s *InMemoryStore) GetAPIKey(_ context.Context, id string) (*APIKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	k, ok := s.apiKeys[id]
	if !ok {
		return nil, ErrInvalidAPIKey
	}
	return copyAPIKey(k), nil
}

func (s *InMemoryStore) GetAPIKeyByHash(_ context.Context, hash string) (*APIKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, k := range s.apiKeys {
		if k.KeyHash == hash {
			return copyAPIKey(k), nil
		}
	}
	return nil, ErrInvalidAPIKey
}

func (s *InMemoryStore) ListAPIKeysByTenant(_ context.Context, tenantID string) ([]*APIKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*APIKey
	for _, k := range s.apiKeys {
		if k.TenantID == tenantID {
			result = append(result, copyAPIKey(k))
		}
	}
	return result, nil
}

func (s *InMemoryStore) UpdateAPIKey(_ context.Context, key *APIKey) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.apiKeys[key.ID]; !exists {
		return ErrInvalidAPIKey
	}
	s.apiKeys[key.ID] = copyAPIKey(key)
	return nil
}

func (s *InMemoryStore) UpdateAPIKeyLastUsed(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	k, exists := s.apiKeys[id]
	if !exists {
		return ErrInvalidAPIKey
	}
	now := time.Now()
	k.LastUsedAt = &now
	return nil
}

func (s *InMemoryStore) RevokeAPIKey(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	k, exists := s.apiKeys[id]
	if !exists {
		return ErrInvalidAPIKey
	}
	k.IsActive = false
	return nil
}

func (s *InMemoryStore) DeleteAPIKey(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.apiKeys[id]; !exists {
		return ErrInvalidAPIKey
	}
	delete(s.apiKeys, id)
	return nil
}

func (s *InMemoryStore) CountAPIKeysByTenant(_ context.Context, tenantID string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	count := 0
	for _, k := range s.apiKeys {
		if k.TenantID == tenantID {
			count++
		}
	}
	return count, nil
}

// Deep copy helpers to prevent external mutation.
func copyTenant(t *Tenant) *Tenant {
	cp := *t
	return &cp
}

func copyUser(u *User) *User {
	cp := *u
	if u.Roles != nil {
		cp.Roles = make([]string, len(u.Roles))
		copy(cp.Roles, u.Roles)
	}
	return &cp
}

func copyAPIKey(k *APIKey) *APIKey {
	cp := *k
	if k.Scopes != nil {
		cp.Scopes = make([]string, len(k.Scopes))
		copy(cp.Scopes, k.Scopes)
	}
	return &cp
}

// Ensure compile-time interface compliance.
var (
	_ Store = (*PostgresStore)(nil)
	_ Store = (*InMemoryStore)(nil)
)
