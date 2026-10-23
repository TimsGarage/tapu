<script>
	import { onMount } from 'svelte';
	import {
		users,
		services,
		clients,
		error,
		notice,
		clientSecret,
		newUser,
		newService,
		newClient,
		loadAdminPage,
		createUser,
		setUserState,
		resetPassword,
		toggleUserService,
		createService,
		toggleService,
		createClient,
		toggleClient,
		rotateSecret
	} from '../../lib/admin.svelte';

	onMount(loadAdminPage);
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
	{#if $error}
		<p class="message error" role="alert">{$error}</p>
	{/if}

	{#if $notice}
		<p class="message success">{$notice}</p>
	{/if}

	{#if $clientSecret}
		<section class="secret">
			<h2>Save this client secret now</h2>
			<p>It will not be shown again.</p>
			<code>{$clientSecret}</code>
		</section>
	{/if}

	<section>
		<h2>Users</h2>

		<form class="grid" on:submit|preventDefault={createUser}>
			<input bind:value={newUser.username} placeholder="Username" autocomplete="off" required />
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

		{#each $users as user (user.id)}
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

				<button on:click={() => resetPassword(user)}> Reset password </button>

				<details>
					<summary>Service access</summary>

					{#each $services as service (service.id)}
						<label>
							<input
								type="checkbox"
								checked={user.serviceIds?.includes(service.id)}
								on:change={() => toggleUserService(user, service)}
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
			<input bind:value={newService.key} placeholder="project-a" required />
			<input bind:value={newService.name} placeholder="Project A" required />
			<input bind:value={newService.audience} placeholder="project-a-api" required />
			<button type="submit">Create service</button>
		</form>

		{#each $services as service (service.id)}
			<article>
				<div>
					<strong>{service.name}</strong>
					<small>{service.key} · {service.audience}</small>
				</div>

				<span class:disabled={!service.enabled}>
					{service.enabled ? 'Enabled' : 'Disabled'}
				</span>

				<button on:click={() => toggleService(service)}>
					{service.enabled ? 'Disable' : 'Enable'}
				</button>
			</article>
		{/each}
	</section>

	<section>
		<h2>OAuth clients</h2>

		<form class="grid" on:submit|preventDefault={createClient}>
			<input bind:value={newClient.name} placeholder="Project A web" required />

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
				{#each $services as service (service.id)}
					<option value={service.id}>{service.name}</option>
				{/each}
			</select>

			<button type="submit">Create client</button>
		</form>

		{#each $clients as client (client.id)}
			<article>
				<div>
					<strong>{client.name}</strong>
					<small>{client.clientId}</small>
					<small>{client.clientType}</small>
				</div>

				<span class:disabled={!client.enabled}>
					{client.enabled ? 'Enabled' : 'Disabled'}
				</span>

				<button on:click={() => toggleClient(client)}>
					{client.enabled ? 'Disable' : 'Enable'}
				</button>

				{#if client.clientType === 'confidential'}
					<button on:click={() => rotateSecret(client)}> Rotate secret </button>
				{/if}

				<details>
					<summary>Configuration</summary>

					<strong>Redirect URIs</strong>
					<ul>
						{#each client.redirectUris as redirectURI (redirectURI)}
							<li><code>{redirectURI}</code></li>
						{/each}
					</ul>
				</details>
			</article>
		{/each}
	</section>
</main>
