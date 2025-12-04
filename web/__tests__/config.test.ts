import { config } from '@/lib/config';

describe('Config Module', () => {
  const originalEnv = process.env;

  beforeEach(() => {
    jest.resetModules();
    process.env = { ...originalEnv };
  });

  afterAll(() => {
    process.env = originalEnv;
  });

  describe('apiBaseUrl', () => {
    it('should have a default apiBaseUrl', () => {
      expect(config.apiBaseUrl).toBeDefined();
      expect(typeof config.apiBaseUrl).toBe('string');
    });

    it('should default to localhost:8090 when env var is not set', () => {
      // When NEXT_PUBLIC_OFFGRIDFLOW_API_URL is not set, it should use default
      expect(config.apiBaseUrl).toBe('http://localhost:8090');
    });
  });
});
