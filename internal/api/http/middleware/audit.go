package middleware

import (
	"net/http"

	"github.com/example/offgridflow/internal/audit"
	"github.com/example/offgridflow/internal/auth"
)

// AuditMiddleware logs all authentication and authorization events
func AuditMiddleware(auditLogger *audit.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Wrap response writer to capture status code
			wrapped := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}

			// Process request
			next.ServeHTTP(wrapped, r)

			// Get auth context
			tenant, _ := auth.TenantFromContext(ctx)
			user, _ := auth.UserFromContext(ctx)
			apiKey, _ := auth.APIKeyFromContext(ctx)

			// Determine actor
			actorID := ""
			actorType := ""
			tenantID := ""

			if tenant != nil {
				tenantID = tenant.ID
			}

			if user != nil {
				actorID = user.ID
				actorType = "user"
			} else if apiKey != nil {
				actorID = apiKey.ID
				actorType = "apikey"
			} else {
				actorID = getClientIP(r)
				actorType = "anonymous"
			}

			// Only log certain paths and methods
			if shouldAuditRequest(r) {
				status := "success"
				if wrapped.statusCode >= 400 {
					status = "failure"
				}

				event := audit.Event{
					EventType: determineEventType(r, wrapped.statusCode),
					ActorID:   actorID,
					ActorType: actorType,
					TenantID:  tenantID,
					Resource:  r.URL.Path,
					Action:    r.Method,
					Status:    status,
					IPAddress: getClientIP(r),
					UserAgent: r.UserAgent(),
					Details: map[string]interface{}{
						"method":      r.Method,
						"path":        r.URL.Path,
						"status_code": wrapped.statusCode,
					},
				}

				// Log asynchronously to not block response
				go auditLogger.Log(ctx, event)
			}
		})
	}
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func shouldAuditRequest(r *http.Request) bool {
	// Audit these paths
	auditPaths := []string{
		"/api/auth/",
		"/api/organization/",
		"/api/users/",
		"/api/apikeys/",
		"/api/settings/",
	}

	for _, path := range auditPaths {
		if len(r.URL.Path) >= len(path) && r.URL.Path[:len(path)] == path {
			return true
		}
	}

	// Audit all non-GET requests
	if r.Method != "GET" {
		return true
	}

	return false
}

func determineEventType(r *http.Request, statusCode int) audit.EventType {
	// Map paths and methods to event types
	path := r.URL.Path

	if path == "/api/auth/login" {
		if statusCode == http.StatusOK {
			return audit.EventLogin
		}
		return audit.EventLoginFailed
	}

	if path == "/api/auth/logout" {
		return audit.EventLogout
	}

	if path == "/api/apikeys" && r.Method == "POST" {
		return audit.EventAPIKeyCreated
	}

	if r.Method == "DELETE" && len(path) > len("/api/apikeys/") && path[:len("/api/apikeys/")] == "/api/apikeys/" {
		return audit.EventAPIKeyRevoked
	}

	if r.Method == "POST" && path == "/api/users" {
		return audit.EventUserCreated
	}

	if r.Method == "PATCH" && len(path) > len("/api/users/") && path[:len("/api/users/")] == "/api/users/" {
		return audit.EventUserUpdated
	}

	if r.Method == "DELETE" {
		return audit.EventDataDeleted
	}

	if path == "/api/export" {
		return audit.EventDataExported
	}

	// Default to permission check
	if statusCode == http.StatusForbidden {
		return audit.EventPermissionDenied
	}

	return audit.EventType("http.request")
}
