import { test, expect } from '@playwright/test';

test.describe('Real Web Application Tests', () => {
  test('Test Demo Bank Login', async ({ page }) => {
    // Navigate to demo bank login page
    await page.goto('https://demo.testfire.net/login.jsp');
    
    // Fill in login form
    await page.fill('#uid', 'admin');
    await page.fill('#passw', 'admin');
    
    // Click login button
    await page.click('input[type="submit"]');
    
    // Verify successful login - should redirect to home page
    await expect(page).toHaveURL(/.*home.*/);
    
    // Verify we're logged in by checking for logout link
    await expect(page.locator('a[href*="logout"]')).toBeVisible();
  });

  test('Test Demo Bank Invalid Login', async ({ page }) => {
    // Navigate to demo bank login page
    await page.goto('https://demo.testfire.net/login.jsp');
    
    // Fill in invalid credentials
    await page.fill('#uid', 'invalid');
    await page.fill('#passw', 'wrongpassword');
    
    // Click login button
    await page.click('input[type="submit"]');
    
    // Verify error message appears
    await expect(page.locator('text=Login Failed')).toBeVisible();
  });

  test('Test Demo Bank Navigation', async ({ page }) => {
    // Navigate to demo bank
    await page.goto('https://demo.testfire.net/');
    
    // Verify main navigation elements
    await expect(page.locator('text=Online Banking Demo')).toBeVisible();
    await expect(page.locator('a[href*="login"]')).toBeVisible();
    await expect(page.locator('a[href*="about"]')).toBeVisible();
  });

  test('Test Demo E-commerce Site', async ({ page }) => {
    // Navigate to demo e-commerce site
    await page.goto('https://demo.opencart.com/');
    
    // Verify site loads
    await expect(page.locator('text=Your Store')).toBeVisible();
    
    // Search for a product
    await page.fill('#search input[name="search"]', 'phone');
    await page.click('#search button');
    
    // Verify search results
    await expect(page.locator('text=Search')).toBeVisible();
  });

  test('Test Demo Form Site', async ({ page }) => {
    // Navigate to demo form site
    await page.goto('https://demoqa.com/');
    
    // Verify site loads
    await expect(page.locator('text=ToolsQA')).toBeVisible();
    
    // Navigate to forms
    await page.click('text=Forms');
    
    // Verify forms page loads
    await expect(page.locator('text=Practice Form')).toBeVisible();
  });
}); 