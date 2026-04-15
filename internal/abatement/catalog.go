package abatement

import "fmt"

var frameworkCatalog = map[Framework]FrameworkMeta{
	FrameworkSB253: {
		Key:            FrameworkSB253,
		Label:          "California SB 253",
		ShortLabel:     "SB 253",
		PenaltyHeading: "Potential penalty exposure",
		PenaltyBody:    "SB 253: Up to $500,000 per reporting year for non-filing or materially deficient disclosure.",
	},
	FrameworkCSRD: {
		Key:            FrameworkCSRD,
		Label:          "CSRD / ESRS E1",
		ShortLabel:     "CSRD",
		PenaltyHeading: "Potential penalty exposure",
		PenaltyBody:    "CSRD: Member-state enforcement varies, but incomplete evidence trails and unsupported disclosures can trigger remediation, delay, and enforcement risk.",
	},
	FrameworkSEC: {
		Key:            FrameworkSEC,
		Label:          "SEC Climate Disclosure",
		ShortLabel:     "SEC",
		PenaltyHeading: "Potential penalty exposure",
		PenaltyBody:    "SEC climate disclosure: unsupported filings can create disclosure, control, and enforcement exposure. Treat this as a governance workplan, not legal advice.",
	},
	FrameworkIFRS: {
		Key:            FrameworkIFRS,
		Label:          "IFRS S2",
		ShortLabel:     "IFRS S2",
		PenaltyHeading: "Potential penalty exposure",
		PenaltyBody:    "IFRS S2: market enforcement depends on the filing regime, but weak evidence trails increase assurance and disclosure risk.",
	},
	FrameworkCBAM: {
		Key:            FrameworkCBAM,
		Label:          "EU CBAM",
		ShortLabel:     "CBAM",
		PenaltyHeading: "Potential penalty exposure",
		PenaltyBody:    "CBAM: incomplete supplier and imported-goods evidence can expose the declarant to corrective action, rejected submissions, and financial risk.",
	},
}

func FrameworkMetadata(framework Framework) (FrameworkMeta, error) {
	meta, ok := frameworkCatalog[framework]
	if !ok {
		return FrameworkMeta{}, fmt.Errorf("unsupported framework %q", framework)
	}
	return meta, nil
}

