package handlers

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/offgridflow/internal/auth"
	"github.com/example/offgridflow/internal/ingestion"
)

func TestCSVIngestionHandler_SuccessMultipart(t *testing.T) {
	store := ingestion.NewInMemoryActivityStore()
	handler := NewCSVIngestionHandler(CSVIngestionHandlerConfig{
		Store:        store,
		DefaultOrgID: "org-default",
	})

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fileWriter, err := writer.CreateFormFile("file", "utility.csv")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	_, _ = fileWriter.Write([]byte("meter_id,location,period_start,period_end,kwh\nM1,US-WEST,2025-01-01,2025-01-31,1000"))
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/ingestion/upload/csv", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	// Provide tenant context to set org ID
	req = req.WithContext(auth.WithTenant(context.Background(), &auth.Tenant{ID: "org-tenant"}))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected status 202, got %d", rr.Code)
	}

	activities, err := store.List(context.Background())
	if err != nil {
		t.Fatalf("failed to list activities: %v", err)
	}
	if len(activities) != 1 {
		t.Fatalf("expected 1 activity saved, got %d", len(activities))
	}
	if activities[0].OrgID != "org-tenant" {
		t.Fatalf("expected org ID org-tenant, got %s", activities[0].OrgID)
	}
}

func TestCSVIngestionHandler_MissingOrg(t *testing.T) {
	store := ingestion.NewInMemoryActivityStore()
	handler := NewCSVIngestionHandler(CSVIngestionHandlerConfig{
		Store: store,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/ingestion/upload/csv", bytes.NewBufferString("meter_id,location,period_start,period_end,kwh\nM1,US-WEST,2025-01-01,2025-01-31,1000"))
	req.Header.Set("Content-Type", "text/csv")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing org, got %d", rr.Code)
	}
}

func TestCSVIngestionHandler_MethodNotAllowed(t *testing.T) {
	handler := NewCSVIngestionHandler(CSVIngestionHandlerConfig{Store: ingestion.NewInMemoryActivityStore()})
	req := httptest.NewRequest(http.MethodGet, "/api/ingestion/upload/csv", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}
