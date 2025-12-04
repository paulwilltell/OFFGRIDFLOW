package auth

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

// MockStore is a simple in-memory store for testing
type MockStore struct {
	tenants    map[string]*Tenant
	users      map[string]*User
	apiKeys    map[string]*APIKey
	keyHash    map[string]*APIKey
	keyCounter int
	mu         sync.Mutex
}

func NewMockStore() *MockStore {
	return &MockStore{
		tenants: make(map[string]*Tenant),
		users:   make(map[string]*User),
		apiKeys: make(map[string]*APIKey),
		keyHash: make(map[string]*APIKey),
	}
}

func (m *MockStore) GetTenant(ctx context.Context, id string) (*Tenant, error) {
	t, ok := m.tenants[id]
	if !ok {
		return nil, errors.New("tenant not found")
	}
	return t, nil
}

func (m *MockStore) CreateTenant(ctx context.Context, tenant *Tenant) error {
	if tenant.ID == "" {
		tenant.ID = "tenant_" + tenant.Name
	}
	tenant.CreatedAt = time.Now()
	tenant.UpdatedAt = time.Now()
	m.tenants[tenant.ID] = tenant
	return nil
}

func (m *MockStore) UpdateTenant(ctx context.Context, tenant *Tenant) error {
	if _, ok := m.tenants[tenant.ID]; !ok {
		return errors.New("tenant not found")
	}
	tenant.UpdatedAt = time.Now()
	m.tenants[tenant.ID] = tenant
	return nil
}

func (m *MockStore) GetUser(ctx context.Context, id string) (*User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	return u, nil
}

func (m *MockStore) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, errors.New("user not found")
}

func (m *MockStore) CreateUser(ctx context.Context, user *User) error {
	if user.ID == "" {
		user.ID = "user_" + user.Email
	}
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	m.users[user.ID] = user
	return nil
}

func (m *MockStore) UpdateUser(ctx context.Context, user *User) error {
	if _, ok := m.users[user.ID]; !ok {
		return errors.New("user not found")
	}
	user.UpdatedAt = time.Now()
	m.users[user.ID] = user
	return nil
}

func (m *MockStore) ListUsersByTenant(ctx context.Context, tenantID string) ([]*User, error) {
	var users []*User
	for _, u := range m.users {
		if u.TenantID == tenantID {
			users = append(users, u)
		}
	}
	return users, nil
}

func (m *MockStore) GetAPIKeyByHash(ctx context.Context, hash string) (*APIKey, error) {
	key, ok := m.keyHash[hash]
	if !ok {
		return nil, ErrInvalidAPIKey
	}
	return key, nil
}

func (m *MockStore) CreateAPIKey(ctx context.Context, key *APIKey) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if key.ID == "" {
		m.keyCounter++
		key.ID = fmt.Sprintf("key_%d", m.keyCounter)
	}
	key.CreatedAt = time.Now()
	m.apiKeys[key.ID] = key
	m.keyHash[key.KeyHash] = key
	return nil
}

func (m *MockStore) RevokeAPIKey(ctx context.Context, keyID string) error {
	key, ok := m.apiKeys[keyID]
	if !ok {
		return errors.New("API key not found")
	}
	key.IsActive = false
	return nil
}

func (m *MockStore) UpdateAPIKeyLastUsed(ctx context.Context, keyID string) error {
	key, ok := m.apiKeys[keyID]
	if !ok {
		return errors.New("API key not found")
	}
	now := time.Now()
	key.LastUsedAt = &now
	return nil
}

func (m *MockStore) ListAPIKeysByTenant(ctx context.Context, tenantID string) ([]*APIKey, error) {
	var keys []*APIKey
	for _, k := range m.apiKeys {
		if k.TenantID == tenantID {
			keys = append(keys, k)
		}
	}
	return keys, nil
}

func (m *MockStore) CountAPIKeysByTenant(ctx context.Context, tenantID string) (int, error) {
	count := 0
	for _, k := range m.apiKeys {
		if k.TenantID == tenantID {
			count++
		}
	}
	return count, nil
}

func (m *MockStore) GetAPIKey(ctx context.Context, id string) (*APIKey, error) {
	key, ok := m.apiKeys[id]
	if !ok {
		return nil, errors.New("API key not found")
	}
	return key, nil
}

