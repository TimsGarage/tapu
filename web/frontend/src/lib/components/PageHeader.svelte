<script>
	let { eyebrow = '', title, description = '', actions } = $props();
</script>

<header class="page-header">
	<div class="page-header__content">
		<div>
			{#if eyebrow}
				<p class="page-header__eyebrow">{eyebrow}</p>
			{/if}

			<h1>{title}</h1>

			{#if description}
				<p>{description}</p>
			{/if}
		</div>

		{#if actions}
			<div class="page-header__actions">
				{@render actions()}
			</div>
		{/if}
	</div>
</header>
