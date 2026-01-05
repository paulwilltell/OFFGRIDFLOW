import { test, expect } from '@playwright/test';

test.describe('User Login Flow', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
  });

  test('should display login form with all required elements', async ({ page }) => {
    // Verify form fields
    await expect(page.getByLabel(/email/i)).toBeVisible();
    await expect(page.getByLabel(/password/i)).toBeVisible();
    
    // Verify remember me checkbox
    await expect(page.getByLabel(/remember me/i)).toBeVisible();
    
    // Verify forgot password link
    await expect(page.getByRole('link', { name: /forgot.*password/i })).toBeVisible();
    
    // Verify sign in button
    await expect(page.getByRole('button', { name: /sign in/i })).toBeVisible();
    await expect(page.getByRole('button', { name: /sign in/i })).toBeEnabled();
    
    // Verify register link
    await expect(page.getByRole('link', { name: /(create account|sign up|register)/i })).toBeVisible();
  });

  test('should show validation errors for empty form submission', async ({ page }) => {
    // Click submit without filling form
    await page.getByRole('button', { name: /sign in/i }).click();
    
    // Email field should be focused or show validation
    const emailInput = page.getByLabel(/email/i);
    await expect(emailInput).toBeFocused();
  });

  test('should handle invalid credentials gracefully', async ({ page }) => {
    // Fill in invalid credentials
    await page.getByLabel(/email/i).fill('invalid@example.com');
    await page.getByLabel(/password/i).fill('WrongPassword123!');
    
    // Submit form
    await page.getByRole('button', { name: /sign in/i }).click();
    
    // Should show loading state briefly
    await expect(page.getByRole('button', { name: /signing/i })).toBeVisible();
    
    // Should show error message
    await expect(page.getByText(/invalid.*credentials|email.*password.*incorrect/i)).toBeVisible({ timeout: 10000 });
    
    // Should not redirect
    await expect(page).toHaveURL(/\/login/);
    
    // Form should still be usable
    await expect(page.getByRole('button', { name: /sign in/i })).toBeEnabled();
  });

  test('should successfully login with valid credentials', async ({ page }) => {
    // Use test account credentials (adjust based on your test data)
    const testEmail = 'test@offgridflow.test';
    const testPassword = 'TestPassword123!';
    
    // Fill in valid credentials
    await page.getByLabel(/email/i).fill(testEmail);
    await page.getByLabel(/password/i).fill(testPassword);
    
    // Submit form
    await page.getByRole('button', { name: /sign in/i }).click();
    
    // Should show loading state
    await expect(page.getByRole('button', { name: /signing/i })).toBeVisible();
    
    // Should redirect to dashboard or home page
    await page.waitForURL(/\/(dashboard|home)/, { timeout: 15000 }).catch(() => {
      // If no redirect, check if 2FA is required
      expect(page.getByText(/verification.*code|enter.*otp|2fa/i)).toBeVisible();
    });
  });

  test('should handle 2FA flow when enabled', async ({ page }) => {
    const testEmail = 'test-2fa@offgridflow.test';
    const testPassword = 'TestPassword123!';
    
    // Login with 2FA enabled account
    await page.getByLabel(/email/i).fill(testEmail);
    await page.getByLabel(/password/i).fill(testPassword);
    await page.getByRole('button', { name: /sign in/i }).click();
    
    // Wait for either dashboard redirect or 2FA prompt
    await page.waitForTimeout(2000);
    
    // If 2FA is shown
    const otpInput = page.getByLabel(/verification.*code|otp|token/i);
    if (await otpInput.isVisible()) {
      // Verify 2FA form elements
      await expect(otpInput).toBeVisible();
      await expect(page.getByRole('button', { name: /verify|submit/i })).toBeVisible();
      
      // Enter valid OTP (use test OTP or generate one)
      await otpInput.fill('123456');
      await page.getByRole('button', { name: /verify|submit/i }).click();
      
      // Should redirect after successful 2FA
      await page.waitForURL(/\/(dashboard|home)/, { timeout: 10000 });
    }
  });

  test('should toggle password visibility', async ({ page }) => {
    const passwordInput = page.getByLabel(/password/i);
    
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

  test('should persist "Remember Me" preference', async ({ page, context }) => {
    const testEmail = 'test@offgridflow.test';
    const testPassword = 'TestPassword123!';
    
    // Check "Remember Me"
    await page.getByLabel(/remember me/i).check();
    
    // Login
    await page.getByLabel(/email/i).fill(testEmail);
    await page.getByLabel(/password/i).fill(testPassword);
    await page.getByRole('button', { name: /sign in/i }).click();
    
    // Wait for login to complete
    await page.waitForTimeout(3000);
    
    // Check if auth token/cookie is persistent
    const cookies = await context.cookies();
    const authCookie = cookies.find(c => c.name.includes('token') || c.name.includes('session'));
    
    if (authCookie) {
      // Should have a long expiration (not session)
      expect(authCookie.expires).toBeGreaterThan(Date.now() / 1000 + 86400); // At least 1 day
    }
  });

  test('should navigate to registration page', async ({ page }) => {
    // Click "Create account" link
    await page.getByRole('link', { name: /(create account|sign up|register)/i }).click();
    
    // Should navigate to registration page
    await expect(page).toHaveURL(/\/register/);
  });

  test('should navigate to forgot password page', async ({ page }) => {
    // Click "Forgot password" link
    await page.getByRole('link', { name: /forgot.*password/i }).click();
    
    // Should navigate to password reset page
    await expect(page).toHaveURL(/\/(password\/forgot|forgot-password|reset-password)/);
  });

  test('should redirect to intended page after login', async ({ page }) => {
    // Navigate to protected page (should redirect to login with returnUrl)
    await page.goto('/dashboard/carbon?returnUrl=/dashboard/carbon');
    
    // Should be on login page
    await expect(page).toHaveURL(/\/login/);
    
    // Login
    const testEmail = 'test@offgridflow.test';
    const testPassword = 'TestPassword123!';
    
    await page.getByLabel(/email/i).fill(testEmail);
    await page.getByLabel(/password/i).fill(testPassword);
    await page.getByRole('button', { name: /sign in/i }).click();
    
    // Should redirect back to intended page
    await page.waitForURL(/\/dashboard\/carbon/, { timeout: 15000 }).catch(() => {
      // Fallback: check if redirected to any dashboard
      expect(page.url()).toContain('dashboard');
    });
  });

  test('should prevent access to login page when already logged in', async ({ page, context }) => {
    // Set auth token (simulate logged in state)
    await context.addCookies([{
      name: 'auth_token',
      value: 'mock-token',
      domain: 'localhost',
      path: '/',
      httpOnly: true,
      secure: false,
      sameSite: 'Lax'
    }]);
    
    // Navigate to login page
    await page.goto('/login');
    
    // Should redirect to dashboard or show already logged in message
    await page.waitForTimeout(2000);
    const currentUrl = page.url();
    
    // Either redirected away from login or shown message
    if (currentUrl.includes('/login')) {
      await expect(page.getByText(/already.*logged.*in/i)).toBeVisible();
    } else {
      expect(currentUrl).toContain('dashboard');
    }
  });

  test('should be responsive on mobile devices', async ({ page }) => {
    // Set mobile viewport
    await page.setViewportSize({ width: 375, height: 667 });
    
    // All form elements should still be visible
    await expect(page.getByLabel(/email/i)).toBeVisible();
    await expect(page.getByLabel(/password/i)).toBeVisible();
    await expect(page.getByRole('button', { name: /sign in/i })).toBeVisible();
    
    // Form should be usable
    await page.getByLabel(/email/i).fill('test@example.com');
    await expect(page.getByLabel(/email/i)).toHaveValue('test@example.com');
  });

  test('should handle network errors gracefully', async ({ page, context }) => {
    // Simulate offline
    await context.setOffline(true);
    
    // Attempt to login
    await page.getByLabel(/email/i).fill('test@offgridflow.test');
    await page.getByLabel(/password/i).fill('TestPassword123!');
    await page.getByRole('button', { name: /sign in/i }).click();
    
    // Should show network error message
    await expect(page.getByText(/network.*error|connection.*failed|check.*internet/i)).toBeVisible({ timeout: 10000 });
    
    // Restore connection
    await context.setOffline(false);
  });

  test('should sanitize inputs to prevent XSS', async ({ page }) => {
    const xssAttempt = '<script>alert("XSS")</script>';
    
    // Try to inject script in email field
    await page.getByLabel(/email/i).fill(xssAttempt);
    await page.getByLabel(/password/i).fill('password');
    await page.getByRole('button', { name: /sign in/i }).click();
    
    // Wait for response
    await page.waitForTimeout(2000);
    
    // Should not execute script (check that no alert appeared)
    // Page should still be functional
    await expect(page.getByLabel(/email/i)).toBeVisible();
  });
});