func (m *MockStore) UpdateAPIKey(ctx context.Context, key *APIKey) error {
	if _, ok := m.apiKeys[key.ID]; !ok {
		return errors.New("API key not found")
	}
	m.apiKeys[key.ID] = key
	return nil
}

func (m *MockStore) DeleteAPIKey(ctx context.Context, id string) error {
	delete(m.apiKeys, id)
	delete(m.keyHash, "")
	return nil
}

func (m *MockStore) GetTenantByName(ctx context.Context, name string) (*Tenant, error) {
	for _, t := range m.tenants {
		if t.Name == name {
			return t, nil
		}
	}
	return nil, errors.New("tenant not found")
}

func (m *MockStore) DeleteTenant(ctx context.Context, id string) error {
	delete(m.tenants, id)
	return nil
}

func (m *MockStore) ListTenants(ctx context.Context) ([]*Tenant, error) {
	var tenants []*Tenant
	for _, t := range m.tenants {
		tenants = append(tenants, t)
	}
	return tenants, nil
}

func (m *MockStore) UpdateUserLastLogin(ctx context.Context, id string) error {
	u, ok := m.users[id]
	if !ok {
		return errors.New("user not found")
	}
	now := time.Now()
	u.LastLoginAt = &now
	return nil
}

func (m *MockStore) UpdateUserPassword(ctx context.Context, id, passwordHash string) error {
	u, ok := m.users[id]
	if !ok {
		return errors.New("user not found")
	}
	u.PasswordHash = passwordHash
	return nil
}

func (m *MockStore) DeleteUser(ctx context.Context, id string) error {
	delete(m.users, id)
	return nil
}

func TestService_CreateTenant(t *testing.T) {
	store := NewMockStore()
	svc := NewService(store, nil)
	ctx := context.Background()

	tenant, user, rawKey, err := svc.CreateTenant(ctx, "Test Corp", "admin@test.com", "Admin User")
	if err != nil {
		t.Fatalf("CreateTenant() error = %v", err)
	}

	if tenant == nil {
		t.Fatal("Expected tenant to be created")
	}
	if user == nil {
		t.Fatal("Expected user to be created")
	}
	if rawKey == "" {
		t.Fatal("Expected API key to be generated")
	}

	// Verify tenant
	if tenant.Name != "Test Corp" {
		t.Errorf("Expected tenant name 'Test Corp', got %s", tenant.Name)
	}
	if !tenant.IsActive {
		t.Error("Expected tenant to be active")
	}

	// Verify user
	if user.Email != "admin@test.com" {
		t.Errorf("Expected user email 'admin@test.com', got %s", user.Email)
	}
	if user.Role != "admin" {
		t.Errorf("Expected user role 'admin', got %s", user.Role)
	}
}

func TestService_ValidateAPIKey(t *testing.T) {
	store := NewMockStore()
	svc := NewService(store, nil)
	ctx := context.Background()

	// Create tenant and API key
	_, _, rawKey, err := svc.CreateTenant(ctx, "Test Corp", "admin@test.com", "Admin")
	if err != nil {
		t.Fatalf("CreateTenant() error = %v", err)
	}

	// Validate the key
	tenant, user, key, err := svc.ValidateAPIKey(ctx, rawKey)
	if err != nil {
		t.Fatalf("ValidateAPIKey() error = %v", err)
	}

	if tenant == nil {
		t.Error("Expected tenant to be returned")
	}
	if user == nil {
		t.Error("Expected user to be returned")
	}
	if key == nil {
		t.Error("Expected key to be returned")
	}

	// Test invalid key
	_, _, _, err = svc.ValidateAPIKey(ctx, "invalid_key")
	if err == nil {
		t.Error("Expected error for invalid API key")
	}
	if !errors.Is(err, ErrInvalidAPIKey) {
		t.Errorf("Expected ErrInvalidAPIKey, got %v", err)
	}
}

func TestService_ValidateAPIKey_Expired(t *testing.T) {
	store := NewMockStore()
	svc := NewService(store, nil)
	ctx := context.Background()

	// Create tenant
	tenant, user, _, err := svc.CreateTenant(ctx, "Test Corp", "admin@test.com", "Admin")
	if err != nil {
		t.Fatalf("CreateTenant() error = %v", err)
	}

	// Create expired key
	pastTime := time.Now().Add(-1 * time.Hour)
	rawKey, apiKey, err := GenerateAPIKey("live", tenant.ID, user.ID, "Expired Key", []string{"*"}, &pastTime)
	if err != nil {
		t.Fatalf("GenerateAPIKey() error = %v", err)
	}

	if err := store.CreateAPIKey(ctx, apiKey); err != nil {
		t.Fatalf("CreateAPIKey() error = %v", err)
	}

	// Validate expired key
	_, _, _, err = svc.ValidateAPIKey(ctx, rawKey)
	if err == nil {
		t.Error("Expected error for expired API key")
	}
	if !errors.Is(err, ErrInvalidAPIKey) {
		t.Errorf("Expected ErrInvalidAPIKey, got %v", err)
	}
}

