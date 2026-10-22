import { test, expect } from "@playwright/test";
import fs from "node:fs";

const generated = JSON.parse(fs.readFileSync(".generated.json", "utf8"));

test.describe("IAM authentication", () => {
	test("rejects invalid credentials", async ({ page }) => {
		await page.goto("/login");

		await page.getByLabel("Username").fill(generated.username);
		await page.getByLabel("Password").fill("definitely-the-wrong-password");

		await page
			.getByRole("button", {
				name: "Sign in",
			})
			.click();

		await expect(page.getByRole("alert")).toContainText(
			"Invalid username or password.",
		);
	});

	test("logs in and logs out", async ({ page }) => {
		await page.goto("/login");

		await page.getByLabel("Username").fill(generated.username);
		await page.getByLabel("Password").fill(generated.password);

		await page
			.getByRole("button", {
				name: "Sign in",
			})
			.click();

		await expect(page).toHaveURL(/\/account$/);
		await expect(page.getByText(generated.username)).toBeVisible();

		await page
			.getByRole("button", {
				name: "Sign out",
			})
			.click();

		await expect(page).toHaveURL(/\/login$/);
	});

	test("redirects unauthenticated users from account", async ({ page }) => {
		await page.goto("/account");

		await expect(page).toHaveURL(/\/login/);
	});
});
