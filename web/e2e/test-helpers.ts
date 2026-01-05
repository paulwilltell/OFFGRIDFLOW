import { Page, expect } from '@playwright/test';

/**
 * Test helper utilities for E2E tests
 */

export interface TestUser {
  email: string;
  password: string;
  firstName?: string;
  lastName?: string;
}

/**
 * Login helper function
 */
export async function loginUser(page: Page, user: TestUser) {
  await page.goto('/login');
  await page.getByLabel(/email/i).fill(user.email);
  await page.getByLabel(/password/i).fill(user.password);
  await page.getByRole('button', { name: /sign in/i }).click();
  
  // Wait for redirect to dashboard or 2FA
  await page.waitForTimeout(3000);
}

/**
 * Register a new user
 */
export async function registerUser(page: Page, user: TestUser) {
  await page.goto('/register');
  
  await page.getByLabel(/first name/i).fill(user.firstName || 'Test');
  await page.getByLabel(/last name/i).fill(user.lastName || 'User');
  await page.getByLabel(/email/i).fill(user.email);
  await page.getByLabel(/^password$/i).fill(user.password);
  await page.getByLabel(/confirm password/i).fill(user.password);
  
  await page.getByRole('button', { name: /create account/i }).click();
  
  // Wait for email verification screen
  await page.waitForTimeout(3000);
}

/**
 * Logout helper function
 */
export async function logoutUser(page: Page) {
  const userMenu = page.getByRole('button', { name: /profile|account|user|menu/i });
  
  if (await userMenu.isVisible()) {
    await userMenu.click();
    await page.waitForTimeout(500);
  }
  
  const logoutButton = page.getByRole('button', { name: /logout|sign out/i });
  await logoutButton.click();
  
  await page.waitForURL(/\/login/, { timeout: 10000 });
}

/**
 * Check for error messages
 */
export async function expectNoErrors(page: Page) {
  const errorMessages = page.getByText(/error|failed|something went wrong/i);
  const count = await errorMessages.count();
  
  if (count > 0) {
    const errorText = await errorMessages.first().textContent();
    console.log(`Unexpected error found: ${errorText}`);
  }
  
  expect(count).toBe(0);
}

/**
 * Wait for API call to complete
 */
export async function waitForApiResponse(page: Page, endpoint: string | RegExp, timeout = 10000) {
  return page.waitForResponse(
    response => {
      const url = response.url();
      const matches = typeof endpoint === 'string' 
        ? url.includes(endpoint) 
        : endpoint.test(url);
      return matches && response.status() < 400;
    },
    { timeout }
  );
}

/**
 * Generate unique test email
 */
export function generateTestEmail(prefix = 'test'): string {
  const timestamp = Date.now();
  const random = Math.floor(Math.random() * 1000);
  return `${prefix}.${timestamp}.${random}@offgridflow.test`;
}

/**
 * Generate secure test password
 */
export function generateTestPassword(): string {
  return `TestPassword${Date.now()}!`;
}

/**
 * Check if element is in viewport
 */
export async function isInViewport(page: Page, selector: string): Promise<boolean> {
  return page.locator(selector).evaluate((element) => {
    const rect = element.getBoundingClientRect();
    return (
      rect.top >= 0 &&
      rect.left >= 0 &&
      rect.bottom <= (window.innerHeight || document.documentElement.clientHeight) &&
      rect.right <= (window.innerWidth || document.documentElement.clientWidth)
    );
  });
}

/**
 * Take screenshot on failure
 */
export async function takeScreenshotOnFailure(page: Page, testName: string) {
  const screenshotPath = `screenshots/failure-${testName}-${Date.now()}.png`;
  await page.screenshot({ path: screenshotPath, fullPage: true });
  console.log(`Screenshot saved: ${screenshotPath}`);
}

/**
 * Mock API response
 */
export async function mockApiEndpoint(
  page: Page,
  endpoint: string | RegExp,
  response: any,
  status = 200
) {
  await page.route(endpoint, route => {
    route.fulfill({
      status,
      contentType: 'application/json',
      body: JSON.stringify(response)
    });
  });
}

/**
 * Clear all cookies and storage
 */
export async function clearSession(page: Page) {
  await page.context().clearCookies();
  await page.evaluate(() => {
    localStorage.clear();
    sessionStorage.clear();
  });
}

/**
 * Fill form field by label
 */
export async function fillFormField(page: Page, label: string | RegExp, value: string) {
  const field = page.getByLabel(label);
  await field.fill(value);
  await expect(field).toHaveValue(value);
}

/**
 * Submit form and wait for response
 */
export async function submitForm(page: Page, buttonName: string | RegExp) {
  const button = page.getByRole('button', { name: buttonName });
  await button.click();
  await page.waitForTimeout(1000);
}

/**
 * Navigate to section and wait for load
 */
export async function navigateToSection(page: Page, sectionName: string | RegExp) {
  const link = page.getByRole('link', { name: sectionName }).first();
  await link.click();
  await page.waitForLoadState('networkidle');
  await page.waitForTimeout(1000);
}

/**
 * Check if user is logged in
 */
export async function isLoggedIn(page: Page): Promise<boolean> {
  // Check for auth cookie or token
  const cookies = await page.context().cookies();
  const hasAuthCookie = cookies.some(c => 
    c.name.includes('token') || 
    c.name.includes('session') || 
    c.name.includes('auth')
  );
  
  return hasAuthCookie;
}

/**
 * Wait for page to be fully loaded
 */
export async function waitForFullLoad(page: Page) {
  await page.waitForLoadState('domcontentloaded');
  await page.waitForLoadState('networkidle');
  await page.waitForTimeout(1000);
}

/**
 * Test data factory
 */
export const TestData = {
  users: {
    valid: (): TestUser => ({
      email: generateTestEmail('valid'),
      password: 'ValidPassword123!',
      firstName: 'Valid',
      lastName: 'User'
    }),
    
    with2FA: (): TestUser => ({
      email: 'test-2fa@offgridflow.test',
      password: 'TestPassword123!',
      firstName: '2FA',
      lastName: 'User'
    }),
    
    admin: (): TestUser => ({
      email: 'admin@offgridflow.test',
      password: 'AdminPassword123!',
      firstName: 'Admin',
      lastName: 'User'
    })
  },
  
  activities: {
    electricity: {
      name: 'Office Electricity',
      category: 'energy',
      amount: 1000,
      unit: 'kWh'
    },
    
    travel: {
      name: 'Business Flight',
      category: 'travel',
      amount: 500,
      unit: 'km'
    }
  }
};
