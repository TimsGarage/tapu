<script>
	import { onMount } from 'svelte';

	import { Login } from '../../lib/auth.svelte';

	let loading = $state(true);
	let loadError = $state('');
	let csrfToken = $state('');
	let returnTo = $state('');

	const loginError = $derived.by(() => {
		const value = new URLSearchParams(window.location.search).get('error');

		switch (value) {
			case 'invalid_credentials':
				return 'Invalid username or password.';

			case 'rate_limited':
				return 'Too many login attempts. Please try again later.';

			default:
				return '';
		}
	});

	onMount(async () => {
		const res = await Login();
		if (res.error !== '') {
			loadError = res.error;
		} else {
			csrfToken = res.csrfToken;
			returnTo = res.returnTo;
		}
		loading = false;
	});
</script>

<svelte:head>
	<title>Sign in</title>
</svelte:head>

<section class="auth-card">
	<h1>Sign in</h1>

	{#if loginError}
		<p class="error" role="alert">
			{loginError}
		</p>
	{/if}

	{#if loadError}
		<p class="error" role="alert">
			{loadError}
		</p>
	{:else if loading}
		<p>Preparing sign-in…</p>
	{:else}
		<form method="post" action="/api/web/login">
			<input type="hidden" name="csrf_token" bind:value={csrfToken} />

			{#if returnTo}
				<input type="hidden" name="return_to" bind:value={returnTo} />
			{/if}

			<label for="username">Username</label>
			<input
				id="username"
				name="username"
				type="text"
				autocomplete="username"
				maxlength="255"
				required
				autofocus
			/>

			<label for="password">Password</label>
			<input
				id="password"
				name="password"
				type="password"
				autocomplete="current-password"
				required
			/>

			<button type="submit"> Sign in </button>
		</form>
	{/if}
</section>
