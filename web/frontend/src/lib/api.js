let csrfToken = '';

export async function initializeSession() {
	const response = await fetch('/api/admin/session', {
		credentials: 'same-origin'
	});

	if (!response.ok) {
		throw new Error('Unable to load the administrative session.');
	}

	const session = await response.json();
	csrfToken = session.csrfToken;

	return session;
}

export async function api(path, options = {}) {
	const headers = new Headers(options.headers || {});

	if (options.body !== undefined) {
		headers.set('Content-Type', 'application/json');
	}

	if (!['GET', 'HEAD'].includes(options.method || 'GET')) {
		headers.set('X-CSRF-Token', csrfToken);
	}

	const response = await fetch(`/api/admin${path}`, {
		...options,
		headers,
		credentials: 'same-origin',
		body: options.body === undefined ? undefined : JSON.stringify(options.body)
	});

	if (response.status === 204) {
		return null;
	}

	const value = await response.json().catch(() => ({}));

	if (!response.ok) {
		throw new Error(value.error || 'Request failed.');
	}

	return value;
}
