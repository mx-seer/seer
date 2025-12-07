import { test, expect } from '@playwright/test';

test.describe('Dashboard', () => {
	test('should load the dashboard page', async ({ page }) => {
		await page.goto('/');
		// Dashboard has navbar with Seer branding
		await expect(page.getByRole('link', { name: 'Seer' })).toBeVisible();
		// Dashboard has Filters card
		await expect(page.getByText('Filters')).toBeVisible();
	});

	test('should display stats cards', async ({ page }) => {
		await page.goto('/');
		// Wait for stats to load
		await expect(page.getByText('Total Opportunities')).toBeVisible();
		await expect(page.locator('.stat')).toHaveCount(4);
	});

	test('should have working navigation', async ({ page }) => {
		await page.goto('/');

		// Navigate to Sources
		await page.click('a[href="/sources"]');
		await page.waitForURL('/sources');
		await expect(page.getByText('Sources').first()).toBeVisible();

		// Navigate to Reports
		await page.click('a[href="/reports"]');
		await page.waitForURL('/reports');
		await expect(page.getByText('Reports').first()).toBeVisible();

		// Navigate back to Dashboard
		await page.click('a[href="/"]');
		await page.waitForURL('/');
		await expect(page.getByText('Filters')).toBeVisible();
	});
});

test.describe('Sources', () => {
	test('should load the sources page', async ({ page }) => {
		await page.goto('/sources');
		await expect(page.locator('h1:has-text("Sources")')).toBeVisible();
	});

	test('should display add source button', async ({ page }) => {
		await page.goto('/sources');
		await expect(page.getByRole('button', { name: 'Add Source' })).toBeVisible();
	});

	test('should open add source modal', async ({ page }) => {
		await page.goto('/sources');
		await page.getByRole('button', { name: 'Add Source' }).click();
		await expect(page.locator('.modal')).toBeVisible();
		await expect(page.getByText('Add New Source')).toBeVisible();
	});
});

test.describe('Reports', () => {
	test('should load the reports page', async ({ page }) => {
		await page.goto('/reports');
		await expect(page.locator('h1:has-text("Reports")')).toBeVisible();
	});

	test('should display generate button', async ({ page }) => {
		await page.goto('/reports');
		await expect(page.getByRole('button', { name: 'Generate Report' })).toBeVisible();
	});
});

test.describe('API Health', () => {
	test('should return healthy status', async ({ request }) => {
		const response = await request.get('/api/health');
		expect(response.status()).toBe(200);

		const data = await response.json();
		expect(data.status).toBe('ok');
	});
});
