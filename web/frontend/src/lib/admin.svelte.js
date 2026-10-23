import { writable } from 'svelte/store';
import { api, initializeSession } from './api';

export const users = writable([]);
export const services = writable([]);
export const clients = writable([]);
export const error = writable('');
export const notice = writable('');
export const clientSecret = writable('');

export const newUser = $state({
	username: '',
	password: '',
	isAdmin: false
});

export const newService = $state({
	key: '',
	name: '',
	audience: ''
});

export const newClient = $state({
	name: '',
	clientType: 'confidential',
	redirectURI: '',
	serviceId: ''
});

export async function loadAdminPage() {
	error.set('');

	try {
		await initializeSession();

		let [_users, _services, _clients] = await Promise.all([
			api('/users'),
			api('/services'),
			api('/clients')
		]);
		for (const _user of _users) {
			_user.serviceIds = (await api(`/users/${_user.id}/services`)).map((service) => service.id);
		}
		users.set(_users);
		services.set(_services);
		clients.set(_clients);
	} catch (caught) {
		error.set(caught.message);
	}
}

export async function createUser() {
	await run(async () => {
		await api('/users', {
			method: 'POST',
			body: newUser
		});

		newUser.username = '';
		newUser.password = '';
		newUser.isAdmin = false;

		notice.set('User created.');
		await loadAdminPage();
	});
}

export async function setUserState(user) {
	await run(async () => {
		await api(`/users/${user.id}`, {
			method: 'PATCH',
			body: {
				isAdmin: user.isAdmin,
				status: user.status
			}
		});

		notice.set('User updated.');
		await loadAdminPage();
	});
}

export async function resetPassword(user) {
	const password = window.prompt(`New password for ${user.username}:`);

	if (!password) {
		return;
	}

	await run(async () => {
		await api(`/users/${user.id}/password`, {
			method: 'POST',
			body: { password }
		});

		notice.set('Password reset and existing sessions revoked.');
	});
}

export async function toggleUserService(user, service) {
	let hasAccess = false;
	if (user.serviceIds) {
		hasAccess = user.serviceIds.includes(service.id);
	}

	await run(async () => {
		await api(`/users/${user.id}/services/${service.id}`, {
			method: hasAccess ? 'DELETE' : 'PUT'
		});

		notice.set(hasAccess ? 'Access revoked.' : 'Access granted.');
		await loadAdminPage();
	});
}

export async function createService() {
	await run(async () => {
		await api('/services', {
			method: 'POST',
			body: newService
		});

		newService.audience = '';
		newService.key = '';
		newService.name = '';

		notice.set('Service created.');
		await loadAdminPage();
	});
}

export async function toggleService(service) {
	await run(async () => {
		await api(`/services/${service.id}`, {
			method: 'PATCH',
			body: { enabled: !service.enabled }
		});

		notice.set(service.enabled ? 'Service disabled.' : 'Service enabled.');

		await loadAdminPage();
	});
}

export async function createClient() {
	await run(async () => {
		const body = {
			name: newClient.name,
			clientType: newClient.clientType,
			redirectUris: [newClient.redirectURI],
			serviceIds: newClient.serviceId ? [newClient.serviceId] : []
		};

		const created = await api('/clients', {
			method: 'POST',
			body
		});

		clientSecret.set(created.clientSecret || '');

		newClient.name = '';
		newClient.clientType = 'confidential';
		newClient.redirectURI = '';
		newClient.serviceId = '';

		notice.set('OAuth client created.');
		await loadAdminPage();
	});
}

export async function toggleClient(client) {
	await run(async () => {
		await api(`/clients/${client.id}`, {
			method: 'PATCH',
			body: { enabled: !client.enabled }
		});

		notice.set(client.enabled ? 'Client disabled.' : 'Client enabled.');

		await loadAdminPage();
	});
}

export async function rotateSecret(client) {
	if (!window.confirm(`Rotate the secret for ${client.name}?`)) {
		return;
	}

	await run(async () => {
		const result = await api(`/clients/${client.id}/rotate-secret`, {
			method: 'POST',
			body: {}
		});

		clientSecret.set(result.clientSecret);
		notice.set('Client secret rotated.');
	});
}

async function run(operation) {
	error.set('');
	notice.set('');
	clientSecret.set('');

	try {
		await operation();
	} catch (caught) {
		error.set(caught.message);
	}
}
