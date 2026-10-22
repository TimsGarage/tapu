import { test, expect } from "@playwright/test";
import fs from "node:fs";

const generated = JSON.parse(fs.readFileSync(".generated.json", "utf8"));

async function loginToIAM(page) {
	await page.goto("/login");

	await page.getByLabel("Username").fill(generated.username);

	await page.getByLabel("Password").fill(generated.password);

	await page
		.getByRole("button", {
			name: "Sign in",
		})
		.click();

	const alert = page.getByRole("alert");

	if (await alert.isVisible().catch(() => false)) {
		throw new Error(`IAM login failed: ${await alert.textContent()}`);
	}

	await expect(page).toHaveURL(/\/account$/);
}

test.describe("OAuth and SSO", () => {
	test("completes Authorization Code with PKCE", async ({ page }) => {
		await loginToIAM(page);

		await page.goto("http://127.0.0.1:4100/login");

		await expect(page.getByText("Authentication successful")).toBeVisible();

		await expect(page.locator("#authenticated")).toHaveText("true");

		await expect(page.locator("#access-token")).not.toHaveText("");

		await expect(page.locator("#id-token")).not.toHaveText("");

		await expect(page.locator("#scope")).toHaveText("openid profile");
	});

	test("uses IAM SSO without another password prompt", async ({ page }) => {
		await loginToIAM(page);

		await page.goto("http://127.0.0.1:4100/login");

		await expect(page.getByText("Authentication successful")).toBeVisible();

		await expect(page).not.toHaveURL(/127\.0\.0\.1:8080\/login/);
	});

	test("rejects a mismatched PKCE verifier", async ({ page, context }) => {
		await loginToIAM(page);

		const authorizationURL = new URL("/oauth/authorize", generated.iamURL);

		authorizationURL.search = new URLSearchParams({
			response_type: "code",
			client_id: generated.clientID,
			redirect_uri: generated.redirectURI,
			scope: "openid profile",
			audience: generated.audience,
			state: "test-state",
			nonce: "test-nonce",
			code_challenge: "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
			code_challenge_method: "S256",
		});

		await page.goto(authorizationURL.toString());

		await expect(page).toHaveURL(/127\.0\.0\.1:4100\/callback/);

		const callbackURL = new URL(page.url());
		const code = callbackURL.searchParams.get("code");

		expect(code).toBeTruthy();

		const response = await context.request.post(
			new URL("/oauth/token", generated.iamURL).toString(),
			{
				form: {
					grant_type: "authorization_code",
					client_id: generated.clientID,
					code,
					redirect_uri: generated.redirectURI,
					code_verifier:
						"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				},
			},
		);

		expect(response.status()).toBe(400);

		const body = await response.json();
		expect(body.error).toBe("invalid_grant");
	});
});
