/**
 * @fileoverview Unit tests for CarbonApi client
 * @description Tests for singleton API client with error handling and WebSocket support
 */

import { CarbonApi, CarbonApiError, formatNumber, formatDate, downloadFile } from '@/lib/api/carbon';

// Mock fetch globally
const mockFetch = jest.fn();
global.fetch = mockFetch as any;

// Mock WebSocket
let lastWebSocketUrl: string | null = null;
class MockWebSocket {
  onopen: (() => void) | null = null;
  onmessage: ((event: { data: string }) => void) | null = null;
  onclose: (() => void) | null = null;
  onerror: ((error: Error) => void) | null = null;
  readyState = 1;

  constructor(url: string) {
    lastWebSocketUrl = url;
  }

  close = jest.fn();
  send = jest.fn();
}

(global as any).WebSocket = MockWebSocket;

describe('CarbonApi', () => {
  let api: CarbonApi;

  beforeEach(() => {
    jest.clearAllMocks();
    lastWebSocketUrl = null;
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
  });

  describe('API Requests', () => {
    it('should fetch emissions successfully', async () => {
      const mockData = {
        data: {
          id: '1',
          tenantId: 'tenant-1',
          total: 12450,
          scope1: 3200,
          scope2: 5800,
          scope3: 3450,
        },
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockData),
      });

      const result = await api.getEmissions('tenant-1', 'monthly');

      expect(result).toEqual(mockData.data);
      const calledUrl = mockFetch.mock.calls[0][0] as string;
      expect(calledUrl).toContain('/api/v1/emissions');
      expect(calledUrl).toContain('tenantId=tenant-1');
      expect(calledUrl).toContain('timeframe=monthly');
    });

    it('should throw CarbonApiError on HTTP error', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 500,
        json: () =>
          Promise.resolve({
            message: 'Server error',
            code: 'SERVER_ERROR',
            requestId: 'req-123',
          }),
      });

      await expect(api.getEmissions('tenant-1', 'monthly')).rejects.toEqual(
        expect.objectContaining({
          name: 'CarbonApiError',
          message: 'Server error',
          code: 'SERVER_ERROR',
          status: 500,
          requestId: 'req-123',
        })
      );
    });

    it('should handle network errors', async () => {
      mockFetch.mockRejectedValueOnce(new Error('Network error'));

      await expect(api.getEmissions('tenant-1', 'monthly')).rejects.toEqual(
        expect.objectContaining({
          code: 'NETWORK_ERROR',
        })
      );
    });

    it('should map abort errors to timeout', async () => {
      mockFetch.mockRejectedValueOnce(new DOMException('Aborted', 'AbortError'));

      await expect(api.getEmissions('tenant-1', 'monthly')).rejects.toEqual(
        expect.objectContaining({
          code: 'TIMEOUT',
          status: 408,
        })
      );
    });
  });

  describe('Metrics Endpoint', () => {
    it('should fetch metrics successfully', async () => {
      const mockMetrics = {
        data: {
          totalEmissions: 12450,
          reduction: 8.5,
          intensity: 249,
        },
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockMetrics),
      });

      const result = await api.getMetrics('tenant-1');

      expect(result).toEqual(mockMetrics.data);
    });
  });

  describe('Report Generation', () => {
    it('should generate compliance report', async () => {
      const mockReport = {
        data: {
          reportId: 'report-123',
          format: 'pdf',
          url: 'https://example.com/report.pdf',
        },
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockReport),
      });

      const result = await api.generateComplianceReport('tenant-1', 'pdf', [1, 2, 3]);

      expect(result).toEqual(mockReport.data);
      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/compliance/report'),
        expect.objectContaining({
          method: 'POST',
          body: expect.stringContaining('includeScopes'),
        })
      );
    });
  });

  describe('WebSocket Subscriptions', () => {
    it('should create WebSocket subscription', () => {
      const callback = jest.fn();
      const unsubscribe = api.subscribeToUpdates('tenant-1', callback);

      expect(typeof unsubscribe).toBe('function');
      expect(lastWebSocketUrl).toContain('tenant=tenant-1');
    });
  });
});

describe('CarbonApiError', () => {
  it('should create error with message and status', () => {
    const error = new CarbonApiError('Not found', 'NOT_FOUND', 404, 'req-1');

    expect(error.message).toBe('Not found');
    expect(error.code).toBe('NOT_FOUND');
    expect(error.status).toBe(404);
    expect(error.requestId).toBe('req-1');
    expect(error.name).toBe('CarbonApiError');
  });

  it('should serialize to JSON', () => {
    const error = new CarbonApiError('Bad request', 'BAD_REQUEST', 400);
    expect(error.toJSON()).toEqual({
      name: 'CarbonApiError',
      message: 'Bad request',
      code: 'BAD_REQUEST',
      status: 400,
      requestId: undefined,
    });
  });
});

describe('Utility Functions', () => {
  describe('formatNumber', () => {
    it('should format large numbers with suffixes', () => {
      expect(formatNumber(1234567)).toBe('1.2M');
      expect(formatNumber(1234.56)).toBe('1.2K');
    });

    it('should format small numbers with fixed decimals', () => {
      expect(formatNumber(999)).toBe('999.00');
      expect(formatNumber(0)).toBe('0.00');
      expect(formatNumber(-1234)).toBe('-1234.00');
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
      expect(URL.createObjectURL).toHaveBeenCalled();

      createElementSpy.mockRestore();
      appendChildSpy.mockRestore();
      removeChildSpy.mockRestore();
    });
  });
});
