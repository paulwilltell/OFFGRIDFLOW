# 🚀 Quick Start Guide - OffGridFlow Frontend

## ⚡ 5-Minute Setup

### 1. Navigate to Project
```bash
cd C:\Users\pault\OffGridFlow\web
```

### 2. Install Dependencies
```bash
npm install
```

### 3. Start Development Server
```bash
npm run dev
```

### 4. Open Your Browser
Navigate to: **http://localhost:3000/showcase**

---

## 📦 What's Been Installed

### Core Framework
- ✅ Next.js 14
- ✅ React 18
- ✅ TypeScript

### UI Framework
- ✅ Chakra UI (Design System)
- ✅ Framer Motion (Animations)
- ✅ Emotion (CSS-in-JS)

### Charts & Visualization
- ✅ Recharts (Charts Library)
- ✅ html2canvas (Export)
- ✅ jsPDF (PDF Export)

### Data & Tables
- ✅ TanStack Table (Data Tables)
- ✅ React DatePicker
- ✅ React Dropzone (File Upload)

### Internationalization
- ✅ react-i18next
- ✅ i18next

### Notifications
- ✅ React Toastify

### Onboarding
- ✅ intro.js-react

---

## 🎯 What You Get

### ✅ 40+ Production-Ready Components
1. **Design System** - Complete theme with tokens
2. **Charts** - 4 chart types (Line, Bar, Pie, Heatmap)
3. **Loading States** - 6 skeleton variants
4. **Notifications** - Toast system
5. **Empty States** - 4 state variants
6. **Data Table** - Sortable, filterable, paginated
7. **Date Picker** - Range selection
8. **File Upload** - Drag & drop with progress
9. **Multi-Step Wizard** - Complex forms
10. **Notification Bell** - Real-time alerts
11. **Search** - Autocomplete
12. **Tree View** - Hierarchical data
13. **Dashboard Widgets** - 7 widget types
14. **Responsive Layout** - Mobile-first
15. **Theme Controls** - Dark/Light mode
16. **Onboarding** - Welcome tour & checklist
17. **Performance Utils** - Optimization tools

---

## 📖 Documentation Files

| File | Purpose |
|------|---------|
| `IMPLEMENTATION_SUMMARY.md` | Complete overview of all implementations |
| `FRONTEND_COMPONENTS_README.md` | Detailed component documentation |
| `QUICK_REFERENCE.md` | Quick usage examples |
| `COMPONENT_GALLERY.md` | Visual guide with examples |
| `QUICK_START.md` | This file - quick setup guide |

---

## 🎨 Your First Component

### 1. Import and Use a Component
```tsx
// app/page.tsx
import { KPICard } from '@/app/components/DashboardWidgets';

export default function HomePage() {
  return (
    <KPICard
      title="Total Emissions"
      value="45,234"
      unit="tCO2e"
      trend={-12.5}
      icon={TrendingDown}
      colorScheme="green"
    />
  );
}
```

### 2. Use a Chart
```tsx
import { EmissionsTrendChart } from '@/app/components/Charts';

const data = [
  { month: 'Jan', scope1: 1200, scope2: 800, scope3: 2000 },
  { month: 'Feb', scope1: 1300, scope2: 850, scope3: 2100 }
];

<EmissionsTrendChart data={data} />
```

### 3. Show a Toast Notification
```tsx
import { toast } from '@/app/components/Toast';

const handleSave = () => {
  toast.success('Data saved successfully!');
};
```

---

## 🏗️ Project Structure

```
web/
├── app/
│   ├── components/          # All components
│   │   ├── Charts.tsx      # Visualization components
│   │   ├── DashboardWidgets.tsx
│   │   ├── DataTable.tsx
│   │   ├── FileUpload.tsx
│   │   └── ... (15+ more)
│   ├── showcase/           # Component showcase
│   │   └── page.tsx
│   ├── layout.tsx          # Root layout
│   └── providers.tsx       # App providers
├── lib/
│   ├── theme.ts           # Chakra UI theme
│   └── i18n.ts           # Translations
├── public/               # Static assets
└── Documentation files
```

---

## 🎯 Common Tasks

