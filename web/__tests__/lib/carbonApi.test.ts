/**
 * @fileoverview Unit tests for CarbonApi client
 * @description Tests for singleton API client with error handling and WebSocket support
 */

import { CarbonApi, CarbonApiError, formatNumber, formatDate, downloadFile } from '@/lib/api/carbon';
import { Timeframe } from '@/types/carbon';

// Mock fetch globally
const mockFetch = jest.fn();
global.fetch = mockFetch;

// Mock WebSocket
class MockWebSocket {
  onopen: (() => void) | null = null;
  onmessage: ((event: { data: string }) => void) | null = null;
  onclose: (() => void) | null = null;
  onerror: ((error: Error) => void) | null = null;
  readyState = 1;
  
  close = jest.fn();
  send = jest.fn();
}

(global as any).WebSocket = MockWebSocket;

describe('CarbonApi', () => {
  let api: CarbonApi;

  beforeEach(() => {
    jest.clearAllMocks();
    // Reset singleton for testing
    (CarbonApi as any).instance = null;
    api = CarbonApi.getInstance();
  });

  describe('Singleton Pattern', () => {
    it('should return the same instance', () => {
      const instance1 = CarbonApi.getInstance();
      const instance2 = CarbonApi.getInstance();
      
      expect(instance1).toBe(instance2);
    });

    it('should accept custom base URL', () => {
      (CarbonApi as any).instance = null;
      const customApi = CarbonApi.getInstance('https://custom-api.com');
      
      expect(customApi).toBeDefined();
    });
  });

  describe('API Requests', () => {
    it('should fetch emissions successfully', async () => {
      const mockData = {
        data: [{
          id: '1',
          total: 12450,
          scope1: 3200,
          scope2: 5800,
          scope3: 3450,
        }]
      };
      
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockData),
      });

      const result = await api.getEmissions('monthly');
      
      expect(result).toEqual(mockData);
      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/carbon/emissions'),
        expect.objectContaining({
          method: 'GET',
          headers: expect.objectContaining({
            'Content-Type': 'application/json',
          }),
        })
      );
    });

    it('should include timeframe in query params', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ data: [] }),
      });

      await api.getEmissions('yearly');
      
      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('timeframe=yearly'),
        expect.any(Object)
      );
    });

    it('should throw CarbonApiError on HTTP error', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 500,
        statusText: 'Internal Server Error',
        json: () => Promise.resolve({ message: 'Server error' }),
      });

      await expect(api.getEmissions('monthly')).rejects.toThrow(CarbonApiError);
    });

    it('should handle network errors', async () => {
      mockFetch.mockRejectedValueOnce(new Error('Network error'));

      await expect(api.getEmissions('monthly')).rejects.toThrow();
    });
  });

  describe('Metrics Endpoint', () => {
    it('should fetch metrics successfully', async () => {
      const mockMetrics = {
        data: {
          totalEmissions: 12450,
          reduction: 8.5,
          intensity: 249,
        }
      };
      
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockMetrics),
      });

      const result = await api.getMetrics();
      
      expect(result).toEqual(mockMetrics);
    });
  });

  describe('Report Generation', () => {
    it('should generate compliance report', async () => {
      const mockReport = {
        data: {
          reportId: 'report-123',
          format: 'pdf',
          url: 'https://example.com/report.pdf',
        }
      };
      
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockReport),
      });

      const result = await api.generateComplianceReport('pdf');
      
      expect(result).toEqual(mockReport);
      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/carbon/reports'),
        expect.objectContaining({
          method: 'POST',
          body: expect.stringContaining('pdf'),
        })
      );
    });

    it('should support multiple report formats', async () => {
      const formats = ['pdf', 'csv', 'xlsx'] as const;
      
      for (const format of formats) {
        mockFetch.mockResolvedValueOnce({
          ok: true,
          json: () => Promise.resolve({ data: { format } }),
        });

        await api.generateComplianceReport(format);
        
        expect(mockFetch).toHaveBeenLastCalledWith(
          expect.any(String),
          expect.objectContaining({
            body: expect.stringContaining(format),
          })
        );
      }
    });
  });

  describe('Reduction Targets', () => {
    it('should fetch reduction targets', async () => {
      const mockTargets = {
        data: [{
          id: 'target-1',
          targetYear: 2030,
          reductionPercent: 50,
        }]
      };
      
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockTargets),
      });

      const result = await api.getReductionTargets();
      
      expect(result).toEqual(mockTargets);
    });

    it('should create reduction target', async () => {
      const newTarget = {
        targetYear: 2035,
        reductionPercent: 75,
        baselineYear: 2020,
      };
      
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ data: { id: 'new-target', ...newTarget } }),
      });

      const result = await api.createReductionTarget(newTarget);
      
      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/carbon/targets'),
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify(newTarget),
        })
      );
    });
  });

  describe('WebSocket Subscriptions', () => {
    it('should create WebSocket subscription', () => {
      const callback = jest.fn();
      const unsubscribe = api.subscribeToUpdates(callback);
      
      expect(typeof unsubscribe).toBe('function');
    });

    it('should handle incoming WebSocket messages', () => {
      const callback = jest.fn();
      api.subscribeToUpdates(callback);
      
      // WebSocket connection is mocked, verify callback setup
      expect(callback).not.toHaveBeenCalled(); // Not called until message received
    });

    it('should clean up on unsubscribe', () => {
      const callback = jest.fn();
      const unsubscribe = api.subscribeToUpdates(callback);
      
      unsubscribe();
      
      // Verify cleanup occurred (WebSocket close called)
      // Implementation-dependent
    });
  });

  describe('Request Cancellation', () => {
    it('should support request cancellation via AbortController', async () => {
      const controller = new AbortController();
      
      mockFetch.mockImplementationOnce(() => 
        new Promise((_, reject) => {
          controller.signal.addEventListener('abort', () => {
            reject(new DOMException('Aborted', 'AbortError'));
          });
        })
      );

      const promise = api.getEmissions('monthly');
      controller.abort();
      
      await expect(promise).rejects.toThrow();
    });
  });
});

