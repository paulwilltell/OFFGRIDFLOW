package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/example/offgridflow/internal/api/http/middleware"
	"github.com/example/offgridflow/internal/api/http/responders"
	"github.com/example/offgridflow/internal/compliance"
)

// ComplianceHandlerDeps holds dependencies for compliance handlers.
type ComplianceHandlerDeps struct {
	ComplianceService *compliance.Service
}

// NewCSRDComplianceHandler creates an HTTP handler for GET /api/compliance/csrd.
func NewCSRDComplianceHandler(deps *ComplianceHandlerDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			responders.Error(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
			return
		}

		if deps == nil || deps.ComplianceService == nil {
			responders.Error(w, http.StatusServiceUnavailable, "compliance_unavailable", "compliance service not configured")
			return
		}

		ctx := r.Context()

		// Parse query parameters
		orgID := r.URL.Query().Get("org_id")
		if orgID == "" {
			tenantID, ok := middleware.MustGetTenantID(w, r)
			if !ok {
				return
			}
			orgID = tenantID
		}

		year := time.Now().Year()
		if yearStr := r.URL.Query().Get("year"); yearStr != "" {
			if y, err := strconv.Atoi(yearStr); err == nil && y > 2000 && y <= time.Now().Year()+1 {
				year = y
			}
		}

		// Generate CSRD report
		report, err := deps.ComplianceService.GenerateCSRDReport(ctx, orgID, year)
		if err != nil {
			responders.Error(w, http.StatusInternalServerError, "csrd_report_failed", err.Error())
			return
		}

		responders.JSON(w, http.StatusOK, report)
	}
}

// NewSECComplianceHandler creates an HTTP handler for GET /api/compliance/sec.
func NewSECComplianceHandler(deps *ComplianceHandlerDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			responders.Error(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
			return
		}

		if deps == nil || deps.ComplianceService == nil {
			responders.Error(w, http.StatusServiceUnavailable, "compliance_unavailable", "compliance service not configured")
			return
		}

		ctx := r.Context()

		// Parse query parameters
		orgID := r.URL.Query().Get("org_id")
		if orgID == "" {
			tenantID, ok := middleware.MustGetTenantID(w, r)
			if !ok {
				return
			}
			orgID = tenantID
		}

		orgName := r.URL.Query().Get("org_name")
		if orgName == "" {
			orgName = "Demo Organization"
		}

		cik := r.URL.Query().Get("cik")
		if cik == "" {
			cik = "0000000000" // Demo CIK
		}

		year := time.Now().Year()
		if yearStr := r.URL.Query().Get("year"); yearStr != "" {
			if y, err := strconv.Atoi(yearStr); err == nil && y > 2000 && y <= time.Now().Year()+1 {
				year = y
			}
		}

		// Generate SEC report
		report, err := deps.ComplianceService.GenerateSECReport(ctx, orgID, orgName, cik, year)
		if err != nil {
			responders.Error(w, http.StatusInternalServerError, "sec_report_failed", err.Error())
			return
		}

		responders.JSON(w, http.StatusOK, report)
	}
}

// NewCaliforniaComplianceHandler creates an HTTP handler for GET /api/compliance/california.
func NewCaliforniaComplianceHandler(deps *ComplianceHandlerDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			responders.Error(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
			return
		}

		if deps == nil || deps.ComplianceService == nil {
			responders.Error(w, http.StatusServiceUnavailable, "compliance_unavailable", "compliance service not configured")
			return
		}

		ctx := r.Context()

		// Parse query parameters
		orgID := r.URL.Query().Get("org_id")
		if orgID == "" {
			tenantID, ok := middleware.MustGetTenantID(w, r)
			if !ok {
				return
			}
			orgID = tenantID
		}

		orgName := r.URL.Query().Get("org_name")
		if orgName == "" {
			orgName = "Demo Organization"
		}

		year := time.Now().Year()
		if yearStr := r.URL.Query().Get("year"); yearStr != "" {
			if y, err := strconv.Atoi(yearStr); err == nil && y > 2000 && y <= time.Now().Year()+1 {
				year = y
			}
		}

		// Generate California report
		report, err := deps.ComplianceService.GenerateCaliforniaReport(ctx, orgID, orgName, year)
		if err != nil {
			responders.Error(w, http.StatusInternalServerError, "california_report_failed", err.Error())
			return
		}

		responders.JSON(w, http.StatusOK, report)
	}
}

