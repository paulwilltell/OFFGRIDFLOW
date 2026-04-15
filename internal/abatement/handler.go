package abatement

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"

	"github.com/example/offgridflow/internal/api/http/responders"
	"github.com/example/offgridflow/internal/auth"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.service == nil {
		responders.Error(w, http.StatusServiceUnavailable, "abatement_unavailable", "abatement service not configured")
		return
	}

	tenant, ok := auth.TenantFromContext(r.Context())
	if !ok || tenant == nil {
		responders.Unauthorized(w, "unauthorized", "authentication required")
		return
	}
	user, userOK := auth.UserFromContext(r.Context())
	if !userOK || user == nil {
		responders.Unauthorized(w, "unauthorized", "user context required")
		return
	}

	framework, suffix, err := parseFrameworkPath(r.URL.Path)
	if err != nil {
		responders.BadRequest(w, "invalid_framework", err.Error())
		return
	}

	switch {
	case suffix == "" && r.Method == http.MethodGet:
		h.handleDashboard(w, r, tenant.ID, framework)
	case suffix == "/evaluate" && r.Method == http.MethodPost:
		h.handleEvaluate(w, r, tenant.ID, user.ID, framework)
	case suffix == "/self-certify" && r.Method == http.MethodPost:
		h.handleSelfCertify(w, r, tenant.ID, user.ID, framework)
	case suffix == "/report" && r.Method == http.MethodGet:
		h.handleReport(w, r, tenant.ID, user.ID, framework)
	case strings.HasPrefix(suffix, "/evidence/") && r.Method == http.MethodGet:
		h.handleEvidence(w, r, tenant.ID)
	default:
		responders.MethodNotAllowed(w, http.MethodGet, http.MethodPost)
	}
}

func (h *Handler) handleDashboard(w http.ResponseWriter, r *http.Request, tenantID string, framework Framework) {
	dashboard, err := h.service.BuildDashboard(r.Context(), tenantID, framework)
	if err != nil {
		responders.Error(w, http.StatusInternalServerError, "abatement_dashboard_failed", err.Error())
		return
	}
	responders.JSON(w, http.StatusOK, dashboard)
}

func (h *Handler) handleEvaluate(w http.ResponseWriter, r *http.Request, tenantID, userID string, framework Framework) {
	req, err := decodeEvaluateRequest(r)
	if err != nil {
		responders.BadRequest(w, "invalid_request", err.Error())
		return
	}

	card, err := h.service.Evaluate(r.Context(), tenantID, userID, framework, req)
	if err != nil {
		responders.Error(w, http.StatusInternalServerError, "abatement_evaluation_failed", err.Error())
		return
	}

	responders.JSON(w, http.StatusOK, map[string]any{
		"risk": card,
		"evaluation": map[string]any{
			"status":          card.EngineStatus,
			"feedback":        card.EngineFeedback,
			"criteriaChecked": card.CriteriaChecked,
		},
	})
}

func (h *Handler) handleSelfCertify(w http.ResponseWriter, r *http.Request, tenantID, userID string, framework Framework) {
	var req SelfCertificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responders.BadRequest(w, "invalid_request", "invalid JSON body")
		return
	}

	card, err := h.service.SelfCertify(r.Context(), tenantID, userID, framework, req)
	if err != nil {
		responders.Error(w, http.StatusInternalServerError, "abatement_self_certification_failed", err.Error())
		return
	}
	responders.JSON(w, http.StatusOK, map[string]any{"risk": card})
}

func (h *Handler) handleReport(w http.ResponseWriter, r *http.Request, tenantID, userID string, framework Framework) {
	content, filename, err := h.service.GenerateReport(r.Context(), tenantID, userID, framework)
	if err != nil {
		responders.Error(w, http.StatusInternalServerError, "abatement_report_failed", err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	w.Header().Set("Content-Length", strconv.Itoa(len(content)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(content)
}

func (h *Handler) handleEvidence(w http.ResponseWriter, r *http.Request, tenantID string) {
	evidenceID := strings.TrimPrefix(r.URL.Path, "/api/abatement/")
	parts := strings.Split(evidenceID, "/")
	if len(parts) < 3 {
		responders.NotFound(w, "abatement evidence")
		return
	}
	evidenceID = parts[len(parts)-1]
	record, err := h.service.GetEvidence(r.Context(), tenantID, evidenceID)
	if err != nil {
		responders.NotFound(w, err.Error())
		return
	}
	w.Header().Set("Content-Type", record.MimeType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", record.FileName))
	w.Header().Set("Content-Length", strconv.FormatInt(record.SizeBytes, 10))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(record.Content)
}

func parseFrameworkPath(path string) (Framework, string, error) {
	trimmed := strings.TrimPrefix(path, "/api/abatement/")
	trimmed = strings.Trim(trimmed, "/")
	if trimmed == "" {
		return "", "", fmt.Errorf("framework is required")
	}
	parts := strings.Split(trimmed, "/")
	framework := Framework(parts[0])
	if _, err := FrameworkMetadata(framework); err != nil {
		return "", "", err
	}
	if len(parts) == 1 {
		return framework, "", nil
	}
	return framework, "/" + strings.Join(parts[1:], "/"), nil
}

func decodeEvaluateRequest(r *http.Request) (EvaluateRequest, error) {
	contentType := r.Header.Get("Content-Type")
	if mediaType, _, err := mime.ParseMediaType(contentType); err == nil && strings.HasPrefix(mediaType, "multipart/") {
		return decodeMultipartEvaluateRequest(r)
	}

	var req struct {
		ComplianceCheckID string `json:"compliance_check_id"`
		Completed         bool   `json:"completed"`
		Justification     string `json:"justification"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return EvaluateRequest{}, fmt.Errorf("invalid JSON body")
	}

	if strings.TrimSpace(req.ComplianceCheckID) == "" {
		return EvaluateRequest{}, fmt.Errorf("compliance_check_id is required")
	}

	return EvaluateRequest{
		ComplianceCheckID: req.ComplianceCheckID,
		Completed:         req.Completed,
		Justification:     req.Justification,
	}, nil
}

func decodeMultipartEvaluateRequest(r *http.Request) (EvaluateRequest, error) {
	if err := r.ParseMultipartForm(12 << 20); err != nil {
		return EvaluateRequest{}, fmt.Errorf("invalid multipart form")
	}

	req := EvaluateRequest{
		ComplianceCheckID: strings.TrimSpace(r.FormValue("compliance_check_id")),
		Justification:     strings.TrimSpace(r.FormValue("justification")),
	}
	req.Completed = parseBoolFormValue(r.FormValue("completed"))

	if req.ComplianceCheckID == "" {
		return EvaluateRequest{}, fmt.Errorf("compliance_check_id is required")
	}

	if r.MultipartForm != nil && r.MultipartForm.File != nil {
		files := r.MultipartForm.File["evidence"]
		req.Evidence = make([]EvidenceUpload, 0, len(files))
		for _, header := range files {
			file, err := header.Open()
			if err != nil {
				return EvaluateRequest{}, fmt.Errorf("open uploaded file %s: %w", header.Filename, err)
			}
			content, readErr := io.ReadAll(io.LimitReader(file, 10<<20))
			_ = file.Close()
			if readErr != nil {
				return EvaluateRequest{}, fmt.Errorf("read uploaded file %s: %w", header.Filename, readErr)
			}
			req.Evidence = append(req.Evidence, EvidenceUpload{
				FileName: header.Filename,
				MimeType: header.Header.Get("Content-Type"),
				Content:  content,
			})
		}
	}

	return req, nil
}

func parseBoolFormValue(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
