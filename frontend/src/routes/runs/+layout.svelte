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
	:global(.page) {
		max-width: 40rem;
		margin: 0 auto;
		padding: 2rem 1.25rem 4rem;
	}
	:global(.page .brand) {
		margin: 0 0 0.35rem;
		font-size: 1.75rem;
		font-weight: 700;
	}
	:global(.page .brand a) {
		color: inherit;
		text-decoration: none;
	}
	:global(.page .crumbs) {
		margin: 0 0 0.5rem;
		font-size: 0.9rem;
		color: #94a3b8;
	}
	:global(.page .crumbs a) {
		color: #5eead4;
	}
	:global(.page h1) {
		margin: 0 0 0.35rem;
		font-size: 1.25rem;
	}
	:global(.page .lede) {
		margin: 0 0 1rem;
		color: #94a3b8;
	}
	:global(.page .err) {
		color: #fca5a5;
	}
	:global(.page .ok) {
		color: #99f6e4;
	}
	:global(.page .muted) {
		color: #94a3b8;
	}

	:global(.page .form) {
		display: flex;
		flex-direction: column;
		gap: 0.9rem;
	}
	:global(.page .form label) {
		display: flex;
		flex-direction: column;
		gap: 0.35rem;
		font-size: 0.9rem;
		color: #cbd5e1;
	}
	:global(.page .form select),
	:global(.page .form textarea),
	:global(.page .form input) {
		padding: 0.55rem 0.65rem;
		border-radius: 0.4rem;
		border: 1px solid #334155;
		background: #0f172a;
		color: inherit;
		font: inherit;
	}
	:global(.page .form button) {
		align-self: flex-start;
		padding: 0.6rem 1rem;
		border: none;
		border-radius: 0.4rem;
		background: #0f766e;
		color: #ecfdf5;
		font-weight: 600;
		cursor: pointer;
		font: inherit;
	}
	:global(.page .form button:disabled) {
		opacity: 0.5;
		cursor: not-allowed;
	}
</style>
