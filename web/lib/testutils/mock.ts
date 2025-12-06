/**
 * Mock HTTP utilities for testing API clients
 * @packageDocumentation
 */

import { rest } from 'msw';
import { setupServer } from 'msw/node';

/**
 * Mock response configuration
 */
export interface MockResponse<T = any> {
  status?: number;
  data?: T;
  error?: string;
  delay?: number;
}

/**
 * Creates a mock HTTP response
 */
export const createMockResponse = <T = any>(config: MockResponse<T> = {}) => {
  const { status = 200, data, error, delay = 0 } = config;

  return {
    status,
    body: error ? { error } : data,
    delay,
  };
};

/**
 * Mock server setup for tests
 */
export const createMockServer = () => {
  return setupServer();
};

/**
 * Creates a successful mock response
 */
export const mockSuccess = <T = any>(data: T, status = 200): MockResponse<T> => ({
  status,
  data,
});

/**
 * Creates an error mock response
 */
export const mockError = (message: string, status = 500): MockResponse => ({
  status,
  error: message,
});

/**
 * Creates a 401 Unauthorized mock response
 */
export const mockUnauthorized = (): MockResponse => ({
  status: 401,
  error: 'Unauthorized',
});

/**
 * Creates a 404 Not Found mock response
 */
export const mockNotFound = (): MockResponse => ({
  status: 404,
  error: 'Not found',
});

/**
 * Creates a 429 Too Many Requests mock response
 */
export const mockRateLimited = (): MockResponse => ({
  status: 429,
  error: 'Rate limit exceeded',
});

/**
 * Creates a mock handler for GET requests
 */
export const mockGet = <T = any>(path: string, response: MockResponse<T>) => {
  return rest.get(path, (req, res, ctx) => {
    const { status = 200, data, error, delay = 0 } = response;

    return res(
      ctx.delay(delay),
      ctx.status(status),
      ctx.json(error ? { error } : data)
    );
  });
};

/**
 * Creates a mock handler for POST requests
 */
export const mockPost = <T = any>(path: string, response: MockResponse<T>) => {
  return rest.post(path, (req, res, ctx) => {
    const { status = 200, data, error, delay = 0 } = response;

    return res(
      ctx.delay(delay),
      ctx.status(status),
      ctx.json(error ? { error } : data)
    );
  });
};

/**
 * Creates a mock handler for PUT requests
 */
export const mockPut = <T = any>(path: string, response: MockResponse<T>) => {
  return rest.put(path, (req, res, ctx) => {
    const { status = 200, data, error, delay = 0 } = response;

    return res(
      ctx.delay(delay),
      ctx.status(status),
      ctx.json(error ? { error } : data)
    );
  });
};

/**
 * Creates a mock handler for DELETE requests
 */
export const mockDelete = <T = any>(path: string, response: MockResponse<T>) => {
  return rest.delete(path, (req, res, ctx) => {
    const { status = 200, data, error, delay = 0 } = response;

    return res(
      ctx.delay(delay),
      ctx.status(status),
      ctx.json(error ? { error } : data)
    );
  });
};

/**
 * Type-safe test assertion helpers
 */
export const assertStatus = (actual: number, expected: number) => {
  if (actual !== expected) {
    throw new Error(`Expected status ${expected} but got ${actual}`);
  }
};

export const assertError = (error: any) => {
  if (!error) {
    throw new Error('Expected error but got none');
  }
};

export const assertNoError = (error: any) => {
  if (error) {
    throw new Error(`Unexpected error: ${error.message}`);
  }
};

export const assertDefined = <T>(value: T | undefined | null): asserts value is T => {
  if (value === undefined || value === null) {
    throw new Error('Expected value to be defined');
  }
};
