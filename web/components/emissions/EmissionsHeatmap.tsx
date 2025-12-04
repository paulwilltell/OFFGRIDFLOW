'use client';

import { useEffect, useState } from 'react';
import { api, ApiRequestError } from '../../lib/api';

interface HeatmapCell {
  date: string;
  hour: number;
  value: number;
  intensity: number;
}

interface HeatmapResponse {
  data: HeatmapCell[];
  max: number;
  min: number;
  dates: string[];
  hours: number[];
}

interface EmissionsHeatmapProps {
  height?: number;
  period?: 'week' | 'month';
}

export default function EmissionsHeatmap({ height = 400, period = 'week' }: EmissionsHeatmapProps) {
  const [heatmapData, setHeatmapData] = useState<HeatmapResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchHeatmapData = async () => {
      setLoading(true);
      setError(null);
      
      try {
        const response = await api.get<HeatmapResponse>(`/api/emissions/heatmap?period=${period}`);
        setHeatmapData(response);
      } catch (err) {
        if (err instanceof ApiRequestError) {
          setError(err.message);
        } else {
          setError('Failed to load emissions heatmap data');
        }
      } finally {
        setLoading(false);
      }
    };

    fetchHeatmapData();
  }, [period]);

  if (loading) {
    return (
      <div style={{ height, display: 'flex', alignItems: 'center', justifyContent: 'center', background: '#0f172a', borderRadius: '12px' }}>
        <div style={{ textAlign: 'center' }}>
          <div style={{ marginBottom: '1rem', fontSize: '2rem' }}>🗓️</div>
          <div style={{ color: '#888' }}>Loading emissions heatmap...</div>
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
            Endpoint: /api/emissions/heatmap?period={period}
          </div>
        </div>
      </div>
    );
  }

  if (!heatmapData || heatmapData.data.length === 0) {
    return (
      <div style={{ height, display: 'flex', alignItems: 'center', justifyContent: 'center', background: '#0f172a', borderRadius: '12px' }}>
        <div style={{ textAlign: 'center', padding: '2rem' }}>
          <div style={{ fontSize: '3rem', marginBottom: '1rem' }}>🗓️</div>
          <div style={{ color: '#888', fontSize: '1.1rem', marginBottom: '0.5rem' }}>No temporal pattern data</div>
          <div style={{ color: '#666', fontSize: '0.85rem' }}>Upload hourly emissions data to see patterns</div>
        </div>
      </div>
    );
  }

  const { data, max, min, dates, hours } = heatmapData;

  const dataMap = new Map<string, HeatmapCell>();
  data.forEach(cell => {
    const key = `${cell.date}-${cell.hour}`;
    dataMap.set(key, cell);
  });

  const getColor = (intensity: number): string => {
    if (intensity === 0) return '#1a1f36';
    
    const colors = [
      '#1e3a8a', '#1e40af', '#3b82f6', '#60a5fa',
      '#38bdf8', '#22d3ee', '#06b6d4', '#14b8a6',
      '#10b981', '#84cc16', '#eab308', '#f59e0b',
      '#f97316', '#ef4444',
    ];
    
    const index = Math.floor(intensity * (colors.length - 1));
    return colors[Math.min(index, colors.length - 1)];
  };

  const cellSize = Math.min(40, (height - 100) / 24);
  const cellGap = 2;

  return (
    <div style={{ background: '#0f172a', padding: '1.5rem', borderRadius: '12px', border: '1px solid #1d2940' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
        <h3 style={{ margin: 0, fontSize: '1.1rem', color: '#8aa9ff' }}>
          Emissions Temporal Pattern - {period === 'week' ? 'Last 7 Days' : 'Last 30 Days'}
        </h3>
        <div style={{ fontSize: '0.85rem', color: '#888' }}>
          Range: {min.toFixed(2)} - {max.toFixed(2)} tCO2e
        </div>
      </div>

      <div style={{ overflowX: 'auto', overflowY: 'hidden' }}>
        <div style={{ minWidth: `${dates.length * (cellSize + cellGap) + 100}px` }}>
          <div style={{ display: 'flex', marginBottom: '0.5rem', paddingLeft: '60px' }}>
            {dates.map((date) => (
              <div 
                key={date}
                style={{ 
                  width: cellSize,
                  marginRight: cellGap,
                  fontSize: '0.7rem',
                  color: '#888',
                  textAlign: 'center',
                  transform: 'rotate(-45deg)',
                  transformOrigin: 'left bottom',
                  whiteSpace: 'nowrap',
                }}
              >
                {new Date(date).toLocaleDateString(undefined, { month: 'short', day: 'numeric' })}
              </div>
            ))}
          </div>

          {hours.map((hour) => (
            <div key={hour} style={{ display: 'flex', alignItems: 'center', marginBottom: cellGap }}>
              <div style={{ width: '60px', fontSize: '0.75rem', color: '#888', textAlign: 'right', paddingRight: '0.5rem' }}>
                {hour.toString().padStart(2, '0')}:00
              </div>
              {dates.map((date) => {
                const key = `${date}-${hour}`;
                const cell = dataMap.get(key);
                const intensity = cell?.intensity || 0;
                const value = cell?.value || 0;
                
                return (
                  <div
                    key={key}
                    title={`${date} ${hour}:00 - ${value.toFixed(3)} tCO2e`}
                    style={{
                      width: cellSize,
                      height: cellSize,
                      marginRight: cellGap,
                      background: getColor(intensity),
                      borderRadius: '3px',
                      cursor: value > 0 ? 'pointer' : 'default',
                      transition: 'transform 0.2s',
                    }}
                    onMouseEnter={(e) => {
                      if (value > 0) {
                        e.currentTarget.style.transform = 'scale(1.1)';
                      }
                    }}
                    onMouseLeave={(e) => {
                      e.currentTarget.style.transform = 'scale(1)';
                    }}
                  />
                );
              })}
            </div>
          ))}

          <div style={{ marginTop: '1.5rem', display: 'flex', alignItems: 'center', gap: '0.5rem', paddingLeft: '60px' }}>
            <span style={{ fontSize: '0.8rem', color: '#888' }}>Less</span>
            <div style={{ display: 'flex', gap: '2px' }}>
              {[0, 0.2, 0.4, 0.6, 0.8, 1].map((intensity, idx) => (
                <div
                  key={idx}
                  style={{
                    width: '30px',
                    height: '15px',
                    background: getColor(intensity),
                    borderRadius: '2px',
                  }}
                />
              ))}
            </div>
            <span style={{ fontSize: '0.8rem', color: '#888' }}>More</span>
          </div>
        </div>
      </div>
    </div>
  );
}
