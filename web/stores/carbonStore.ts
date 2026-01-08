import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { immer } from 'zustand/middleware/immer';
import { CarbonApi } from '@/lib/api/carbon';
import { 
  EmissionData, 
  ComplianceStatus, 
  CarbonMetrics,
  DataSource,
  ReductionTarget,
  Timeframe,
  ComplianceStatusType
} from '@/types/carbon';

// ============================================================================
// Re-export types for convenience
// ============================================================================

export type { 
  EmissionData, 
  ComplianceStatus, 
  CarbonMetrics, 
  DataSource, 
  ReductionTarget,
  Timeframe 
} from '@/types/carbon';

// ============================================================================
// Store State Interface
// ============================================================================

interface CarbonState {
  // State
  emissions: EmissionData | null;
  complianceStatus: ComplianceStatus;
  metrics: CarbonMetrics;
  dataSources: DataSource[];
  reductionTargets: ReductionTarget[];
  isLoading: boolean;
  error: string | null;
  lastUpdated: string | null;
  
  // Actions
  fetchEmissions: (tenantId: string, timeframe: Timeframe) => Promise<void>;
  updateMetrics: (update: Partial<CarbonMetrics>) => void;
  updateEmission: (update: Partial<EmissionData>) => void;
  setComplianceStatus: (status: Partial<ComplianceStatus>) => void;
  addDataSource: (source: DataSource) => void;
  removeDataSource: (sourceId: string) => void;
  updateDataSourceStatus: (sourceId: string, status: DataSource['status']) => void;
  calculateIntensity: () => number;
  reset: () => void;
}

// ============================================================================
// Analytics Helper
// ============================================================================

const trackEvent = (name: string, properties: Record<string, unknown>) => {
  if (typeof window !== 'undefined' && (window as unknown as { gtag?: Function }).gtag) {
    (window as unknown as { gtag: Function }).gtag('event', name, properties);
  }
  if (process.env.NODE_ENV === 'development') {
    console.log('[Analytics]', name, properties);
  }
};

const captureException = (error: unknown, context: Record<string, unknown>) => {
  console.error('[Error]', error, context);
  // TODO: Integrate with Sentry or other error tracking
};

// ============================================================================
// Initial State
// ============================================================================

const initialComplianceStatus: ComplianceStatus = {
  csrd: 'pending' as ComplianceStatusType,
  sec: 'complete' as ComplianceStatusType,
  sb253: 'in_progress' as ComplianceStatusType,
  ifrs: 'pending' as ComplianceStatusType,
  cbam: 'pending' as ComplianceStatusType,
};

const initialMetrics: CarbonMetrics = {
  totalEmissions: 0,
  carbonIntensity: 0,
  benchmarkComparison: 0,
  reductionTarget: 0,
  progress: 0,
  revenue: 0,
  employees: 0,
  facilities: 0,
};

// ============================================================================
// Zustand Store with Immer
// ============================================================================

export const useCarbonStore = create<CarbonState>()(
  persist(
    immer((set, get) => ({
      // Initial state
      emissions: null,
      complianceStatus: initialComplianceStatus,
      metrics: initialMetrics,
      dataSources: [],
      reductionTargets: [],
      isLoading: false,
      error: null,
      lastUpdated: null,
      
      // Actions
      fetchEmissions: async (tenantId: string, timeframe: Timeframe) => {
        set({ isLoading: true, error: null });
        
        try {
          const api = CarbonApi.getInstance();
          const emissions = await api.getEmissions(tenantId, timeframe);
          
          set((state) => {
            state.emissions = emissions;
            state.metrics.totalEmissions = emissions.total;
            state.dataSources = emissions.dataSources || [];
            state.lastUpdated = new Date().toISOString();
            
            // Calculate carbon intensity if revenue exists
            if (state.metrics.revenue > 0) {
              state.metrics.carbonIntensity = 
                (emissions.total / state.metrics.revenue) * 1000000;
            }
          });
          
          trackEvent('emissions_fetched', { tenantId, timeframe });
          
        } catch (error) {
          const errorMessage = error instanceof Error ? error.message : 'Failed to fetch emissions';
          set({ error: errorMessage });
          captureException(error, { tenantId, timeframe });
          
          // Use mock data in development
          if (process.env.NODE_ENV === 'development') {
            set((state) => {
              state.emissions = generateMockEmissions(tenantId);
              state.dataSources = generateMockDataSources();
              state.metrics = generateMockMetrics();
              state.lastUpdated = new Date().toISOString();
              state.error = null;
            });
          }
        } finally {
          set({ isLoading: false });
        }
      },
      
      updateMetrics: (update: Partial<CarbonMetrics>) => {
        set((state) => {
          Object.assign(state.metrics, update);
          
          // Recalculate intensity if relevant fields changed
          if (update.revenue !== undefined || update.totalEmissions !== undefined) {
            const { emissions, metrics } = state;
            if (emissions && metrics.revenue > 0) {
              state.metrics.carbonIntensity = 
                (emissions.total / metrics.revenue) * 1000000;
            }
          }
        });
      },

      updateEmission: (update: Partial<EmissionData>) => {
        set((state) => {
          if (state.emissions) {
            Object.assign(state.emissions, update);
            state.lastUpdated = new Date().toISOString();
          }
        });
      },

      setComplianceStatus: (status: Partial<ComplianceStatus>) => {
        set((state) => {
          Object.assign(state.complianceStatus, status);
        });
      },

      addDataSource: (source: DataSource) => {
        set((state) => {
          state.dataSources.push(source);
        });
        trackEvent('data_source_added', { sourceType: source.type, sourceName: source.name });
      },

      removeDataSource: (sourceId: string) => {
        set((state) => {
          state.dataSources = state.dataSources.filter(s => s.id !== sourceId);
        });
        trackEvent('data_source_removed', { sourceId });
      },

      updateDataSourceStatus: (sourceId: string, status: DataSource['status']) => {
        set((state) => {
          const source = state.dataSources.find(s => s.id === sourceId);
          if (source) {
            source.status = status;
            source.lastSync = new Date();
          }
        });
      },
      
      calculateIntensity: () => {
        const { emissions, metrics } = get();
        if (!emissions || !metrics.revenue) return 0;
        return (emissions.total / metrics.revenue) * 1000000;
      },
      
      reset: () => {
        set({
          emissions: null,
          complianceStatus: initialComplianceStatus,
          metrics: initialMetrics,
          dataSources: [],
          reductionTargets: [],
          isLoading: false,
          error: null,
          lastUpdated: null,
        });
      },
    })),
    {
      name: 'carbon-storage',
      partialize: (state) => ({
        metrics: state.metrics,
        complianceStatus: state.complianceStatus,
        reductionTargets: state.reductionTargets,
      }),
    }
  )
);

