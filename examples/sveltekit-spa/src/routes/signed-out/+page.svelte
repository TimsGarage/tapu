<script>
	import { auth } from '#lib/auth.svelte.js';
</script>

<svelte:head>
	<title>Signed out</title>
</svelte:head>

<section>
	<h1>Signed out locally</h1>

	<p>
		The tokens stored by this browser tab have been removed. The central IAM session remains active.
	</p>

	<p>
		Signing in again may complete without another password prompt because the IAM still provides
		SSO.
	</p>

	<button type="button" onclick={() => auth.login()}> Sign in again </button>
</section>
