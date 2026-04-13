'use client';

export function DashboardPreview() {
  return (
    <div className="relative mx-auto mt-16 max-w-5xl">
      {/* Browser chrome */}
      <div className="overflow-hidden rounded-xl border border-gray-700/50 bg-gray-900 shadow-2xl shadow-black/50">
        {/* Title bar */}
        <div className="flex items-center gap-2 border-b border-gray-800 bg-gray-900 px-4 py-2.5">
          <div className="flex gap-1.5">
            <div className="h-3 w-3 rounded-full bg-red-500/70" />
            <div className="h-3 w-3 rounded-full bg-yellow-500/70" />
            <div className="h-3 w-3 rounded-full bg-green-500/70" />
          </div>
          <div className="ml-4 flex-1 rounded-md bg-gray-800 px-3 py-1 text-center text-[11px] text-gray-500">
            off-grid-flow.com/dashboard/carbon
          </div>
        </div>

        {/* Dashboard content */}
        <div className="flex">
          {/* Sidebar */}
          <div className="hidden w-44 border-r border-gray-800 bg-gray-900/80 p-3 sm:block">
            <div className="mb-4 text-xs font-bold text-white">OffGridFlow</div>
            <div className="space-y-1">
              {['Dashboard', 'Emissions', 'Compliance', 'Settings'].map((item, i) => (
                <div
                  key={item}
                  className={`rounded-md px-2.5 py-1.5 text-[11px] ${
                    i === 0 ? 'bg-primary-600/10 text-primary-400' : 'text-gray-500'
                  }`}
                >
                  {item}
                </div>
              ))}
            </div>
          </div>

          {/* Main area */}
          <div className="flex-1 p-4">
            {/* Header */}
            <div className="mb-4 flex items-center justify-between">
              <div>
                <div className="text-sm font-semibold text-white">Carbon Dashboard</div>
                <div className="text-[10px] text-gray-600">Updated 2 minutes ago</div>
              </div>
              <div className="rounded-md bg-gray-800 px-2.5 py-1 text-[10px] text-gray-400">Export Report</div>
            </div>

            {/* KPI Row */}
            <div className="mb-4 grid grid-cols-4 gap-2">
              {[
                { label: 'Total Emissions', value: '21.36', unit: 'tCO2e', change: '-12.4%' },
                { label: 'Scope 1', value: '0', unit: 'tCO2e', change: '' },
                { label: 'Scope 2', value: '21.36', unit: 'tCO2e', change: '-8.2%' },
                { label: 'Scope 3', value: '—', unit: '', change: '' },
              ].map((kpi) => (
                <div key={kpi.label} className="rounded-lg border border-gray-800 bg-gray-800/40 p-2.5">
                  <div className="text-[9px] uppercase tracking-wider text-gray-500">{kpi.label}</div>
                  <div className="mt-1 flex items-baseline gap-1">
                    <span className="text-lg font-bold text-white">{kpi.value}</span>
                    <span className="text-[9px] text-gray-600">{kpi.unit}</span>
                  </div>
                  {kpi.change && (
                    <div className="mt-0.5 text-[9px] text-green-400">{kpi.change} vs last period</div>
                  )}
                </div>
              ))}
            </div>

            {/* Chart + Sidebar */}
            <div className="grid grid-cols-3 gap-3">
              {/* Chart area */}
              <div className="col-span-2 rounded-lg border border-gray-800 bg-gray-800/40 p-3">
                <div className="mb-2 text-[10px] uppercase tracking-wider text-gray-500">Emissions Trend</div>
                {/* Fake chart bars */}
                <div className="flex items-end gap-1.5" style={{ height: '80px' }}>
                  {[45, 52, 48, 65, 72, 68, 55, 60, 50, 42, 38, 44].map((h, i) => (
                    <div
                      key={i}
                      className="flex-1 rounded-t bg-gradient-to-t from-primary-600/60 to-primary-400/40"
                      style={{ height: `${h}%` }}
                    />
                  ))}
                </div>
                <div className="mt-1.5 flex justify-between text-[8px] text-gray-600">
                  <span>Jan</span><span>Mar</span><span>May</span><span>Jul</span><span>Sep</span><span>Dec</span>
                </div>
              </div>

              {/* Scope breakdown */}
              <div className="rounded-lg border border-gray-800 bg-gray-800/40 p-3">
                <div className="mb-2 text-[10px] uppercase tracking-wider text-gray-500">By Scope</div>
                {/* Stacked bar */}
                <div className="mb-3 flex h-2 overflow-hidden rounded-full bg-gray-700">
                  <div className="bg-amber-500" style={{ width: '100%' }} />
                </div>
                {[
                  { name: 'Scope 1', pct: '0%', color: 'bg-red-500' },
                  { name: 'Scope 2', pct: '100%', color: 'bg-amber-500' },
                  { name: 'Scope 3', pct: '—', color: 'bg-blue-500' },
                ].map((s) => (
                  <div key={s.name} className="flex items-center justify-between py-1">
                    <div className="flex items-center gap-1.5">
                      <div className={`h-1.5 w-1.5 rounded-full ${s.color}`} />
                      <span className="text-[10px] text-gray-400">{s.name}</span>
                    </div>
                    <span className="text-[10px] font-medium text-white">{s.pct}</span>
                  </div>
                ))}

                {/* Compliance status */}
                <div className="mt-3 border-t border-gray-800 pt-3">
                  <div className="mb-2 text-[10px] uppercase tracking-wider text-gray-500">Compliance</div>
                  {[
                    { name: 'CSRD', status: 'Ready', color: 'text-green-400' },
                    { name: 'SEC', status: 'Pending', color: 'text-gray-500' },
                    { name: 'SB 253', status: 'Ready', color: 'text-green-400' },
                  ].map((fw) => (
                    <div key={fw.name} className="flex items-center justify-between py-0.5">
                      <span className="text-[10px] text-gray-400">{fw.name}</span>
                      <span className={`text-[9px] ${fw.color}`}>{fw.status}</span>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Glow effect behind the mockup */}
      <div className="pointer-events-none absolute -inset-10 -z-10 bg-gradient-to-b from-primary-600/5 via-transparent to-transparent blur-3xl" />
    </div>
  );
}
