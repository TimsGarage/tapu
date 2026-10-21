<script>
	import { onMount } from "svelte";
	import { api, initializeSession } from "./api.js";

	let users = [];
	let services = [];
	let clients = [];
	let error = "";
	let notice = "";
	let clientSecret = "";

	let newUser = {
		username: "",
		password: "",
		isAdmin: false,
	};

	let newService = {
		key: "",
		name: "",
		audience: "",
	};

	let newClient = {
		name: "",
		clientType: "confidential",
		redirectURI: "",
		serviceId: "",
	};

	onMount(load);

	async function load() {
		error = "";

		try {
			await initializeSession();

			[users, services, clients] = await Promise.all([
				api("/users"),
				api("/services"),
				api("/clients"),
			]);

			for (const user of users) {
				user.serviceIds = (await api(`/users/${user.id}/services`)).map(
					(service) => service.id,
				);
			}
		} catch (caught) {
			error = caught.message;
		}
	}

	async function createUser() {
		await run(async () => {
			await api("/users", {
				method: "POST",
				body: newUser,
			});

			newUser = {
				username: "",
				password: "",
				isAdmin: false,
			};

			notice = "User created.";
			await load();
		});
	}

	async function setUserState(user) {
		await run(async () => {
			await api(`/users/${user.id}`, {
				method: "PATCH",
				body: {
					isAdmin: user.isAdmin,
					status: user.status,
				},
			});

			notice = "User updated.";
			await load();
		});
	}

	async function resetPassword(user) {
		const password = window.prompt(`New password for ${user.username}:`);

		if (!password) {
			return;
		}

		await run(async () => {
			await api(`/users/${user.id}/password`, {
				method: "POST",
				body: { password },
			});

			notice = "Password reset and existing sessions revoked.";
		});
	}

	async function toggleUserService(user, service) {
		let hasAccess = false;
		if (user.serviceIds) {
			hasAccess = user.serviceIds.includes(service.id);
		}

		await run(async () => {
			await api(`/users/${user.id}/services/${service.id}`, {
				method: hasAccess ? "DELETE" : "PUT",
			});

			notice = hasAccess ? "Access revoked." : "Access granted.";
			await load();
		});
	}

	async function createService() {
		await run(async () => {
			await api("/services", {
				method: "POST",
				body: newService,
			});

			newService = {
				key: "",
				name: "",
				audience: "",
			};

			notice = "Service created.";
			await load();
		});
	}

	async function toggleService(service) {
		await run(async () => {
			await api(`/services/${service.id}`, {
				method: "PATCH",
				body: { enabled: !service.enabled },
			});

			notice = service.enabled ? "Service disabled." : "Service enabled.";

			await load();
		});
	}

	async function createClient() {
		await run(async () => {
			const body = {
				name: newClient.name,
				clientType: newClient.clientType,
				redirectUris: [newClient.redirectURI],
				serviceIds: newClient.serviceId ? [newClient.serviceId] : [],
			};

			const created = await api("/clients", {
				method: "POST",
				body,
			});

			clientSecret = created.clientSecret || "";

			newClient = {
				name: "",
				clientType: "confidential",
				redirectURI: "",
				serviceId: "",
			};

			notice = "OAuth client created.";
			await load();
		});
	}

	async function toggleClient(client) {
		await run(async () => {
			await api(`/clients/${client.id}`, {
				method: "PATCH",
				body: { enabled: !client.enabled },
			});

			notice = client.enabled ? "Client disabled." : "Client enabled.";

			await load();
		});
	}

	async function rotateSecret(client) {
		if (!window.confirm(`Rotate the secret for ${client.name}?`)) {
			return;
		}

		await run(async () => {
			const result = await api(`/clients/${client.id}/rotate-secret`, {
				method: "POST",
				body: {},
			});

			clientSecret = result.clientSecret;
			notice = "Client secret rotated.";
		});
	}

	async function run(operation) {
		error = "";
		notice = "";
		clientSecret = "";

		try {
			await operation();
		} catch (caught) {
			error = caught.message;
		}
	}
</script>

<svelte:head>
	<title>IAM administration</title>
</svelte:head>

