'use client';

import React from 'react';

type SkeletonType = 'text' | 'card' | 'chart' | 'metric' | 'table' | 'calendar' | 'globe';

interface LoadingSkeletonProps {
  type: SkeletonType;
  count?: number;
  className?: string;
}

export function LoadingSkeleton({ type, count = 1, className = '' }: LoadingSkeletonProps) {
  const items = Array.from({ length: count }, (_, i) => i);

  return (
    <div className={`animate-pulse ${className}`}>
      {items.map((index) => (
        <SkeletonItem key={index} type={type} />
      ))}
    </div>
  );
}

function SkeletonItem({ type }: { type: SkeletonType }) {
  switch (type) {
    case 'text':
      return <TextSkeleton />;
    case 'card':
      return <CardSkeleton />;
    case 'chart':
      return <ChartSkeleton />;
    case 'metric':
      return <MetricSkeleton />;
    case 'table':
      return <TableSkeleton />;
    case 'calendar':
      return <CalendarSkeleton />;
    case 'globe':
      return <GlobeSkeleton />;
    default:
      return <CardSkeleton />;
  }
}

function TextSkeleton() {
  return (
    <div className="space-y-2">
      <div className="h-4 bg-gray-700/50 rounded w-3/4" />
      <div className="h-4 bg-gray-700/50 rounded w-1/2" />
    </div>
  );
}

function CardSkeleton() {
  return (
    <div className="bg-gray-800/50 rounded-xl p-6 border border-gray-700/50">
      <div className="h-4 bg-gray-700/50 rounded w-1/3 mb-4" />
      <div className="space-y-3">
        <div className="h-8 bg-gray-700/50 rounded w-2/3" />
        <div className="h-4 bg-gray-700/50 rounded w-1/2" />
      </div>
    </div>
  );
}

function ChartSkeleton() {
  return (
    <div className="bg-gray-800/50 rounded-xl p-6 border border-gray-700/50">
      <div className="h-4 bg-gray-700/50 rounded w-1/4 mb-6" />
      <div className="flex items-end justify-between h-64 gap-2">
        {[40, 65, 45, 80, 55, 70, 60, 75, 50, 85, 65, 70].map((height, i) => (
          <div
            key={i}
            className="flex-1 bg-gray-700/50 rounded-t"
            style={{ height: `${height}%` }}
          />
        ))}
      </div>
      <div className="flex justify-between mt-4">
        {['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun'].map((_, i) => (
          <div key={i} className="h-3 bg-gray-700/50 rounded w-8" />
        ))}
      </div>
    </div>
  );
}

function MetricSkeleton() {
  return (
    <div className="bg-gray-800/50 rounded-xl p-5 border border-gray-700/50">
      <div className="flex items-center justify-between mb-3">
        <div className="w-10 h-10 bg-gray-700/50 rounded-lg" />
        <div className="w-16 h-5 bg-gray-700/50 rounded" />
      </div>
      <div className="h-8 bg-gray-700/50 rounded w-2/3 mb-2" />
      <div className="h-4 bg-gray-700/50 rounded w-1/2" />
    </div>
  );
}

function TableSkeleton() {
  return (
    <div className="bg-gray-800/50 rounded-xl border border-gray-700/50 overflow-hidden">
      {/* Header */}
      <div className="flex gap-4 p-4 border-b border-gray-700/50">
        {[1, 2, 3, 4].map((i) => (
          <div key={i} className="h-4 bg-gray-700/50 rounded flex-1" />
        ))}
      </div>
      {/* Rows */}
      {[1, 2, 3, 4, 5].map((row) => (
        <div key={row} className="flex gap-4 p-4 border-b border-gray-700/30 last:border-0">
          {[1, 2, 3, 4].map((col) => (
            <div
              key={col}
              className="h-4 bg-gray-700/50 rounded flex-1"
              style={{ width: `${Math.random() * 30 + 50}%` }}
            />
          ))}
        </div>
      ))}
    </div>
  );
}

function CalendarSkeleton() {
  return (
    <div className="bg-gray-800/50 rounded-xl p-6 border border-gray-700/50">
      <div className="flex justify-between items-center mb-6">
        <div className="h-5 bg-gray-700/50 rounded w-32" />
        <div className="flex gap-2">
          <div className="w-8 h-8 bg-gray-700/50 rounded" />
          <div className="w-8 h-8 bg-gray-700/50 rounded" />
        </div>
      </div>
      {/* Calendar grid */}
      <div className="grid grid-cols-7 gap-2">
        {/* Day headers */}
        {['S', 'M', 'T', 'W', 'T', 'F', 'S'].map((_, i) => (
          <div key={i} className="h-6 bg-gray-700/50 rounded" />
        ))}
        {/* Days */}
        {Array.from({ length: 35 }, (_, i) => (
          <div key={i} className="h-10 bg-gray-700/50 rounded" />
        ))}
      </div>
    </div>
  );
}

function GlobeSkeleton() {
  return (
    <div className="bg-gray-800/50 rounded-xl p-6 border border-gray-700/50 flex items-center justify-center">
      <div className="relative">
        {/* Globe circle */}
        <div className="w-64 h-64 rounded-full bg-gray-700/30 border-2 border-gray-700/50" />
        {/* Latitude lines */}
        <div className="absolute inset-0 flex items-center justify-center">
          <div className="w-full h-px bg-gray-700/50" />
        </div>
        <div className="absolute inset-0 flex items-center justify-center rotate-45">
          <div className="w-full h-px bg-gray-700/50" />
        </div>
        <div className="absolute inset-0 flex items-center justify-center -rotate-45">
          <div className="w-full h-px bg-gray-700/50" />
        </div>
        {/* Center dot */}
        <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2">
          <div className="w-4 h-4 bg-gray-700/50 rounded-full animate-ping" />
        </div>
      </div>
    </div>
  );
}

// Dashboard-specific skeleton
export function DashboardSkeleton() {
  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-900 via-gray-900 to-gray-800 p-6">
      {/* Header */}
      <div className="flex justify-between items-center mb-8">
        <div className="space-y-2">
          <div className="h-8 bg-gray-700/50 rounded w-64 animate-pulse" />
          <div className="h-4 bg-gray-700/50 rounded w-48 animate-pulse" />
        </div>
        <div className="flex gap-3">
          <div className="h-10 bg-gray-700/50 rounded w-32 animate-pulse" />
          <div className="h-10 bg-gray-700/50 rounded w-24 animate-pulse" />
        </div>
      </div>

      {/* Grid */}
      <div className="grid grid-cols-12 gap-6">
        {/* Left column - Metrics */}
        <div className="col-span-3 space-y-4">
          <LoadingSkeleton type="metric" count={4} className="space-y-4" />
        </div>

        {/* Center column - Charts */}
        <div className="col-span-6 space-y-6">
          <LoadingSkeleton type="chart" />
          <LoadingSkeleton type="globe" />
        </div>

        {/* Right column - Calendar & Insights */}
        <div className="col-span-3 space-y-4">
          <LoadingSkeleton type="calendar" />
          <LoadingSkeleton type="card" count={2} className="space-y-4" />
        </div>
      </div>
    </div>
  );
}

export default LoadingSkeleton;
