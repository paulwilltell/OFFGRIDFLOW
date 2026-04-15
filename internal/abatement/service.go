package abatement

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/example/offgridflow/internal/compliance"
	"github.com/example/offgridflow/internal/ingestion"
	"github.com/jung-kurt/gofpdf"
)

type Service struct {
	db                *sql.DB
	activityStore     ingestion.ActivityStore
	complianceService *compliance.Service
	store             *Store
}

func NewService(db *sql.DB, activityStore ingestion.ActivityStore, complianceService *compliance.Service) *Service {
	return &Service{
		db:                db,
		activityStore:     activityStore,
		complianceService: complianceService,
		store:             NewStore(db),
	}
}

func (s *Service) BuildDashboard(ctx context.Context, tenantID string, framework Framework) (*Dashboard, error) {
	meta, err := FrameworkMetadata(framework)
	if err != nil {
		return nil, err
	}

	definitions, err := FrameworkDefinitions(framework)
	if err != nil {
		return nil, err
	}
	if err := s.store.SyncDefinitions(ctx, tenantID, framework, definitions); err != nil {
		return nil, err
	}

	activities, err := s.activityStore.ListByOrg(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("load tenant activities: %w", err)
	}
	year := latestReportingYear(activities)
	filtered := filterActivitiesByYear(activities, year)

	facts, err := s.buildRiskFacts(ctx, tenantID, framework, year, filtered)
	if err != nil {
		return nil, err
	}

	storedItems, err := s.store.ListActionItems(ctx, tenantID, framework)
	if err != nil {
		return nil, err
	}
	itemByCheckID := make(map[string]StoredActionItem, len(storedItems))
	for _, item := range storedItems {
		itemByCheckID[item.ComplianceCheckID] = item
	}

	risks := make([]RiskCard, 0, len(definitions))
	summary := RiskSummary{}
	progress := ProgressSummary{}
	for _, def := range definitions {
		if !def.IsTriggered(facts) {
			continue
		}

		item := itemByCheckID[def.CheckID]
		evidence, err := s.store.ListEvidenceRecords(ctx, tenantID, item.ID, framework)
		if err != nil {
			return nil, err
		}
		card := RiskCard{
			ID:                    item.ID,
			ComplianceCheckID:     def.CheckID,
			Title:                 def.Title,
			Severity:              def.Severity,
			Priority:              def.Priority,
			Description:           def.Description,
			AcceptanceCriteria:    def.AcceptanceCriteria,
			RequiredEvidenceTypes: def.RequiredEvidenceTypes,
			Justification:         item.Justification,
			Evidence:              evidence,
			EngineStatus:          item.EngineStatus,
			EngineFeedback:        item.EngineFeedback,
			CriteriaChecked:       item.CriteriaChecked,
			Completed:             item.Completed,
			SelfCertified:         item.SelfCertified,
			CertifiedAt:           item.CertifiedAt,
			UpdatedAt:             &item.UpdatedAt,
		}
		risks = append(risks, card)

		switch def.Priority {
		case PriorityHigh:
			summary.High++
		case PriorityMedium:
			summary.Medium++
		default:
			summary.Low++
		}

		if def.Severity == SeverityBlocker {
			progress.CriticalTotal++
			if card.Completed || card.SelfCertified {
				progress.CriticalAddressed++
			}
		}
	}

	sort.Slice(risks, func(i, j int) bool {
		pi := priorityRank(risks[i].Priority)
		pj := priorityRank(risks[j].Priority)
		if pi != pj {
			return pi < pj
		}
		return risks[i].Title < risks[j].Title
	})

	if progress.CriticalTotal == 0 {
		progress.PercentAddressed = 100
	} else {
		progress.PercentAddressed = int(float64(progress.CriticalAddressed) / float64(progress.CriticalTotal) * 100)
	}

	return &Dashboard{
		Framework:     meta,
		Summary:       summary,
		Progress:      progress,
		Disclaimer:    GuidanceDisclaimer,
		ReportingYear: year,
		Risks:         risks,
		GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (s *Service) Evaluate(ctx context.Context, tenantID, userID string, framework Framework, req EvaluateRequest) (*RiskCard, error) {
	if strings.TrimSpace(req.ComplianceCheckID) == "" {
		return nil, fmt.Errorf("compliance_check_id is required")
	}

	definitions, err := FrameworkDefinitions(framework)
	if err != nil {
		return nil, err
	}
	if err := s.store.SyncDefinitions(ctx, tenantID, framework, definitions); err != nil {
		return nil, err
	}

	item, err := s.store.GetActionItemByCheckID(ctx, tenantID, framework, req.ComplianceCheckID)
	if err != nil {
		return nil, err
	}

	definition, err := findDefinition(definitions, req.ComplianceCheckID)
	if err != nil {
		return nil, err
	}
	evaluation := definition.Evaluator(req.Justification, req.Evidence)

	updated, err := s.store.SaveEvaluation(
		ctx,
		tenantID,
		userID,
		framework,
		item.ID,
		req.Justification,
		req.Completed,
		evaluation,
		req.Evidence,
	)
	if err != nil {
		return nil, err
	}

	evidence, err := s.store.ListEvidenceRecords(ctx, tenantID, updated.ID, framework)
	if err != nil {
		return nil, err
	}

	card := &RiskCard{
		ID:                    updated.ID,
		ComplianceCheckID:     updated.ComplianceCheckID,
		Title:                 updated.Title,
		Severity:              updated.Severity,
		Priority:              updated.Priority,
		Description:           updated.Description,
		AcceptanceCriteria:    updated.AcceptanceCriteria,
		RequiredEvidenceTypes: updated.RequiredEvidenceTypes,
		Justification:         updated.Justification,
		Evidence:              evidence,
		EngineStatus:          updated.EngineStatus,
		EngineFeedback:        updated.EngineFeedback,
		CriteriaChecked:       updated.CriteriaChecked,
		Completed:             updated.Completed,
		SelfCertified:         updated.SelfCertified,
		CertifiedAt:           updated.CertifiedAt,
		UpdatedAt:             &updated.UpdatedAt,
	}
	return card, nil
}

func (s *Service) SelfCertify(ctx context.Context, tenantID, userID string, framework Framework, req SelfCertificationRequest) (*RiskCard, error) {
	updated, err := s.store.SetSelfCertified(ctx, tenantID, userID, framework, req.ActionItemID, req.ComplianceCheckID, req.SelfCertified)
	if err != nil {
		return nil, err
	}
	evidence, err := s.store.ListEvidenceRecords(ctx, tenantID, updated.ID, framework)
	if err != nil {
		return nil, err
	}
	return &RiskCard{
		ID:                    updated.ID,
		ComplianceCheckID:     updated.ComplianceCheckID,
		Title:                 updated.Title,
		Severity:              updated.Severity,
		Priority:              updated.Priority,
		Description:           updated.Description,
		AcceptanceCriteria:    updated.AcceptanceCriteria,
		RequiredEvidenceTypes: updated.RequiredEvidenceTypes,
		Justification:         updated.Justification,
		Evidence:              evidence,
		EngineStatus:          updated.EngineStatus,
		EngineFeedback:        updated.EngineFeedback,
		CriteriaChecked:       updated.CriteriaChecked,
		Completed:             updated.Completed,
		SelfCertified:         updated.SelfCertified,
		CertifiedAt:           updated.CertifiedAt,
		UpdatedAt:             &updated.UpdatedAt,
	}, nil
}

func (s *Service) GenerateReport(ctx context.Context, tenantID, userID string, framework Framework) ([]byte, string, error) {
	dashboard, err := s.BuildDashboard(ctx, tenantID, framework)
	if err != nil {
		return nil, "", err
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetTitle(fmt.Sprintf("%s Risk Abatement Workplan", dashboard.Framework.Label), false)
	pdf.SetAuthor("OffGridFlow", false)
	pdf.SetCreator("OffGridFlow Risk Abatement Engine", false)
	pdf.AliasNbPages("")
	pdf.SetMargins(15, 18, 15)
	pdf.SetAutoPageBreak(true, 18)
	pdf.SetFooterFunc(func() {
		pdf.SetY(-14)
		pdf.SetFont("Arial", "I", 8)
		pdf.SetTextColor(95, 95, 95)
		pdf.MultiCell(0, 3.5, GuidanceDisclaimer, "", "L", false)
		pdf.SetX(15)
		pdf.CellFormat(0, 4, fmt.Sprintf("Page %d/{nb}", pdf.PageNo()), "", 0, "R", false, 0, "")
	})

	pdf.AddPage()
	pdf.SetFont("Arial", "B", 18)
	pdf.Cell(0, 10, fmt.Sprintf("%s Risk Abatement Workplan", dashboard.Framework.Label))
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(0, 6, "This report summarizes identified risks, customer-provided justifications, self-certification choices, and OffGridFlow assessment results. Final filing decisions remain with the customer.", "", "L", false)
	pdf.Ln(2)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 7, "Summary")
	pdf.Ln(8)
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(60, 6, "High-priority risks", "1", 0, "L", false, 0, "")
	pdf.CellFormat(20, 6, fmt.Sprintf("%d", dashboard.Summary.High), "1", 1, "R", false, 0, "")
	pdf.CellFormat(60, 6, "Medium-priority risks", "1", 0, "L", false, 0, "")
	pdf.CellFormat(20, 6, fmt.Sprintf("%d", dashboard.Summary.Medium), "1", 1, "R", false, 0, "")
	pdf.CellFormat(60, 6, "Low-priority risks", "1", 0, "L", false, 0, "")
	pdf.CellFormat(20, 6, fmt.Sprintf("%d", dashboard.Summary.Low), "1", 1, "R", false, 0, "")
	pdf.CellFormat(60, 6, "Critical risks addressed", "1", 0, "L", false, 0, "")
	pdf.CellFormat(20, 6, fmt.Sprintf("%d/%d", dashboard.Progress.CriticalAddressed, dashboard.Progress.CriticalTotal), "1", 1, "R", false, 0, "")
	pdf.Ln(4)
	pdf.MultiCell(0, 5, dashboard.Framework.PenaltyBody, "", "L", false)

	for _, risk := range dashboard.Risks {
		pdf.AddPage()
		pdf.SetFont("Arial", "B", 15)
		pdf.MultiCell(0, 7, risk.Title, "", "L", false)
		pdf.Ln(1)

		pdf.SetFont("Arial", "", 10)
		pdf.CellFormat(30, 6, "Severity", "1", 0, "L", false, 0, "")
		pdf.CellFormat(45, 6, titleCase(string(risk.Severity)), "1", 0, "L", false, 0, "")
		pdf.CellFormat(30, 6, "Priority", "1", 0, "L", false, 0, "")
		pdf.CellFormat(45, 6, titleCase(string(risk.Priority)), "1", 1, "L", false, 0, "")
		pdf.CellFormat(30, 6, "Completed", "1", 0, "L", false, 0, "")
		pdf.CellFormat(45, 6, yesNo(risk.Completed), "1", 0, "L", false, 0, "")
		pdf.CellFormat(30, 6, "Self-certified", "1", 0, "L", false, 0, "")
		pdf.CellFormat(45, 6, yesNo(risk.SelfCertified), "1", 1, "L", false, 0, "")
		pdf.Ln(4)

		pdf.SetFont("Arial", "B", 11)
		pdf.Cell(0, 6, "Risk description")
		pdf.Ln(7)
		pdf.SetFont("Arial", "", 10)
		pdf.MultiCell(0, 5.5, risk.Description, "", "L", false)
		pdf.Ln(2)

		pdf.SetFont("Arial", "B", 11)
		pdf.Cell(0, 6, "Acceptance criteria")
		pdf.Ln(7)
		pdf.SetFont("Arial", "", 10)
		for _, criterion := range risk.AcceptanceCriteria {
			pdf.MultiCell(0, 5, "• "+criterion, "", "L", false)
		}
		pdf.Ln(2)

		pdf.SetFont("Arial", "B", 11)
		pdf.Cell(0, 6, "Customer justification")
		pdf.Ln(7)
		pdf.SetFont("Arial", "", 10)
		if strings.TrimSpace(risk.Justification) == "" {
			pdf.MultiCell(0, 5, "No justification submitted.", "", "L", false)
		} else {
			pdf.MultiCell(0, 5, risk.Justification, "", "L", false)
		}
		pdf.Ln(2)

		pdf.SetFont("Arial", "B", 11)
		pdf.Cell(0, 6, "OffGridFlow assessment")
		pdf.Ln(7)
		pdf.SetFont("Arial", "", 10)
		statusLabel := "No assessment yet"
		if risk.EngineStatus != "" {
			statusLabel = strings.ReplaceAll(string(risk.EngineStatus), "_", " ")
		}
		pdf.MultiCell(0, 5, fmt.Sprintf("Status: %s", titleCase(statusLabel)), "", "L", false)
		if strings.TrimSpace(risk.EngineFeedback) != "" {
			pdf.MultiCell(0, 5, risk.EngineFeedback, "", "L", false)
		}
		if len(risk.CriteriaChecked) > 0 {
			pdf.MultiCell(0, 5, "Criteria checked: "+strings.Join(risk.CriteriaChecked, ", "), "", "L", false)
		}
		if len(risk.Evidence) > 0 {
			pdf.Ln(2)
			pdf.MultiCell(0, 5, "Evidence files:", "", "L", false)
			for _, file := range risk.Evidence {
				pdf.MultiCell(0, 5, fmt.Sprintf("• %s (%s, %d bytes)", file.FileName, file.MimeType, file.SizeBytes), "", "L", false)
			}
		}
	}

	if err := s.logReportGenerated(ctx, tenantID, userID, framework, len(dashboard.Risks)); err != nil {
		return nil, "", err
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, "", fmt.Errorf("render abatement report pdf: %w", err)
	}

	filename := fmt.Sprintf("%s-risk-abatement-%d.pdf", framework, dashboard.ReportingYear)
	return buf.Bytes(), filename, nil
}

func (s *Service) GetEvidence(ctx context.Context, tenantID, evidenceID string, framework Framework) (*StoredEvidence, error) {
	return s.store.GetEvidence(ctx, tenantID, evidenceID, framework)
}

func (s *Service) buildRiskFacts(ctx context.Context, tenantID string, framework Framework, year int, activities []ingestion.Activity) (RiskFacts, error) {
	facts := RiskFacts{
		ReportingYear: year,
	}

	if s.complianceService != nil {
		summary, err := s.complianceService.GenerateSummary(ctx, tenantID, year)
		if err != nil {
			return facts, fmt.Errorf("generate compliance summary for abatement: %w", err)
		}
		key := complianceSummaryKey(framework)
		frameworkSummary, ok := summary.Frameworks[key]
		if ok {
			facts.Scope1Ready = frameworkSummary.Scope1
			facts.Scope2Ready = frameworkSummary.Scope2
			facts.Scope3Ready = frameworkSummary.Scope3
		} else {
			facts.Scope1Ready = summary.Totals.Scope1Tons > 0
			facts.Scope2Ready = summary.Totals.Scope2Tons > 0
			facts.Scope3Ready = summary.Totals.Scope3Tons > 0
		}
	}

	var measuredCount int
	var purchaseCount int
	var purchaseWithSupplierData int
	var importEvidenceCount int
	for _, activity := range activities {
		if isElectricityActivity(activity) {
			facts.HasElectricityActivities = true
		}
		if isPurchaseActivity(activity) {
			facts.HasPurchaseActivities = true
			purchaseCount++
		}
		if hasSupplierMetadata(activity.Metadata) {
			facts.HasSupplierMetadata = true
			if isPurchaseActivity(activity) {
				purchaseWithSupplierData++
			}
		}
		if hasMarketBasedSignals(activity.Metadata) {
			facts.HasMarketBasedSignals = true
		}
		if strings.EqualFold(activity.DataQuality, "measured") {
			measuredCount++
		}
		if isImportedGoodsEvidence(activity.Metadata) {
			importEvidenceCount++
		}
	}

	if len(activities) > 0 {
		facts.MeasuredDataRatio = float64(measuredCount) / float64(len(activities))
	}
	if purchaseCount > 0 {
		facts.SupplierSpecificRatio = float64(purchaseWithSupplierData) / float64(purchaseCount)
		facts.ImportedGoodsRatio = float64(importEvidenceCount) / float64(purchaseCount)
	}

	if s.db != nil {
		snapshotCount, err := s.countLockedFactorSnapshots(ctx, tenantID, year)
		if err != nil {
			return facts, err
		}
		facts.HasLockedFactorSnapshot = snapshotCount > 0

		approvalCount, err := s.countApprovals(ctx, tenantID, year)
		if err != nil {
			return facts, err
		}
		facts.HasApprovalWorkflow = approvalCount > 0
	}

	return facts, nil
}

func (s *Service) countLockedFactorSnapshots(ctx context.Context, tenantID string, year int) (int, error) {
	if s.db == nil {
		return 0, nil
	}
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
	end := time.Date(year, 12, 31, 0, 0, 0, 0, time.UTC).Format("2006-01-02")

	var count int
	err := s.db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM factor_snapshots
WHERE organization_id = $1
  AND status = 'locked'
  AND reporting_period_start <= $3
  AND reporting_period_end >= $2`, tenantID, start, end).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count locked factor snapshots: %w", err)
	}
	return count, nil
}

func (s *Service) countApprovals(ctx context.Context, tenantID string, year int) (int, error) {
	if s.db == nil {
		return 0, nil
	}
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)

	var count int
	err := s.db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM approval_workflow
WHERE tenant_id = $1
  AND updated_at >= $2
  AND updated_at < $3`, tenantID, start, end).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count approval workflow records: %w", err)
	}
	return count, nil
}

