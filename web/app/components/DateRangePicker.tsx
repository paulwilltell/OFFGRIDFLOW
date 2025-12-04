'use client';

import DatePicker from 'react-datepicker';
import 'react-datepicker/dist/react-datepicker.css';
import { Box, FormLabel, useColorMode } from '@chakra-ui/react';
import { useState } from 'react';

interface DateRangePickerProps {
  startDate: Date | null;
  endDate: Date | null;
  onStartDateChange: (date: Date | null) => void;
  onEndDateChange: (date: Date | null) => void;
  label?: string;
}

export function DateRangePicker({
  startDate,
  endDate,
  onStartDateChange,
  onEndDateChange,
  label = 'Date Range',
}: DateRangePickerProps) {
  const { colorMode } = useColorMode();
  const isDark = colorMode === 'dark';

  return (
    <Box>
      {label && <FormLabel mb={2}>{label}</FormLabel>}
      <Box
        className={isDark ? 'dark-datepicker' : ''}
        display="flex"
        gap={4}
        flexWrap="wrap"
      >
        <Box flex={1} minW="200px">
          <DatePicker
            selected={startDate}
            onChange={onStartDateChange}
            selectsStart
            startDate={startDate}
            endDate={endDate}
            placeholderText="Start Date"
            dateFormat="MMM d, yyyy"
            className="custom-datepicker"
          />
        </Box>
        <Box flex={1} minW="200px">
          <DatePicker
            selected={endDate}
            onChange={onEndDateChange}
            selectsEnd
            startDate={startDate}
            endDate={endDate}
            minDate={startDate}
            placeholderText="End Date"
            dateFormat="MMM d, yyyy"
            className="custom-datepicker"
          />
        </Box>
      </Box>
      <style jsx global>{`
        .custom-datepicker {
          width: 100%;
          padding: 0.5rem;
          border: 1px solid ${isDark ? '#4a5568' : '#e2e8f0'};
          border-radius: 0.375rem;
          background-color: ${isDark ? '#2d3748' : '#ffffff'};
          color: ${isDark ? '#f7fafc' : '#1a202c'};
        }
        .custom-datepicker:focus {
          outline: none;
          border-color: #059669;
          box-shadow: 0 0 0 1px #059669;
        }
        .dark-datepicker .react-datepicker {
          background-color: #2d3748;
          border-color: #4a5568;
        }
        .dark-datepicker .react-datepicker__header {
          background-color: #1a202c;
          border-bottom-color: #4a5568;
        }
        .dark-datepicker .react-datepicker__current-month,
        .dark-datepicker .react-datepicker__day-name {
          color: #f7fafc;
        }
        .dark-datepicker .react-datepicker__day {
          color: #e2e8f0;
        }
        .dark-datepicker .react-datepicker__day:hover {
          background-color: #4a5568;
        }
        .dark-datepicker .react-datepicker__day--selected {
          background-color: #059669;
          color: white;
        }
      `}</style>
    </Box>
  );
}
