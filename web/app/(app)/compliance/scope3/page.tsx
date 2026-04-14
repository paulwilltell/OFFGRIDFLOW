'use client';

import { useState } from 'react';
import Link from 'next/link';
import { useRequireAuth } from '@/lib/session';

const scope3Categories = [
  { id: 1, name: 'Purchased Goods & Services', upstream: true, description: 'Emissions from production of purchased goods and services', method: 'spend-based', typical: '40-60% of total Scope 3 for most companies' },
  { id: 2, name: 'Capital Goods', upstream: true, description: 'Emissions from production of capital goods purchased', method: 'spend-based', typical: '5-15%' },
  { id: 3, name: 'Fuel & Energy-Related Activities', upstream: true, description: 'Emissions not included in Scope 1 or 2 (T&D losses, upstream fuel)', method: 'average-data', typical: '3-8%' },
  { id: 4, name: 'Upstream Transportation & Distribution', upstream: true, description: 'Transportation and distribution of purchased goods', method: 'distance-based', typical: '2-10%' },
  { id: 5, name: 'Waste Generated in Operations', upstream: true, description: 'Disposal and treatment of waste', method: 'waste-type', typical: '1-3%' },
  { id: 6, name: 'Business Travel', upstream: true, description: 'Employee travel for business purposes', method: 'distance-based', typical: '1-5%' },
  { id: 7, name: 'Employee Commuting', upstream: true, description: 'Transportation of employees between home and work', method: 'average-data', typical: '2-5%' },
  { id: 8, name: 'Upstream Leased Assets', upstream: true, description: 'Operation of assets leased by the reporting company', method: 'asset-specific', typical: '1-10%' },
  { id: 9, name: 'Downstream Transportation & Distribution', upstream: false, description: 'Transportation and distribution of sold products', method: 'distance-based', typical: '2-8%' },
  { id: 10, name: 'Processing of Sold Products', upstream: false, description: 'Processing of intermediate products by third parties', method: 'average-data', typical: 'Varies by sector' },
  { id: 11, name: 'Use of Sold Products', upstream: false, description: 'End use of goods and services sold', method: 'product-specific', typical: '10-50% for energy products' },
  { id: 12, name: 'End-of-Life Treatment of Sold Products', upstream: false, description: 'Waste disposal and treatment of products sold', method: 'waste-type', typical: '1-3%' },
  { id: 13, name: 'Downstream Leased Assets', upstream: false, description: 'Operation of assets owned and leased to others', method: 'asset-specific', typical: 'Varies' },
  { id: 14, name: 'Franchises', upstream: false, description: 'Operation of franchises', method: 'franchise-specific', typical: 'Franchise companies only' },
  { id: 15, name: 'Investments', upstream: false, description: 'Investments including equity and debt', method: 'investment-specific', typical: 'Financial institutions only' },
];

