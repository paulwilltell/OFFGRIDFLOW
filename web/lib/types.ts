/**
 * Shared TypeScript types for OffGridFlow API responses
 */

// -----------------------------------------------------------------------------
// Emissions Types
// -----------------------------------------------------------------------------

export interface Scope2Emission {
  id: string;
  meterId: string;
  location: string;
  region: string;
  quantityKWh: number;
  emissionsKgCO2e: number;
  emissionsTonsCO2e: number;
  emissionFactor: number;
  methodology: 'location-based' | 'market-based';
  dataSource: string;
  dataQuality: string;
  periodStart: string;
  periodEnd: string;
  calculatedAt?: string;
}

export interface Scope2Summary {
  scope: 'SCOPE2';
  totalKWh: number;
  totalEmissionsKgCO2e: number;
  totalEmissionsTonsCO2e: number;
  averageEmissionFactor: number;
  activityCount: number;
  regionBreakdown?: Record<string, number>;
  periodStart?: string;
  periodEnd?: string;
  timestamp: string;
}

export interface EmissionsTotals {
  scope1Tons: number;
  scope2Tons: number;
  scope3Tons: number;
  totalTons: number;
}

// -----------------------------------------------------------------------------
// Compliance Types
// -----------------------------------------------------------------------------

export interface ValidationInfo {
  valid: boolean;
  errors: string[];
  warnings: string[];
}

export interface CSRDComplianceResponse {
  standard: string;
  orgId: string;
  year: number;
  totals: EmissionsTotals;
  metrics: Record<string, unknown> & {
    validation?: ValidationInfo;
  };
  status: 'ok' | 'incomplete' | 'warnings';
  timestamp: string;
}

export interface FrameworkStatus {
  name: string;
  status: 'ok' | 'partial' | 'no_data' | 'not_started' | 'not_applicable';
  scope1Ready?: boolean;
  scope2Ready?: boolean;
  scope3Ready?: boolean;
}

export interface ComplianceSummary {
  frameworks: {
    csrd: FrameworkStatus;
    sec: FrameworkStatus;
    cbam: FrameworkStatus;
    california: FrameworkStatus;
  };
  totals: {
    scope1Tons: number;
    scope2Tons: number;
    scope3Tons: number;
  };
  timestamp: string;
}

// -----------------------------------------------------------------------------
// Dashboard Types
// -----------------------------------------------------------------------------

export interface ModeResponse {
  mode: 'normal' | 'offline' | 'degraded';
}

export interface HealthResponse {
  status: 'ok' | 'degraded' | 'unhealthy';
  timestamp: string;
  service: string;
}

export interface ScheduleStatus {
  interval?: string;
  last_run_at?: string;
  next_run_at?: string;
}

// -----------------------------------------------------------------------------
// AI Chat Types
// -----------------------------------------------------------------------------

export interface ChatRequest {
  prompt: string;
  context?: string;
}

export interface ChatResponse {
  output: string;
  source: 'cloud' | 'local';
}

// -----------------------------------------------------------------------------
// Pagination Types
// -----------------------------------------------------------------------------

export interface PageInfo {
  page: number;
  perPage: number;
  total: number;
  totalPages: number;
  hasNext: boolean;
  hasPrev: boolean;
}

export interface PaginatedResponse<T> {
  data: T[];
  pageInfo: PageInfo;
}
