'use client';

import React, { useEffect, useState, useCallback, useMemo, Suspense, memo } from 'react';
import dynamic from 'next/dynamic';
import { useRouter } from 'next/navigation';
import { useCarbonStore } from '@/stores/carbonStore';
import { useCompliance } from '@/hooks/useCompliance';
import RealTimeProvider, { useRealTime } from '@/providers/RealTimeDataProvider';
import { LoadingSkeleton, DashboardSkeleton } from '@/components/ui/LoadingSkeleton';
import { CarbonMetrics } from '@/components/metrics/CarbonMetrics';
import { ComplianceCalendar } from '@/components/compliance/ComplianceCalendar';
import { CarbonApi, downloadFile, formatNumber } from '@/lib/api/carbon';
import { EmissionData, DataSource, Timeframe, ReductionTarget } from '@/types/carbon';
import ErrorBoundary from '@/components/ErrorBoundary';

// Dynamic imports for code splitting
const DataGlobe = dynamic(() => import('@/components/visualizations/DataGlobe'), {
  loading: () => <LoadingSkeleton type="globe" />,
  ssr: false,
});

const EmissionChart = dynamic(() => import('@/components/charts/EmissionChartJS'), {
  loading: () => <LoadingSkeleton type="chart" />,
});

const AdvancedChart = dynamic(() => import('@/components/charts/AdvancedChart'), {
  loading: () => <LoadingSkeleton type="chart" />,
});

// ============================================================================
// Props Interface
// ============================================================================

interface CarbonDashboardProps {
  tenantId?: string;
  timeframe?: Timeframe;
  onDataChange?: (data: EmissionData) => void;
}

// ============================================================================
// Dashboard Header Component
// ============================================================================

const DashboardHeader = memo(function DashboardHeader({
  title,
  subtitle,
  onExport,
  onSettings,
}: {
  title: string;
  subtitle?: string;
  onExport: () => void;
  onSettings: () => void;
}) {
  const { lastUpdated, isLoading } = useCarbonStore();

  return (
    <div className="flex items-center justify-between mb-6">
      <div>
        <h1 className="text-2xl font-bold text-white">{title}</h1>
        <p className="text-sm text-gray-400 mt-1">
          {subtitle || (lastUpdated ? (
            <>Last updated: {new Date(lastUpdated).toLocaleString()}</>
          ) : (
            'Loading data...'
          ))}
          {isLoading && (
            <span className="ml-2 inline-flex items-center">
              <span className="w-2 h-2 bg-green-500 rounded-full animate-pulse" />
              <span className="ml-1 text-green-400">Syncing</span>
            </span>
          )}
        </p>
      </div>
      <div className="flex items-center gap-3">
        <ExportButton onExport={onExport} />
        <button
          onClick={onSettings}
          className="px-4 py-2 text-sm font-medium text-gray-300 bg-gray-700/50 hover:bg-gray-700 rounded-lg border border-gray-600/50 transition-colors"
        >
          <span className="flex items-center gap-2">
            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
            </svg>
            Settings
          </span>
        </button>
      </div>
    </div>
  );
});

// ============================================================================
// Export Button Component
// ============================================================================

const ExportButton = memo(function ExportButton({
  onExport,
}: {
  onExport: (format: 'pdf' | 'csv' | 'excel') => void;
}) {
  const [isOpen, setIsOpen] = useState(false);

  return (
    <div className="relative">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="px-4 py-2 text-sm font-medium text-gray-300 bg-gray-700/50 hover:bg-gray-700 rounded-lg border border-gray-600/50 transition-colors"
      >
        <span className="flex items-center gap-2">
          <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
          </svg>
          Export
        </span>
      </button>
      {isOpen && (
        <div className="absolute right-0 mt-2 w-40 bg-gray-800 border border-gray-700 rounded-lg shadow-xl z-50">
          {(['pdf', 'csv', 'excel'] as const).map((format) => (
            <button
              key={format}
              onClick={() => {
                onExport(format);
                setIsOpen(false);
              }}
              className="w-full px-4 py-2 text-sm text-left text-gray-300 hover:bg-gray-700 first:rounded-t-lg last:rounded-b-lg"
            >
              Export as {format.toUpperCase()}
            </button>
          ))}
        </div>
      )}
    </div>
  );
});