### Task: Add Dark Mode Toggle
```tsx
import { ThemeControls } from '@/app/components/ThemeControls';

<ThemeControls />
```

### Task: Show Loading State
```tsx
import { CardSkeleton } from '@/app/components/LoadingSkeletons';

{loading ? <CardSkeleton /> : <YourComponent />}
```

### Task: Display Data Table
```tsx
import { DataTable } from '@/app/components/DataTable';

<DataTable
  data={yourData}
  columns={[
    { id: 'name', header: 'Name' },
    { id: 'value', header: 'Value' }
  ]}
/>
```

### Task: Upload Files
```tsx
import { FileUpload } from '@/app/components/FileUpload';

<FileUpload
  accept=".csv,.xlsx"
  onUpload={(files) => console.log(files)}
/>
```

### Task: Create Multi-Step Form
```tsx
import { MultiStepWizard } from '@/app/components/MultiStepWizard';

const steps = [
  { id: 'step1', title: 'Step 1', content: <Step1Form /> },
  { id: 'step2', title: 'Step 2', content: <Step2Form /> }
];

<MultiStepWizard steps={steps} onComplete={handleComplete} />
```

---

## 🌍 Internationalization

### Change Language
```tsx
import { useTranslation } from 'react-i18next';

const { t, i18n } = useTranslation();

// Change language
i18n.changeLanguage('es'); // Spanish
i18n.changeLanguage('de'); // German
i18n.changeLanguage('fr'); // French
i18n.changeLanguage('en'); // English
```

### Use Translations
```tsx
const { t } = useTranslation();

<Heading>{t('dashboard.title')}</Heading>
```

---

## 🎨 Theming

### Use Theme Colors
```tsx
import { useColorModeValue } from '@chakra-ui/react';

const bg = useColorModeValue('white', 'gray.800');
const color = useColorModeValue('gray.800', 'white');
```

### Custom Theme Colors
```tsx
// lib/theme.ts - already configured!
colors: {
  brand: {
    50: '#f0fdf4',
    500: '#059669', // Primary green
    900: '#064e3b'
  }
}
```

---

## 📊 Data Visualization Quick Guide

### Line Chart
```tsx
<EmissionsTrendChart data={monthlyData} />
```

### Bar Chart
```tsx
<ScopeBreakdownChart data={scopeData} />
```

### Pie Chart
```tsx
<EmissionSourcesPieChart data={sourceData} />
```

### Heat Map
```tsx
<TemporalHeatMap data={hourlyData} />
```

---

## 🚨 Error Handling

### Show Error State
```tsx
import { ErrorState } from '@/app/components/EmptyStates';

{error && (
  <ErrorState
    title="Failed to load data"
    description={error.message}
    onRetry={refetch}
  />
)}
```

### Toast Error
```tsx
import { toast } from '@/app/components/Toast';

try {
  await saveData();
} catch (error) {
  toast.error('Failed to save');
}
```

---

## ♿ Accessibility

### All Components Include:
- ✅ ARIA labels
- ✅ Keyboard navigation
- ✅ Screen reader support
- ✅ Color contrast compliance (WCAG 2.1 AA)
- ✅ Focus management

### Test Accessibility
```bash
# Use browser DevTools
# Check keyboard navigation
# Use screen reader (NVDA, JAWS, VoiceOver)
```

---

## 📱 Responsive Design

### Breakpoint Usage
```tsx
// Hide on mobile
<Box display={{ base: 'none', md: 'block' }}>
  Desktop content
</Box>

// Show on mobile only
<Box display={{ base: 'block', md: 'none' }}>
  Mobile content
</Box>

// Responsive grid
<SimpleGrid columns={{ base: 1, md: 2, lg: 4 }} spacing={4}>
  <Card />
  <Card />
  <Card />
  <Card />
</SimpleGrid>
```

---

## 🔧 Customization

### Modify Theme
Edit `lib/theme.ts`:
```typescript
export const theme = extendTheme({
  colors: {
    brand: {
      500: '#YOUR_COLOR' // Change primary color
    }
  }
});
```

### Add Language
Edit `lib/i18n.ts`:
```typescript
const resources = {
  en: { translation: {...} },
  es: { translation: {...} },
  // Add your language
  pt: { translation: {...} }
};
```

