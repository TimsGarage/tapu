import { test, expect } from "@playwright/test";

test("API rejects an unregistered origin", async ({ request }) => {
	const response = await request.get("http://127.0.0.1:3000/api/auth/me", {
		headers: {
			Origin: "https://evil.example.net",
		},
	});

	expect(response.status()).toBe(403);
});

test("API handles credentialed preflight", async ({ request }) => {
	const response = await request.fetch(
		"http://127.0.0.1:3000/api/auth/logout",
		{
			method: "OPTIONS",
			headers: {
				Origin: "http://127.0.0.1:5173",
				"Access-Control-Request-Method": "POST",
				"Access-Control-Request-Headers": "X-CSRF-Token",
			},
		},
	);

	expect(response.status()).toBe(204);
	expect(response.headers()["access-control-allow-credentials"]).toBe("true");
});
