# Quick Reference Guide - OffGridFlow Frontend Components

## 🚀 Quick Start

### 1. Install Dependencies (Already Done)
```bash
npm install
```

### 2. Start Development Server
```bash
cd C:\Users\pault\OffGridFlow\web
npm run dev
```

### 3. View Component Showcase
Navigate to: `http://localhost:3000/showcase`

## 📦 Component Import Guide

### Charts & Visualizations
```tsx
import {
  EmissionsTrendChart,
  ScopeBreakdownChart,
  EmissionSourcesPieChart,
  TemporalHeatMap
} from '@/app/components/Charts';
```

### Dashboard Widgets
```tsx
import {
  KPICard,
  ExecutiveSummaryWidget,
  RecentActivityFeed,
  ComplianceDeadlinesWidget,
  DataSourceHealthWidget,
  CarbonReductionTargetsWidget,
  QuickActionsWidget
} from '@/app/components/DashboardWidgets';
```

### UI Components
```tsx
import { DataTable } from '@/app/components/DataTable';
import { DateRangePicker } from '@/app/components/DateRangePicker';
import { FileUpload } from '@/app/components/FileUpload';
import { MultiStepWizard } from '@/app/components/MultiStepWizard';
import { NotificationBell } from '@/app/components/NotificationBell';
import { SearchWithAutocomplete } from '@/app/components/SearchWithAutocomplete';
import { TreeView } from '@/app/components/TreeView';
```

### UX Patterns
```tsx
import { toast } from '@/app/components/Toast';
import { EmptyState, NoDataEmptyState } from '@/app/components/EmptyStates';
import { ConfirmationDialog, DeleteConfirmationDialog } from '@/app/components/ConfirmationDialog';
import { CardSkeleton, TableSkeleton } from '@/app/components/LoadingSkeletons';
```

### Layout & Theme
```tsx
import { ResponsiveLayout } from '@/app/components/ResponsiveLayout';
import { ThemeAndLanguageControls } from '@/app/components/ThemeControls';
import { DesignSystemProvider } from '@/app/components/DesignSystemProvider';
```

### Onboarding
```tsx
import {
  WelcomeModal,
  OnboardingTour,
  SetupChecklist,
  useOnboarding
} from '@/app/components/Onboarding';
```

## 💡 Common Use Cases

### Display KPI Cards
```tsx
<SimpleGrid columns={{ base: 1, md: 2, lg: 4 }} spacing={4}>
  <KPICard
    title="Total Emissions"
    value="12,450 tCO2e"
    change={12}
    trend="down"
    subtitle="vs last month"
    icon="🌍"
  />
</SimpleGrid>
```

### Show Emissions Trend Chart
```tsx
const data = [
  { date: 'Jan', scope1: 1200, scope2: 2300, scope3: 3400, total: 6900 },
  // ... more data
];

<EmissionsTrendChart data={data} />
```

### Create Data Table
```tsx
const columns = [
  { accessorKey: 'date', header: 'Date' },
  { accessorKey: 'source', header: 'Source' },
  { accessorKey: 'amount', header: 'Amount' },
];

<DataTable
  data={records}
  columns={columns}
  enableSorting
  enableFiltering
  enablePagination
/>
```

### Show Notifications
```tsx
// Success
toast.success('Data saved successfully!');

// Error
toast.error('Failed to load data');

// Warning
toast.warning('Please review your input');

// Info
toast.info('New features available');

// Promise-based
await toast.promise(
  apiCall(),
  {
    pending: 'Saving...',
    success: 'Saved!',
    error: 'Failed to save'
  }
);
```

### File Upload
```tsx
<FileUpload
  onUpload={async (files) => {
    // Upload files to API
    await uploadToAPI(files);
  }}
  accept={{
    'text/csv': ['.csv'],
    'application/pdf': ['.pdf'],
  }}
  maxSize={10485760} // 10MB
/>
```

### Date Range Selection
```tsx
const [startDate, setStartDate] = useState<Date | null>(null);
const [endDate, setEndDate] = useState<Date | null>(null);

<DateRangePicker
  startDate={startDate}
  endDate={endDate}
  onStartDateChange={setStartDate}
  onEndDateChange={setEndDate}
  label="Select Period"
/>
```

### Search with Autocomplete
```tsx
<SearchWithAutocomplete
  onSearch={async (query) => {
    const results = await searchAPI(query);
    return results;
  }}
  onSelect={(result) => {
    console.log('Selected:', result);
  }}
  placeholder="Search emissions records..."
/>
```

### Confirmation Dialog
```tsx
const { isOpen, onOpen, onClose } = useDisclosure();

<Button onClick={onOpen}>Delete</Button>

<DeleteConfirmationDialog
  isOpen={isOpen}
  onClose={onClose}
  onConfirm={async () => {
    await deleteRecord();
    toast.success('Record deleted');
    onClose();
  }}
  itemName="Emission Record #123"
/>
```

### Multi-Step Wizard
```tsx
const steps = [
  {
    title: 'Basic Info',
    content: <BasicInfoForm />,
    validate: () => validateBasicInfo(),
  },
  {
    title: 'Details',
    content: <DetailsForm />,
    validate: () => validateDetails(),
  },
  {
    title: 'Review',
    content: <ReviewForm />,
  },
];

<MultiStepWizard
  steps={steps}
  onComplete={() => {
    toast.success('Wizard completed!');
  }}
  onCancel={() => {
    // Handle cancel
  }}
/>
```

### Responsive Layout
```tsx
<ResponsiveLayout>
  <YourPageContent />
</ResponsiveLayout>
```

