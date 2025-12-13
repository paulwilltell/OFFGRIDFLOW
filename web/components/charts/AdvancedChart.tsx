'use client';

import React, { memo, useState, useMemo, useCallback } from 'react';

interface ChartDataPoint {
  label: string;
  value: number;
  secondaryValue?: number;
  category?: string;
  date?: Date;
}

interface AdvancedChartProps {
  data: ChartDataPoint[];
  type?: 'line' | 'area' | 'bar' | 'scatter' | 'radar';
  title?: string;
  xAxisLabel?: string;
  yAxisLabel?: string;
  showLegend?: boolean;
  showGrid?: boolean;
  animate?: boolean;
  height?: number;
  colors?: string[];
  onDataPointClick?: (point: ChartDataPoint, index: number) => void;
}

// SVG-based advanced chart with multiple visualization types
export const AdvancedChart = memo(function AdvancedChart({
  data,
  type = 'line',
  title,
  xAxisLabel,
  yAxisLabel,
  showLegend = true,
  showGrid = true,
  animate = true,
  height = 300,
  colors = ['#22c55e', '#3b82f6', '#f59e0b', '#ef4444', '#8b5cf6'],
  onDataPointClick,
}: AdvancedChartProps) {
  const [hoveredIndex, setHoveredIndex] = useState<number | null>(null);
  const [selectedIndex, setSelectedIndex] = useState<number | null>(null);

  const padding = { top: 40, right: 40, bottom: 60, left: 60 };
  const width = 600;
  const chartWidth = width - padding.left - padding.right;
  const chartHeight = height - padding.top - padding.bottom;

  // Calculate scales
  const { xScale, yScale, maxValue, minValue } = useMemo(() => {
    const values = data.map(d => d.value);
    const max = Math.max(...values) * 1.1;
    const min = Math.min(0, Math.min(...values));
    
    return {
      xScale: (i: number) => (i / (data.length - 1 || 1)) * chartWidth,
      yScale: (v: number) => chartHeight - ((v - min) / (max - min || 1)) * chartHeight,
      maxValue: max,
      minValue: min,
    };
  }, [data, chartWidth, chartHeight]);

  // Generate path for line/area charts
  const linePath = useMemo(() => {
    if (data.length === 0) return '';
    
    const points = data.map((d, i) => ({
      x: xScale(i),
      y: yScale(d.value),
    }));

    // Create smooth curve using cubic bezier
    let path = `M ${points[0].x} ${points[0].y}`;
    
    for (let i = 1; i < points.length; i++) {
      const prev = points[i - 1];
      const curr = points[i];
      const tension = 0.3;
      
      const cp1x = prev.x + (curr.x - prev.x) * tension;
      const cp1y = prev.y;
      const cp2x = curr.x - (curr.x - prev.x) * tension;
      const cp2y = curr.y;
      
      path += ` C ${cp1x} ${cp1y}, ${cp2x} ${cp2y}, ${curr.x} ${curr.y}`;
    }

    return path;
  }, [data, xScale, yScale]);

  // Generate area path
  const areaPath = useMemo(() => {
    if (data.length === 0 || type !== 'area') return '';
    
    const firstX = xScale(0);
    const lastX = xScale(data.length - 1);
    const baseY = yScale(0);
    
    return `${linePath} L ${lastX} ${baseY} L ${firstX} ${baseY} Z`;
  }, [linePath, xScale, yScale, data.length, type]);

  // Generate grid lines
  const gridLines = useMemo(() => {
    if (!showGrid) return { horizontal: [], vertical: [] };
    
    const horizontal = [];
    const vertical = [];
    const gridCount = 5;

    for (let i = 0; i <= gridCount; i++) {
      const y = (i / gridCount) * chartHeight;
      const value = maxValue - (i / gridCount) * (maxValue - minValue);
      horizontal.push({ y, value });
    }

    for (let i = 0; i < data.length; i += Math.ceil(data.length / 6)) {
      const x = xScale(i);
      vertical.push({ x, label: data[i]?.label || '' });
    }

    return { horizontal, vertical };
  }, [showGrid, chartHeight, maxValue, minValue, data, xScale]);

  const handlePointClick = useCallback((point: ChartDataPoint, index: number) => {
    setSelectedIndex(index === selectedIndex ? null : index);
    onDataPointClick?.(point, index);
  }, [selectedIndex, onDataPointClick]);

  // Render different chart types
  const renderChart = () => {
    switch (type) {
      case 'bar':
        return renderBarChart();
      case 'scatter':
        return renderScatterChart();
      case 'radar':
        return renderRadarChart();
      case 'area':
        return renderAreaChart();
      case 'line':
      default:
        return renderLineChart();
    }
  };

  const renderLineChart = () => (
    <g>
      {/* Line */}
      <path
        d={linePath}
        fill="none"
        stroke={colors[0]}
        strokeWidth={2}
        strokeLinecap="round"
        strokeLinejoin="round"
        className={animate ? 'animate-draw' : ''}
      />
      
      {/* Data points */}
      {data.map((point, i) => (
        <g key={i}>
          <circle
            cx={xScale(i)}
            cy={yScale(point.value)}
            r={hoveredIndex === i || selectedIndex === i ? 6 : 4}
            fill={hoveredIndex === i || selectedIndex === i ? colors[0] : '#1f2937'}
            stroke={colors[0]}
            strokeWidth={2}
            className="cursor-pointer transition-all duration-150"
            onMouseEnter={() => setHoveredIndex(i)}
            onMouseLeave={() => setHoveredIndex(null)}
            onClick={() => handlePointClick(point, i)}
          />
        </g>
      ))}
    </g>
  );

  const renderAreaChart = () => (
    <g>
      {/* Gradient definition */}
      <defs>
        <linearGradient id="areaGradient" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor={colors[0]} stopOpacity={0.4} />
          <stop offset="100%" stopColor={colors[0]} stopOpacity={0} />
        </linearGradient>
      </defs>
      
      {/* Area fill */}
      <path
        d={areaPath}
        fill="url(#areaGradient)"
        className={animate ? 'animate-fade-in' : ''}
      />
      
      {/* Line */}
      <path
        d={linePath}
        fill="none"
        stroke={colors[0]}
        strokeWidth={2}
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      
      {/* Data points */}
      {data.map((point, i) => (
        <circle
          key={i}
          cx={xScale(i)}
          cy={yScale(point.value)}
          r={hoveredIndex === i ? 5 : 3}
          fill={colors[0]}
          className="cursor-pointer transition-all duration-150"
          onMouseEnter={() => setHoveredIndex(i)}
          onMouseLeave={() => setHoveredIndex(null)}
          onClick={() => handlePointClick(point, i)}
        />
      ))}
    </g>
  );

  const renderBarChart = () => {
    const barWidth = (chartWidth / data.length) * 0.7;
    const barGap = (chartWidth / data.length) * 0.3;
    
    return (
      <g>
        {data.map((point, i) => {
          const barHeight = chartHeight - yScale(point.value);
          const x = (i / data.length) * chartWidth + barGap / 2;
          const isActive = hoveredIndex === i || selectedIndex === i;
          
          return (
            <g key={i}>
              <rect
                x={x}
                y={yScale(point.value)}
                width={barWidth}
                height={barHeight}
                fill={isActive ? colors[0] : `${colors[0]}cc`}
                rx={4}
                className="cursor-pointer transition-all duration-150"
                onMouseEnter={() => setHoveredIndex(i)}
                onMouseLeave={() => setHoveredIndex(null)}
                onClick={() => handlePointClick(point, i)}
                style={animate ? {
                  animation: `grow-up 0.5s ease-out ${i * 0.05}s both`,
                  transformOrigin: 'bottom',
                } : undefined}
              />
              {isActive && (
                <rect
                  x={x}
                  y={yScale(point.value)}
                  width={barWidth}
                  height={barHeight}
                  fill="none"
                  stroke={colors[0]}
                  strokeWidth={2}
                  rx={4}
                />
              )}
            </g>
          );
        })}
      </g>
    );
  };

  const renderScatterChart = () => (
    <g>
      {data.map((point, i) => {
        const isActive = hoveredIndex === i || selectedIndex === i;
        
        return (
          <g key={i}>
            {isActive && (
              <circle
                cx={xScale(i)}
                cy={yScale(point.value)}
                r={12}
                fill={`${colors[0]}33`}
                className="animate-pulse"
              />
            )}
            <circle
              cx={xScale(i)}
              cy={yScale(point.value)}
              r={isActive ? 8 : 6}
              fill={colors[Math.floor(point.value / maxValue * colors.length) % colors.length]}
              stroke="#1f2937"
              strokeWidth={2}
              className="cursor-pointer transition-all duration-150"
              onMouseEnter={() => setHoveredIndex(i)}
              onMouseLeave={() => setHoveredIndex(null)}
              onClick={() => handlePointClick(point, i)}
            />
          </g>
        );
      })}
    </g>
  );

  const renderRadarChart = () => {
    const centerX = chartWidth / 2;
    const centerY = chartHeight / 2;
    const maxRadius = Math.min(chartWidth, chartHeight) / 2 - 20;
    const angleStep = (2 * Math.PI) / data.length;

    // Generate polygon points
    const points = data.map((point, i) => {
      const angle = i * angleStep - Math.PI / 2;
      const radius = (point.value / maxValue) * maxRadius;
      return {
        x: centerX + radius * Math.cos(angle),
        y: centerY + radius * Math.sin(angle),
      };
    });

    const pathD = points.map((p, i) => 
      `${i === 0 ? 'M' : 'L'} ${p.x} ${p.y}`
    ).join(' ') + ' Z';

    return (
      <g>
        {/* Grid rings */}
        {[0.25, 0.5, 0.75, 1].map((scale) => (
          <polygon
            key={scale}
            points={data.map((_, i) => {
              const angle = i * angleStep - Math.PI / 2;
              const radius = scale * maxRadius;
              return `${centerX + radius * Math.cos(angle)},${centerY + radius * Math.sin(angle)}`;
            }).join(' ')}
            fill="none"
            stroke="rgba(75, 85, 99, 0.3)"
            strokeWidth={1}
          />
        ))}
        
        {/* Axis lines */}
        {data.map((_, i) => {
          const angle = i * angleStep - Math.PI / 2;
          return (
            <line
              key={i}
              x1={centerX}
              y1={centerY}
              x2={centerX + maxRadius * Math.cos(angle)}
              y2={centerY + maxRadius * Math.sin(angle)}
              stroke="rgba(75, 85, 99, 0.3)"
              strokeWidth={1}
            />
          );
        })}
        
        {/* Data polygon */}
        <defs>
          <linearGradient id="radarGradient" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor={colors[0]} stopOpacity={0.4} />
            <stop offset="100%" stopColor={colors[0]} stopOpacity={0.1} />
          </linearGradient>
        </defs>
        <path
          d={pathD}
          fill="url(#radarGradient)"
          stroke={colors[0]}
          strokeWidth={2}
        />
        
        {/* Data points */}
        {points.map((point, i) => (
          <circle
            key={i}
            cx={point.x}
            cy={point.y}
            r={hoveredIndex === i ? 6 : 4}
            fill={colors[0]}
            stroke="#1f2937"
            strokeWidth={2}
            className="cursor-pointer"
            onMouseEnter={() => setHoveredIndex(i)}
            onMouseLeave={() => setHoveredIndex(null)}
            onClick={() => handlePointClick(data[i], i)}
          />
        ))}
        
        {/* Labels */}
        {data.map((point, i) => {
          const angle = i * angleStep - Math.PI / 2;
          const labelRadius = maxRadius + 20;
          const x = centerX + labelRadius * Math.cos(angle);
          const y = centerY + labelRadius * Math.sin(angle);
          
          return (
            <text
              key={i}
              x={x}
              y={y}
              textAnchor="middle"
              dominantBaseline="middle"
              className="text-xs fill-gray-400"
            >
              {point.label}
            </text>
          );
        })}
      </g>
    );
  };

  return (
    <div className="bg-gray-800/50 rounded-xl border border-gray-700/50 p-6">
      {title && (
        <h3 className="text-sm font-semibold text-gray-400 uppercase tracking-wider mb-4">
          {title}
        </h3>
      )}
      
      <svg
        viewBox={`0 0 ${width} ${height}`}
        className="w-full"
        style={{ height }}
      >
        {/* Chart area */}
        <g transform={`translate(${padding.left}, ${padding.top})`}>
          {/* Grid */}
          {showGrid && type !== 'radar' && (
            <g>
              {gridLines.horizontal.map((line, i) => (
                <g key={i}>
                  <line
                    x1={0}
                    y1={line.y}
                    x2={chartWidth}
                    y2={line.y}
                    stroke="rgba(75, 85, 99, 0.3)"
                    strokeDasharray="4,4"
                  />
                  <text
                    x={-10}
                    y={line.y}
                    textAnchor="end"
                    dominantBaseline="middle"
                    className="text-xs fill-gray-500"
                  >
                    {line.value.toFixed(0)}
                  </text>
                </g>
              ))}
              {gridLines.vertical.map((line, i) => (
                <g key={i}>
                  <line
                    x1={line.x}
                    y1={0}
                    x2={line.x}
                    y2={chartHeight}
                    stroke="rgba(75, 85, 99, 0.2)"
                    strokeDasharray="4,4"
                  />
                  <text
                    x={line.x}
                    y={chartHeight + 20}
                    textAnchor="middle"
                    className="text-xs fill-gray-500"
                  >
                    {line.label}
                  </text>
                </g>
              ))}
            </g>
          )}
          
          {/* Chart content */}
          {renderChart()}
        </g>
        
        {/* Axis labels */}
        {xAxisLabel && (
          <text
            x={width / 2}
            y={height - 10}
            textAnchor="middle"
            className="text-xs fill-gray-400"
          >
            {xAxisLabel}
          </text>
        )}
        {yAxisLabel && (
          <text
            x={15}
            y={height / 2}
            textAnchor="middle"
            transform={`rotate(-90, 15, ${height / 2})`}
            className="text-xs fill-gray-400"
          >
            {yAxisLabel}
          </text>
        )}
      </svg>

      {/* Tooltip */}
      {hoveredIndex !== null && data[hoveredIndex] && (
        <div className="mt-4 p-3 bg-gray-900/80 rounded-lg border border-gray-700">
          <div className="text-sm font-medium text-white">
            {data[hoveredIndex].label}
          </div>
          <div className="text-lg font-bold text-green-400">
            {data[hoveredIndex].value.toLocaleString()}
          </div>
          {data[hoveredIndex].category && (
            <div className="text-xs text-gray-400 mt-1">
              {data[hoveredIndex].category}
            </div>
          )}
        </div>
      )}

      {/* Legend */}
      {showLegend && (
        <div className="mt-4 flex items-center justify-center gap-4 flex-wrap">
          <div className="flex items-center gap-2">
            <span 
              className="w-3 h-3 rounded-full" 
              style={{ backgroundColor: colors[0] }} 
            />
            <span className="text-xs text-gray-400">Value</span>
          </div>
        </div>
      )}

      {/* CSS animations */}
      <style jsx>{`
        @keyframes grow-up {
          from {
            transform: scaleY(0);
          }
          to {
            transform: scaleY(1);
          }
        }
        .animate-draw {
          stroke-dasharray: 2000;
          stroke-dashoffset: 2000;
          animation: draw 1.5s ease-out forwards;
        }
        @keyframes draw {
          to {
            stroke-dashoffset: 0;
          }
        }
        .animate-fade-in {
          animation: fade-in 0.5s ease-out forwards;
        }
        @keyframes fade-in {
          from {
            opacity: 0;
          }
          to {
            opacity: 1;
          }
        }
      `}</style>
    </div>
  );
});

export default AdvancedChart;
