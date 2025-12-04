import { api, ApiRequestError } from './api';

export interface BillingPlan {
  id: string;
  price_id: string;
  name: string;
  amount_cents: number;
  interval: 'month' | 'year';
  features: string[];
}

export interface BillingPlansResponse {
  plans: BillingPlan[];
}

export interface SubscriptionResponse {
  plan_id: string | null;
  status: string | null;
  current_period_end: string | null;
  seats_used?: number;
  seats_included?: number;
  is_trial?: boolean;
  subscribed?: boolean;
}

export interface CheckoutResponse {
  checkout_url: string;
}

export interface PortalResponse {
  portal_url: string;
}

export async function getPlans(): Promise<BillingPlansResponse> {
  // Backend does not yet expose a plans catalog; provide static defaults.
  return Promise.resolve({
    plans: [
      {
        id: 'basic',
        price_id: 'price_basic',
        name: 'Basic',
        amount_cents: 1900,
        interval: 'month',
        features: ['Scope 2 tracking', 'Email support'],
      },
      {
        id: 'pro',
        price_id: 'price_pro',
        name: 'Pro',
        amount_cents: 4900,
        interval: 'month',
        features: ['All compliance frameworks', 'Priority support'],
      },
      {
        id: 'enterprise',
        price_id: 'price_enterprise',
        name: 'Enterprise',
        amount_cents: 9900,
        interval: 'month',
        features: ['Dedicated success', 'Custom SLAs'],
      },
    ],
  });
}

export async function getSubscription(): Promise<SubscriptionResponse> {
  const status = await api.get<{
    subscribed: boolean;
    plan?: string | null;
    status?: string | null;
    currentPeriodEnd?: string | null;
  }>('/api/billing/status');

  return {
    plan_id: status.plan ?? null,
    status: status.status ?? (status.subscribed ? 'active' : null),
    current_period_end: status.currentPeriodEnd ?? null,
    subscribed: status.subscribed,
  };
}

export async function createCheckoutSession(planId: string, successUrl: string, cancelUrl: string): Promise<string> {
  const response = await api.post<CheckoutResponse>('/api/billing/checkout', {
    plan_id: planId,
    success_url: successUrl,
    cancel_url: cancelUrl,
  });
  return response.checkout_url;
}

export async function createPortalSession(returnUrl: string): Promise<string> {
  const response = await api.post<PortalResponse>('/api/billing/portal', {
    return_url: returnUrl,
  });
  return response.portal_url;
}

export async function hasActiveSubscription(): Promise<boolean> {
  try {
    const status = await getSubscription();
    return status.status === 'active' || status.status === 'trialing';
  } catch (e) {
    if (e instanceof ApiRequestError && e.status === 401) {
      return false;
    }
    throw e;
  }
}

export function formatSubscriptionStatus(status: string | null | undefined): string {
  if (!status) return 'None';

  const statusMap: Record<string, string> = {
    active: 'Active',
    trialing: 'Trial',
    past_due: 'Past Due',
    canceled: 'Canceled',
    unpaid: 'Unpaid',
  };

  return statusMap[status] || status;
}

export function formatPeriodEnd(dateString: string | null | undefined): string {
  if (!dateString) return 'N/A';

  const date = new Date(dateString);
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  });
}

export function formatPrice(amountCents: number, interval: 'month' | 'year'): string {
  return `$${(amountCents / 100).toFixed(2)}/${interval}`;
}
