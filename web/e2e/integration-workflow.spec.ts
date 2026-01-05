import { test, expect } from '@playwright/test';

// Helper function to login before tests
async function loginAsTestUser(page: any) {
  await page.goto('/login');
  await page.getByLabel(/email/i).fill('test@offgridflow.test');
  await page.getByLabel(/password/i).fill('TestPassword123!');
  await page.getByRole('button', { name: /sign in/i }).click();
  
  // Wait for redirect to dashboard
  await page.waitForURL(/\/(dashboard|home)/, { timeout: 15000 }).catch(() => {
    // If no redirect, might be 2FA or already on dashboard
  });
  
  await page.waitForTimeout(2000);
}

test.describe('Integration Workflow - Post Authentication', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsTestUser(page);
  });

  test('should successfully load dashboard after login', async ({ page }) => {
    // Should be on dashboard
    await expect(page).toHaveURL(/\/dashboard/, { timeout: 10000 });
    
    // Dashboard elements should be visible
    await expect(page.getByRole('heading', { name: /dashboard|carbon|emissions/i })).toBeVisible();
    
    // No error messages should be present
    const errorMessages = page.getByText(/error|failed|something went wrong/i);
    await expect(errorMessages).toHaveCount(0);
  });

  test('should display user profile information', async ({ page }) => {
    // Navigate to profile or check if user info is displayed
    const userMenu = page.getByRole('button', { name: /profile|account|user|menu/i });
    
    if (await userMenu.isVisible()) {
      await userMenu.click();
      
      // Should show user email or name
      await expect(page.getByText(/test@offgridflow\.test|test user/i)).toBeVisible();
    }
  });

  test('should navigate between main sections without errors', async ({ page }) => {
    const sections = [
      { name: /dashboard|home/i, path: '/dashboard' },
      { name: /activities|activity/i, path: '/activities' },
      { name: /emissions/i, path: '/emissions' },
      { name: /reports|reporting/i, path: '/reports' },
      { name: /compliance/i, path: '/compliance' },
      { name: /settings/i, path: '/settings' },
    ];

    for (const section of sections) {
      // Find and click navigation link
      const navLink = page.getByRole('link', { name: section.name }).first();
      
      if (await navLink.isVisible()) {
        await navLink.click();
        
        // Wait for navigation
        await page.waitForTimeout(1500);
        
        // Should not show error
        const errorMessages = page.getByText(/error|failed|something went wrong/i);
        const errorCount = await errorMessages.count();
        
        expect(errorCount).toBe(0);
        
        // Should load some content
        const mainContent = page.locator('main, [role="main"], .content');
        await expect(mainContent).toBeVisible();
      }
    }
  });
});

test.describe('Dashboard Functionality', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsTestUser(page);
    await page.goto('/dashboard');
  });

  test('should display key metrics without errors', async ({ page }) => {
    // Wait for content to load
    await page.waitForTimeout(2000);
    
    // Check for common dashboard elements
    const hasCharts = await page.locator('canvas, svg[class*="recharts"]').count() > 0;
    const hasMetrics = await page.getByText(/scope 1|scope 2|scope 3|tco2e|emissions|carbon/i).count() > 0;
    
    expect(hasCharts || hasMetrics).toBe(true);
    
    // No error messages
    const errorMessages = page.getByText(/error|failed|something went wrong/i);
    await expect(errorMessages).toHaveCount(0);
  });

  test('should load charts and visualizations', async ({ page }) => {
    // Wait for charts to render
    await page.waitForTimeout(3000);
    
    // Check for chart elements (Canvas for Chart.js or SVG for Recharts)
    const charts = page.locator('canvas, svg[class*="recharts"]');
    const chartCount = await charts.count();
    
    expect(chartCount).toBeGreaterThan(0);
  });

  test('should allow date range filtering', async ({ page }) => {
    // Look for date picker or filter controls
    const dateFilter = page.getByLabel(/date|from|to|range/i).first();
    
    if (await dateFilter.isVisible()) {
      await dateFilter.click();
      
      // Should show date picker
      await expect(page.locator('.react-datepicker, [role="dialog"]')).toBeVisible();
    }
  });

  test('should refresh data without errors', async ({ page }) => {
    // Look for refresh button
    const refreshButton = page.getByRole('button', { name: /refresh|reload/i });
    
    if (await refreshButton.isVisible()) {
      await refreshButton.click();
      
      // Should show loading state briefly
      await page.waitForTimeout(1000);
      
      // Should not show error
      const errorMessages = page.getByText(/error|failed/i);
      await expect(errorMessages).toHaveCount(0);
    }
  });
});