func (s *Service) logReportGenerated(ctx context.Context, tenantID, userID string, framework Framework, riskCount int) error {
	tx, err := s.store.beginTenantTx(ctx, tenantID)
	if err != nil {
		return err
	}
	defer rollback(tx)

	if err := s.store.logAuditEventTx(ctx, tx, tenantID, userID, "abatement.report_generated", string(framework), map[string]any{
		"framework":  framework,
		"risk_count": riskCount,
	}); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit abatement report audit log: %w", err)
	}
	return nil
}

func latestReportingYear(activities []ingestion.Activity) int {
	year := 0
	for _, activity := range activities {
		if activity.PeriodStart.IsZero() {
			continue
		}
		if activity.PeriodStart.Year() > year {
			year = activity.PeriodStart.Year()
		}
	}
	if year == 0 {
		return time.Now().Year()
	}
	return year
}

func filterActivitiesByYear(activities []ingestion.Activity, year int) []ingestion.Activity {
	filtered := make([]ingestion.Activity, 0, len(activities))
	for _, activity := range activities {
		if activity.PeriodStart.IsZero() {
			filtered = append(filtered, activity)
			continue
		}
		if activity.PeriodStart.Year() == year {
			filtered = append(filtered, activity)
		}
	}
	if len(filtered) == 0 {
		return activities
	}
	return filtered
}

