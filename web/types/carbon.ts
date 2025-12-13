// ============================================================================
// Carbon Types - Type Safety for OffGridFlow
// ============================================================================

export interface EmissionData {
  id: string;
  tenantId: string;
  total: number;
  scope1: number;
  scope2: number;
  scope3: number;
  intensity: number;
  timeframe: Timeframe;
  dataSources: DataSource[];
  updatedAt: Date;
  methodology: Methodology;
  uncertainty: number;
  region: Region;
}

export interface CarbonMetrics {
  totalEmissions: number;
  carbonIntensity: number;
  reductionTarget: number;
  progress: number;
  revenue: number;
  employees?: number;
  facilities?: number;
  comparison?: {
    industryAverage: number;
    percentile: number;
  };
}

export interface ComplianceStatus {
  csrd: ComplianceStatusType;
  sec: ComplianceStatusType;
  sb253: ComplianceStatusType;
  ifrs: ComplianceStatusType;
  cbam: ComplianceStatusType;
}

export type ComplianceStatusType = 
  | 'complete' 
  | 'in_progress' 
  | 'pending' 
  | 'at_risk' 
  | 'overdue';

export type Timeframe = 'daily' | 'weekly' | 'monthly' | 'quarterly' | 'yearly';

export interface DataSource {
  id: string;
  type: DataSourceType;
  name: string;
  lastSync: Date | string;
  status: DataSourceStatus;
  emissions: number;
  coordinates?: {
    lat: number;
    lng: number;
  };
}

export type DataSourceType = 'aws' | 'azure' | 'gcp' | 'sap' | 'manual' | 'iot' | 'api' | 'import';
export type DataSourceStatus = 'active' | 'inactive' | 'error';

export type Methodology = 
  | 'ghg_protocol' 
  | 'iso_14064' 
  | 'epa_method' 
  | 'defra'
  | 'custom';

export type Region = 
  | 'north_america'
  | 'europe'
  | 'asia_pacific'
  | 'latin_america'
  | 'middle_east'
  | 'africa'
  | 'global';

// ============================================================================
// Utility Types
// ============================================================================

export type RequireAtLeastOne<T, Keys extends keyof T = keyof T> = 
  Pick<T, Exclude<keyof T, Keys>> &
  {
    [K in Keys]-?: Required<Pick<T, K>> & Partial<Pick<T, Exclude<Keys, K>>>
  }[Keys];

export type DeepPartial<T> = {
  [P in keyof T]?: T[P] extends object ? DeepPartial<T[P]> : T[P];
};

// ============================================================================
// API Response Types
// ============================================================================

export interface ApiResponse<T> {
  data: T;
  meta: {
    requestId: string;
    timestamp: Date;
    version: string;
  };
  pagination?: PaginationMeta;
}

export interface PaginationMeta {
  page: number;
  limit: number;
  total: number;
  pages: number;
}

export interface ApiError {
  code: string;
  message: string;
  details?: Record<string, unknown>;
  requestId: string;
}

// ============================================================================
// Chart Types
// ============================================================================

export interface ChartPoint {
  datasetIndex: number;
  index: number;
  value: EmissionData;
}

export interface ChartDataset {
  label: string;
  data: number[];
  borderColor: string;
  backgroundColor?: string;
  borderWidth?: number;
  tension?: number;
  fill?: boolean;
}

// ============================================================================
// Reduction Target Types
// ============================================================================

export interface ReductionTarget {
  id: string;
  name: string;
  targetValue: number;
  currentValue: number;
  deadline: string;
  status: 'on_track' | 'at_risk' | 'behind' | 'achieved';
  baselineYear?: number;
  baselineValue?: number;
}

// ============================================================================
// Compliance Types
// ============================================================================

export interface ComplianceDeadline {
  id: string;
  framework: ComplianceFramework;
  title: string;
  description?: string;
  dueDate: Date | string;
  status: ComplianceStatusType;
  priority: 'low' | 'medium' | 'high' | 'critical';
  assignee?: string;
  requirements?: string[];
}

export type ComplianceFramework = 
  | 'CSRD'
  | 'SEC'
  | 'SB253'
  | 'IFRS'
  | 'CBAM'
  | 'CDP'
  | 'TCFD'
  | 'GRI';

// ============================================================================
// Report Types
// ============================================================================

export interface ComplianceReport {
  id: string;
  tenantId: string;
  format: ReportFormat;
  generatedAt: Date;
  scopes: number[];
  data: EmissionData;
  compliance: ComplianceStatus;
  url?: string;
}

export type ReportFormat = 'pdf' | 'csv' | 'excel' | 'json';

// ============================================================================
// Real-time Update Types
// ============================================================================

export interface RealTimeUpdate {
  type: 'emission' | 'compliance' | 'metric' | 'alert';
  payload: unknown;
  timestamp: Date;
  tenantId: string;
}

export interface EmissionUpdate {
  source: string;
  value: number;
  scope: 1 | 2 | 3;
  timestamp: Date;
}

// ============================================================================
// Analytics Types
// ============================================================================

export interface AnalyticsEvent {
  name: string;
  properties: Record<string, unknown>;
  timestamp: Date;
  userId?: string;
  tenantId?: string;
}

// ============================================================================
// Filter & Query Types
// ============================================================================

export interface EmissionFilter {
  timeframe?: Timeframe;
  scopes?: number[];
  sources?: string[];
  region?: Region;
  startDate?: Date;
  endDate?: Date;
}

export interface QueryOptions {
  page?: number;
  limit?: number;
  sortBy?: string;
  sortOrder?: 'asc' | 'desc';
  filters?: EmissionFilter;
}
