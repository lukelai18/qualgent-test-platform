import { test, expect } from '@playwright/test';

test.describe('Onboarding Flow Tests', () => {
  test('New user can complete onboarding flow', async ({ page }) => {
    // Start onboarding
    await page.goto('http://localhost:8000');
    // Navigate to onboarding page
    await page.evaluate(() => {
        showPage('onboarding-page');
    });
    
    // Step 1: Welcome screen
    await expect(page.locator('[data-testid="get-started-button"]')).toBeVisible();
    await page.click('[data-testid="get-started-button"]');
    
    // Step 2: Profile setup
    await page.fill('[data-testid="first-name"]', 'John');
    await page.fill('[data-testid="last-name"]', 'Doe');
    await page.fill('[data-testid="email"]', 'john.doe@example.com');
    await page.click('[data-testid="next-button"]');
    
    // Step 3: Preferences
    await page.click('[data-testid="preference-option-1"]');
    await page.click('[data-testid="preference-option-3"]');
    await page.click('[data-testid="next-button"]');
    
    // Step 4: Verification
    await expect(page.locator('[data-testid="verification-code"]')).toBeVisible();
    await page.fill('[data-testid="verification-input"]', '123456');
    await page.click('[data-testid="verify-button"]');
    
    // Verify completion
    await expect(page.locator('[data-testid="onboarding-complete"]')).toBeVisible();
  });

  test('User can skip optional onboarding steps', async ({ page }) => {
    // Start onboarding
    await page.goto('http://localhost:8000');
    // Navigate to onboarding page
    await page.evaluate(() => {
        showPage('onboarding-page');
    });
    
    // Skip to end
    await page.click('[data-testid="skip-onboarding"]');
    
    // Verify skip confirmation (browser will show confirm dialog)
    // await page.click('[data-testid="confirm-skip"]');
    
    // Verify redirected to dashboard
    await expect(page.locator('[data-testid="welcome-message"]')).toBeVisible();
  });

  test('Onboarding form validation works', async ({ page }) => {
    // Start onboarding
    await page.goto('http://localhost:8000');
    // Navigate to onboarding page
    await page.evaluate(() => {
        showPage('onboarding-page');
    });
    await page.click('[data-testid="get-started-button"]');
    
    // Try to proceed without filling required fields
    await page.click('[data-testid="next-button"]');
    
    // Verify validation messages
    await expect(page.locator('[data-testid="first-name-error"]')).toBeVisible();
    await expect(page.locator('[data-testid="email-error"]')).toBeVisible();
  });
}); 