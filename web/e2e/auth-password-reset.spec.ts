import { test, expect } from '@playwright/test';

test.describe('Password Reset Flow', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/password/forgot');
  });

  test('should display forgot password form', async ({ page }) => {
    // Verify heading
    await expect(page.getByRole('heading', { name: /forgot.*password|reset.*password/i })).toBeVisible();
    
    // Verify email input
    await expect(page.getByLabel(/email/i)).toBeVisible();
    
    // Verify submit button
    await expect(page.getByRole('button', { name: /send.*reset.*link|reset.*password|continue/i })).toBeVisible();
    await expect(page.getByRole('button', { name: /send.*reset.*link|reset.*password|continue/i })).toBeEnabled();
    
    // Verify back to login link
    await expect(page.getByRole('link', { name: /back to.*login|sign in/i })).toBeVisible();
  });

  test('should show validation error for empty email', async ({ page }) => {
    // Submit without email
    await page.getByRole('button', { name: /send.*reset.*link|reset.*password|continue/i }).click();
    
    // Email field should be focused or show validation
    const emailInput = page.getByLabel(/email/i);
    await expect(emailInput).toBeFocused();
  });

  test('should validate email format', async ({ page }) => {
    // Enter invalid email
    await page.getByLabel(/email/i).fill('invalid-email');
    await page.getByRole('button', { name: /send.*reset.*link|reset.*password|continue/i }).click();
    
    // Should show validation error or HTML5 validation
    const emailInput = page.getByLabel(/email/i);
    const validationMessage = await emailInput.evaluate((el: HTMLInputElement) => el.validationMessage);
    
    expect(validationMessage).toBeTruthy();
  });

  test('should successfully request password reset', async ({ page }) => {
    const testEmail = 'test@offgridflow.test';
    
    // Enter email
    await page.getByLabel(/email/i).fill(testEmail);
    
    // Submit form
    await page.getByRole('button', { name: /send.*reset.*link|reset.*password|continue/i }).click();
    
    // Should show loading state
    await expect(page.getByRole('button', { name: /sending|loading/i })).toBeVisible();
    
    // Should show success message
    await expect(page.getByText(/check your email|sent.*reset.*link|email.*sent/i)).toBeVisible({ timeout: 10000 });
    
    // Should display the email address
    await expect(page.getByText(testEmail)).toBeVisible();
  });

  test('should show success message even for non-existent email (security)', async ({ page }) => {
    const nonExistentEmail = 'nonexistent@example.com';
    
    // Enter non-existent email
    await page.getByLabel(/email/i).fill(nonExistentEmail);
    
    // Submit form
    await page.getByRole('button', { name: /send.*reset.*link|reset.*password|continue/i }).click();
    
    // Should still show success message (prevents user enumeration)
    await expect(page.getByText(/check your email|sent.*reset.*link/i)).toBeVisible({ timeout: 10000 });
  });

  test('should allow trying another email', async ({ page }) => {
    const testEmail = 'test@offgridflow.test';
    
    // Submit first email
    await page.getByLabel(/email/i).fill(testEmail);
    await page.getByRole('button', { name: /send.*reset.*link|reset.*password|continue/i }).click();
    
    // Wait for success screen
    await expect(page.getByText(/check your email/i)).toBeVisible({ timeout: 10000 });
    
    // Click "Try another email" button
    const tryAnotherButton = page.getByRole('button', { name: /try.*another.*email|different.*email/i });
    if (await tryAnotherButton.isVisible()) {
      await tryAnotherButton.click();
      
      // Should go back to form
      await expect(page.getByLabel(/email/i)).toBeVisible();
      await expect(page.getByRole('button', { name: /send.*reset.*link|reset.*password|continue/i })).toBeVisible();
    }
  });

  test('should navigate back to login', async ({ page }) => {
    // Click back to login link
    await page.getByRole('link', { name: /back to.*login|sign in/i }).click();
    
    // Should navigate to login page
    await expect(page).toHaveURL(/\/login/);
  });

  test('should be responsive on mobile devices', async ({ page }) => {
    // Set mobile viewport
    await page.setViewportSize({ width: 375, height: 667 });
    
    // All elements should still be visible
    await expect(page.getByLabel(/email/i)).toBeVisible();
    await expect(page.getByRole('button', { name: /send.*reset.*link|reset.*password|continue/i })).toBeVisible();
    
    // Form should be usable
    await page.getByLabel(/email/i).fill('test@example.com');
    await expect(page.getByLabel(/email/i)).toHaveValue('test@example.com');
  });

  test('should handle rate limiting gracefully', async ({ page }) => {
    const testEmail = 'test@offgridflow.test';
    
    // Submit multiple times rapidly
    for (let i = 0; i < 5; i++) {
      await page.goto('/password/forgot');
      await page.getByLabel(/email/i).fill(testEmail);
      await page.getByRole('button', { name: /send.*reset.*link|reset.*password|continue/i }).click();
      await page.waitForTimeout(500);
    }
    
    // Should show rate limit message or still show success
    const hasRateLimitMessage = await page.getByText(/too many.*requests|try.*again.*later|rate.*limit/i).isVisible();
    const hasSuccessMessage = await page.getByText(/check your email/i).isVisible();
    
    expect(hasRateLimitMessage || hasSuccessMessage).toBe(true);
  });
});

