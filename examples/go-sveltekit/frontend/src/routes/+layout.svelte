<script>
	import '../app.css';
	import { onMount } from 'svelte';

	import { auth } from '#lib/auth.svelte.js';

	let { children } = $props();

	onMount(() => {
		auth.initialize();
	});
</script>

<header>
	<a class="brand" href="/"> Tapu Go + SvelteKit </a>

	<nav aria-label="Main navigation">
		<a href="/">Home</a>
		<a href="/todos">Todos</a>
	</nav>

	<div class="session">
		{#if auth.loading}
			<span>Loading…</span>
		{:else if auth.authenticated}
			<span>{auth.username}</span>

			<button type="button" onclick={() => auth.logout()}> Sign out </button>
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
