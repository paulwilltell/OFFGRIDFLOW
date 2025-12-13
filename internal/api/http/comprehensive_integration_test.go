package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"

	apihttp "github.com/example/offgridflow/internal/api/http"
	"github.com/example/offgridflow/internal/auth"
	"github.com/example/offgridflow/internal/db"
	"github.com/example/offgridflow/internal/emissions"
	"github.com/example/offgridflow/internal/emissions/factors"
	"github.com/example/offgridflow/internal/ingestion"
)

// TestFullEmissionsCalculationFlow tests the complete flow:
// User registration → login → create activity → calculate emissions → generate report
func TestFullEmissionsCalculationFlow(t *testing.T) {
	// Setup test database
	testDB := setupTestDB(t)
	defer testDB.Close()

	// Setup test dependencies
	router := setupTestRouter(t, testDB)

	var authToken string

	// Step 1: Register new user
	t.Run("user registration", func(t *testing.T) {
		email := uniqueEmail("test")
		name := fmt.Sprintf("Test User %d", time.Now().UnixNano())

		payload := map[string]string{
			"email":    email,
			"password": "SecurePass123!",
			"name":     name,
		}

		resp := makeRequest(t, router, "POST", "/api/auth/register", payload)
		assertStatus(t, resp, http.StatusCreated)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		assert(t, result["user"] != nil, "user should be in response")
		assert(t, result["token"] != nil, "token should be in response")

		// Use token returned at registration for subsequent steps
		if tok, ok := result["token"].(string); ok {
			authToken = tok
		}
	})

	// Step 2: Login to get auth token
	t.Run("user login", func(t *testing.T) {
		if authToken != "" {
			t.Skip("login covered by registration token")
		}

		payload := map[string]string{
			"email":    "test@offgridflow.com",
			"password": "SecurePass123!",
		}

		resp := makeRequest(t, router, "POST", "/api/auth/login", payload)
		assertStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		if tok, ok := result["token"].(string); ok {
			authToken = tok
		}
		assert(t, authToken != "", "token should not be empty")
	})

	// Step 3: Create activity data
	var activityID string
	t.Run("create activity", func(t *testing.T) {
		payload := map[string]interface{}{
			"name":  "Office Electricity",
			"type":  "electricity",
			"value": 1000.0,
			"unit":  "kWh",
			"date":  "2025-01-01",
		}

		resp := makeAuthRequest(t, router, "POST", "/api/emissions/activities", payload, authToken)
		assertStatus(t, resp, http.StatusCreated)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		activity := result["activity"].(map[string]interface{})
		activityID = activity["id"].(string)

		assert(t, activityID != "", "activity ID should not be empty")
		assert(t, activity["emissions"] != nil, "emissions should be calculated")
	})

	// Step 4: Calculate Scope 2 emissions
	t.Run("calculate scope 2 emissions", func(t *testing.T) {
		resp := makeAuthRequest(t, router, "GET", "/api/emissions/scope2/summary", nil, authToken)
		assertStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		summary := result["summary"].(map[string]interface{})
		assert(t, summary["total_emissions"].(float64) > 0, "total emissions should be positive")
		assert(t, summary["total_activities"].(float64) >= 1, "should have at least one activity")
	})

	// Step 5: Generate compliance report
	t.Run("generate compliance report", func(t *testing.T) {
		payload := map[string]interface{}{
			"year":   2025,
			"format": "pdf",
		}

		resp := makeAuthRequest(t, router, "POST", "/api/compliance/csrd", payload, authToken)
		assertStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		assert(t, result["report_id"] != nil, "report ID should be present")
		assert(t, result["status"] == "generated", "report should be generated")
	})

	// Step 6: Verify data isolation (create second user)
	t.Run("multi-tenant isolation", func(t *testing.T) {
		// Register second user
		email2 := uniqueEmail("user2")
		name2 := fmt.Sprintf("User Two %d", time.Now().UnixNano())

		payload := map[string]string{
			"email":    email2,
			"password": "SecurePass456!",
			"name":     name2,
		}

		resp := makeRequest(t, router, "POST", "/api/auth/register", payload)
		assertStatus(t, resp, http.StatusCreated)

		var regResult map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&regResult)
		user2Token := regResult["token"].(string)

		// Try to access first user's activities
		resp = makeAuthRequest(t, router, "GET", "/api/emissions/activities", nil, user2Token)
		assertStatus(t, resp, http.StatusOK)

		var actResult map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&actResult)

		activities := actResult["activities"].([]interface{})
		assert(t, len(activities) == 0, "user 2 should not see user 1's activities")
	})
}

