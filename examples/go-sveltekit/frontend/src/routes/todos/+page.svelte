<script>
	import { onMount } from 'svelte';

	import { apiRequest, APIError } from '#lib/api.js';
	import { auth } from '#lib/auth.svelte.js';

	let todos = $state([]);
	let newTitle = $state('');
	let loading = $state(false);
	let error = $state('');

	onMount(async () => {
		if (auth.authenticated) {
			await loadTodos();
		}
	});

	async function loadTodos() {
		loading = true;
		error = '';

		try {
			todos = await apiRequest('/api/todos/');
		} catch (caught) {
			handleError(caught);
		} finally {
			loading = false;
		}
	}

	async function createTodo() {
		const title = newTitle.trim();

		if (!title) {
			return;
		}

		error = '';

		try {
			const created = await apiRequest('/api/todos/', {
				method: 'POST',
				headers: {
					'X-CSRF-Token': auth.csrfToken
				},
				body: { title }
			});

			todos = [...todos, created];
			newTitle = '';
		} catch (caught) {
			handleError(caught);
		}
	}

	async function toggleTodo(todo) {
		error = '';

		try {
			const updated = await apiRequest(`/api/todos/${todo.id}`, {
				method: 'PUT',
				headers: {
					'X-CSRF-Token': auth.csrfToken
				},
				body: {
					title: todo.title,
					completed: !todo.completed
				}
			});

			todos = todos.map((item) => (item.id === updated.id ? updated : item));
		} catch (caught) {
			handleError(caught);
		}
	}

	async function deleteTodo(todo) {
		error = '';

		try {
			await apiRequest(`/api/todos/${todo.id}`, {
				method: 'DELETE',
				headers: {
					'X-CSRF-Token': auth.csrfToken
				}
			});

			todos = todos.filter((item) => item.id !== todo.id);
		} catch (caught) {
			handleError(caught);
		}
	}

	function handleError(caught) {
		if (caught instanceof APIError && caught.status === 401) {
			auth.user = null;
			auth.csrfToken = '';
			error = 'Your session has expired.';
			return;
		}

		error = caught instanceof Error ? caught.message : 'Request failed.';
	}
</script>

<svelte:head>
	<title>Todos</title>
</svelte:head>

<section>
	<h1>Protected todos</h1>

	{#if auth.loading}
		<p>Loading authentication state…</p>
	{:else if !auth.authenticated}
		<p>You must sign in before using the todo API.</p>

		<button type="button" onclick={() => auth.login()}> Sign in </button>
	{:else}
		<form
			onsubmit={(event) => {
				event.preventDefault();
				createTodo();
			}}
		>
			<label for="new-todo">New todo</label>

			<div class="todo-form">
				<input id="new-todo" bind:value={newTitle} maxlength="200" required />

				<button type="submit"> Add </button>
			</div>
		</form>

		{#if loading}
			<p>Loading todos…</p>
		{/if}

		{#if error}
			<p class="error" role="alert">
				{error}
			</p>
		{/if}

		{#if !loading && todos.length === 0}
			<p>No todos yet.</p>
		{/if}

		<ul class="todos">
			{#each todos as todo (todo.id)}
				<li>
					<label>
						<input type="checkbox" checked={todo.completed} onchange={() => toggleTodo(todo)} />

						<span class:completed={todo.completed}>
							{todo.title}
						</span>
					</label>

					<button type="button" onclick={() => deleteTodo(todo)}> Delete </button>
				</li>
			{/each}
		</ul>
	{/if}
</section>
