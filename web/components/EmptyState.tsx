'use client';

import { ReactNode } from 'react';

interface EmptyStateProps {
  /** Icon to display (default: 📭) */
  icon?: ReactNode;
  /** Heading text */
  title: string;
  /** Descriptive text */
  description?: string;
  /** Call-to-action button */
  action?: ReactNode;
  /** Compact variant for cards/sidebars */
  compact?: boolean;
}

/**
 * Polished empty state component for lists, tables, and dashboard widgets.
 * Use when there's no data to display rather than showing nothing or a raw message.
 */
export default function EmptyState({
  icon = '📭',
  title,
  description,
  action,
  compact = false,
}: EmptyStateProps) {
  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        padding: compact ? '2rem 1rem' : '4rem 2rem',
        textAlign: 'center',
        background: 'linear-gradient(135deg, #0f172a 0%, #1e293b 100%)',
        borderRadius: '12px',
        border: '1px solid rgba(148, 163, 184, 0.1)',
        minHeight: compact ? 'auto' : '240px',
      }}
    >
      <div
        style={{
          fontSize: compact ? '2.5rem' : '3.5rem',
          marginBottom: compact ? '0.75rem' : '1rem',
          filter: 'grayscale(0.2)',
        }}
        aria-hidden="true"
      >
        {icon}
      </div>

      <h3
        style={{
          fontSize: compact ? '1rem' : '1.25rem',
          fontWeight: 600,
          color: '#e2e8f0',
          margin: 0,
          marginBottom: description ? '0.5rem' : action ? '1rem' : 0,
        }}
      >
        {title}
      </h3>

      {description && (
        <p
          style={{
            fontSize: compact ? '0.85rem' : '0.95rem',
            color: '#94a3b8',
            margin: 0,
            marginBottom: action ? '1.25rem' : 0,
            maxWidth: '320px',
            lineHeight: 1.5,
          }}
        >
          {description}
        </p>
      )}

      {action}
    </div>
  );
}

/**
 * Pre-configured empty states for common scenarios
 */
export const EmptyStates = {
  NoActivities: (
    <EmptyState
      icon="📊"
      title="No activities yet"
      description="Start tracking your emissions by adding your first activity."
    />
  ),
  NoEmissions: (
    <EmptyState
      icon="🌱"
      title="No emissions recorded"
      description="Once you add activities, emissions will be calculated automatically."
    />
  ),
  NoReports: (
    <EmptyState
      icon="📋"
      title="No compliance reports"
      description="Generate your first report to see it here."
    />
  ),
  NoConnectors: (
    <EmptyState
      icon="🔌"
      title="No data sources connected"
      description="Connect cloud providers or ERP systems to import data automatically."
    />
  ),
  SearchNoResults: (
    <EmptyState
      icon="🔍"
      title="No results found"
      description="Try adjusting your search or filters."
      compact
    />
  ),
  LoadingError: (
    <EmptyState
      icon="⚠️"
      title="Unable to load data"
      description="Please try again later or contact support if the issue persists."
    />
  ),
};