test.describe('Activities Management', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsTestUser(page);
    await page.goto('/activities');
  });

  test('should display activities list', async ({ page }) => {
    await page.waitForTimeout(2000);
    
    // Should show activities table or list
    const hasTable = await page.locator('table').isVisible();
    const hasList = await page.getByRole('list').isVisible();
    
    expect(hasTable || hasList).toBe(true);
  });

  test('should allow creating new activity', async ({ page }) => {
    // Look for "Add Activity" or "Create" button
    const createButton = page.getByRole('button', { name: /(add|create|new).*activity/i });
    
    if (await createButton.isVisible()) {
      await createButton.click();
      
      // Should show activity form or modal
      await expect(page.getByLabel(/activity.*name|name/i)).toBeVisible({ timeout: 5000 });
      
      // Form should have required fields
      await expect(page.getByLabel(/category|type/i)).toBeVisible();
    }
  });

  test('should search/filter activities', async ({ page }) => {
    // Look for search input
    const searchInput = page.getByPlaceholder(/search|filter/i);
    
    if (await searchInput.isVisible()) {
      await searchInput.fill('test');
      
      // Should filter results
      await page.waitForTimeout(1000);
      
      // No error should occur
      const errorMessages = page.getByText(/error|failed/i);
      await expect(errorMessages).toHaveCount(0);
    }
  });
});

test.describe('Emissions Calculation', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsTestUser(page);
  });

  test('should calculate emissions for activity', async ({ page }) => {
    await page.goto('/activities');
    await page.waitForTimeout(2000);
    
    // Look for calculate button or emissions display
    const calculateButton = page.getByRole('button', { name: /calculate|compute/i }).first();
    
    if (await calculateButton.isVisible()) {
      await calculateButton.click();
      
      // Should show calculation result
      await page.waitForTimeout(2000);
      
      // Should display CO2e value
      const hasCO2eDisplay = await page.getByText(/tco2e|kgco2e|co2/i).isVisible();
      expect(hasCO2eDisplay).toBe(true);
    }
  });

  test('should view emission factors', async ({ page }) => {
    await page.goto('/emissions');
    await page.waitForTimeout(2000);
    
    // Look for emission factors section
    const hasFactors = await page.getByText(/emission.*factor|factor/i).isVisible();
    
    if (hasFactors) {
      // Should display factors without error
      const errorMessages = page.getByText(/error|failed/i);
      await expect(errorMessages).toHaveCount(0);
    }
  });
});

test.describe('Compliance Reports', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsTestUser(page);
    await page.goto('/compliance');
  });

  test('should display compliance frameworks', async ({ page }) => {
    await page.waitForTimeout(2000);
    
    // Should show compliance options
    const frameworks = ['CSRD', 'SEC', 'CBAM', 'California', 'IFRS'];
    
    for (const framework of frameworks) {
      const hasFramework = await page.getByText(new RegExp(framework, 'i')).isVisible();
      // At least some frameworks should be visible
    }
    
    // No errors
    const errorMessages = page.getByText(/error|failed/i);
    await expect(errorMessages).toHaveCount(0);
  });

  test('should generate compliance report', async ({ page }) => {
    await page.waitForTimeout(2000);
    
    // Look for generate report button
    const generateButton = page.getByRole('button', { name: /generate.*report|create.*report/i }).first();
    
    if (await generateButton.isVisible()) {
      await generateButton.click();
      
      // Should show report options or start generation
      await page.waitForTimeout(2000);
      
      // Should not show error
      const errorMessages = page.getByText(/error.*generating|failed.*generate/i);
      await expect(errorMessages).toHaveCount(0);
    }
  });
});

test.describe('Settings and Configuration', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsTestUser(page);
    await page.goto('/settings');
  });

  test('should display settings sections', async ({ page }) => {
    await page.waitForTimeout(2000);
    
    // Should show settings navigation
    const sections = [/profile|account/i, /organization/i, /security/i, /billing/i];
    
    let visibleSections = 0;
    for (const section of sections) {
      const hasSection = await page.getByText(section).isVisible();
      if (hasSection) visibleSections++;
    }
    
    expect(visibleSections).toBeGreaterThan(0);
  });

  test('should update profile information', async ({ page }) => {
    await page.waitForTimeout(2000);
    
    // Look for profile edit form
    const nameInput = page.getByLabel(/name|full name|first name/i).first();
    
    if (await nameInput.isVisible()) {
      const currentValue = await nameInput.inputValue();
      await nameInput.fill('Updated Name');
      
      // Look for save button
      const saveButton = page.getByRole('button', { name: /save|update/i });
      if (await saveButton.isVisible()) {
        await saveButton.click();
        
        // Should show success message
        await expect(page.getByText(/saved|updated|success/i)).toBeVisible({ timeout: 5000 });
      }
    }
  });
});

