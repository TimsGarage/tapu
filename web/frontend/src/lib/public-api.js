export class APIError extends Error {
	constructor(status, message) {
		super(message);
		this.name = 'APIError';
		this.status = status;
	}
}

export async function apiRequest(path, options = {}) {
	const headers = new Headers(options.headers || {});

	headers.set('Accept', 'application/json');

	const response = await fetch(path, {
		...options,
		credentials: 'same-origin',
		headers
	});

	const body = await response.json().catch(() => ({}));

	if (!response.ok) {
		throw new APIError(response.status, body.error || 'Request failed.');
	}

	return body;
}
