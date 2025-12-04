import {
  login,
  logout,
  register,
  getCurrentUser,
  changePassword,
  getSession,
  getToken,
  isAuthenticated,
  hasRole,
  useAuth,
} from '@/lib/auth';
import { api } from '@/lib/api';

// Mock the API module
jest.mock('@/lib/api', () => ({
  api: {
    get: jest.fn(),
    post: jest.fn(),
  },
  ACCESS_TOKEN_KEY: 'offgridflow_access_token',
  REFRESH_TOKEN_KEY: 'offgridflow_refresh_token',
  TENANT_ID_KEY: 'offgridflow_tenant_id',
}));

const mockedApi = api as jest.Mocked<typeof api>;

// Mock localStorage
const localStorageMock = (() => {
  let store: Record<string, string> = {};
  return {
    getItem: jest.fn((key: string) => store[key] || null),
    setItem: jest.fn((key: string, value: string) => {
      store[key] = value;
    }),
    removeItem: jest.fn((key: string) => {
      delete store[key];
    }),
    clear: jest.fn(() => {
      store = {};
    }),
  };
})();
Object.defineProperty(window, 'localStorage', { value: localStorageMock });

describe('Auth Module', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    localStorageMock.clear();
  });

  describe('register', () => {
    it('should register a new user and store session', async () => {
      const mockResponse = {
        token: 'new-token',
        user: {
          id: '1',
          email: 'new@example.com',
          name: 'New User',
          role: 'user',
        },
      };
      mockedApi.post.mockResolvedValueOnce(mockResponse);

      const result = await register({
        email: 'new@example.com',
        password: 'password123',
        name: 'New User',
      });

      expect(mockedApi.post).toHaveBeenCalledWith('/api/auth/register', {
        email: 'new@example.com',
        password: 'password123',
        name: 'New User',
      });
      expect(result).toEqual(mockResponse);
      expect(localStorageMock.setItem).toHaveBeenCalledWith(
        'offgridflow_access_token',
        'new-token'
      );
    });
  });

  describe('login', () => {
    it('should login and store session', async () => {
      const mockResponse = {
        token: 'login-token',
        user: {
          id: '1',
          email: 'user@example.com',
          name: 'User',
          role: 'user',
        },
      };
      mockedApi.post.mockResolvedValueOnce(mockResponse);

      const result = await login({
        email: 'user@example.com',
        password: 'password123',
      });

      expect(mockedApi.post).toHaveBeenCalledWith('/api/auth/login', {
        email: 'user@example.com',
        password: 'password123',
      });
      expect(result).toEqual(mockResponse);
    });

  });

  describe('logout', () => {
    it('should call logout API and clear session', async () => {
      localStorageMock.setItem('offgridflow_access_token', 'existing-token');
      mockedApi.post.mockResolvedValueOnce({});

      await logout();

      expect(mockedApi.post).toHaveBeenCalledWith('/api/auth/logout', {});
      expect(localStorageMock.removeItem).toHaveBeenCalledWith('offgridflow_access_token');
    });

    it('should clear session even if API call fails', async () => {
      localStorageMock.setItem('offgridflow_access_token', 'existing-token');
      mockedApi.post.mockRejectedValueOnce(new Error('Network error'));

      await logout();

      expect(localStorageMock.removeItem).toHaveBeenCalledWith('offgridflow_access_token');
    });
  });

  describe('getCurrentUser', () => {
    it('should fetch current user', async () => {
      const mockUser = {
        token: 'token',
        user: {
          id: '1',
          email: 'user@example.com',
          name: 'User',
          role: 'admin',
        },
      };
      mockedApi.get.mockResolvedValueOnce(mockUser);

      const result = await getCurrentUser();

      expect(mockedApi.get).toHaveBeenCalledWith('/api/auth/me');
      expect(result).toEqual(mockUser);
    });

    it('should return null on error', async () => {
      mockedApi.get.mockRejectedValueOnce(new Error('Unauthorized'));

      const result = await getCurrentUser();

      expect(result).toBeNull();
    });
  });

  describe('changePassword', () => {
    it('should call change password API', async () => {
      mockedApi.post.mockResolvedValueOnce({});

      await changePassword('oldPassword', 'newPassword');

      expect(mockedApi.post).toHaveBeenCalledWith('/api/auth/change-password', {
        current_password: 'oldPassword',
        new_password: 'newPassword',
      });
    });
  });

  describe('getSession', () => {
    it('should return null when no session stored', () => {
      const result = getSession();
      expect(result).toBeNull();
    });

    it('should return parsed session when stored', () => {
      const session = {
        user: { id: '1', email: 'test@example.com', name: 'Test', role: 'user' },
        tenant: null,
        token: 'token',
      };
      localStorageMock.getItem.mockReturnValueOnce(JSON.stringify(session));

      const result = getSession();

      expect(result).toEqual(session);
    });

    it('should return null for invalid JSON', () => {
      localStorageMock.getItem.mockReturnValueOnce('invalid-json');

      const result = getSession();

      expect(result).toBeNull();
    });
  });

  describe('getToken', () => {
    it('should return token from localStorage', () => {
      localStorageMock.getItem.mockReturnValueOnce('stored-token');

      const result = getToken();

      expect(result).toBe('stored-token');
    });

    it('should return null when no token', () => {
      localStorageMock.getItem.mockReturnValueOnce(null);

      const result = getToken();

      expect(result).toBeNull();
    });
  });

  describe('isAuthenticated', () => {
    it('should return true when session exists', () => {
      const session = {
        user: { id: '1', email: 'test@example.com', name: 'Test', role: 'user' },
        tenant: null,
      };
      localStorageMock.getItem.mockReturnValueOnce(JSON.stringify(session));

      expect(isAuthenticated()).toBe(true);
    });

    it('should return false when no session', () => {
      localStorageMock.getItem.mockReturnValueOnce(null);

      expect(isAuthenticated()).toBe(false);
    });
  });

  describe('hasRole', () => {
    it('should return true for matching role', () => {
      const session = {
        user: { id: '1', email: 'test@example.com', name: 'Test', role: 'admin' },
        tenant: null,
      };
      localStorageMock.getItem.mockReturnValueOnce(JSON.stringify(session));

      expect(hasRole('admin')).toBe(true);
    });

    it('should return false for non-matching role', () => {
      const session = {
        user: { id: '1', email: 'test@example.com', name: 'Test', role: 'user' },
        tenant: null,
      };
      localStorageMock.getItem.mockReturnValueOnce(JSON.stringify(session));

      expect(hasRole('admin')).toBe(false);
    });

    it('should return false when no session', () => {
      localStorageMock.getItem.mockReturnValueOnce(null);

      expect(hasRole('admin')).toBe(false);
    });
  });

  describe('useAuth hook', () => {
    it('should return auth state and methods', () => {
      localStorageMock.getItem.mockReturnValueOnce(null);

      const result = useAuth();

      expect(result).toHaveProperty('isAuthenticated');
      expect(result).toHaveProperty('user');
      expect(result).toHaveProperty('tenant');
      expect(result).toHaveProperty('login');
      expect(result).toHaveProperty('logout');
      expect(result).toHaveProperty('register');
    });
  });
});
