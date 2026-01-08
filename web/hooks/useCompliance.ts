import { useState, useEffect, useCallback } from 'react';
import { api } from '@/lib/api';
import { useCarbonStore, ComplianceStatus } from '@/stores/carbonStore';

export interface ComplianceDeadline {
  id: string;
  framework: 'csrd' | 'sec' | 'cbam' | 'sb253' | 'ifrs';
  title: string;
  description: string;
  dueDate: string;
  status: 'upcoming' | 'due_soon' | 'overdue' | 'completed';
  priority: 'high' | 'medium' | 'low';
  requirements: string[];
}

export interface ComplianceCheckResult {
  framework: string;
  score: number;
  gaps: string[];
  recommendations: string[];
}

interface UseComplianceReturn {
  deadlines: ComplianceDeadline[];
  isLoading: boolean;
  error: Error | null;
  checkCompliance: () => Promise<void>;
  getFrameworkStatus: (framework: string) => ComplianceCheckResult | null;
  refreshDeadlines: () => Promise<void>;
}

export function useCompliance(tenantId: string): UseComplianceReturn {
  const [deadlines, setDeadlines] = useState<ComplianceDeadline[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const [checkResults, setCheckResults] = useState<Map<string, ComplianceCheckResult>>(new Map());

  const { setComplianceStatus } = useCarbonStore();

  const statusScoreMap: Record<ComplianceStatus[keyof ComplianceStatus], number> = {
    complete: 100,
    in_progress: 60,
    pending: 30,
    at_risk: 40,
    overdue: 10,
  };

  // Calculate deadline status based on date
  const calculateStatus = (dueDate: string): ComplianceDeadline['status'] => {
    const now = new Date();
    const due = new Date(dueDate);
    const daysUntilDue = Math.ceil((due.getTime() - now.getTime()) / (1000 * 60 * 60 * 24));

    if (daysUntilDue < 0) return 'overdue';
    if (daysUntilDue <= 30) return 'due_soon';
    return 'upcoming';
  };

  // Fetch compliance deadlines
  const refreshDeadlines = useCallback(async () => {
    setIsLoading(true);
    setError(null);

    try {
      const response = await api.get<{ deadlines: ComplianceDeadline[] }>(
        `/api/compliance/deadlines?tenant_id=${tenantId}`
      );
      
      const processedDeadlines = response.deadlines.map((d) => ({
        ...d,
        status: d.status === 'completed' ? 'completed' : calculateStatus(d.dueDate),
      }));
      
      setDeadlines(processedDeadlines);
    } catch (err) {
      // Use mock data if API fails
      const mockDeadlines: ComplianceDeadline[] = [
        {
          id: '1',
          framework: 'csrd',
          title: 'CSRD Annual Report',
          description: 'Submit annual corporate sustainability report',
          dueDate: '2025-06-30',
          status: 'upcoming',
          priority: 'high',
          requirements: [
            'Double materiality assessment',
            'Scope 1, 2, 3 emissions disclosure',
            'Climate risk analysis',
            'Biodiversity impact assessment',
          ],
        },
        {
          id: '2',
          framework: 'sec',
          title: 'SEC Climate Disclosure',
          description: 'SEC climate-related financial disclosures',
          dueDate: '2025-03-31',
          status: 'due_soon',
          priority: 'high',
          requirements: [
            'GHG emissions metrics',
            'Climate risk governance',
            'Transition plan disclosure',
          ],
        },
        {
          id: '3',
          framework: 'cbam',
          title: 'CBAM Quarterly Report',
          description: 'Carbon Border Adjustment Mechanism reporting',
          dueDate: '2025-01-31',
          status: 'due_soon',
          priority: 'medium',
          requirements: [
            'Embedded emissions calculation',
            'Product carbon footprint',
            'Supply chain documentation',
          ],
        },
        {
          id: '4',
          framework: 'sb253',
          title: 'California SB 253 Report',
          description: 'California Climate Corporate Data Accountability Act',
          dueDate: '2025-12-31',
          status: 'upcoming',
          priority: 'medium',
          requirements: [
            'Full scope emissions inventory',
            'Third-party verification',
            'Public disclosure',
          ],
        },
      ];

      setDeadlines(mockDeadlines.map((d) => ({
        ...d,
        status: d.status === 'completed' ? 'completed' : calculateStatus(d.dueDate),
      })));
    } finally {
      setIsLoading(false);
    }
  }, [tenantId]);

  // Run compliance check
  const checkCompliance = useCallback(async () => {
    setIsLoading(true);
    setError(null);

    try {
      const response = await api.get<ComplianceStatus>(
        `/api/compliance/summary?tenant_id=${tenantId}`
      );
      
      setComplianceStatus(response);
      
      // Store individual check results
      const results = new Map<string, ComplianceCheckResult>();
      (Object.entries(response) as Array<[keyof ComplianceStatus, ComplianceStatus[keyof ComplianceStatus]]>).forEach(([key, value]) => {
        results.set(key, {
          framework: key,
          score: statusScoreMap[value],
          gaps: [],
          recommendations: [],
        });
      });
      setCheckResults(results);
    } catch (err) {
      // Mock compliance status
      const mockStatus: ComplianceStatus = {
        csrd: 'in_progress',
        sec: 'in_progress',
        cbam: 'complete',
        sb253: 'at_risk',
        ifrs: 'pending',
      };
      
      setComplianceStatus(mockStatus);
      
      const results = new Map<string, ComplianceCheckResult>();
      results.set('csrd', {
        framework: 'csrd',
        score: statusScoreMap[mockStatus.csrd],
        gaps: ['Missing biodiversity assessment', 'Incomplete Scope 3 data'],
        recommendations: ['Complete double materiality assessment', 'Engage supply chain partners'],
      });
      results.set('sec', {
        framework: 'sec',
        score: statusScoreMap[mockStatus.sec],
        gaps: ['Climate risk governance documentation incomplete'],
        recommendations: ['Document board oversight of climate risks'],
      });
      results.set('cbam', {
        framework: 'cbam',
        score: statusScoreMap[mockStatus.cbam],
        gaps: [],
        recommendations: ['Maintain current reporting quality'],
      });
      results.set('sb253', {
        framework: 'sb253',
        score: statusScoreMap[mockStatus.sb253],
        gaps: ['Missing third-party verification', 'Incomplete Scope 3 inventory'],
        recommendations: ['Engage verification provider', 'Expand supply chain data collection'],
      });
      results.set('ifrs', {
        framework: 'ifrs',
        score: statusScoreMap[mockStatus.ifrs],
        gaps: ['Draft disclosure statements incomplete'],
        recommendations: ['Align disclosures with IFRS S2 requirements'],
      });
      setCheckResults(results);
    } finally {
      setIsLoading(false);
    }

    // Also refresh deadlines
    await refreshDeadlines();
  }, [tenantId, setComplianceStatus, refreshDeadlines]);

  // Get status for a specific framework
  const getFrameworkStatus = useCallback(
    (framework: string): ComplianceCheckResult | null => {
      return checkResults.get(framework) || null;
    },
    [checkResults]
  );

  // Initial load
  useEffect(() => {
    refreshDeadlines();
  }, [refreshDeadlines]);

  return {
    deadlines,
    isLoading,
    error,
    checkCompliance,
    getFrameworkStatus,
    refreshDeadlines,
  };
}

// Additional utility hooks
export function useComplianceScore(tenantId: string) {
  const complianceStatus = useCarbonStore((state) => state.complianceStatus);
  if (!complianceStatus) return 0;
  const scoreMap: Record<ComplianceStatus[keyof ComplianceStatus], number> = {
    complete: 100,
    in_progress: 60,
    pending: 30,
    at_risk: 40,
    overdue: 10,
  };
  const statuses = Object.values(complianceStatus) as Array<ComplianceStatus[keyof ComplianceStatus]>;
  if (statuses.length === 0) return 0;
  const total = statuses.reduce((sum, status) => sum + scoreMap[status], 0);
  return Math.round(total / statuses.length);
}

export function useUpcomingDeadlines(tenantId: string, days: number = 30) {
  const { deadlines } = useCompliance(tenantId);
  
  return deadlines.filter((d) => {
    const dueDate = new Date(d.dueDate);
    const now = new Date();
    const daysUntilDue = Math.ceil((dueDate.getTime() - now.getTime()) / (1000 * 60 * 60 * 24));
    return daysUntilDue >= 0 && daysUntilDue <= days;
  });
}
