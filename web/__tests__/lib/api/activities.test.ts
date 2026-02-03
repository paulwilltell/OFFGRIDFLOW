/**
 * @jest-environment jsdom
 */

import { setupServer } from 'msw/node';
import { http, HttpResponse } from 'msw';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8090';
const server = setupServer();

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

// Types
interface Activity {
  id: string;
  name: string;
  type: string;
  value: number;
  unit: string;
  date: string;
  emissions?: number;
}

interface ActivityResponse {
  activities: Activity[];
  pagination?: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
  };
}

// Mock API client functions (these would normally be imported)
const getActivities = async (options: any = {}): Promise<ActivityResponse> => {
  const { page, limit, type, startDate, endDate, sortBy, sortOrder, signal, timeout, retry } = options;
  
  const params = new URLSearchParams();
  if (page) params.append('page', page.toString());
  if (limit) params.append('limit', limit.toString());
  if (type) params.append('type', type);
  if (startDate) params.append('startDate', startDate);
  if (endDate) params.append('endDate', endDate);
  if (sortBy) params.append('sortBy', sortBy);
  if (sortOrder) params.append('sortOrder', sortOrder);

  const token = typeof window !== 'undefined' ? localStorage.getItem('token') : null;
  
  try {
    const controller = signal || new AbortController();
    const timeoutId = timeout ? setTimeout(() => controller.abort(), timeout) : null;

    const response = await fetch(`${API_BASE}/api/v1/activities?${params}`, {
      headers: {
        ...(token && { Authorization: `Bearer ${token}` }),
      },
      signal: controller.signal,
    });

    if (timeoutId) clearTimeout(timeoutId);

    if (!response.ok) {
      if (response.status === 429 && retry) {
        await new Promise(resolve => setTimeout(resolve, 1000));
        return getActivities({ ...options, retry: false });
      }
      const error = await response.json();
      throw new Error(error.error || `HTTP ${response.status}`);
    }

    return response.json();
  } catch (error: any) {
    if (error.name === 'AbortError') {
      throw new Error('Request aborted');
    }
    throw error;
  }
};

const createActivity = async (activity: Omit<Activity, 'id' | 'emissions'>): Promise<{ activity: Activity }> => {
  const token = typeof window !== 'undefined' ? localStorage.getItem('token') : null;
  
  const response = await fetch(`${API_BASE}/api/v1/activities`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...(token && { Authorization: `Bearer ${token}` }),
    },
    body: JSON.stringify(activity),
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || `HTTP ${response.status}`);
  }

  return response.json();
};

