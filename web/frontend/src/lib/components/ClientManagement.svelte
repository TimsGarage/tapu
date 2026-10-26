<script>
	import {
		clients,
		services,
		newClient,
		createClient,
		toggleClient,
		rotateSecret
	} from '#lib/admin.svelte.js';

	import EmptyState from '#lib/components/EmptyState.svelte';
	import Panel from '#lib/components/Panel.svelte';
	import StatusBadge from '#lib/components/StatusBadge.svelte';

	function submitCreateClient(event) {
		event.preventDefault();
		createClient(event);
	}
</script>

<Panel
	title="OAuth 2.0 Clients"
	description="Register applications, configure redirect URIs, and rotate secrets."
	flush
>
	{#snippet meta()}
		<StatusBadge>{$clients.length} total</StatusBadge>
	{/snippet}

	<form class="panel__section form-grid--compact" onsubmit={submitCreateClient}>
		<label class="field">
			<span>Client name</span>
			<input class="input" bind:value={newClient.name} placeholder="Project A web" required />
		</label>

		<label class="field">
			<span>Client type</span>
			<select class="select" bind:value={newClient.clientType}>
				<option value="confidential">Confidential</option>
				<option value="public">Public</option>
			</select>
		</label>

		<label class="field">
			<span>Redirect URI</span>
			<input
				class="input"
				bind:value={newClient.redirectURI}
				type="url"
				placeholder="https://example.com/auth/callback"
				required
			/>
		</label>

		<label class="field">
			<span>Associated service</span>
			<select class="select" bind:value={newClient.serviceId}>
				<option value="">No service</option>

				{#each $services as service (service.id)}
					<option value={service.id}>{service.name}</option>
				{/each}
			</select>
		</label>

		<button class="button button--primary" type="submit"> Create client </button>
	</form>

	{#if $clients.length === 0}
		<EmptyState
			title="No OAuth clients registered"
			description="Register an OAuth application with the form above."
			icon="O"
		/>
	{:else}
		<div class="entity-list">
			{#each $clients as client (client.id)}
				<article class="entity-row">
					<div class="entity-row__identity">
						<div class="entity-row__icon" aria-hidden="true">
							{client.name.slice(0, 2).toUpperCase()}
						</div>

						<div class="entity-row__content">
							<div class="form-actions">
								<strong>{client.name}</strong>
								<StatusBadge>{client.clientType}</StatusBadge>
								<StatusBadge type={client.enabled ? 'success' : 'danger'}>
									{client.enabled ? 'Enabled' : 'Disabled'}
								</StatusBadge>
							</div>

							<code>Client ID: {client.clientId}</code>
						</div>
					</div>

					<div class="entity-row__actions">
						{#if client.clientType === 'confidential'}
							<button class="button" type="button" onclick={() => rotateSecret(client)}>
								Rotate secret
							</button>
						{/if}
						<button
							class="button"
							class:button--success={!client.enabled}
							class:button--danger={client.enabled}
							type="button"
							onclick={() => toggleClient(client)}
						>
							{client.enabled ? 'Disable' : 'Enable'}
						</button>
					</div>

					<details>
						<summary>
							<span>Client configuration</span>
							<small>{client.redirectUris.length} redirect URI(s)</small>
						</summary>

						<div class="details-content">
							<div class="form-stack">
								{#each client.redirectUris as redirectURI (redirectURI)}
									<code class="metadata-item">{redirectURI}</code>
								{/each}
							</div>
						</div>
					</details>
				</article>
			{/each}
		</div>
	{/if}
</Panel>
