package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/example/offgridflow/internal/api/http/middleware"
	"github.com/example/offgridflow/internal/api/http/responders"
	"github.com/example/offgridflow/internal/auth"
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

		tenantID, ok := middleware.MustGetTenantID(w, r)
		if !ok {
			return
		}
		orgID := tenantID

		// Default to the most recent year with data; customers upload prior-year
		// reporting data, so defaulting to the current calendar year would export
		// an empty paid report. An explicit ?year= always overrides.
		year := 0
		if yearParam := r.URL.Query().Get("year"); yearParam != "" {
			if parsed, err := strconv.Atoi(yearParam); err == nil && parsed > 2000 && parsed <= time.Now().Year()+1 {
				year = parsed
			}
		}
		if year == 0 {
			if latest := deps.ComplianceService.LatestDataYear(ctx, orgID); latest > 0 {
				year = latest
			} else {
				year = time.Now().Year()
			}
		}

		format := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format")))
		if format == "" {
			format = "pdf"
		}

		// Build the full traceable inventory (the audit-ready report backbone).
		inv, err := deps.ComplianceService.GenerateInventory(ctx, orgID, year)
		if err != nil {
			responders.Error(w, http.StatusInternalServerError, "export_inventory_failed", "failed to generate inventory")
			return
		}
		// Use the organization's real name on the report when available.
		if tenant, ok := auth.TenantFromContext(ctx); ok && tenant != nil && tenant.Name != "" {
			inv.OrgName = tenant.Name
		}

		var data []byte
		var contentType string

		switch format {
		case "pdf":
			data, err = compliance.ExportInventoryReportPDF(inv)
			contentType = "application/pdf"
		case "csv":
			data, err = compliance.ExportInventoryCSV(inv)
			contentType = "text/csv"
		case "xbrl":
			summary, serr := deps.ComplianceService.GenerateSummary(ctx, orgID, year)
			if serr != nil {
				responders.Error(w, http.StatusInternalServerError, "export_summary_failed", "failed to generate summary")
				return
			}
			data, err = compliance.ExportSummaryToXBRL(summary)
			contentType = "application/xml"
		default:
			responders.Error(w, http.StatusBadRequest, "unsupported_format", "supported formats: pdf, csv, xbrl")
			return
		}

		if err != nil {
			responders.Error(w, http.StatusInternalServerError, "export_failed", "export generation failed")
			return
		}

		filename := fmt.Sprintf("ghg-emissions-inventory-%d-%s.%s", year, time.Now().UTC().Format("20060102"), format)
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		w.Write(data)
	}
}
