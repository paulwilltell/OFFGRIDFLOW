// Carbon Dashboard Component Architecture
// Exports all components for the OffGridFlow Carbon Dashboard

// Main Dashboard
export { CarbonDashboard } from './CarbonDashboard';

// Stores
export {
  useCarbonStore,
  useEmissions,
  useMetrics,
  useComplianceStatus,
  useCarbonLoading,
} from '@/stores/carbonStore';
export type {
  DataSource,
  EmissionData,
  FlatEmissions,
  DashboardMetrics,
  ReductionTarget,
  ComplianceStatus,
} from '@/stores/carbonStore';

// Hooks
export { useCompliance } from '@/hooks/useCompliance';
export type { ComplianceDeadline, ComplianceCheckResult } from '@/hooks/useCompliance';

// Providers
export { RealTimeDataProvider, useRealTimeData } from '@/providers/RealTimeDataProvider';

// UI Components
export { LoadingSkeleton, DashboardSkeleton } from './ui/LoadingSkeleton';

// Metrics Components
export { CarbonMetrics } from './metrics/CarbonMetrics';

// Compliance Components
export { ComplianceCalendar } from './compliance/ComplianceCalendar';

// Chart Components
export { EmissionChart } from './charts/EmissionChart';
export { AdvancedChart } from './charts/AdvancedChart';

// Visualization Components
export { DataGlobe } from './visualizations/DataGlobe';
