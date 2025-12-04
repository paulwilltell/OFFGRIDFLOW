'use client';

import Link from 'next/link';
import { FormEvent, useState } from 'react';
import { api } from '@/lib/api';
import { useRequireAuth } from '@/lib/session';

export default function SecuritySettingsPage() {
  const session = useRequireAuth();
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  if (session.loading || !session.isAuthenticated) {
    return (
      <div className="p-8">
        <h1 className="text-xl font-semibold">Security</h1>
        <p className="mt-2 text-sm text-gray-500">Checking your session...</p>
      </div>
    );
  }

  const onSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError(null);
    setSuccess(null);
    setLoading(true);
    try {
      await api.post('/api/auth/change-password', {
        current_password: currentPassword,
        new_password: newPassword,
      });
      setSuccess('Password updated successfully');
      setCurrentPassword('');
      setNewPassword('');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to change password');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="max-w-3xl mx-auto py-10 px-4 space-y-8">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-semibold">Security settings</h1>
          <p className="mt-2 text-gray-600 dark:text-gray-400">
            Manage your password. API keys are available under Auth → Keys, and billing controls remain in Billing.
          </p>
        </div>
        <Link
          href="/settings"
          className="inline-flex items-center rounded-md border border-gray-200 bg-white px-3 py-1.5 text-xs font-medium text-gray-700 shadow-sm hover:bg-gray-50"
        >
          ⟵ Back to settings
        </Link>
      </div>

      <form onSubmit={onSubmit} className="rounded-lg border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900/50 p-6 space-y-4 shadow-sm">
        <h2 className="text-lg font-medium">Change password</h2>
        {error && <div className="text-sm text-red-600 dark:text-red-400">{error}</div>}
        {success && <div className="text-sm text-emerald-600">{success}</div>}
        <label className="block space-y-1">
          <span className="text-sm text-gray-700 dark:text-gray-300">Current password</span>
          <input
            type="password"
            className="w-full rounded-md border border-gray-300 dark:border-gray-700 bg-transparent px-3 py-2 text-sm"
            value={currentPassword}
            onChange={(e) => setCurrentPassword(e.target.value)}
            required
          />
        </label>
        <label className="block space-y-1">
          <span className="text-sm text-gray-700 dark:text-gray-300">New password</span>
          <input
            type="password"
            className="w-full rounded-md border border-gray-300 dark:border-gray-700 bg-transparent px-3 py-2 text-sm"
            value={newPassword}
            onChange={(e) => setNewPassword(e.target.value)}
            required
            minLength={8}
          />
          <span className="text-xs text-gray-500">Use 8+ chars with upper, lower, and digits.</span>
        </label>
        <button
          type="submit"
          disabled={loading}
          className="inline-flex items-center justify-center rounded-md bg-emerald-600 px-4 py-2 text-sm font-medium text-white shadow hover:bg-emerald-700 disabled:opacity-50"
        >
          {loading ? 'Updating…' : 'Update password'}
        </button>
      </form>
    </div>
  );
}
