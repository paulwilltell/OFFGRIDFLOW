'use client';

import React, { useState, memo, useMemo } from 'react';
import { ComplianceDeadline } from '@/hooks/useCompliance';
import { ComplianceStatus } from '@/stores/carbonStore';

interface ComplianceCalendarProps {
  deadlines: ComplianceDeadline[];
  complianceStatus: ComplianceStatus | null;
}

const DAYS = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
const MONTHS = [
  'January', 'February', 'March', 'April', 'May', 'June',
  'July', 'August', 'September', 'October', 'November', 'December'
];

export const ComplianceCalendar = memo(function ComplianceCalendar({
  deadlines,
  complianceStatus,
}: ComplianceCalendarProps) {
  const [currentDate, setCurrentDate] = useState(new Date());
  const [selectedDate, setSelectedDate] = useState<Date | null>(null);

  const { year, month, days, startDay } = useMemo(() => {
    const y = currentDate.getFullYear();
    const m = currentDate.getMonth();
    const firstDay = new Date(y, m, 1);
    const lastDay = new Date(y, m + 1, 0);
    const daysInMonth = lastDay.getDate();
    const startDay = firstDay.getDay();

    return {
      year: y,
      month: m,
      days: daysInMonth,
      startDay,
    };
  }, [currentDate]);

  const deadlinesByDate = useMemo(() => {
    const map = new Map<string, ComplianceDeadline[]>();
    deadlines.forEach((deadline) => {
      const dateKey = deadline.dueDate.split('T')[0];
      if (!map.has(dateKey)) {
        map.set(dateKey, []);
      }
      map.get(dateKey)!.push(deadline);
    });
    return map;
  }, [deadlines]);

  const overallScore = useMemo(() => {
    if (!complianceStatus) return null;
    const scoreMap: Record<ComplianceStatus[keyof ComplianceStatus], number> = {
      complete: 100,
      in_progress: 60,
      pending: 30,
      at_risk: 40,
      overdue: 10,
    };
    const statuses = Object.values(complianceStatus) as Array<ComplianceStatus[keyof ComplianceStatus]>;
    if (statuses.length === 0) return null;
    const total = statuses.reduce((sum, status) => sum + scoreMap[status], 0);
    return Math.round(total / statuses.length);
  }, [complianceStatus]);

  const navigateMonth = (direction: 'prev' | 'next') => {
    setCurrentDate((prev) => {
      const newDate = new Date(prev);
      if (direction === 'prev') {
        newDate.setMonth(newDate.getMonth() - 1);
      } else {
        newDate.setMonth(newDate.getMonth() + 1);
      }
      return newDate;
    });
  };

  const getDateKey = (day: number) => {
    return `${year}-${String(month + 1).padStart(2, '0')}-${String(day).padStart(2, '0')}`;
  };

  const isToday = (day: number) => {
    const today = new Date();
    return (
      today.getFullYear() === year &&
      today.getMonth() === month &&
      today.getDate() === day
    );
  };

  const selectedDateDeadlines = useMemo(() => {
    if (!selectedDate) return [];
    const key = selectedDate.toISOString().split('T')[0];
    return deadlinesByDate.get(key) || [];
  }, [selectedDate, deadlinesByDate]);

  return (
    <div className="bg-gray-800/50 rounded-xl border border-gray-700/50 overflow-hidden">
      {/* Header */}
      <div className="p-4 border-b border-gray-700/50">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-sm font-semibold text-gray-400 uppercase tracking-wider">
            Compliance Calendar
          </h3>
          {overallScore !== null && (
            <div className="flex items-center gap-2">
              <span className="text-xs text-gray-500">Overall Score</span>
              <span className={`text-sm font-bold ${
                overallScore >= 80 ? 'text-green-400' :
                overallScore >= 60 ? 'text-yellow-400' :
                'text-red-400'
              }`}>
                {overallScore}%
              </span>
            </div>
          )}
        </div>

        {/* Month navigation */}
        <div className="flex items-center justify-between">
          <button
            onClick={() => navigateMonth('prev')}
            className="p-2 rounded-lg hover:bg-gray-700/50 transition-colors"
          >
            <ChevronLeftIcon />
          </button>
          <span className="text-white font-semibold">
            {MONTHS[month]} {year}
          </span>
          <button
            onClick={() => navigateMonth('next')}
            className="p-2 rounded-lg hover:bg-gray-700/50 transition-colors"
          >
            <ChevronRightIcon />
          </button>
        </div>
      </div>

      {/* Calendar Grid */}
      <div className="p-4">
        {/* Day headers */}
        <div className="grid grid-cols-7 gap-1 mb-2">
          {DAYS.map((day) => (
            <div
              key={day}
              className="text-center text-xs font-medium text-gray-500 py-2"
            >
              {day}
            </div>
          ))}
        </div>

        {/* Days grid */}
        <div className="grid grid-cols-7 gap-1">
          {/* Empty cells for days before month start */}
          {Array.from({ length: startDay }, (_, i) => (
            <div key={`empty-${i}`} className="aspect-square" />
          ))}

          {/* Day cells */}
          {Array.from({ length: days }, (_, i) => {
            const day = i + 1;
            const dateKey = getDateKey(day);
            const dayDeadlines = deadlinesByDate.get(dateKey) || [];
            const hasDeadlines = dayDeadlines.length > 0;
            const today = isToday(day);

            return (
              <button
                key={day}
                onClick={() => {
                  const date = new Date(year, month, day);
                  setSelectedDate(date);
                }}
                className={`
                  aspect-square rounded-lg flex flex-col items-center justify-center
                  text-sm transition-all relative
                  ${today ? 'bg-green-500/20 text-green-400 font-bold' : 'text-gray-300 hover:bg-gray-700/50'}
                  ${hasDeadlines ? 'ring-1 ring-inset ring-opacity-50' : ''}
                  ${hasDeadlines && dayDeadlines.some(d => d.status === 'overdue') ? 'ring-red-500' : ''}
                  ${hasDeadlines && dayDeadlines.some(d => d.status === 'due_soon') ? 'ring-yellow-500' : ''}
                  ${hasDeadlines && !dayDeadlines.some(d => d.status === 'overdue' || d.status === 'due_soon') ? 'ring-blue-500' : ''}
                `}
              >
                <span>{day}</span>
                {hasDeadlines && (
                  <div className="flex gap-0.5 mt-1">
                    {dayDeadlines.slice(0, 3).map((d, idx) => (
                      <span
                        key={idx}
                        className={`w-1 h-1 rounded-full ${
                          d.status === 'overdue' ? 'bg-red-500' :
                          d.status === 'due_soon' ? 'bg-yellow-500' :
                          'bg-blue-500'
                        }`}
                      />
                    ))}
                  </div>
                )}
              </button>
            );
          })}
        </div>
      </div>

      {/* Selected date details */}
      {selectedDate && selectedDateDeadlines.length > 0 && (
        <div className="p-4 border-t border-gray-700/50 space-y-3">
          <div className="flex items-center justify-between">
            <h4 className="text-sm font-semibold text-white">
              {selectedDate.toLocaleDateString('en-US', { month: 'long', day: 'numeric', year: 'numeric' })}
            </h4>
            <button
              onClick={() => setSelectedDate(null)}
              className="text-gray-400 hover:text-white"
            >
              <CloseIcon />
            </button>
          </div>
          {selectedDateDeadlines.map((deadline) => (
            <DeadlineCard key={deadline.id} deadline={deadline} />
          ))}
        </div>
      )}

      {/* Upcoming deadlines summary */}
      <div className="p-4 border-t border-gray-700/50">
        <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-3">
          Upcoming Deadlines
        </h4>
        <div className="space-y-2">
          {deadlines
            .filter((d) => d.status !== 'completed')
            .sort((a, b) => new Date(a.dueDate).getTime() - new Date(b.dueDate).getTime())
            .slice(0, 3)
            .map((deadline) => (
              <div
                key={deadline.id}
                className="flex items-center justify-between p-2 rounded-lg bg-gray-700/30"
              >
                <div className="flex items-center gap-2">
                  <FrameworkBadge framework={deadline.framework} />
                  <span className="text-sm text-gray-300 truncate max-w-[120px]">
                    {deadline.title}
                  </span>
                </div>
                <span className={`text-xs font-medium ${
                  deadline.status === 'overdue' ? 'text-red-400' :
                  deadline.status === 'due_soon' ? 'text-yellow-400' :
                  'text-gray-400'
                }`}>
                  {formatDateShort(deadline.dueDate)}
                </span>
              </div>
            ))}
        </div>
      </div>
    </div>
  );
});

