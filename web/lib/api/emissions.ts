/**
 * Emissions API Client
 * 
 * Handles Scope 1, 2, and 3 emissions calculations and reporting
 */

import { APIError, RequestOptions } from './activities';
import { config } from '../config';

const API_BASE = config.apiBaseUrl;

/**
 * Emissions data types
 */
export interface EmissionsSummary {
  scope1: number;
  scope2: number;
  scope3: number;
  total: number;
  period: {
    startDate: string;
    endDate: string;
  };
  breakdown: {
    electricity?: number;
    naturalGas?: number;
    fuel?: number;
    waste?: number;
    travel?: number;
  };
}

export interface Scope2Summary {
  total_emissions: number;
  total_activities: number;
  by_location?: Record<string, number>;
  by_month?: Record<string, number>;
}

/**
 * Build headers
 */
const buildHeaders = (): HeadersInit => {
  return {
    'Content-Type': 'application/json',
  };
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
      error.error || `Request failed`,
      response.status,
      error.details
    );
  }

  return response.json();
};

/**
 * Get Scope 2 emissions summary
 */
export const getScope2Summary = async (
  params: {
    startDate?: string;
    endDate?: string;
  } = {},
  options: RequestOptions = {}
): Promise<{ summary: Scope2Summary }> => {
  const searchParams = new URLSearchParams();
  if (params.startDate) searchParams.append('startDate', params.startDate);
  if (params.endDate) searchParams.append('endDate', params.endDate);

  const query = searchParams.toString();
  const url = `/api/emissions/scope2/summary${query ? `?${query}` : ''}`;

  const response = await fetch(`${API_BASE}${url}`, {
    method: 'GET',
    headers: buildHeaders(),
    credentials: 'include',
    signal: options.signal,
  });

  return handleResponse(response);
};

/**
 * Get complete emissions summary (all scopes)
 */
export const getEmissionsSummary = async (
  params: {
    startDate?: string;
    endDate?: string;
  } = {},
  options: RequestOptions = {}
): Promise<EmissionsSummary> => {
  const searchParams = new URLSearchParams();
  if (params.startDate) searchParams.append('startDate', params.startDate);
  if (params.endDate) searchParams.append('endDate', params.endDate);

  const query = searchParams.toString();
  const url = `/api/emissions/summary${query ? `?${query}` : ''}`;

  const response = await fetch(`${API_BASE}${url}`, {
    method: 'GET',
    headers: buildHeaders(),
    credentials: 'include',
    signal: options.signal,
  });

  return handleResponse(response);
};

/**
 * Calculate emissions for a specific activity
 */
export const calculateEmissions = async (
  activity: {
    type: string;
    value: number;
    unit: string;
    location?: string;
  },
  options: RequestOptions = {}
): Promise<{ emissions: number; scope: number }> => {
  const response = await fetch(`${API_BASE}/api/emissions/calculate`, {
    method: 'POST',
    headers: buildHeaders(),
    credentials: 'include',
    body: JSON.stringify(activity),
    signal: options.signal,
  });

  return handleResponse(response);
};

/**
 * Get emissions trends over time
 */
export const getEmissionsTrends = async (
  params: {
    startDate: string;
    endDate: string;
    granularity: 'day' | 'week' | 'month' | 'year';
  },
  options: RequestOptions = {}
): Promise<{
  trends: Array<{
    date: string;
    scope1: number;
    scope2: number;
    scope3: number;
    total: number;
  }>;
}> => {
  const searchParams = new URLSearchParams({
    startDate: params.startDate,
    endDate: params.endDate,
    granularity: params.granularity,
  });

  const response = await fetch(
    `${API_BASE}/api/emissions/trends?${searchParams}`,
    {
      method: 'GET',
      headers: buildHeaders(),
      credentials: 'include',
      signal: options.signal,
    }
  );

  return handleResponse(response);
};

/**
 * Get emissions by category
 */
export const getEmissionsByCategory = async (
  params: {
    startDate?: string;
    endDate?: string;
  } = {},
  options: RequestOptions = {}
): Promise<{
  categories: Array<{
    name: string;
    emissions: number;
    percentage: number;
  }>;
}> => {
  const searchParams = new URLSearchParams();
  if (params.startDate) searchParams.append('startDate', params.startDate);
  if (params.endDate) searchParams.append('endDate', params.endDate);

  const query = searchParams.toString();
  const url = `/api/emissions/by-category${query ? `?${query}` : ''}`;

  const response = await fetch(`${API_BASE}${url}`, {
    method: 'GET',
    headers: buildHeaders(),
    credentials: 'include',
    signal: options.signal,
  });

  return handleResponse(response);
};