### Internationalization
```tsx
import { useTranslation } from 'react-i18next';

function MyComponent() {
  const { t, i18n } = useTranslation();

  return (
    <Box>
      <Heading>{t('dashboard.title')}</Heading>
      <Button onClick={() => i18n.changeLanguage('es')}>
        Español
      </Button>
    </Box>
  );
}
```

### Loading States
```tsx
{isLoading ? (
  <CardSkeleton />
) : (
  <MyCard data={data} />
)}

{isTableLoading ? (
  <TableSkeleton rows={5} columns={4} />
) : (
  <DataTable data={data} columns={columns} />
)}
```

### Empty States
```tsx
{data.length === 0 ? (
  <NoDataEmptyState
    onRefresh={() => refetchData()}
  />
) : (
  <DataTable data={data} columns={columns} />
)}
```

### Tree View
```tsx
const orgStructure = [
  {
    id: '1',
    label: 'Company',
    children: [
      {
        id: '1-1',
        label: 'Department A',
        children: [
          { id: '1-1-1', label: 'Team 1' },
          { id: '1-1-2', label: 'Team 2' },
        ],
      },
    ],
  },
];

<TreeView
  data={orgStructure}
  selectable
  multiSelect
  defaultExpanded={['1']}
  onSelect={(node) => {
    console.log('Selected:', node.label);
  }}
/>
```

## 🎨 Theme Customization

### Access Theme Colors
```tsx
import { useColorMode, useTheme } from '@chakra-ui/react';

function MyComponent() {
  const { colorMode } = useColorMode();
  const theme = useTheme();
  
  const isDark = colorMode === 'dark';
  const brandColor = theme.colors.brand[500];
  
  return <Box bg={isDark ? 'gray.800' : 'white'}>...</Box>;
}
```

### Toggle Dark Mode
```tsx
import { useColorMode } from '@chakra-ui/react';

const { colorMode, toggleColorMode } = useColorMode();

<Button onClick={toggleColorMode}>
  {colorMode === 'light' ? '🌙' : '☀️'}
</Button>
```

## 📱 Responsive Design

### Responsive Props
```tsx
<SimpleGrid
  columns={{ base: 1, md: 2, lg: 3, xl: 4 }}
  spacing={{ base: 4, md: 6 }}
>
  {/* Cards */}
</SimpleGrid>

<Box
  display={{ base: 'block', md: 'flex' }}
  p={{ base: 4, md: 6, lg: 8 }}
>
  {/* Content */}
</Box>
```

### Breakpoint Values
- `base`: Mobile (< 640px)
- `sm`: 640px+
- `md`: 768px+
- `lg`: 1024px+
- `xl`: 1280px+
- `2xl`: 1536px+

## 🔧 Performance Tips

### Code Splitting
```tsx
import dynamic from 'next/dynamic';

const HeavyChart = dynamic(() => import('@/components/HeavyChart'), {
  loading: () => <ChartSkeleton />,
  ssr: false,
});
```

### Virtual Scrolling
```tsx
import { useVirtualScroll } from '@/app/components/PerformanceUtils';

const { visibleItems, totalHeight, offsetY, onScroll } = useVirtualScroll(
  items.length,
  itemHeight,
  containerHeight
);
```

### Image Optimization
```tsx
import { OptimizedImage } from '@/app/components/PerformanceUtils';

<OptimizedImage
  src="/chart.png"
  alt="Emissions chart"
  width={800}
  height={400}
  priority={false}
/>
```

## 🐛 Debugging

### Enable Console Logs in Dev
Console logs are automatically removed in production builds.

### Check Theme Values
```tsx
import { useTheme } from '@chakra-ui/react';

const theme = useTheme();
console.log('Theme:', theme);
```

### i18n Debugging
```tsx
import { useTranslation } from 'react-i18next';

const { t, i18n } = useTranslation();
console.log('Current language:', i18n.language);
console.log('Translation:', t('dashboard.title'));
```

## 📖 Additional Resources

- [Chakra UI Docs](https://chakra-ui.com/docs)
- [Recharts Docs](https://recharts.org/en-US/)
- [TanStack Table Docs](https://tanstack.com/table/v8)
- [react-i18next Docs](https://react.i18next.com/)
- [Next.js Docs](https://nextjs.org/docs)

## 🎯 Testing the Implementation

### 1. Start the development server:
```bash
cd C:\Users\pault\OffGridFlow\web
npm run dev
```

### 2. Visit the showcase page:
```
http://localhost:3000/showcase
```

### 3. Test all features:
- ✅ Dark mode toggle
- ✅ Language switching
- ✅ Charts with export
- ✅ Data table sorting/filtering
- ✅ File upload
- ✅ Toast notifications
- ✅ Date range picker
- ✅ Search autocomplete
- ✅ Tree view

### 4. Check responsive design:
- Open browser DevTools
- Toggle device toolbar
- Test on mobile, tablet, desktop sizes

### 5. Verify accessibility:
- Tab through all interactive elements
- Test with screen reader (NVDA, JAWS)
- Check color contrast ratios

## 🚨 Common Issues & Solutions

### Issue: Charts not rendering
**Solution:** Ensure data format matches expected interface

### Issue: Dark mode not working
**Solution:** Check ColorModeScript in root layout

### Issue: Translations not showing
**Solution:** Verify i18n initialization in providers

### Issue: Build errors
**Solution:** Run `npm install` to ensure all dependencies are installed

### Issue: Performance issues
**Solution:** Enable code splitting and lazy loading

## 📞 Need Help?

Check the comprehensive README at:
`web/FRONTEND_COMPONENTS_README.md`