interface DeadlineCardProps {
  deadline: ComplianceDeadline;
}

function DeadlineCard({ deadline }: DeadlineCardProps) {
  const [expanded, setExpanded] = useState(false);

  return (
    <div className="p-3 rounded-lg bg-gray-700/30 border border-gray-700/50">
      <div className="flex items-start justify-between mb-2">
        <div className="flex items-center gap-2">
          <FrameworkBadge framework={deadline.framework} />
          <span className="text-sm font-medium text-white">{deadline.title}</span>
        </div>
        <StatusBadge status={deadline.status} />
      </div>
      
      <p className="text-xs text-gray-400 mb-2">{deadline.description}</p>
      
      {deadline.requirements.length > 0 && (
        <>
          <button
            onClick={() => setExpanded(!expanded)}
            className="text-xs text-blue-400 hover:text-blue-300 flex items-center gap-1"
          >
            {expanded ? 'Hide' : 'Show'} requirements ({deadline.requirements.length})
            <ChevronDownIcon className={`w-3 h-3 transition-transform ${expanded ? 'rotate-180' : ''}`} />
          </button>
          
          {expanded && (
            <ul className="mt-2 space-y-1">
              {deadline.requirements.map((req, idx) => (
                <li key={idx} className="text-xs text-gray-400 flex items-center gap-2">
                  <span className="w-1 h-1 rounded-full bg-gray-500" />
                  {req}
                </li>
              ))}
            </ul>
          )}
        </>
      )}
    </div>
  );
}

