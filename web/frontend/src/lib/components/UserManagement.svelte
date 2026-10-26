<script>
	import {
		users,
		services,
		newUser,
		createUser,
		setUserState,
		resetPassword,
		toggleUserService
	} from '#lib/admin.svelte.js';

	import EmptyState from '#lib/components/EmptyState.svelte';
	import Panel from '#lib/components/Panel.svelte';
	import StatusBadge from '#lib/components/StatusBadge.svelte';

	function submitCreateUser(event) {
		event.preventDefault();
		createUser(event);
	}
</script>

<Panel
	title="Users"
	description="Create identities, manage account state, and assign service access."
	flush
>
	{#snippet meta()}
		<StatusBadge>{$users.length} total</StatusBadge>
	{/snippet}

	<form class="panel__section form-grid--compact" onsubmit={submitCreateUser}>
		<label class="field">
			<span>Username</span>
			<input
				class="input"
				bind:value={newUser.username}
				placeholder="name@company.io"
				autocomplete="off"
				required
			/>
		</label>

		<label class="field">
			<span>Initial password</span>
			<input
				class="input"
				bind:value={newUser.password}
				type="password"
				placeholder="Minimum 12 characters"
				autocomplete="new-password"
				minlength="12"
				required
			/>
		</label>

		<label class="checkbox">
			<input bind:checked={newUser.isAdmin} type="checkbox" />
			<span>Administrator</span>
		</label>

		<button class="button button--primary" type="submit"> Create user </button>
	</form>

	{#if $users.length === 0}
		<EmptyState
			title="No users found"
			description="Create the first user with the form above."
			icon="U"
		/>
	{:else}
		<div class="entity-list panel__section">
			{#each $users as user (user.id)}
				<article class="entity-row">
					<div class="entity-row__identity">
						<div class="entity-row__icon" aria-hidden="true">
							{user.username.slice(0, 2).toUpperCase()}
						</div>

						<div class="entity-row__content">
							<div class="form-actions">
								<strong>{user.username}</strong>

								{#if user.isAdmin}
									<StatusBadge type="info">Administrator</StatusBadge>
								{/if}
							</div>

							<code>{user.id}</code>
						</div>
					</div>

					<div class="entity-row__actions">
						<label class="checkbox">
							<input bind:checked={user.isAdmin} type="checkbox" />
							<span>Administrator</span>
						</label>

						<select class="select" bind:value={user.status}>
							<option value="active">Active</option>
							<option value="disabled">Disabled</option>
						</select>

						<button class="button" type="button" onclick={() => setUserState(user)}> Save </button>

						<button class="button" type="button" onclick={() => resetPassword(user)}>
							Reset password
						</button>
					</div>

					<details>
						<summary>
							<span>Service access</span>
							<small>{user.serviceIds?.length ?? 0} assigned</small>
						</summary>

						<div class="details-content form-grid">
							{#each $services as service (service.id)}
								<label class="checkbox">
									<input
										type="checkbox"
										checked={user.serviceIds?.includes(service.id)}
										onchange={() => toggleUserService(user, service)}
									/>
									<span>{service.name}</span>
								</label>
							{/each}

							{#if $services.length === 0}
								<p class="text-muted">No services are available.</p>
							{/if}
						</div>
					</details>
				</article>
			{/each}
		</div>
	{/if}
</Panel>
