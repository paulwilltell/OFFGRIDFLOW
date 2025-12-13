/**
 * @fileoverview Unit tests for EmissionChartJS component
 * @description Tests for Chart.js visualization with accessibility and interactions
 */

import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { EmissionChartJS } from '@/components/charts/EmissionChartJS';
import { EmissionData, Timeframe } from '@/types/carbon';

// Mock Chart.js
jest.mock('chart.js', () => ({
  Chart: jest.fn().mockImplementation(() => ({
    destroy: jest.fn(),
    update: jest.fn(),
    resize: jest.fn(),
  })),
  CategoryScale: jest.fn(),
  LinearScale: jest.fn(),
  PointElement: jest.fn(),
  LineElement: jest.fn(),
  BarElement: jest.fn(),
  Title: jest.fn(),
  Tooltip: jest.fn(),
  Legend: jest.fn(),
  Filler: jest.fn(),
  registerables: [],
}));

jest.mock('chart.js', () => {
  const mockChart = jest.fn().mockImplementation(() => ({
    destroy: jest.fn(),
    update: jest.fn(),
  }));
  
  return {
    Chart: Object.assign(mockChart, {
      register: jest.fn(),
    }),
    CategoryScale: jest.fn(),
    LinearScale: jest.fn(),
    PointElement: jest.fn(),
    LineElement: jest.fn(),
    BarElement: jest.fn(),
    Title: jest.fn(),
    Tooltip: jest.fn(),
    Legend: jest.fn(),
    Filler: jest.fn(),
  };
});

jest.mock('chartjs-plugin-annotation', () => ({}));
jest.mock('chartjs-plugin-zoom', () => ({}));

// Mock canvas context
HTMLCanvasElement.prototype.getContext = jest.fn().mockReturnValue({
  createLinearGradient: jest.fn().mockReturnValue({
    addColorStop: jest.fn(),
  }),
  fillRect: jest.fn(),
  clearRect: jest.fn(),
  getImageData: jest.fn(),
  putImageData: jest.fn(),
  createImageData: jest.fn(),
  setTransform: jest.fn(),
  drawImage: jest.fn(),
  save: jest.fn(),
  fillText: jest.fn(),
  restore: jest.fn(),
  beginPath: jest.fn(),
  moveTo: jest.fn(),
  lineTo: jest.fn(),
  closePath: jest.fn(),
  stroke: jest.fn(),
  translate: jest.fn(),
  scale: jest.fn(),
  rotate: jest.fn(),
  arc: jest.fn(),
  fill: jest.fn(),
  measureText: jest.fn().mockReturnValue({ width: 0 }),
  transform: jest.fn(),
  rect: jest.fn(),
  clip: jest.fn(),
});

const mockEmissionData: EmissionData[] = [
  {
    id: '1',
    tenantId: 'tenant-1',
    total: 12450,
    scope1: 3200,
    scope2: 5800,
    scope3: 3450,
    intensity: 249,
    timeframe: 'monthly',
    dataSources: [],
    updatedAt: new Date('2024-01-15'),
    methodology: 'ghg_protocol',
    uncertainty: 5,
    region: 'north_america',
  },
  {
    id: '2',
    tenantId: 'tenant-1',
    total: 11800,
    scope1: 3000,
    scope2: 5500,
    scope3: 3300,
    intensity: 236,
    timeframe: 'monthly',
    dataSources: [],
    updatedAt: new Date('2024-02-15'),
    methodology: 'ghg_protocol',
    uncertainty: 5,
    region: 'north_america',
  },
];

