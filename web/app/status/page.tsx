'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';

interface ServiceStatus {
  name: string;
  status: 'operational' | 'degraded' | 'down' | 'checking';
  latency?: number;
  lastChecked?: string;
}

const API_URL = process.env.NEXT_PUBLIC_OFFGRIDFLOW_API_URL || 'https://offgridflow-api-v2-production.up.railway.app';

export default function StatusPage() {
  const [services, setServices] = useState<ServiceStatus[]>([
    { name: 'API Server', status: 'checking' },
    { name: 'Database', status: 'checking' },
    { name: 'Web Application', status: 'operational', lastChecked: new Date().toISOString() },
  ]);

  useEffect(() => {
    const checkHealth = async () => {
      const start = Date.now();
      try {
        const res = await fetch(`${API_URL}/health`, { signal: AbortSignal.timeout(10000) });
        const latency = Date.now() - start;
        const data = await res.json();

        setServices(prev => prev.map(s => {
          if (s.name === 'API Server') {
            return { ...s, status: res.ok ? 'operational' : 'degraded', latency, lastChecked: new Date().toISOString() };
          }
          if (s.name === 'Database') {
            return { ...s, status: data.status === 'ok' ? 'operational' : 'degraded', lastChecked: new Date().toISOString() };
          }
          return s;
        }));
      } catch {
        setServices(prev => prev.map(s =>
          s.name === 'API Server' || s.name === 'Database'
            ? { ...s, status: 'down', lastChecked: new Date().toISOString() }
            : s
        ));
      }
    };

    checkHealth();
    const interval = setInterval(checkHealth, 30000);
    return () => clearInterval(interval);
  }, []);

  const allOperational = services.every(s => s.status === 'operational');

  return (
    <div className="min-h-screen bg-dark-900 text-gray-100">
      <nav className="border-b border-gray-800/50 bg-dark-900/80 backdrop-blur-md">
        <div className="mx-auto flex max-w-3xl items-center justify-between px-6 py-4">
          <Link href="/" className="text-xl font-bold text-white">OffGridFlow</Link>
          <span className="text-xs text-gray-500">System Status</span>
        </div>
      </nav>

      <div className="mx-auto max-w-3xl px-6 py-16">
        {/* Overall status */}
        <div className={`mb-10 rounded-xl border p-6 text-center ${
          allOperational ? 'border-green-500/20 bg-green-500/5' : 'border-amber-500/20 bg-amber-500/5'
        }`}>
          <div className={`mb-2 text-lg font-bold ${allOperational ? 'text-green-400' : 'text-amber-400'}`}>
            {allOperational ? 'All Systems Operational' : 'Experiencing Issues'}
          </div>
          <div className="text-xs text-gray-500">
            Last checked: {new Date().toLocaleString()}
          </div>
        </div>

        {/* Service list */}
        <div className="space-y-3">
          {services.map(service => (
            <div key={service.name} className="flex items-center justify-between rounded-lg border border-gray-800 bg-gray-800/30 px-5 py-4">
              <div className="flex items-center gap-3">
                <div className={`h-2.5 w-2.5 rounded-full ${
                  service.status === 'operational' ? 'bg-green-500' :
                  service.status === 'degraded' ? 'bg-amber-500' :
                  service.status === 'down' ? 'bg-red-500' : 'bg-gray-500 animate-pulse'
                }`} />
                <span className="text-sm font-medium text-white">{service.name}</span>
              </div>
              <div className="flex items-center gap-4">
                {service.latency !== undefined && (
                  <span className="text-xs text-gray-500">{service.latency}ms</span>
                )}
                <span className={`text-xs capitalize ${
                  service.status === 'operational' ? 'text-green-400' :
                  service.status === 'degraded' ? 'text-amber-400' :
                  service.status === 'down' ? 'text-red-400' : 'text-gray-500'
                }`}>
                  {service.status}
                </span>
              </div>
            </div>
          ))}
        </div>

        {/* SLOs */}
        <div className="mt-10">
          <h2 className="mb-4 text-sm font-semibold text-gray-400">Service Level Objectives</h2>
          <div className="space-y-3">
            {[
              { metric: 'API Availability', target: '99.9%', window: 'Monthly' },
              { metric: 'API Response Time (p95)', target: '< 500ms', window: 'Hourly' },
              { metric: 'Report Generation', target: '< 30 seconds', window: 'Per request' },
              { metric: 'Data Ingestion', target: '< 60 seconds', window: 'Per upload' },
              { metric: 'Planned Maintenance Window', target: 'Sundays 02:00–06:00 UTC', window: 'Weekly' },
            ].map(slo => (
              <div key={slo.metric} className="flex items-center justify-between rounded-lg border border-gray-800 bg-gray-900/30 px-5 py-3">
                <span className="text-sm text-gray-300">{slo.metric}</span>
                <div className="flex items-center gap-4">
                  <span className="text-xs text-gray-500">{slo.window}</span>
                  <span className="rounded bg-gray-800 px-2 py-0.5 text-xs font-medium text-white">{slo.target}</span>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Incident communication */}
        <div className="mt-10">
          <h2 className="mb-4 text-sm font-semibold text-gray-400">Incident Communication</h2>
          <div className="rounded-lg border border-gray-800 bg-gray-900/30 p-5">
            <p className="text-sm text-gray-400">
              For service disruptions, contact{' '}
              <a href="mailto:contact@off-grid-flow.com" className="text-primary-400 hover:underline">
                contact@off-grid-flow.com
              </a>
              . Critical incidents are communicated via email to all affected customers within 30 minutes of detection.
            </p>
            <div className="mt-4 space-y-2 text-xs text-gray-500">
              <div>Response time for P1 incidents: &lt; 30 minutes</div>
              <div>Response time for P2 incidents: &lt; 4 hours</div>
              <div>Post-incident review: within 72 hours</div>
            </div>
          </div>
        </div>

        <div className="mt-10 text-center text-xs text-gray-600">
          <Link href="/" className="hover:text-gray-400">Back to OffGridFlow</Link>
        </div>
      </div>
    </div>
  );
}
