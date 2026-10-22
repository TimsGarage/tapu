<script>
	import { onMount } from 'svelte';

	import { auth } from '#lib/auth.svelte.js';
	import '../app.css';

	let { children } = $props();

	onMount(() => {
		auth.initialize();
	});
</script>

<svelte:head>
	<meta name="description" content="Tapu SvelteKit client-side OAuth example" />
</svelte:head>

<header>
	<a class="brand" href="/"> Tapu SvelteKit SPA </a>

	<nav aria-label="Main navigation">
		<a href="/">Home</a>
		<a href="/protected">Protected</a>
	</nav>

	<div class="session">
		{#if auth.loading}
			<span>Loading…</span>
		{:else if auth.authenticated}
			<span>{auth.username}</span>

			<button type="button" onclick={() => auth.logout()}> Sign out locally </button>
		{:else}
			<button type="button" onclick={() => auth.login()}> Sign in </button>
		{/if}
	</div>
</header>

<main>
	{#if auth.error}
		<p class="error" role="alert">
			{auth.error}
		</p>
	{/if}

	{@render children()}
</main>
