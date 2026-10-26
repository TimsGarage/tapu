<script>
	import { services, newService, createService, toggleService } from '#lib/admin.svelte.js';

	import EmptyState from '#lib/components/EmptyState.svelte';
	import Panel from '#lib/components/Panel.svelte';
	import StatusBadge from '#lib/components/StatusBadge.svelte';

	function submitCreateService(event) {
		event.preventDefault();
		createService(event);
	}
</script>

<Panel title="Services" description="Manage protected resource servers and token audiences." flush>
	{#snippet meta()}
		<StatusBadge>{$services.length} total</StatusBadge>
	{/snippet}

	<form class="panel__section form-grid--compact" onsubmit={submitCreateService}>
		<label class="field">
			<span>Service key</span>
			<input class="input" bind:value={newService.key} placeholder="project-a" required />
		</label>

		<label class="field">
			<span>Display name</span>
			<input class="input" bind:value={newService.name} placeholder="Project A" required />
		</label>

		<label class="field">
			<span>Token audience</span>
			<input class="input" bind:value={newService.audience} placeholder="project-a-api" required />
		</label>

		<button class="button button--primary" type="submit"> Create service </button>
	</form>

	{#if $services.length === 0}
		<EmptyState
			title="No services registered"
			description="Register a protected resource server with the form above."
			icon="S"
		/>
	{:else}
		<div class="entity-list">
			{#each $services as service (service.id)}
				<article class="entity-row">
					<div class="entity-row__identity">
						<div class="entity-row__icon" aria-hidden="true">S</div>

						<div class="entity-row__content">
							<strong
								>{service.name}
								<StatusBadge type={service.enabled ? 'success' : 'danger'}>
									{service.enabled ? 'Enabled' : 'Disabled'}
								</StatusBadge></strong
							>
							<small>
								Key: {service.key} · Audience: {service.audience}
							</small>
						</div>
					</div>

					<div class="entity-row__actions">
						<button
							class="button"
							class:button--success={!service.enabled}
							class:button--danger={service.enabled}
							type="button"
							onclick={() => toggleService(service)}
						>
							{service.enabled ? 'Disable' : 'Enable'}
						</button>
					</div>
				</article>
			{/each}
		</div>
	{/if}
</Panel>