<header>
	<div>
		<h1>IAM administration</h1>
		<p>Manage users, services, access, and OAuth clients.</p>
	</div>
	<a class="button" href="/account">Account and sign out</a>
</header>

<main>
	{#if error}
		<p class="message error" role="alert">{error}</p>
	{/if}

	{#if notice}
		<p class="message success">{notice}</p>
	{/if}

	{#if clientSecret}
		<section class="secret">
			<h2>Save this client secret now</h2>
			<p>It will not be shown again.</p>
			<code>{clientSecret}</code>
		</section>
	{/if}

	<section>
		<h2>Users</h2>

		<form class="grid" on:submit|preventDefault={createUser}>
			<input
				bind:value={newUser.username}
				placeholder="Username"
				autocomplete="off"
				required
			/>
			<input
				bind:value={newUser.password}
				type="password"
				placeholder="Initial password"
				autocomplete="new-password"
				minlength="12"
				required
			/>
			<label>
				<input bind:checked={newUser.isAdmin} type="checkbox" />
				Administrator
			</label>
			<button type="submit">Create user</button>
		</form>

		{#each users as user}
			<article>
				<div>
					<strong>{user.username}</strong>
					<small>{user.id}</small>
				</div>

				<label>
					<input bind:checked={user.isAdmin} type="checkbox" />
					Administrator
				</label>

				<select bind:value={user.status}>
					<option value="active">Active</option>
					<option value="disabled">Disabled</option>
				</select>

				<button on:click={() => setUserState(user)}> Save </button>

				<button on:click={() => resetPassword(user)}>
					Reset password
				</button>

				<details>
					<summary>Service access</summary>

					{#each services as service}
						<label>
							<input
								type="checkbox"
								checked={user.serviceIds?.includes(service.id)}
								on:change={() =>
									toggleUserService(user, service)}
							/>
							{service.name}
						</label>
					{/each}
				</details>
			</article>
		{/each}
	</section>

	<section>
		<h2>Services</h2>

		<form class="grid" on:submit|preventDefault={createService}>
			<input
				bind:value={newService.key}
				placeholder="project-a"
				required
			/>
			<input
				bind:value={newService.name}
				placeholder="Project A"
				required
			/>
			<input
				bind:value={newService.audience}
				placeholder="project-a-api"
				required
			/>
			<button type="submit">Create service</button>
		</form>

		{#each services as service}
			<article>
				<div>
					<strong>{service.name}</strong>
					<small>{service.key} · {service.audience}</small>
				</div>

				<span class:disabled={!service.enabled}>
					{service.enabled ? "Enabled" : "Disabled"}
				</span>

				<button on:click={() => toggleService(service)}>
					{service.enabled ? "Disable" : "Enable"}
				</button>
			</article>
		{/each}
	</section>

	<section>
		<h2>OAuth clients</h2>

		<form class="grid" on:submit|preventDefault={createClient}>
			<input
				bind:value={newClient.name}
				placeholder="Project A web"
				required
			/>

			<select bind:value={newClient.clientType}>
				<option value="confidential">Confidential</option>
				<option value="public">Public</option>
			</select>

			<input
				bind:value={newClient.redirectURI}
				placeholder="https://project-a.example.com/auth/callback"
				required
			/>

			<select bind:value={newClient.serviceId}>
				<option value="">No service</option>
				{#each services as service}
					<option value={service.id}>{service.name}</option>
				{/each}
			</select>

			<button type="submit">Create client</button>
		</form>

		{#each clients as client}
			<article>
				<div>
					<strong>{client.name}</strong>
					<small>{client.clientId}</small>
					<small>{client.clientType}</small>
				</div>

				<span class:disabled={!client.enabled}>
					{client.enabled ? "Enabled" : "Disabled"}
				</span>

				<button on:click={() => toggleClient(client)}>
					{client.enabled ? "Disable" : "Enable"}
				</button>

				{#if client.clientType === "confidential"}
					<button on:click={() => rotateSecret(client)}>
						Rotate secret
					</button>
				{/if}

				<details>
					<summary>Configuration</summary>

					<strong>Redirect URIs</strong>
					<ul>
						{#each client.redirectUris as redirectURI}
							<li><code>{redirectURI}</code></li>
						{/each}
					</ul>
				</details>
			</article>
		{/each}
	</section>
</main>
