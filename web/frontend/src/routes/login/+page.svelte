<script>
	import { onMount } from 'svelte';
	import { page } from '$app/state';

	import { Login } from '#lib/auth.svelte.js';
	import Alert from '#lib/components/Alert.svelte';
	import AppFooter from '#lib/components/AppFooter.svelte';
	import AppHeader from '#lib/components/AppHeader.svelte';
	import LoadingState from '#lib/components/LoadingState.svelte';
	import Panel from '#lib/components/Panel.svelte';
	import StatusBadge from '#lib/components/StatusBadge.svelte';

	let loading = $state(true);
	let loadError = $state('');
	let csrfToken = $state('');
	let returnTo = $state('');
	let showPassword = $state(false);

	const loginError = $derived.by(() => {
		switch (page.url.searchParams.get('error')) {
			case 'invalid_credentials':
				return 'Invalid username or password.';

			case 'rate_limited':
				return 'Too many login attempts. Please try again later.';

			default:
				return '';
		}
	});

	onMount(() => {
		void prepareLogin();
	});

	async function prepareLogin() {
		try {
			const response = await Login();

			if (response.error) {
				loadError = response.error;
			} else {
				csrfToken = response.csrfToken;
				returnTo = response.returnTo;
			}
		} catch {
			loadError = 'Unable to prepare sign-in. Please refresh the page and try again.';
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Sign In - IAM Platform</title>
	<meta
		name="description"
		content="Authenticate into the enterprise identity and access management platform."
	/>
</svelte:head>

<div class="app-shell">
	<AppHeader label="">
		{#snippet trailing()}
			<!-- <a class="text-link" href="/docs">Help &amp; Docs</a> -->

			<StatusBadge type="success">Systems Normal</StatusBadge>
		{/snippet}
	</AppHeader>

	<main class="page-main page-main--centered">
		<div class="page-stack">
			<Panel
				title="Authenticate into IAM"
				description="Zero-trust access for OAuth 2.0 resources and cryptographic keys."
			>
				<div class="form-stack">
					{#if loginError}
						<Alert title="Sign-in failed">
							<p>{loginError}</p>
						</Alert>
					{/if}

					{#if loadError}
						<Alert title="Unable to prepare sign-in">
							<p>{loadError}</p>
						</Alert>
					{:else if loading}
						<LoadingState
							title="Preparing sign-in"
							description="Establishing a secure authentication session…"
						/>
					{:else}
						<form class="form-stack" method="post" action="/api/web/login">
							<input type="hidden" name="csrf_token" value={csrfToken} />

							{#if returnTo}
								<input type="hidden" name="return_to" value={returnTo} />
							{/if}

							<label class="field">
								<span>Username</span>
								<input
									class="input monospace"
									name="username"
									type="text"
									placeholder="max-mustermann"
									autocomplete="username"
									maxlength="255"
									required
									autofocus
								/>
							</label>

							<label class="field">
								<span>Password</span>
								<input
									class="input monospace"
									name="password"
									type={showPassword ? 'text' : 'password'}
									placeholder="••••••••••••"
									autocomplete="current-password"
									required
								/>
							</label>

							<div class="split wrap">
								<button
									class="text-link"
									type="button"
									onclick={() => (showPassword = !showPassword)}
								>
									{showPassword ? 'Hide password' : 'Show password'}
								</button>
							</div>

							<button class="button button--primary button--block" type="submit"> Sign in </button>

							<div class="split wrap text-muted">
								<small class="monospace">Realm: placeholder</small>
							</div>
						</form>
					{/if}
				</div>
			</Panel>
			<!-- 
			<p class="text-muted" style="margin: 0; text-align: center; font-size: 0.72rem;">
				Protected by OpenID Connect 1.0 and WebAuthn / FIDO Alliance specifications.
			</p> -->
		</div>
	</main>

	<!-- <AppFooter
		links={[
			{ href: '/security-policy', label: 'Security Policy' },
			{ href: '/public-keys', label: 'Public Key Directory' },
			{ href: '/support', label: 'Support Desk' }
		]}
	/> -->
</div>
