package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/example/offgridflow/internal/api/http/middleware"
	"github.com/example/offgridflow/internal/api/http/responders"
	"github.com/example/offgridflow/internal/compliance"
)

// NewComplianceExportHandler exposes download links for compliance summaries.
func NewComplianceExportHandler(deps *ComplianceHandlerDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			responders.MethodNotAllowed(w, http.MethodGet)
			return
		}

		if deps == nil || deps.ComplianceService == nil {
			responders.Error(w, http.StatusServiceUnavailable, "compliance_unavailable", "compliance service not configured")
			return
		}

		ctx := r.Context()

		orgID := r.URL.Query().Get("org_id")
		if orgID == "" {
			tenantID, ok := middleware.MustGetTenantID(w, r)
			if !ok {
				return
			}
			orgID = tenantID
		}

		year := time.Now().Year()
		if yearParam := r.URL.Query().Get("year"); yearParam != "" {
			if parsed, err := strconv.Atoi(yearParam); err == nil && parsed > 2000 && parsed <= time.Now().Year()+1 {
				year = parsed
			}
		}

		format := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format")))
		if format == "" {
			format = "pdf"
		}

		summary, err := deps.ComplianceService.GenerateSummary(ctx, orgID, year)
		if err != nil {
			responders.Error(w, http.StatusInternalServerError, "export_summary_failed", err.Error())
			return
		}

		var data []byte
		var contentType string

		switch format {
		case "pdf":
			data, err = compliance.ExportSummaryToPDF(summary)
			contentType = "application/pdf"
		case "xbrl":
			data, err = compliance.ExportSummaryToXBRL(summary)
			contentType = "application/xml"
		default:
			responders.Error(w, http.StatusBadRequest, "unsupported_format", "supported formats: pdf, xbrl")
			return
		}

		if err != nil {
			responders.Error(w, http.StatusInternalServerError, "export_failed", err.Error())
			return
		}

		filename := fmt.Sprintf("compliance-summary-%s.%s", time.Now().UTC().Format("20060102"), format)
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		w.Write(data)
	}
}
