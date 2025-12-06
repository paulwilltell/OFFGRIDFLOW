/**
 * Compliance API Client
 * 
 * Handles regulatory compliance reporting (CSRD, SEC, California, CBAM, IFRS)
 */

import { APIError, RequestOptions } from './activities';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8090';

/**
 * Compliance report types
 */
export interface ComplianceReport {
  id: string;
  type: 'csrd' | 'sec' | 'california' | 'cbam' | 'ifrs';
  year: number;
  status: 'draft' | 'generated' | 'submitted' | 'approved';
  generatedAt: string;
  data: any;
  downloadUrl?: string;
}

export interface ComplianceSummary {
  csrd: {
    required: boolean;
    status: string;
    completeness: number;
  };
  sec: {
    required: boolean;
    status: string;
    completeness: number;
  };
  california: {
    required: boolean;
    status: string;
    completeness: number;
  };
  cbam: {
    required: boolean;
    status: string;
    completeness: number;
  };
  ifrs: {
    required: boolean;
    status: string;
    completeness: number;
  };
}

/**
 * Get auth token
 */
const getAuthToken = (): string | null => {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem('auth_token');
};

/**
 * Build headers
 */
const buildHeaders = (): HeadersInit => {
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
  };

  const token = getAuthToken();
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  return headers;
};

/**
 * Handle response
 */
const handleResponse = async <T>(response: Response): Promise<T> => {
  if (!response.ok) {
    const error = await response.json().catch(() => ({
      error: `HTTP ${response.status}`,
    }));

    throw new APIError(
      error.error || 'Request failed',
      response.status,
      error.details
    );
  }

  return response.json();
};

/**
 * Generate CSRD report
 */
export const generateCSRDReport = async (
  params: {
    year: number;
    format: 'pdf' | 'json';
  },
  options: RequestOptions = {}
): Promise<ComplianceReport> => {
  const response = await fetch(`${API_BASE}/api/compliance/csrd`, {
    method: 'POST',
    headers: buildHeaders(),
    body: JSON.stringify(params),
    signal: options.signal,
  });

  return handleResponse(response);
};

/**
 * Generate SEC Climate Disclosure report
 */
export const generateSECReport = async (
  params: {
    year: number;
    format: 'pdf' | 'json';
  },
  options: RequestOptions = {}
): Promise<ComplianceReport> => {
  const response = await fetch(`${API_BASE}/api/compliance/sec`, {
    method: 'POST',
    headers: buildHeaders(),
    body: JSON.stringify(params),
    signal: options.signal,
  });

  return handleResponse(response);
};

/**
 * Generate California Climate Disclosure report
 */
export const generateCaliforniaReport = async (
  params: {
    year: number;
    format: 'pdf' | 'json';
  },
  options: RequestOptions = {}
): Promise<ComplianceReport> => {
  const response = await fetch(`${API_BASE}/api/compliance/california`, {
    method: 'POST',
    headers: buildHeaders(),
    body: JSON.stringify(params),
    signal: options.signal,
  });

  return handleResponse(response);
};

/**
 * Generate CBAM report
 */
export const generateCBAMReport = async (
  params: {
    year: number;
    format: 'pdf' | 'json';
  },
  options: RequestOptions = {}
): Promise<ComplianceReport> => {
  const response = await fetch(`${API_BASE}/api/compliance/cbam`, {
    method: 'POST',
    headers: buildHeaders(),
    body: JSON.stringify(params),
    signal: options.signal,
  });

  return handleResponse(response);
};

/**
 * Generate IFRS S2 report
 */
export const generateIFRSReport = async (
  params: {
    year: number;
    format: 'pdf' | 'json';
  },
  options: RequestOptions = {}
): Promise<ComplianceReport> => {
  const response = await fetch(`${API_BASE}/api/compliance/ifrs`, {
    method: 'POST',
    headers: buildHeaders(),
    body: JSON.stringify(params),
    signal: options.signal,
  });

  return handleResponse(response);
};

/**
 * Get compliance summary (all frameworks)
 */
export const getComplianceSummary = async (
  options: RequestOptions = {}
): Promise<ComplianceSummary> => {
  const response = await fetch(`${API_BASE}/api/compliance/summary`, {
    method: 'GET',
    headers: buildHeaders(),
    signal: options.signal,
  });

  return handleResponse(response);
};

/**
 * Export compliance report
 */
export const exportComplianceReport = async (
  reportId: string,
  format: 'pdf' | 'json' | 'xlsx',
  options: RequestOptions = {}
): Promise<Blob> => {
  const response = await fetch(
    `${API_BASE}/api/compliance/export?reportId=${reportId}&format=${format}`,
    {
      method: 'GET',
      headers: buildHeaders(),
      signal: options.signal,
    }
  );

  if (!response.ok) {
    throw new APIError('Export failed', response.status);
  }

  return response.blob();
};

/**
 * Get compliance reports history
 */
export const getComplianceReports = async (
  params: {
    type?: string;
    year?: number;
    status?: string;
  } = {},
  options: RequestOptions = {}
): Promise<{ reports: ComplianceReport[] }> => {
  const searchParams = new URLSearchParams();
  if (params.type) searchParams.append('type', params.type);
  if (params.year) searchParams.append('year', params.year.toString());
  if (params.status) searchParams.append('status', params.status);

  const query = searchParams.toString();
  const url = `/api/compliance/reports${query ? `?${query}` : ''}`;

  const response = await fetch(`${API_BASE}${url}`, {
    method: 'GET',
    headers: buildHeaders(),
    signal: options.signal,
  });

  return handleResponse(response);
};
