'use client';

import React, { useEffect, useState } from 'react';
import { cn } from '@/lib/utils';

// ============================================================
// Types
// ============================================================

type ServiceStatus = 'operational' | 'degraded' | 'outage' | 'maintenance';

interface ServiceHealth {
  name: string;
  description: string;
  status: ServiceStatus;
  latency?: number;
  lastChecked: Date;
  uptime: number; // percentage
}

interface Incident {
  id: string;
  title: string;
  status: 'investigating' | 'identified' | 'monitoring' | 'resolved';
  severity: 'minor' | 'major' | 'critical';
  createdAt: Date;
  updatedAt: Date;
  affectedServices: string[];
  updates: IncidentUpdate[];
}

interface IncidentUpdate {
  timestamp: Date;
  status: string;
  message: string;
}

interface MaintenanceWindow {
  id: string;
  title: string;
  description: string;
  scheduledStart: Date;
  scheduledEnd: Date;
  affectedServices: string[];
}

interface HistoricalUptime {
  date: Date;
  uptime: number;
  incidents: number;
}

// ============================================================
// Status Configuration
// ============================================================

const STATUS_CONFIG: Record<ServiceStatus, { label: string; color: string; bgColor: string; icon: string }> = {
  operational: {
    label: 'Operational',
    color: 'text-green-600 dark:text-green-400',
    bgColor: 'bg-green-100 dark:bg-green-900/30',
    icon: '✓',
  },
  degraded: {
    label: 'Degraded Performance',
    color: 'text-yellow-600 dark:text-yellow-400',
    bgColor: 'bg-yellow-100 dark:bg-yellow-900/30',
    icon: '⚠',
  },
  outage: {
    label: 'Major Outage',
    color: 'text-red-600 dark:text-red-400',
    bgColor: 'bg-red-100 dark:bg-red-900/30',
    icon: '✕',
  },
  maintenance: {
    label: 'Under Maintenance',
    color: 'text-blue-600 dark:text-blue-400',
    bgColor: 'bg-blue-100 dark:bg-blue-900/30',
    icon: '🔧',
  },
};

const SEVERITY_CONFIG = {
  minor: { color: 'text-yellow-600', bgColor: 'bg-yellow-100 dark:bg-yellow-900/30' },
  major: { color: 'text-orange-600', bgColor: 'bg-orange-100 dark:bg-orange-900/30' },
  critical: { color: 'text-red-600', bgColor: 'bg-red-100 dark:bg-red-900/30' },
};

// ============================================================
// Components
// ============================================================

function StatusIndicator({ status }: { status: ServiceStatus }) {
  const config = STATUS_CONFIG[status];
  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-sm font-medium',
        config.color,
        config.bgColor
      )}
    >
      <span className="text-xs">{config.icon}</span>
      {config.label}
    </span>
  );
}

function OverallStatus({ services }: { services: ServiceHealth[] }) {
  const getOverallStatus = (): ServiceStatus => {
    if (services.some(s => s.status === 'outage')) return 'outage';
    if (services.some(s => s.status === 'maintenance')) return 'maintenance';
    if (services.some(s => s.status === 'degraded')) return 'degraded';
    return 'operational';
  };

  const status = getOverallStatus();
  const config = STATUS_CONFIG[status];

  return (
    <div
      className={cn(
        'rounded-xl p-6 mb-8 border-2',
        status === 'operational' && 'border-green-200 dark:border-green-800 bg-green-50 dark:bg-green-950/20',
        status === 'degraded' && 'border-yellow-200 dark:border-yellow-800 bg-yellow-50 dark:bg-yellow-950/20',
        status === 'outage' && 'border-red-200 dark:border-red-800 bg-red-50 dark:bg-red-950/20',
        status === 'maintenance' && 'border-blue-200 dark:border-blue-800 bg-blue-50 dark:bg-blue-950/20'
      )}
    >
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-gray-900 dark:text-white">
            {status === 'operational' && 'All Systems Operational'}
            {status === 'degraded' && 'Experiencing Degraded Performance'}
            {status === 'outage' && 'Service Disruption'}
            {status === 'maintenance' && 'Scheduled Maintenance'}
          </h2>
          <p className="text-gray-600 dark:text-gray-400 mt-1">
            Last updated: {new Date().toLocaleString()}
          </p>
        </div>
        <div className={cn('text-6xl', config.color)}>{config.icon}</div>
      </div>
    </div>
  );
}

