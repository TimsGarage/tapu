<script>
	import { getUserState } from '../../lib/auth.svelte';
</script>

<svelte:head>
	<title>Account</title>
</svelte:head>

{#await getUserState()}
	<h2>Loading...</h2>
{:then UserState}
	<section>
		<h1>Your account</h1>

		{#if UserState.error}
			<p class="error" role="alert">
				{error}
			</p>
		{:else if UserState.user}
			<dl>
				<dt>Username</dt>
				<dd>{UserState.user.username}</dd>

				<dt>User ID</dt>
				<dd>{UserState.user.id}</dd>

				<dt>Administrator</dt>
				<dd>{UserState.user.isAdmin ? 'Yes' : 'No'}</dd>
			</dl>

			{#if UserState.user.isAdmin}
				<p>
					<a href="/admin"> Open administration </a>
				</p>
			{/if}

			<form method="post" action="/api/web/logout">
				<input type="hidden" name="csrf_token" value={UserState.csrfToken} />

				<button type="submit"> Sign out </button>
			</form>
		{:else}
			<h3>You are not logged in (Live should auto redirect to login)</h3>
			<a href="/login">Go to login</a>
		{/if}
	</section>
{/await}
