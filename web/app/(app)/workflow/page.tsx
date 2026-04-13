'use client';

import { useCallback, useEffect, useMemo, useState } from 'react';
import { api } from '@/lib/api';
import styles from './page.module.css';

type WorkflowTask = {
  id: string;
  title: string;
  assignee: string;
  status: string;
  createdAt?: string;
  dueAt?: string;
  metadata?: Record<string, string>;
};

const statusOptions = ['pending', 'in_progress', 'done'];

export default function WorkflowPage() {
  const [tasks, setTasks] = useState<WorkflowTask[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [toast, setToast] = useState<string | null>(null);

  const [title, setTitle] = useState('');
  const [assignee, setAssignee] = useState('');
  const [dueAt, setDueAt] = useState('');

  const loadTasks = useMemo(
    () => async () => {
      setLoading(true);
      setError(null);
      try {
        const data = await api.get<WorkflowTask[]>('/api/workflow/tasks');
        setTasks(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load tasks');
      } finally {
        setLoading(false);
      }
    },
    [],
  );

  useEffect(() => {
    void loadTasks();
  }, [loadTasks]);

  const createTask = useCallback(async () => {
    if (!title.trim()) return;
    setLoading(true);
    setError(null);
    try {
      await api.post('/api/workflow/tasks', {
        title: title.trim(),
        assignee: assignee.trim(),
        due_at: dueAt ? new Date(dueAt).toISOString() : undefined,
      });
      setTitle('');
      setAssignee('');
      setDueAt('');
      setToast('Task created');
      await loadTasks();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create task');
    } finally {
      setLoading(false);
    }
  }, [title, assignee, dueAt, loadTasks]);

  const updateStatus = useCallback(
    async (id: string, status: string) => {
      setLoading(true);
      setError(null);
      try {
        await api.patch('/api/workflow/tasks', { id, status });
        setToast('Task updated');
        await loadTasks();
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to update task');
      } finally {
        setLoading(false);
      }
    },
    [loadTasks],
  );

  return (
    <div className={styles.container}>
      <div className={styles.headerRow}>
        <div>
          <p className={styles.eyebrow}>Workflow</p>
          <h1 className={styles.title}>Operational Tasks</h1>
          <p className={styles.muted}>
            Track compliance and ingestion tasks. Create quick tasks and update status without leaving OffGridFlow.
          </p>
        </div>
        <button onClick={loadTasks} className={styles.secondaryButton} disabled={loading}>
          {loading ? 'Refreshing…' : 'Refresh'}
        </button>
      </div>

      {toast && <div className={styles.toast}>{toast}</div>}
      {error && <div className={styles.error}>{error}</div>}

      <div className={styles.card}>
        <h2 className={styles.cardTitle}>New task</h2>
        <div className={styles.formGrid}>
          <label className={styles.label}>
            Title
            <input
              className={styles.input}
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="Re-run ingestion, review CSRD data…"
            />
          </label>
          <label className={styles.label}>
            Assignee
            <input
              className={styles.input}
              value={assignee}
              onChange={(e) => setAssignee(e.target.value)}
              placeholder="owner@example.com"
            />
          </label>
          <label className={styles.label}>
            Due date
            <input
              className={styles.input}
              type="date"
              value={dueAt}
              onChange={(e) => setDueAt(e.target.value)}
            />
          </label>
        </div>
        <button className={styles.primaryButton} onClick={createTask} disabled={loading || !title.trim()}>
          Create task
        </button>
      </div>

      <div className={styles.tableWrapper}>
        <table className={styles.table}>
          <thead>
            <tr>
              <th>Task</th>
              <th>Assignee</th>
              <th>Status</th>
              <th>Due</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {tasks.map((t) => (
              <tr key={t.id}>
                <td>
                  <div className={styles.taskTitle}>{t.title}</div>
                  <div className={styles.meta}>
                    Created {t.createdAt ? new Date(t.createdAt).toLocaleDateString() : '—'}
                  </div>
                </td>
                <td>{t.assignee || 'Unassigned'}</td>
                <td>
                  <span className={`${styles.badge} ${statusClass(t.status)}`}>{t.status}</span>
                </td>
                <td>{t.dueAt ? new Date(t.dueAt).toLocaleDateString() : '—'}</td>
                <td className={styles.actionsCell}>
                  {statusOptions.map((s) => (
                    <button
                      key={s}
                      onClick={() => updateStatus(t.id, s)}
                      className={styles.secondaryButton}
                      disabled={loading || t.status === s}
                    >
                      {s.replace('_', ' ')}
                    </button>
                  ))}
                </td>
              </tr>
            ))}
            {tasks.length === 0 && (
              <tr>
                <td colSpan={5} className={styles.muted}>
                  No tasks yet.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function statusClass(status: string) {
  switch (status) {
    case 'done':
      return styles.badgeSuccess;
    case 'in_progress':
      return styles.badgeInfo;
    case 'pending':
    default:
      return styles.badgeMuted;
  }
}
