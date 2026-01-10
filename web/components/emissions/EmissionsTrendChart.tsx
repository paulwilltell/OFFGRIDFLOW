'use client';

import { useEffect, useState } from 'react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import { api, ApiRequestError } from '../../lib/api';

interface TrendDataPoint {
  date: string;
  scope1: number;
  scope2: number;
  scope3: number;
  total: number;
}

interface TrendResponse {
  data: TrendDataPoint[];
  period: string;
}

interface EmissionsTrendChartProps {
  period?: 'week' | 'month' | 'quarter' | 'year';
  height?: number;
}

export default function EmissionsTrendChart({ period = 'year', height = 400 }: EmissionsTrendChartProps) {
  const [data, setData] = useState<TrendDataPoint[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchTrendData = async () => {
      setLoading(true);
      setError(null);
      
      try {
        const response = await api.get<TrendResponse>(`/api/emissions/trend?period=${period}`);
        setData(response.data || []);
      } catch (err) {
        if (err instanceof ApiRequestError) {
          setError(err.message);
        } else {
          setError('Failed to load emissions trend data');
        }
      } finally {
        setLoading(false);
      }
    };

    fetchTrendData();
  }, [period]);

  if (loading) {
    return (
      <div style={{ height, display: 'flex', alignItems: 'center', justifyContent: 'center', background: '#0f172a', borderRadius: '12px' }}>
        <div style={{ textAlign: 'center' }}>
          <div style={{ marginBottom: '1rem', fontSize: '2rem' }}>📊</div>
          <div style={{ color: '#888' }}>Loading emissions trend...</div>
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
            Endpoint: /api/emissions/trend?period={period}
          </div>
        </div>
      </div>
    );
  }

  if (data.length === 0) {
    return (
      <div style={{ height, display: 'flex', alignItems: 'center', justifyContent: 'center', background: '#0f172a', borderRadius: '12px' }}>
        <div style={{ textAlign: 'center', padding: '2rem' }}>
          <div style={{ fontSize: '3rem', marginBottom: '1rem' }}>📭</div>
          <div style={{ color: '#888', fontSize: '1.1rem', marginBottom: '0.5rem' }}>No trend data available</div>
          <div style={{ color: '#666', fontSize: '0.85rem' }}>Upload emissions data to see trends over time</div>
        </div>
      </div>
    );
  }

  return (
    <div style={{ background: '#0f172a', padding: '1.5rem', borderRadius: '12px', border: '1px solid #1d2940' }}>
      <h3 style={{ margin: '0 0 1rem 0', fontSize: '1.1rem', color: '#8aa9ff' }}>
        Emissions Trend - {period.charAt(0).toUpperCase() + period.slice(1)}
      </h3>
      <ResponsiveContainer width="100%" height={height}>
        <LineChart data={data} margin={{ top: 5, right: 30, left: 20, bottom: 5 }}>
          <CartesianGrid strokeDasharray="3 3" stroke="#1d2940" />
          <XAxis 
            dataKey="date" 
            stroke="#888" 
            style={{ fontSize: '0.85rem' }}
            tickFormatter={(value) => {
              const date = new Date(value);
              return date.toLocaleDateString(undefined, { month: 'short', day: 'numeric' });
            }}
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
            labelFormatter={(value) => new Date(value).toLocaleDateString()}
            formatter={(value) => [`${(value as number)?.toFixed(2) ?? '0.00'} tCO2e`, '']}
          />
          <Legend 
            wrapperStyle={{ fontSize: '0.9rem' }}
            iconType="line"
          />
          <Line 
            type="monotone" 
            dataKey="scope1" 
            stroke="#ef4444" 
            strokeWidth={2}
            name="Scope 1"
            dot={{ fill: '#ef4444', r: 4 }}
            activeDot={{ r: 6 }}
          />
          <Line 
            type="monotone" 
            dataKey="scope2" 
            stroke="#3b82f6" 
            strokeWidth={2}
            name="Scope 2"
            dot={{ fill: '#3b82f6', r: 4 }}
            activeDot={{ r: 6 }}
          />
          <Line 
            type="monotone" 
            dataKey="scope3" 
            stroke="#10b981" 
            strokeWidth={2}
            name="Scope 3"
            dot={{ fill: '#10b981', r: 4 }}
            activeDot={{ r: 6 }}
          />
          <Line 
            type="monotone" 
            dataKey="total" 
            stroke="#8aa9ff" 
            strokeWidth={3}
            name="Total"
            dot={{ fill: '#8aa9ff', r: 5 }}
            activeDot={{ r: 7 }}
          />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}
