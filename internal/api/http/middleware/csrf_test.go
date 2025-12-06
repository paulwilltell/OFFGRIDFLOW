package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCSRFMiddleware_Wrap(t *testing.T) {
	mw := NewCSRFMiddleware(CSRFMiddlewareConfig{
		TokenTTL:        time.Minute,
		CleanupInterval: time.Hour,
		ExemptPaths: []string{
			"/api/billing/webhook",
		},
	})

	handler := mw.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("rejects missing token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/secure", nil)
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
		if res.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", res.Code)
		}
	})

	t.Run("rejects invalid token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/secure", nil)
		req.AddCookie(&http.Cookie{Name: mw.CookieName(), Value: "bad"})
		req.Header.Set(mw.HeaderName(), "also-bad")
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
		if res.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", res.Code)
		}
	})

	t.Run("allows valid token", func(t *testing.T) {
		token, err := mw.GenerateToken()
		if err != nil {
			t.Fatalf("failed to generate token: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/api/secure", nil)
		req.AddCookie(&http.Cookie{Name: mw.CookieName(), Value: token})
		req.Header.Set(mw.HeaderName(), token)
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
		if res.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", res.Code)
		}
	})

	t.Run("allows safe methods", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/secure", nil)
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
		if res.Code != http.StatusOK {
			t.Fatalf("expected 200 for GET, got %d", res.Code)
		}
	})

	t.Run("skips exempt path", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/billing/webhook", nil)
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
		if res.Code != http.StatusOK {
			t.Fatalf("expected 200 for exempt path, got %d", res.Code)
		}
	})
}
