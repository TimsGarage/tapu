<script>
	let {
		title = '',
		description = '',
		variant = 'default',
		flush = false,
		meta,
		children
	} = $props();
</script>

<section class="panel" class:panel--danger={variant === 'danger'}>
	{#if title || description || meta}
		<header class="panel__header">
			<div>
				{#if title}
					<h2 class="panel__title">{title}</h2>
				{/if}

				{#if description}
					<p class="panel__description">{description}</p>
				{/if}
			</div>

			{#if meta}
				{@render meta()}
			{/if}
		</header>
	{/if}

	<div class="panel__body" class:panel__body--flush={flush}>
		{@render children()}
	</div>
</section>
