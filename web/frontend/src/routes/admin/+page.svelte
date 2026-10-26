<script>
	import { onMount } from 'svelte';

	import {
		users,
		services,
		clients,
		error,
		notice,
		clientSecret,
		loadAdminPage
	} from '#lib/admin.svelte.js';

	import Alert from '#lib/components/Alert.svelte';
	import AppFooter from '#lib/components/AppFooter.svelte';
	import AppHeader from '#lib/components/AppHeader.svelte';
	import LoadingState from '#lib/components/LoadingState.svelte';
	import PageHeader from '#lib/components/PageHeader.svelte';
	import Panel from '#lib/components/Panel.svelte';
	import StatusBadge from '#lib/components/StatusBadge.svelte';

	import ClientManagement from '#lib/components/ClientManagement.svelte';
	import ServiceManagement from '#lib/components/ServiceManagement.svelte';
	import UserManagement from '#lib/components/UserManagement.svelte';

	let loading = $state(true);
	let initializationError = $state('');
	let secretCopied = $state(false);

	onMount(() => {
		void initialize();
	});

	async function initialize() {
		try {
			await loadAdminPage();
		} catch (caughtError) {
			initializationError =
				caughtError instanceof Error ? caughtError.message : 'Unable to load administration data.';
		} finally {
			loading = false;
		}
	}

	async function copyClientSecret() {
		if (!$clientSecret || !navigator.clipboard) return;

		await navigator.clipboard.writeText($clientSecret);
		secretCopied = true;

		window.setTimeout(() => {
			secretCopied = false;
		}, 2000);
	}
</script>

<svelte:head>
	<title>IAM Administration - IAM Console</title>
	<meta name="description" content="Manage IAM users, protected services, and OAuth 2.0 clients." />
</svelte:head>

<div class="app-shell">
	<AppHeader
		label="Administration"
		homeHref="/admin"
		navigation={[
			{ href: '#users', label: 'Users' },
			{ href: '#services', label: 'Services' },
			{ href: '#clients', label: 'OAuth Clients' }
		]}
	>
		{#snippet trailing()}
			<a class="button" href="/account">My Account</a>
		{/snippet}
	</AppHeader>

	<PageHeader
		eyebrow="IAM System / Administration"
		title="IAM Administration"
		description="Manage users, service access, protected resource servers, and OAuth 2.0 clients."
	>
		{#snippet actions()}
			<StatusBadge>{$users.length} users</StatusBadge>
			<StatusBadge>{$services.length} services</StatusBadge>
			<StatusBadge>{$clients.length} clients</StatusBadge>
		{/snippet}
	</PageHeader>

	<main class="page-main">
		<div class="page-stack">
			{#if initializationError}
				<Alert title="Unable to load administration data">
					<p>{initializationError}</p>
				</Alert>
			{/if}

			{#if $error}
				<Alert title="Administration request failed">
					<p>{$error}</p>
				</Alert>
			{/if}

			{#if $notice}
				<Alert type="success" title="Request completed">
					<p>{$notice}</p>
				</Alert>
			{/if}

			{#if $clientSecret}
				<Panel
					title="Save this client secret now"
					description="For security reasons, this secret will not be displayed again."
				>
					<div class="form-stack">
						<code class="metadata-item" style="overflow-wrap: anywhere;">
							{$clientSecret}
						</code>

						<div class="form-actions">
							<button class="button button--primary" type="button" onclick={copyClientSecret}>
								{secretCopied ? 'Copied' : 'Copy secret'}
							</button>
						</div>
					</div>
				</Panel>
			{/if}

			{#if loading}
				<Panel>
					<LoadingState
						title="Loading administration"
						description="Retrieving users, services, and OAuth clients…"
					/>
				</Panel>
			{:else if !initializationError}
				<section id="users">
					<UserManagement />
				</section>

				<section id="services">
					<ServiceManagement />
				</section>

				<section id="clients">
					<ClientManagement />
				</section>
			{/if}
		</div>
	</main>

	<AppFooter
		links={[
			{ href: '/account', label: 'My Account' },
			{ href: '/security', label: 'Security Policy' },
			{ href: '/docs/oauth', label: 'OAuth 2.0 Documentation' }
		]}
	/>
</div>
