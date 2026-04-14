'use client';

import { useCallback, useEffect, useState } from 'react';
import { api } from '@/lib/api';
import { useRequireAuth } from '@/lib/session';

interface Alert {
  id: string;
  title: string;
  description: string;
  priority: 'critical' | 'high' | 'medium' | 'low';
  status: 'open' | 'in_progress' | 'escalated' | 'resolved' | 'dismissed';
  source_type: 'anomaly' | 'threshold' | 'deadline' | 'compliance' | 'connector' | 'approval';
  category: string;
  assigned_to: string | null;
  due_date: string | null;
  created_at: string;
  updated_at: string;
}

interface AlertComment {
  id: string;
  alert_id: string;
  author: string;
  content: string;
  created_at: string;
}

interface AlertSummary {
  open: number;
  in_progress: number;
  escalated: number;
  resolved: number;
  dismissed: number;
  critical_active: number;
}

interface AlertsResponse {
  alerts: Alert[];
  count: number;
  summary: AlertSummary;
}

interface CommentsResponse {
  comments: AlertComment[];
}

const priorityConfig: Record<string, { label: string; bg: string; text: string; dot: string }> = {
  critical: { label: 'Critical', bg: 'bg-red-500/10', text: 'text-red-400', dot: 'bg-red-500' },
  high: { label: 'High', bg: 'bg-orange-500/10', text: 'text-orange-400', dot: 'bg-orange-500' },
  medium: { label: 'Medium', bg: 'bg-yellow-500/10', text: 'text-yellow-400', dot: 'bg-yellow-500' },
  low: { label: 'Low', bg: 'bg-gray-500/10', text: 'text-gray-400', dot: 'bg-gray-500' },
};

const statusConfig: Record<string, { label: string; bg: string; text: string }> = {
  open: { label: 'Open', bg: 'bg-blue-500/10', text: 'text-blue-400' },
  in_progress: { label: 'In Progress', bg: 'bg-amber-500/10', text: 'text-amber-400' },
  escalated: { label: 'Escalated', bg: 'bg-red-500/10', text: 'text-red-400' },
  resolved: { label: 'Resolved', bg: 'bg-green-500/10', text: 'text-green-400' },
  dismissed: { label: 'Dismissed', bg: 'bg-gray-500/10', text: 'text-gray-400' },
};

const sourceLabels: Record<string, string> = {
  anomaly: 'Data Anomaly',
  threshold: 'Threshold Breach',
  deadline: 'Deadline',
  compliance: 'Compliance',
  connector: 'Connector',
  approval: 'Approval',
};

type FilterTab = 'all' | 'open' | 'in_progress' | 'escalated' | 'resolved';

const filterTabs: { key: FilterTab; label: string }[] = [
  { key: 'all', label: 'All' },
  { key: 'open', label: 'Open' },
  { key: 'in_progress', label: 'In Progress' },
  { key: 'escalated', label: 'Escalated' },
  { key: 'resolved', label: 'Resolved' },
];

