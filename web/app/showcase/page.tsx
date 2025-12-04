'use client';

import {
  Box,
  SimpleGrid,
  VStack,
  Heading,
  Text,
  Button,
  Container,
} from '@chakra-ui/react';
import { useState } from 'react';
import { EmissionsTrendChart, ScopeBreakdownChart } from '../components/Charts';
import { KPICard } from '../components/DashboardWidgets';
import { DataTable } from '../components/DataTable';
import { toast } from '../components/Toast';
import { ThemeAndLanguageControls } from '../components/ThemeControls';
import { ColumnDef } from '@tanstack/react-table';

const trendData = [
  { date: 'Jan', scope1: 1200, scope2: 2300, scope3: 3400, total: 6900 },
  { date: 'Feb', scope1: 1150, scope2: 2250, scope3: 3300, total: 6700 },
];

const scopeData = [
  { scope: 'Scope 1', emissions: 1000, percentage: 16 },
  { scope: 'Scope 2', emissions: 2100, percentage: 35 },
];

interface EmissionRecord {
  id: string;
  date: string;
  source: string;
}

const tableData: EmissionRecord[] = [
  { id: '1', date: '2024-12-01', source: 'Office Electricity' },
];

export default function ShowcasePage() {
  const columns: ColumnDef<EmissionRecord>[] = [
    { accessorKey: 'date', header: 'Date' },
    { accessorKey: 'source', header: 'Source' },
  ];

  return (
    <Container maxW="container.xl" py={8}>
      <VStack spacing={8} align="stretch">
        <Box>
          <Heading size="2xl" mb={2}>
            Component Showcase
          </Heading>
          <Text mb={4}>OffGridFlow Frontend Components</Text>
          <ThemeAndLanguageControls />
        </Box>

        <SimpleGrid columns={{ base: 1, md: 4 }} spacing={4}>
          <KPICard title="Emissions" value="6,100 tCO2e" icon="🌍" />
          <KPICard title="Sources" value="12" icon="🔌" color="blue" />
          <KPICard title="Compliance" value="85%" icon="📊" color="green" />
          <KPICard title="Progress" value="88%" icon="🎯" color="purple" />
        </SimpleGrid>

        <EmissionsTrendChart data={trendData} />
        <ScopeBreakdownChart data={scopeData} />
        <DataTable data={tableData} columns={columns} />

        <SimpleGrid columns={4} spacing={3}>
          <Button onClick={() => toast.success('Success!')}>Success</Button>
          <Button onClick={() => toast.error('Error!')}>Error</Button>
          <Button onClick={() => toast.warning('Warning!')}>Warning</Button>
          <Button onClick={() => toast.info('Info!')}>Info</Button>
        </SimpleGrid>
      </VStack>
    </Container>
  );
}
