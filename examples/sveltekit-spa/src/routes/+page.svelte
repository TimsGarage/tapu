<script>
	import { auth } from '#lib/auth.svelte.js';
</script>

<svelte:head>
	<title>Tapu SvelteKit SPA</title>
</svelte:head>

<section>
	<h1>SvelteKit client-side authentication</h1>

	<p>This application authenticates directly with Tapu using Authorization Code with PKCE.</p>

	{#if auth.loading}
		<p>Loading authentication state…</p>
	{:else if auth.authenticated}
		<p>
			Signed in as
			<strong>{auth.username}</strong>.
		</p>

		<dl>
			<dt>Subject</dt>
			<dd>{auth.subject}</dd>

			<dt>Session expires</dt>
			<dd>
				{auth.expiresAt?.toLocaleString()}
			</dd>
		</dl>
	{:else}
		<p>You are not signed in.</p>

		<button type="button" onclick={() => auth.login()}> Sign in </button>
	{/if}
</section>
