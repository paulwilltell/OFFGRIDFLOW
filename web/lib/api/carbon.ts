import { 
  EmissionData, 
  CarbonMetrics, 
  ApiResponse, 
  ApiError,
  Timeframe,
  QueryOptions,
  ComplianceReport,
  ReportFormat
} from '@/types/carbon';

// ============================================================================
// Configuration
// ============================================================================

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8090';
const API_VERSION = 'v1';

interface RequestConfig extends RequestInit {
  params?: Record<string, string | number | boolean | undefined>;
  timeout?: number;
}

// ============================================================================
// Carbon API Client (Singleton)
// ============================================================================

export class CarbonApi {
  private static instance: CarbonApi;
  private baseUrl: string;
  private token: string | null = null;

  private constructor() {
    this.baseUrl = `${API_BASE_URL}/api/${API_VERSION}`;
  }

  static getInstance(): CarbonApi {
    if (!CarbonApi.instance) {
      CarbonApi.instance = new CarbonApi();
    }
    return CarbonApi.instance;
  }

  // ============================================================================
  // Authentication
  // ============================================================================

  setToken(token: string): void {
    this.token = token;
  }

  clearToken(): void {
    this.token = null;
  }

  // ============================================================================
  // HTTP Methods
  // ============================================================================

  private async request<T>(
    endpoint: string, 
    config: RequestConfig = {}
  ): Promise<T> {
    const { params, timeout = 30000, ...fetchConfig } = config;
    
    // Build URL with query params
    let url = `${this.baseUrl}${endpoint}`;
    if (params) {
      const searchParams = new URLSearchParams();
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined) {
          searchParams.append(key, String(value));
        }
      });
      const queryString = searchParams.toString();
      if (queryString) {
        url += `?${queryString}`;
      }
    }

    // Set up headers
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      ...fetchConfig.headers,
    };

    if (this.token) {
      (headers as Record<string, string>)['Authorization'] = `Bearer ${this.token}`;
    }

    // Create abort controller for timeout
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), timeout);

    try {
      const response = await fetch(url, {
        ...fetchConfig,
        headers,
        signal: controller.signal,
      });

      clearTimeout(timeoutId);

      if (!response.ok) {
        const error = await response.json() as ApiError;
        throw new CarbonApiError(
          error.message || `HTTP ${response.status}`,
          error.code || 'UNKNOWN_ERROR',
          response.status,
          error.requestId
        );
      }

      const data = await response.json() as ApiResponse<T>;
      return data.data;
    } catch (error) {
      clearTimeout(timeoutId);
      
      if (error instanceof CarbonApiError) {
        throw error;
      }
      
      if (error instanceof Error) {
        if (error.name === 'AbortError') {
          throw new CarbonApiError('Request timeout', 'TIMEOUT', 408);
        }
        throw new CarbonApiError(error.message, 'NETWORK_ERROR', 0);
      }
      
      throw new CarbonApiError('Unknown error occurred', 'UNKNOWN_ERROR', 0);
    }
  }

  private async get<T>(endpoint: string, params?: Record<string, string | number | boolean | undefined>): Promise<T> {
    return this.request<T>(endpoint, { method: 'GET', params });
  }

  private async post<T>(endpoint: string, data?: unknown): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'POST',
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  private async put<T>(endpoint: string, data?: unknown): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'PUT',
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  private async delete<T>(endpoint: string): Promise<T> {
    return this.request<T>(endpoint, { method: 'DELETE' });
  }

  // ============================================================================
  // Emissions API
  // ============================================================================

  async getEmissions(tenantId: string, timeframe: Timeframe): Promise<EmissionData> {
    return this.get<EmissionData>('/emissions', { tenantId, timeframe });
  }

  async getEmissionsSummary(tenantId: string, options?: QueryOptions): Promise<EmissionData> {
    return this.get<EmissionData>('/emissions/summary', {
      tenantId,
      timeframe: options?.filters?.timeframe,
      page: options?.page,
      limit: options?.limit,
    });
  }

  async getEmissionsByScope(
    tenantId: string, 
    scope: 1 | 2 | 3, 
    timeframe: Timeframe
  ): Promise<EmissionData[]> {
    return this.get<EmissionData[]>(`/emissions/scope/${scope}`, { tenantId, timeframe });
  }

  async createEmission(tenantId: string, data: Partial<EmissionData>): Promise<EmissionData> {
    return this.post<EmissionData>('/emissions', { tenantId, ...data });
  }

  async updateEmission(emissionId: string, data: Partial<EmissionData>): Promise<EmissionData> {
    return this.put<EmissionData>(`/emissions/${emissionId}`, data);
  }

  async deleteEmission(emissionId: string): Promise<void> {
    return this.delete<void>(`/emissions/${emissionId}`);
  }

  // ============================================================================
  // Metrics API
  // ============================================================================

  async getMetrics(tenantId: string): Promise<CarbonMetrics> {
    return this.get<CarbonMetrics>('/metrics/carbon', { tenantId });
  }

  async updateMetrics(tenantId: string, data: Partial<CarbonMetrics>): Promise<CarbonMetrics> {
    return this.put<CarbonMetrics>('/metrics/carbon', { tenantId, ...data });
  }

  async getIndustryBenchmark(industry: string): Promise<{ average: number; percentile: number }> {
    return this.get<{ average: number; percentile: number }>('/metrics/benchmark', { industry });
  }

  // ============================================================================
  // Compliance API
  // ============================================================================

  async getComplianceStatus(tenantId: string): Promise<Record<string, unknown>> {
    return this.get<Record<string, unknown>>('/compliance/status', { tenantId });
  }

  async getComplianceDeadlines(tenantId: string): Promise<unknown[]> {
    return this.get<unknown[]>('/compliance/deadlines', { tenantId });
  }

  async generateComplianceReport(
    tenantId: string,
    format: ReportFormat,
    scopes: number[]
  ): Promise<ComplianceReport> {
    return this.post<ComplianceReport>('/compliance/report', {
      tenantId,
      format,
      includeScopes: scopes,
    });
  }

  // ============================================================================
  // Data Sources API
  // ============================================================================

  async getDataSources(tenantId: string): Promise<unknown[]> {
    return this.get<unknown[]>('/data-sources', { tenantId });
  }

  async addDataSource(tenantId: string, source: unknown): Promise<unknown> {
    return this.post<unknown>('/data-sources', { tenantId, ...source as object });
  }

  async syncDataSource(sourceId: string): Promise<{ status: string; lastSync: Date }> {
    return this.post<{ status: string; lastSync: Date }>(`/data-sources/${sourceId}/sync`);
  }

  async deleteDataSource(sourceId: string): Promise<void> {
    return this.delete<void>(`/data-sources/${sourceId}`);
  }

  // ============================================================================
  // Real-time Subscription
  // ============================================================================

  subscribeToUpdates(
    tenantId: string,
    onUpdate: (data: unknown) => void,
    onError?: (error: Error) => void
  ): () => void {
    const wsUrl = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8090/ws';
    const ws = new WebSocket(`${wsUrl}?tenant=${tenantId}`);

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        onUpdate(data);
      } catch (error) {
        console.error('Failed to parse WebSocket message:', error);
      }
    };

    ws.onerror = (event) => {
      onError?.(new Error('WebSocket error'));
    };

    ws.onclose = () => {
      console.log('WebSocket connection closed');
    };

    // Return unsubscribe function
    return () => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.close();
      }
    };
  }
}

