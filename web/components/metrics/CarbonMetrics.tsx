'use client';

import React, { memo } from 'react';
import { CarbonMetrics as CarbonMetricsType } from '@/stores/carbonStore';

interface MetricCardProps {
  id: string;
  label: string;
  value: string | number;
  unit?: string;
  trend?: 'up' | 'down' | 'stable';
  trendValue?: number;
  icon: React.ReactNode;
  onClick?: (id: string, value: number) => void;
  color?: 'green' | 'blue' | 'orange' | 'red' | 'purple';
}

const colorClasses = {
  green: 'from-green-500/20 to-green-600/5 border-green-500/30',
  blue: 'from-blue-500/20 to-blue-600/5 border-blue-500/30',
  orange: 'from-orange-500/20 to-orange-600/5 border-orange-500/30',
  red: 'from-red-500/20 to-red-600/5 border-red-500/30',
  purple: 'from-purple-500/20 to-purple-600/5 border-purple-500/30',
};

const iconColorClasses = {
  green: 'bg-green-500/20 text-green-400',
  blue: 'bg-blue-500/20 text-blue-400',
  orange: 'bg-orange-500/20 text-orange-400',
  red: 'bg-red-500/20 text-red-400',
  purple: 'bg-purple-500/20 text-purple-400',
};

const MetricCard = memo(function MetricCard({
  id,
  label,
  value,
  unit,
  trend,
  trendValue,
  icon,
  onClick,
  color = 'green',
}: MetricCardProps) {
  const handleClick = () => {
    if (onClick) {
      onClick(id, typeof value === 'number' ? value : parseFloat(String(value)) || 0);
    }
  };

  return (
    <button
      onClick={handleClick}
      className={`
        w-full p-5 rounded-xl border bg-gradient-to-br transition-all duration-200
        hover:scale-[1.02] hover:shadow-lg hover:shadow-black/20
        focus:outline-none focus:ring-2 focus:ring-green-500/50
        ${colorClasses[color]}
      `}
    >
      <div className="flex items-start justify-between mb-3">
        <div className={`p-2.5 rounded-lg ${iconColorClasses[color]}`}>
          {icon}
        </div>
        {trend && trendValue !== undefined && (
          <div
            className={`flex items-center gap-1 text-sm font-medium ${
              trend === 'down' ? 'text-green-400' : trend === 'up' ? 'text-red-400' : 'text-gray-400'
            }`}
          >
            {trend === 'down' ? (
              <TrendDownIcon />
            ) : trend === 'up' ? (
              <TrendUpIcon />
            ) : (
              <TrendStableIcon />
            )}
            <span>{Math.abs(trendValue).toFixed(1)}%</span>
          </div>
        )}
      </div>
      
      <div className="text-left">
        <div className="text-2xl font-bold text-white mb-1">
          {typeof value === 'number' ? value.toLocaleString(undefined, { maximumFractionDigits: 1 }) : value}
          {unit && <span className="text-sm font-normal text-gray-400 ml-1">{unit}</span>}
        </div>
        <div className="text-sm text-gray-400">{label}</div>
      </div>
    </button>
  );
});

interface CarbonMetricsProps {
  metrics: CarbonMetricsType;
  onMetricClick?: (metricId: string, value: number) => void;
}

export const CarbonMetrics = memo(function CarbonMetrics({
  metrics,
  onMetricClick,
}: CarbonMetricsProps) {
  return (
    <div className="space-y-4">
      <h3 className="text-sm font-semibold text-gray-400 uppercase tracking-wider mb-4">
        Key Metrics
      </h3>
      
      <div className="grid gap-4">
        <MetricCard
          id="carbon-intensity"
          label="Carbon Intensity"
          value={metrics.carbonIntensity}
          unit="tCO₂e/$M"
          trend={metrics.benchmarkComparison < 0 ? 'down' : 'up'}
          trendValue={metrics.benchmarkComparison}
          color="green"
          icon={<CarbonIcon />}
          onClick={onMetricClick}
        />
        
        <MetricCard
          id="revenue"
          label="Revenue Base"
          value={`$${(metrics.revenue / 1000000).toFixed(1)}M`}
          color="blue"
          icon={<RevenueIcon />}
          onClick={onMetricClick}
        />
        
        <MetricCard
          id="employees"
          label="Employees"
          value={metrics.employees}
          color="purple"
          icon={<EmployeesIcon />}
          onClick={onMetricClick}
        />
        
        <MetricCard
          id="facilities"
          label="Facilities"
          value={metrics.facilities}
          color="orange"
          icon={<FacilitiesIcon />}
          onClick={onMetricClick}
        />
      </div>

      {/* Reduction Targets */}
      {metrics.reductionTargets.length > 0 && (
        <div className="mt-6">
          <h4 className="text-sm font-semibold text-gray-400 uppercase tracking-wider mb-3">
            Reduction Targets
          </h4>
          <div className="space-y-3">
            {metrics.reductionTargets.map((target) => (
              <TargetCard key={target.id} target={target} />
            ))}
          </div>
        </div>
      )}
    </div>
  );
});

interface TargetCardProps {
  target: CarbonMetricsType['reductionTargets'][0];
}

function TargetCard({ target }: TargetCardProps) {
  const progress = target.targetValue === 0 
    ? 0 
    : ((target.currentValue - target.targetValue) / target.currentValue) * 100;
  
  const statusColors = {
    on_track: 'bg-green-500',
    at_risk: 'bg-yellow-500',
    behind: 'bg-red-500',
  };

  return (
    <div className="p-4 rounded-lg bg-gray-800/50 border border-gray-700/50">
      <div className="flex items-center justify-between mb-2">
        <span className="text-sm font-medium text-white">{target.name}</span>
        <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${
          target.status === 'on_track' ? 'bg-green-500/20 text-green-400' :
          target.status === 'at_risk' ? 'bg-yellow-500/20 text-yellow-400' :
          'bg-red-500/20 text-red-400'
        }`}>
          {target.status.replace('_', ' ')}
        </span>
      </div>
      
      <div className="mb-2">
        <div className="h-2 bg-gray-700 rounded-full overflow-hidden">
          <div
            className={`h-full rounded-full transition-all ${statusColors[target.status]}`}
            style={{ width: `${Math.min(100, Math.max(0, 100 - progress))}%` }}
          />
        </div>
      </div>
      
      <div className="flex justify-between text-xs text-gray-400">
        <span>Current: {target.currentValue.toLocaleString()} tCO₂e</span>
        <span>Target: {target.targetValue.toLocaleString()} tCO₂e</span>
      </div>
    </div>
  );
}

// Icons
function CarbonIcon() {
  return (
    <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
    </svg>
  );
}

function RevenueIcon() {
  return (
    <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
    </svg>
  );
}

function EmployeesIcon() {
  return (
    <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
    </svg>
  );
}

function FacilitiesIcon() {
  return (
    <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
    </svg>
  );
}

function TrendUpIcon() {
  return (
    <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
    </svg>
  );
}

function TrendDownIcon() {
  return (
    <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 17h8m0 0V9m0 8l-8-8-4 4-6-6" />
    </svg>
  );
}

function TrendStableIcon() {
  return (
    <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 12h14" />
    </svg>
  );
}

export default CarbonMetrics;
