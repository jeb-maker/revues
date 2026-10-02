<script lang="ts">
	const title = 'Revues';
</script>

<svelte:head>
	<title>{title}</title>
</svelte:head>

<main class="shell">
	<p class="brand">{title}</p>
	<h1>API + SvelteKit</h1>
	<p class="lede">
		Scaffold rewrite — l’UI métier arrivera via les work packages suivants.
	</p>
</main>

<style>
	:global(body) {
		margin: 0;
		min-height: 100vh;
		font-family: 'Segoe UI', system-ui, sans-serif;
		background: linear-gradient(160deg, #0f172a 0%, #1e293b 45%, #0f766e 100%);
		color: #f8fafc;
	}
	.shell {
		max-width: 40rem;
		margin: 0 auto;
		padding: 20vh 1.5rem 4rem;
	}
	.brand {
		margin: 0 0 0.75rem;
		font-size: clamp(2.5rem, 8vw, 4rem);
		font-weight: 700;
		letter-spacing: -0.04em;
		line-height: 1;
	}
	h1 {
		margin: 0 0 0.75rem;
		font-size: 1.25rem;
		font-weight: 500;
		color: #99f6e4;
	}
	.lede {
		margin: 0;
		max-width: 28rem;
		line-height: 1.5;
		color: #cbd5e1;
	}
</style>
