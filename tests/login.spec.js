import { test, expect } from '@playwright/test';

test.describe('Login Flow Tests', () => {
  test('User can login with valid credentials', async ({ page }) => {
    // Navigate to login page
    await page.goto('http://localhost:8000');
    
    // Fill in login form
    await page.fill('[data-testid="username"]', 'testuser@example.com');
    await page.fill('[data-testid="password"]', 'password123');
    
    // Click login button
    await page.click('[data-testid="login-button"]');
    
    // Verify successful login
    await expect(page.locator('[data-testid="welcome-message"]')).toBeVisible();
    await expect(page.locator('[data-testid="user-profile"]')).toContainText('testuser');
  });

  test('User sees error with invalid credentials', async ({ page }) => {
    // Navigate to login page
    await page.goto('http://localhost:8000');
    
    // Fill in invalid credentials
    await page.fill('[data-testid="username"]', 'invalid@example.com');
    await page.fill('[data-testid="password"]', 'wrongpassword');
    
    // Click login button
    await page.click('[data-testid="login-button"]');
    
    // Verify error message
    await expect(page.locator('[data-testid="error-message"]')).toBeVisible();
    await expect(page.locator('[data-testid="error-message"]')).toContainText('Invalid credentials');
  });

  test('Login form validation works', async ({ page }) => {
    // Navigate to login page
    await page.goto('http://localhost:8000');
    
    // Try to submit empty form
    await page.click('[data-testid="login-button"]');
    
    // Verify validation messages
    await expect(page.locator('[data-testid="username-error"]')).toBeVisible();
    await expect(page.locator('[data-testid="password-error"]')).toBeVisible();
  });
}); 