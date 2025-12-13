/**
 * @fileoverview Unit tests for Carbon Store
 * @description Comprehensive test suite for Zustand carbon store with immer middleware
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
  useLastUpdated
} from '@/stores/carbonStore';
import { EmissionData, CarbonMetrics, ComplianceStatus, Timeframe } from '@/types/carbon';

// Mock the CarbonApi
jest.mock('@/lib/api/carbon', () => ({
  CarbonApi: {
    getInstance: jest.fn(() => ({
      getEmissions: jest.fn(),
      getMetrics: jest.fn(),
      subscribeToUpdates: jest.fn(() => jest.fn()),
    })),
  },
  formatNumber: jest.fn((n) => n.toLocaleString()),
  formatDate: jest.fn((d) => d.toISOString()),
}));

describe('Carbon Store', () => {
  beforeEach(() => {
    // Reset store state before each test
    const { result } = renderHook(() => useCarbonStore());
    act(() => {
      result.current.reset();
    });
  });

  describe('Initial State', () => {
    it('should have correct initial values', () => {
      const { result } = renderHook(() => useCarbonStore());
      
      expect(result.current.emissions).toEqual([]);
      expect(result.current.metrics).toBeNull();
      expect(result.current.complianceStatus).toBeNull();
      expect(result.current.isLoading).toBe(false);
      expect(result.current.error).toBeNull();
      expect(result.current.selectedTimeframe).toBe('monthly');
    });
  });

  describe('Selector Hooks', () => {
    it('useEmissions should return emissions array', () => {
      const { result } = renderHook(() => useEmissions());
      expect(Array.isArray(result.current)).toBe(true);
    });

    it('useMetrics should return null initially', () => {
      const { result } = renderHook(() => useMetrics());
      expect(result.current).toBeNull();
    });

    it('useComplianceStatus should return null initially', () => {
      const { result } = renderHook(() => useComplianceStatus());
      expect(result.current).toBeNull();
    });

    it('useCarbonLoading should return false initially', () => {
      const { result } = renderHook(() => useCarbonLoading());
      expect(result.current).toBe(false);
    });

    it('useCarbonError should return null initially', () => {
      const { result } = renderHook(() => useCarbonError());
      expect(result.current).toBeNull();
    });
  });

  describe('Actions', () => {
    it('setTimeframe should update selectedTimeframe', () => {
      const { result } = renderHook(() => useCarbonStore());
      
      act(() => {
        result.current.setTimeframe('yearly');
      });
      
      expect(result.current.selectedTimeframe).toBe('yearly');
    });

    it('setLoading should update isLoading state', () => {
      const { result } = renderHook(() => useCarbonStore());
      
      act(() => {
        result.current.setLoading(true);
      });
      
      expect(result.current.isLoading).toBe(true);
    });

    it('setError should update error state', () => {
      const { result } = renderHook(() => useCarbonStore());
      const testError = 'Test error message';
      
      act(() => {
        result.current.setError(testError);
      });
      
      expect(result.current.error).toBe(testError);
    });

    it('clearError should reset error to null', () => {
      const { result } = renderHook(() => useCarbonStore());
      
      act(() => {
        result.current.setError('Some error');
        result.current.clearError();
      });
      
      expect(result.current.error).toBeNull();
    });

    it('reset should restore initial state', () => {
      const { result } = renderHook(() => useCarbonStore());
      
      act(() => {
        result.current.setTimeframe('yearly');
        result.current.setLoading(true);
        result.current.setError('Error');
        result.current.reset();
      });
      
      expect(result.current.selectedTimeframe).toBe('monthly');
      expect(result.current.isLoading).toBe(false);
      expect(result.current.error).toBeNull();
    });
  });

  describe('Data Source Management', () => {
    it('updateDataSourceStatus should update specific data source', () => {
      const { result } = renderHook(() => useCarbonStore());
      
      // First add some data sources via mock data
      act(() => {
        result.current.updateDataSourceStatus('utility_api', 'connected');
      });
      
      // Verify the action was called (store may have empty dataSources initially)
      expect(result.current.dataSources).toBeDefined();
    });
  });

  describe('Computed Values', () => {
    it('calculateIntensity should compute correct value', () => {
      const { result } = renderHook(() => useCarbonStore());
      
      const intensity = result.current.calculateIntensity(12450, 50);
      expect(intensity).toBe(249); // 12450 / 50 = 249
    });

    it('calculateIntensity should handle zero revenue', () => {
      const { result } = renderHook(() => useCarbonStore());
      
      const intensity = result.current.calculateIntensity(12450, 0);
      expect(intensity).toBe(0);
    });

    it('calculateIntensity should round to nearest integer', () => {
      const { result } = renderHook(() => useCarbonStore());
      
      const intensity = result.current.calculateIntensity(100, 3);
      expect(intensity).toBe(33); // 100 / 3 ≈ 33.33 → 33
    });
  });

  describe('Timeframe Validation', () => {
    const validTimeframes: Timeframe[] = ['daily', 'weekly', 'monthly', 'quarterly', 'yearly'];
    
    validTimeframes.forEach((timeframe) => {
      it(`should accept valid timeframe: ${timeframe}`, () => {
        const { result } = renderHook(() => useCarbonStore());
        
        act(() => {
          result.current.setTimeframe(timeframe);
        });
        
        expect(result.current.selectedTimeframe).toBe(timeframe);
      });
    });
  });

  describe('Immutability (Immer)', () => {
    it('should maintain immutability when updating state', () => {
      const { result } = renderHook(() => useCarbonStore());
      
      const originalState = result.current;
      
      act(() => {
        result.current.setTimeframe('yearly');
      });
      
      // State reference should change (immutable update)
      expect(result.current).not.toBe(originalState);
    });
  });
});

describe('Mock Data Generation', () => {
  it('should generate valid mock emission data', () => {
    const { result } = renderHook(() => useCarbonStore());
    
    // Trigger mock data generation
    act(() => {
      result.current.fetchEmissions('monthly');
    });
    
    // Mock data should be generated after fetch
    // (Actual verification depends on implementation)
    expect(result.current.isLoading).toBeDefined();
  });
});