// NewCBAMComplianceHandler creates an HTTP handler for GET /api/compliance/cbam.
func NewCBAMComplianceHandler(deps *ComplianceHandlerDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			responders.Error(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
			return
		}

		if deps == nil || deps.ComplianceService == nil {
			responders.Error(w, http.StatusServiceUnavailable, "compliance_unavailable", "compliance service not configured")
			return
		}

		ctx := r.Context()

		// Parse query parameters
		orgID := r.URL.Query().Get("org_id")
		if orgID == "" {
			tenantID, ok := middleware.MustGetTenantID(w, r)
			if !ok {
				return
			}
			orgID = tenantID
		}

		year := time.Now().Year()
		if yearStr := r.URL.Query().Get("year"); yearStr != "" {
			if y, err := strconv.Atoi(yearStr); err == nil && y > 2000 && y <= time.Now().Year()+1 {
				year = y
			}
		}

		quarter := 1
		if quarterStr := r.URL.Query().Get("quarter"); quarterStr != "" {
			if q, err := strconv.Atoi(quarterStr); err == nil && q >= 1 && q <= 4 {
				quarter = q
			}
		}

		// Generate CBAM report
		report, err := deps.ComplianceService.GenerateCBAMReport(ctx, orgID, quarter, year)
		if err != nil {
			responders.Error(w, http.StatusInternalServerError, "cbam_report_failed", err.Error())
			return
		}

		responders.JSON(w, http.StatusOK, report)
	}
}

// NewIFRSComplianceHandler creates an HTTP handler for GET /api/compliance/ifrs.
func NewIFRSComplianceHandler(deps *ComplianceHandlerDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			responders.Error(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
			return
		}

		if deps == nil || deps.ComplianceService == nil {
			responders.Error(w, http.StatusServiceUnavailable, "compliance_unavailable", "compliance service not configured")
			return
		}

		ctx := r.Context()

		// Parse query parameters
		orgID := r.URL.Query().Get("org_id")
		if orgID == "" {
			tenantID, ok := middleware.MustGetTenantID(w, r)
			if !ok {
				return
			}
			orgID = tenantID
		}

		orgName := r.URL.Query().Get("org_name")
		if orgName == "" {
			orgName = "Demo Organization"
		}

		year := time.Now().Year()
		if yearStr := r.URL.Query().Get("year"); yearStr != "" {
			if y, err := strconv.Atoi(yearStr); err == nil && y > 2000 && y <= time.Now().Year()+1 {
				year = y
			}
		}

		// Generate IFRS report
		report, err := deps.ComplianceService.GenerateIFRSReport(ctx, orgID, orgName, year)
		if err != nil {
			responders.Error(w, http.StatusInternalServerError, "ifrs_report_failed", err.Error())
			return
		}

		responders.JSON(w, http.StatusOK, report)
	}
}

// NewComplianceSummaryHandler creates an HTTP handler for GET /api/compliance/summary.
func NewComplianceSummaryHandler(deps *ComplianceHandlerDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			responders.Error(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
			return
		}

		if deps == nil || deps.ComplianceService == nil {
			responders.Error(w, http.StatusServiceUnavailable, "compliance_unavailable", "compliance service not configured")
			return
		}

		ctx := r.Context()

		// Parse query parameters
		orgID := r.URL.Query().Get("org_id")
		if orgID == "" {
			tenantID, ok := middleware.MustGetTenantID(w, r)
			if !ok {
				return
			}
			orgID = tenantID
		}

		year := time.Now().Year()
		if yearStr := r.URL.Query().Get("year"); yearStr != "" {
			if y, err := strconv.Atoi(yearStr); err == nil && y > 2000 && y <= time.Now().Year()+1 {
				year = y
			}
		}

		// Generate compliance summary
		summary, err := deps.ComplianceService.GenerateSummary(ctx, orgID, year)
		if err != nil {
			responders.Error(w, http.StatusInternalServerError, "summary_failed", err.Error())
			return
		}

		responders.JSON(w, http.StatusOK, summary)
	}
}