func complianceSummaryKey(framework Framework) string {
	switch framework {
	case FrameworkSB253:
		return "california"
	case FrameworkIFRS:
		return "ifrs_s2"
	default:
		return string(framework)
	}
}

func isElectricityActivity(activity ingestion.Activity) bool {
	category := strings.ToLower(strings.TrimSpace(activity.Category))
	unit := strings.ToLower(strings.TrimSpace(activity.Unit))
	return activity.IsElectricity() ||
		unit == "kwh" || unit == "mwh" ||
		strings.Contains(category, "electric") ||
		strings.Contains(category, "steam") ||
		strings.Contains(category, "heating") ||
		strings.Contains(category, "cooling")
}

func isPurchaseActivity(activity ingestion.Activity) bool {
	source := strings.ToLower(strings.TrimSpace(activity.Source))
	category := strings.ToLower(strings.TrimSpace(activity.Category))
	return source == string(ingestion.SourcePurchases) ||
		strings.Contains(category, "purchase") ||
		strings.Contains(category, "supplier") ||
		strings.Contains(category, "procurement")
}

func hasSupplierMetadata(metadata map[string]string) bool {
	for key, value := range metadata {
		keyLower := strings.ToLower(key)
		valueLower := strings.ToLower(value)
		if strings.Contains(keyLower, "supplier") || strings.Contains(keyLower, "vendor") {
			if strings.TrimSpace(valueLower) != "" {
				return true
			}
		}
		if strings.Contains(keyLower, "pcf") || strings.Contains(keyLower, "product_carbon") {
			return true
		}
	}
	return false
}