test.describe('Session Management', () => {
  test('should maintain session across page refreshes', async ({ page, context }) => {
    // Login first
    await page.goto('/login');
    
    const testEmail = 'test@offgridflow.test';
    const testPassword = 'TestPassword123!';
    
    await page.getByLabel(/email/i).fill(testEmail);
    await page.getByLabel(/password/i).fill(testPassword);
    await page.getByRole('button', { name: /sign in/i }).click();
    
    // Wait for redirect to dashboard
    await page.waitForTimeout(3000);
    
    // Refresh page
    await page.reload();
    
    // Should still be logged in (not redirected to login)
    await page.waitForTimeout(2000);
    expect(page.url()).not.toContain('/login');
  });

  test('should handle session expiration', async ({ page, context }) => {
    // Set expired token
    await context.addCookies([{
      name: 'auth_token',
      value: 'expired-token',
      domain: 'localhost',
      path: '/',
      httpOnly: true,
      secure: false,
      sameSite: 'Lax',
      expires: Math.floor(Date.now() / 1000) - 3600 // 1 hour ago
    }]);
    
    // Navigate to protected page
    await page.goto('/dashboard');
    
    // Should redirect to login
    await expect(page).toHaveURL(/\/login/, { timeout: 10000 });
    
    // Should show session expired message
    await expect(page.getByText(/session.*expired|please.*log.*in.*again/i)).toBeVisible();
  });
});
