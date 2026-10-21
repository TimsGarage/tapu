<script>
	import { onMount } from "svelte";
	import { auth, loadAuthentication, login, logout } from "./auth.js";

	let error = "";

	onMount(async () => {
		try {
			await loadAuthentication();
		} catch (caught) {
			error = caught.message;
		}
	});

	async function signOut() {
		error = "";

		try {
			await logout($auth.csrfToken);
		} catch (caught) {
			error = caught.message;
		}
	}
</script>

{#if error}
	<p role="alert">{error}</p>
{/if}

{#if $auth.loading}
	<p>Loading authentication…</p>
{:else if $auth.authenticated}
	<p>
		Signed in as
		<strong>
			{$auth.user.username || $auth.user.subject}
		</strong>
	</p>

	<button type="button" on:click={signOut}>
		Sign out of this application
	</button>
{:else}
	<button type="button" on:click={login}> Sign in </button>
{/if}
