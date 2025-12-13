package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/offgridflow/internal/api/http/handlers"
	"github.com/example/offgridflow/internal/auth"
	"github.com/example/offgridflow/internal/ingestion"
)

func TestActivitiesHandler_List_EmptyIsArray(t *testing.T) {
	store := ingestion.NewInMemoryActivityStore()

	h := handlers.NewActivitiesHandler(handlers.ActivitiesHandlerConfig{
		Store:        store,
		DefaultOrgID: "",
	})

	req := httptest.NewRequest(http.MethodGet, "/api/emissions/activities", nil)
	// Provide a tenant so orgID is deterministic.
	req = req.WithContext(auth.WithTenant(req.Context(), &auth.Tenant{ID: "org-empty"}))

	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var body map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	acts, ok := body["activities"]
	if !ok {
		t.Fatalf("expected 'activities' key in response")
	}

	// When empty, it must still decode as a JSON array.
	if _, ok := acts.([]interface{}); !ok {
		t.Fatalf("expected activities to be a JSON array, got %T (%v)", acts, acts)
	}
}
