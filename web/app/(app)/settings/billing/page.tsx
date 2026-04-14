'use client';

import { useState, useEffect, Suspense } from 'react';
import EliteInquiryModal from '@/components/EliteInquiryModal';
import { useRouter, useSearchParams } from 'next/navigation';
import Link from 'next/link';
import {
  BillingPlan,
  BillingPlansResponse,
  SubscriptionResponse,
  getPlans,
  getSubscription,
  createCheckoutSession,
  createPortalSession,
  formatSubscriptionStatus,
  formatPeriodEnd,
  formatPrice,
} from '@/lib/billing';
import { useRequireAuth } from '@/lib/session';

// Group plans by name so we can show monthly/annual side by side per tier
type PlanGroup = {
  name: string;
  monthly: BillingPlan | null;
  annual: BillingPlan | null;
};

function groupPlans(plans: BillingPlan[]): PlanGroup[] {
  const map = new Map<string, PlanGroup>();
  for (const plan of plans) {
    if (!map.has(plan.name)) {
      map.set(plan.name, { name: plan.name, monthly: null, annual: null });
    }
    const group = map.get(plan.name)!;
    if (plan.interval === 'month') group.monthly = plan;
    else group.annual = plan;
  }
  return Array.from(map.values());
}

function BillingContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const session = useRequireAuth();

  const success = searchParams.get('success') === 'true';
  const canceled = searchParams.get('canceled') === 'true';

  const [subscription, setSubscription] = useState<SubscriptionResponse | null>(null);
  const [plans, setPlans] = useState<BillingPlan[]>([]);
  const [billingInterval, setBillingInterval] = useState<'month' | 'year'>('year');
  const [loading, setLoading] = useState(true);
  const [checkoutLoading, setCheckoutLoading] = useState<string | null>(null);
  const [portalLoading, setPortalLoading] = useState(false);
  const [eliteModalOpen, setEliteModalOpen] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!session.loading && !session.isAuthenticated) {
      router.replace('/login?returnTo=/settings/billing');
    }
  }, [session.loading, session.isAuthenticated, router]);

  useEffect(() => {
    if (!session.loading && session.isAuthenticated) {
      void loadBilling();
    }
  }, [session.loading, session.isAuthenticated]);

  const loadBilling = async () => {
    setLoading(true);
    setError(null);
    try {
      const [subscriptionRes, plansRes] = await Promise.all([
        getSubscription(),
        getPlans(),
      ]);
      setSubscription(subscriptionRes);
      setPlans((plansRes as BillingPlansResponse).plans ?? []);
    } catch {
      setError('Failed to load billing data');
    } finally {
      setLoading(false);
    }
  };

  const handleSubscribe = async (planId: string) => {
    setCheckoutLoading(planId);
    setError(null);
    try {
      const baseUrl = window.location.origin;
      const checkoutUrl = await createCheckoutSession(
        planId,
        `${baseUrl}/settings/billing?success=true`,
        `${baseUrl}/settings/billing?canceled=true`,
      );
      window.location.href = checkoutUrl;
    } catch {
      setError('Failed to start checkout. Please try again.');
      setCheckoutLoading(null);
    }
  };

  const handleManageSubscription = async () => {
    setPortalLoading(true);
    setError(null);
    try {
      const portalUrl = await createPortalSession(`${window.location.origin}/settings/billing`);
      window.location.href = portalUrl;
    } catch {
      setError('Failed to open billing portal. Please try again.');
      setPortalLoading(false);
    }
  };

  if (session.loading || !session.isAuthenticated || loading) {
    return (
      <div className="min-h-screen bg-gray-50 dark:bg-gray-900 flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-green-500" />
      </div>
    );
  }

  const planGroups = groupPlans(plans);
  const hasMonthlyPlans = planGroups.some((group) => group.monthly);
  const hasManagedSubscription = Boolean(subscription?.subscribed && subscription?.plan_id && subscription?.status);

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-7xl mx-auto">

        {/* Header */}
        <div className="text-center">
          <h1 className="text-3xl font-extrabold text-gray-900 dark:text-white sm:text-4xl">
            Billing &amp; Subscription
          </h1>
          <p className="mt-4 text-lg text-gray-600 dark:text-gray-400">
            Transparent carbon accounting. Priced for value.
          </p>
        </div>

        {/* Alerts */}
        {success && (
          <div className="mt-8 max-w-xl mx-auto rounded-md bg-green-50 dark:bg-green-900/50 p-4 flex items-start gap-3">
            <svg className="h-5 w-5 text-green-400 mt-0.5 shrink-0" viewBox="0 0 20 20" fill="currentColor">
              <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
            </svg>
            <p className="text-sm font-medium text-green-800 dark:text-green-200">Subscription activated successfully!</p>
          </div>
        )}
        {canceled && (
          <div className="mt-8 max-w-xl mx-auto rounded-md bg-yellow-50 dark:bg-yellow-900/50 p-4">
            <p className="text-sm font-medium text-yellow-800 dark:text-yellow-200">Checkout canceled. No charges were made.</p>
          </div>
        )}
        {error && (
          <div className="mt-8 max-w-xl mx-auto rounded-md bg-red-50 dark:bg-red-900/50 p-4">
            <p className="text-sm text-red-700 dark:text-red-200">{error}</p>
          </div>
        )}

        {/* Current subscription */}
        {subscription && (
          <div className="mt-10 max-w-xl mx-auto bg-white dark:bg-gray-800 rounded-xl shadow p-6">
            <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Current Subscription</h2>
            <dl className="space-y-3 text-sm">
              <div className="flex justify-between">
                <dt className="text-gray-500 dark:text-gray-400">Plan</dt>
                <dd className="font-medium text-gray-900 dark:text-white capitalize">{subscription.plan_id?.replace('_', ' ') ?? '—'}</dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-gray-500 dark:text-gray-400">Status</dt>
                <dd>
                  <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                    subscription.status === 'active' || subscription.status === 'trialing'
                      ? 'bg-green-100 text-green-800 dark:bg-green-800 dark:text-green-100'
                      : 'bg-yellow-100 text-yellow-800 dark:bg-yellow-800 dark:text-yellow-100'
                  }`}>
                    {formatSubscriptionStatus(subscription.status)}
                  </span>
                </dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-gray-500 dark:text-gray-400">Next billing date</dt>
                <dd className="font-medium text-gray-900 dark:text-white">{formatPeriodEnd(subscription.current_period_end)}</dd>
              </div>
            </dl>
            {hasManagedSubscription ? (
              <button
                onClick={handleManageSubscription}
                disabled={portalLoading}
                className="mt-5 w-full py-2 px-4 border border-gray-300 dark:border-gray-600 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-200 bg-white dark:bg-gray-700 hover:bg-gray-50 dark:hover:bg-gray-600 disabled:opacity-50 transition-colors"
              >
                {portalLoading ? 'Loading…' : 'Manage Subscription'}
              </button>
            ) : (
              <p className="mt-5 text-sm text-gray-500 dark:text-gray-400">
                No active subscription yet. Choose a plan below to start checkout.
              </p>
            )}
          </div>
        )}

        {/* Monthly / Annual toggle */}
        <div className="mt-12 flex flex-col items-center gap-3">
          <h2 className="text-2xl font-bold text-gray-900 dark:text-white">
            {subscription ? 'Change Plan' : 'Choose Your Plan'}
          </h2>
          {hasMonthlyPlans && (
            <div className="flex items-center gap-3 mt-2">
              <span className={`text-sm font-medium ${billingInterval === 'month' ? 'text-gray-900 dark:text-white' : 'text-gray-400 dark:text-gray-500'}`}>
                Monthly
              </span>
              <button
                onClick={() => setBillingInterval(billingInterval === 'month' ? 'year' : 'month')}
                className={`relative inline-flex h-7 w-14 items-center rounded-full transition-colors focus:outline-none ${
                  billingInterval === 'year' ? 'bg-green-600' : 'bg-gray-300 dark:bg-gray-600'
                }`}
              >
                <span className={`inline-block h-5 w-5 transform rounded-full bg-white shadow transition-transform ${
                  billingInterval === 'year' ? 'translate-x-8' : 'translate-x-1'
                }`} />
              </button>
              <span className={`text-sm font-medium ${billingInterval === 'year' ? 'text-gray-900 dark:text-white' : 'text-gray-400 dark:text-gray-500'}`}>
                Annual
              </span>
            </div>
          )}
        </div>

        {/* Plan cards */}
        <div className="mt-8 grid gap-8 lg:grid-cols-3 max-w-6xl mx-auto">
          {planGroups.map((group) => {
            const plan = billingInterval === 'month' ? (group.monthly ?? group.annual) : (group.annual ?? group.monthly);
            if (!plan) return null;
            return (
              <PlanCard
                key={group.name}
                plan={plan}
                currentPlanId={subscription?.plan_id ?? null}
                loading={checkoutLoading === plan.id}
                onSubscribe={() => handleSubscribe(plan.id)}
                onEliteInquiry={() => setEliteModalOpen(true)}
              />
            );
          })}
        </div>

        <div className="mt-12 text-center">
          <Link href="/settings" className="text-sm text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200">
            ← Back to Settings
          </Link>
        </div>

        <EliteInquiryModal
          isOpen={eliteModalOpen}
          onClose={() => setEliteModalOpen(false)}
          userEmail={session.user?.email ?? ''}
          userName={session.user?.name ?? ''}
          companyName={session.user?.company ?? ''}
          currentPlan={subscription?.plan_id ?? ''}
        />
      </div>
    </div>
  );
}

interface PlanCardProps {
  plan: BillingPlan;
  currentPlanId: string | null;
  loading: boolean;
  onSubscribe: () => void;
  onEliteInquiry?: () => void;
}

function PlanCard({ plan, currentPlanId, loading, onSubscribe, onEliteInquiry }: PlanCardProps) {
  const isCustomQuote = plan.id === 'global' || plan.amount_cents === 0;
  const isCommand = plan.name === 'Compliance Pro';
  const isCurrent = currentPlanId === plan.id;

  return (
    <div className={`relative rounded-2xl border bg-white dark:bg-gray-800 p-8 flex flex-col shadow-sm transition-shadow hover:shadow-md ${
      isCommand
        ? 'border-green-500 ring-2 ring-green-500 ring-offset-2 dark:ring-offset-gray-900'
        : 'border-gray-200 dark:border-gray-700'
    }`}>
      {isCommand && (
        <div className="absolute -top-4 left-1/2 -translate-x-1/2">
          <span className="bg-green-500 text-white text-xs font-bold px-4 py-1 rounded-full uppercase tracking-wide">
            Most Popular
          </span>
        </div>
      )}

      <div className="text-center mb-6">
        <h3 className="text-xl font-bold text-gray-900 dark:text-white">{plan.name}</h3>
        <div className="mt-4">
          {isCustomQuote ? (
            <div>
              <span className="text-3xl font-extrabold text-gray-900 dark:text-white">Contact Us</span>
              <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">Custom pricing for complex multi-region deployments</p>
            </div>
          ) : (
            <div>
              <span className="text-4xl font-extrabold text-gray-900 dark:text-white">
                {formatPrice(plan.amount_cents, plan.interval)}
              </span>
              <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">billed annually</p>
            </div>
          )}
        </div>
      </div>

      <ul className="space-y-3 flex-1 mb-8">
        {plan.features.map((feature, i) => (
          <li key={i} className="flex items-start gap-3">
            <svg className="h-5 w-5 text-green-500 shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
            </svg>
            <span className={`text-sm ${feature === '2 months free vs monthly' ? 'text-green-600 dark:text-green-400 font-medium' : 'text-gray-600 dark:text-gray-300'}`}>
              {feature}
            </span>
          </li>
        ))}
      </ul>

      <div>
        {isCurrent ? (
          <button disabled className="w-full py-3 px-4 rounded-xl text-sm font-medium bg-gray-100 dark:bg-gray-700 text-gray-500 dark:text-gray-400 cursor-not-allowed">
            Current Plan
          </button>
        ) : isCustomQuote ? (
          <button
            onClick={onEliteInquiry}
            className="w-full py-3 px-4 rounded-xl text-sm font-medium text-white bg-gray-800 dark:bg-gray-700 hover:bg-gray-900 dark:hover:bg-gray-600 transition-colors"
          >
            Request Global Pricing
          </button>
        ) : (
          <button
            onClick={onSubscribe}
            disabled={loading}
            className={`w-full py-3 px-4 rounded-xl text-sm font-medium transition-colors disabled:opacity-50 ${
              isCommand
                ? 'bg-green-600 text-white hover:bg-green-700'
                : 'bg-gray-900 dark:bg-white text-white dark:text-gray-900 hover:bg-gray-800 dark:hover:bg-gray-100'
            }`}
          >
            {loading ? 'Loading…' : `Get ${plan.name}`}
          </button>
        )}
      </div>
    </div>
  );
}

export default function BillingPage() {
  return (
    <Suspense fallback={
      <div className="min-h-screen bg-gray-50 dark:bg-gray-900 flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-green-500" />
      </div>
    }>
      <BillingContent />
    </Suspense>
  );
}