func FrameworkDefinitions(framework Framework) ([]RiskDefinition, error) {
	switch framework {
	case FrameworkSB253:
		return []RiskDefinition{
			{
				CheckID:               "missing_scope1_activity_data",
				Title:                 "Missing Scope 1 activity support",
				Severity:              SeverityBlocker,
				Priority:              PriorityHigh,
				Description:           "Direct-emissions source data is missing or incomplete for the current reporting year.",
				AcceptanceCriteria:    []string{"Identify the affected fuel or direct-emission source.", "Explain the corrective action taken.", "Attach the source record or equivalent evidence."},
				RequiredEvidenceTypes: []string{"pdf", "csv", "image"},
				IsTriggered:           func(facts RiskFacts) bool { return !facts.Scope1Ready },
				Evaluator:             evaluateScope1Action,
			},
			{
				CheckID:               "missing_scope2_market_based_data",
				Title:                 "Missing Scope 2 market-based data",
				Severity:              SeverityWarning,
				Priority:              PriorityHigh,
				Description:           "Electricity activity exists, but market-based instruments are not documented for the reporting year.",
				AcceptanceCriteria:    []string{"Describe the REC, contract, tariff, or residual-mix approach.", "State which sites or accounts are covered.", "Attach the supporting document."},
				RequiredEvidenceTypes: []string{"pdf", "csv", "image"},
				IsTriggered:           func(facts RiskFacts) bool { return facts.HasElectricityActivities && !facts.HasMarketBasedSignals },
				Evaluator:             evaluateScope2MarketBased,
			},
			{
				CheckID:               "missing_scope3_supplier_data",
				Title:                 "Missing Scope 3 supplier evidence",
				Severity:              SeverityBlocker,
				Priority:              PriorityHigh,
				Description:           "Supplier or procurement data does not yet support a defendable Scope 3 position.",
				AcceptanceCriteria:    []string{"Identify the supplier or spend category resolved.", "Describe the primary or supplier-specific data collected.", "Attach the supporting supplier export, questionnaire, or schedule."},
				RequiredEvidenceTypes: []string{"pdf", "csv"},
				IsTriggered: func(facts RiskFacts) bool {
					return !facts.Scope3Ready || !facts.HasPurchaseActivities || !facts.HasSupplierMetadata
				},
				Evaluator: evaluateScope3SupplierAction,
			},
			{
				CheckID:               "missing_factor_snapshot",
				Title:                 "Missing factor snapshot",
				Severity:              SeverityWarning,
				Priority:              PriorityMedium,
				Description:           "No locked factor version is present for the reporting year, which weakens reproducibility.",
				AcceptanceCriteria:    []string{"Record the factor version used for the reporting period.", "Lock the factor set for the year.", "Attach the export or screenshot that shows the frozen version."},
				RequiredEvidenceTypes: []string{"pdf", "image"},
				IsTriggered:           func(facts RiskFacts) bool { return !facts.HasLockedFactorSnapshot },
				Evaluator:             evaluateFactorSnapshotAction,
			},
			{
				CheckID:               "missing_approval_workflow",
				Title:                 "Missing approval workflow evidence",
				Severity:              SeverityWarning,
				Priority:              PriorityMedium,
				Description:           "No review or approval trail exists for the disclosure package.",
				AcceptanceCriteria:    []string{"Identify the reviewer or approver.", "Describe the workflow step completed.", "Attach the signoff export, ticket, or supporting record."},
				RequiredEvidenceTypes: []string{"pdf", "image"},
				IsTriggered:           func(facts RiskFacts) bool { return !facts.HasApprovalWorkflow },
				Evaluator:             evaluateApprovalWorkflowAction,
			},
			{
				CheckID:               "low_measured_data_ratio",
				Title:                 "Low measured-data coverage",
				Severity:              SeverityWarning,
				Priority:              PriorityLow,
				Description:           "A high share of the current inventory still relies on estimated or default inputs.",
				AcceptanceCriteria:    []string{"Describe which measured records were added.", "State the coverage improvement achieved.", "Attach the invoice, meter, or source export supporting the update."},
				RequiredEvidenceTypes: []string{"pdf", "csv"},
				IsTriggered:           func(facts RiskFacts) bool { return facts.MeasuredDataRatio > 0 && facts.MeasuredDataRatio < 0.6 },
				Evaluator:             evaluateMeasuredDataAction,
			},
		}, nil
	case FrameworkCSRD, FrameworkIFRS:
		return []RiskDefinition{
			{
				CheckID:               "missing_scope1_activity_data",
				Title:                 "Missing Scope 1 activity support",
				Severity:              SeverityBlocker,
				Priority:              PriorityHigh,
				Description:           "Direct-emissions source data is incomplete for the reporting year.",
				AcceptanceCriteria:    []string{"Identify the affected Scope 1 source.", "Explain the corrective action taken.", "Attach source evidence."},
				RequiredEvidenceTypes: []string{"pdf", "csv", "image"},
				IsTriggered:           func(facts RiskFacts) bool { return !facts.Scope1Ready },
				Evaluator:             evaluateScope1Action,
			},
			{
				CheckID:               "missing_scope2_market_based_data",
				Title:                 "Missing Scope 2 market-based support",
				Severity:              SeverityWarning,
				Priority:              PriorityHigh,
				Description:           "Market-based electricity support is not documented for the current year.",
				AcceptanceCriteria:    []string{"Identify the market-based instrument.", "State covered sites or meters.", "Attach the supporting document."},
				RequiredEvidenceTypes: []string{"pdf", "image"},
				IsTriggered:           func(facts RiskFacts) bool { return facts.HasElectricityActivities && !facts.HasMarketBasedSignals },
				Evaluator:             evaluateScope2MarketBased,
			},
			{
				CheckID:               "missing_scope3_supplier_data",
				Title:                 "Missing Scope 3 supplier evidence",
				Severity:              SeverityBlocker,
				Priority:              PriorityHigh,
				Description:           "Supplier data and value-chain evidence are not yet sufficient for the current inventory.",
				AcceptanceCriteria:    []string{"Identify the supplier or spend category resolved.", "Describe the data source collected.", "Attach the supporting document or export."},
				RequiredEvidenceTypes: []string{"pdf", "csv"},
				IsTriggered:           func(facts RiskFacts) bool { return !facts.Scope3Ready || facts.SupplierSpecificRatio < 0.5 },
				Evaluator:             evaluateScope3SupplierAction,
			},
			{
				CheckID:               "missing_factor_snapshot",
				Title:                 "Missing factor snapshot",
				Severity:              SeverityWarning,
				Priority:              PriorityMedium,
				Description:           "No locked factor version exists for the reporting year.",
				AcceptanceCriteria:    []string{"Record the factor version used.", "Lock the reporting-period snapshot.", "Attach the evidence trace."},
				RequiredEvidenceTypes: []string{"pdf", "image"},
				IsTriggered:           func(facts RiskFacts) bool { return !facts.HasLockedFactorSnapshot },
				Evaluator:             evaluateFactorSnapshotAction,
			},
			{
				CheckID:               "missing_approval_workflow",
				Title:                 "Missing approval workflow evidence",
				Severity:              SeverityWarning,
				Priority:              PriorityMedium,
				Description:           "No review or signoff record is present for the disclosure package.",
				AcceptanceCriteria:    []string{"Identify the reviewer or approver.", "Describe the step completed.", "Attach the signoff record."},
				RequiredEvidenceTypes: []string{"pdf", "image"},
				IsTriggered:           func(facts RiskFacts) bool { return !facts.HasApprovalWorkflow },
				Evaluator:             evaluateApprovalWorkflowAction,
			},
			{
				CheckID:               "low_measured_data_ratio",
				Title:                 "Low measured-data coverage",
				Severity:              SeverityWarning,
				Priority:              PriorityMedium,
				Description:           "Measured data coverage remains low relative to total reported activity.",
				AcceptanceCriteria:    []string{"Describe the measured records added.", "State the coverage improvement.", "Attach the supporting source file."},
				RequiredEvidenceTypes: []string{"pdf", "csv"},
				IsTriggered:           func(facts RiskFacts) bool { return facts.MeasuredDataRatio > 0 && facts.MeasuredDataRatio < 0.7 },
				Evaluator:             evaluateMeasuredDataAction,
			},
		}, nil
	case FrameworkSEC:
		return []RiskDefinition{
			{
				CheckID:               "missing_scope1_activity_data",
				Title:                 "Missing Scope 1 activity support",
				Severity:              SeverityBlocker,
				Priority:              PriorityHigh,
				Description:           "Direct-emissions source data is incomplete for the reporting year.",
				AcceptanceCriteria:    []string{"Identify the source.", "Describe the fix.", "Attach source evidence."},
				RequiredEvidenceTypes: []string{"pdf", "csv", "image"},
				IsTriggered:           func(facts RiskFacts) bool { return !facts.Scope1Ready },
				Evaluator:             evaluateScope1Action,
			},
			{
				CheckID:               "missing_scope2_activity_data",
				Title:                 "Missing Scope 2 activity support",
				Severity:              SeverityBlocker,
				Priority:              PriorityHigh,
				Description:           "Purchased-energy records are incomplete for the reporting year.",
				AcceptanceCriteria:    []string{"Identify the energy record or invoice.", "Describe the corrective action.", "Attach the supporting bill or export."},
				RequiredEvidenceTypes: []string{"pdf", "csv", "image"},
				IsTriggered:           func(facts RiskFacts) bool { return !facts.Scope2Ready },
				Evaluator:             evaluateScope2ActivityAction,
			},
			{
				CheckID:               "missing_factor_snapshot",
				Title:                 "Missing factor snapshot",
				Severity:              SeverityWarning,
				Priority:              PriorityMedium,
				Description:           "No locked factor version exists for the reporting year.",
				AcceptanceCriteria:    []string{"Record the factor version used.", "Lock the reporting-period snapshot.", "Attach the evidence trace."},
				RequiredEvidenceTypes: []string{"pdf", "image"},
				IsTriggered:           func(facts RiskFacts) bool { return !facts.HasLockedFactorSnapshot },
				Evaluator:             evaluateFactorSnapshotAction,
			},
			{
				CheckID:               "missing_approval_workflow",
				Title:                 "Missing approval workflow evidence",
				Severity:              SeverityWarning,
				Priority:              PriorityMedium,
				Description:           "No review or signoff trail exists for the disclosure package.",
				AcceptanceCriteria:    []string{"Identify the reviewer or approver.", "Describe the completed workflow step.", "Attach the signoff record."},
				RequiredEvidenceTypes: []string{"pdf", "image"},
				IsTriggered:           func(facts RiskFacts) bool { return !facts.HasApprovalWorkflow },
				Evaluator:             evaluateApprovalWorkflowAction,
			},
			{
				CheckID:               "low_measured_data_ratio",
				Title:                 "Low measured-data coverage",
				Severity:              SeverityWarning,
				Priority:              PriorityLow,
				Description:           "Measured-source coverage remains low relative to total inventory activity.",
				AcceptanceCriteria:    []string{"Describe the measured records added.", "State the coverage improvement.", "Attach the supporting records."},
				RequiredEvidenceTypes: []string{"pdf", "csv"},
				IsTriggered:           func(facts RiskFacts) bool { return facts.MeasuredDataRatio > 0 && facts.MeasuredDataRatio < 0.65 },
				Evaluator:             evaluateMeasuredDataAction,
			},
		}, nil
	case FrameworkCBAM:
		return []RiskDefinition{
			{
				CheckID:               "missing_imported_goods_evidence",
				Title:                 "Missing imported-goods evidence",
				Severity:              SeverityBlocker,
				Priority:              PriorityHigh,
				Description:           "Imported-goods traceability is incomplete for the current reporting period.",
				AcceptanceCriteria:    []string{"Identify the imported good and classification.", "Describe the customs or supplier record provided.", "Attach the supporting document."},
				RequiredEvidenceTypes: []string{"pdf", "csv", "image"},
				IsTriggered:           func(facts RiskFacts) bool { return facts.HasPurchaseActivities && facts.ImportedGoodsRatio < 0.5 },
				Evaluator:             evaluateImportedGoodsAction,
			},
			{
				CheckID:               "missing_supplier_specific_cbam_data",
				Title:                 "Missing supplier-specific CBAM data",
				Severity:              SeverityBlocker,
				Priority:              PriorityHigh,
				Description:           "Supplier-specific emissions support is incomplete for imported goods in scope.",
				AcceptanceCriteria:    []string{"Identify the supplier-specific emissions source.", "Describe the calculation boundary or verification basis.", "Attach the supporting file."},
				RequiredEvidenceTypes: []string{"pdf", "csv"},
				IsTriggered:           func(facts RiskFacts) bool { return facts.HasPurchaseActivities && facts.SupplierSpecificRatio < 0.5 },
				Evaluator:             evaluateSupplierSpecificCBAMAction,
			},
			{
				CheckID:               "missing_factor_snapshot",
				Title:                 "Missing factor snapshot",
				Severity:              SeverityWarning,
				Priority:              PriorityMedium,
				Description:           "No locked factor version exists for the reporting period.",
				AcceptanceCriteria:    []string{"Record the factor version used.", "Lock the reporting-period snapshot.", "Attach the evidence trace."},
				RequiredEvidenceTypes: []string{"pdf", "image"},
				IsTriggered:           func(facts RiskFacts) bool { return !facts.HasLockedFactorSnapshot },
				Evaluator:             evaluateFactorSnapshotAction,
			},
			{
				CheckID:               "missing_approval_workflow",
				Title:                 "Missing approval workflow evidence",
				Severity:              SeverityWarning,
				Priority:              PriorityMedium,
				Description:           "No review or signoff trail exists for the CBAM submission package.",
				AcceptanceCriteria:    []string{"Identify the reviewer or approver.", "Describe the completed workflow step.", "Attach the signoff record."},
				RequiredEvidenceTypes: []string{"pdf", "image"},
				IsTriggered:           func(facts RiskFacts) bool { return !facts.HasApprovalWorkflow },
				Evaluator:             evaluateApprovalWorkflowAction,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported framework %q", framework)
	}
}
