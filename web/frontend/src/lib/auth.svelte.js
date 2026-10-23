import { goto } from '$app/navigation';
import { SvelteURLSearchParams } from 'svelte/reactivity';
import { apiRequest } from './public-api';

const UserState = $state({
	user: undefined,
	csrfToken: undefined,
	error: '',
	loading: true
});

export const getUserState = async () => {
	if (!UserState.user) {
		await loadAccount();
	}
	return UserState;
};

const loadAccount = async () => {
	try {
		const result = await apiRequest('/api/web/account');

		UserState.user = result.user;
		UserState.csrfToken = result.csrfToken;
		UserState.error = '';
	} catch (caught) {
		if (caught && typeof caught === 'object' && caught.status === 401) {
			goto('/login');
			return;
		}

		UserState.user = undefined;
		UserState.csrfToken = undefined;
		UserState.error = caught instanceof Error ? caught.message : 'Unable to load the account.';
	} finally {
		UserState.loading = false;
	}
};

export const Login = async () => {
	try {
		const query = new SvelteURLSearchParams(window.location.search);

		const requestedReturnTo = query.get('return_to') || '';

		const suffix = requestedReturnTo ? `?return_to=${encodeURIComponent(requestedReturnTo)}` : '';

		const result = await apiRequest(`/api/web/login${suffix}`);

		if (result.authenticated) {
			window.location.replace(result.redirectTo);
			return;
		}

		let csrfToken = result.csrfToken;
		let returnTo = result.returnTo || '';
		return { csrfToken, returnTo, error: '' };
	} catch (error) {
		let loadError = error instanceof Error ? error.message : 'Unable to prepare the login form.';
		return { csrfToken: '', returnTo: '', error: loadError };
	}
};