function ServiceCard({ service }: { service: ServiceHealth }) {
  const config = STATUS_CONFIG[service.status];

  return (
    <div className="flex items-center justify-between py-4 border-b border-gray-200 dark:border-gray-700 last:border-0">
      <div className="flex-1">
        <h3 className="font-medium text-gray-900 dark:text-white">{service.name}</h3>
        <p className="text-sm text-gray-500 dark:text-gray-400">{service.description}</p>
        {service.latency && (
          <p className="text-xs text-gray-400 dark:text-gray-500 mt-1">
            Response time: {service.latency}ms
          </p>
        )}
      </div>
      <div className="flex items-center gap-4">
        <div className="text-right">
          <div className="text-sm font-medium text-gray-900 dark:text-white">
            {service.uptime.toFixed(2)}%
          </div>
          <div className="text-xs text-gray-500">30-day uptime</div>
        </div>
        <StatusIndicator status={service.status} />
      </div>
    </div>
  );
}

function IncidentCard({ incident }: { incident: Incident }) {
  const [expanded, setExpanded] = useState(false);
  const severityConfig = SEVERITY_CONFIG[incident.severity];

  return (
    <div
      className={cn(
        'rounded-lg border p-4 mb-4',
        incident.status === 'resolved'
          ? 'border-gray-200 dark:border-gray-700'
          : 'border-orange-200 dark:border-orange-800'
      )}
    >
      <div className="flex items-start justify-between">
        <div className="flex-1">
          <div className="flex items-center gap-2 mb-2">
            <span
              className={cn(
                'px-2 py-0.5 rounded text-xs font-medium',
                severityConfig.bgColor,
                severityConfig.color
              )}
            >
              {incident.severity.toUpperCase()}
            </span>
            <span className="text-xs text-gray-500 dark:text-gray-400">
              {incident.status === 'resolved' ? 'Resolved' : 'Active'}
            </span>
          </div>
          <h3 className="font-medium text-gray-900 dark:text-white">{incident.title}</h3>
          <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
            Affected: {incident.affectedServices.join(', ')}
          </p>
          <p className="text-xs text-gray-400 mt-1">
            Started: {incident.createdAt.toLocaleString()}
          </p>
        </div>
        <button
          onClick={() => setExpanded(!expanded)}
          className="text-sm text-blue-600 hover:text-blue-700 dark:text-blue-400"
        >
          {expanded ? 'Hide updates' : 'Show updates'}
        </button>
      </div>

      {expanded && incident.updates.length > 0 && (
        <div className="mt-4 pl-4 border-l-2 border-gray-200 dark:border-gray-700">
          {incident.updates.map((update, idx) => (
            <div key={idx} className="mb-3 last:mb-0">
              <div className="flex items-center gap-2">
                <span className="text-xs font-medium text-gray-600 dark:text-gray-400">
                  {update.status}
                </span>
                <span className="text-xs text-gray-400">
                  {update.timestamp.toLocaleString()}
                </span>
              </div>
              <p className="text-sm text-gray-700 dark:text-gray-300 mt-1">{update.message}</p>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

function MaintenanceCard({ maintenance }: { maintenance: MaintenanceWindow }) {
  return (
    <div className="rounded-lg border border-blue-200 dark:border-blue-800 bg-blue-50 dark:bg-blue-950/20 p-4 mb-4">
      <div className="flex items-center gap-2 mb-2">
        <span className="text-blue-600 dark:text-blue-400">🔧</span>
        <span className="text-sm font-medium text-blue-700 dark:text-blue-300">
          Scheduled Maintenance
        </span>
      </div>
      <h3 className="font-medium text-gray-900 dark:text-white">{maintenance.title}</h3>
      <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">{maintenance.description}</p>
      <div className="flex items-center gap-4 mt-2 text-xs text-gray-500">
        <span>Start: {maintenance.scheduledStart.toLocaleString()}</span>
        <span>End: {maintenance.scheduledEnd.toLocaleString()}</span>
      </div>
      <p className="text-xs text-gray-400 mt-2">
        Affected: {maintenance.affectedServices.join(', ')}
      </p>
    </div>
  );
}

function UptimeGraph({ history }: { history: HistoricalUptime[] }) {
  // Show last 90 days as small bars
  const displayDays = history.slice(-90);

  return (
    <div className="mt-8">
      <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
        90-Day Uptime History
      </h3>
      <div className="flex gap-0.5">
        {displayDays.map((day, idx) => {
          let bgColor = 'bg-green-500';
          if (day.uptime < 99.9) bgColor = 'bg-yellow-500';
          if (day.uptime < 99) bgColor = 'bg-orange-500';
          if (day.uptime < 95) bgColor = 'bg-red-500';

          return (
            <div
              key={idx}
              className={cn('w-2 h-8 rounded-sm', bgColor)}
              title={`${day.date.toLocaleDateString()}: ${day.uptime.toFixed(2)}% uptime${
                day.incidents > 0 ? `, ${day.incidents} incident(s)` : ''
              }`}
            />
          );
        })}
      </div>
      <div className="flex justify-between mt-2 text-xs text-gray-500">
        <span>90 days ago</span>
        <span>Today</span>
      </div>
      <div className="flex items-center gap-4 mt-4 text-xs">
        <div className="flex items-center gap-1">
          <div className="w-3 h-3 rounded bg-green-500" />
          <span className="text-gray-600 dark:text-gray-400">100%</span>
        </div>
        <div className="flex items-center gap-1">
          <div className="w-3 h-3 rounded bg-yellow-500" />
          <span className="text-gray-600 dark:text-gray-400">99-99.9%</span>
        </div>
        <div className="flex items-center gap-1">
          <div className="w-3 h-3 rounded bg-orange-500" />
          <span className="text-gray-600 dark:text-gray-400">95-99%</span>
        </div>
        <div className="flex items-center gap-1">
          <div className="w-3 h-3 rounded bg-red-500" />
          <span className="text-gray-600 dark:text-gray-400">&lt;95%</span>
        </div>
      </div>
    </div>
  );
}

function SubscribeForm() {
  const [email, setEmail] = useState('');
  const [subscribed, setSubscribed] = useState(false);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    // In production, this would call an API
    setSubscribed(true);
    setEmail('');
  };

  return (
    <div className="mt-8 p-6 rounded-xl bg-gray-100 dark:bg-gray-800">
      <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">
        Subscribe to Updates
      </h3>
      <p className="text-sm text-gray-600 dark:text-gray-400 mb-4">
        Get notified when we have scheduled maintenance or experience issues.
      </p>
      {subscribed ? (
        <div className="flex items-center gap-2 text-green-600 dark:text-green-400">
          <span>✓</span>
          <span>You're subscribed to status updates!</span>
        </div>
      ) : (
        <form onSubmit={handleSubmit} className="flex gap-2">
          <input
            type="email"
            value={email}
            onChange={e => setEmail(e.target.value)}
            placeholder="Enter your email"
            required
            className="flex-1 px-4 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-900 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          />
          <button
            type="submit"
            className="px-6 py-2 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-lg transition-colors"
          >
            Subscribe
          </button>
        </form>
      )}
    </div>
  );
}

// ============================================================
// Main Component
// ============================================================

interface StatusPageProps {
  apiEndpoint?: string;
  refreshInterval?: number; // milliseconds
}

export default function StatusPage({
  apiEndpoint = '/api/status',
  refreshInterval = 60000, // 1 minute
}: StatusPageProps) {
  const [services, setServices] = useState<ServiceHealth[]>([]);
  const [incidents, setIncidents] = useState<Incident[]>([]);
  const [maintenance, setMaintenance] = useState<MaintenanceWindow[]>([]);
  const [history, setHistory] = useState<HistoricalUptime[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchStatus = async () => {
      try {
        // In production, fetch from API
        // const response = await fetch(apiEndpoint);
        // const data = await response.json();

        // Demo data for development
        setServices([
          {
            name: 'API',
            description: 'Core REST and GraphQL APIs',
            status: 'operational',
            latency: 45,
            lastChecked: new Date(),
            uptime: 99.98,
          },
          {
            name: 'Dashboard',
            description: 'Web application and user interface',
            status: 'operational',
            latency: 120,
            lastChecked: new Date(),
            uptime: 99.99,
          },
          {
            name: 'Data Ingestion',
            description: 'Connectors for AWS, Azure, GCP, SAP',
            status: 'operational',
            latency: 230,
            lastChecked: new Date(),
            uptime: 99.95,
          },
          {
            name: 'Calculation Engine',
            description: 'Emissions calculation and reporting',
            status: 'operational',
            latency: 85,
            lastChecked: new Date(),
            uptime: 99.99,
          },
          {
            name: 'Background Jobs',
            description: 'Report generation and scheduled tasks',
            status: 'operational',
            latency: undefined,
            lastChecked: new Date(),
            uptime: 99.97,
          },
        ]);

        setIncidents([
          {
            id: '1',
            title: 'Elevated API latency',
            status: 'resolved',
            severity: 'minor',
            createdAt: new Date(Date.now() - 3 * 24 * 60 * 60 * 1000),
            updatedAt: new Date(Date.now() - 3 * 24 * 60 * 60 * 1000 + 2 * 60 * 60 * 1000),
            affectedServices: ['API'],
            updates: [
              {
                timestamp: new Date(Date.now() - 3 * 24 * 60 * 60 * 1000),
                status: 'Investigating',
                message: 'We are investigating reports of increased API response times.',
              },
              {
                timestamp: new Date(Date.now() - 3 * 24 * 60 * 60 * 1000 + 30 * 60 * 1000),
                status: 'Identified',
                message: 'The issue has been identified as a database connection pool exhaustion.',
              },
              {
                timestamp: new Date(Date.now() - 3 * 24 * 60 * 60 * 1000 + 2 * 60 * 60 * 1000),
                status: 'Resolved',
                message:
                  'The issue has been resolved. Connection pool limits have been increased.',
              },
            ],
          },
        ]);

        setMaintenance([]);

        // Generate 90 days of uptime history
        const historyData: HistoricalUptime[] = [];
        for (let i = 89; i >= 0; i--) {
          const date = new Date();
          date.setDate(date.getDate() - i);
          historyData.push({
            date,
            uptime: 99.5 + Math.random() * 0.5,
            incidents: Math.random() > 0.95 ? 1 : 0,
          });
        }
        setHistory(historyData);

        setLoading(false);
      } catch (err) {
        setError('Failed to load status information');
        setLoading(false);
      }
    };

    fetchStatus();
    const interval = setInterval(fetchStatus, refreshInterval);
    return () => clearInterval(interval);
  }, [apiEndpoint, refreshInterval]);

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 dark:bg-gray-900 flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-gray-50 dark:bg-gray-900 flex items-center justify-center">
        <div className="text-center">
          <div className="text-6xl mb-4">⚠️</div>
          <h2 className="text-xl font-semibold text-gray-900 dark:text-white">{error}</h2>
          <p className="text-gray-600 dark:text-gray-400 mt-2">Please try again later.</p>
        </div>
      </div>
    );
  }

  const activeIncidents = incidents.filter(i => i.status !== 'resolved');
  const resolvedIncidents = incidents.filter(i => i.status === 'resolved');

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      <div className="max-w-4xl mx-auto px-4 py-12">
        {/* Header */}
        <header className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
            OffGridFlow Status
          </h1>
          <p className="text-gray-600 dark:text-gray-400 mt-2">
            Real-time system status and incident history
          </p>
        </header>

        {/* Overall Status */}
        <OverallStatus services={services} />

        {/* Scheduled Maintenance */}
        {maintenance.length > 0 && (
          <section className="mb-8">
            <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-4">
              Scheduled Maintenance
            </h2>
            {maintenance.map(m => (
              <MaintenanceCard key={m.id} maintenance={m} />
            ))}
          </section>
        )}

        {/* Active Incidents */}
        {activeIncidents.length > 0 && (
          <section className="mb-8">
            <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-4">
              Active Incidents
            </h2>
            {activeIncidents.map(incident => (
              <IncidentCard key={incident.id} incident={incident} />
            ))}
          </section>
        )}

        {/* Services */}
        <section className="mb-8">
          <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-4">
            System Components
          </h2>
          <div className="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-4">
            {services.map(service => (
              <ServiceCard key={service.name} service={service} />
            ))}
          </div>
        </section>

        {/* Uptime History */}
        <section className="mb-8">
          <UptimeGraph history={history} />
        </section>

        {/* Past Incidents */}
        {resolvedIncidents.length > 0 && (
          <section className="mb-8">
            <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-4">
              Past Incidents
            </h2>
            {resolvedIncidents.map(incident => (
              <IncidentCard key={incident.id} incident={incident} />
            ))}
          </section>
        )}

        {/* Subscribe */}
        <SubscribeForm />

        {/* Footer */}
        <footer className="mt-12 text-center text-sm text-gray-500 dark:text-gray-400">
          <p>
            Need support?{' '}
            <a href="/support" className="text-blue-600 hover:text-blue-700 dark:text-blue-400">
              Contact us
            </a>
          </p>
          <p className="mt-2">
            © {new Date().getFullYear()} OffGridFlow. All rights reserved.
          </p>
        </footer>
      </div>
    </div>
  );
}
