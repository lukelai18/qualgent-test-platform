import { test, expect } from '@playwright/test';

test.describe('End-to-End User Journey', () => {
  test('Complete user journey from signup to purchase', async ({ page }) => {
    // 1. User Registration
    await page.goto('http://localhost:8000');
    // Navigate to signup page
    await page.evaluate(() => {
        showPage('signup-page');
    });
    await page.fill('[data-testid="email"]', 'newuser@example.com');
    await page.fill('[data-testid="password"]', 'securepassword123');
    await page.fill('[data-testid="confirm-password"]', 'securepassword123');
    await page.click('[data-testid="signup-button"]');
    
    // Verify email verification
    await expect(page.locator('[data-testid="verification-sent"]')).toBeVisible();
    
    // 2. Email Verification (simulate)
    // Navigate to dashboard after successful registration
    await page.evaluate(() => {
        showPage('dashboard-page');
    });
    await expect(page.locator('[data-testid="verification-success"]')).toBeVisible();
    
    // 3. Complete Profile
    await page.fill('[data-testid="full-name"]', 'Jane Smith');
    await page.fill('[data-testid="phone"]', '+1234567890');
    await page.click('[data-testid="save-profile"]');
    
    // 4. Browse Products
    await page.evaluate(() => {
        showPage('products-page');
    });
    await expect(page.locator('[data-testid="product-grid"]')).toBeVisible();
    
    // Search for a product
    await page.fill('[data-testid="search-input"]', 'laptop');
    await page.click('[data-testid="search-button"]');
    
    // 5. Add to Cart
    await page.click('[data-testid="product-item-1"]');
    await page.click('[data-testid="add-to-cart"]');
    await expect(page.locator('[data-testid="cart-count"]')).toContainText('1');
    
    // 6. Checkout Process
    await page.evaluate(() => {
        showPage('cart-page');
    });
    await page.click('[data-testid="proceed-checkout"]');
    
    // Fill shipping information
    await page.fill('[data-testid="shipping-name"]', 'Jane Smith');
    await page.fill('[data-testid="shipping-address"]', '123 Main St');
    await page.fill('[data-testid="shipping-city"]', 'New York');
    await page.fill('[data-testid="shipping-zip"]', '10001');
    await page.click('[data-testid="save-shipping"]');
    
    // 7. Payment (simulate)
    await page.fill('[data-testid="card-number"]', '4111111111111111');
    await page.fill('[data-testid="card-expiry"]', '12/25');
    await page.fill('[data-testid="card-cvv"]', '123');
    await page.click('[data-testid="pay-button"]');
    
    // 8. Order Confirmation
    await expect(page.locator('[data-testid="order-success"]')).toBeVisible();
    await expect(page.locator('[data-testid="order-number"]')).toBeVisible();
    
    // 9. View Order History
    await page.evaluate(() => {
        showPage('orders-page');
    });
    await expect(page.locator('[data-testid="order-history"]')).toBeVisible();
  });

  test('User can recover from failed payment', async ({ page }) => {
    // Start checkout process
    await page.goto('http://localhost:8000');
    await page.evaluate(() => {
        showPage('cart-page');
    });
    await page.click('[data-testid="proceed-checkout"]');
    
    // Fill payment with invalid card
    await page.fill('[data-testid="card-number"]', '4000000000000002'); // Declined card
    await page.fill('[data-testid="card-expiry"]', '12/25');
    await page.fill('[data-testid="card-cvv"]', '123');
    await page.click('[data-testid="pay-button"]');
    
    // Verify payment failure
    await expect(page.locator('[data-testid="payment-error"]')).toBeVisible();
    await expect(page.locator('[data-testid="payment-error"]')).toContainText('Payment failed');
    
    // Try again with valid card
    await page.fill('[data-testid="card-number"]', '4111111111111111');
    await page.click('[data-testid="pay-button"]');
    
    // Verify success
    await expect(page.locator('[data-testid="order-success"]')).toBeVisible();
  });
}); 