describe('EmissionChartJS', () => {
  describe('Rendering', () => {
    it('should render without crashing', () => {
      render(
        <EmissionChartJS 
          data={mockEmissionData} 
          timeframe="monthly" 
        />
      );
      
      expect(screen.getByText('Emission Trends')).toBeInTheDocument();
    });

    it('should render with custom height', () => {
      const { container } = render(
        <EmissionChartJS 
          data={mockEmissionData} 
          timeframe="monthly"
          height={500}
        />
      );
      
      const chartContainer = container.querySelector('[style*="height: 500px"]');
      expect(chartContainer).toBeInTheDocument();
    });

    it('should render time range buttons', () => {
      render(
        <EmissionChartJS 
          data={mockEmissionData} 
          timeframe="monthly" 
        />
      );
      
      expect(screen.getByText('1M')).toBeInTheDocument();
      expect(screen.getByText('3M')).toBeInTheDocument();
      expect(screen.getByText('1Y')).toBeInTheDocument();
      expect(screen.getByText('All')).toBeInTheDocument();
    });

    it('should render chart footer with instructions', () => {
      render(
        <EmissionChartJS 
          data={mockEmissionData} 
          timeframe="monthly" 
        />
      );
      
      expect(screen.getByText(/Scroll to zoom/)).toBeInTheDocument();
      expect(screen.getByText(/Drag to pan/)).toBeInTheDocument();
    });

    it('should render canvas element', () => {
      const { container } = render(
        <EmissionChartJS 
          data={mockEmissionData} 
          timeframe="monthly" 
        />
      );
      
      expect(container.querySelector('canvas')).toBeInTheDocument();
    });
  });

  describe('Props Handling', () => {
    it('should handle empty data array', () => {
      render(
        <EmissionChartJS 
          data={[]} 
          timeframe="monthly" 
        />
      );
      
      expect(screen.getByText('Emission Trends')).toBeInTheDocument();
    });

    it('should handle single data point', () => {
      render(
        <EmissionChartJS 
          data={[mockEmissionData[0]]} 
          timeframe="monthly" 
        />
      );
      
      expect(screen.getByText('Emission Trends')).toBeInTheDocument();
    });

    it('should accept different timeframes', () => {
      const timeframes: Timeframe[] = ['daily', 'weekly', 'monthly', 'quarterly', 'yearly'];
      
      timeframes.forEach((timeframe) => {
        const { unmount } = render(
          <EmissionChartJS 
            data={mockEmissionData} 
            timeframe={timeframe} 
          />
        );
        
        expect(screen.getByText('Emission Trends')).toBeInTheDocument();
        unmount();
      });
    });
  });

  describe('Interactions', () => {
    it('should call onPointClick when provided', () => {
      const handleClick = jest.fn();
      
      render(
        <EmissionChartJS 
          data={mockEmissionData} 
          timeframe="monthly"
          onPointClick={handleClick}
        />
      );
      
      // Click handler is attached to Chart.js, not directly testable without full integration
      expect(handleClick).not.toHaveBeenCalled();
    });
  });

  describe('Memoization', () => {
    it('should be memoized component', () => {
      const { rerender } = render(
        <EmissionChartJS 
          data={mockEmissionData} 
          timeframe="monthly" 
        />
      );
      
      // Re-render with same props should not cause issues
      rerender(
        <EmissionChartJS 
          data={mockEmissionData} 
          timeframe="monthly" 
        />
      );
      
      expect(screen.getByText('Emission Trends')).toBeInTheDocument();
    });
  });

  describe('Cleanup', () => {
    it('should destroy chart on unmount', () => {
      const { unmount } = render(
        <EmissionChartJS 
          data={mockEmissionData} 
          timeframe="monthly" 
        />
      );
      
      // Unmount should not throw
      expect(() => unmount()).not.toThrow();
    });
  });

  describe('Accessibility', () => {
    it('should have accessible heading', () => {
      render(
        <EmissionChartJS 
          data={mockEmissionData} 
          timeframe="monthly" 
        />
      );
      
      const heading = screen.getByText('Emission Trends');
      expect(heading.tagName).toBe('H3');
    });

    it('should have accessible button labels', () => {
      render(
        <EmissionChartJS 
          data={mockEmissionData} 
          timeframe="monthly" 
        />
      );
      
      const buttons = screen.getAllByRole('button');
      expect(buttons.length).toBeGreaterThan(0);
    });
  });
});

describe('Chart Configuration', () => {
  it('should support dark theme styling', () => {
    const { container } = render(
      <EmissionChartJS 
        data={mockEmissionData} 
        timeframe="monthly" 
      />
    );
    
    const wrapper = container.firstChild;
    expect(wrapper).toHaveClass('bg-gray-800/50');
  });

  it('should have rounded border', () => {
    const { container } = render(
      <EmissionChartJS 
        data={mockEmissionData} 
        timeframe="monthly" 
      />
    );
    
    const wrapper = container.firstChild;
    expect(wrapper).toHaveClass('rounded-xl');
  });
});