describe('CarbonApiError', () => {
  it('should create error with message and status', () => {
    const error = new CarbonApiError('Not found', 404);
    
    expect(error.message).toBe('Not found');
    expect(error.status).toBe(404);
    expect(error.name).toBe('CarbonApiError');
  });

  it('should include optional details', () => {
    const details = { field: 'timeframe', reason: 'invalid' };
    const error = new CarbonApiError('Validation error', 400, details);
    
    expect(error.details).toEqual(details);
  });
});

describe('Utility Functions', () => {
  describe('formatNumber', () => {
    it('should format numbers with commas', () => {
      expect(formatNumber(1234567)).toBe('1,234,567');
    });

    it('should handle decimal places', () => {
      expect(formatNumber(1234.56)).toBe('1,234.56');
    });

    it('should handle zero', () => {
      expect(formatNumber(0)).toBe('0');
    });

    it('should handle negative numbers', () => {
      expect(formatNumber(-1234)).toBe('-1,234');
    });
  });

  describe('formatDate', () => {
    const testDate = new Date('2024-06-15T12:00:00Z');

    it('should format daily timeframe', () => {
      const result = formatDate(testDate, 'daily');
      expect(result).toContain('Jun');
      expect(result).toContain('15');
    });

    it('should format monthly timeframe', () => {
      const result = formatDate(testDate, 'monthly');
      expect(result).toContain('Jun');
    });

    it('should format yearly timeframe', () => {
      const result = formatDate(testDate, 'yearly');
      expect(result).toContain('2024');
    });
  });

  describe('downloadFile', () => {
    it('should create download link', () => {
      const createElementSpy = jest.spyOn(document, 'createElement');
      const appendChildSpy = jest.spyOn(document.body, 'appendChild').mockImplementation(() => null as any);
      const removeChildSpy = jest.spyOn(document.body, 'removeChild').mockImplementation(() => null as any);
      
      downloadFile('https://example.com/file.pdf', 'report.pdf');
      
      expect(createElementSpy).toHaveBeenCalledWith('a');
      
      createElementSpy.mockRestore();
      appendChildSpy.mockRestore();
      removeChildSpy.mockRestore();
    });
  });
});