export default function AlertsPage() {
  useRequireAuth();

  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [summary, setSummary] = useState<AlertSummary>({
    open: 0,
    in_progress: 0,
    escalated: 0,
    resolved: 0,
    dismissed: 0,
    critical_active: 0,
  });
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState<FilterTab>('all');
  const [actionLoading, setActionLoading] = useState<string | null>(null);
  const [expandedComments, setExpandedComments] = useState<string | null>(null);
  const [comments, setComments] = useState<Record<string, AlertComment[]>>({});
  const [commentsLoading, setCommentsLoading] = useState<string | null>(null);
  const [commentInputs, setCommentInputs] = useState<Record<string, string>>({});
  const [showCommentForm, setShowCommentForm] = useState<string | null>(null);
  const [escalateTarget, setEscalateTarget] = useState<Record<string, string>>({});
  const [showEscalateForm, setShowEscalateForm] = useState<string | null>(null);

  const loadAlerts = useCallback(async (tab: FilterTab) => {
    try {
      const params = tab !== 'all' ? `?status=${tab}` : '';
      const res = await api.get<AlertsResponse>(`/api/audit/alerts${params}`);
      setAlerts(res.alerts || []);
      setSummary(res.summary || {
        open: 0,
        in_progress: 0,
        escalated: 0,
        resolved: 0,
        dismissed: 0,
        critical_active: 0,
      });
    } catch {
      setAlerts([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    setLoading(true);
    loadAlerts(activeTab);
  }, [activeTab, loadAlerts]);

  const loadComments = useCallback(async (alertId: string) => {
    setCommentsLoading(alertId);
    try {
      const res = await api.get<CommentsResponse>(`/api/audit/alerts/${alertId}/comments`);
      setComments(prev => ({ ...prev, [alertId]: res.comments || [] }));
    } catch {
      setComments(prev => ({ ...prev, [alertId]: [] }));
    } finally {
      setCommentsLoading(null);
    }
  }, []);

  const toggleComments = useCallback((alertId: string) => {
    if (expandedComments === alertId) {
      setExpandedComments(null);
    } else {
      setExpandedComments(alertId);
      if (!comments[alertId]) {
        loadComments(alertId);
      }
    }
  }, [expandedComments, comments, loadComments]);

  const handleAction = useCallback(async (
    alertId: string,
    action: string,
    extraBody?: Record<string, string>,
  ) => {
    setActionLoading(alertId);
    try {
      await api.put(`/api/audit/alerts/${alertId}`, { action, ...extraBody });
      await loadAlerts(activeTab);
      if (action === 'comment' && expandedComments === alertId) {
        await loadComments(alertId);
      }
    } catch (err) {
      console.error(`Alert action "${action}" failed:`, err);
    } finally {
      setActionLoading(null);
    }
  }, [activeTab, expandedComments, loadAlerts, loadComments]);

  const submitComment = useCallback(async (alertId: string) => {
    const content = (commentInputs[alertId] || '').trim();
    if (!content) return;
    await handleAction(alertId, 'comment', { content });
    setCommentInputs(prev => ({ ...prev, [alertId]: '' }));
    setShowCommentForm(null);
  }, [commentInputs, handleAction]);

  const submitEscalate = useCallback(async (alertId: string) => {
    const assignedTo = (escalateTarget[alertId] || '').trim();
    await handleAction(alertId, 'escalate', { assigned_to: assignedTo });
    setEscalateTarget(prev => ({ ...prev, [alertId]: '' }));
    setShowEscalateForm(null);
  }, [escalateTarget, handleAction]);

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    });
  };

  const formatDateTime = (dateStr: string) => {
    return new Date(dateStr).toLocaleString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: 'numeric',
      minute: '2-digit',
    });
  };

  const isDueSoon = (dueDate: string | null) => {
    if (!dueDate) return false;
    const diff = new Date(dueDate).getTime() - Date.now();
    return diff > 0 && diff < 48 * 60 * 60 * 1000;
  };

  const isOverdue = (dueDate: string | null) => {
    if (!dueDate) return false;
    return new Date(dueDate).getTime() < Date.now();
  };

  const summaryItems = [
    { label: 'Open', value: summary.open, color: 'text-blue-400', bg: 'bg-blue-500/10' },
    { label: 'In Progress', value: summary.in_progress, color: 'text-amber-400', bg: 'bg-amber-500/10' },
    { label: 'Escalated', value: summary.escalated, color: 'text-red-400', bg: 'bg-red-500/10' },
    { label: 'Critical Active', value: summary.critical_active, color: 'text-red-400', bg: 'bg-red-500/10' },
    { label: 'Resolved', value: summary.resolved, color: 'text-green-400', bg: 'bg-green-500/10' },
  ];

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-xl font-bold text-white">Alert Actions</h1>
        <p className="mt-1 text-xs text-gray-500">
          Trackable action items triggered by data quality anomalies, compliance deadlines, and system events
        </p>
      </div>

      {/* Summary Bar */}
      <div className="mb-6 grid grid-cols-5 gap-3">
        {summaryItems.map(item => (
          <div
            key={item.label}
            className="rounded-lg border border-gray-800 bg-gray-800/30 px-4 py-3"
          >
            <div className={`text-2xl font-bold ${item.color}`}>{item.value}</div>
            <div className="mt-0.5 text-xs text-gray-500">{item.label}</div>
          </div>
        ))}
      </div>

      {/* Filter Tabs */}
      <div className="mb-5 flex items-center gap-1 rounded-lg border border-gray-800 bg-gray-900/50 p-1">
        {filterTabs.map(tab => (
          <button
            key={tab.key}
            onClick={() => setActiveTab(tab.key)}
            className={`rounded-md px-4 py-1.5 text-xs font-medium transition ${
              activeTab === tab.key
                ? 'bg-primary-600 text-white'
                : 'text-gray-400 hover:bg-gray-800 hover:text-white'
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* Alert List */}
      {loading ? (
        <div className="py-20 text-center text-gray-500">Loading alerts...</div>
      ) : alerts.length === 0 ? (
        <div className="rounded-xl border border-gray-800 bg-gray-800/30 p-12 text-center">
          <div className="mb-3 text-3xl">&#9888;&#65039;</div>
          <h3 className="mb-2 text-base font-semibold text-white">No Alerts Found</h3>
          <p className="text-sm text-gray-400">
            {activeTab === 'all'
              ? 'There are no alert actions at this time. Alerts are generated automatically from data quality checks, compliance deadlines, and system events.'
              : `No alerts with status "${filterTabs.find(t => t.key === activeTab)?.label}" at this time.`}
          </p>
        </div>
      ) : (
        <div className="space-y-3">
          {alerts.map(alert => {
            const priority = priorityConfig[alert.priority] || priorityConfig.low;
            const status = statusConfig[alert.status] || statusConfig.open;
            const isExpanded = expandedComments === alert.id;
            const alertComments = comments[alert.id] || [];
            const isLoadingComments = commentsLoading === alert.id;
            const isActionLoading = actionLoading === alert.id;
            const isTerminal = alert.status === 'resolved' || alert.status === 'dismissed';

            return (
              <div
                key={alert.id}
                className="rounded-xl border border-gray-800 bg-gray-800/30"
              >
                <div className="p-5">
                  {/* Header Row */}
                  <div className="flex items-start justify-between gap-4">
                    <div className="min-w-0 flex-1">
                      <div className="flex flex-wrap items-center gap-2">
                        <span className={`inline-flex items-center gap-1.5 rounded px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide ${priority.bg} ${priority.text}`}>
                          <span className={`h-1.5 w-1.5 rounded-full ${priority.dot}`} />
                          {priority.label}
                        </span>
                        <span className={`rounded px-2 py-0.5 text-[10px] font-medium ${status.bg} ${status.text}`}>
                          {status.label}
                        </span>
                        <span className="rounded bg-gray-700/50 px-2 py-0.5 text-[10px] font-medium text-gray-300">
                          {sourceLabels[alert.source_type] || alert.source_type}
                        </span>
                      </div>
                      <h3 className="mt-2 text-sm font-semibold text-white">{alert.title}</h3>
                      <p className="mt-1 text-xs leading-relaxed text-gray-400">{alert.description}</p>
                    </div>
                  </div>

                  {/* Metadata Row */}
                  <div className="mt-3 flex flex-wrap items-center gap-x-4 gap-y-1 text-[11px] text-gray-500">
                    <span>Category: <span className="text-gray-400">{alert.category}</span></span>
                    <span className="text-gray-700">|</span>
                    <span>
                      Assigned:{' '}
                      <span className={alert.assigned_to ? 'text-gray-300' : 'text-gray-600'}>
                        {alert.assigned_to || 'Unassigned'}
                      </span>
                    </span>
                    <span className="text-gray-700">|</span>
                    <span>Created: <span className="text-gray-400">{formatDate(alert.created_at)}</span></span>
                    {alert.due_date && (
                      <>
                        <span className="text-gray-700">|</span>
                        <span>
                          Due:{' '}
                          <span
                            className={
                              isOverdue(alert.due_date)
                                ? 'font-medium text-red-400'
                                : isDueSoon(alert.due_date)
                                  ? 'font-medium text-amber-400'
                                  : 'text-gray-400'
                            }
                          >
                            {formatDate(alert.due_date)}
                            {isOverdue(alert.due_date) && ' (overdue)'}
                            {isDueSoon(alert.due_date) && !isOverdue(alert.due_date) && ' (due soon)'}
                          </span>
                        </span>
                      </>
                    )}
                  </div>

                  {/* Action Buttons */}
                  {!isTerminal && (
                    <div className="mt-4 flex flex-wrap items-center gap-2 border-t border-gray-800 pt-4">
                      {!alert.assigned_to && (
                        <button
                          onClick={() => handleAction(alert.id, 'assign')}
                          disabled={isActionLoading}
                          className="rounded-lg bg-primary-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-primary-500 disabled:opacity-50"
                        >
                          Take
                        </button>
                      )}
                      <button
                        onClick={() => handleAction(alert.id, 'resolve')}
                        disabled={isActionLoading}
                        className="rounded-lg bg-green-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-green-500 disabled:opacity-50"
                      >
                        Resolve
                      </button>
                      <button
                        onClick={() => handleAction(alert.id, 'dismiss')}
                        disabled={isActionLoading}
                        className="rounded-lg bg-gray-700 px-3 py-1.5 text-xs font-medium text-gray-300 hover:bg-gray-600 disabled:opacity-50"
                      >
                        Dismiss
                      </button>
                      <button
                        onClick={() => {
                          setShowEscalateForm(showEscalateForm === alert.id ? null : alert.id);
                          setShowCommentForm(null);
                        }}
                        disabled={isActionLoading}
                        className="rounded-lg bg-red-600/20 px-3 py-1.5 text-xs font-medium text-red-400 hover:bg-red-600/30 disabled:opacity-50"
                      >
                        Escalate
                      </button>
                      <button
                        onClick={() => {
                          setShowCommentForm(showCommentForm === alert.id ? null : alert.id);
                          setShowEscalateForm(null);
                        }}
                        disabled={isActionLoading}
                        className="rounded-lg bg-gray-700 px-3 py-1.5 text-xs font-medium text-gray-300 hover:bg-gray-600 disabled:opacity-50"
                      >
                        Comment
                      </button>
                    </div>
                  )}

                  {/* Escalate Form */}
                  {showEscalateForm === alert.id && (
                    <div className="mt-3 flex items-center gap-2">
                      <input
                        type="text"
                        placeholder="Assign to (email or name, leave blank for manager)"
                        value={escalateTarget[alert.id] || ''}
                        onChange={e =>
                          setEscalateTarget(prev => ({ ...prev, [alert.id]: e.target.value }))
                        }
                        className="flex-1 rounded-lg border border-gray-700 bg-gray-900 px-3 py-1.5 text-xs text-white placeholder-gray-600 focus:border-red-500 focus:outline-none focus:ring-1 focus:ring-red-500"
                      />
                      <button
                        onClick={() => submitEscalate(alert.id)}
                        disabled={isActionLoading}
                        className="rounded-lg bg-red-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-red-500 disabled:opacity-50"
                      >
                        Confirm Escalate
                      </button>
                      <button
                        onClick={() => setShowEscalateForm(null)}
                        className="rounded-lg px-3 py-1.5 text-xs text-gray-500 hover:text-gray-300"
                      >
                        Cancel
                      </button>
                    </div>
                  )}

                  {/* Comment Form */}
                  {showCommentForm === alert.id && (
                    <div className="mt-3 flex items-center gap-2">
                      <input
                        type="text"
                        placeholder="Write a comment..."
                        value={commentInputs[alert.id] || ''}
                        onChange={e =>
                          setCommentInputs(prev => ({ ...prev, [alert.id]: e.target.value }))
                        }
                        onKeyDown={e => {
                          if (e.key === 'Enter' && !e.shiftKey) {
                            e.preventDefault();
                            submitComment(alert.id);
                          }
                        }}
                        className="flex-1 rounded-lg border border-gray-700 bg-gray-900 px-3 py-1.5 text-xs text-white placeholder-gray-600 focus:border-primary-500 focus:outline-none focus:ring-1 focus:ring-primary-500"
                      />
                      <button
                        onClick={() => submitComment(alert.id)}
                        disabled={isActionLoading || !(commentInputs[alert.id] || '').trim()}
                        className="rounded-lg bg-primary-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-primary-500 disabled:opacity-50"
                      >
                        Send
                      </button>
                      <button
                        onClick={() => setShowCommentForm(null)}
                        className="rounded-lg px-3 py-1.5 text-xs text-gray-500 hover:text-gray-300"
                      >
                        Cancel
                      </button>
                    </div>
                  )}

                  {/* Comments Toggle */}
                  <div className="mt-3">
                    <button
                      onClick={() => toggleComments(alert.id)}
                      className="text-[11px] font-medium text-gray-500 hover:text-gray-300"
                    >
                      {isExpanded ? 'Hide Comments' : 'View Comments'}
                    </button>
                  </div>
                </div>

                {/* Expanded Comments Section */}
                {isExpanded && (
                  <div className="border-t border-gray-800 bg-gray-900/40 px-5 py-4">
                    {isLoadingComments ? (
                      <div className="py-4 text-center text-xs text-gray-500">Loading comments...</div>
                    ) : alertComments.length === 0 ? (
                      <div className="py-4 text-center text-xs text-gray-600">No comments yet.</div>
                    ) : (
                      <div className="space-y-3">
                        {alertComments.map(comment => (
                          <div key={comment.id} className="flex gap-3">
                            <div className="flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full bg-gray-700 text-[10px] font-bold text-gray-300">
                              {comment.author
                                .split(' ')
                                .map(n => n[0])
                                .join('')
                                .slice(0, 2)
                                .toUpperCase()}
                            </div>
                            <div className="min-w-0 flex-1">
                              <div className="flex items-center gap-2">
                                <span className="text-xs font-medium text-gray-300">{comment.author}</span>
                                <span className="text-[10px] text-gray-600">{formatDateTime(comment.created_at)}</span>
                              </div>
                              <p className="mt-0.5 text-xs leading-relaxed text-gray-400">{comment.content}</p>
                            </div>
                          </div>
                        ))}
                      </div>
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
