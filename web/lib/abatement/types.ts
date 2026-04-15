export type AbatementFramework = 'sb253' | 'csrd' | 'sec' | 'ifrs' | 'cbam';

export type RiskSeverity = 'blocker' | 'warning';
export type RiskPriority = 'high' | 'medium' | 'low';
export type EngineStatus = 'recommended' | 'needs_clarification' | 'insufficient';

export interface FrameworkMeta {
  key: AbatementFramework;
  label: string;
  short_label: string;
  penalty_heading: string;
  penalty_body: string;
}

export interface EvidenceRecord {
  id: string;
  url: string;
  fileName: string;
  mimeType: string;
  sizeBytes: number;
  createdAt: string;
}

export interface RiskCardData {
  id: string;
  complianceCheckId: string;
  title: string;
  severity: RiskSeverity;
  priority: RiskPriority;
  description: string;
  acceptanceCriteria: string[];
  requiredEvidenceTypes: string[];
  justification: string;
  evidence: EvidenceRecord[];
  engineStatus?: EngineStatus;
  engineFeedback?: string;
  criteriaChecked: string[];
  completed: boolean;
  selfCertified: boolean;
  certifiedAt?: string | null;
  updatedAt?: string | null;
}

export interface RiskSummary {
  high: number;
  medium: number;
  low: number;
}

export interface ProgressSummary {
  criticalAddressed: number;
  criticalTotal: number;
  percentAddressed: number;
}

export interface AbatementDashboardData {
  framework: FrameworkMeta;
  summary: RiskSummary;
  progress: ProgressSummary;
  disclaimer: string;
  reportingYear: number;
  risks: RiskCardData[];
  generatedAt: string;
}

export interface AbatementEvaluation {
  status: EngineStatus;
  feedback: string;
  criteriaChecked: string[];
}

export interface EvaluateAbatementResponse {
  risk: RiskCardData;
  evaluation: AbatementEvaluation;
}

export interface SelfCertifyResponse {
  risk: RiskCardData;
}
