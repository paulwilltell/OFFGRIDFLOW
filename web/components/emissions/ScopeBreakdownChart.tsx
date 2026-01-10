'use client';

import { useEffect, useState } from 'react';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer, Cell } from 'recharts';
import { api, ApiRequestError } from '../../lib/api';

interface ScopeData {
  scope: string;
  emissions: number;
  percentage: number;
  activities: number;
}

interface ScopeBreakdownResponse {
  data: ScopeData[];
  total: number;
}

interface ScopeBreakdownChartProps {
  height?: number;
  startDate?: string;
  endDate?: string;
}

const SCOPE_COLORS = {
  'Scope 1': '#ef4444',
  'Scope 2': '#3b82f6',
  'Scope 3': '#10b981',
};

export default function ScopeBreakdownChart({ height = 400, startDate, endDate }: ScopeBreakdownChartProps) {
  const [data, setData] = useState<ScopeData[]>([]);
  const [total, setTotal] = useState<number>(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchScopeData = async () => {
      setLoading(true);
      setError(null);
      
      try {
        const params = new URLSearchParams();
        if (startDate) params.set('start_date', startDate);
        if (endDate) params.set('end_date', endDate);
        
        const queryString = params.toString();
        const endpoint = `/api/emissions/scopes${queryString ? `?${queryString}` : ''}`;
        
        const response = await api.get<ScopeBreakdownResponse>(endpoint);
        setData(response.data || []);
        setTotal(response.total || 0);
      } catch (err) {
        if (err instanceof ApiRequestError) {
          setError(err.message);
        } else {
          setError('Failed to load scope breakdown data');
        }
      } finally {
        setLoading(false);
      }
    };

    fetchScopeData();
  }, [startDate, endDate]);

  if (loading) {
    return (
      <div style={{ height, display: 'flex', alignItems: 'center', justifyContent: 'center', background: '#0f172a', borderRadius: '12px' }}>
        <div style={{ textAlign: 'center' }}>
          <div style={{ marginBottom: '1rem', fontSize: '2rem' }}>📊</div>
          <div style={{ color: '#888' }}>Loading scope breakdown...</div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div style={{ height, display: 'flex', alignItems: 'center', justifyContent: 'center', background: '#0f172a', borderRadius: '12px', border: '1px solid #7f1d1d' }}>
        <div style={{ textAlign: 'center', padding: '2rem' }}>
          <div style={{ color: '#fecaca', marginBottom: '0.5rem', fontSize: '1.1rem' }}>⚠️ Error</div>
          <div style={{ color: '#fca5a5', fontSize: '0.9rem' }}>{error}</div>
          <div style={{ color: '#888', fontSize: '0.8rem', marginTop: '0.5rem' }}>
            Endpoint: /api/emissions/scopes
          </div>
        </div>
      </div>
    );
  }

  if (data.length === 0) {
    return (
      <div style={{ height, display: 'flex', alignItems: 'center', justifyContent: 'center', background: '#0f172a', borderRadius: '12px' }}>
        <div style={{ textAlign: 'center', padding: '2rem' }}>
          <div style={{ fontSize: '3rem', marginBottom: '1rem' }}>📊</div>
          <div style={{ color: '#888', fontSize: '1.1rem', marginBottom: '0.5rem' }}>No scope data available</div>
          <div style={{ color: '#666', fontSize: '0.85rem' }}>Upload emissions data to see scope breakdown</div>
        </div>
      </div>
    );
  }

  return (
    <div style={{ background: '#0f172a', padding: '1.5rem', borderRadius: '12px', border: '1px solid #1d2940' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
        <h3 style={{ margin: 0, fontSize: '1.1rem', color: '#8aa9ff' }}>
          Emissions by Scope
        </h3>
        <div style={{ fontSize: '0.9rem', color: '#888' }}>
          Total: <span style={{ color: '#fff', fontWeight: 600 }}>{total.toFixed(2)} tCO2e</span>
        </div>
      </div>
      
      <ResponsiveContainer width="100%" height={height}>
        <BarChart data={data} margin={{ top: 20, right: 30, left: 20, bottom: 5 }}>
          <CartesianGrid strokeDasharray="3 3" stroke="#1d2940" />
          <XAxis 
            dataKey="scope" 
            stroke="#888" 
            style={{ fontSize: '0.9rem' }}
          />
          <YAxis 
            stroke="#888" 
            style={{ fontSize: '0.85rem' }}
            label={{ value: 'tCO2e', angle: -90, position: 'insideLeft', style: { fill: '#888' } }}
          />
          <Tooltip 
            contentStyle={{ 
              background: '#1d2940', 
              border: '1px solid #374151', 
              borderRadius: '8px',
              color: '#fff'
            }}
            formatter={(value, name, props) => {
              const numValue = (value as number) ?? 0;
              const percentage = props.payload?.percentage ?? 0;
              return [
                <>
                  <div>{numValue.toFixed(2)} tCO2e</div>
                  <div style={{ fontSize: '0.85rem', color: '#888' }}>{percentage.toFixed(1)}% of total</div>
                  <div style={{ fontSize: '0.85rem', color: '#888' }}>{props.payload?.activities ?? 0} activities</div>
                </>,
                ''
              ];
            }}
            labelStyle={{ color: '#8aa9ff', fontWeight: 600 }}
          />
          <Legend 
            wrapperStyle={{ fontSize: '0.9rem' }}
            content={() => null}
          />
          <Bar dataKey="emissions" name="Emissions" radius={[8, 8, 0, 0]}>
            {data.map((entry, index) => (
              <Cell key={`cell-${index}`} fill={SCOPE_COLORS[entry.scope as keyof typeof SCOPE_COLORS] || '#8aa9ff'} />
            ))}
          </Bar>
        </BarChart>
      </ResponsiveContainer>

      {/* Percentage breakdown */}
      <div style={{ display: 'flex', gap: '1rem', marginTop: '1rem', flexWrap: 'wrap' }}>
        {data.map((scope) => (
          <div 
            key={scope.scope}
            style={{
              flex: '1',
              minWidth: '150px',
              padding: '0.75rem',
              background: '#1d2940',
              borderRadius: '8px',
              borderLeft: `4px solid ${SCOPE_COLORS[scope.scope as keyof typeof SCOPE_COLORS] || '#8aa9ff'}`,
            }}
          >
            <div style={{ fontSize: '0.85rem', color: '#888', marginBottom: '0.25rem' }}>{scope.scope}</div>
            <div style={{ fontSize: '1.3rem', fontWeight: 700 }}>{scope.percentage.toFixed(1)}%</div>
            <div style={{ fontSize: '0.75rem', color: '#666', marginTop: '0.25rem' }}>
              {scope.emissions.toFixed(2)} tCO2e • {scope.activities} activities
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
