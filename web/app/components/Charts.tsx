'use client';

import {
  LineChart,
  Line,
  BarChart,
  Bar,
  PieChart,
  Pie,
  Cell,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  TooltipProps,
} from 'recharts';
import { Box, Button, ButtonGroup, useColorMode } from '@chakra-ui/react';
import html2canvas from 'html2canvas';
import jsPDF from 'jspdf';
import { useRef } from 'react';

const COLORS = ['#059669', '#0ea5e9', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899'];

interface EmissionTrendData {
  date: string;
  scope1: number;
  scope2: number;
  scope3: number;
  total: number;
}

interface ScopeBreakdownData {
  scope: string;
  emissions: number;
  percentage: number;
}

interface EmissionSourceData {
  source: string;
  value: number;
  color: string;
}

interface HeatMapData {
  hour: number;
  day: string;
  value: number;
}

export function EmissionsTrendChart({ data }: { data: EmissionTrendData[] }) {
  const chartRef = useRef<HTMLDivElement>(null);
  const { colorMode } = useColorMode();
  const isDark = colorMode === 'dark';

  const exportAsPNG = async () => {
    if (chartRef.current) {
      const canvas = await html2canvas(chartRef.current);
      const link = document.createElement('a');
      link.download = 'emissions-trend.png';
      link.href = canvas.toDataURL();
      link.click();
    }
  };

  const exportAsPDF = async () => {
    if (chartRef.current) {
      const canvas = await html2canvas(chartRef.current);
      const imgData = canvas.toDataURL('image/png');
      const pdf = new jsPDF('landscape');
      const imgWidth = 280;
      const imgHeight = (canvas.height * imgWidth) / canvas.width;
      pdf.addImage(imgData, 'PNG', 10, 10, imgWidth, imgHeight);
      pdf.save('emissions-trend.pdf');
    }
  };

  return (
    <Box ref={chartRef} p={4} bg={isDark ? 'gray.800' : 'white'} borderRadius="lg" boxShadow="md">
      <ButtonGroup size="sm" mb={4}>
        <Button onClick={exportAsPNG}>Export PNG</Button>
        <Button onClick={exportAsPDF}>Export PDF</Button>
      </ButtonGroup>
      <ResponsiveContainer width="100%" height={400}>
        <LineChart data={data}>
          <CartesianGrid strokeDasharray="3 3" stroke={isDark ? '#374151' : '#e5e7eb'} />
          <XAxis
            dataKey="date"
            stroke={isDark ? '#9ca3af' : '#6b7280'}
            style={{ fontSize: '12px' }}
          />
          <YAxis
            stroke={isDark ? '#9ca3af' : '#6b7280'}
            style={{ fontSize: '12px' }}
            label={{ value: 'tCO2e', angle: -90, position: 'insideLeft' }}
          />
          <Tooltip
            contentStyle={{
              backgroundColor: isDark ? '#1f2937' : '#ffffff',
              border: `1px solid ${isDark ? '#374151' : '#e5e7eb'}`,
              borderRadius: '8px',
            }}
            labelStyle={{ color: isDark ? '#f3f4f6' : '#111827' }}
          />
          <Legend />
          <Line
            type="monotone"
            dataKey="scope1"
            stroke="#ef4444"
            strokeWidth={2}
            name="Scope 1"
            dot={{ r: 4 }}
            activeDot={{ r: 6 }}
          />
          <Line
            type="monotone"
            dataKey="scope2"
            stroke="#0ea5e9"
            strokeWidth={2}
            name="Scope 2"
            dot={{ r: 4 }}
            activeDot={{ r: 6 }}
          />
          <Line
            type="monotone"
            dataKey="scope3"
            stroke="#f59e0b"
            strokeWidth={2}
            name="Scope 3"
            dot={{ r: 4 }}
            activeDot={{ r: 6 }}
          />
          <Line
            type="monotone"
            dataKey="total"
            stroke="#059669"
            strokeWidth={3}
            name="Total"
            dot={{ r: 5 }}
            activeDot={{ r: 7 }}
          />
        </LineChart>
      </ResponsiveContainer>
    </Box>
  );
}

export function ScopeBreakdownChart({ data }: { data: ScopeBreakdownData[] }) {
  const chartRef = useRef<HTMLDivElement>(null);
  const { colorMode } = useColorMode();
  const isDark = colorMode === 'dark';

  const exportAsPNG = async () => {
    if (chartRef.current) {
      const canvas = await html2canvas(chartRef.current);
      const link = document.createElement('a');
      link.download = 'scope-breakdown.png';
      link.href = canvas.toDataURL();
      link.click();
    }
  };

  const exportAsPDF = async () => {
    if (chartRef.current) {
      const canvas = await html2canvas(chartRef.current);
      const imgData = canvas.toDataURL('image/png');
      const pdf = new jsPDF();
      const imgWidth = 190;
      const imgHeight = (canvas.height * imgWidth) / canvas.width;
      pdf.addImage(imgData, 'PNG', 10, 10, imgWidth, imgHeight);
      pdf.save('scope-breakdown.pdf');
    }
  };

  return (
    <Box ref={chartRef} p={4} bg={isDark ? 'gray.800' : 'white'} borderRadius="lg" boxShadow="md">
      <ButtonGroup size="sm" mb={4}>
        <Button onClick={exportAsPNG}>Export PNG</Button>
        <Button onClick={exportAsPDF}>Export PDF</Button>
      </ButtonGroup>
      <ResponsiveContainer width="100%" height={400}>
        <BarChart data={data}>
          <CartesianGrid strokeDasharray="3 3" stroke={isDark ? '#374151' : '#e5e7eb'} />
          <XAxis
            dataKey="scope"
            stroke={isDark ? '#9ca3af' : '#6b7280'}
            style={{ fontSize: '12px' }}
          />
          <YAxis
            stroke={isDark ? '#9ca3af' : '#6b7280'}
            style={{ fontSize: '12px' }}
            label={{ value: 'tCO2e', angle: -90, position: 'insideLeft' }}
          />
          <Tooltip
            contentStyle={{
              backgroundColor: isDark ? '#1f2937' : '#ffffff',
              border: `1px solid ${isDark ? '#374151' : '#e5e7eb'}`,
              borderRadius: '8px',
            }}
            labelStyle={{ color: isDark ? '#f3f4f6' : '#111827' }}
          />
          <Legend />
          <Bar dataKey="emissions" name="Emissions (tCO2e)">
            {data.map((entry, index) => (
              <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
            ))}
          </Bar>
        </BarChart>
      </ResponsiveContainer>
    </Box>
  );
}

export function EmissionSourcesPieChart({ data }: { data: EmissionSourceData[] }) {
  const chartRef = useRef<HTMLDivElement>(null);
  const { colorMode } = useColorMode();
  const isDark = colorMode === 'dark';

  const exportAsPNG = async () => {
    if (chartRef.current) {
      const canvas = await html2canvas(chartRef.current);
      const link = document.createElement('a');
      link.download = 'emission-sources.png';
      link.href = canvas.toDataURL();
      link.click();
    }
  };

  const exportAsPDF = async () => {
    if (chartRef.current) {
      const canvas = await html2canvas(chartRef.current);
      const imgData = canvas.toDataURL('image/png');
      const pdf = new jsPDF();
      const imgWidth = 190;
      const imgHeight = (canvas.height * imgWidth) / canvas.width;
      pdf.addImage(imgData, 'PNG', 10, 10, imgWidth, imgHeight);
      pdf.save('emission-sources.pdf');
    }
  };

  const CustomTooltip = ({ active, payload }: TooltipProps<number, string>) => {
    if (active && payload && payload.length) {
      return (
        <Box
          bg={isDark ? 'gray.700' : 'white'}
          p={3}
          borderRadius="md"
          boxShadow="lg"
          border="1px"
          borderColor={isDark ? 'gray.600' : 'gray.200'}
        >
          <p style={{ margin: 0, fontWeight: 'bold' }}>{payload[0].name}</p>
          <p style={{ margin: 0, color: payload[0].payload.color }}>
            {payload[0].value} tCO2e ({((payload[0].value as number / data.reduce((sum, d) => sum + d.value, 0)) * 100).toFixed(1)}%)
          </p>
        </Box>
      );
    }
    return null;
  };

  return (
    <Box ref={chartRef} p={4} bg={isDark ? 'gray.800' : 'white'} borderRadius="lg" boxShadow="md">
      <ButtonGroup size="sm" mb={4}>
        <Button onClick={exportAsPNG}>Export PNG</Button>
        <Button onClick={exportAsPDF}>Export PDF</Button>
      </ButtonGroup>
      <ResponsiveContainer width="100%" height={400}>
        <PieChart>
          <Pie
            data={data}
            cx="50%"
            cy="50%"
            labelLine={false}
            label={({ name, percent }) => `${name}: ${(percent * 100).toFixed(1)}%`}
            outerRadius={120}
            fill="#8884d8"
            dataKey="value"
          >
            {data.map((entry, index) => (
              <Cell key={`cell-${index}`} fill={entry.color || COLORS[index % COLORS.length]} />
            ))}
          </Pie>
          <Tooltip content={<CustomTooltip />} />
        </PieChart>
      </ResponsiveContainer>
    </Box>
  );
}

export function TemporalHeatMap({ data }: { data: HeatMapData[] }) {
  const chartRef = useRef<HTMLDivElement>(null);
  const { colorMode } = useColorMode();
  const isDark = colorMode === 'dark';

  // Group data by day
  const days = Array.from(new Set(data.map((d) => d.day)));
  const hours = Array.from(new Set(data.map((d) => d.hour))).sort((a, b) => a - b);

  const maxValue = Math.max(...data.map((d) => d.value));

  const getColor = (value: number) => {
    const intensity = value / maxValue;
    if (intensity > 0.75) return '#dc2626';
    if (intensity > 0.5) return '#f59e0b';
    if (intensity > 0.25) return '#fbbf24';
    return '#059669';
  };

  const exportAsPNG = async () => {
    if (chartRef.current) {
      const canvas = await html2canvas(chartRef.current);
      const link = document.createElement('a');
      link.download = 'temporal-heatmap.png';
      link.href = canvas.toDataURL();
      link.click();
    }
  };

  const exportAsPDF = async () => {
    if (chartRef.current) {
      const canvas = await html2canvas(chartRef.current);
      const imgData = canvas.toDataURL('image/png');
      const pdf = new jsPDF('landscape');
      const imgWidth = 280;
      const imgHeight = (canvas.height * imgWidth) / canvas.width;
      pdf.addImage(imgData, 'PNG', 10, 10, imgWidth, imgHeight);
      pdf.save('temporal-heatmap.pdf');
    }
  };

  return (
    <Box ref={chartRef} p={4} bg={isDark ? 'gray.800' : 'white'} borderRadius="lg" boxShadow="md">
      <ButtonGroup size="sm" mb={4}>
        <Button onClick={exportAsPNG}>Export PNG</Button>
        <Button onClick={exportAsPDF}>Export PDF</Button>
      </ButtonGroup>
      <Box overflowX="auto">
        <Box display="inline-block" minW="100%">
          <Box display="grid" gridTemplateColumns={`80px repeat(${hours.length}, 40px)`} gap={1}>
            <Box />
            {hours.map((hour) => (
              <Box key={hour} textAlign="center" fontSize="xs" color={isDark ? 'gray.400' : 'gray.600'}>
                {hour}:00
              </Box>
            ))}
            {days.map((day) => (
              <>
                <Box key={`${day}-label`} fontSize="sm" fontWeight="medium" display="flex" alignItems="center" color={isDark ? 'gray.300' : 'gray.700'}>
                  {day}
                </Box>
                {hours.map((hour) => {
                  const dataPoint = data.find((d) => d.day === day && d.hour === hour);
                  const value = dataPoint?.value || 0;
                  return (
                    <Box
                      key={`${day}-${hour}`}
                      h="40px"
                      bg={getColor(value)}
                      borderRadius="sm"
                      display="flex"
                      alignItems="center"
                      justifyContent="center"
                      fontSize="xs"
                      color="white"
                      fontWeight="bold"
                      cursor="pointer"
                      title={`${day} ${hour}:00 - ${value.toFixed(2)} tCO2e`}
                      _hover={{ transform: 'scale(1.1)', transition: 'transform 0.2s' }}
                    >
                      {value > 0 ? value.toFixed(0) : ''}
                    </Box>
                  );
                })}
              </>
            ))}
          </Box>
        </Box>
      </Box>
    </Box>
  );
}
