# OffGridFlow E2E Test Suite

Comprehensive end-to-end testing suite for validating authentication flows and integration workflows in the OffGridFlow web application.

## Overview

This test suite uses [Playwright](https://playwright.dev/) to automate browser testing across Chromium, Firefox, and WebKit. It validates:

- ✅ **Authentication Flows**: Registration, login, 2FA, password reset
- ✅ **User Experience**: No error messages during normal flows
- ✅ **Integration Workflows**: Dashboard, activities, emissions, compliance, settings
- ✅ **Error Handling**: Graceful degradation and user-friendly errors
- ✅ **Accessibility**: Keyboard navigation and responsive design
- ✅ **Security**: XSS prevention, session management, rate limiting

## Test Coverage

### Authentication Tests (`auth-*.spec.ts`)

#### 1. Registration Flow (`auth-registration.spec.ts`)
- ✅ Display and validate registration form
- ✅ Password strength validation (min 8 chars)
- ✅ Password confirmation matching
- ✅ Successful user registration
- ✅ Duplicate email handling
- ✅ Email verification flow
- ✅ Toggle password visibility
- ✅ Responsive design (mobile)

#### 2. Login Flow (`auth-login.spec.ts`)
- ✅ Display login form with all elements
- ✅ Invalid credentials handling
- ✅ Successful login with valid credentials
- ✅ 2FA verification flow
- ✅ "Remember Me" functionality
- ✅ Session persistence across refreshes
- ✅ Redirect to intended page after login
- ✅ Session expiration handling
- ✅ Network error handling
- ✅ XSS prevention

#### 3. Password Reset Flow (`auth-password-reset.spec.ts`)
- ✅ Request password reset
- ✅ Email validation
- ✅ Success message for security (prevents user enumeration)
- ✅ Reset password with valid token
- ✅ Password strength validation on reset
- ✅ Invalid/expired token handling
- ✅ Complete password reset journey

### Integration Tests (`integration-workflow.spec.ts`)

#### Post-Authentication Workflows
- ✅ Dashboard loading after login
- ✅ User profile information display
- ✅ Navigation between sections without errors
- ✅ Dashboard metrics and charts
- ✅ Date range filtering
- ✅ Activities management (CRUD)
- ✅ Emissions calculation
- ✅ Compliance reports generation
- ✅ Settings and configuration
- ✅ Data integration (cloud connectors)
- ✅ Logout flow

#### Error Handling
- ✅ API errors displayed gracefully
- ✅ 401 unauthorized redirect
- ✅ 404 page handling
- ✅ Network failure recovery

#### Performance & UX
- ✅ Dashboard loads within 5 seconds
- ✅ Loading states for async operations
- ✅ Keyboard navigation support

## Quick Start

### Installation

```bash
# Install dependencies
cd web
npm install

# Install Playwright browsers
npx playwright install --with-deps
```

### Running Tests

#### Run all tests (default: Chromium)
```bash
npm run test:e2e
```

#### Run tests in specific browser
```bash
# Chromium
npx playwright test --project=chromium

# Firefox
npx playwright test --project=firefox

# WebKit (Safari)
npx playwright test --project=webkit

# All browsers
npx playwright test
```

#### Run tests with UI mode (recommended for development)
```bash
npm run test:e2e:ui
```

#### Run tests in headed mode (watch browser)
```bash
npx playwright test --headed
```

#### Run tests in debug mode
```bash
npm run test:e2e:debug
```

#### Run specific test file
```bash
npx playwright test e2e/auth-login.spec.ts
```

#### Run tests matching pattern
```bash
npx playwright test --grep "login"
```

### View Test Reports

```bash
# Open HTML report
npm run test:e2e:report

# Or
npx playwright show-report
```

## Using Test Scripts

### PowerShell (Windows)
```powershell
# Run tests in Chromium
.\scripts\run-e2e-tests.ps1

# Run tests in Firefox with headed mode
.\scripts\run-e2e-tests.ps1 -Browser firefox -Headed

# Run tests in UI mode
.\scripts\run-e2e-tests.ps1 -UI

# Show report after tests
.\scripts\run-e2e-tests.ps1 -Report
```

### Bash (Linux/macOS)
```bash
# Make script executable
chmod +x scripts/run-e2e-tests.sh

# Run tests in Chromium
./scripts/run-e2e-tests.sh

# Run tests in Firefox with headed mode
./scripts/run-e2e-tests.sh --browser firefox --headed

# Run tests in UI mode
./scripts/run-e2e-tests.sh --ui

# Show report after tests
./scripts/run-e2e-tests.sh --report
```

## Test Helpers

The `test-helpers.ts` file provides utility functions:

```typescript
import { loginUser, registerUser, logoutUser, expectNoErrors } from './test-helpers';

// Login helper
await loginUser(page, {
  email: 'test@example.com',
  password: 'Password123!'
});

// Register new user
await registerUser(page, {
  email: 'new@example.com',
  password: 'Password123!',
  firstName: 'John',
  lastName: 'Doe'
});

// Check for errors
await expectNoErrors(page);

// Generate unique test data
const email = generateTestEmail('test');
const password = generateTestPassword();
```

## Configuration

Edit `playwright.config.ts` to customize:

```typescript
export default defineConfig({
  testDir: './e2e',
  timeout: 30000,
  retries: process.env.CI ? 2 : 0,
  use: {
    baseURL: 'http://localhost:3000',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },
  // ... more config
});
```

## Environment Variables

Create `.env.test` for test-specific configuration:

```bash
# Base URL
PLAYWRIGHT_BASE_URL=http://localhost:3000

# Test user credentials
TEST_USER_EMAIL=test@offgridflow.test
TEST_USER_PASSWORD=TestPassword123!

# API endpoint (if different from base URL)
API_BASE_URL=http://localhost:8080
```

## Best Practices

### Writing Tests

1. **Use Descriptive Test Names**
   ```typescript
   test('should successfully login with valid credentials', async ({ page }) => {
     // ...
   });
   ```

2. **Use Page Object Model for Reusability**
   ```typescript
   class LoginPage {
     async login(email: string, password: string) {
       await this.page.getByLabel(/email/i).fill(email);
       await this.page.getByLabel(/password/i).fill(password);
       await this.page.getByRole('button', { name: /sign in/i }).click();
     }
   }
   ```

3. **Wait for Elements Properly**
   ```typescript
   // Good - wait for element to be visible
   await expect(page.getByText('Dashboard')).toBeVisible();
   
   // Avoid - arbitrary timeouts
   await page.waitForTimeout(5000); // Only use when necessary
   ```

4. **Use Locators that Match User Behavior**
   ```typescript
   // Good - accessible locators
   page.getByRole('button', { name: /sign in/i })
   page.getByLabel(/email/i)
   
   // Avoid - brittle selectors
   page.locator('#login-btn')
   page.locator('.css-class-xyz')
   ```

5. **Clean Up After Tests**
   ```typescript
   test.afterEach(async ({ page }) => {
     await clearSession(page);
   });
   ```

### Debugging

1. **Use Playwright Inspector**
   ```bash
   npx playwright test --debug
   ```

2. **Use UI Mode**
   ```bash
   npx playwright test --ui
   ```

3. **Generate Tests with Codegen**
   ```bash
   npm run test:e2e:codegen
   # Opens browser and records your actions
   ```

4. **View Trace Files**
   ```bash
   npx playwright show-trace trace.zip
   ```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: E2E Tests

on: [push, pull_request]

jobs:
  e2e:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Node.js
        uses: actions/setup-node@v3
        with:
          node-version: '18'
          
      - name: Install dependencies
        run: |
          cd web
          npm ci
          npx playwright install --with-deps
          
      - name: Run E2E tests
        run: |
          cd web
          npm run test:e2e
          
      - name: Upload test results
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: playwright-report
          path: web/playwright-report/
```

## Test Data Management

### Using Test Fixtures

```typescript
import { test as base } from '@playwright/test';

type Fixtures = {
  authenticatedPage: Page;
};

export const test = base.extend<Fixtures>({
  authenticatedPage: async ({ page }, use) => {
    await loginUser(page, TestData.users.valid());
    await use(page);
    await logoutUser(page);
  },
});

// Use in tests
test('should access dashboard', async ({ authenticatedPage }) => {
  await authenticatedPage.goto('/dashboard');
  // Already logged in!
});
```

## Troubleshooting

### Common Issues

**Issue**: Tests fail with "Timeout waiting for element"
- **Solution**: Increase timeout or check if element selector is correct
- Use `await page.pause()` to inspect page state

**Issue**: Tests pass locally but fail in CI
- **Solution**: Check for race conditions, ensure proper waits
- Use `waitForLoadState('networkidle')` before assertions

**Issue**: Browser doesn't close after tests
- **Solution**: Ensure proper cleanup in `afterEach` hooks
- Check for uncaught exceptions

**Issue**: Tests are flaky
- **Solution**: Avoid `waitForTimeout`, use proper element waits
- Use `waitForResponse` for API calls
- Add retries in config: `retries: 2`

## Performance Optimization

1. **Run tests in parallel**
   ```bash
   npx playwright test --workers=4
   ```

2. **Use test sharding for large suites**
   ```bash
   npx playwright test --shard=1/4
   ```

3. **Skip unnecessary tests**
   ```typescript
   test.skip('slow test', async ({ page }) => {
     // ...
   });
   ```

## Maintenance

### Regular Tasks

- [ ] Update Playwright: `npm update @playwright/test`
- [ ] Update browser binaries: `npx playwright install`
- [ ] Review and update test data
- [ ] Check for deprecated API usage
- [ ] Update screenshots if UI changes

### Test Health Metrics

Monitor these metrics to ensure test suite quality:

- **Pass Rate**: Should be > 95%
- **Execution Time**: Should be < 10 minutes for full suite
- **Flakiness**: Should be < 2% (tests passing sometimes, failing others)
- **Coverage**: Should cover all critical user journeys

## Resources

- [Playwright Documentation](https://playwright.dev/)
- [Playwright Best Practices](https://playwright.dev/docs/best-practices)
- [Playwright API Reference](https://playwright.dev/docs/api/class-playwright)
- [Playwright Discord Community](https://discord.gg/playwright)

## Support

For issues or questions:

1. Check this documentation
2. Review Playwright documentation
3. Check existing test examples
4. Open an issue in the project repository

---

**Last Updated**: December 27, 2025  
**Test Suite Version**: 1.0.0  
**Playwright Version**: 1.48.2
