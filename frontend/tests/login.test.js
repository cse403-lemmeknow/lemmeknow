import { expect, test } from '@playwright/test';

test.beforeEach(async ({ page }) => {
	await page.route('**/api/user/', (route) => {
		route.fulfill({
			status: 200,
			contentType: 'application/json',
			body: JSON.stringify({ userId: 1234, name: 'Alex', groups: [1234] })
		});
	});
});

test('Page has header and message', async ({ page }) => {
	await page.goto('http://localhost:5173/');

	expect(await page.textContent('h1')).toBe('LemmeKnow');
	expect(await page.textContent('h3')).toContain(
		'SCHEDULE HANGOUTS, PLAN ROAD-TRIPS, SHARE CALENDARS, EVERYTHING, EVERYWHERE, ALL AT ONCE.'
	);
});
