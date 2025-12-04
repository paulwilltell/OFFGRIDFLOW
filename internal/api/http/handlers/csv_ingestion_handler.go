package handlers

import (
	"io"
	"net/http"
	"strings"

	"github.com/example/offgridflow/internal/api/http/responders"
	"github.com/example/offgridflow/internal/auth"
	"github.com/example/offgridflow/internal/ingestion"
	"github.com/example/offgridflow/internal/ingestion/sources/csv_upload"
)

// CSVIngestionHandlerConfig configures the CSV ingestion upload endpoint.
type CSVIngestionHandlerConfig struct {
	Store         ingestion.ActivityStore
	DefaultOrgID  string
	MaxUploadSize int64 // optional max upload size in bytes; 0 means unlimited
}

// NewCSVIngestionHandler handles POST /api/ingestion/upload/csv requests.
// It accepts either multipart/form-data (field "file") or raw text/csv bodies.
func NewCSVIngestionHandler(cfg CSVIngestionHandlerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			responders.MethodNotAllowed(w, http.MethodPost)
			return
		}

		if cfg.Store == nil {
			responders.ServiceUnavailable(w, "ingestion store not configured", 0)
			return
		}

		if cfg.MaxUploadSize > 0 {
			r.Body = http.MaxBytesReader(w, r.Body, cfg.MaxUploadSize)
		}

		orgID := resolveOrgID(r, cfg.DefaultOrgID)
		if orgID == "" {
			responders.BadRequest(w, "missing_org_id", "org_id or tenant context is required")
			return
		}

		reader, cleanup, err := extractCSVReader(r)
		if err != nil {
			responders.BadRequest(w, "invalid_csv_upload", err.Error())
			return
		}
		if cleanup != nil {
			defer cleanup()
		}

		adapter := csv_upload.NewAdapter(cfg.Store)
		activities, ingestErr := adapter.IngestUtilityCSV(r.Context(), reader, orgID)
		if ingestErr != nil {
			responders.BadRequest(w, "csv_ingest_failed", ingestErr.Error())
			return
		}

		responders.Accepted(w, map[string]any{
			"status":          "accepted",
			"activitiesSaved": len(activities),
			"orgId":           orgID,
		})
	}
}

// resolveOrgID determines the target org/tenant for the upload.
func resolveOrgID(r *http.Request, defaultOrg string) string {
	// Query params override defaults
	orgID := strings.TrimSpace(r.URL.Query().Get("org_id"))
	if orgID == "" {
		orgID = strings.TrimSpace(r.URL.Query().Get("tenant_id"))
	}

	if orgID == "" {
		if tenant, ok := auth.TenantFromContext(r.Context()); ok && tenant != nil && tenant.ID != "" {
			orgID = tenant.ID
		}
	}

	if orgID == "" {
		orgID = strings.TrimSpace(defaultOrg)
	}

	return orgID
}

// extractCSVReader retrieves the CSV data from either multipart upload or raw body.
func extractCSVReader(r *http.Request) (io.Reader, func(), error) {
	contentType := r.Header.Get("Content-Type")

	if strings.HasPrefix(contentType, "multipart/form-data") {
		if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB default buffer
			return nil, nil, err
		}
		file, _, err := r.FormFile("file")
		if err != nil {
			return nil, nil, err
		}
		cleanup := func() {
			_ = file.Close()
		}
		return file, cleanup, nil
	}

	// For non-multipart requests, use body directly.
	if r.Body == nil {
		return nil, nil, io.ErrUnexpectedEOF
	}

	return r.Body, func() {}, nil
}
