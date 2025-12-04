package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/offgridflow/internal/auth"
)

func TestPasswordResetHappyPath(t *testing.T) {
	store := auth.NewInMemoryStore()
	pw, _ := auth.HashPassword("OldPass123")
	user := &auth.User{
		ID:           "user-1",
		TenantID:     "tenant-1",
		Email:        "test@example.com",
		Name:         "Test User",
		PasswordHash: pw,
		IsActive:     true,
		Role:         "admin",
		Roles:        []string{"admin"},
	}
	if err := store.CreateUser(context.Background(), user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	svc := auth.NewService(store, nil)
	token, _, err := svc.CreatePasswordResetToken(context.Background(), user.Email)
	if err != nil {
		t.Fatalf("create reset token: %v", err)
	}

	handler := NewPasswordResetHandler(svc, nil)
	body, _ := json.Marshal(map[string]string{
		"token":        token,
		"new_password": "NewPass1234",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/password/reset", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}

	updated, err := store.GetUser(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	if updated.PasswordHash == pw {
		t.Fatalf("password hash did not change")
	}
}

func TestPasswordResetInvalidToken(t *testing.T) {
	store := auth.NewInMemoryStore()
	svc := auth.NewService(store, nil)
	handler := NewPasswordResetHandler(svc, nil)

	body, _ := json.Marshal(map[string]string{
		"token":        "badtoken",
		"new_password": "NewPass1234",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/password/reset", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}
