<script>
	let { title, description, icon = '◇', actions } = $props();
</script>

<div class="empty-state">
	<span class="empty-state__icon" aria-hidden="true">{icon}</span>
	<strong>{title}</strong>
	<p>{description}</p>

	{#if actions}
		<div class="form-actions">
			{@render actions()}
		</div>
	{/if}
</div>