// ============================================================================
// Scope Breakdown Component
// ============================================================================

const ScopeBreakdown = memo(function ScopeBreakdown({ 
  scopes 
}: { 
  scopes?: { scope1: number; scope2: number; scope3: number } 
}) {
  const { emissions } = useCarbonStore();
  const data = scopes || emissions;
  
  if (!data) return null;
  
  const total = data.scope1 + data.scope2 + data.scope3;
  
  const scopeData = [
    {
      name: 'Scope 1',
      description: 'Direct emissions',
      value: data.scope1,
      percentage: total > 0 ? (data.scope1 / total) * 100 : 0,
      color: 'bg-red-500',
    },
    {
      name: 'Scope 2',
      description: 'Indirect energy',
      value: data.scope2,
      percentage: total > 0 ? (data.scope2 / total) * 100 : 0,
      color: 'bg-yellow-500',
    },
    {
      name: 'Scope 3',
      description: 'Value chain',
      value: data.scope3,
      percentage: total > 0 ? (data.scope3 / total) * 100 : 0,
      color: 'bg-blue-500',
    },
  ];

  return (
    <div className="bg-gray-800/50 rounded-xl border border-gray-700/50 p-6">
      <h3 className="text-sm font-semibold text-gray-400 uppercase tracking-wider mb-4">
        Emission Scopes
      </h3>
      
      {/* Stacked bar */}
      <div className="h-4 rounded-full overflow-hidden bg-gray-700 mb-6 flex">
        {scopeData.map((scope) => (
          <div
            key={scope.name}
            className={`${scope.color} transition-all duration-500`}
            style={{ width: `${scope.percentage}%` }}
          />
        ))}
      </div>

      {/* Scope details */}
      <div className="space-y-4">
        {scopeData.map((scope) => (
          <div key={scope.name} className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <span className={`w-3 h-3 rounded-full ${scope.color}`} />
              <div>
                <span className="text-sm font-medium text-white">{scope.name}</span>
                <span className="text-xs text-gray-500 ml-2">{scope.description}</span>
              </div>
            </div>
            <div className="text-right">
              <span className="text-sm font-semibold text-white">
                {formatNumber(scope.value)}
              </span>
              <span className="text-xs text-gray-500 ml-1">tCO2e</span>
              <span className="text-xs text-gray-400 ml-2">
                ({scope.percentage.toFixed(1)}%)
              </span>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
});

// ============================================================================
// Carbon Intensity Card Component
// ============================================================================

const CarbonIntensityCard = memo(function CarbonIntensityCard({ 
  value 
}: { 
  value?: number 
}) {
  const { metrics } = useCarbonStore();
  const carbonIntensity = value ?? metrics?.carbonIntensity ?? 0;
  
  const intensityMetrics = [
    {
      label: 'Per Revenue',
      value: carbonIntensity.toFixed(2),
      unit: 'tCO2e/$M',
      change: -8.3,
    },
    {
      label: 'Per Employee',
      value: ((metrics?.totalEmissions || 0) / (metrics?.employees || 1)).toFixed(2),
      unit: 'tCO2e',
      change: -5.2,
    },
    {
      label: 'Per Facility',
      value: ((metrics?.totalEmissions || 0) / (metrics?.facilities || 1)).toFixed(2),
      unit: 'tCO2e',
      change: -12.1,
    },
  ];

  return (
    <div className="bg-gray-800/50 rounded-xl border border-gray-700/50 p-6">
      <h3 className="text-sm font-semibold text-gray-400 uppercase tracking-wider mb-4">
        Carbon Intensity
      </h3>
      
      <div className="space-y-4">
        {intensityMetrics.map((metric) => (
          <div key={metric.label} className="flex items-center justify-between">
            <span className="text-sm text-gray-400">{metric.label}</span>
            <div className="flex items-center gap-2">
              <span className="text-lg font-semibold text-white">{metric.value}</span>
              <span className="text-xs text-gray-500">{metric.unit}</span>
              <span className={`text-xs px-1.5 py-0.5 rounded ${
                metric.change < 0 ? 'bg-green-500/20 text-green-400' : 'bg-red-500/20 text-red-400'
              }`}>
                {metric.change > 0 ? '+' : ''}{metric.change}%
              </span>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
});

// Reduction Targets Component
const ReductionTargets = memo(function ReductionTargets({ targets }: { targets?: ReductionTarget[] }) {
  const defaultTargets = [
    { year: 2025, target: 15, current: 12.3, color: 'green' },
    { year: 2030, target: 50, current: 12.3, color: 'blue' },
    { year: 2050, target: 100, current: 12.3, color: 'purple' },
  ] as const;

  const statusColorMap: Record<ReductionTarget['status'], 'green' | 'blue' | 'orange' | 'red' | 'purple'> = {
    achieved: 'green',
    on_track: 'blue',
    at_risk: 'orange',
    behind: 'red',
  };

  const displayTargets = targets && targets.length > 0
    ? targets.map((target) => {
        const parsed = Date.parse(target.deadline);
        const year = Number.isNaN(parsed) ? target.deadline : new Date(parsed).getFullYear();
        return {
          year,
          target: target.targetValue,
          current: target.currentValue,
          color: statusColorMap[target.status] ?? 'purple',
        };
      })
    : defaultTargets;

  return (
    <div className="bg-gray-800/50 rounded-xl border border-gray-700/50 p-6">
      <h3 className="text-sm font-semibold text-gray-400 uppercase tracking-wider mb-4">
        Reduction Targets
      </h3>
      
      <div className="space-y-4">
        {displayTargets.map((target) => (
          <div key={target.year}>
            <div className="flex items-center justify-between mb-2">
              <span className="text-sm text-white font-medium">{target.year}</span>
              <span className="text-xs text-gray-400">
                {target.current}% / {target.target}%
              </span>
            </div>
            <div className="h-2 bg-gray-700 rounded-full overflow-hidden">
              <div
                className={`h-full rounded-full transition-all duration-500 ${
                  target.color === 'green' ? 'bg-green-500' :
                  target.color === 'blue' ? 'bg-blue-500' :
                  'bg-purple-500'
                }`}
                style={{ width: `${(target.current / target.target) * 100}%` }}
              />
            </div>
          </div>
        ))}
      </div>
    </div>
  );
});

// ============================================================================
// Main Dashboard Content
// ============================================================================

const DashboardContent = memo(function DashboardContent({
  tenantId = 'default',
  timeframe = 'monthly',
  onDataChange,
}: CarbonDashboardProps) {
  const router = useRouter();
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);
  
  // Zustand store for global state
  const { 
    emissions, 
    complianceStatus, 
    metrics,
    reductionTargets,
    dataSources,
    fetchEmissions,
    updateMetrics 
  } = useCarbonStore((state) => ({
    emissions: state.emissions,
    complianceStatus: state.complianceStatus,
    metrics: state.metrics,
    reductionTargets: state.reductionTargets,
    dataSources: state.dataSources,
    fetchEmissions: state.fetchEmissions,
    updateMetrics: state.updateMetrics
  }));

  // Custom hooks for business logic
  const { deadlines, checkCompliance } = useCompliance(tenantId);
  const { subscribe, isConnected } = useRealTime();
  
  // State for modals
  const [selectedDataSource, setSelectedDataSource] = useState<DataSource | null>(null);
  
  // Memoized calculations
  const carbonIntensity = useMemo(() => {
    if (!emissions || !metrics.revenue) return 0;
    return (emissions.total / metrics.revenue) * 1000000;
  }, [emissions, metrics.revenue]);

  // Async data fetching with error handling
  useEffect(() => {
    const loadDashboardData = async () => {
      try {
        setIsLoading(true);
        await Promise.all([
          fetchEmissions(tenantId, timeframe),
          checkCompliance()
        ]);
      } catch (err) {
        setError(err instanceof Error ? err : new Error('Unknown error'));
        console.error('Failed to load dashboard:', err);
      } finally {
        setIsLoading(false);
      }
    };

    loadDashboardData();
  }, [tenantId, timeframe, fetchEmissions, checkCompliance]);

  // Subscribe to real-time updates
  useEffect(() => {
    const handleUpdate = (update: Partial<EmissionData>) => {
      updateMetrics(update as Partial<typeof metrics>);
      if (emissions && onDataChange) {
        onDataChange(emissions);
      }
    };

    const unsubscribe = subscribe(tenantId, handleUpdate);

    return unsubscribe;
  }, [subscribe, tenantId, updateMetrics, emissions, onDataChange]);

  // Event handlers with useCallback
  const handleMetricClick = useCallback((metricId: string, value?: number) => {
    // Analytics tracking
    if (typeof window !== 'undefined' && (window as any).gtag) {
      (window as any).gtag('event', 'metric_click', {
        metric_id: metricId,
        value: value,
        tenant_id: tenantId
      });
    }
    
    // Open detailed view
    router.push(`/dashboard/metrics/${metricId}`);
  }, [tenantId, router]);

  const handleExport = useCallback(async (format: 'pdf' | 'csv' | 'excel') => {
    try {
      const api = CarbonApi.getInstance();
      const report = await api.generateComplianceReport(tenantId, format, [1, 2, 3]);
      
      if (report.url) {
        window.open(report.url, '_blank');
      } else {
        // Fallback: export current data as JSON
        const data = JSON.stringify({ emissions, metrics, complianceStatus }, null, 2);
        downloadFile(data, `offgridflow-report-${Date.now()}.json`);
      }
    } catch (err) {
      console.error('Failed to export report:', err);
      // Fallback export
      const data = JSON.stringify({ emissions, metrics, complianceStatus }, null, 2);
      downloadFile(data, `offgridflow-report-${Date.now()}.json`);
    }
  }, [tenantId, emissions, metrics, complianceStatus]);

  const handleSettings = useCallback(() => {
    router.push('/dashboard/settings');
  }, [router]);

  const handleNodeClick = useCallback((node: DataSource) => {
    setSelectedDataSource(node);
  }, []);

  if (isLoading) {
    return <DashboardSkeleton />;
  }

  if (error) {
    return (
      <div className="min-h-screen bg-gray-900 p-6">
        <div className="max-w-2xl mx-auto mt-20 bg-red-900/20 border border-red-700/50 rounded-xl p-8 text-center">
          <h2 className="text-xl font-bold text-red-400 mb-2">Error Loading Dashboard</h2>
          <p className="text-red-300/70 mb-4">{error.message}</p>
          <button
            onClick={() => fetchEmissions(tenantId, timeframe)}
            className="px-6 py-2 bg-red-600 hover:bg-red-500 text-white rounded-lg transition-colors"
          >
            Retry
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="carbon-dashboard min-h-screen bg-gray-900">
      <div className="max-w-7xl mx-auto p-6">
        {/* Header */}
        <DashboardHeader 
          title="Carbon Intelligence Dashboard"
          subtitle={emissions?.updatedAt ? `Last updated: ${new Date(emissions.updatedAt).toLocaleString()}` : undefined}
          onExport={() => handleExport('pdf')}
          onSettings={handleSettings}
        />

        {/* Connection Status */}
        <div className="mb-6 flex items-center gap-2">
          <span className={`w-2 h-2 rounded-full ${isConnected ? 'bg-green-500' : 'bg-yellow-500'}`} />
          <span className="text-xs text-gray-400">
            {isConnected ? 'Real-time data connected' : 'Using cached data'}
          </span>
        </div>
        
        <div className="dashboard-grid grid grid-cols-1 lg:grid-cols-12 gap-6">
          {/* Left Column - Metrics */}
          <div className="metrics-section lg:col-span-3 space-y-6">
            <CarbonMetrics 
              metrics={metrics}
              reductionTargets={reductionTargets}
              onMetricClick={handleMetricClick}
            />
            
            <ComplianceCalendar 
              deadlines={deadlines}
              complianceStatus={complianceStatus}
            />
          </div>
          
          {/* Center Column - Visualizations */}
          <div className="visualization-section lg:col-span-6 space-y-6">
            <Suspense fallback={<LoadingSkeleton type="chart" />}>
              <EmissionChart 
                data={emissions ? [emissions] : []}
                timeframe={timeframe}
                height={400}
              />
            </Suspense>
            
            <Suspense fallback={<LoadingSkeleton type="globe" />}>
              <DataGlobe 
                nodes={dataSources}
                onNodeClick={handleNodeClick}
              />
            </Suspense>
          </div>
          
          {/* Right Column - Insights */}
          <div className="insights-section lg:col-span-3 space-y-6">
            <CarbonIntensityCard value={carbonIntensity} />
            {emissions && (
              <ScopeBreakdown scopes={{ 
                scope1: emissions.scope1, 
                scope2: emissions.scope2, 
                scope3: emissions.scope3 
              }} />
            )}
            <ReductionTargets targets={metrics.reductionTarget ? [
              {
                id: '1',
                name: '2025 Net Zero Goal',
                targetValue: metrics.reductionTarget,
                currentValue: emissions?.total || 0,
                deadline: '2025-12-31',
                status: 'at_risk' as const,
              }
            ] : undefined} />
          </div>
        </div>

        {/* Selected Data Source Modal */}
        {selectedDataSource && (
          <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
            <div className="bg-gray-800 border border-gray-700 rounded-xl p-6 max-w-md w-full mx-4">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-semibold text-white">
                  {selectedDataSource.name}
                </h3>
                <button
                  onClick={() => setSelectedDataSource(null)}
                  className="text-gray-400 hover:text-white"
                >
                  <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
              <div className="space-y-3">
                <div className="flex justify-between">
                  <span className="text-gray-400">Type</span>
                  <span className="text-white">{selectedDataSource.type}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-400">Status</span>
                  <span className={`px-2 py-0.5 rounded text-sm ${
                    selectedDataSource.status === 'active' ? 'bg-green-500/20 text-green-400' :
                    selectedDataSource.status === 'error' ? 'bg-red-500/20 text-red-400' :
                    'bg-yellow-500/20 text-yellow-400'
                  }`}>
                    {selectedDataSource.status}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-400">Last Sync</span>
                  <span className="text-white">
                    {new Date(selectedDataSource.lastSync).toLocaleString()}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-400">Emissions</span>
                  <span className="text-white">
                    {formatNumber(selectedDataSource.emissions)} tCO₂e
                  </span>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
});

DashboardContent.displayName = 'DashboardContent';

// ============================================================================
// Main Dashboard Component with Provider
// ============================================================================

export const CarbonDashboard: React.FC<CarbonDashboardProps> = memo(({
  tenantId = 'default',
  timeframe = 'monthly',
  onDataChange,
}) => {
  const wsUrl = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8090/ws';

  return (
    <RealTimeProvider tenantId={tenantId} baseUrl={wsUrl}>
      <ErrorBoundary componentName="CarbonDashboard">
        <DashboardContent 
          tenantId={tenantId}
          timeframe={timeframe}
          onDataChange={onDataChange}
        />
      </ErrorBoundary>
    </RealTimeProvider>
  );
});

CarbonDashboard.displayName = 'CarbonDashboard';

export default CarbonDashboard;