// TestRateLimitingFlow tests rate limiting across multiple requests
func TestRateLimitingFlow(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	router := setupTestRouter(t, testDB)

	// Register and login
	emailRL := uniqueEmail("ratelimit")
	nameRL := fmt.Sprintf("Rate Test %d", time.Now().UnixNano())

	registerPayload := map[string]string{
		"email":    emailRL,
		"password": "Pass123!",
		"name":     nameRL,
	}

	makeRequest(t, router, "POST", "/api/auth/register", registerPayload)

	loginPayload := map[string]string{
		"email":    emailRL,
		"password": "Pass123!",
	}

	resp := makeRequest(t, router, "POST", "/api/auth/login", loginPayload)
	var loginResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&loginResult)
	token := loginResult["token"].(string)

	// Make requests until rate limited (free tier: 5 req/s)
	var lastStatus int
	for i := 0; i < 10; i++ {
		resp := makeAuthRequest(t, router, "GET", "/api/emissions/activities", nil, token)
		lastStatus = resp.Code

		if lastStatus == http.StatusTooManyRequests {
			t.Logf("Rate limited after %d requests", i+1)
			break
		}
	}

	assert(t, lastStatus == http.StatusTooManyRequests, "should eventually be rate limited")

	// Wait for rate limit to reset
	time.Sleep(2 * time.Second)

	// Should be able to make requests again
	resp = makeAuthRequest(t, router, "GET", "/api/emissions/activities", nil, token)
	assert(t, resp.Code == http.StatusOK, "rate limit should reset after cooldown")
}

// TestDatabaseFailover tests graceful degradation when database is unavailable
func TestDatabaseFailover(t *testing.T) {
	// Setup router with no database (should fall back to in-memory)
	router := setupTestRouterNoDB(t)

	// Health check should show degraded but functional
	resp := makeRequest(t, router, "GET", "/readyz", nil)
	assertStatus(t, resp, http.StatusServiceUnavailable)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	assert(t, result["status"] == "degraded", "should be in degraded mode")
	assert(t, result["checks"].(map[string]interface{})["database"] == "not_configured", "DB should be unavailable")

	// But basic endpoints should still work with in-memory store
	emailOff := uniqueEmail("offline")
	nameOff := fmt.Sprintf("Offline User %d", time.Now().UnixNano())

	registerPayload := map[string]string{
		"email":    emailOff,
		"password": "Pass123!",
		"name":     nameOff,
	}

	resp = makeRequest(t, router, "POST", "/api/auth/register", registerPayload)
	assertStatus(t, resp, http.StatusCreated)
}

// TestConcurrentRequests tests handling of concurrent requests
func TestConcurrentRequests(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	router := setupTestRouter(t, testDB)

	// Register user
	emailCon := uniqueEmail("concurrent")
	nameCon := fmt.Sprintf("Concurrent User %d", time.Now().UnixNano())
	registerPayload := map[string]string{
		"email":    emailCon,
		"password": "Pass123!",
		"name":     nameCon,
	}
	makeRequest(t, router, "POST", "/api/auth/register", registerPayload)

	loginPayload := map[string]string{
		"email":    emailCon,
		"password": "Pass123!",
	}
	resp := makeRequest(t, router, "POST", "/api/auth/login", loginPayload)
	var loginResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&loginResult)
	token := loginResult["token"].(string)

	// Make concurrent activity creation requests
	const numRequests = 50
	results := make(chan *httptest.ResponseRecorder, numRequests)

	for i := 0; i < numRequests; i++ {
		go func(index int) {
			payload := map[string]interface{}{
				"name":  "Concurrent Activity",
				"type":  "electricity",
				"value": float64(index),
				"unit":  "kWh",
				"date":  "2025-01-01",
			}

			resp := makeAuthRequest(t, router, "POST", "/api/emissions/activities", payload, token)
			results <- resp
		}(i)
	}

	// Collect results
	successCount := 0
	for i := 0; i < numRequests; i++ {
		resp := <-results
		if resp.Code == http.StatusCreated {
			successCount++
		}
	}

	assert(t, successCount == numRequests, "all concurrent requests should succeed")

	// Verify all activities were created
	resp = makeAuthRequest(t, router, "GET", "/api/emissions/activities", nil, token)
	var actResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&actResult)

	activities := actResult["activities"].([]interface{})
	assert(t, len(activities) == numRequests, "should have all activities")
}

