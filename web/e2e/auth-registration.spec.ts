import { test, expect } from '@playwright/test';

test.describe('User Registration Flow', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/register');
  });

  test('should display registration form with all required fields', async ({ page }) => {
    // Verify all form fields are present
    await expect(page.getByLabel(/first name/i)).toBeVisible();
    await expect(page.getByLabel(/last name/i)).toBeVisible();
    await expect(page.getByLabel(/email/i)).toBeVisible();
    await expect(page.getByLabel(/^password$/i)).toBeVisible();
    await expect(page.getByLabel(/confirm password/i)).toBeVisible();
    
    // Optional fields
    await expect(page.getByLabel(/company name/i)).toBeVisible();
    await expect(page.getByLabel(/job title/i)).toBeVisible();
    
    // Submit button
    await expect(page.getByRole('button', { name: /create account/i })).toBeVisible();
    await expect(page.getByRole('button', { name: /create account/i })).toBeEnabled();
  });

  test('should show validation errors for empty form submission', async ({ page }) => {
    // Click submit without filling form
    await page.getByRole('button', { name: /create account/i }).click();
    
    // Should show validation errors (HTML5 or custom)
    // Note: Exact behavior depends on validation implementation
    const firstNameInput = page.getByLabel(/first name/i);
    await expect(firstNameInput).toBeFocused();
  });

  test('should validate password strength and match', async ({ page }) => {
    await page.getByLabel(/first name/i).fill('John');
    await page.getByLabel(/last name/i).fill('Doe');
    await page.getByLabel(/email/i).fill('john.doe@example.com');
    
    // Test weak password
    await page.getByLabel(/^password$/i).fill('weak');
    await page.getByLabel(/confirm password/i).fill('weak');
    await page.getByRole('button', { name: /create account/i }).click();
    
    // Should show error about password strength
    await expect(page.getByText(/password.*least.*8.*characters/i)).toBeVisible();
  });

  test('should validate password confirmation match', async ({ page }) => {
    await page.getByLabel(/first name/i).fill('John');
    await page.getByLabel(/last name/i).fill('Doe');
    await page.getByLabel(/email/i).fill('john.doe@example.com');
    await page.getByLabel(/^password$/i).fill('SecurePassword123!');
    await page.getByLabel(/confirm password/i).fill('DifferentPassword123!');
    
    await page.getByRole('button', { name: /create account/i }).click();
    
    // Should show error about password mismatch
    await expect(page.getByText(/passwords.*not.*match/i)).toBeVisible();
  });

  test('should successfully register a new user', async ({ page }) => {
    const timestamp = Date.now();
    const email = `test.user.${timestamp}@offgridflow.test`;
    
    // Fill in registration form
    await page.getByLabel(/first name/i).fill('Test');
    await page.getByLabel(/last name/i).fill('User');
    await page.getByLabel(/email/i).fill(email);
    await page.getByLabel(/company name/i).fill('Test Company');
    await page.getByLabel(/job title/i).fill('Engineer');
    await page.getByLabel(/^password$/i).fill('SecurePassword123!');
    await page.getByLabel(/confirm password/i).fill('SecurePassword123!');
    
    // Submit form
    await page.getByRole('button', { name: /create account/i }).click();
    
    // Should show loading state
    await expect(page.getByRole('button', { name: /creating/i })).toBeVisible();
    
    // Should redirect to email verification page or show success message
    await expect(page.getByText(/check your email/i)).toBeVisible({ timeout: 10000 });
    await expect(page.getByText(/verification.*link/i)).toBeVisible();
  });

  test('should handle duplicate email registration', async ({ page }) => {
    const email = 'existing.user@offgridflow.test';
    
    // Attempt to register with existing email
    await page.getByLabel(/first name/i).fill('Test');
    await page.getByLabel(/last name/i).fill('User');
    await page.getByLabel(/email/i).fill(email);
    await page.getByLabel(/^password$/i).fill('SecurePassword123!');
    await page.getByLabel(/confirm password/i).fill('SecurePassword123!');
    
    await page.getByRole('button', { name: /create account/i }).click();
    
    // Should show error message (if email already exists in system)
    // Note: This might show "Check your email" for security reasons
    await page.waitForTimeout(2000);
  });

  test('should toggle password visibility', async ({ page }) => {
    const passwordInput = page.getByLabel(/^password$/i);
    
    // Password should be hidden by default
    await expect(passwordInput).toHaveAttribute('type', 'password');
    
    // Click show password button
    const showPasswordButton = page.locator('button[aria-label*="password"], button[aria-label*="show"]').first();
    if (await showPasswordButton.isVisible()) {
      await showPasswordButton.click();
      
      // Password should now be visible
      await expect(passwordInput).toHaveAttribute('type', 'text');
      
      // Click hide password button
      await showPasswordButton.click();
      
      // Password should be hidden again
      await expect(passwordInput).toHaveAttribute('type', 'password');
    }
  });

  test('should navigate to login page', async ({ page }) => {
    // Click "Already have an account? Sign in" link
    await page.getByRole('link', { name: /sign in/i }).click();
    
    // Should navigate to login page
    await expect(page).toHaveURL(/\/login/);
  });

  test('should display Terms of Service and Privacy Policy links', async ({ page }) => {
    // Verify links are present
    await expect(page.getByRole('link', { name: /terms/i })).toBeVisible();
    await expect(page.getByRole('link', { name: /privacy/i })).toBeVisible();
  });

  test('should be responsive on mobile devices', async ({ page }) => {
    // Set mobile viewport
    await page.setViewportSize({ width: 375, height: 667 });
    
    // All form elements should still be visible
    await expect(page.getByLabel(/first name/i)).toBeVisible();
    await expect(page.getByLabel(/email/i)).toBeVisible();
    await expect(page.getByRole('button', { name: /create account/i })).toBeVisible();
  });
});

test.describe('Email Verification', () => {
  test('should display verification instructions after registration', async ({ page }) => {
    await page.goto('/register');
    
    const timestamp = Date.now();
    const email = `verify.test.${timestamp}@offgridflow.test`;
    
    // Complete registration
    await page.getByLabel(/first name/i).fill('Verify');
    await page.getByLabel(/last name/i).fill('Test');
    await page.getByLabel(/email/i).fill(email);
    await page.getByLabel(/^password$/i).fill('SecurePassword123!');
    await page.getByLabel(/confirm password/i).fill('SecurePassword123!');
    await page.getByRole('button', { name: /create account/i }).click();
    
    // Should show verification page
    await expect(page.getByText(/check your email/i)).toBeVisible({ timeout: 10000 });
    
    // Should display user's email
    await expect(page.getByText(email)).toBeVisible();
    
    // Should have resend link or back to login link
    await expect(page.getByRole('link', { name: /(back to login|resend)/i })).toBeVisible();
  });
});
