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
  checkout_url?: string;
  url?: string;
}

export interface PortalResponse {
  portal_url?: string;
  url?: string;
}

export async function getPlans(): Promise<BillingPlansResponse> {
  try {
    const response = await api.get<BillingPlansResponse>('/api/billing/plans');
    if (response?.plans?.length) {
      return response;
    }
  } catch {
    // Fallback to static defaults if the API is unavailable.
  }

  return {
    plans: [
      // Carbon Starter — Monthly
      {
        id: 'basic_monthly',
        price_id: 'price_basic_monthly',
        name: 'Carbon Starter',
        amount_cents: 208300,
        interval: 'month',
        features: [
          'Up to 5 users',
          'Scope 1 & 2 emissions tracking',
          'CSV data import',
          'Basic GHG reports',
          'Annual methodology review',
          'Email support',
          '$1,000/site/year for additional sites',
        ],
      },
      // Carbon Starter — Annual
      {
        id: 'basic_annual',
        price_id: 'price_basic_annual',
        name: 'Carbon Starter',
        amount_cents: 2500000,
        interval: 'year',
        features: [
          'Up to 5 users',
          'Scope 1 & 2 emissions tracking',
          'CSV data import',
          'Basic GHG reports',
          'Annual methodology review',
          'Email support',
          '$1,000/site/year for additional sites',
          '2 months free vs monthly',
        ],
      },
      // Carbon Command — Monthly
      {
        id: 'pro_monthly',
        price_id: 'price_pro_monthly',
        name: 'Carbon Command',
        amount_cents: 375000,
        interval: 'month',
        features: [
          'Unlimited users',
          'Scope 1, 2 & 3 emissions tracking',
          'Cloud API connectors — AWS, Azure, GCP, SAP, Utilities',
          'CSRD, SEC & SB 253 compliance reports',
          'Live emissions dashboard',
          'Advanced analytics & forecasting',
          'Dedicated account manager',
          'Quarterly business reviews',
          '$800/site/year for additional sites',
        ],
      },
      // Carbon Command — Annual
      {
        id: 'pro_annual',
        price_id: 'price_pro_annual',
        name: 'Carbon Command',
        amount_cents: 4500000,
        interval: 'year',
        features: [
          'Unlimited users',
          'Scope 1, 2 & 3 emissions tracking',
          'Cloud API connectors — AWS, Azure, GCP, SAP, Utilities',
          'CSRD, SEC & SB 253 compliance reports',
          'Live emissions dashboard',
          'Advanced analytics & forecasting',
          'Dedicated account manager',
          'Quarterly business reviews',
          '$800/site/year for additional sites',
          '2 months free vs monthly',
        ],
      },
      // Carbon Command Elite — Contact Us
      {
        id: 'enterprise',
        price_id: 'price_enterprise',
        name: 'Carbon Command Elite',
        amount_cents: 0,
        interval: 'year',
        features: [
          'Everything in Carbon Command',
          'All global frameworks — CSRD, SEC, CBAM, IFRS S2, GRI, CDP, CA SB 253',
          'Multi-region compliance (EU, UK, CA & more)',
          'Custom calculation methodologies',
          'On-site implementation support',
          'White-label branding & SSO',
          'Executive dashboard & board reporting',
          'Dedicated customer success manager',
          '99.9% SLA guarantee',
          'Custom pricing — contact us',
        ],
      },
    ],
  };
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
    plan: planId,
    plan_id: planId,
    success_url: successUrl,
    cancel_url: cancelUrl,
  });

  const checkoutUrl = response.checkout_url ?? response.url;
  if (!checkoutUrl) {
    throw new Error('Checkout URL missing from billing API response');
  }
  return checkoutUrl;
}

export async function createPortalSession(returnUrl: string): Promise<string> {
  const response = await api.post<PortalResponse>('/api/billing/portal', {
    return_url: returnUrl,
  });

  const portalUrl = response.portal_url ?? response.url;
  if (!portalUrl) {
    throw new Error('Portal URL missing from billing API response');
  }
  return portalUrl;
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
    timeZone: 'UTC',
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  });
}

export function formatPrice(amountCents: number, interval: 'month' | 'year'): string {
  if (amountCents === 0) return 'Contact Us';
  const dollars = amountCents / 100;
  const formatted = dollars % 1 === 0
    ? `${dollars.toLocaleString('en-US')}`
    : `${dollars.toLocaleString('en-US', { minimumFractionDigits: 2 })}`;
  return `${formatted}/${interval}`;
}
