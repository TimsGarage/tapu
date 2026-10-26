<script>
	let { type = 'error', title = '', children } = $props();

	const icon = $derived(type === 'success' ? '✓' : type === 'warning' ? '!' : '!');
</script>

<div
	class="alert"
	class:alert--error={type === 'error'}
	class:alert--success={type === 'success'}
	class:alert--warning={type === 'warning'}
	role={type === 'error' ? 'alert' : 'status'}
>
	<span class="alert__icon" aria-hidden="true">{icon}</span>

	<div>
		{#if title}
			<strong>{title}</strong>
		{/if}

		{@render children()}
	</div>
</div>
