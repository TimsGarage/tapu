<script>
	let { type = 'neutral', children } = $props();
</script>

<span
	class="badge"
	class:badge--success={type === 'success'}
	class:badge--danger={type === 'danger'}
	class:badge--info={type === 'info'}
	class:badge--neutral={type === 'neutral'}
>
	{@render children()}
</span>
