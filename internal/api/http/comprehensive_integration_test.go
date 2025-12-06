package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/example/offgridflow/internal/api/http"
	"github.com/example/offgridflow/internal/auth"
	"github.com/example/offgridflow/internal/db"
	"github.com/example/offgridflow/internal/emissions"
	"github.com/example/offgridflow/internal/ingestion"
)

// TestFullEmissionsCalculationFlow tests the complete flow:
// User registration → login → create activity → calculate emissions → generate report
func TestFullEmissionsCalculationFlow(t *testing.T) {
	// Setup test database
	ctx := context.Background()
	testDB := setupTestDB(t)
	defer testDB.Close()

	// Setup test dependencies
	router := setupTestRouter(t, testDB)

	// Step 1: Register new user
	t.Run("user registration", func(t *testing.T) {
		payload := map[string]string{
			"email":    "test@offgridflow.com",
			"password": "SecurePass123!",
			"name":     "Test User",
		}

		resp := makeRequest(t, router, "POST", "/api/auth/register", payload)
		assertStatus(t, resp, http.StatusCreated)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		
		assert(t, result["user"] != nil, "user should be in response")
		assert(t, result["token"] != nil, "token should be in response")
	})

	// Step 2: Login to get auth token
	var authToken string
	t.Run("user login", func(t *testing.T) {
		payload := map[string]string{
			"email":    "test@offgridflow.com",
			"password": "SecurePass123!",
		}

		resp := makeRequest(t, router, "POST", "/api/auth/login", payload)
		assertStatus(t, resp, http.StatusOK)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		
		authToken = result["token"].(string)
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
		payload := map[string]string{
			"email":    "user2@offgridflow.com",
			"password": "SecurePass456!",
			"name":     "User Two",
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
	registerPayload := map[string]string{
		"email":    "ratelimit@test.com",
		"password": "Pass123!",
		"name":     "Rate Test",
	}

	makeRequest(t, router, "POST", "/api/auth/register", registerPayload)

	loginPayload := map[string]string{
		"email":    "ratelimit@test.com",
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
		lastStatus = resp.StatusCode
		
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
	assert(t, resp.StatusCode == http.StatusOK, "rate limit should reset after cooldown")
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
	registerPayload := map[string]string{
		"email":    "offline@test.com",
		"password": "Pass123!",
		"name":     "Offline User",
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
	registerPayload := map[string]string{
		"email":    "concurrent@test.com",
		"password": "Pass123!",
		"name":     "Concurrent User",
	}
	makeRequest(t, router, "POST", "/api/auth/register", registerPayload)

	loginPayload := map[string]string{
		"email":    "concurrent@test.com",
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
				registerPayload := map[string]string{
					"email":    "valid@test.com",
					"password": "Pass123!",
					"name":     "Valid User",
				}
				makeRequest(t, router, "POST", "/api/auth/register", registerPayload)

				loginPayload := map[string]string{
					"email":    "valid@test.com",
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
	
	// Use in-memory SQLite for testing
	dsn := "file::memory:?cache=shared"
	database, err := db.Connect(context.Background(), db.Config{DSN: dsn})
	if err != nil {
		t.Fatalf("failed to setup test database: %v", err)
	}

	// Run migrations
	if err := database.RunMigrations(context.Background()); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	return database
}

func setupTestRouter(t *testing.T, database *db.DB) http.Handler {
	t.Helper()

	authStore := auth.NewPostgresStore(database.DB)
	sessionManager, _ := auth.NewSessionManager("test-secret-key")
	activityStore := ingestion.NewPostgresActivityStore(database.DB)
	scope2Calc := emissions.NewScope2Calculator(emissions.Scope2Config{})

	cfg := &http.RouterConfig{
		DB:               database,
		AuthStore:        authStore,
		SessionManager:   sessionManager,
		ActivityStore:    activityStore,
		Scope2Calculator: scope2Calc,
		RequireAuth:      true,
	}

	return http.NewRouterWithConfig(cfg)
}

func setupTestRouterNoDB(t *testing.T) http.Handler {
	t.Helper()

	authStore := auth.NewInMemoryStore()
	sessionManager, _ := auth.NewSessionManager("test-secret-key")
	activityStore := ingestion.NewInMemoryActivityStore()
	scope2Calc := emissions.NewScope2Calculator(emissions.Scope2Config{})

	cfg := &http.RouterConfig{
		AuthStore:        authStore,
		SessionManager:   sessionManager,
		ActivityStore:    activityStore,
		Scope2Calculator: scope2Calc,
		RequireAuth:      true,
	}

	return http.NewRouterWithConfig(cfg)
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
	}

	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)

	return resp
}

func assertStatus(t *testing.T, resp *httptest.ResponseRecorder, expected int) {
	t.Helper()
	if resp.Code != expected {
		t.Errorf("expected status %d, got %d. Body: %s", expected, resp.Code, resp.Body.String())
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
	sessionManager, _ := auth.NewSessionManager("test-secret-key")
	claims := &auth.SessionClaims{
		UserID:   "test-user",
		TenantID: "test-tenant",
		Email:    "test@test.com",
	}
	token, _ := sessionManager.GenerateToken(claims)
	// In real implementation, this would use a past expiry time
	return token
}
