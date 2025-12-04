import { api, ApiRequestError, createClient, ACCESS_TOKEN_KEY } from '@/lib/api';

// Mock fetch globally
const mockFetch = jest.fn();
global.fetch = mockFetch;

// Mock localStorage
const localStorageMock = {
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
  clear: jest.fn(),
};
Object.defineProperty(window, 'localStorage', { value: localStorageMock });

describe('API Client', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    localStorageMock.getItem.mockReturnValue(null);
  });

  describe('createClient', () => {
    it('should create a client with default base URL', () => {
      const client = createClient();
      expect(client).toBeDefined();
      expect(client.get).toBeDefined();
      expect(client.post).toBeDefined();
      expect(client.put).toBeDefined();
      expect(client.patch).toBeDefined();
      expect(client.delete).toBeDefined();
    });

    it('should create a client with custom base URL', () => {
      const client = createClient('https://custom-api.example.com');
      expect(client).toBeDefined();
    });
  });

  describe('GET requests', () => {
    it('should make a GET request', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        text: () => Promise.resolve(JSON.stringify({ data: 'test' })),
      });

      const client = createClient('http://localhost:8090');
      const result = await client.get('/api/test');

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8090/api/test',
        expect.objectContaining({
          method: 'GET',
          credentials: 'include',
        })
      );
      expect(result).toEqual({ data: 'test' });
    });

    it('should include auth token if present', async () => {
      localStorageMock.getItem.mockImplementation((key: string) => {
        if (key === ACCESS_TOKEN_KEY) return 'test-token';
        return null;
      });

      mockFetch.mockResolvedValueOnce({
        ok: true,
        text: () => Promise.resolve('{}'),
      });

      const client = createClient('http://localhost:8090');
      await client.get('/api/test');

      expect(mockFetch).toHaveBeenCalledWith(
        expect.any(String),
        expect.objectContaining({
          headers: expect.objectContaining({
            Authorization: 'Bearer test-token',
          }),
        })
      );
    });
  });

  describe('POST requests', () => {
    it('should make a POST request with body', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        text: () => Promise.resolve(JSON.stringify({ success: true })),
      });

      const client = createClient('http://localhost:8090');
      const result = await client.post('/api/test', { foo: 'bar' });

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8090/api/test',
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify({ foo: 'bar' }),
        })
      );
      expect(result).toEqual({ success: true });
    });
  });

  describe('Error handling', () => {
    it('should throw ApiRequestError on non-ok response', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 401,
        json: () => Promise.resolve({ code: 'unauthorized', message: 'Not authenticated' }),
      });

      const client = createClient('http://localhost:8090');

      await expect(client.get('/api/protected')).rejects.toThrow(ApiRequestError);
    });

    it('should handle empty responses', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        text: () => Promise.resolve(''),
      });

      const client = createClient('http://localhost:8090');
      const result = await client.get('/api/empty');

      expect(result).toEqual({});
    });
  });
});

describe('ApiRequestError', () => {
  it('should create an error with correct properties', () => {
    const error = new ApiRequestError(404, {
      code: 'not_found',
      message: 'Resource not found',
    });

    expect(error.message).toBe('Resource not found');
    expect(error.code).toBe('not_found');
    expect(error.status).toBe(404);
  });
});