// TestAuthenticationFlow tests various authentication scenarios
func TestAuthenticationFlow(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	router := setupTestRouter(t, testDB)

	tests := []struct {
		name       string
		setup      func() string // Returns token
		endpoint   string
		wantStatus int
	}{
		{
			name: "valid token",
			setup: func() string {
				emailV := uniqueEmail("valid")
				nameV := fmt.Sprintf("Valid User %d", time.Now().UnixNano())
				registerPayload := map[string]string{
					"email":    emailV,
					"password": "Pass123!",
					"name":     nameV,
				}
				makeRequest(t, router, "POST", "/api/auth/register", registerPayload)

				loginPayload := map[string]string{
					"email":    emailV,
					"password": "Pass123!",
				}
				resp := makeRequest(t, router, "POST", "/api/auth/login", loginPayload)
				var result map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&result)
				return result["token"].(string)
			},
			endpoint:   "/api/emissions/activities",
			wantStatus: http.StatusOK,
		},
		{
			name: "expired token",
			setup: func() string {
				return generateExpiredToken(t)
			},
			endpoint:   "/api/emissions/activities",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "invalid token",
			setup: func() string {
				return "invalid.jwt.token"
			},
			endpoint:   "/api/emissions/activities",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "no token",
			setup: func() string {
				return ""
			},
			endpoint:   "/api/emissions/activities",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tt.setup()
			resp := makeAuthRequest(t, router, "GET", tt.endpoint, nil, token)
			assertStatus(t, resp, tt.wantStatus)
		})
	}
}

// Helper functions

func setupTestDB(t *testing.T) *db.DB {
	t.Helper()

	// Use local Postgres started via docker-compose for integration tests
	// Connect to local Postgres (docker-compose) for integration tests
	dsn := "postgresql://offgridflow:changeme@localhost:5432/offgridflow?sslmode=disable"
	database, err := db.Connect(context.Background(), db.Config{DSN: dsn})
	if err != nil {
		t.Fatalf("failed to setup test database: %v", err)
	}

	// If the tables are not present, run migrations; otherwise assume docker init applied schema
	var tenantsExists bool
	if err := database.QueryRowContext(context.Background(), "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='tenants')").Scan(&tenantsExists); err != nil {
		t.Fatalf("failed to check existing tables: %v", err)
	}

	// Run migrations to ensure schema is up-to-date (idempotent)
	if err := database.RunMigrations(context.Background()); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	return database
}

func setupTestRouter(t *testing.T, database *db.DB) http.Handler {
	t.Helper()

	authStore := auth.NewPostgresStore(database.DB)
	sessionManager, _ := auth.NewSessionManager("test-secret-key-0123456789012345")
	activityStore := ingestion.NewPostgresActivityStore(database.DB)
	// Setup emissions factor registry (in-memory for tests)
	factorRegistry := factors.NewInMemoryRegistry(factors.DefaultRegistryConfig())

	scope2Calc := emissions.NewScope2Calculator(emissions.Scope2Config{Registry: factorRegistry})

	cfg := &apihttp.RouterConfig{
		DB:               database,
		AuthStore:        authStore,
		SessionManager:   sessionManager,
		ActivityStore:    activityStore,
		Scope2Calculator: scope2Calc,
		FactorRegistry:   factorRegistry,
		RequireAuth:      true,
	}

	return apihttp.NewRouterWithConfig(cfg)
}

