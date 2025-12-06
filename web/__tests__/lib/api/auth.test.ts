/**
 * @jest-environment jsdom
 */

import { setupServer } from 'msw/node';
import { rest } from 'msw';
import {
  register,
  login,
  logout,
  getCurrentUser,
  changePassword,
  forgotPassword,
  resetPassword,
  createAPIKey,
  listAPIKeys,
  revokeAPIKey,
  isAuthenticated,
  setAuthToken,
  clearAuthToken,
  getAuthToken,
} from '@/lib/api/auth';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8090';
const server = setupServer();

beforeAll(() => server.listen());
afterEach(() => {
  server.resetHandlers();
  clearAuthToken();
});
afterAll(() => server.close());

describe('Auth API Client', () => {
  describe('register', () => {
    it('should register user and save token', async () => {
      const mockUser = {
        id: '1',
        email: 'test@example.com',
        name: 'Test User',
        role: 'user',
        tenantId: 'tenant-1',
        createdAt: '2025-01-01T00:00:00Z',
      };

      const mockToken = 'jwt-token-123';

      server.use(
        rest.post(`${API_BASE}/api/auth/register`, (req, res, ctx) => {
          return res(
            ctx.status(201),
            ctx.json({
              user: mockUser,
              token: mockToken,
              expiresAt: '2025-01-08T00:00:00Z',
            })
          );
        })
      );

      const result = await register({
        email: 'test@example.com',
        password: 'SecurePass123!',
        name: 'Test User',
      });

      expect(result.user.email).toBe('test@example.com');
      expect(result.token).toBe(mockToken);
      expect(getAuthToken()).toBe(mockToken);
    });

    it('should handle registration errors', async () => {
      server.use(
        rest.post(`${API_BASE}/api/auth/register`, (req, res, ctx) => {
          return res(
            ctx.status(400),
            ctx.json({ error: 'Email already exists' })
          );
        })
      );

      await expect(
        register({
          email: 'existing@example.com',
          password: 'pass',
          name: 'Test',
        })
      ).rejects.toThrow('Email already exists');
    });
  });

  describe('login', () => {
    it('should login user and save token', async () => {
      const mockToken = 'jwt-login-token';

      server.use(
        rest.post(`${API_BASE}/api/auth/login`, (req, res, ctx) => {
          return res(
            ctx.status(200),
            ctx.json({
              user: {
                id: '1',
                email: 'test@example.com',
                name: 'Test User',
                role: 'user',
                tenantId: 'tenant-1',
                createdAt: '2025-01-01T00:00:00Z',
              },
              token: mockToken,
              expiresAt: '2025-01-08T00:00:00Z',
            })
          );
        })
      );

      const result = await login({
        email: 'test@example.com',
        password: 'SecurePass123!',
      });

      expect(result.token).toBe(mockToken);
      expect(getAuthToken()).toBe(mockToken);
    });

    it('should handle invalid credentials', async () => {
      server.use(
        rest.post(`${API_BASE}/api/auth/login`, (req, res, ctx) => {
          return res(
            ctx.status(401),
            ctx.json({ error: 'Invalid credentials' })
          );
        })
      );

      await expect(
        login({
          email: 'wrong@example.com',
          password: 'wrong',
        })
      ).rejects.toThrow('Invalid credentials');
    });
  });

  describe('logout', () => {
    it('should logout and clear token', async () => {
      setAuthToken('some-token');

      server.use(
        rest.post(`${API_BASE}/api/auth/logout`, (req, res, ctx) => {
          return res(ctx.status(200));
        })
      );

      await logout();

      expect(getAuthToken()).toBeNull();
    });

    it('should clear token even if request fails', async () => {
      setAuthToken('some-token');

      server.use(
        rest.post(`${API_BASE}/api/auth/logout`, (req, res, ctx) => {
          return res(ctx.status(500));
        })
      );

      await logout();

      expect(getAuthToken()).toBeNull();
    });
  });

  describe('getCurrentUser', () => {
    it('should get current user info', async () => {
      setAuthToken('valid-token');

      const mockUser = {
        id: '1',
        email: 'test@example.com',
        name: 'Test User',
        role: 'user',
        tenantId: 'tenant-1',
        createdAt: '2025-01-01T00:00:00Z',
      };

      server.use(
        rest.get(`${API_BASE}/api/auth/me`, (req, res, ctx) => {
          const auth = req.headers.get('Authorization');
          if (auth !== 'Bearer valid-token') {
            return res(ctx.status(401), ctx.json({ error: 'Unauthorized' }));
          }

          return res(ctx.status(200), ctx.json({ user: mockUser }));
        })
      );

      const result = await getCurrentUser();

      expect(result.user.email).toBe('test@example.com');
    });

    it('should handle unauthorized', async () => {
      setAuthToken('invalid-token');

      server.use(
        rest.get(`${API_BASE}/api/auth/me`, (req, res, ctx) => {
          return res(ctx.status(401), ctx.json({ error: 'Unauthorized' }));
        })
      );

      await expect(getCurrentUser()).rejects.toThrow('Unauthorized');
    });
  });

  describe('changePassword', () => {
    it('should change password successfully', async () => {
      setAuthToken('valid-token');

      server.use(
        rest.post(`${API_BASE}/api/auth/change-password`, (req, res, ctx) => {
          return res(
            ctx.status(200),
            ctx.json({ message: 'Password changed successfully' })
          );
        })
      );

      const result = await changePassword({
        currentPassword: 'OldPass123!',
        newPassword: 'NewPass456!',
      });

      expect(result.message).toBe('Password changed successfully');
    });

    it('should handle incorrect current password', async () => {
      setAuthToken('valid-token');

      server.use(
        rest.post(`${API_BASE}/api/auth/change-password`, (req, res, ctx) => {
          return res(
            ctx.status(400),
            ctx.json({ error: 'Current password is incorrect' })
          );
        })
      );

      await expect(
        changePassword({
          currentPassword: 'wrong',
          newPassword: 'NewPass456!',
        })
      ).rejects.toThrow('Current password is incorrect');
    });
  });

  describe('forgotPassword', () => {
    it('should send password reset email', async () => {
      server.use(
        rest.post(`${API_BASE}/api/auth/password/forgot`, (req, res, ctx) => {
          return res(
            ctx.status(200),
            ctx.json({ message: 'Password reset email sent' })
          );
        })
      );

      const result = await forgotPassword('test@example.com');

      expect(result.message).toBe('Password reset email sent');
    });

    it('should handle non-existent email gracefully', async () => {
      server.use(
        rest.post(`${API_BASE}/api/auth/password/forgot`, (req, res, ctx) => {
          // Security: return success even if email doesn't exist
          return res(
            ctx.status(200),
            ctx.json({ message: 'Password reset email sent' })
          );
        })
      );

      const result = await forgotPassword('nonexistent@example.com');

      expect(result.message).toBe('Password reset email sent');
    });
  });

  describe('resetPassword', () => {
    it('should reset password with valid token', async () => {
      server.use(
        rest.post(`${API_BASE}/api/auth/password/reset`, (req, res, ctx) => {
          return res(
            ctx.status(200),
            ctx.json({ message: 'Password reset successfully' })
          );
        })
      );

      const result = await resetPassword({
        token: 'reset-token-123',
        newPassword: 'NewSecurePass456!',
      });

      expect(result.message).toBe('Password reset successfully');
    });

    it('should handle invalid reset token', async () => {
      server.use(
        rest.post(`${API_BASE}/api/auth/password/reset`, (req, res, ctx) => {
          return res(
            ctx.status(400),
            ctx.json({ error: 'Invalid or expired reset token' })
          );
        })
      );

      await expect(
        resetPassword({
          token: 'invalid-token',
          newPassword: 'NewPass',
        })
      ).rejects.toThrow('Invalid or expired reset token');
    });
  });

  describe('API Keys', () => {
    beforeEach(() => {
      setAuthToken('valid-token');
    });

    it('should create API key', async () => {
      const mockKey = {
        id: 'key-1',
        name: 'Production API Key',
        key: 'ogf_prod_abc123def456',
        scopes: ['read:emissions', 'write:emissions'],
        createdAt: '2025-01-01T00:00:00Z',
      };

      server.use(
        rest.post(`${API_BASE}/api/auth/keys`, (req, res, ctx) => {
          return res(ctx.status(201), ctx.json({ key: mockKey }));
        })
      );

      const result = await createAPIKey({
        name: 'Production API Key',
        scopes: ['read:emissions', 'write:emissions'],
      });

      expect(result.key.name).toBe('Production API Key');
      expect(result.key.key).toContain('ogf_prod_');
    });

    it('should list API keys', async () => {
      const mockKeys = [
        {
          id: 'key-1',
          name: 'Key 1',
          keyHash: 'hash1',
          scopes: ['read:emissions'],
          createdAt: '2025-01-01T00:00:00Z',
        },
        {
          id: 'key-2',
          name: 'Key 2',
          keyHash: 'hash2',
          scopes: ['read:emissions', 'write:emissions'],
          createdAt: '2025-01-02T00:00:00Z',
        },
      ];

      server.use(
        rest.get(`${API_BASE}/api/auth/keys`, (req, res, ctx) => {
          return res(ctx.status(200), ctx.json({ keys: mockKeys }));
        })
      );

      const result = await listAPIKeys();

      expect(result.keys).toHaveLength(2);
      expect(result.keys[0].name).toBe('Key 1');
    });

    it('should revoke API key', async () => {
      server.use(
        rest.delete(`${API_BASE}/api/auth/keys/key-1`, (req, res, ctx) => {
          return res(
            ctx.status(200),
            ctx.json({ message: 'API key revoked' })
          );
        })
      );

      const result = await revokeAPIKey('key-1');

      expect(result.message).toBe('API key revoked');
    });

    it('should handle revoking non-existent key', async () => {
      server.use(
        rest.delete(`${API_BASE}/api/auth/keys/invalid`, (req, res, ctx) => {
          return res(ctx.status(404), ctx.json({ error: 'Key not found' }));
        })
      );

      await expect(revokeAPIKey('invalid')).rejects.toThrow('Key not found');
    });
  });

  describe('utility functions', () => {
    it('should check authentication status', () => {
      expect(isAuthenticated()).toBe(false);

      setAuthToken('some-token');
      expect(isAuthenticated()).toBe(true);

      clearAuthToken();
      expect(isAuthenticated()).toBe(false);
    });

    it('should manage token storage', () => {
      const token = 'test-token-123';

      setAuthToken(token);
      expect(getAuthToken()).toBe(token);

      clearAuthToken();
      expect(getAuthToken()).toBeNull();
    });
  });
});
