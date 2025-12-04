# 🎨 Component Gallery - Visual Guide

## 📋 Table of Contents
1. [Design System](#design-system)
2. [Data Visualizations](#data-visualizations)
3. [UX Patterns](#ux-patterns)
4. [Advanced Components](#advanced-components)
5. [Dashboard Widgets](#dashboard-widgets)
6. [Layout Components](#layout-components)
7. [Accessibility Features](#accessibility-features)
8. [Onboarding Components](#onboarding-components)

---

## 🎨 Design System

### Theme Configuration
```typescript
// Import theme
import theme from '@/lib/theme';

// Brand Colors
primary: #059669 (Green)
secondary: #0ea5e9 (Blue)  
warning: #d97706 (Amber)
danger: #dc2626 (Red)
success: #10b981 (Emerald)
```

### Typography Scale
```
xs:   12px / 1rem
sm:   14px / 1.25rem
md:   16px / 1.5rem
lg:   18px / 1.75rem
xl:   20px / 1.75rem
2xl:  24px / 2rem
3xl:  30px / 2.25rem
4xl:  36px / 2.5rem
5xl:  48px / 1
6xl:  60px / 1
7xl:  72px / 1
8xl:  84px / 1
9xl:  96px / 1
```

### Spacing System
```
1:  0.25rem (4px)
2:  0.5rem (8px)
3:  0.75rem (12px)
4:  1rem (16px)
6:  1.5rem (24px)
8:  2rem (32px)
12: 3rem (48px)
16: 4rem (64px)
```

### Breakpoints
```
sm:  640px  (Mobile)
md:  768px  (Tablet)
lg:  1024px (Desktop)
xl:  1280px (Large Desktop)
2xl: 1536px (Extra Large)
```

---

## 📊 Data Visualizations

### 1. Emissions Trend Chart (Line Chart)
```tsx
<EmissionsTrendChart
  data={[
    { month: 'Jan', scope1: 1200, scope2: 800, scope3: 2000 },
    { month: 'Feb', scope1: 1300, scope2: 850, scope3: 2100 }
  ]}
/>
```
**Features:**
- Multi-line support (Scope 1, 2, 3)
- Interactive tooltips
- Responsive container
- Dark mode compatible
- Export to PNG/PDF

### 2. Scope Breakdown Chart (Bar Chart)
```tsx
<ScopeBreakdownChart
  data={[
    { category: 'Energy', scope1: 500, scope2: 300, scope3: 800 },
    { category: 'Transport', scope1: 200, scope2: 100, scope3: 600 }
  ]}
/>
```
**Features:**
- Stacked bars
- Color-coded by scope
- Responsive
- Tooltip with breakdown

### 3. Emission Sources Pie Chart
```tsx
<EmissionSourcesPieChart
  data={[
    { name: 'Electricity', value: 4000 },
    { name: 'Natural Gas', value: 3000 },
    { name: 'Transportation', value: 2000 }
  ]}
/>
```
**Features:**
- Interactive segments
- Percentage labels
- Color-coded
- Export capability

### 4. Temporal Heat Map
```tsx
<TemporalHeatMap
  data={Array.from({ length: 24 }, (_, hour) => ({
    hour,
    Monday: Math.random() * 100,
    Tuesday: Math.random() * 100,
    // ... other days
  }))}
/>
```
**Features:**
- Hour x Day grid
- Color gradient (low to high)
- Tooltips with values
- Responsive layout

---

## 🎯 UX Patterns

### 1. Loading Skeletons
```tsx
// Card Skeleton
<CardSkeleton />

// Table Skeleton
<TableSkeleton rows={5} columns={4} />

// Chart Skeleton
<ChartSkeleton />

// List Skeleton
<ListSkeleton items={3} />

// Form Skeleton
<FormSkeleton fields={4} />

// Dashboard Skeleton
<DashboardSkeleton />
```

### 2. Toast Notifications
```tsx
import { toast } from '@/app/components/Toast';

// Success
toast.success('Data saved successfully!');

// Error
toast.error('Failed to save data');

// Warning
toast.warning('This action cannot be undone');

// Info
toast.info('New updates available');

// Promise
toast.promise(
  saveData(),
  {
    pending: 'Saving...',
    success: 'Saved!',
    error: 'Failed!'
  }
);
```

### 3. Empty States
```tsx
// No Data
<EmptyState
  title="No emissions data"
  description="Start by adding your first data source"
  actionLabel="Add Data Source"
  onAction={() => {}}
/>

// No Results
<NoResults
  title="No results found"
  description="Try adjusting your filters"
  onClear={() => {}}
/>

// Error State
<ErrorState
  title="Failed to load data"
  description="Please try again"
  onRetry={() => {}}
/>

// Not Found (404)
<NotFound
  title="Page not found"
  description="The page you're looking for doesn't exist"
/>
```

### 4. Confirmation Dialog
```tsx
const { isOpen, onOpen, onClose } = useDisclosure();
const [isDeleting, setIsDeleting] = useState(false);

<ConfirmationDialog
  isOpen={isOpen}
  onClose={onClose}
  title="Delete Emission Record"
  message="Are you sure? This action cannot be undone."
  confirmLabel="Delete"
  onConfirm={async () => {
    setIsDeleting(true);
    await deleteRecord();
    setIsDeleting(false);
  }}
  isLoading={isDeleting}
  isDanger
/>
```

---

## 🔧 Advanced Components

### 1. Data Table
```tsx
<DataTable
  data={[
    { id: 1, name: 'Record 1', scope: 'Scope 1', emissions: 1200 },
    { id: 2, name: 'Record 2', scope: 'Scope 2', emissions: 800 }
  ]}
  columns={[
    { id: 'name', header: 'Name' },
    { id: 'scope', header: 'Scope' },
    { id: 'emissions', header: 'Emissions (tCO2e)' }
  ]}
  onRowClick={(row) => console.log(row)}
/>
```
**Features:**
- Sorting (ascending/descending)
- Filtering (global search)
- Pagination
- Column visibility toggle
- Row selection
- Keyboard navigation

### 2. Date Range Picker
```tsx
const [startDate, setStartDate] = useState<Date | null>(null);
const [endDate, setEndDate] = useState<Date | null>(null);

<DateRangePicker
  startDate={startDate}
  endDate={endDate}
  onChange={(start, end) => {
    setStartDate(start);
    setEndDate(end);
  }}
/>
```

### 3. File Upload
```tsx
<FileUpload
  accept=".csv,.xlsx"
  maxSize={10 * 1024 * 1024} // 10MB
  onUpload={(files) => {
    files.forEach(file => {
      console.log('Uploading:', file.name);
    });
  }}
/>
```
**Features:**
- Drag & drop
- Progress bar
- File preview
- Validation (size, type)
- Multiple files

### 4. Multi-Step Wizard
```tsx
const steps = [
  {
    id: 'info',
    title: 'Basic Information',
    content: <InfoForm />
  },
  {
    id: 'data',
    title: 'Emissions Data',
    content: <DataForm />
  },
  {
    id: 'review',
    title: 'Review',
    content: <ReviewForm />
  }
];

<MultiStepWizard
  steps={steps}
  onComplete={(data) => console.log('Completed:', data)}
/>
```

### 5. Notification Bell
```tsx
const notifications = [
  {
    id: '1',
    title: 'New compliance deadline',
    message: 'CDP report due in 30 days',
    timestamp: new Date(),
    read: false,
    type: 'warning'
  }
];

<NotificationBell
  notifications={notifications}
  onNotificationClick={(notif) => {}}
  onMarkAllRead={() => {}}
/>
```

### 6. Search with Autocomplete
```tsx
const suggestions = [
  'Electricity consumption',
  'Natural gas usage',
  'Transportation emissions'
];

<SearchWithAutocomplete
  suggestions={suggestions}
  onSearch={(query) => console.log('Search:', query)}
  placeholder="Search emissions sources..."
/>
```

### 7. Tree View
```tsx
const treeData = [
  {
    id: '1',
    label: 'Company',
    children: [
      {
        id: '1-1',
        label: 'Headquarters',
        children: [
          { id: '1-1-1', label: 'Building A' },
          { id: '1-1-2', label: 'Building B' }
        ]
      },
      { id: '1-2', label: 'Factory' }
    ]
  }
];

<TreeView
  data={treeData}
  onNodeSelect={(node) => console.log('Selected:', node)}
/>
```

---

## 📊 Dashboard Widgets

### 1. KPI Card
```tsx
<KPICard
  title="Total Emissions"
  value="45,234"
  unit="tCO2e"
  trend={-12.5}
  icon={TrendingDown}
  colorScheme="green"
/>
```

### 2. Executive Summary Widget
```tsx
<ExecutiveSummaryWidget
  totalEmissions={45234}
  reductionTarget={30}
  currentReduction={12.5}
  complianceStatus="compliant"
/>
```

### 3. Recent Activity Feed
```tsx
<RecentActivityFeed
  activities={[
    {
      id: '1',
      type: 'upload',
      description: 'Uploaded Q1 emissions data',
      user: 'John Doe',
      timestamp: new Date()
    }
  ]}
/>
```

### 4. Compliance Deadlines Widget
```tsx
<ComplianceDeadlinesWidget
  deadlines={[
    {
      id: '1',
      name: 'CDP Climate Report',
      dueDate: new Date('2024-12-31'),
      status: 'pending',
      priority: 'high'
    }
  ]}
/>
```

### 5. Data Source Health Widget
```tsx
<DataSourceHealthWidget
  sources={[
    { name: 'Energy Meter API', status: 'healthy', lastSync: new Date() },
    { name: 'Fleet Management', status: 'warning', lastSync: new Date() }
  ]}
/>
```

### 6. Carbon Reduction Targets Widget
```tsx
<CarbonReductionTargetsWidget
  targets={[
    {
      name: '2030 Net Zero',
      targetReduction: 100,
      currentReduction: 35,
      deadline: new Date('2030-12-31')
    }
  ]}
/>
```

### 7. Quick Actions Widget
```tsx
<QuickActionsWidget
  actions={[
    { label: 'Add Emissions Data', icon: Plus, onClick: () => {} },
    { label: 'Generate Report', icon: FileText, onClick: () => {} }
  ]}
/>
```

---

## 🏗️ Layout Components

### Responsive Layout
```tsx
<ResponsiveLayout>
  <Box p={4}>
    <Heading>Dashboard</Heading>
    {/* Your content */}
  </Box>
</ResponsiveLayout>
```

**Features:**
- Collapsible sidebar (desktop)
- Mobile drawer menu
- Responsive header
- Breadcrumbs
- User menu

---

## ♿ Accessibility Features

### Theme Controls
```tsx
<ThemeControls />
```

**Features:**
- Dark/Light mode toggle
- Language selector (EN, ES, DE, FR)
- System preference detection
- Persistent settings
- Keyboard accessible

### ARIA Labels
All components include:
- `aria-label` for icons
- `aria-describedby` for help text
- `role` attributes
- `tabindex` for navigation

### Keyboard Navigation
- `Tab` - Navigate forward
- `Shift+Tab` - Navigate backward
- `Enter/Space` - Activate
- `Escape` - Close dialogs
- Arrow keys - Navigate lists/trees

---

## 🎓 Onboarding Components

### Welcome Tour
```tsx
const [showTour, setShowTour] = useState(true);

<OnboardingTour
  isOpen={showTour}
  onComplete={() => setShowTour(false)}
/>
```

### Welcome Modal
```tsx
<WelcomeModal
  isOpen={isFirstVisit}
  onClose={() => setIsFirstVisit(false)}
/>
```

### Setup Checklist
```tsx
const [tasks, setTasks] = useState([
  { id: '1', label: 'Connect data source', completed: false },
  { id: '2', label: 'Add first emission record', completed: false },
  { id: '3', label: 'Set reduction targets', completed: false }
]);

<SetupChecklist
  tasks={tasks}
  onTaskComplete={(taskId) => {
    setTasks(tasks.map(t =>
      t.id === taskId ? { ...t, completed: true } : t
    ));
  }}
/>
```

---

## 🎨 Component Composition Examples

### Complete Dashboard Page
```tsx
import { ResponsiveLayout } from '@/app/components/ResponsiveLayout';
import {
  KPICard,
  ExecutiveSummaryWidget,
  RecentActivityFeed,
  ComplianceDeadlinesWidget,
  QuickActionsWidget
} from '@/app/components/DashboardWidgets';

export default function DashboardPage() {
  return (
    <ResponsiveLayout>
      <SimpleGrid columns={{ base: 1, md: 2, lg: 4 }} spacing={4}>
        <KPICard title="Total Emissions" value="45,234" unit="tCO2e" trend={-12.5} />
        <KPICard title="Scope 1" value="12,456" unit="tCO2e" trend={-5.2} />
        <KPICard title="Scope 2" value="18,234" unit="tCO2e" trend={-15.3} />
        <KPICard title="Scope 3" value="14,544" unit="tCO2e" trend={-8.1} />
      </SimpleGrid>
      
      <Grid templateColumns={{ base: '1fr', lg: '2fr 1fr' }} gap={4} mt={4}>
        <ExecutiveSummaryWidget />
        <QuickActionsWidget />
      </Grid>
      
      <Grid templateColumns={{ base: '1fr', lg: '1fr 1fr' }} gap={4} mt={4}>
        <RecentActivityFeed />
        <ComplianceDeadlinesWidget />
      </Grid>
    </ResponsiveLayout>
  );
}
```

### Data Entry Form with Validation
```tsx
import { MultiStepWizard } from '@/app/components/MultiStepWizard';
import { FileUpload } from '@/app/components/FileUpload';
import { DateRangePicker } from '@/app/components/DateRangePicker';
import { toast } from '@/app/components/Toast';

const steps = [
  {
    id: 'upload',
    title: 'Upload Data',
    content: (
      <FileUpload
        accept=".csv,.xlsx"
        onUpload={(files) => {
          toast.success(`Uploaded ${files.length} file(s)`);
        }}
      />
    )
  },
  {
    id: 'period',
    title: 'Select Period',
    content: (
      <DateRangePicker
        startDate={startDate}
        endDate={endDate}
        onChange={(start, end) => {
          setStartDate(start);
          setEndDate(end);
        }}
      />
    )
  }
];

<MultiStepWizard steps={steps} onComplete={handleSubmit} />
```

### Search & Filter Page
```tsx
import { SearchWithAutocomplete } from '@/app/components/SearchWithAutocomplete';
import { DataTable } from '@/app/components/DataTable';
import { EmptyState } from '@/app/components/EmptyStates';

const [searchQuery, setSearchQuery] = useState('');
const [data, setData] = useState([]);

return (
  <Box>
    <SearchWithAutocomplete
      onSearch={setSearchQuery}
      suggestions={['Electricity', 'Gas', 'Transport']}
    />
    
    {data.length === 0 ? (
      <EmptyState
        title="No results found"
        description="Try adjusting your search"
      />
    ) : (
      <DataTable data={data} columns={columns} />
    )}
  </Box>
);
```

---

## 🎯 Best Practices

### 1. Always Wrap with Provider
```tsx
// app/layout.tsx
import { DesignSystemProvider } from '@/app/components/DesignSystemProvider';

export default function RootLayout({ children }) {
  return (
    <html>
      <body>
        <DesignSystemProvider>
          {children}
        </DesignSystemProvider>
      </body>
    </html>
  );
}
```

### 2. Use Loading States
```tsx
import { CardSkeleton } from '@/app/components/LoadingSkeletons';

const [loading, setLoading] = useState(true);

return loading ? <CardSkeleton /> : <DataCard data={data} />;
```

### 3. Handle Errors Gracefully
```tsx
import { ErrorState } from '@/app/components/EmptyStates';

{error && (
  <ErrorState
    title="Failed to load"
    description={error.message}
    onRetry={refetch}
  />
)}
```

### 4. Provide User Feedback
```tsx
import { toast } from '@/app/components/Toast';

const handleSave = async () => {
  try {
    await saveData();
    toast.success('Data saved successfully!');
  } catch (error) {
    toast.error('Failed to save data');
  }
};
```

### 5. Confirm Destructive Actions
```tsx
import { ConfirmationDialog } from '@/app/components/ConfirmationDialog';

<ConfirmationDialog
  isOpen={isOpen}
  onClose={onClose}
  title="Delete Record"
  message="This cannot be undone"
  onConfirm={handleDelete}
  isDanger
/>
```

---

## 📱 Mobile-First Examples

### Responsive Grid
```tsx
<SimpleGrid columns={{ base: 1, md: 2, lg: 3, xl: 4 }} spacing={4}>
  <KPICard />
  <KPICard />
  <KPICard />
  <KPICard />
</SimpleGrid>
```

### Responsive Stack
```tsx
<Stack direction={{ base: 'column', md: 'row' }} spacing={4}>
  <Button>Action 1</Button>
  <Button>Action 2</Button>
</Stack>
```

### Hide on Mobile
```tsx
<Box display={{ base: 'none', md: 'block' }}>
  Desktop only content
</Box>
```

### Show on Mobile Only
```tsx
<Box display={{ base: 'block', md: 'none' }}>
  Mobile only content
</Box>
```

---

## 🌍 Internationalization Examples

### Using Translations
```tsx
import { useTranslation } from 'react-i18next';

function MyComponent() {
  const { t } = useTranslation();
  
  return (
    <Heading>{t('dashboard.title')}</Heading>
  );
}
```

### Language Switcher
```tsx
import { ThemeControls } from '@/app/components/ThemeControls';

<ThemeControls /> // Includes language selector
```

---

## 🎨 Theming Examples

### Using Theme Tokens
```tsx
import { useColorModeValue } from '@chakra-ui/react';

function ThemedCard() {
  const bg = useColorModeValue('white', 'gray.800');
  const color = useColorModeValue('gray.800', 'white');
  
  return (
    <Box bg={bg} color={color}>
      Content
    </Box>
  );
}
```

### Dark Mode Toggle
```tsx
import { ThemeControls } from '@/app/components/ThemeControls';

<ThemeControls /> // Includes dark mode toggle
```

---

## 🚀 Performance Tips

### 1. Lazy Load Components
```tsx
import dynamic from 'next/dynamic';

const Charts = dynamic(() => import('@/app/components/Charts'), {
  loading: () => <ChartSkeleton />
});
```

### 2. Virtual Scrolling for Large Lists
```tsx
import { useVirtualScroll } from '@/app/components/PerformanceUtils';

const { visibleItems, containerRef } = useVirtualScroll(largeDataset, 50);
```

### 3. Memoize Expensive Computations
```tsx
import { useMemo } from 'react';

const processedData = useMemo(
  () => expensiveOperation(data),
  [data]
);
```

---

## 🎉 Ready to Use!

All components are production-ready and fully documented. Start building your OffGridFlow platform with confidence!

### Quick Links
- [Full Documentation](./FRONTEND_COMPONENTS_README.md)
- [Quick Reference](./QUICK_REFERENCE.md)
- [Implementation Summary](./IMPLEMENTATION_SUMMARY.md)
- [Live Showcase](/showcase)
