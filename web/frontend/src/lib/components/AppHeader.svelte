<script>
	let { label = '', homeHref = '/', navigation = [], trailing } = $props();
</script>

<header class="app-header">
	<div class="app-header__left">
		<a class="brand" href={homeHref} aria-label="IAM Console home">
			<span class="brand__icon" aria-hidden="true">
				<svg viewBox="0 0 24 24">
					<path
						d="M12 2 4.5 5.25v5.5c0 5.08 3.2 9.67 7.5 11.25 4.3-1.58 7.5-6.17 7.5-11.25v-5.5L12 2Zm0 2.18 5.5 2.38v4.19c0 3.94-2.3 7.69-5.5 9.13-3.2-1.44-5.5-5.19-5.5-9.13V6.56L12 4.18Z"
					/>
				</svg>
			</span>

			<span>IAM Console</span>

			{#if label}
				<span class="brand__label">{label}</span>
			{/if}
		</a>

		{#if navigation.length > 0}
			<nav class="app-header__nav" aria-label="Primary navigation">
				{#each navigation as item}
					<a href={item.href} aria-current={item.current ? 'page' : undefined}>
						{item.label}
					</a>
				{/each}
			</nav>
		{/if}
	</div>

	{#if trailing}
		<div class="app-header__right">
			{@render trailing()}
		</div>
	{/if}
</header>
