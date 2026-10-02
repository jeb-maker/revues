<script lang="ts">
	let { children } = $props();
</script>

{@render children()}

<style>
	:global(body) {
		margin: 0;
		min-height: 100vh;
		font-family: var(--mb-font-sans, 'Segoe UI', system-ui, sans-serif);
		background: linear-gradient(160deg, #0f172a 0%, #1e293b 50%, #0f766e 100%);
		color: #f8fafc;
	}
	:global(.admin-page) {
		max-width: 40rem;
		margin: 0 auto;
		padding: 2rem 1.25rem 4rem;
	}
	:global(.admin-page .brand) {
		margin: 0 0 0.35rem;
		font-size: 1.75rem;
		font-weight: 700;
	}
	:global(.admin-page .brand a) {
		color: inherit;
		text-decoration: none;
	}
	:global(.admin-page .crumbs) {
		margin: 0 0 0.5rem;
		font-size: 0.9rem;
		color: #94a3b8;
	}
	:global(.admin-page .crumbs a),
	:global(.admin-page a.cfg),
	:global(.admin-page a.back) {
		color: #5eead4;
	}
	:global(.admin-page h1) {
		margin: 0 0 0.35rem;
		font-size: 1.35rem;
	}
	:global(.admin-page .lede) {
		margin: 0 0 1.25rem;
		color: #94a3b8;
		line-height: 1.5;
	}
	:global(.admin-page .err) {
		color: #fca5a5;
	}
	:global(.admin-page .ok) {
		color: #6ee7b7;
	}
	:global(.admin-page .muted) {
		color: #94a3b8;
	}
</style>
