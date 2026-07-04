package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/offgridflow/internal/api/http/handlers"
	"github.com/example/offgridflow/internal/auth"
)

// newRegisterHandler builds an AuthHandlers backed by in-memory stores for
// registration tests.
func newRegisterHandler(t *testing.T, cfg handlers.AuthHandlersConfig) *handlers.AuthHandlers {
	t.Helper()
	if cfg.AuthStore == nil {
		cfg.AuthStore = auth.NewInMemoryStore()
	}
	if cfg.SessionManager == nil {
		sm, err := auth.NewSessionManager("test-secret-key-for-registration")
		if err != nil {
			t.Fatalf("session manager: %v", err)
		}
		cfg.SessionManager = sm
	}
	return handlers.NewAuthHandlers(cfg)
}

func postRegister(t *testing.T, h *handlers.AuthHandlers, body map[string]any) *httptest.ResponseRecorder {
	t.Helper()
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Register(rr, req)
	return rr
}

// When email verification is required but no email sender is configured (the
// production reality before SMTP is wired up), registration must NOT dead-end
// with a 503. It should auto-verify the user and let signup complete.
func TestRegisterAutoVerifiesWhenNoEmailSender(t *testing.T) {
	h := newRegisterHandler(t, handlers.AuthHandlersConfig{
		RequireEmailVerification:  true, // production default
		EmailSender:               nil,  // SMTP not configured
		AllowDevVerificationToken: false,
	})

	rr := postRegister(t, h, map[string]any{
		"name":         "Jane Doe",
		"email":        "jane@example.com",
		"password":     "StrongPass123!",
		"company_name": "Acme Inc.",
	})

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201 (auto-verified signup), got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if rv, _ := resp["requires_verification"].(bool); rv {
		t.Errorf("expected requires_verification false when no email sender, got true")
	}
	user, _ := resp["user"].(map[string]any)
	if user == nil {
		t.Fatalf("no user in response: %s", rr.Body.String())
	}
	if verified, _ := user["email_verified"].(bool); !verified {
		t.Errorf("expected user auto-verified (email_verified=true), got false")
	}
}

// Registration must reject a missing name with a clear field error (the frontend
// now always sends one, but the API contract must hold).
func TestRegisterRequiresName(t *testing.T) {
	h := newRegisterHandler(t, handlers.AuthHandlersConfig{})
	rr := postRegister(t, h, map[string]any{
		"email":    "noname@example.com",
		"password": "StrongPass123!",
	})
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing name, got %d: %s", rr.Code, rr.Body.String())
	}
}

// A valid registration with a name succeeds end to end.
func TestRegisterHappyPath(t *testing.T) {
	h := newRegisterHandler(t, handlers.AuthHandlersConfig{RequireEmailVerification: false})
	rr := postRegister(t, h, map[string]any{
		"name":         "Paul T",
		"email":        "paul@example.com",
		"password":     "StrongPass123!",
		"company_name": "OffGridFlow",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
}