// ============================================================================
// Custom Error Class
// ============================================================================

export class CarbonApiError extends Error {
  constructor(
    message: string,
    public code: string,
    public status: number,
    public requestId?: string
  ) {
    super(message);
    this.name = 'CarbonApiError';
  }

  toJSON() {
    return {
      name: this.name,
      message: this.message,
      code: this.code,
      status: this.status,
      requestId: this.requestId,
    };
  }
}

// ============================================================================
// Utility Functions
// ============================================================================

export function formatNumber(value: number): string {
  if (value >= 1000000) {
    return `${(value / 1000000).toFixed(1)}M`;
  }
  if (value >= 1000) {
    return `${(value / 1000).toFixed(1)}K`;
  }
  return value.toFixed(2);
}

export function formatDate(date: Date | string, timeframe: Timeframe): string {
  const d = new Date(date);
  
  switch (timeframe) {
    case 'daily':
      return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
    case 'weekly':
      return `Week ${Math.ceil(d.getDate() / 7)}`;
    case 'monthly':
      return d.toLocaleDateString('en-US', { month: 'short', year: '2-digit' });
    case 'quarterly':
      return `Q${Math.ceil((d.getMonth() + 1) / 3)} ${d.getFullYear()}`;
    case 'yearly':
      return d.getFullYear().toString();
    default:
      return d.toLocaleDateString();
  }
}

export function downloadFile(data: Blob | string, filename: string): void {
  const blob = typeof data === 'string' ? new Blob([data], { type: 'text/plain' }) : data;
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
}

// ============================================================================
// Default Export
// ============================================================================

export default CarbonApi;
