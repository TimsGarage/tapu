<script>
	let { links = [], year = new Date().getFullYear() } = $props();
</script>

<footer class="app-footer">
	<div class="app-footer__content">
		<span>
			© {year} IAM Security Systems. Enterprise Identity &amp; Access Management.
		</span>

		{#if links.length > 0}
			<nav aria-label="Footer links">
				{#each links as link}
					<a href={link.href}>{link.label}</a>
				{/each}
			</nav>
		{/if}
	</div>
</footer>
