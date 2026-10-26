<script>
	let { title = 'Loading', description = 'Please wait…' } = $props();
</script>

<div class="loading-state" role="status" aria-live="polite">
	<span class="spinner" aria-hidden="true"></span>
	<strong>{title}</strong>
	<p>{description}</p>
</div>
