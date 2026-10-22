import { browser } from '$app/env';

import { UserManager, WebStorageStateStore } from 'oidc-client-ts';

import { callbackURL, config } from './config.js';

let userManager;

export function getUserManager() {
	if (!browser) {
		throw new Error('The OIDC client is available only in the browser.');
	}

	if (!userManager) {
		userManager = new UserManager({
			authority: config.issuerURL,
			client_id: config.clientID,
			redirect_uri: callbackURL,

			response_type: 'code',
			response_mode: 'query',
			scope: 'openid profile',

			loadUserInfo: true,
			automaticSilentRenew: false,
			monitorSession: false,

			userStore: new WebStorageStateStore({
				store: window.sessionStorage
			}),

			stateStore: new WebStorageStateStore({
				store: window.sessionStorage
			}),

			extraQueryParams: {
				audience: config.audience
			}
		});
	}

	return userManager;
}
