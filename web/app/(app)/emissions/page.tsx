'use client';

import { useState, useRef, type ChangeEvent, type DragEvent } from 'react';
import { useRouter } from 'next/navigation';
import { api } from '@/lib/api';
import { useRequireAuth } from '@/lib/session';
import { recordAuditEvent } from '@/lib/auditLog';

type UploadResponse = { status: string; activitiesSaved: number; orgId: string };

const TEMPLATE_CSV = [
  'meter_id,location,period_start,period_end,kwh',
  'METER-001,US-CA,2025-01-01,2025-01-31,12500',
  'METER-001,US-CA,2025-02-01,2025-02-28,11800',
].join('\n');

export default function UploadPage() {
  const session = useRequireAuth();
  const router = useRouter();
  const inputRef = useRef<HTMLInputElement>(null);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [dragging, setDragging] = useState(false);

  const processFile = async (file: File) => {
    setError(null);

    if (!file.name.toLowerCase().endsWith('.csv')) {
      setError(`"${file.name}" is not a CSV. Export your data as a .csv and try again.`);
      return;
    }
    if (file.size === 0) {
      setError(`"${file.name}" is empty. Export your data and upload the CSV again.`);
      return;
    }
    if (file.size > 50 * 1024 * 1024) {
      setError(`"${file.name}" is larger than the 50 MB limit. Split it into smaller batches by period.`);
      return;
    }

    setUploading(true);
    recordAuditEvent('data.import.started', {
      entityType: 'csv_upload',
      metadata: { file_name: file.name, file_size_bytes: file.size },
    });

    try {
      // Send the raw file to the server, which auto-detects the GHG scope from
      // the columns (electricity / fuel / travel / spend), structures every row,
      // and saves the activities. No client-side schema assumptions.
      const result = await api.upload<UploadResponse>('/api/ingestion/upload/csv', file);

      recordAuditEvent('data.import.completed', {
        entityType: 'csv_upload',
        metadata: { file_name: file.name, activities_saved: result.activitiesSaved },
      });

      if (!result.activitiesSaved) {
        setError('No rows could be structured from that file. Check the header row and try again.');
        setUploading(false);
        return;
      }

      // Success — go to the Review step to see the calculated footprint.
      router.push('/dashboard/carbon');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Upload failed');
      recordAuditEvent('data.import.failed', { entityType: 'csv_upload', metadata: { file_name: file.name } });
      setUploading(false);
    }
  };

  const onInputChange = (e: ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) processFile(file);
    e.target.value = '';
  };

  const onDrop = (e: DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    setDragging(false);
    const file = e.dataTransfer.files?.[0];
    if (file) processFile(file);
  };

  const downloadTemplate = () => {
    const blob = new Blob([TEMPLATE_CSV], { type: 'text/csv' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'offgridflow-template.csv';
    a.click();
    URL.revokeObjectURL(url);
  };

  if (!session?.isAuthenticated) return null;

  return (
    <div className="mx-auto w-full max-w-[720px]">
      <h1 className="mb-[9px] text-center text-[26px] font-bold tracking-[-0.02em]">Add your emissions data</h1>
      <p className="mx-auto mb-9 max-w-[540px] text-center text-[15px] leading-[1.55]" style={{ color: '#6a7a71' }}>
        Upload any activity data — utility bills, fuel, business travel, commuting, freight, waste, or supplier
        spend. We auto-detect the scope, map your columns, apply emission factors, and calculate your
        Scope&nbsp;1, 2 &amp; 3 footprint.
      </p>

      {error && (
        <div className="mb-5 rounded-lg border p-3 text-[13.5px]" style={{ background: '#fef2f2', borderColor: '#fecaca', color: '#991b1b' }}>
          {error}
        </div>
      )}

      {/* Dropzone */}
      <div
        onDragOver={(e) => { e.preventDefault(); setDragging(true); }}
        onDragLeave={() => setDragging(false)}
        onDrop={onDrop}
        onClick={() => !uploading && inputRef.current?.click()}
        className="flex cursor-pointer flex-col items-center rounded-[14px] border-[1.5px] border-dashed bg-white px-10 py-[54px] text-center transition"
        style={{ borderColor: dragging ? '#2f6b50' : '#bcd0c4' }}
      >
        <div className="mb-[18px] flex h-[58px] w-[58px] items-center justify-center rounded-[14px] text-[26px]" style={{ background: '#e8f0ea', color: '#2f6b50' }}>
          ⬆
        </div>
        {uploading ? (
          <>
            <div className="mb-[6px] text-[17px] font-semibold">Uploading…</div>
            <div className="text-[14px]" style={{ color: '#8a978f' }}>
              Structuring your data &amp; detecting scope
            </div>
          </>
        ) : (
          <>
            <div className="mb-[6px] text-[17px] font-semibold">Drag &amp; drop your file here</div>
            <div className="mb-5 text-[14px]" style={{ color: '#8a978f' }}>CSV · up to 50&nbsp;MB</div>
            <div className="flex h-[42px] items-center rounded-[9px] px-[22px] text-[14.5px] font-semibold text-white" style={{ background: '#1d3b2e' }}>
              Browse files
            </div>
          </>
        )}
        <input ref={inputRef} type="file" accept=".csv" className="hidden" onChange={onInputChange} />
      </div>

      {/* Alt sources */}
      <div className="my-[18px] mt-[26px] flex items-center gap-4">
        <span className="h-px flex-1" style={{ background: '#e4e9e5' }} />
        <span className="text-[12.5px] font-medium" style={{ color: '#9aa79f' }}>NOT SURE OF THE FORMAT?</span>
        <span className="h-px flex-1" style={{ background: '#e4e9e5' }} />
      </div>
      <div className="grid grid-cols-1 gap-[14px] sm:grid-cols-2">
        <button onClick={downloadTemplate} className="rounded-[11px] border bg-white p-[18px] text-left" style={{ borderColor: '#e8ece8' }}>
          <div className="mb-[9px] text-[20px]">📄</div>
          <div className="mb-[3px] text-[14.5px] font-semibold">Download CSV template</div>
          <div className="text-[12.5px] leading-[1.45]" style={{ color: '#8a978f' }}>Pre-built sheet with the right columns</div>
        </button>
        <div className="rounded-[11px] border bg-white p-[18px]" style={{ borderColor: '#e8ece8', opacity: 0.65 }}>
          <div className="mb-[9px] text-[20px]">⚡</div>
          <div className="mb-[3px] text-[14.5px] font-semibold">Connect utility account</div>
          <div className="text-[12.5px] leading-[1.45]" style={{ color: '#8a978f' }}>PG&amp;E, Duke, National Grid — coming soon</div>
        </div>
      </div>
    </div>
  );
}
