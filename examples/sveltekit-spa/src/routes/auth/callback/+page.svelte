<script>
	import { onMount } from 'svelte';

	import { auth } from '#lib/auth.svelte.js';

	let callbackError = $state('');

	onMount(async () => {
		try {
			await auth.completeLogin();
		} catch (error) {
			callbackError = error instanceof Error ? error.message : 'Authentication failed.';
		}
	});
</script>

<svelte:head>
	<title>Completing sign-in</title>
</svelte:head>

<section>
	<h1>Completing sign-in</h1>

	{#if callbackError}
		<p class="error" role="alert">
			{callbackError}
		</p>

		<a href="/">Return home</a>
	{:else}
		<p>Please wait while the authentication response is verified.</p>
	{/if}
</section>
