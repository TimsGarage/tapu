import { browser } from '$app/env';
import { goto } from '$app/navigation';

import { config } from './config.js';
import { getUserManager } from './oidc.js';

class AuthenticationState {
	user = $state(null);
	loading = $state(true);
	initialized = $state(false);
	error = $state('');

	authenticated = $derived(Boolean(this.user && !this.user.expired));

	username = $derived(this.user?.profile?.preferred_username || this.user?.profile?.sub || '');

	subject = $derived(this.user?.profile?.sub || '');

	expiresAt = $derived(this.user?.expires_at ? new Date(this.user.expires_at * 1000) : null);

	async initialize() {
		if (!browser || this.initialized) {
			return;
		}

		this.loading = true;
		this.error = '';

		try {
			const manager = getUserManager();
			this.user = await manager.getUser();

			if (this.user?.expired) {
				await manager.removeUser();
				this.user = null;
			}

			this.initialized = true;
		} catch (error) {
			this.error = messageFrom(error);
		} finally {
			this.loading = false;
		}
	}

	async login(returnTo = currentLocalPath()) {
		this.error = '';

		try {
			await getUserManager().signinRedirect({
				state: {
					returnTo: validReturnTo(returnTo) ? returnTo : '/'
				}
			});
		} catch (error) {
			this.error = messageFrom(error);
		}
	}

	async completeLogin() {
		this.loading = true;
		this.error = '';

		try {
			const user = await getUserManager().signinRedirectCallback();

			this.user = user;
			this.initialized = true;

			const returnTo = validReturnTo(user.state?.returnTo) ? user.state.returnTo : '/';

			await goto(returnTo, {
				replaceState: true
			});
		} catch (error) {
			this.error = messageFrom(error);
			throw error;
		} finally {
			this.loading = false;
		}
	}

	async logout() {
		this.loading = true;
		this.error = '';

		try {
			await getUserManager().removeUser();
			this.user = null;

			await goto('/signed-out', {
				replaceState: true
			});
		} catch (error) {
			this.error = messageFrom(error);
		} finally {
			this.loading = false;
		}
	}

	async loadUserInfo() {
		const accessToken = this.user?.access_token;

		if (!accessToken) {
			throw new Error('No access token is available.');
		}

		const response = await fetch(`${config.issuerURL}/oauth/userinfo`, {
			headers: {
				Accept: 'application/json',
				Authorization: `Bearer ${accessToken}`
			}
		});

		if (!response.ok) {
			throw new Error(`UserInfo failed with HTTP ${response.status}`);
		}

		return response.json();
	}
}

function currentLocalPath() {
	if (!browser) {
		return '/';
	}

	return window.location.pathname + window.location.search;
}

function validReturnTo(value) {
	if (!browser || typeof value !== 'string' || !value.startsWith('/') || value.startsWith('//')) {
		return false;
	}

	try {
		const parsed = new URL(value, window.location.origin);

		return parsed.origin === window.location.origin;
	} catch {
		return false;
	}
}

function messageFrom(error) {
	return error instanceof Error ? error.message : 'An unexpected authentication error occurred.';
}

export const auth = new AuthenticationState();
