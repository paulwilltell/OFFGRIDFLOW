'use client';

import { useCallback, useEffect, useState } from 'react';
import Link from 'next/link';
import { api } from '@/lib/api';
import { useRequireAuth } from '@/lib/session';

type User = {
  id: string;
  email: string;
  name?: string;
  role?: string;
};

export default function UsersPage() {
  const session = useRequireAuth();
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [toast, setToast] = useState<string | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await api.get<User[]>('/api/users');
      setUsers(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load users');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (!session.isAuthenticated) return;
    void load();
  }, [session.isAuthenticated, load]);

  const deleteUser = async (id: string) => {
    setLoading(true);
    setError(null);
    try {
      await api.delete(`/api/users?id=${encodeURIComponent(id)}`);
      setToast('User deleted');
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete user');
    } finally {
      setLoading(false);
    }
  };

  if (session.loading || !session.isAuthenticated) {
    return (
      <div className="p-8">
        <h1 className="text-xl font-semibold">User management</h1>
        <p className="mt-2 text-sm text-gray-500">Checking your session...</p>
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto py-10 px-4 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-semibold">User management</h1>
          <p className="mt-1 text-gray-600 dark:text-gray-400">
            View users in this tenant and remove stale accounts. Admins can add users via the API.
          </p>
        </div>
        <Link
          href="/settings"
          className="inline-flex items-center rounded-md border border-gray-200 bg-white px-3 py-1.5 text-xs font-medium text-gray-700 shadow-sm hover:bg-gray-50"
        >
          ⟵ Back to settings
        </Link>
      </div>

      {toast && <div className="text-sm text-emerald-600">{toast}</div>}
      {error && <div className="text-sm text-red-600">{error}</div>}

      <div className="rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900/60 shadow-sm">
        <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-800">
          <thead className="bg-gray-50 dark:bg-gray-800/50">
            <tr>
              <th className="px-4 py-2 text-left text-xs font-semibold text-gray-600 dark:text-gray-300">Name</th>
              <th className="px-4 py-2 text-left text-xs font-semibold text-gray-600 dark:text-gray-300">Email</th>
              <th className="px-4 py-2 text-left text-xs font-semibold text-gray-600 dark:text-gray-300">Role</th>
              <th className="px-4 py-2 text-right text-xs font-semibold text-gray-600 dark:text-gray-300">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100 dark:divide-gray-800">
            {users.map((u) => (
              <tr key={u.id}>
                <td className="px-4 py-2 text-sm">{u.name || '—'}</td>
                <td className="px-4 py-2 text-sm">{u.email}</td>
                <td className="px-4 py-2 text-sm">{u.role || 'member'}</td>
                <td className="px-4 py-2 text-right">
                  <button
                    onClick={() => deleteUser(u.id)}
                    className="text-xs text-red-600 hover:underline disabled:opacity-50"
                    disabled={loading}
                  >
                    Delete
                  </button>
                </td>
              </tr>
            ))}
            {users.length === 0 && (
              <tr>
                <td colSpan={4} className="px-4 py-6 text-sm text-gray-500 text-center">
                  No users found for this tenant.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
