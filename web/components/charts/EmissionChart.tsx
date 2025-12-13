'use client';

import React, { memo, useMemo, useState } from 'react';
import { EmissionData, EmissionScope } from '@/stores/carbonStore';

interface EmissionChartProps {
  data: EmissionData | null;
  timeframe: 'monthly' | 'quarterly' | 'yearly';
  height?: number;
}

type ChartView = 'trend' | 'breakdown' | 'comparison';

export const EmissionChart = memo(function EmissionChart({
  data,
  timeframe,
  height = 400,
}: EmissionChartProps) {
  const [view, setView] = useState<ChartView>('trend');
  const [hoveredBar, setHoveredBar] = useState<number | null>(null);

  // Generate mock historical data for visualization
  const chartData = useMemo(() => {
    if (!data) return [];
    
    const periods = timeframe === 'monthly' ? 12 : timeframe === 'quarterly' ? 4 : 5;
    const labels = timeframe === 'monthly' 
      ? ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']
      : timeframe === 'quarterly'
      ? ['Q1', 'Q2', 'Q3', 'Q4']
      : ['2021', '2022', '2023', '2024', '2025'];

    return labels.slice(0, periods).map((label, index) => {
      // Simulate decreasing trend with some variation
      const baseValue = data.total * (1 + (periods - index) * 0.05);
      const variation = (Math.random() - 0.5) * 0.1 * baseValue;
      const total = Math.max(0, baseValue + variation);
      
      // Distribute across scopes (approximate ratios)
      const scope1Ratio = data.scopes.scope1 / data.total;
      const scope2Ratio = data.scopes.scope2 / data.total;
      const scope3Ratio = data.scopes.scope3 / data.total;

      return {
        label,
        total: Math.round(total),
        scope1: Math.round(total * scope1Ratio),
        scope2: Math.round(total * scope2Ratio),
        scope3: Math.round(total * scope3Ratio),
      };
    });
  }, [data, timeframe]);

  const maxValue = useMemo(() => {
    return Math.max(...chartData.map(d => d.total)) * 1.1;
  }, [chartData]);

  if (!data) {
    return (
      <div className="bg-gray-800/50 rounded-xl border border-gray-700/50 p-6" style={{ height }}>
        <div className="h-full flex items-center justify-center text-gray-500">
          No emission data available
        </div>
      </div>
    );
  }

  return (
    <div className="bg-gray-800/50 rounded-xl border border-gray-700/50 overflow-hidden">
      {/* Header */}
      <div className="p-4 border-b border-gray-700/50">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="text-lg font-semibold text-white">Emissions Overview</h3>
            <p className="text-sm text-gray-400">
              Total: {data.total.toLocaleString()} tCO₂e
              <span className={`ml-2 ${data.trend === 'down' ? 'text-green-400' : 'text-red-400'}`}>
                {data.trend === 'down' ? '↓' : '↑'} {Math.abs(data.percentageChange).toFixed(1)}%
              </span>
            </p>
          </div>
          
          {/* View toggles */}
          <div className="flex gap-1 bg-gray-700/50 rounded-lg p-1">
            {(['trend', 'breakdown', 'comparison'] as ChartView[]).map((v) => (
              <button
                key={v}
                onClick={() => setView(v)}
                className={`px-3 py-1.5 text-xs font-medium rounded-md transition-colors ${
                  view === v
                    ? 'bg-green-500/20 text-green-400'
                    : 'text-gray-400 hover:text-white hover:bg-gray-600/50'
                }`}
              >
                {v.charAt(0).toUpperCase() + v.slice(1)}
              </button>
            ))}
          </div>
        </div>
      </div>

      {/* Chart Area */}
      <div className="p-6" style={{ height: height - 80 }}>
        {view === 'trend' && (
          <TrendChart
            data={chartData}
            maxValue={maxValue}
            hoveredBar={hoveredBar}
            setHoveredBar={setHoveredBar}
          />
        )}
        {view === 'breakdown' && (
          <BreakdownChart scopes={data.scopes} total={data.total} />
        )}
        {view === 'comparison' && (
          <ComparisonChart data={chartData} maxValue={maxValue} />
        )}
      </div>

      {/* Legend */}
      <div className="px-6 pb-4">
        <div className="flex items-center justify-center gap-6">
          <LegendItem color="bg-green-500" label="Scope 1" />
          <LegendItem color="bg-blue-500" label="Scope 2" />
          <LegendItem color="bg-purple-500" label="Scope 3" />
        </div>
      </div>
    </div>
  );
});

interface TrendChartProps {
  data: Array<{ label: string; total: number; scope1: number; scope2: number; scope3: number }>;
  maxValue: number;
  hoveredBar: number | null;
  setHoveredBar: (index: number | null) => void;
}

