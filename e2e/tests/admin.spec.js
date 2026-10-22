import { test, expect } from "@playwright/test";
import fs from "node:fs";

const generated = JSON.parse(fs.readFileSync(".generated.json", "utf8"));

async function login(page, username, password) {
	await page.goto("/login");

	await page.getByLabel("Username").fill(username);

	await page.getByLabel("Password").fill(password);

	await page
		.getByRole("button", {
			name: "Sign in",
		})
		.click();
}

test.describe("IAM administration", () => {
	test("administrator can access the admin page", async ({ page }) => {
		await login(page, generated.adminUsername, generated.adminPassword);

		await expect(page).toHaveURL(/\/account$/);

		await page.goto("/admin");

		await expect(
			page.getByRole("heading", {
				name: "IAM administration",
			}),
		).toBeVisible();
	});

	test("normal user cannot access the admin page", async ({ page }) => {
		await login(page, generated.username, generated.password);

		await expect(page).toHaveURL(/\/account$/);

		const response = await page.goto("/admin");

		expect(response).not.toBeNull();
		expect(response.status()).toBe(403);

		await expect(
			page.getByText("Forbidden", {
				exact: true,
			}),
		).toBeVisible();
	});
});