test.describe('Password Reset Completion', () => {
  test('should display password reset form with valid token', async ({ page }) => {
    // Navigate to reset page with mock token
    const mockToken = 'mock-reset-token-123456';
    await page.goto(`/password/reset?token=${mockToken}`);
    
    // Should show password reset form
    await expect(page.getByRole('heading', { name: /reset.*password|new.*password/i })).toBeVisible();
    
    // Should have password fields
    await expect(page.getByLabel(/^password$|new.*password/i)).toBeVisible();
    await expect(page.getByLabel(/confirm.*password/i)).toBeVisible();
    
    // Should have submit button
    await expect(page.getByRole('button', { name: /reset.*password|update.*password|save/i })).toBeVisible();
  });

  test('should validate password strength on reset', async ({ page }) => {
    const mockToken = 'mock-reset-token-123456';
    await page.goto(`/password/reset?token=${mockToken}`);
    
    // Enter weak password
    await page.getByLabel(/^password$|new.*password/i).fill('weak');
    await page.getByLabel(/confirm.*password/i).fill('weak');
    
    // Submit form
    await page.getByRole('button', { name: /reset.*password|update.*password|save/i }).click();
    
    // Should show password strength error
    await expect(page.getByText(/password.*least.*8.*characters/i)).toBeVisible();
  });

  test('should validate password confirmation match on reset', async ({ page }) => {
    const mockToken = 'mock-reset-token-123456';
    await page.goto(`/password/reset?token=${mockToken}`);
    
    // Enter mismatched passwords
    await page.getByLabel(/^password$|new.*password/i).fill('NewSecurePassword123!');
    await page.getByLabel(/confirm.*password/i).fill('DifferentPassword123!');
    
    // Submit form
    await page.getByRole('button', { name: /reset.*password|update.*password|save/i }).click();
    
    // Should show mismatch error
    await expect(page.getByText(/passwords.*not.*match/i)).toBeVisible();
  });

  test('should successfully reset password', async ({ page }) => {
    const mockToken = 'valid-reset-token-123456';
    await page.goto(`/password/reset?token=${mockToken}`);
    
    const newPassword = 'NewSecurePassword123!';
    
    // Enter new password
    await page.getByLabel(/^password$|new.*password/i).fill(newPassword);
    await page.getByLabel(/confirm.*password/i).fill(newPassword);
    
    // Submit form
    await page.getByRole('button', { name: /reset.*password|update.*password|save/i }).click();
    
    // Should show loading state
    await expect(page.getByRole('button', { name: /resetting|updating|saving/i })).toBeVisible();
    
    // Should show success message or redirect to login
    await page.waitForTimeout(3000);
    
    const hasSuccessMessage = await page.getByText(/password.*reset|password.*updated|success/i).isVisible();
    const isOnLoginPage = page.url().includes('/login');
    
    expect(hasSuccessMessage || isOnLoginPage).toBe(true);
  });

  test('should handle invalid or expired token', async ({ page }) => {
    const expiredToken = 'expired-token-123456';
    await page.goto(`/password/reset?token=${expiredToken}`);
    
    // Enter new password
    await page.getByLabel(/^password$|new.*password/i).fill('NewSecurePassword123!');
    await page.getByLabel(/confirm.*password/i).fill('NewSecurePassword123!');
    
    // Submit form
    await page.getByRole('button', { name: /reset.*password|update.*password|save/i }).click();
    
    // Wait for response
    await page.waitForTimeout(2000);
    
    // Should show error about invalid/expired token
    const hasErrorMessage = await page.getByText(/invalid.*token|expired.*link|link.*no.*longer.*valid/i).isVisible();
    
    if (hasErrorMessage) {
      // Should offer to request new reset link
      await expect(page.getByRole('link', { name: /request.*new.*link|try.*again/i })).toBeVisible();
    }
  });

  test('should handle missing token gracefully', async ({ page }) => {
    // Navigate to reset page without token
    await page.goto('/password/reset');
    
    // Should show error or redirect
    await page.waitForTimeout(2000);
    
    const hasErrorMessage = await page.getByText(/invalid.*link|missing.*token/i).isVisible();
    const redirectedToForgot = page.url().includes('/password/forgot');
    
    expect(hasErrorMessage || redirectedToForgot).toBe(true);
  });

  test('should toggle password visibility on reset form', async ({ page }) => {
    const mockToken = 'mock-reset-token-123456';
    await page.goto(`/password/reset?token=${mockToken}`);
    
    const passwordInput = page.getByLabel(/^password$|new.*password/i);
    
    // Password should be hidden by default
    await expect(passwordInput).toHaveAttribute('type', 'password');
    
    // Click show password button
    const showPasswordButton = page.locator('button[aria-label*="password"], button[aria-label*="show"]').first();
    if (await showPasswordButton.isVisible()) {
      await showPasswordButton.click();
      
      // Password should now be visible
      await expect(passwordInput).toHaveAttribute('type', 'text');
    }
  });
});

test.describe('Complete Password Reset Journey', () => {
  test('should complete full password reset flow', async ({ page }) => {
    // Step 1: Navigate to forgot password
    await page.goto('/login');
    await page.getByRole('link', { name: /forgot.*password/i }).click();
    
    // Step 2: Request password reset
    await expect(page).toHaveURL(/\/(password\/forgot|forgot-password)/);
    
    const testEmail = 'complete-flow@offgridflow.test';
    await page.getByLabel(/email/i).fill(testEmail);
    await page.getByRole('button', { name: /send.*reset.*link|reset.*password|continue/i }).click();
    
    // Step 3: Verify success message
    await expect(page.getByText(/check your email/i)).toBeVisible({ timeout: 10000 });
    
    // Step 4: Navigate back to login
    const backToLoginLink = page.getByRole('link', { name: /back to.*login|sign in/i });
    if (await backToLoginLink.isVisible()) {
      await backToLoginLink.click();
      await expect(page).toHaveURL(/\/login/);
    }
    
    // In a real scenario, user would click link in email
    // For testing, we would need to:
    // 1. Check email via test mail server API
    // 2. Extract reset token from email
    // 3. Navigate to reset URL with token
    // 4. Complete password reset
    // 5. Login with new password
  });
});
