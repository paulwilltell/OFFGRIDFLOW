package abatement

import (
	"regexp"
	"strings"
)

var marketInstrumentPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\brenewable energy certificate(s)?\b`),
	regexp.MustCompile(`(?i)\brec(s)?\b`),
	regexp.MustCompile(`(?i)\bgreen power\b`),
	regexp.MustCompile(`(?i)\bppa\b`),
	regexp.MustCompile(`(?i)\bpower purchase agreement(s)?\b`),
	regexp.MustCompile(`(?i)\bguarantee of origin\b`),
	regexp.MustCompile(`(?i)\beac(s)?\b`),
	regexp.MustCompile(`(?i)\bresidual mix\b`),
}

type evaluatorFunc func(justification string, evidence []EvidenceUpload) Evaluation

func evaluateScope1Action(justification string, evidence []EvidenceUpload) Evaluation {
	checks := matchKeywords(justification,
		keywordCheck("source data reference", "fuel invoice", "fuel log", "stationary combustion", "fleet fuel", "gas bill", "meter read"),
		keywordCheck("operational boundary", "owned vehicle", "leased vehicle", "boiler", "generator", "facility"),
		keywordCheck("resolution statement", "uploaded", "reconciled", "backfilled", "corrected", "completed"),
	)
	hasEvidence := hasEvidence(evidence)

	switch {
	case len(checks) >= 2 && hasEvidence:
		return Evaluation{
			Status:          StatusRecommended,
			Feedback:        "The explanation identifies the direct-emissions source, describes the corrective action, and includes supporting evidence.",
			CriteriaChecked: append(checks, "supporting evidence"),
		}
	case len(checks) >= 2:
		return Evaluation{
			Status:          StatusNeedsClarification,
			Feedback:        "The corrective action is described, but supporting source evidence is still missing. Add invoices, meter logs, or fuel records so reviewers can trace the change.",
			CriteriaChecked: checks,
		}
	default:
		return Evaluation{
			Status:          StatusInsufficient,
			Feedback:        "The explanation does not clearly identify the Scope 1 source, the correction performed, and the evidence trail. Describe the source, what changed, and attach the underlying record.",
			CriteriaChecked: checks,
		}
	}
}

func evaluateScope2ActivityAction(justification string, evidence []EvidenceUpload) Evaluation {
	checks := matchKeywords(justification,
		keywordCheck("energy record reference", "utility bill", "invoice", "meter", "kwh", "mwh", "electricity"),
		keywordCheck("location or account detail", "account", "site", "facility", "office", "meter id"),
		keywordCheck("resolution statement", "uploaded", "imported", "corrected", "mapped", "reconciled"),
	)
	hasEvidence := hasEvidence(evidence)

	switch {
	case len(checks) >= 2 && hasEvidence:
		return Evaluation{
			Status:          StatusRecommended,
			Feedback:        "The explanation references the purchased-energy records, the site or account they cover, and includes evidence that supports the correction.",
			CriteriaChecked: append(checks, "supporting evidence"),
		}
	case len(checks) >= 2:
		return Evaluation{
			Status:          StatusNeedsClarification,
			Feedback:        "The action is plausible, but a reviewer will expect the underlying bill, meter export, or invoice. Attach the supporting record to strengthen the file.",
			CriteriaChecked: checks,
		}
	default:
		return Evaluation{
			Status:          StatusInsufficient,
			Feedback:        "Scope 2 remediation should identify the purchased-energy record, the affected account or site, and what was corrected. The current explanation does not do that yet.",
			CriteriaChecked: checks,
		}
	}
}

func evaluateScope2MarketBased(justification string, evidence []EvidenceUpload) Evaluation {
	hasInstrument := hasMarketInstrumentReference(justification)
	hasCoverage := containsAny(justification,
		"office", "facility", "site", "meter", "account", "all locations", "three offices", "portfolio",
	)
	hasEvidence := hasEvidence(evidence)

	checks := make([]string, 0, 3)
	if hasInstrument {
		checks = append(checks, "market-based instrument reference")
	}
	if hasCoverage {
		checks = append(checks, "portfolio coverage detail")
	}
	if hasEvidence {
		checks = append(checks, "supporting evidence")
	}

	switch {
	case hasInstrument && hasCoverage && hasEvidence:
		return Evaluation{
			Status:          StatusRecommended,
			Feedback:        "Market-based instruments are referenced, the covered sites are identified, and supporting evidence was uploaded. This appears sufficient for draft reporting support.",
			CriteriaChecked: checks,
		}
	case hasInstrument && hasEvidence:
		return Evaluation{
			Status:          StatusNeedsClarification,
			Feedback:        "Market-based instruments are referenced and evidence is present, but the explanation does not clearly state which meters, sites, or reporting period are covered. Add coverage detail before relying on it.",
			CriteriaChecked: checks,
		}
	case hasInstrument && !hasEvidence:
		return Evaluation{
			Status:          StatusNeedsClarification,
			Feedback:        "Market-based instruments are mentioned, but no supporting evidence was uploaded. Reviewers typically expect certificates, contracts, or supplier documentation.",
			CriteriaChecked: checks,
		}
	default:
		return Evaluation{
			Status:          StatusInsufficient,
			Feedback:        "No market-based instrument is described. For market-based Scope 2 support, describe the REC, contract, tariff, or residual-mix evidence and attach the supporting file.",
			CriteriaChecked: checks,
		}
	}
}

func hasMarketInstrumentReference(justification string) bool {
	for _, pattern := range marketInstrumentPatterns {
		if pattern.MatchString(justification) {
			return true
		}
	}
	return false
}

func evaluateScope3SupplierAction(justification string, evidence []EvidenceUpload) Evaluation {
	checks := matchKeywords(justification,
		keywordCheck("supplier identification", "supplier", "vendor", "supplier name", "supplier id", "procurement"),
		keywordCheck("primary-data method", "pcf", "questionnaire", "supplier data", "activity data", "emission factor", "survey"),
		keywordCheck("resolution statement", "uploaded", "mapped", "collected", "completed", "reconciled"),
	)
	hasEvidence := hasEvidence(evidence)

	switch {
	case len(checks) >= 2 && hasEvidence:
		return Evaluation{
			Status:          StatusRecommended,
			Feedback:        "The explanation identifies the supplier data source, describes the correction, and includes evidence. This is suitable for a draft audit trail.",
			CriteriaChecked: append(checks, "supporting evidence"),
		}
	case len(checks) >= 2:
		return Evaluation{
			Status:          StatusNeedsClarification,
			Feedback:        "Supplier remediation is described, but there is no attached evidence. Add supplier questionnaires, factor schedules, or reconciled exports to support the claim.",
			CriteriaChecked: checks,
		}
	default:
		return Evaluation{
			Status:          StatusInsufficient,
			Feedback:        "The explanation does not show which supplier data was corrected or how it now supports Scope 3 reporting. Identify the supplier source and attach the supporting export or document.",
			CriteriaChecked: checks,
		}
	}
}

func evaluateFactorSnapshotAction(justification string, evidence []EvidenceUpload) Evaluation {
	checks := matchKeywords(justification,
		keywordCheck("factor version reference", "factor snapshot", "factor version", "frozen factor", "locked snapshot", "registry version"),
		keywordCheck("period coverage", "reporting year", "period", "inventory", "baseline"),
	)
	hasEvidence := hasEvidence(evidence)

	switch {
	case len(checks) == 2 && hasEvidence:
		return Evaluation{
			Status:          StatusRecommended,
			Feedback:        "The explanation references a locked factor version for the reporting period and includes evidence supporting reproducibility.",
			CriteriaChecked: append(checks, "supporting evidence"),
		}
	case len(checks) >= 1:
		return Evaluation{
			Status:          StatusNeedsClarification,
			Feedback:        "A factor version is referenced, but the reporting period and supporting evidence are not fully documented. Add the locked version identifier and attach the export or screenshot.",
			CriteriaChecked: checks,
		}
	default:
		return Evaluation{
			Status:          StatusInsufficient,
			Feedback:        "No locked factor version or reporting-period snapshot is described. Document the frozen factor set used for the reporting period and attach supporting evidence.",
			CriteriaChecked: checks,
		}
	}
}

func evaluateApprovalWorkflowAction(justification string, evidence []EvidenceUpload) Evaluation {
	checks := matchKeywords(justification,
		keywordCheck("review participant", "reviewed by", "approved by", "controller", "cfo", "sustainability lead", "signoff"),
		keywordCheck("workflow step", "submitted", "reviewed", "approved", "rejected", "commented", "escalated"),
	)
	hasEvidence := hasEvidence(evidence)

	switch {
	case len(checks) == 2 && hasEvidence:
		return Evaluation{
			Status:          StatusRecommended,
			Feedback:        "The explanation identifies the reviewer or approver, describes the workflow step, and includes evidence of the decision.",
			CriteriaChecked: append(checks, "supporting evidence"),
		}
	case len(checks) >= 1:
		return Evaluation{
			Status:          StatusNeedsClarification,
			Feedback:        "The workflow is described, but the evidence trail is incomplete. Attach a signoff export, ticket, or approval record before relying on this control.",
			CriteriaChecked: checks,
		}
	default:
		return Evaluation{
			Status:          StatusInsufficient,
			Feedback:        "No review or approval control is documented. Describe who reviewed the item, what decision was made, and attach the evidence trail.",
			CriteriaChecked: checks,
		}
	}
}

func evaluateMeasuredDataAction(justification string, evidence []EvidenceUpload) Evaluation {
	checks := matchKeywords(justification,
		keywordCheck("measured source", "invoice", "meter", "erp", "general ledger", "utility export", "statement"),
		keywordCheck("coverage statement", "all sites", "remaining sites", "sample", "reporting period", "coverage"),
	)
	hasEvidence := hasEvidence(evidence)

	switch {
	case len(checks) == 2 && hasEvidence:
		return Evaluation{
			Status:          StatusRecommended,
			Feedback:        "Measured-source coverage is described with supporting evidence. This supports a stronger data-quality position for the item.",
			CriteriaChecked: append(checks, "supporting evidence"),
		}
	case len(checks) >= 1:
		return Evaluation{
			Status:          StatusNeedsClarification,
			Feedback:        "The explanation suggests measured data was added, but reviewers will still expect supporting exports or invoices and a clearer statement of coverage.",
			CriteriaChecked: checks,
		}
	default:
		return Evaluation{
			Status:          StatusInsufficient,
			Feedback:        "The explanation does not describe measured data sources or improved coverage. Explain which records were added and provide the supporting files.",
			CriteriaChecked: checks,
		}
	}
}

func evaluateImportedGoodsAction(justification string, evidence []EvidenceUpload) Evaluation {
	checks := matchKeywords(justification,
		keywordCheck("goods traceability", "customs", "import declaration", "commodity code", "cn code", "hs code", "country of origin"),
		keywordCheck("supplier document", "mill certificate", "supplier statement", "commercial invoice", "shipping document"),
	)
	hasEvidence := hasEvidence(evidence)

	switch {
	case len(checks) >= 2 && hasEvidence:
		return Evaluation{
			Status:          StatusRecommended,
			Feedback:        "Imported-goods traceability is documented and supporting records were uploaded. This materially improves draft CBAM support.",
			CriteriaChecked: append(checks, "supporting evidence"),
		}
	case len(checks) >= 1:
		return Evaluation{
			Status:          StatusNeedsClarification,
			Feedback:        "Some CBAM traceability details are present, but customs and supplier evidence is still incomplete. Add the missing import or supplier documents.",
			CriteriaChecked: checks,
		}
	default:
		return Evaluation{
			Status:          StatusInsufficient,
			Feedback:        "The explanation does not document the imported good, commodity classification, or supplier evidence. Add that information before relying on the remediation.",
			CriteriaChecked: checks,
		}
	}
}

func evaluateSupplierSpecificCBAMAction(justification string, evidence []EvidenceUpload) Evaluation {
	checks := matchKeywords(justification,
		keywordCheck("supplier-specific emissions", "supplier-specific", "actual emissions", "pcf", "plant data", "mill data", "verified intensity"),
		keywordCheck("method traceability", "calculation method", "boundary", "assumptions", "period", "verification"),
	)
	hasEvidence := hasEvidence(evidence)

	switch {
	case len(checks) == 2 && hasEvidence:
		return Evaluation{
			Status:          StatusRecommended,
			Feedback:        "Supplier-specific emissions data and the supporting method trace are both described with evidence attached.",
			CriteriaChecked: append(checks, "supporting evidence"),
		}
	case len(checks) >= 1:
		return Evaluation{
			Status:          StatusNeedsClarification,
			Feedback:        "Supplier-specific data is mentioned, but the method boundary or supporting evidence is incomplete. Add the missing supporting file and methodological detail.",
			CriteriaChecked: checks,
		}
	default:
		return Evaluation{
			Status:          StatusInsufficient,
			Feedback:        "No supplier-specific emissions evidence or method trace is documented. Add the supplier record, calculation boundary, and supporting file.",
			CriteriaChecked: checks,
		}
	}
}

type keywordCriterion struct {
	label    string
	keywords []string
}

func keywordCheck(label string, keywords ...string) keywordCriterion {
	return keywordCriterion{label: label, keywords: keywords}
}

func matchKeywords(justification string, checks ...keywordCriterion) []string {
	found := make([]string, 0, len(checks))
	for _, check := range checks {
		if containsAny(justification, check.keywords...) {
			found = append(found, check.label)
		}
	}
	return found
}

func containsAny(haystack string, needles ...string) bool {
	normalized := strings.ToLower(haystack)
	for _, needle := range needles {
		if strings.Contains(normalized, strings.ToLower(needle)) {
			return true
		}
	}
	return false
}

func hasEvidence(evidence []EvidenceUpload) bool {
	for _, file := range evidence {
		if len(file.Content) == 0 {
			continue
		}
		return true
	}
	return false
}