func hasMarketBasedSignals(metadata map[string]string) bool {
	for key, value := range metadata {
		pair := strings.ToLower(key + " " + value)
		if strings.Contains(pair, "rec") ||
			strings.Contains(pair, "renewable energy certificate") ||
			strings.Contains(pair, "green power") ||
			strings.Contains(pair, "power purchase agreement") ||
			strings.Contains(pair, "ppa") ||
			strings.Contains(pair, "market-based") {
			return true
		}
	}
	return false
}

func isImportedGoodsEvidence(metadata map[string]string) bool {
	for key, value := range metadata {
		pair := strings.ToLower(key + " " + value)
		if strings.Contains(pair, "country of origin") ||
			strings.Contains(pair, "commodity code") ||
			strings.Contains(pair, "cn code") ||
			strings.Contains(pair, "hs code") ||
			strings.Contains(pair, "import") ||
			strings.Contains(pair, "customs") {
			return true
		}
	}
	return false
}

func priorityRank(priority RiskPriority) int {
	switch priority {
	case PriorityHigh:
		return 0
	case PriorityMedium:
		return 1
	default:
		return 2
	}
}

func findDefinition(definitions []RiskDefinition, checkID string) (RiskDefinition, error) {
	for _, definition := range definitions {
		if definition.CheckID == checkID {
			return definition, nil
		}
	}
	return RiskDefinition{}, fmt.Errorf("unsupported compliance check %q", checkID)
}

func yesNo(value bool) string {
	if value {
		return "Yes"
	}
	return "No"
}

func titleCase(s string) string {
	if s == "" {
		return s
	}
	parts := strings.Fields(s)
	for i, part := range parts {
		runes := []rune(strings.ToLower(part))
		if len(runes) > 0 {
			runes[0] = []rune(strings.ToUpper(string(runes[0])))[0]
		}
		parts[i] = string(runes)
	}
	return strings.Join(parts, " ")
}
