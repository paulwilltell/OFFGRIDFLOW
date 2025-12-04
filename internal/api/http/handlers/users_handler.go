package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/example/offgridflow/internal/api/http/middleware"
	"github.com/example/offgridflow/internal/api/http/responders"
	"github.com/example/offgridflow/internal/auth"
)

// UsersHandlerConfig configures user CRUD endpoints.
type UsersHandlerConfig struct {
	Store auth.Store
}

// NewUsersHandler provides minimal user CRUD (list/create) for Phase 2 start.
func NewUsersHandler(cfg UsersHandlerConfig) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		if cfg.Store == nil {
			responders.Error(w, http.StatusServiceUnavailable, "users_unavailable", "user store not configured")
			return
		}

		tenantID, ok := middleware.MustGetTenantID(w, r)
		if !ok {
			return
		}

		switch r.Method {
		case http.MethodGet:
			users, err := cfg.Store.ListUsersByTenant(r.Context(), tenantID)
			if err != nil {
				responders.InternalError(w, "failed to list users")
				return
			}
			responders.JSON(w, http.StatusOK, users)
		case http.MethodPost:
			user, ok := auth.UserFromContext(r.Context())
			if !ok || user == nil || !user.HasRole("admin") {
				responders.Forbidden(w, "insufficient_role", "admin role required")
				return
			}
			var req auth.User
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				responders.BadRequest(w, "invalid_request", "invalid JSON payload")
				return
			}
			if req.Email == "" || req.PasswordHash == "" {
				responders.BadRequest(w, "missing_fields", "email and password_hash required")
				return
			}
			req.TenantID = tenantID
			if err := cfg.Store.CreateUser(r.Context(), &req); err != nil {
				responders.InternalError(w, "failed to create user")
				return
			}
			responders.JSON(w, http.StatusCreated, req)
		case http.MethodPut:
			user, ok := auth.UserFromContext(r.Context())
			if !ok || user == nil || !user.HasRole("admin") {
				responders.Forbidden(w, "insufficient_role", "admin role required")
				return
			}
			var req auth.User
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				responders.BadRequest(w, "invalid_request", "invalid JSON payload")
				return
			}
			if req.ID == "" {
				responders.BadRequest(w, "missing_id", "user id required")
				return
			}
			req.TenantID = tenantID
			if err := cfg.Store.UpdateUser(r.Context(), &req); err != nil {
				responders.InternalError(w, "failed to update user")
				return
			}
			responders.JSON(w, http.StatusOK, req)
		case http.MethodDelete:
			user, ok := auth.UserFromContext(r.Context())
			if !ok || user == nil || !user.HasRole("admin") {
				responders.Forbidden(w, "insufficient_role", "admin role required")
				return
			}
			id := r.URL.Query().Get("id")
			if id == "" {
				responders.BadRequest(w, "missing_id", "user id required")
				return
			}
			if err := cfg.Store.DeleteUser(r.Context(), id); err != nil {
				responders.InternalError(w, "failed to delete user")
				return
			}
			responders.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
		default:
			responders.MethodNotAllowed(w, http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete)
		}
	})

	mux.HandleFunc("/api/users/invite", func(w http.ResponseWriter, r *http.Request) {
		user, ok := auth.UserFromContext(r.Context())
		if !ok || user == nil || !user.HasRole("admin") {
			responders.Forbidden(w, "insufficient_role", "admin role required")
			return
		}
		if r.Method != http.MethodPost {
			responders.MethodNotAllowed(w, http.MethodPost)
			return
		}
		responders.JSON(w, http.StatusNotImplemented, map[string]string{
			"status":  "not_implemented",
			"message": "invite flow pending mail provider",
		})
	})

	return mux
}
