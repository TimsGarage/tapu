import { env } from '$env/dynamic/public';

function requiredURL(name, value) {
	const normalized = value?.trim().replace(/\/$/, '');

	if (!normalized) {
		throw new Error(`Missing required environment variable: ${name}`);
	}

	let parsed;

	try {
		parsed = new URL(normalized);
	} catch {
		throw new Error(`${name} must be a valid URL`);
	}

	if (parsed.protocol !== 'http:' && parsed.protocol !== 'https:') {
		throw new Error(`${name} must use HTTP or HTTPS`);
	}

	return normalized;
}

function required(name, value) {
	const normalized = value?.trim();

	if (!normalized) {
		throw new Error(`Missing required environment variable: ${name}`);
	}

	return normalized;
}

export const config = Object.freeze({
	issuerURL: requiredURL('PUBLIC_IAM_ISSUER_URL', env.PUBLIC_IAM_ISSUER_URL),

	clientID: required('PUBLIC_IAM_CLIENT_ID', env.PUBLIC_IAM_CLIENT_ID),

	audience: required('PUBLIC_IAM_AUDIENCE', env.PUBLIC_IAM_AUDIENCE),

	appOrigin: requiredURL('PUBLIC_APP_ORIGIN', env.PUBLIC_APP_ORIGIN)
});

export const callbackURL = `${config.appOrigin}/auth/callback`;