func TestService_RevokeAPIKey(t *testing.T) {
	store := NewMockStore()
	svc := NewService(store, nil)
	ctx := context.Background()

	// Create tenant and key
	_, _, rawKey, err := svc.CreateTenant(ctx, "Test Corp", "admin@test.com", "Admin")
	if err != nil {
		t.Fatalf("CreateTenant() error = %v", err)
	}

	// Validate before revocation
	_, _, key, err := svc.ValidateAPIKey(ctx, rawKey)
	if err != nil {
		t.Fatalf("ValidateAPIKey() before revocation error = %v", err)
	}

	// Revoke the key
	if err := svc.RevokeAPIKey(ctx, key.ID); err != nil {
		t.Fatalf("RevokeAPIKey() error = %v", err)
	}

	// Validate after revocation
	_, _, _, err = svc.ValidateAPIKey(ctx, rawKey)
	if err == nil {
		t.Error("Expected error for revoked API key")
	}
	if !errors.Is(err, ErrInvalidAPIKey) {
		t.Errorf("Expected ErrInvalidAPIKey, got %v", err)
	}
}

func TestService_CreateAPIKey_MaxLimit(t *testing.T) {
	store := NewMockStore()
	config := DefaultServiceConfig()
	config.MaxKeysPerTenant = 2
	svc := NewServiceWithConfig(store, nil, config)
	ctx := context.Background()

	// Create tenant
	tenant, user, _, err := svc.CreateTenant(ctx, "Test Corp", "admin@test.com", "Admin")
	if err != nil {
		t.Fatalf("CreateTenant() error = %v", err)
	}

	// First key already created, create one more
	_, _, err = svc.CreateAPIKey(ctx, tenant.ID, user.ID, "Key 2", nil, nil)
	if err != nil {
		t.Fatalf("CreateAPIKey() second key error = %v", err)
	}

	// Third key should fail
	_, _, err = svc.CreateAPIKey(ctx, tenant.ID, user.ID, "Key 3", nil, nil)
	if err == nil {
		t.Error("Expected error when exceeding max keys")
	}
}

func TestService_Authorize(t *testing.T) {
	store := NewMockStore()
	svc := NewService(store, nil)
	ctx := context.Background()

	adminUser := &User{ID: "1", Email: "admin@test.com", Role: "admin"}
	viewerUser := &User{ID: "2", Email: "viewer@test.com", Role: "viewer"}

	// Admin should be authorized
	allowed, err := svc.Authorize(ctx, adminUser, ActionDelete, ResourceUsers)
	if err != nil {
		t.Fatalf("Authorize() error = %v", err)
	}
	if !allowed {
		t.Error("Admin should be authorized to delete users")
	}

	// Viewer should not be authorized
	allowed, err = svc.Authorize(ctx, viewerUser, ActionDelete, ResourceUsers)
	if err != nil {
		t.Fatalf("Authorize() error = %v", err)
	}
	if allowed {
		t.Error("Viewer should not be authorized to delete users")
	}
}

func TestService_RequirePermission(t *testing.T) {
	store := NewMockStore()
	svc := NewService(store, nil)
	ctx := context.Background()

	adminUser := &User{ID: "1", Email: "admin@test.com", Role: "admin"}
	viewerUser := &User{ID: "2", Email: "viewer@test.com", Role: "viewer"}

	// Admin should pass
	err := svc.RequirePermission(ctx, adminUser, ActionDelete, ResourceUsers)
	if err != nil {
		t.Errorf("Admin RequirePermission() error = %v", err)
	}

	// Viewer should fail with ErrForbidden
	err = svc.RequirePermission(ctx, viewerUser, ActionDelete, ResourceUsers)
	if err == nil {
		t.Error("Expected error for viewer")
	}
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("Expected ErrForbidden, got %v", err)
	}
}
