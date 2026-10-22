<script>
	import { auth } from '#lib/auth.svelte.js';

	let userInfo = $state(null);
	let userInfoError = $state('');
	let loadingUserInfo = $state(false);

	async function loadUserInfo() {
		loadingUserInfo = true;
		userInfoError = '';

		try {
			userInfo = await auth.loadUserInfo();
		} catch (error) {
			userInfoError = error instanceof Error ? error.message : 'Unable to load UserInfo.';
		} finally {
			loadingUserInfo = false;
		}
	}
</script>

<svelte:head>
	<title>Protected view</title>
</svelte:head>

<section>
	<h1>Protected view</h1>

	{#if auth.loading}
		<p>Loading authentication state…</p>
	{:else if !auth.authenticated}
		<p>You must sign in to view this page.</p>

		<button type="button" onclick={() => auth.login('/protected')}> Sign in </button>
	{:else}
		<p>
			Welcome,
			<strong>{auth.username}</strong>.
		</p>

		<dl>
			<dt>Subject</dt>
			<dd>{auth.subject}</dd>

			<dt>Scopes</dt>
			<dd>{auth.user.scope}</dd>

			<dt>Expires</dt>
			<dd>{auth.expiresAt?.toLocaleString()}</dd>
		</dl>

		<button type="button" disabled={loadingUserInfo} onclick={loadUserInfo}>
			{loadingUserInfo ? 'Loading…' : 'Load UserInfo'}
		</button>

		{#if userInfoError}
			<p class="error" role="alert">
				{userInfoError}
			</p>
		{/if}

		{#if userInfo}
			<pre>{JSON.stringify(userInfo, null, 2)}</pre>
		{/if}
	{/if}
</section>