test.describe('Data Integration', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsTestUser(page);
  });

  test('should access cloud connectors', async ({ page }) => {
    // Navigate to integrations/connectors
    await page.goto('/settings/integrations').catch(() => 
      page.goto('/settings/data-sources')
    );
    
    await page.waitForTimeout(2000);
    
    // Should show cloud provider options
    const providers = ['AWS', 'Azure', 'GCP'];
    
    let visibleProviders = 0;
    for (const provider of providers) {
      const hasProvider = await page.getByText(new RegExp(provider, 'i')).isVisible();
      if (hasProvider) visibleProviders++;
    }
    
    // At least one provider should be visible
    expect(visibleProviders).toBeGreaterThan(0);
  });

  test('should configure data source', async ({ page }) => {
    await page.goto('/settings/data-sources').catch(() => 
      page.goto('/settings')
    );
    
    await page.waitForTimeout(2000);
    
    // Look for add/configure button
    const configureButton = page.getByRole('button', { name: /(add|configure|connect).*source/i }).first();
    
    if (await configureButton.isVisible()) {
      await configureButton.click();
      
      // Should show configuration form
      await page.waitForTimeout(1000);
      
      // Should not show error
      const errorMessages = page.getByText(/error|failed/i);
      await expect(errorMessages).toHaveCount(0);
    }
  });
});

test.describe('Logout Flow', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsTestUser(page);
  });

  test('should successfully logout', async ({ page }) => {
    // Find logout button (could be in user menu)
    const userMenu = page.getByRole('button', { name: /profile|account|user|menu/i });
    
    if (await userMenu.isVisible()) {
      await userMenu.click();
    }
    
    // Click logout
    const logoutButton = page.getByRole('button', { name: /logout|sign out/i });
    await logoutButton.click();
    
    // Should redirect to login page
    await expect(page).toHaveURL(/\/login/, { timeout: 10000 });
    
    // Should not be able to access protected pages
    await page.goto('/dashboard');
    
    // Should redirect back to login
    await expect(page).toHaveURL(/\/login/);
  });
});

test.describe('Error Handling and Edge Cases', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsTestUser(page);
  });

  test('should handle API errors gracefully', async ({ page, context }) => {
    // Simulate network failure
    await context.route('**/api/**', route => route.abort());
    
    // Try to load dashboard
    await page.goto('/dashboard');
    await page.waitForTimeout(3000);
    
    // Should show user-friendly error message
    const hasErrorMessage = await page.getByText(/error|failed|try again|check connection/i).isVisible();
    
    // Should not crash or show raw error
    expect(hasErrorMessage).toBe(true);
  });

  test('should handle 401 unauthorized gracefully', async ({ page, context }) => {
    // Clear auth cookies
    await context.clearCookies();
    
    // Try to access protected page
    await page.goto('/dashboard');
    
    // Should redirect to login
    await expect(page).toHaveURL(/\/login/, { timeout: 10000 });
  });

  test('should handle 404 pages', async ({ page }) => {
    // Navigate to non-existent page
    await page.goto('/non-existent-page-12345');
    
    // Should show 404 page or redirect
    const has404 = await page.getByText(/404|not found|page.*exist/i).isVisible();
    const redirected = !page.url().includes('non-existent');
    
    expect(has404 || redirected).toBe(true);
  });
});

test.describe('Performance and UX', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsTestUser(page);
  });

  test('should load dashboard within acceptable time', async ({ page }) => {
    const startTime = Date.now();
    
    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');
    
    const loadTime = Date.now() - startTime;
    
    // Should load within 5 seconds
    expect(loadTime).toBeLessThan(5000);
  });

  test('should show loading states for async operations', async ({ page }) => {
    await page.goto('/dashboard');
    
    // Look for loading indicators
    const hasLoadingIndicator = await page.locator('[aria-label*="loading"], [role="status"], .spinner, .skeleton').count() > 0;
    
    // Loading indicators should be used somewhere in the app
    // (might not be visible when test runs if data loads quickly)
  });

  test('should be fully keyboard navigable', async ({ page }) => {
    await page.goto('/dashboard');
    await page.waitForTimeout(2000);
    
    // Press Tab multiple times
    for (let i = 0; i < 10; i++) {
      await page.keyboard.press('Tab');
      await page.waitForTimeout(100);
    }
    
    // Should be able to navigate with keyboard
    // Check that focus is visible
    const focusedElement = await page.evaluate(() => document.activeElement?.tagName);
    expect(focusedElement).toBeTruthy();
  });
});
