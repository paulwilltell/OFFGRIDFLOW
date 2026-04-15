package abatement

import "time"

type Framework string

const (
	FrameworkSB253 Framework = "sb253"
	FrameworkCSRD  Framework = "csrd"
	FrameworkSEC   Framework = "sec"
	FrameworkIFRS  Framework = "ifrs"
	FrameworkCBAM  Framework = "cbam"
)

type RiskSeverity string

const (
	SeverityBlocker RiskSeverity = "blocker"
	SeverityWarning RiskSeverity = "warning"
)

type RiskPriority string

const (
	PriorityHigh   RiskPriority = "high"
	PriorityMedium RiskPriority = "medium"
	PriorityLow    RiskPriority = "low"
)

type EngineStatus string

const (
	StatusRecommended       EngineStatus = "recommended"
	StatusNeedsClarification EngineStatus = "needs_clarification"
	StatusInsufficient      EngineStatus = "insufficient"
)

type FrameworkMeta struct {
	Key            Framework `json:"key"`
	Label          string    `json:"label"`
	ShortLabel     string    `json:"short_label"`
	PenaltyHeading string    `json:"penalty_heading"`
	PenaltyBody    string    `json:"penalty_body"`
}

type Evaluation struct {
	Status          EngineStatus `json:"status"`
	Feedback        string       `json:"feedback"`
	CriteriaChecked []string     `json:"criteriaChecked"`
}

type EvidenceUpload struct {
	FileName string
	MimeType string
	Content  []byte
}

type EvidenceRecord struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	FileName  string    `json:"fileName"`
	MimeType  string    `json:"mimeType"`
	SizeBytes int64     `json:"sizeBytes"`
	CreatedAt time.Time `json:"createdAt"`
}

type RiskCard struct {
	ID                    string        `json:"id"`
	ComplianceCheckID     string        `json:"complianceCheckId"`
	Title                 string        `json:"title"`
	Severity              RiskSeverity  `json:"severity"`
	Priority              RiskPriority  `json:"priority"`
	Description           string        `json:"description"`
	AcceptanceCriteria    []string      `json:"acceptanceCriteria"`
	RequiredEvidenceTypes []string      `json:"requiredEvidenceTypes"`
	Justification         string        `json:"justification"`
	Evidence              []EvidenceRecord `json:"evidence"`
	EngineStatus          EngineStatus  `json:"engineStatus,omitempty"`
	EngineFeedback        string        `json:"engineFeedback,omitempty"`
	CriteriaChecked       []string      `json:"criteriaChecked"`
	Completed             bool          `json:"completed"`
	SelfCertified         bool          `json:"selfCertified"`
	CertifiedAt           *time.Time    `json:"certifiedAt,omitempty"`
	UpdatedAt             *time.Time    `json:"updatedAt,omitempty"`
}

type RiskSummary struct {
	High   int `json:"high"`
	Medium int `json:"medium"`
	Low    int `json:"low"`
}

type ProgressSummary struct {
	CriticalAddressed int `json:"criticalAddressed"`
	CriticalTotal     int `json:"criticalTotal"`
	PercentAddressed  int `json:"percentAddressed"`
}

type Dashboard struct {
	Framework     FrameworkMeta    `json:"framework"`
	Summary       RiskSummary      `json:"summary"`
	Progress      ProgressSummary  `json:"progress"`
	Disclaimer    string           `json:"disclaimer"`
	ReportingYear int              `json:"reportingYear"`
	Risks         []RiskCard       `json:"risks"`
	GeneratedAt   string           `json:"generatedAt"`
}

type EvaluateRequest struct {
	ComplianceCheckID string
	Completed         bool
	Justification     string
	Evidence          []EvidenceUpload
}

type SelfCertificationRequest struct {
	ActionItemID    string `json:"action_item_id"`
	SelfCertified   bool   `json:"self_certified"`
	ComplianceCheckID string `json:"compliance_check_id,omitempty"`
}

type StoredActionItem struct {
	ID                    string
	TenantID              string
	Framework             Framework
	ComplianceCheckID     string
	Title                 string
	Severity              RiskSeverity
	Priority              RiskPriority
	Description           string
	AcceptanceCriteria    []string
	RequiredEvidenceTypes []string
	Justification         string
	EvidenceURLs          []string
	EngineStatus          EngineStatus
	EngineFeedback        string
	CriteriaChecked       []string
	Completed             bool
	SelfCertified         bool
	CertifiedAt           *time.Time
	UpdatedAt             time.Time
}

type RiskFacts struct {
	ReportingYear            int
	Scope1Ready              bool
	Scope2Ready              bool
	Scope3Ready              bool
	HasElectricityActivities bool
	HasPurchaseActivities    bool
	HasSupplierMetadata      bool
	HasMarketBasedSignals    bool
	HasLockedFactorSnapshot  bool
	HasApprovalWorkflow      bool
	MeasuredDataRatio        float64
	SupplierSpecificRatio    float64
	ImportedGoodsRatio       float64
}

type RiskDefinition struct {
	CheckID               string
	Title                 string
	Severity              RiskSeverity
	Priority              RiskPriority
	Description           string
	AcceptanceCriteria    []string
	RequiredEvidenceTypes []string
	IsTriggered           func(facts RiskFacts) bool
	Evaluator             func(justification string, evidence []EvidenceUpload) Evaluation
}

const GuidanceDisclaimer = "This tool provides guidance only. OffGridFlow does not file, assure, or guarantee compliance. Final decisions are yours."

