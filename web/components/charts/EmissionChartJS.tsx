'use client';

import React, { memo, useRef, useEffect, useCallback } from 'react';
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  LineController,
  BarElement,
  BarController,
  Title,
  Tooltip,
  Legend,
  Filler,
  ChartOptions,
  ChartData,
} from 'chart.js';
import annotationPlugin from 'chartjs-plugin-annotation';
import zoomPlugin from 'chartjs-plugin-zoom';
import { EmissionData, Timeframe } from '@/types/carbon';
import { formatNumber, formatDate } from '@/lib/api/carbon';

// Register ChartJS components (controllers required in Chart.js v4)
ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  LineController,
  BarElement,
  BarController,
  Title,
  Tooltip,
  Legend,
  Filler,
  annotationPlugin,
  zoomPlugin
);

// ============================================================================
// Types
// ============================================================================

interface ChartPoint {
  datasetIndex: number;
  index: number;
  value: EmissionData;
}

interface EmissionChartProps {
  data: EmissionData[];
  timeframe: Timeframe;
  height?: number;
  onPointClick?: (point: ChartPoint) => void;
}

// ============================================================================
// Component
// ============================================================================

export const EmissionChartJS: React.FC<EmissionChartProps> = memo(({
  data,
  timeframe,
  height = 400,
  onPointClick
}) => {
  const chartRef = useRef<ChartJS | null>(null);
  const canvasRef = useRef<HTMLCanvasElement>(null);

  // Return only real data. Never fabricate historical points — showing a
  // customer 12 months of randomly-generated emissions values is a direct
  // compliance and legal liability in a carbon reporting tool.
  const chartData = useCallback((): EmissionData[] => {
    return Array.isArray(data) ? data : [];
  }, [data]);

  useEffect(() => {
    if (!canvasRef.current) return;

    const ctx = canvasRef.current.getContext('2d');
    if (!ctx) return;

    // Destroy previous chart
    if (chartRef.current) {
      chartRef.current.destroy();
    }

    const emissions = chartData();
    if (emissions.length === 0) return;

    // Create gradient for fill
    const gradient = ctx.createLinearGradient(0, 0, 0, height);
    gradient.addColorStop(0, 'rgba(0, 229, 197, 0.3)');
    gradient.addColorStop(1, 'rgba(0, 229, 197, 0.0)');

    const chartDataConfig: ChartData<'line'> = {
      labels: emissions.map(d => formatDate(d.updatedAt, timeframe)),
      datasets: [
        {
          label: 'Scope 1 (Direct)',
          data: emissions.map(d => d.scope1),
          borderColor: 'rgb(0, 229, 197)',
          backgroundColor: gradient,
          borderWidth: 3,
          tension: 0.4,
          fill: true,
          pointRadius: 4,
          pointHoverRadius: 6,
        },
        {
          label: 'Scope 2 (Energy)',
          data: emissions.map(d => d.scope2),
          borderColor: 'rgb(139, 92, 246)',
          backgroundColor: 'transparent',
          borderWidth: 2,
          tension: 0.4,
          fill: false,
          pointRadius: 3,
          pointHoverRadius: 5,
        },
        {
          label: 'Scope 3 (Value Chain)',
          data: emissions.map(d => d.scope3),
          borderColor: 'rgb(251, 191, 36)',
          backgroundColor: 'transparent',
          borderWidth: 2,
          tension: 0.4,
          fill: false,
          pointRadius: 3,
          pointHoverRadius: 5,
        }
      ]
    };

    const options: ChartOptions<'line'> = {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: {
          position: 'top',
          labels: {
            color: '#94A3B8',
            font: {
              size: 12,
              family: 'Inter, system-ui, sans-serif'
            },
            usePointStyle: true,
            pointStyle: 'circle',
            padding: 20,
          }
        },
        tooltip: {
          mode: 'index',
          intersect: false,
          backgroundColor: 'rgba(15, 23, 42, 0.95)',
          titleColor: '#F1F5F9',
          bodyColor: '#CBD5E1',
          borderColor: 'rgba(0, 229, 197, 0.3)',
          borderWidth: 1,
          padding: 12,
          cornerRadius: 8,
          displayColors: true,
          callbacks: {
            label: (context) => {
              let label = context.dataset.label || '';
              if (label) {
                label += ': ';
              }
              const yValue = context.parsed.y ?? 0;
              label += formatNumber(yValue) + ' tCO₂e';
              return label;
            },
            footer: (tooltipItems) => {
              const total = tooltipItems.reduce((sum, item) => sum + (item.parsed.y ?? 0), 0);
              return `Total: ${formatNumber(total)} tCO₂e`;
            }
          }
        },
        annotation: {
          annotations: {
            targetLine: {
              type: 'line',
              yMin: 5000,
              yMax: 5000,
              borderColor: 'rgba(239, 68, 68, 0.5)',
              borderWidth: 2,
              borderDash: [5, 5],
              label: {
                content: 'Target',
                display: true,
                position: 'end',
                backgroundColor: 'rgba(239, 68, 68, 0.8)',
                color: '#fff',
                font: {
                  size: 10,
                },
                padding: 4,
              }
            }
          }
        },
        zoom: {
          pan: {
            enabled: true,
            mode: 'x'
          },
          zoom: {
            wheel: {
              enabled: true
            },
            pinch: {
              enabled: true
            },
            mode: 'x'
          }
        }
      },
      scales: {
        x: {
          grid: {
            color: 'rgba(255, 255, 255, 0.05)',
          },
          border: {
            display: false,
          },
          ticks: {
            color: '#64748B',
            font: {
              size: 11
            },
            maxRotation: 45
          }
        },
        y: {
          beginAtZero: false,
          grid: {
            color: 'rgba(255, 255, 255, 0.05)',
          },
          border: {
            display: false,
          },
          ticks: {
            color: '#64748B',
            font: {
              size: 11
            },
            callback: (value) => {
              return formatNumber(value as number) + ' t';
            }
          },
          title: {
            display: true,
            text: 'Emissions (tCO₂e)',
            color: '#94A3B8',
            font: {
              size: 12,
              weight: 'bold'
            }
          }
        }
      },
      interaction: {
        intersect: false,
        mode: 'nearest'
      },
      onClick: (event, elements) => {
        if (elements.length > 0 && onPointClick) {
          const element = elements[0];
          const point: ChartPoint = {
            datasetIndex: element.datasetIndex,
            index: element.index,
            value: emissions[element.index]
          };
          onPointClick(point);
        }
      },
      animation: {
        duration: 1000,
        easing: 'easeOutQuart'
      }
    };

    chartRef.current = new ChartJS(ctx, {
      type: 'line',
      data: chartDataConfig,
      options,
    });

    // Cleanup on unmount
    return () => {
      if (chartRef.current) {
        chartRef.current.destroy();
      }
    };
  }, [chartData, timeframe, height, onPointClick]);

  return (
    <div 
      className="bg-gray-800/50 rounded-xl border border-gray-700/50 p-6"
      role="region"
      aria-label="Emission trends chart"
    >
      <div className="flex items-center justify-between mb-4">
        <h3 
          id="chart-title"
          className="text-sm font-semibold text-gray-400 uppercase tracking-wider"
        >
          Emission Trends
        </h3>
        <div 
          className="flex items-center gap-2"
          role="group"
          aria-label="Time range selector"
        >
          <button 
            className="px-3 py-1 text-xs font-medium text-gray-400 hover:text-white hover:bg-gray-700 rounded transition-colors focus:outline-none focus:ring-2 focus:ring-cyan-500 focus:ring-offset-1 focus:ring-offset-gray-900"
            aria-label="Show 1 month of data"
          >
            1M
          </button>
          <button 
            className="px-3 py-1 text-xs font-medium text-gray-400 hover:text-white hover:bg-gray-700 rounded transition-colors focus:outline-none focus:ring-2 focus:ring-cyan-500 focus:ring-offset-1 focus:ring-offset-gray-900"
            aria-label="Show 3 months of data"
          >
            3M
          </button>
          <button 
            className="px-3 py-1 text-xs font-medium bg-green-500/20 text-green-400 rounded focus:outline-none focus:ring-2 focus:ring-cyan-500 focus:ring-offset-1 focus:ring-offset-gray-900"
            aria-label="Show 1 year of data"
            aria-pressed="true"
          >
            1Y
          </button>
          <button 
            className="px-3 py-1 text-xs font-medium text-gray-400 hover:text-white hover:bg-gray-700 rounded transition-colors focus:outline-none focus:ring-2 focus:ring-cyan-500 focus:ring-offset-1 focus:ring-offset-gray-900"
            aria-label="Show all available data"
          >
            All
          </button>
        </div>
      </div>
      
      <div style={{ position: 'relative', height: `${height}px` }}>
        <canvas
          ref={canvasRef}
          style={{ width: '100%', height: '100%' }}
          role="img"
          aria-labelledby="chart-title"
          aria-describedby="chart-description"
        />
        {/* Empty-state overlay — shown when there is insufficient real data
            to plot a trend. We do not fabricate historical points. */}
        {data.length < 2 && (
          <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
            <div className="pointer-events-auto max-w-sm rounded-xl border border-gray-700 bg-gray-900/80 backdrop-blur-sm px-5 py-4 text-center">
              <div className="text-sm font-medium text-white">
                Historical trends will appear once you have multiple reporting periods
              </div>
              <div className="mt-1 text-xs text-gray-400">
                Upload additional monthly or quarterly data to see emissions over time.
              </div>
            </div>
          </div>
        )}
        {/* Screen reader accessible data summary */}
        <div id="chart-description" className="sr-only">
          {data.length < 2
            ? 'Not enough data to display a trend chart. Upload additional reporting periods to view historical emissions.'
            : 'Line chart showing carbon emission trends over time. Displays Scope 1 (direct emissions), Scope 2 (energy), and Scope 3 (value chain) emissions. Use keyboard to navigate data points. Scroll to zoom, drag to pan.'}
        </div>
      </div>

      {/* Chart footer */}
      <div 
        className="mt-4 pt-4 border-t border-gray-700/50 flex items-center justify-between text-xs text-gray-500"
        aria-hidden="true"
      >
        <span>Scroll to zoom • Drag to pan</span>
        <span>Data updated in real-time</span>
      </div>
    </div>
  );
});

EmissionChartJS.displayName = 'EmissionChartJS';

export default EmissionChartJS;
