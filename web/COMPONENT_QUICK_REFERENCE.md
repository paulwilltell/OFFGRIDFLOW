# Quick Reference - Component Usage

## Emissions Visualizations

### Trend Chart
```tsx
import { EmissionsTrendChart } from '@/components/emissions';

<EmissionsTrendChart period="year" height={400} />
```

### Scope Breakdown
```tsx
import { ScopeBreakdownChart } from '@/components/emissions';

<ScopeBreakdownChart 
  height={400} 
  startDate="2024-01-01" 
  endDate="2024-12-31" 
/>
```

### Heatmap
```tsx
import { EmissionsHeatmap } from '@/components/emissions';

<EmissionsHeatmap period="week" height={350} />
```

## Blockchain Components

### Wallet Connect
```tsx
import WalletConnect from '@/components/blockchain/WalletConnect';

<WalletConnect onConnect={(connected) => {
  console.log('Wallet connected:', connected);
}} />
```

### Portfolio
```tsx
import PortfolioOverview from '@/components/blockchain/PortfolioOverview';

<PortfolioOverview portfolio={portfolioData} loading={false} />
```

### Transactions
```tsx
import TransactionHistory from '@/components/blockchain/TransactionHistory';

<TransactionHistory walletConnected={true} />
```

## Whitelabel Theme

### Theme Provider (wrap entire app)
```tsx
import TenantThemeProvider from '@/components/whitelabel/TenantThemeProvider';

<TenantThemeProvider>
  <App />
</TenantThemeProvider>
```

### Logo
```tsx
import TenantLogo from '@/components/whitelabel/TenantLogo';

<TenantLogo width="200px" height="60px" fallback="Company Name" />
```

### Theme Customizer
```tsx
import ThemeCustomizer from '@/components/whitelabel/ThemeCustomizer';

<ThemeCustomizer onThemeChange={(theme) => {
  console.log('Theme changed:', theme);
}} />
```

## Error Boundary

```tsx
import ErrorBoundary from '@/components/ErrorBoundary';

<ErrorBoundary fallback={<CustomError />}>
  <YourComponent />
</ErrorBoundary>
```

## API Endpoints

All components expect these endpoints:

```
# Emissions
GET /api/emissions/trend?period={week|month|quarter|year}
GET /api/emissions/scopes?start_date={date}&end_date={date}
GET /api/emissions/heatmap?period={week|month}

# Blockchain
GET /api/blockchain/portfolio
GET /api/blockchain/transactions

# Tenant
GET /api/tenant/branding
```

## Environment Variables

```bash
# .env.local
NEXT_PUBLIC_OFFGRIDFLOW_API_URL=http://localhost:8090
```

## TypeScript Interfaces

### Emissions
```typescript
interface TrendDataPoint {
  date: string;
  scope1: number;
  scope2: number;
  scope3: number;
  total: number;
}

interface ScopeData {
  scope: string;
  emissions: number;
  percentage: number;
  activities: number;
}

interface HeatmapCell {
  date: string;
  hour: number;
  value: number;
  intensity: number;
}
```

### Blockchain
```typescript
interface Portfolio {
  totalValue: number;
  totalCredits: number;
  credits: CreditHolding[];
}

interface Transaction {
  id: string;
  type: 'mint' | 'buy' | 'sell' | 'transfer';
  timestamp: string;
  amount: number;
  price?: number;
  txHash: string;
  status: 'pending' | 'confirmed' | 'failed';
}
```

### Whitelabel
```typescript
interface TenantBranding {
  primaryColor: string;
  secondaryColor: string;
  logoUrl?: string;
  customCss?: string;
  fontFamily?: string;
}
```
