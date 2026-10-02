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
	:global(.admin-page .form),
	:global(.admin-page .test) {
		display: flex;
		flex-direction: column;
		gap: 0.85rem;
	}
	:global(.admin-page .test) {
		margin-top: 2rem;
		padding-top: 1.25rem;
		border-top: 1px solid #334155;
	}
	:global(.admin-page .test h2) {
		margin: 0;
		font-size: 1rem;
		color: #99f6e4;
	}
	:global(.admin-page label) {
		display: flex;
		flex-direction: column;
		gap: 0.35rem;
		font-size: 0.9rem;
		color: #cbd5e1;
	}
	:global(.admin-page label.check) {
		flex-direction: row;
		align-items: center;
		gap: 0.5rem;
	}
	:global(.admin-page input:not([type='checkbox'])) {
		padding: 0.55rem 0.65rem;
		border-radius: 0.4rem;
		border: 1px solid #334155;
		background: #0f172a;
		color: inherit;
		font: inherit;
	}
	:global(.admin-page .actions) {
		display: flex;
		gap: 0.75rem;
		flex-wrap: wrap;
	}
	:global(.admin-page button) {
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
	:global(.admin-page button:disabled) {
		opacity: 0.5;
		cursor: not-allowed;
	}
	:global(.admin-page button.ghost) {
		background: transparent;
		border: 1px solid #475569;
		color: #cbd5e1;
	}

	:global(.admin-page .list) {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}
	:global(.admin-page .list li) {
		padding: 0.9rem 1rem;
		border-radius: 0.45rem;
		background: rgba(15, 23, 42, 0.45);
		border: 1px solid #334155;
	}
	:global(.admin-page .row) {
		display: flex;
		justify-content: space-between;
		gap: 1rem;
		align-items: flex-start;
	}
	:global(.admin-page .name) {
		margin: 0;
		font-weight: 600;
	}
	:global(.admin-page .desc) {
		margin: 0.25rem 0 0;
		font-size: 0.88rem;
		color: #94a3b8;
	}
	:global(.admin-page .badge) {
		font-size: 0.75rem;
		padding: 0.2rem 0.5rem;
		border-radius: 0.3rem;
		background: #334155;
		color: #cbd5e1;
		white-space: nowrap;
	}
	:global(.admin-page .badge.on) {
		background: #115e59;
		color: #99f6e4;
	}
	:global(.admin-page .cfg) {
		display: inline-block;
		margin-top: 0.65rem;
		font-weight: 600;
		font-size: 0.9rem;
	}
</style>
