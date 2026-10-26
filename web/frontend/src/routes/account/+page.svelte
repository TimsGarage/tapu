<script>
	import { getUserState } from '#lib/auth.svelte.js';

	import Alert from '#lib/components/Alert.svelte';
	import AppFooter from '#lib/components/AppFooter.svelte';
	import AppHeader from '#lib/components/AppHeader.svelte';
	import EmptyState from '#lib/components/EmptyState.svelte';
	import LoadingState from '#lib/components/LoadingState.svelte';
	import PageHeader from '#lib/components/PageHeader.svelte';
	import Panel from '#lib/components/Panel.svelte';
	import StatusBadge from '#lib/components/StatusBadge.svelte';

	const userStatePromise = getUserState();

	function initials(username) {
		return (
			username
				.split(/[\s.@_-]+/)
				.filter(Boolean)
				.slice(0, 2)
				.map((part) => part[0]?.toUpperCase())
				.join('') || 'U'
		);
	}
</script>

<svelte:head>
	<title>User Account - IAM Console</title>
	<meta
		name="description"
		content="Manage your IAM identity, account access, credentials, and sessions."
	/>
</svelte:head>

<div class="app-shell">
	{#await userStatePromise}
		<main class="page-main page-main--centered">
			<Panel>
				<LoadingState
					title="Loading your account"
					description="Retrieving your IAM identity and session information…"
				/>
			</Panel>
		</main>
	{:then userState}
		{#if userState.error}
			<main class="page-main page-main--centered">
				<Alert title="Unable to load your account">
					<p>{userState.error}</p>
				</Alert>
			</main>
		{:else if userState.user}
			<AppHeader
				label="Account"
				homeHref="/account"
				navigation={[
					...(userState.user.isAdmin ? [{ href: '/admin', label: 'Administration' }] : [])
				]}
			>
				{#snippet trailing()}
					<!-- <div class="avatar" aria-hidden="true">
						{initials(userState.user.username)}
					</div> -->

					<form method="post" action="/api/web/logout">
						<input type="hidden" name="csrf_token" value={userState.csrfToken} />
						<button class="button" type="submit">Sign out</button>
					</form>
				{/snippet}
			</AppHeader>

			<PageHeader
				eyebrow="IAM System / Account & Security"
				title="User Account Profile"
				description="Manage your identity attributes, security credentials, sessions, and cryptographic keys."
			>
				{#snippet actions()}
					{#if userState.user.isAdmin}
						<a class="button button--primary" href="/admin"> Open administration </a>
					{/if}

					<button class="button" type="button" disabled> Export identity data </button>
				{/snippet}
			</PageHeader>

			<main class="page-main">
				<div class="page-stack">
					<section class="identity-overview">
						<div class="identity-overview__main">
							<div class="avatar" aria-hidden="true">
								{initials(userState.user.username)}
							</div>

							<div>
								<div class="form-actions">
									<h2>{userState.user.username}</h2>

									<StatusBadge type="info">
										{userState.user.isAdmin ? 'Administrator' : 'User'}
									</StatusBadge>

									<StatusBadge type="success">Active</StatusBadge>
								</div>

								<p>
									{userState.user.isAdmin
										? 'Administrative access is enabled.'
										: 'Standard account access is enabled.'}
								</p>

								<div class="metadata-list" style="margin-top: 0.75rem;">
									<span class="metadata-item">
										User ID: {userState.user.id}
									</span>
									<span class="metadata-item">Realm: Default</span>
								</div>
							</div>
						</div>

						<StatusBadge type="success">Authenticated</StatusBadge>
					</section>

					<div class="content-grid">
						<div class="content-column">
							<Panel
								title="Identity details"
								description="Identity attributes returned by the current user-state endpoint."
							>
								<dl class="definition-grid">
									<div>
										<dt>Username</dt>
										<dd>{userState.user.username}</dd>
									</div>

									<div>
										<dt>User ID</dt>
										<dd class="monospace">{userState.user.id}</dd>
									</div>

									<div>
										<dt>Account type</dt>
										<dd>
											{userState.user.isAdmin ? 'Administrator' : 'Standard user'}
										</dd>
									</div>

									<div>
										<dt>Status</dt>
										<dd class="text-success">Authenticated</dd>
									</div>
								</dl>
							</Panel>

							<Panel
								title="Registered credentials"
								description="SSH, WebAuthn, signing, and mutual TLS credentials."
							>
								<EmptyState
									title="Credential API not connected"
									description="Registered keys will appear here when the credential endpoint is available."
									icon="K"
								/>
							</Panel>

							<Panel
								title="Authorized OAuth applications"
								description="Applications authorized to access resources for this identity."
							>
								<EmptyState
									title="OAuth authorization data unavailable"
									description="Connected applications will appear here when the authorization endpoint is available."
									icon="O"
								/>
							</Panel>
						</div>

						<aside class="content-column" aria-label="Account security">
							<Panel title="Multi-factor authentication">
								<EmptyState
									title="MFA status unavailable"
									description="Authenticator information is not exposed by the current endpoint."
									icon="M"
								/>
							</Panel>

							<Panel title="Active sessions">
								<EmptyState
									title="Session API not connected"
									description="Browsers, CLI sessions, and expiry information will appear here."
									icon="S"
								/>
							</Panel>

							<Panel
								title="Emergency account lockdown"
								description="Invalidate tokens and active sessions if credentials are compromised."
								variant="danger"
							>
								<button class="button button--danger button--block" type="button" disabled>
									Freeze account credentials
								</button>
							</Panel>
						</aside>
					</div>
				</div>
			</main>

			<AppFooter
				links={[
					{ href: '/security', label: 'Security Audit' },
					{ href: '/docs/oauth', label: 'OAuth 2.0 Specifications' },
					{ href: '/docs/pki', label: 'PKI Guidelines' }
				]}
			/>
		{:else}
			<main class="page-main page-main--centered">
				<Panel>
					<EmptyState
						title="You are not signed in"
						description="Sign in to access your account and security settings."
						icon="↪"
					>
						{#snippet actions()}
							<a class="button button--primary" href="/login"> Go to sign in </a>
						{/snippet}
					</EmptyState>
				</Panel>
			</main>
		{/if}
	{:catch caughtError}
		<main class="page-main page-main--centered">
			<Alert title="Unable to load your account">
				<p>
					{caughtError instanceof Error ? caughtError.message : 'An unexpected error occurred.'}
				</p>
			</Alert>
		</main>
	{/await}
</div>