function TrendChart({ data, maxValue, hoveredBar, setHoveredBar }: TrendChartProps) {
  return (
    <div className="h-full flex flex-col">
      {/* Y-axis and bars */}
      <div className="flex-1 flex">
        {/* Y-axis labels */}
        <div className="w-16 flex flex-col justify-between text-right pr-2 py-2">
          {[100, 75, 50, 25, 0].map((pct) => (
            <span key={pct} className="text-xs text-gray-500">
              {Math.round((maxValue * pct) / 100).toLocaleString()}
            </span>
          ))}
        </div>

        {/* Chart area */}
        <div className="flex-1 relative">
          {/* Grid lines */}
          <div className="absolute inset-0 flex flex-col justify-between pointer-events-none">
            {[0, 1, 2, 3, 4].map((i) => (
              <div key={i} className="border-t border-gray-700/30" />
            ))}
          </div>

          {/* Bars */}
          <div className="absolute inset-0 flex items-end justify-between px-2 pb-6">
            {data.map((d, i) => {
              const totalHeight = (d.total / maxValue) * 100;
              const scope1Height = (d.scope1 / d.total) * totalHeight;
              const scope2Height = (d.scope2 / d.total) * totalHeight;
              const scope3Height = (d.scope3 / d.total) * totalHeight;
              const isHovered = hoveredBar === i;

              return (
                <div
                  key={d.label}
                  className="flex-1 flex flex-col items-center relative group"
                  onMouseEnter={() => setHoveredBar(i)}
                  onMouseLeave={() => setHoveredBar(null)}
                >
                  {/* Tooltip */}
                  {isHovered && (
                    <div className="absolute bottom-full mb-2 left-1/2 -translate-x-1/2 z-10">
                      <div className="bg-gray-900 border border-gray-700 rounded-lg p-3 shadow-xl min-w-[140px]">
                        <div className="text-xs font-medium text-white mb-2">{d.label}</div>
                        <div className="space-y-1 text-xs">
                          <div className="flex justify-between">
                            <span className="text-gray-400">Total:</span>
                            <span className="text-white font-medium">{d.total.toLocaleString()}</span>
                          </div>
                          <div className="flex justify-between">
                            <span className="text-green-400">Scope 1:</span>
                            <span className="text-white">{d.scope1.toLocaleString()}</span>
                          </div>
                          <div className="flex justify-between">
                            <span className="text-blue-400">Scope 2:</span>
                            <span className="text-white">{d.scope2.toLocaleString()}</span>
                          </div>
                          <div className="flex justify-between">
                            <span className="text-purple-400">Scope 3:</span>
                            <span className="text-white">{d.scope3.toLocaleString()}</span>
                          </div>
                        </div>
                      </div>
                    </div>
                  )}

                  {/* Stacked bar */}
                  <div
                    className={`w-8 rounded-t transition-all cursor-pointer ${
                      isHovered ? 'opacity-100 scale-105' : 'opacity-80'
                    }`}
                    style={{ height: `${totalHeight}%` }}
                  >
                    <div className="h-full flex flex-col-reverse rounded-t overflow-hidden">
                      <div className="bg-green-500" style={{ height: `${scope1Height}%` }} />
                      <div className="bg-blue-500" style={{ height: `${scope2Height}%` }} />
                      <div className="bg-purple-500" style={{ height: `${scope3Height}%` }} />
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      </div>

      {/* X-axis labels */}
      <div className="flex pl-16">
        {data.map((d) => (
          <div key={d.label} className="flex-1 text-center text-xs text-gray-500">
            {d.label}
          </div>
        ))}
      </div>
    </div>
  );
}

interface BreakdownChartProps {
  scopes: EmissionScope;
  total: number;
}

function BreakdownChart({ scopes, total }: BreakdownChartProps) {
  const data = [
    { name: 'Scope 1', value: scopes.scope1, color: '#22c55e', description: 'Direct emissions' },
    { name: 'Scope 2', value: scopes.scope2, color: '#3b82f6', description: 'Indirect (energy)' },
    { name: 'Scope 3', value: scopes.scope3, color: '#a855f7', description: 'Value chain' },
  ];

  // Calculate percentages and angles for donut chart
  let currentAngle = 0;
  const segments = data.map((d) => {
    const percentage = (d.value / total) * 100;
    const angle = (percentage / 100) * 360;
    const segment = {
      ...d,
      percentage,
      startAngle: currentAngle,
      endAngle: currentAngle + angle,
    };
    currentAngle += angle;
    return segment;
  });

  return (
    <div className="h-full flex items-center justify-center gap-12">
      {/* Donut chart */}
      <div className="relative">
        <svg width="200" height="200" viewBox="0 0 200 200">
          {segments.map((segment, i) => (
            <DonutSegment
              key={segment.name}
              cx={100}
              cy={100}
              radius={80}
              startAngle={segment.startAngle}
              endAngle={segment.endAngle}
              color={segment.color}
            />
          ))}
          {/* Center text */}
          <text x="100" y="95" textAnchor="middle" className="fill-white text-2xl font-bold">
            {total.toLocaleString()}
          </text>
          <text x="100" y="115" textAnchor="middle" className="fill-gray-400 text-xs">
            tCO₂e
          </text>
        </svg>
      </div>

      {/* Legend with details */}
      <div className="space-y-4">
        {data.map((d) => (
          <div key={d.name} className="flex items-center gap-4">
            <div
              className="w-4 h-4 rounded"
              style={{ backgroundColor: d.color }}
            />
            <div>
              <div className="text-sm font-medium text-white">{d.name}</div>
              <div className="text-xs text-gray-400">{d.description}</div>
              <div className="text-lg font-bold text-white">
                {d.value.toLocaleString()} <span className="text-xs text-gray-400">({((d.value / total) * 100).toFixed(1)}%)</span>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

interface DonutSegmentProps {
  cx: number;
  cy: number;
  radius: number;
  startAngle: number;
  endAngle: number;
  color: string;
}

function DonutSegment({ cx, cy, radius, startAngle, endAngle, color }: DonutSegmentProps) {
  const innerRadius = radius * 0.6;
  
  const startRad = (startAngle - 90) * (Math.PI / 180);
  const endRad = (endAngle - 90) * (Math.PI / 180);
  
  const x1 = cx + radius * Math.cos(startRad);
  const y1 = cy + radius * Math.sin(startRad);
  const x2 = cx + radius * Math.cos(endRad);
  const y2 = cy + radius * Math.sin(endRad);
  const x3 = cx + innerRadius * Math.cos(endRad);
  const y3 = cy + innerRadius * Math.sin(endRad);
  const x4 = cx + innerRadius * Math.cos(startRad);
  const y4 = cy + innerRadius * Math.sin(startRad);
  
  const largeArc = endAngle - startAngle > 180 ? 1 : 0;
  
  const d = `
    M ${x1} ${y1}
    A ${radius} ${radius} 0 ${largeArc} 1 ${x2} ${y2}
    L ${x3} ${y3}
    A ${innerRadius} ${innerRadius} 0 ${largeArc} 0 ${x4} ${y4}
    Z
  `;
  
  return <path d={d} fill={color} className="transition-opacity hover:opacity-80" />;
}

interface ComparisonChartProps {
  data: Array<{ label: string; total: number; scope1: number; scope2: number; scope3: number }>;
  maxValue: number;
}

function ComparisonChart({ data, maxValue }: ComparisonChartProps) {
  if (data.length < 2) {
    return (
      <div className="h-full flex items-center justify-center text-gray-500">
        Need more data for comparison
      </div>
    );
  }

  const current = data[data.length - 1];
  const previous = data[data.length - 2];
  const change = ((current.total - previous.total) / previous.total) * 100;

  const comparisons = [
    { label: 'Total', current: current.total, previous: previous.total },
    { label: 'Scope 1', current: current.scope1, previous: previous.scope1 },
    { label: 'Scope 2', current: current.scope2, previous: previous.scope2 },
    { label: 'Scope 3', current: current.scope3, previous: previous.scope3 },
  ];

  return (
    <div className="h-full flex flex-col justify-center space-y-6">
      <div className="text-center">
        <div className="text-3xl font-bold text-white">
          {change > 0 ? '+' : ''}{change.toFixed(1)}%
        </div>
        <div className={`text-sm ${change < 0 ? 'text-green-400' : 'text-red-400'}`}>
          {change < 0 ? 'Decrease' : 'Increase'} vs previous period
        </div>
      </div>

      <div className="space-y-4">
        {comparisons.map((c) => {
          const pctChange = ((c.current - c.previous) / c.previous) * 100;
          return (
            <div key={c.label} className="space-y-2">
              <div className="flex justify-between text-sm">
                <span className="text-gray-400">{c.label}</span>
                <span className={pctChange < 0 ? 'text-green-400' : 'text-red-400'}>
                  {pctChange > 0 ? '+' : ''}{pctChange.toFixed(1)}%
                </span>
              </div>
              <div className="flex gap-2 h-6">
                <div className="flex-1 bg-gray-700/50 rounded overflow-hidden">
                  <div
                    className="h-full bg-gray-500 rounded"
                    style={{ width: `${(c.previous / maxValue) * 100}%` }}
                  />
                </div>
                <div className="flex-1 bg-gray-700/50 rounded overflow-hidden">
                  <div
                    className={`h-full rounded ${pctChange < 0 ? 'bg-green-500' : 'bg-red-500'}`}
                    style={{ width: `${(c.current / maxValue) * 100}%` }}
                  />
                </div>
              </div>
              <div className="flex justify-between text-xs text-gray-500">
                <span>{c.previous.toLocaleString()} (prev)</span>
                <span>{c.current.toLocaleString()} (curr)</span>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}

function LegendItem({ color, label }: { color: string; label: string }) {
  return (
    <div className="flex items-center gap-2">
      <div className={`w-3 h-3 rounded ${color}`} />
      <span className="text-xs text-gray-400">{label}</span>
    </div>
  );
}

export default EmissionChart;
