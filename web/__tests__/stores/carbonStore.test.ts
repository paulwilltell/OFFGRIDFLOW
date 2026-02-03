/**
 * @fileoverview Unit tests for Carbon Store
 * @description Tests for Zustand carbon store with immer middleware
 */

import { act, renderHook } from '@testing-library/react';
import {
  useCarbonStore,
  useEmissions,
  useMetrics,
  useComplianceStatus,
  useDataSources,
  useReductionTargets,
  useCarbonLoading,
  useCarbonError,
  useLastUpdated,
} from '@/stores/carbonStore';
import { EmissionData } from '@/types/carbon';

const mockGetEmissions = jest.fn();

// Mock the CarbonApi
jest.mock('@/lib/api/carbon', () => ({
  CarbonApi: {
    getInstance: jest.fn(() => ({
      getEmissions: (...args: unknown[]) => mockGetEmissions(...args),
      subscribeToUpdates: jest.fn(() => jest.fn()),
    })),
  },
  formatNumber: jest.fn((n) => n.toLocaleString()),
  formatDate: jest.fn((d) => d.toISOString()),
}));

const mockEmission: EmissionData = {
  id: 'emit-1',
  tenantId: 'tenant-1',
  total: 12450,
  scope1: 3200,
  scope2: 5800,
  scope3: 3450,
  intensity: 249,
  timeframe: 'monthly',
  dataSources: [],
  updatedAt: new Date('2024-01-15T00:00:00Z'),
  methodology: 'ghg_protocol',
  uncertainty: 5,
  region: 'north_america',
};

describe('Carbon Store', () => {
  beforeEach(() => {
    const { result } = renderHook(() => useCarbonStore());
    act(() => {
      result.current.reset();
    });
    mockGetEmissions.mockReset();
  });

  describe('Initial State', () => {
    it('should have correct initial values', () => {
      const { result } = renderHook(() => useCarbonStore());

      expect(result.current.emissions).toBeNull();
      expect(result.current.metrics).toEqual(
        expect.objectContaining({
          totalEmissions: 0,
          carbonIntensity: 0,
        })
      );
      expect(result.current.complianceStatus).toEqual(
        expect.objectContaining({
          csrd: 'pending',
        })
      );
      expect(result.current.isLoading).toBe(false);
      expect(result.current.error).toBeNull();
      expect(result.current.lastUpdated).toBeNull();
    });
  });

  describe('Selector Hooks', () => {
    it('useEmissions should return null initially', () => {
      const { result } = renderHook(() => useEmissions());
      expect(result.current).toBeNull();
    });

    it('useMetrics should return metrics object', () => {
      const { result } = renderHook(() => useMetrics());
      expect(result.current).toEqual(
        expect.objectContaining({
          totalEmissions: 0,
        })
      );
    });

    it('useComplianceStatus should return default compliance status', () => {
      const { result } = renderHook(() => useComplianceStatus());
      expect(result.current).toEqual(
        expect.objectContaining({
          csrd: 'pending',
        })
      );
    });

    it('useCarbonLoading should return false initially', () => {
      const { result } = renderHook(() => useCarbonLoading());
      expect(result.current).toBe(false);
    });

    it('useCarbonError should return null initially', () => {
      const { result } = renderHook(() => useCarbonError());
      expect(result.current).toBeNull();
    });

    it('useDataSources should return empty array initially', () => {
      const { result } = renderHook(() => useDataSources());
      expect(result.current).toEqual([]);
    });

    it('useReductionTargets should return empty array initially', () => {
      const { result } = renderHook(() => useReductionTargets());
      expect(result.current).toEqual([]);
    });

    it('useLastUpdated should return null initially', () => {
      const { result } = renderHook(() => useLastUpdated());
      expect(result.current).toBeNull();
    });
  });

  describe('Actions', () => {
    it('updateMetrics should merge metric updates', () => {
      const { result } = renderHook(() => useCarbonStore());

      act(() => {
        result.current.updateMetrics({ revenue: 1000000, totalEmissions: 500 });
      });

      expect(result.current.metrics.revenue).toBe(1000000);
      expect(result.current.metrics.totalEmissions).toBe(500);
    });

    it('setComplianceStatus should merge status updates', () => {
      const { result } = renderHook(() => useCarbonStore());

      act(() => {
        result.current.setComplianceStatus({ csrd: 'complete' });
      });

      expect(result.current.complianceStatus.csrd).toBe('complete');
    });

    it('add/remove data sources should update list', () => {
      const { result } = renderHook(() => useCarbonStore());
      const source = {
        id: 'src-1',
        name: 'Utility API',
        type: 'manual',
        status: 'active',
        lastSync: new Date(),
        emissions: 120,
        coordinates: { lat: 0, lng: 0 },
      };

      act(() => {
        result.current.addDataSource(source);
      });

      expect(result.current.dataSources).toHaveLength(1);

      act(() => {
        result.current.removeDataSource('src-1');
      });

      expect(result.current.dataSources).toHaveLength(0);
    });

    it('updateDataSourceStatus should update existing source', () => {
      const { result } = renderHook(() => useCarbonStore());
      const source = {
        id: 'src-1',
        name: 'Utility API',
        type: 'manual',
        status: 'inactive',
        lastSync: new Date(),
        emissions: 120,
        coordinates: { lat: 0, lng: 0 },
      };

      act(() => {
        result.current.addDataSource(source);
        result.current.updateDataSourceStatus('src-1', 'active');
      });

      expect(result.current.dataSources[0].status).toBe('active');
    });

    it('updateEmission should patch current emissions', () => {
      const { result } = renderHook(() => useCarbonStore());

      act(() => {
        useCarbonStore.setState({ emissions: { ...mockEmission } });
      });

      act(() => {
        result.current.updateEmission({ total: 13000 });
      });

      expect(result.current.emissions?.total).toBe(13000);
    });

    it('reset should restore initial state', () => {
      const { result } = renderHook(() => useCarbonStore());

      act(() => {
        useCarbonStore.setState({ emissions: { ...mockEmission }, error: 'boom' });
        result.current.reset();
      });

      expect(result.current.emissions).toBeNull();
      expect(result.current.error).toBeNull();
      expect(result.current.lastUpdated).toBeNull();
    });
  });

  describe('fetchEmissions', () => {
    it('should store emissions and update metrics', async () => {
      mockGetEmissions.mockResolvedValueOnce({ ...mockEmission });

      const { result } = renderHook(() => useCarbonStore());

      await act(async () => {
        await result.current.fetchEmissions('tenant-1', 'monthly');
      });

      expect(result.current.emissions?.id).toBe('emit-1');
      expect(result.current.metrics.totalEmissions).toBe(12450);
      expect(result.current.lastUpdated).not.toBeNull();
    });
  });

  describe('calculateIntensity', () => {
    it('should compute intensity based on emissions and revenue', () => {
      const { result } = renderHook(() => useCarbonStore());

      act(() => {
        useCarbonStore.setState({
          emissions: { ...mockEmission, total: 12450 },
          metrics: { ...result.current.metrics, revenue: 50 },
        });
      });

      const intensity = result.current.calculateIntensity();
      expect(intensity).toBe(249000000);
    });

    it('should return 0 when revenue is missing', () => {
      const { result } = renderHook(() => useCarbonStore());

      act(() => {
        useCarbonStore.setState({
          emissions: { ...mockEmission, total: 12450 },
          metrics: { ...result.current.metrics, revenue: 0 },
        });
      });

      expect(result.current.calculateIntensity()).toBe(0);
    });
  });
});