export default function Scope3Page() {
  useRequireAuth();
  const [selectedCategory, setSelectedCategory] = useState<number | null>(null);

  const upstream = scope3Categories.filter(c => c.upstream);
  const downstream = scope3Categories.filter(c => !c.upstream);

  return (
    <div>
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-white">Scope 3 — Value Chain Emissions</h1>
          <p className="mt-1 text-xs text-gray-500">GHG Protocol Corporate Value Chain Standard — 15 categories</p>
        </div>
        <Link
          href="/emissions"
          className="rounded-lg border border-gray-700 px-4 py-2 text-sm text-gray-300 hover:border-gray-600 hover:text-white"
        >
          Upload Data
        </Link>
      </div>

      {/* Summary cards */}
      <div className="mb-6 grid grid-cols-3 gap-4">
        <div className="rounded-xl border border-gray-800 bg-gray-800/40 p-4">
          <div className="text-xs uppercase tracking-wider text-gray-500">Total Scope 3</div>
          <div className="mt-1 text-2xl font-bold text-white">—</div>
          <div className="mt-1 text-xs text-gray-600">Upload data to calculate</div>
        </div>
        <div className="rounded-xl border border-gray-800 bg-gray-800/40 p-4">
          <div className="text-xs uppercase tracking-wider text-gray-500">Categories Assessed</div>
          <div className="mt-1 text-2xl font-bold text-white">0 / 15</div>
          <div className="mt-1 text-xs text-gray-600">Complete materiality assessment</div>
        </div>
        <div className="rounded-xl border border-gray-800 bg-gray-800/40 p-4">
          <div className="text-xs uppercase tracking-wider text-gray-500">Data Quality</div>
          <div className="mt-1 text-2xl font-bold text-gray-500">—</div>
          <div className="mt-1 text-xs text-gray-600">No data yet</div>
        </div>
      </div>

      {/* Upstream categories */}
      <div className="mb-6">
        <h2 className="mb-3 text-sm font-semibold text-gray-400">Upstream Activities (Categories 1–8)</h2>
        <div className="space-y-2">
          {upstream.map(cat => (
            <button
              key={cat.id}
              onClick={() => setSelectedCategory(selectedCategory === cat.id ? null : cat.id)}
              className="w-full rounded-lg border border-gray-800 bg-gray-800/30 p-4 text-left transition hover:border-gray-700"
            >
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <div className="flex h-7 w-7 items-center justify-center rounded bg-blue-500/10 text-xs font-bold text-blue-400">
                    {cat.id}
                  </div>
                  <div>
                    <div className="text-sm font-medium text-white">{cat.name}</div>
                    <div className="mt-0.5 text-xs text-gray-500">{cat.description}</div>
                  </div>
                </div>
                <div className="flex items-center gap-4">
                  <span className="rounded bg-gray-700/50 px-2 py-0.5 text-[10px] text-gray-400">{cat.method}</span>
                  <span className="text-sm text-gray-500">—</span>
                </div>
              </div>
              {selectedCategory === cat.id && (
                <div className="mt-3 border-t border-gray-800 pt-3">
                  <div className="grid grid-cols-2 gap-4 text-xs">
                    <div>
                      <span className="text-gray-500">Typical share:</span>
                      <span className="ml-1 text-gray-300">{cat.typical}</span>
                    </div>
                    <div>
                      <span className="text-gray-500">Method:</span>
                      <span className="ml-1 text-gray-300">{cat.method}</span>
                    </div>
                    <div>
                      <span className="text-gray-500">Status:</span>
                      <span className="ml-1 rounded bg-gray-700/50 px-1.5 py-0.5 text-gray-400">Not assessed</span>
                    </div>
                    <div>
                      <span className="text-gray-500">Emissions:</span>
                      <span className="ml-1 text-gray-400">No data</span>
                    </div>
                  </div>
                  <div className="mt-3 flex gap-2">
                    <Link
                      href="/emissions"
                      className="rounded bg-primary-600/10 px-3 py-1.5 text-xs font-medium text-primary-400 hover:bg-primary-600/20"
                    >
                      Upload Data
                    </Link>
                    <button className="rounded bg-gray-700/50 px-3 py-1.5 text-xs text-gray-400 hover:text-white">
                      Mark Not Material
                    </button>
                  </div>
                </div>
              )}
            </button>
          ))}
        </div>
      </div>

      {/* Downstream categories */}
      <div className="mb-6">
        <h2 className="mb-3 text-sm font-semibold text-gray-400">Downstream Activities (Categories 9–15)</h2>
        <div className="space-y-2">
          {downstream.map(cat => (
            <button
              key={cat.id}
              onClick={() => setSelectedCategory(selectedCategory === cat.id ? null : cat.id)}
              className="w-full rounded-lg border border-gray-800 bg-gray-800/30 p-4 text-left transition hover:border-gray-700"
            >
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <div className="flex h-7 w-7 items-center justify-center rounded bg-purple-500/10 text-xs font-bold text-purple-400">
                    {cat.id}
                  </div>
                  <div>
                    <div className="text-sm font-medium text-white">{cat.name}</div>
                    <div className="mt-0.5 text-xs text-gray-500">{cat.description}</div>
                  </div>
                </div>
                <div className="flex items-center gap-4">
                  <span className="rounded bg-gray-700/50 px-2 py-0.5 text-[10px] text-gray-400">{cat.method}</span>
                  <span className="text-sm text-gray-500">—</span>
                </div>
              </div>
              {selectedCategory === cat.id && (
                <div className="mt-3 border-t border-gray-800 pt-3">
                  <div className="grid grid-cols-2 gap-4 text-xs">
                    <div>
                      <span className="text-gray-500">Typical share:</span>
                      <span className="ml-1 text-gray-300">{cat.typical}</span>
                    </div>
                    <div>
                      <span className="text-gray-500">Method:</span>
                      <span className="ml-1 text-gray-300">{cat.method}</span>
                    </div>
                    <div>
                      <span className="text-gray-500">Status:</span>
                      <span className="ml-1 rounded bg-gray-700/50 px-1.5 py-0.5 text-gray-400">Not assessed</span>
                    </div>
                    <div>
                      <span className="text-gray-500">Emissions:</span>
                      <span className="ml-1 text-gray-400">No data</span>
                    </div>
                  </div>
                  <div className="mt-3 flex gap-2">
                    <Link
                      href="/emissions"
                      className="rounded bg-primary-600/10 px-3 py-1.5 text-xs font-medium text-primary-400 hover:bg-primary-600/20"
                    >
                      Upload Data
                    </Link>
                    <button className="rounded bg-gray-700/50 px-3 py-1.5 text-xs text-gray-400 hover:text-white">
                      Mark Not Material
                    </button>
                  </div>
                </div>
              )}
            </button>
          ))}
        </div>
      </div>

      {/* Methodology note */}
      <div className="rounded-lg border border-gray-800 bg-gray-900/30 p-4">
        <h3 className="mb-2 text-xs font-semibold text-gray-400">Methodology</h3>
        <p className="text-xs text-gray-500">
          Scope 3 emissions are categorized per the GHG Protocol Corporate Value Chain (Scope 3) Accounting and Reporting Standard.
          Each category supports multiple calculation methods: spend-based (EEIO factors), average-data, supplier-specific, and hybrid approaches.
          Emission factors sourced from UK DEFRA 2024, EPA, and GHG Protocol guidance. Categories should be assessed for materiality
          before detailed calculation.
        </p>
      </div>
    </div>
  );
}
