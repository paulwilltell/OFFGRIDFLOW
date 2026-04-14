'use client';

import { useEffect, useState } from 'react';
import { api } from '@/lib/api';
import { useRequireAuth, useSession } from '@/lib/session';

interface Approval {
  id: string;
  entity_type: string;
  entity_id: string;
  status: string;
  prepared_by?: string;
  prepared_at?: string;
  reviewed_by?: string;
  reviewed_at?: string;
  review_notes?: string;
  approved_by?: string;
  approved_at?: string;
  approval_notes?: string;
  rejected_by?: string;
  rejected_at?: string;
  rejection_reason?: string;
  created_at: string;
  updated_at: string;
}

const statusConfig: Record<string, { label: string; bg: string; text: string }> = {
  draft: { label: 'Draft', bg: 'bg-gray-500/10', text: 'text-gray-400' },
  submitted: { label: 'Submitted', bg: 'bg-blue-500/10', text: 'text-blue-400' },
  reviewed: { label: 'Reviewed', bg: 'bg-amber-500/10', text: 'text-amber-400' },
  approved: { label: 'Approved', bg: 'bg-green-500/10', text: 'text-green-400' },
  rejected: { label: 'Rejected', bg: 'bg-red-500/10', text: 'text-red-400' },
};

export default function ApprovalsPage() {
  useRequireAuth();
  const { user } = useSession();
  const [approvals, setApprovals] = useState<Approval[]>([]);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState<string | null>(null);
  const [notes, setNotes] = useState('');

  const load = async () => {
    try {
      const res = await api.get<{ approvals: Approval[] }>('/api/audit/approvals');
      setApprovals(res.approvals || []);
    } catch {
      setApprovals([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { load(); }, []);

  const handleAction = async (id: string, action: string) => {
    setActionLoading(id);
    try {
      await api.put(`/api/audit/approvals/${id}`, { action, notes });
      setNotes('');
      await load();
    } catch (err) {
      console.error('Action failed:', err);
    } finally {
      setActionLoading(null);
    }
  };

  const handleCreate = async () => {
    try {
      await api.post('/api/audit/approvals', {
        entity_type: 'report',
        entity_id: `inventory-${new Date().getFullYear()}`,
      });
      await load();
    } catch (err) {
      console.error('Create failed:', err);
    }
  };

  return (
    <div>
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-white">Approval Workflow</h1>
          <p className="mt-1 text-xs text-gray-500">
            Preparer &#8594; Reviewer &#8594; Approver &mdash; track sign-off for every reportable entity
          </p>
        </div>
        <button
          onClick={handleCreate}
          className="rounded-lg bg-primary-600 px-4 py-2 text-sm font-medium text-white hover:bg-primary-500"
        >
          New Approval Request
        </button>
      </div>

      {/* Workflow steps visualization */}
      <div className="mb-6 flex items-center justify-center gap-2 rounded-lg border border-gray-800 bg-gray-900/30 p-4">
        {['Draft', 'Submitted', 'Reviewed', 'Approved'].map((step, i) => (
          <div key={step} className="flex items-center gap-2">
            <div className="flex h-8 w-8 items-center justify-center rounded-full bg-gray-800 text-xs font-bold text-gray-400">
              {i + 1}
            </div>
            <span className="text-xs text-gray-400">{step}</span>
            {i < 3 && <span className="mx-1 text-gray-700">&#8594;</span>}
          </div>
        ))}
      </div>

      {loading ? (
        <div className="py-20 text-center text-gray-500">Loading approvals...</div>
      ) : approvals.length === 0 ? (
        <div className="rounded-xl border border-gray-800 bg-gray-800/30 p-12 text-center">
          <div className="mb-3 text-3xl">&#9989;</div>
          <h3 className="mb-2 text-base font-semibold text-white">No Approval Requests</h3>
          <p className="text-sm text-gray-400">
            Create an approval request to begin the review workflow for a report or inventory period.
          </p>
        </div>
      ) : (
        <div className="space-y-3">
          {approvals.map(a => {
            const s = statusConfig[a.status] || statusConfig.draft;
            return (
              <div key={a.id} className="rounded-xl border border-gray-800 bg-gray-800/30 p-5">
                <div className="flex items-start justify-between">
                  <div>
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-medium text-white">{a.entity_type}: {a.entity_id}</span>
                      <span className={`rounded px-2 py-0.5 text-[10px] font-medium ${s.bg} ${s.text}`}>{s.label}</span>
                    </div>
                    <div className="mt-1 text-xs text-gray-500">Created {new Date(a.created_at).toLocaleString()}</div>
                  </div>
                </div>

                {/* Timeline */}
                <div className="mt-4 space-y-2">
                  {a.prepared_at && (
                    <div className="flex items-center gap-2 text-xs text-gray-400">
                      <span className="h-1.5 w-1.5 rounded-full bg-gray-500" />
                      Prepared {new Date(a.prepared_at).toLocaleString()}
                    </div>
                  )}
                  {a.reviewed_at && (
                    <div className="flex items-center gap-2 text-xs text-gray-400">
                      <span className="h-1.5 w-1.5 rounded-full bg-amber-500" />
                      Reviewed {new Date(a.reviewed_at).toLocaleString()}
                      {a.review_notes && <span className="text-gray-500">&mdash; {a.review_notes}</span>}
                    </div>
                  )}
                  {a.approved_at && (
                    <div className="flex items-center gap-2 text-xs text-green-400">
                      <span className="h-1.5 w-1.5 rounded-full bg-green-500" />
                      Approved {new Date(a.approved_at).toLocaleString()}
                    </div>
                  )}
                  {a.rejected_at && (
                    <div className="flex items-center gap-2 text-xs text-red-400">
                      <span className="h-1.5 w-1.5 rounded-full bg-red-500" />
                      Rejected: {a.rejection_reason}
                    </div>
                  )}
                </div>

                {/* Actions based on current status */}
                {(a.status === 'draft' || a.status === 'submitted' || a.status === 'reviewed') && (
                  <div className="mt-4 flex items-center gap-2 border-t border-gray-800 pt-4">
                    <input
                      type="text"
                      placeholder="Notes (optional)"
                      value={notes}
                      onChange={e => setNotes(e.target.value)}
                      className="flex-1 rounded-lg border border-gray-700 bg-gray-900 px-3 py-1.5 text-xs text-white placeholder-gray-600"
                    />
                    {a.status === 'draft' && (
                      <button
                        onClick={() => handleAction(a.id, 'submit')}
                        disabled={actionLoading === a.id}
                        className="rounded-lg bg-blue-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-blue-500 disabled:opacity-50"
                      >
                        Submit for Review
                      </button>
                    )}
                    {a.status === 'submitted' && (
                      <>
                        <button
                          onClick={() => handleAction(a.id, 'review')}
                          disabled={actionLoading === a.id}
                          className="rounded-lg bg-amber-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-amber-500 disabled:opacity-50"
                        >
                          Mark Reviewed
                        </button>
                        <button
                          onClick={() => handleAction(a.id, 'reject')}
                          disabled={actionLoading === a.id}
                          className="rounded-lg bg-red-600/20 px-3 py-1.5 text-xs font-medium text-red-400 hover:bg-red-600/30 disabled:opacity-50"
                        >
                          Reject
                        </button>
                      </>
                    )}
                    {a.status === 'reviewed' && (
                      <>
                        <button
                          onClick={() => handleAction(a.id, 'approve')}
                          disabled={actionLoading === a.id}
                          className="rounded-lg bg-green-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-green-500 disabled:opacity-50"
                        >
                          Approve
                        </button>
                        <button
                          onClick={() => handleAction(a.id, 'reject')}
                          disabled={actionLoading === a.id}
                          className="rounded-lg bg-red-600/20 px-3 py-1.5 text-xs font-medium text-red-400 hover:bg-red-600/30 disabled:opacity-50"
                        >
                          Reject
                        </button>
                      </>
                    )}
                  </div>
                )}
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
