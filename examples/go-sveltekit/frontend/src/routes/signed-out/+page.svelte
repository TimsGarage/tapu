<script>
	import { auth } from '#lib/auth.svelte.js';
</script>

<svelte:head>
	<title>Signed out</title>
</svelte:head>

<section>
	<h1>Signed out of the application</h1>

	<p>Your local application session has been removed. Your central Tapu session remains active.</p>

	<button type="button" onclick={() => auth.login()}> Sign in again </button>
</section>
