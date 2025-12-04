import {
  getPlans,
  getSubscription,
  formatSubscriptionStatus,
  formatPeriodEnd,
  formatPrice,
  hasActiveSubscription,
} from '@/lib/billing';
import { api, ApiRequestError } from '@/lib/api';

// Mock the API module
jest.mock('@/lib/api', () => ({
  api: {
    get: jest.fn(),
    post: jest.fn(),
  },
  ApiRequestError: class ApiRequestError extends Error {
    code: string;
    status: number;
    constructor(status: number, error: { code: string; message: string }) {
      super(error.message);
      this.code = error.code;
      this.status = status;
    }
  },
}));

const mockedApi = api as jest.Mocked<typeof api>;

describe('Billing Module', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('getPlans', () => {
    it('should fetch billing plans', async () => {
      const mockPlans = {
        plans: [
          { id: 'pro', price_id: 'price_123', name: 'Pro', amount_cents: 9900, interval: 'month' as const, features: [] },
        ],
      };
      mockedApi.get.mockResolvedValueOnce(mockPlans);

      const result = await getPlans();

      expect(mockedApi.get).toHaveBeenCalledWith('/api/billing/plans');
      expect(result).toEqual(mockPlans);
    });
  });

  describe('getSubscription', () => {
    it('should fetch subscription status', async () => {
      const mockSubscription = {
        plan_id: 'pro',
        status: 'active',
        current_period_end: '2024-12-31T00:00:00Z',
      };
      mockedApi.get.mockResolvedValueOnce(mockSubscription);

      const result = await getSubscription();

      expect(mockedApi.get).toHaveBeenCalledWith('/api/billing/subscription');
      expect(result).toEqual(mockSubscription);
    });
  });

  describe('hasActiveSubscription', () => {
    it('should return true for active subscription', async () => {
      mockedApi.get.mockResolvedValueOnce({ status: 'active' });

      const result = await hasActiveSubscription();

      expect(result).toBe(true);
    });

    it('should return true for trialing subscription', async () => {
      mockedApi.get.mockResolvedValueOnce({ status: 'trialing' });

      const result = await hasActiveSubscription();

      expect(result).toBe(true);
    });

    it('should return false for canceled subscription', async () => {
      mockedApi.get.mockResolvedValueOnce({ status: 'canceled' });

      const result = await hasActiveSubscription();

      expect(result).toBe(false);
    });

    it('should return false on 401 error', async () => {
      const error = new (ApiRequestError as any)(401, { code: 'unauthorized', message: 'Not authenticated' });
      mockedApi.get.mockRejectedValueOnce(error);

      const result = await hasActiveSubscription();

      expect(result).toBe(false);
    });

    it('should throw on other errors', async () => {
      const error = new Error('Network error');
      mockedApi.get.mockRejectedValueOnce(error);

      await expect(hasActiveSubscription()).rejects.toThrow('Network error');
    });
  });

  describe('formatSubscriptionStatus', () => {
    it('should format active status', () => {
      expect(formatSubscriptionStatus('active')).toBe('Active');
    });

    it('should format trialing status', () => {
      expect(formatSubscriptionStatus('trialing')).toBe('Trial');
    });

    it('should format past_due status', () => {
      expect(formatSubscriptionStatus('past_due')).toBe('Past Due');
    });

    it('should format canceled status', () => {
      expect(formatSubscriptionStatus('canceled')).toBe('Canceled');
    });

    it('should format unpaid status', () => {
      expect(formatSubscriptionStatus('unpaid')).toBe('Unpaid');
    });

    it('should return None for null/undefined', () => {
      expect(formatSubscriptionStatus(null)).toBe('None');
      expect(formatSubscriptionStatus(undefined)).toBe('None');
    });

    it('should return original value for unknown status', () => {
      expect(formatSubscriptionStatus('unknown_status')).toBe('unknown_status');
    });
  });

  describe('formatPeriodEnd', () => {
    it('should format a valid date string', () => {
      const result = formatPeriodEnd('2024-12-31T00:00:00Z');
      expect(result).toContain('December');
      expect(result).toContain('31');
      expect(result).toContain('2024');
    });

    it('should return N/A for null/undefined', () => {
      expect(formatPeriodEnd(null)).toBe('N/A');
      expect(formatPeriodEnd(undefined)).toBe('N/A');
    });
  });

  describe('formatPrice', () => {
    it('should format monthly price', () => {
      expect(formatPrice(9900, 'month')).toBe('$99.00/month');
    });

    it('should format yearly price', () => {
      expect(formatPrice(99900, 'year')).toBe('$999.00/year');
    });

    it('should handle zero price', () => {
      expect(formatPrice(0, 'month')).toBe('$0.00/month');
    });
  });
});
