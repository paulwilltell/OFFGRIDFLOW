/**
 * Activities API Client
 * 
 * Production-grade API client for managing emission activities
 * with full error handling, retry logic, and TypeScript types.
 */

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8090';

/**
 * Activity data structure
 */
export interface Activity {
  id: string;
  name: string;
  type: 'electricity' | 'natural_gas' | 'fuel' | 'waste' | 'travel';
  value: number;
  unit: string;
  date: string;
  emissions?: number;
  scope?: number;
  location?: string;
  notes?: string;
  createdAt?: string;
  updatedAt?: string;
}

/**
 * API response types
 */
export interface ActivityResponse {
  activities: Activity[];
  pagination?: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
  };
}

export interface SingleActivityResponse {
  activity: Activity;
}

/**
 * API error structure
 */
export class APIError extends Error {
  constructor(
    message: string,
    public status: number,
    public details?: any
  ) {
    super(message);
    this.name = 'APIError';
  }
}

/**
 * Request options
 */
export interface RequestOptions {
  signal?: AbortSignal;
  timeout?: number;
  retry?: boolean;
  cache?: boolean;
}

export interface GetActivitiesOptions extends RequestOptions {
  page?: number;
  limit?: number;
  type?: string;
  startDate?: string;
  endDate?: string;
  sortBy?: 'date' | 'name' | 'emissions' | 'value';
  sortOrder?: 'asc' | 'desc';
}

/**
 * Get authentication token from storage
 */
const getAuthToken = (): string | null => {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem('auth_token');
};

/**
 * Build request headers with authentication
 */
const buildHeaders = (includeAuth = true): HeadersInit => {
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
  };

  if (includeAuth) {
    const token = getAuthToken();
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }
  }

  return headers;
};

/**
 * Handle API response and errors
 */
const handleResponse = async <T>(response: Response): Promise<T> => {
  if (!response.ok) {
    const error = await response.json().catch(() => ({
      error: `HTTP ${response.status}: ${response.statusText}`,
    }));

    throw new APIError(
      error.error || error.message || `Request failed with status ${response.status}`,
      response.status,
      error.details
    );
  }

  // Handle 204 No Content
  if (response.status === 204) {
    return {} as T;
  }

  return response.json();
};

/**
 * Make API request with retry logic
 */
const makeRequest = async <T>(
  endpoint: string,
  options: RequestInit & RequestOptions = {}
): Promise<T> => {
  const { retry = false, timeout, signal, ...fetchOptions } = options;

  // Setup timeout
  const controller = new AbortController();
  const finalSignal = signal || controller.signal;
  
  let timeoutId: NodeJS.Timeout | undefined;
  if (timeout) {
    timeoutId = setTimeout(() => controller.abort(), timeout);
  }

  try {
    const response = await fetch(`${API_BASE}${endpoint}`, {
      ...fetchOptions,
      signal: finalSignal,
    });

    if (timeoutId) clearTimeout(timeoutId);

    // Handle rate limiting with retry
    if (response.status === 429 && retry) {
      await new Promise((resolve) => setTimeout(resolve, 1000));
      return makeRequest<T>(endpoint, { ...options, retry: false });
    }

    return handleResponse<T>(response);
  } catch (error: any) {
    if (timeoutId) clearTimeout(timeoutId);

    if (error.name === 'AbortError') {
      throw new APIError('Request aborted', 0);
    }

    throw error;
  }
};

/**
 * Build query string from parameters
 */
const buildQueryString = (params: Record<string, any>): string => {
  const searchParams = new URLSearchParams();

  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== null) {
      searchParams.append(key, value.toString());
    }
  });

  const query = searchParams.toString();
  return query ? `?${query}` : '';
};

/**
 * Get activities with optional filtering and pagination
 */
export const getActivities = async (
  options: GetActivitiesOptions = {}
): Promise<ActivityResponse> => {
  const {
    page,
    limit,
    type,
    startDate,
    endDate,
    sortBy,
    sortOrder,
    signal,
    timeout,
    retry,
  } = options;

  const queryString = buildQueryString({
    page,
    limit,
    type,
    startDate,
    endDate,
    sortBy,
    sortOrder,
  });

  return makeRequest<ActivityResponse>(`/api/v1/activities${queryString}`, {
    method: 'GET',
    headers: buildHeaders(),
    signal,
    timeout,
    retry,
  });
};

/**
 * Get single activity by ID
 */
export const getActivity = async (
  id: string,
  options: RequestOptions = {}
): Promise<SingleActivityResponse> => {
  return makeRequest<SingleActivityResponse>(`/api/v1/activities/${id}`, {
    method: 'GET',
    headers: buildHeaders(),
    ...options,
  });
};

/**
 * Create new activity
 */
export const createActivity = async (
  activity: Omit<Activity, 'id' | 'emissions' | 'createdAt' | 'updatedAt'>,
  options: RequestOptions = {}
): Promise<SingleActivityResponse> => {
  return makeRequest<SingleActivityResponse>('/api/v1/activities', {
    method: 'POST',
    headers: buildHeaders(),
    body: JSON.stringify(activity),
    ...options,
  });
};

/**
 * Update existing activity
 */
export const updateActivity = async (
  id: string,
  updates: Partial<Activity>,
  options: RequestOptions = {}
): Promise<SingleActivityResponse> => {
  return makeRequest<SingleActivityResponse>(`/api/v1/activities/${id}`, {
    method: 'PUT',
    headers: buildHeaders(),
    body: JSON.stringify(updates),
    ...options,
  });
};

/**
 * Delete activity
 */
export const deleteActivity = async (
  id: string,
  options: RequestOptions = {}
): Promise<void> => {
  return makeRequest<void>(`/api/v1/activities/${id}`, {
    method: 'DELETE',
    headers: buildHeaders(),
    ...options,
  });
};

/**
 * Bulk create activities
 */
export const bulkCreateActivities = async (
  activities: Array<Omit<Activity, 'id' | 'emissions' | 'createdAt' | 'updatedAt'>>,
  options: RequestOptions = {}
): Promise<{ activities: Activity[]; errors?: any[] }> => {
  return makeRequest('/api/v1/activities/bulk', {
    method: 'POST',
    headers: buildHeaders(),
    body: JSON.stringify({ activities }),
    ...options,
  });
};

/**
 * Get activities summary (aggregated data)
 */
export const getActivitiesSummary = async (
  params: {
    startDate?: string;
    endDate?: string;
    groupBy?: 'type' | 'month' | 'scope';
  } = {},
  options: RequestOptions = {}
): Promise<{
  totalEmissions: number;
  totalActivities: number;
  breakdown: Record<string, number>;
}> => {
  const queryString = buildQueryString(params);

  return makeRequest(`/api/v1/activities/summary${queryString}`, {
    method: 'GET',
    headers: buildHeaders(),
    ...options,
  });
};

/**
 * Export activities to CSV
 */
export const exportActivities = async (
  format: 'csv' | 'json' | 'xlsx',
  filters: Omit<GetActivitiesOptions, keyof RequestOptions> = {},
  options: RequestOptions = {}
): Promise<Blob> => {
  const queryString = buildQueryString({ ...filters, format });

  const response = await fetch(
    `${API_BASE}/api/v1/activities/export${queryString}`,
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
