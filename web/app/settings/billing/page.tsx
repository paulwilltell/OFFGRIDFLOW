'use client';

import { useState, useEffect, Suspense } from 'react';
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

function BillingContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const session = useRequireAuth();

  const success = searchParams.get('success') === 'true';
  const canceled = searchParams.get('canceled') === 'true';

  const [subscription, setSubscription] = useState<SubscriptionResponse | null>(null);
  const [plans, setPlans] = useState<BillingPlan[]>([]);
  const [loading, setLoading] = useState(true);
  const [checkoutLoading, setCheckoutLoading] = useState<string | null>(null);
  const [portalLoading, setPortalLoading] = useState(false);
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
    } catch (err) {
      setError('Failed to start checkout. Please try again.');
      setCheckoutLoading(null);
    }
  };

  const handleManageSubscription = async () => {
    setPortalLoading(true);
    setError(null);
    try {
      const returnUrl = `${window.location.origin}/settings/billing`;
      const portalUrl = await createPortalSession(returnUrl);
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

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="text-center">
          <h1 className="text-3xl font-extrabold text-gray-900 dark:text-white sm:text-4xl">
            Billing & Subscription
          </h1>
          <p className="mt-4 text-lg text-gray-600 dark:text-gray-400">
            Manage your subscription and billing settings
          </p>
        </div>

        {/* Success/Cancel Messages */}
        {success && (
          <div className="mt-8 max-w-xl mx-auto rounded-md bg-green-50 dark:bg-green-900/50 p-4">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg className="h-5 w-5 text-green-400" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                </svg>
              </div>
              <div className="ml-3">
                <p className="text-sm font-medium text-green-800 dark:text-green-200">
                  Subscription activated successfully!
                </p>
              </div>
            </div>
          </div>
        )}

        {canceled && (
          <div className="mt-8 max-w-xl mx-auto rounded-md bg-yellow-50 dark:bg-yellow-900/50 p-4">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg className="h-5 w-5 text-yellow-400" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                </svg>
              </div>
              <div className="ml-3">
                <p className="text-sm font-medium text-yellow-800 dark:text-yellow-200">
                  Checkout was canceled. No charges were made.
                </p>
              </div>
            </div>
          </div>
        )}

        {error && (
          <div className="mt-8 max-w-xl mx-auto rounded-md bg-red-50 dark:bg-red-900/50 p-4">
            <p className="text-sm text-red-700 dark:text-red-200">{error}</p>
          </div>
        )}

        {/* Current Subscription Status */}
        {subscription && (
          <div className="mt-12 max-w-xl mx-auto bg-white dark:bg-gray-800 rounded-lg shadow p-6">
            <h2 className="text-lg font-medium text-gray-900 dark:text-white">
              Current Subscription
            </h2>
            <dl className="mt-4 space-y-4">
              <div className="flex justify-between">
                <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">Plan</dt>
                <dd className="text-sm text-gray-900 dark:text-white capitalize">{subscription.plan_id}</dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">Status</dt>
                <dd className="text-sm text-gray-900 dark:text-white">
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
                <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">Next billing date</dt>
                <dd className="text-sm text-gray-900 dark:text-white">
                  {formatPeriodEnd(subscription.current_period_end)}
                </dd>
              </div>
              {(subscription.seats_used !== undefined || subscription.seats_included !== undefined) && (
                <div className="flex justify-between">
                  <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">Seats</dt>
                  <dd className="text-sm text-gray-900 dark:text-white">
                    {subscription.seats_used ?? 0} / {subscription.seats_included ?? '—'}
                  </dd>
                </div>
              )}
            </dl>
            <div className="mt-6">
              <button
                onClick={handleManageSubscription}
                disabled={portalLoading}
                className="w-full flex justify-center py-2 px-4 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm text-sm font-medium text-gray-700 dark:text-gray-200 bg-white dark:bg-gray-700 hover:bg-gray-50 dark:hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-green-500 disabled:opacity-50"
              >
                {portalLoading ? 'Loading...' : 'Manage Subscription'}
              </button>
            </div>
          </div>
        )}

        {/* Pricing Plans */}
        <div className="mt-12">
          <h2 className="text-2xl font-bold text-center text-gray-900 dark:text-white mb-8">
            {subscription ? 'Available Plans' : 'Choose Your Plan'}
          </h2>
          
          <div className="grid gap-8 lg:grid-cols-2 max-w-5xl mx-auto">
            {plans.map((plan) => (
              <PlanCard
                key={plan.id}
                plan={plan}
                currentPlan={subscription?.plan_id || null}
                loading={checkoutLoading === plan.id}
                onSubscribe={() => handleSubscribe(plan.id)}
              />
            ))}
          </div>
        </div>

        {/* Back Link */}
        <div className="mt-12 text-center">
          <Link
            href="/settings"
            className="text-sm text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200"
          >
            ← Back to Settings
          </Link>
        </div>
      </div>
    </div>
  );
}

interface PlanCardProps {
  plan: BillingPlan;
  currentPlan: string | null;
  loading: boolean;
  onSubscribe: () => void;
}

function PlanCard({ plan, currentPlan, loading, onSubscribe }: PlanCardProps) {
  const isCurrent = currentPlan === plan.id;

  return (
    <div className="relative rounded-2xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 p-8">
      <div className="text-center">
        <h3 className="text-xl font-semibold text-gray-900 dark:text-white">{plan.name}</h3>
        <div className="mt-4">
          <span className="text-3xl font-extrabold text-gray-900 dark:text-white">
            {formatPrice(plan.amount_cents, plan.interval)}
          </span>
        </div>
      </div>

      <ul className="mt-6 space-y-3">
        {plan.features.map((feature, index) => (
          <li key={index} className="flex items-start">
            <svg
              className="h-5 w-5 text-green-500 mt-0.5"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
            </svg>
            <span className="ml-3 text-sm text-gray-600 dark:text-gray-300">{feature}</span>
          </li>
        ))}
      </ul>

      <div className="mt-8">
        {isCurrent ? (
          <button
            disabled
            className="w-full py-3 px-4 rounded-lg text-sm font-medium bg-gray-100 dark:bg-gray-700 text-gray-500 dark:text-gray-400 cursor-not-allowed"
          >
            Current Plan
          </button>
        ) : (
          <button
            onClick={onSubscribe}
            disabled={loading}
            className="w-full py-3 px-4 rounded-lg text-sm font-medium transition-colors bg-green-600 text-white hover:bg-green-700 disabled:opacity-50"
          >
            {loading ? 'Loading...' : `Subscribe to ${plan.name}`}
          </button>
        )}
      </div>
    </div>
  );
}

export default function BillingPage() {
  return (
    <Suspense
      fallback={
        <div className="min-h-screen bg-gray-50 dark:bg-gray-900 flex items-center justify-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-green-500" />
        </div>
      }
    >
      <BillingContent />
    </Suspense>
  );
}
