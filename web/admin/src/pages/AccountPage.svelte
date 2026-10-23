<script>
	import { onMount } from "svelte";

	import { apiRequest } from "../lib/public-api.js";

	let loading = $state(true);
	let error = $state("");
	let user = $state(null);
	let csrfToken = $state("");

	onMount(async () => {
		try {
			const result = await apiRequest("/api/web/account");

			user = result.user;
			csrfToken = result.csrfToken;
		} catch (caught) {
			if (caught && typeof caught === "object" && caught.status === 401) {
				window.location.replace("/login");
				return;
			}

			error =
				caught instanceof Error
					? caught.message
					: "Unable to load the account.";
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>Account</title>
</svelte:head>

<section>
	<h1>Your account</h1>

	{#if error}
		<p class="error" role="alert">
			{error}
		</p>
	{:else if loading}
		<p>Loading account…</p>
	{:else if user}
		<dl>
			<dt>Username</dt>
			<dd>{user.username}</dd>

			<dt>User ID</dt>
			<dd>{user.id}</dd>

			<dt>Administrator</dt>
			<dd>{user.isAdmin ? "Yes" : "No"}</dd>
		</dl>

		{#if user.isAdmin}
			<p>
				<a href="/admin"> Open administration </a>
			</p>
		{/if}

		<form method="post" action="/api/web/logout">
			<input type="hidden" name="csrf_token" value={csrfToken} />

			<button type="submit"> Sign out </button>
		</form>
	{/if}
</section>