// ============================================================================
// Mock Data Generators (Development Only)
// ============================================================================

function generateMockEmissions(tenantId: string): EmissionData {
  return {
    id: `emit-${Date.now()}`,
    tenantId,
    total: 12450.5,
    scope1: 3200.0,
    scope2: 5800.5,
    scope3: 3450.0,
    intensity: 249.01,
    timeframe: 'monthly',
    dataSources: generateMockDataSources(),
    updatedAt: new Date(),
    methodology: 'ghg_protocol',
    uncertainty: 5.2,
    region: 'north_america',
  };
}

function generateMockMetrics(): CarbonMetrics {
  return {
    totalEmissions: 12450.5,
    carbonIntensity: 249.01,
    benchmarkComparison: -7.5,
    reductionTarget: 10000,
    progress: 19.7,
    revenue: 50000000,
    employees: 250,
    facilities: 5,
    comparison: {
      industryAverage: 320,
      percentile: 25,
    },
  };
}

function generateMockDataSources(): DataSource[] {
  return [
    {
      id: '1',
      name: 'HQ Building',
      type: 'iot',
      status: 'active',
      lastSync: new Date(),
      emissions: 2500,
      coordinates: { lat: 37.7749, lng: -122.4194 },
    },
    {
      id: '2',
      name: 'AWS Cloud',
      type: 'aws',
      status: 'active',
      lastSync: new Date(),
      emissions: 1800,
      coordinates: { lat: 47.6062, lng: -122.3321 },
    },
    {
      id: '3',
      name: 'Azure Services',
      type: 'azure',
      status: 'active',
      lastSync: new Date(),
      emissions: 1200,
      coordinates: { lat: 52.52, lng: 13.405 },
    },
    {
      id: '4',
      name: 'SAP Integration',
      type: 'sap',
      status: 'active',
      lastSync: new Date(),
      emissions: 3500,
      coordinates: { lat: 49.0069, lng: 8.4037 },
    },
    {
      id: '5',
      name: 'Fleet Vehicles',
      type: 'manual',
      status: 'active',
      lastSync: new Date(),
      emissions: 1950,
      coordinates: { lat: 34.0522, lng: -118.2437 },
    },
    {
      id: '6',
      name: 'Asia Data Center',
      type: 'gcp',
      status: 'inactive',
      lastSync: new Date(Date.now() - 86400000),
      emissions: 1500,
      coordinates: { lat: 35.6762, lng: 139.6503 },
    },
  ];
}

// ============================================================================
// Selector Hooks (Optimized Re-renders)
// ============================================================================

export const useEmissions = () => useCarbonStore((state) => state.emissions);
export const useMetrics = () => useCarbonStore((state) => state.metrics);
export const useComplianceStatus = () => useCarbonStore((state) => state.complianceStatus);
export const useDataSources = () => useCarbonStore((state) => state.dataSources);
export const useReductionTargets = () => useCarbonStore((state) => state.reductionTargets);
export const useCarbonLoading = () => useCarbonStore((state) => state.isLoading);
export const useCarbonError = () => useCarbonStore((state) => state.error);
export const useLastUpdated = () => useCarbonStore((state) => state.lastUpdated);