function FrameworkBadge({ framework }: { framework: string }) {
  const colors = {
    csrd: 'bg-blue-500/20 text-blue-400',
    sec: 'bg-purple-500/20 text-purple-400',
    cbam: 'bg-green-500/20 text-green-400',
    california: 'bg-orange-500/20 text-orange-400',
  };

  return (
    <span className={`px-2 py-0.5 rounded text-xs font-medium uppercase ${colors[framework as keyof typeof colors] || 'bg-gray-500/20 text-gray-400'}`}>
      {framework}
    </span>
  );
}

function StatusBadge({ status }: { status: ComplianceDeadline['status'] }) {
  const colors = {
    upcoming: 'bg-blue-500/20 text-blue-400',
    due_soon: 'bg-yellow-500/20 text-yellow-400',
    overdue: 'bg-red-500/20 text-red-400',
    completed: 'bg-green-500/20 text-green-400',
  };

  return (
    <span className={`px-2 py-0.5 rounded text-xs font-medium ${colors[status]}`}>
      {status.replace('_', ' ')}
    </span>
  );
}

function formatDateShort(dateStr: string): string {
  const date = new Date(dateStr);
  const now = new Date();
  const diff = Math.ceil((date.getTime() - now.getTime()) / (1000 * 60 * 60 * 24));
  
  if (diff < 0) return `${Math.abs(diff)}d overdue`;
  if (diff === 0) return 'Today';
  if (diff === 1) return 'Tomorrow';
  if (diff <= 7) return `${diff}d`;
  return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
}

// Icons
function ChevronLeftIcon() {
  return (
    <svg className="w-5 h-5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
    </svg>
  );
}

function ChevronRightIcon() {
  return (
    <svg className="w-5 h-5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
    </svg>
  );
}

function ChevronDownIcon({ className = 'w-4 h-4' }: { className?: string }) {
  return (
    <svg className={className} fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
    </svg>
  );
}

function CloseIcon() {
  return (
    <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
    </svg>
  );
}

export default ComplianceCalendar;