---

## 🎓 Learning Path

### Beginner
1. ✅ Start with `QUICK_REFERENCE.md`
2. ✅ View `/showcase` page
3. ✅ Copy examples from `COMPONENT_GALLERY.md`

### Intermediate
1. ✅ Read `FRONTEND_COMPONENTS_README.md`
2. ✅ Customize theme in `lib/theme.ts`
3. ✅ Add translations in `lib/i18n.ts`

### Advanced
1. ✅ Review `IMPLEMENTATION_SUMMARY.md`
2. ✅ Optimize with `PerformanceUtils.tsx`
3. ✅ Build custom components

---

## 🐛 Troubleshooting

### Issue: Components not styled
**Solution:** Ensure `DesignSystemProvider` wraps your app
```tsx
// app/layout.tsx
import { DesignSystemProvider } from '@/app/components/DesignSystemProvider';

<DesignSystemProvider>{children}</DesignSystemProvider>
```

### Issue: Dark mode not working
**Solution:** Add ColorModeScript to layout
```tsx
import { ColorModeScript } from '@chakra-ui/react';
import theme from '@/lib/theme';

<ColorModeScript initialColorMode={theme.config.initialColorMode} />
```

### Issue: Translations not loading
**Solution:** Check i18n initialization
```tsx
import '@/lib/i18n'; // Import in layout or _app
```

### Issue: Charts not responsive
**Solution:** Wrap in ResponsiveContainer
```tsx
<ResponsiveContainer width="100%" height={400}>
  <LineChart data={data}>
    {/* chart content */}
  </LineChart>
</ResponsiveContainer>
```

---

## 🎉 Next Steps

### 1. Explore Showcase
```
http://localhost:3000/showcase
```

### 2. Build Your First Page
```tsx
// app/dashboard/page.tsx
import { ResponsiveLayout } from '@/app/components/ResponsiveLayout';
import { KPICard } from '@/app/components/DashboardWidgets';

export default function Dashboard() {
  return (
    <ResponsiveLayout>
      <KPICard
        title="Total Emissions"
        value="45,234"
        unit="tCO2e"
        trend={-12.5}
      />
    </ResponsiveLayout>
  );
}
```

### 3. Connect to Backend
Replace mock data with API calls:
```tsx
const { data, loading, error } = useSWR('/api/emissions', fetcher);
```

### 4. Add Authentication
Protect routes with NextAuth or similar:
```tsx
import { useSession } from 'next-auth/react';

const { data: session } = useSession();
if (!session) return <Login />;
```

---

## 📚 Resources

### Documentation
- **Component Guide**: `FRONTEND_COMPONENTS_README.md`
- **Quick Reference**: `QUICK_REFERENCE.md`
- **Visual Gallery**: `COMPONENT_GALLERY.md`
- **Implementation Summary**: `IMPLEMENTATION_SUMMARY.md`

### External Resources
- [Chakra UI Docs](https://chakra-ui.com)
- [Recharts Docs](https://recharts.org)
- [Next.js Docs](https://nextjs.org/docs)
- [TanStack Table](https://tanstack.com/table)

---

## ✅ Verification Checklist

Before deploying, verify:
- [ ] All components render correctly
- [ ] Dark mode works in all components
- [ ] Responsive design works on mobile
- [ ] Keyboard navigation works
- [ ] Translations load correctly
- [ ] Charts export as PNG/PDF
- [ ] Forms validate properly
- [ ] Loading states display
- [ ] Error states handle failures
- [ ] Toasts show notifications

---

## 🎊 You're All Set!

You now have a production-grade component library for your OffGridFlow carbon accounting platform!

### Quick Links
- 📖 [Full Documentation](./FRONTEND_COMPONENTS_README.md)
- 🎨 [Component Gallery](./COMPONENT_GALLERY.md)
- ⚡ [Quick Reference](./QUICK_REFERENCE.md)
- 📊 [Implementation Summary](./IMPLEMENTATION_SUMMARY.md)
- 🎯 [Live Showcase](http://localhost:3000/showcase)

### Support
- Check documentation files
- Review showcase examples
- Inspect component code
- Test in browser DevTools

**Happy Building! 🚀**
