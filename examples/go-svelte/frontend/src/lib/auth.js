import { writable } from "svelte/store";

const configuredAPIOrigin = import.meta.env.VITE_API_ORIGIN || "";

const apiOrigin = configuredAPIOrigin.replace(/\/$/, "");

function apiURL(path) {
	return `${apiOrigin}${path}`;
}

export const auth = writable({
	loading: true,
	authenticated: false,
	user: null,
	csrfToken: "",
});

export async function loadAuthentication() {
	const response = await fetch(apiURL("/api/auth/me"), {
		credentials: "include",
		headers: {
			Accept: "application/json",
		},
	});

	if (response.status === 401) {
		auth.set({
			loading: false,
			authenticated: false,
			user: null,
			csrfToken: "",
		});

		return;
	}

	if (!response.ok) {
		throw new Error("Unable to load authentication state.");
	}

	const result = await response.json();

	auth.set({
		loading: false,
		authenticated: result.authenticated,
		user: result.user,
		csrfToken: result.csrfToken,
	});
}

export function login() {
	window.location.assign(apiURL("/auth/login"));
}

export async function logout(csrfToken) {
	const response = await fetch(apiURL("/api/auth/logout"), {
		method: "POST",
		credentials: "include",
		headers: {
			"X-CSRF-Token": csrfToken,
		},
	});

	if (!response.ok && response.status !== 401) {
		throw new Error("Unable to sign out.");
	}

	auth.set({
		loading: false,
		authenticated: false,
		user: null,
		csrfToken: "",
	});

	window.location.assign("/");
}
