import { defineEnvVars } from '@sveltejs/kit/env';

export const variables = defineEnvVars({
	PUBLIC_IAM_ISSUER_URL: {
		public: true
	},
	PUBLIC_IAM_CLIENT_ID: {
		public: true
	},
	PUBLIC_IAM_AUDIENCE: {
		public: true
	},
	PUBLIC_APP_ORIGIN: {
		public: true
	}
});