const updateActivity = async (id: string, data: Partial<Activity>): Promise<{ activity: Activity }> => {
  const token = typeof window !== 'undefined' ? localStorage.getItem('token') : null;
  
  const response = await fetch(`${API_BASE}/api/v1/activities/${id}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
      ...(token && { Authorization: `Bearer ${token}` }),
    },
    body: JSON.stringify(data),
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || `HTTP ${response.status}`);
  }

  return response.json();
};

const deleteActivity = async (id: string): Promise<void> => {
  const token = typeof window !== 'undefined' ? localStorage.getItem('token') : null;
  
  const response = await fetch(`${API_BASE}/api/v1/activities/${id}`, {
    method: 'DELETE',
    headers: {
      ...(token && { Authorization: `Bearer ${token}` }),
    },
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || `HTTP ${response.status}`);
  }
};

describe('Activities API Client', () => {
  describe('getActivities', () => {
    it('should fetch activities successfully', async () => {
      const mockActivities: Activity[] = [
        {
          id: '1',
          name: 'Test Activity',
          type: 'electricity',
          value: 100,
          unit: 'kWh',
          date: '2025-01-01',
          emissions: 50,
        },
      ];

      server.use(
        http.get(`${API_BASE}/api/v1/activities`, async () => {
          return HttpResponse.json({ activities: mockActivities }, { status: 200 });
        })
      );

      const result = await getActivities();

      expect(result.activities).toEqual(mockActivities);
      expect(result.activities).toHaveLength(1);
      expect(result.activities[0].name).toBe('Test Activity');
    });

    it('should handle 401 Unauthorized', async () => {
      server.use(
        http.get(`${API_BASE}/api/v1/activities`, async () => {
          return HttpResponse.json({ error: 'Unauthorized' }, { status: 401 });
        })
      );

      await expect(getActivities()).rejects.toThrow('Unauthorized');
    });

    it('should handle network errors', async () => {
      server.use(
        http.get(`${API_BASE}/api/v1/activities`, async () => {
          return HttpResponse.error();
        })
      );

      await expect(getActivities()).rejects.toThrow();
    });

    it('should include auth token in request headers', async () => {
      let capturedHeaders: Headers | undefined;

      server.use(
        http.get(`${API_BASE}/api/v1/activities`, async ({ request }) => {
          capturedHeaders = request.headers;
          return HttpResponse.json({ activities: [] }, { status: 200 });
        })
      );

      // Mock localStorage
      Object.defineProperty(window, 'localStorage', {
        value: {
          getItem: jest.fn(() => 'test-jwt-token'),
          setItem: jest.fn(),
          removeItem: jest.fn(),
        },
        writable: true,
      });

      await getActivities();

      expect(capturedHeaders?.get('Authorization')).toBe('Bearer test-jwt-token');
    });

    it('should handle pagination', async () => {
      server.use(
        http.get(`${API_BASE}/api/v1/activities`, async ({ request }) => {
          const url = new URL(request.url);
          const page = url.searchParams.get('page') || '1';
          const limit = url.searchParams.get('limit') || '10';

          return HttpResponse.json(
            {
              activities: [],
              pagination: {
                page: parseInt(page),
                limit: parseInt(limit),
                total: 100,
                totalPages: 10,
              },
            },
            { status: 200 }
          );
        })
      );

      const result = await getActivities({ page: 2, limit: 20 });

      expect(result.pagination?.page).toBe(2);
      expect(result.pagination?.limit).toBe(20);
    });
  });

  describe('createActivity', () => {
    it('should create activity successfully', async () => {
      const newActivity = {
        name: 'New Activity',
        type: 'electricity',
        value: 150,
        unit: 'kWh',
        date: '2025-01-15',
      };

      const createdActivity: Activity = {
        ...newActivity,
        id: '2',
        emissions: 75,
      };

      server.use(
        http.post(`${API_BASE}/api/v1/activities`, async () => {
          return HttpResponse.json({ activity: createdActivity }, { status: 201 });
        })
      );

      const result = await createActivity(newActivity);

      expect(result.activity.id).toBe('2');
      expect(result.activity.name).toBe('New Activity');
    });

    it('should handle validation errors', async () => {
      server.use(
        http.post(`${API_BASE}/api/v1/activities`, async () => {
          return HttpResponse.json({ error: 'Validation failed' }, { status: 400 });
        })
      );

      const invalidActivity = {
        name: '',
        type: 'electricity',
        value: -10,
        unit: 'kWh',
        date: '2025-01-15',
      };

      await expect(createActivity(invalidActivity)).rejects.toThrow('Validation failed');
    });
  });

  describe('updateActivity', () => {
    it('should update activity successfully', async () => {
      const updatedActivity: Activity = {
        id: '1',
        name: 'Updated Activity',
        type: 'electricity',
        value: 200,
        unit: 'kWh',
        date: '2025-01-01',
        emissions: 100,
      };

      server.use(
        http.put(`${API_BASE}/api/v1/activities/1`, async () => {
          return HttpResponse.json({ activity: updatedActivity }, { status: 200 });
        })
      );

      const result = await updateActivity('1', { name: 'Updated Activity' });

      expect(result.activity.name).toBe('Updated Activity');
    });

    it('should handle 404 Not Found', async () => {
      server.use(
        http.put(`${API_BASE}/api/v1/activities/999`, async () => {
          return HttpResponse.json({ error: 'Activity not found' }, { status: 404 });
        })
      );

      await expect(updateActivity('999', { name: 'Test' })).rejects.toThrow('Activity not found');
    });
  });

  describe('deleteActivity', () => {
    it('should delete activity successfully', async () => {
      server.use(
        http.delete(`${API_BASE}/api/v1/activities/1`, async () => {
          return new HttpResponse(null, { status: 204 });
        })
      );

      await expect(deleteActivity('1')).resolves.not.toThrow();
    });

    it('should handle 403 Forbidden', async () => {
      server.use(
        http.delete(`${API_BASE}/api/v1/activities/1`, async () => {
          return HttpResponse.json({ error: 'Forbidden' }, { status: 403 });
        })
      );

      await expect(deleteActivity('1')).rejects.toThrow('Forbidden');
    });
  });
});
