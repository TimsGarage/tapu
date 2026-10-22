<script>
	import { auth } from '#lib/auth.svelte.js';
</script>

<svelte:head>
	<title>Tapu Go and SvelteKit example</title>
</svelte:head>

<section>
	<h1>Go backend with SvelteKit</h1>

	<p>
		This example keeps OAuth tokens in the Go backend. The browser receives only an opaque
		application session cookie.
	</p>

	{#if auth.loading}
		<p>Loading authentication state…</p>
	{:else if auth.authenticated}
		<p>
			Signed in as
			<strong>{auth.username}</strong>.
		</p>

		<dl>
			<dt>IAM subject</dt>
			<dd>{auth.user.subject}</dd>

			<dt>Audience</dt>
			<dd>{auth.user.audience}</dd>

			<dt>Scopes</dt>
			<dd>{auth.user.scope}</dd>

			<dt>Session expires</dt>
			<dd>
				{new Date(auth.user.sessionExpiresAt).toLocaleString()}
			</dd>
		</dl>
	{:else}
		<p>You are not signed in.</p>

		<button type="button" onclick={() => auth.login()}> Sign in </button>
	{/if}
</section>
