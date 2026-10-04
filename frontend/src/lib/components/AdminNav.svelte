<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		section,
		children
	}: {
		section:
			| 'hub'
			| 'members'
			| 'policies'
			| 'integrations'
			| 'webhooks'
			| 'smtp';
		children: Snippet;
	} = $props();

	// Équipes : icebox produit — API/schéma conservés, hors nav nominale (#295).
	const links = [
		{ href: '/admin', id: 'hub', label: 'Organisation' },
		{ href: '/admin/members', id: 'members', label: 'Membres' },
		{ href: '/admin/settings/policies', id: 'policies', label: 'Politiques' },
		{ href: '/admin/integrations', id: 'integrations', label: 'Intégrations' },
		{ href: '/admin/settings/webhooks', id: 'webhooks', label: 'Webhooks' },
		{ href: '/admin/settings/smtp', id: 'smtp', label: 'SMTP' }
	] as const;
</script>

<nav class="admin-nav" aria-label="Administration">
	{#each links as link (link.id)}
		<a href={link.href} aria-current={section === link.id ? 'page' : undefined}>{link.label}</a>
	{/each}
</nav>

{@render children()}

<style>
	.admin-nav {
		display: flex;
		flex-wrap: wrap;
		gap: var(--mb-space-2) var(--mb-space-4);
		margin: 0 0 var(--mb-space-5);
		padding-bottom: var(--mb-space-3);
		border-bottom: 1px solid var(--mb-color-border);
	}
	.admin-nav a {
		color: var(--mb-color-muted);
		text-decoration: none;
		font-size: var(--mb-font-size-sm);
	}
	.admin-nav a:hover {
		color: var(--mb-color-fg);
	}
	.admin-nav a[aria-current='page'] {
		color: var(--mb-color-fg);
		font-weight: 600;
		text-decoration: underline;
		text-underline-offset: 0.25em;
	}
</style>