func setupTestRouterNoDB(t *testing.T) http.Handler {
	t.Helper()

	authStore := auth.NewInMemoryStore()
	sessionManager, _ := auth.NewSessionManager("test-secret-key-0123456789012345")
	activityStore := ingestion.NewInMemoryActivityStore()

	// In-memory factor registry for no-DB router
	factorRegistry := factors.NewInMemoryRegistry(factors.DefaultRegistryConfig())
	scope2Calc := emissions.NewScope2Calculator(emissions.Scope2Config{Registry: factorRegistry})

	cfg := &apihttp.RouterConfig{
		AuthStore:        authStore,
		SessionManager:   sessionManager,
		ActivityStore:    activityStore,
		Scope2Calculator: scope2Calc,
		FactorRegistry:   factorRegistry,
		RequireAuth:      true,
	}

	return apihttp.NewRouterWithConfig(cfg)
}

func makeRequest(t *testing.T, handler http.Handler, method, path string, payload interface{}) *httptest.ResponseRecorder {
	t.Helper()

	var body *bytes.Buffer
	if payload != nil {
		data, _ := json.Marshal(payload)
		body = bytes.NewBuffer(data)
	} else {
		body = bytes.NewBuffer(nil)
	}

	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)

	return resp
}

func makeAuthRequest(t *testing.T, handler http.Handler, method, path string, payload interface{}, token string) *httptest.ResponseRecorder {
	t.Helper()

	var body *bytes.Buffer
	if payload != nil {
		data, _ := json.Marshal(payload)
		body = bytes.NewBuffer(data)
	} else {
		body = bytes.NewBuffer(nil)
	}

	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
		// Also attach session cookie (handlers set this cookie on login/registration)
		req.AddCookie(&http.Cookie{Name: "offgrid_session", Value: token, Path: "/"})
	}

	// For mutating requests, fetch CSRF token and attach cookie/header
	if method != http.MethodGet && method != http.MethodHead && method != http.MethodOptions {
		// Get CSRF token
		csrfResp := makeRequest(t, handler, "GET", "/api/auth/csrf-token", nil)
		if csrfResp.Code == http.StatusOK {
			var body map[string]interface{}
			json.NewDecoder(csrfResp.Body).Decode(&body)
			if tokenVal, ok := body["csrf_token"].(string); ok && tokenVal != "" {
				req.Header.Set("X-CSRF-Token", tokenVal)
				// extract set-cookie header
				if setCookie := csrfResp.Header().Get("Set-Cookie"); setCookie != "" {
					// parse cookie name/value (simple split)
					// expected format: "csrf_token=VALUE; Path=/;"
					parts := strings.SplitN(setCookie, "=", 2)
					if len(parts) == 2 {
						name := strings.TrimSpace(parts[0])
						rest := parts[1]
						valParts := strings.SplitN(rest, ";", 2)
						value := valParts[0]
						req.AddCookie(&http.Cookie{Name: name, Value: value, Path: "/"})
					}
				}
			}
		}
	}

	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)

	return resp
}

func assertStatus(t *testing.T, resp *httptest.ResponseRecorder, expected int) {
	t.Helper()
	if resp.Code != expected {
		t.Fatalf("expected status %d, got %d. Body: %s", expected, resp.Code, resp.Body.String())
	}
}

func assert(t *testing.T, condition bool, message string) {
	t.Helper()
	if !condition {
		t.Errorf("assertion failed: %s", message)
	}
}

func generateExpiredToken(t *testing.T) string {
	t.Helper()
	// Generate a JWT token that's already expired
	claims := &auth.SessionClaims{
		UserID:   "test-user",
		TenantID: "test-tenant",
		Email:    "test@test.com",
	}
	// Create an expired token by setting expiry in the past
	now := time.Now().Add(-2 * time.Hour)
	claims.IssuedAt = jwt.NewNumericDate(now)
	claims.ExpiresAt = jwt.NewNumericDate(now.Add(1 * time.Hour))

	sessionManager, _ := auth.NewSessionManager("test-secret-key-0123456789012345")
	token, _ := sessionManager.CreateTokenWithClaims(*claims)
	return token
}

// uniqueEmail generates a test-unique email address to avoid collisions across runs.
func uniqueEmail(prefix string) string {
	return fmt.Sprintf("%s+%d@offgridflow.test", prefix, time.Now().UnixNano())
}
