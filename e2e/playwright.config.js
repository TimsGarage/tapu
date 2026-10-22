import { defineConfig, devices } from "@playwright/test";
import fs from "node:fs";

const generated = JSON.parse(fs.readFileSync(".generated.json", "utf8"));

export default defineConfig({
	testDir: "./tests",
	timeout: 30_000,

	expect: {
		timeout: 5_000,
	},

	fullyParallel: false,
	workers: 1,
	retries: process.env.CI ? 2 : 0,

	reporter: [
		["list"],
		[
			"html",
			{
				outputFolder: "playwright-report",
			},
		],
	],

	use: {
		baseURL: generated.iamURL,
		trace: "retain-on-failure",
		screenshot: "only-on-failure",
		video: "retain-on-failure",
	},

	webServer: [
		{
			command: "node receiver/server.mjs",
			cwd: ".",
			env: {
				...process.env,
				E2E_IAM_URL: generated.iamURL,
				E2E_RECEIVER_URL: generated.receiverURL,
				E2E_CLIENT_ID: generated.clientID,
				E2E_AUDIENCE: generated.audience,
			},
			url: generated.receiverURL,
			reuseExistingServer: !process.env.CI,
			timeout: 30_000,
		},
		{
			command: "go run ../examples/go-svelte/cmd/server",
			cwd: ".",
			env: {
				...process.env,
				IAM_ISSUER_URL: generated.iamURL,
				IAM_CLIENT_ID: generated.apiClientID,
				IAM_CLIENT_SECRET: generated.apiClientSecret,
				IAM_REDIRECT_URL: generated.apiRedirectURI,
				IAM_AUDIENCE: generated.audience,
				APP_FRONTEND_URL: "http://127.0.0.1:5173",
				APP_SECURE_COOKIES: "false",
			},
			url: "http://127.0.0.1:3000",
			reuseExistingServer: !process.env.CI,
			timeout: 30_000,
		},
	],

	projects: [
		{
			name: "chromium",
			use: {
				...devices["Desktop Chrome"],
			},
		},
	],
});
