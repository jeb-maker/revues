<script lang="ts" module>
	export type AdminNavSection =
		| 'hub'
		| 'members'
		| 'policies'
		| 'integrations'
		| 'webhooks'
		| 'smtp';

	export function adminNavSection(pathname: string): AdminNavSection {
		if (pathname.startsWith('/admin/members')) return 'members';
		if (pathname.startsWith('/admin/settings/policies')) return 'policies';
		if (pathname.startsWith('/admin/settings/webhooks')) return 'webhooks';
		if (pathname.startsWith('/admin/settings/smtp')) return 'smtp';
		if (pathname.startsWith('/admin/integrations')) return 'integrations';
		return 'hub';
	}
</script>

<script lang="ts">
	let { section }: { section: AdminNavSection } = $props();

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

<style>
	.admin-nav {
		display: flex;
		flex-wrap: nowrap;
		gap: var(--mb-space-2) var(--mb-space-4);
		margin: 0 0 var(--mb-space-5);
		padding-bottom: var(--mb-space-3);
		border-bottom: 1px solid var(--mb-color-border);
		overflow-x: auto;
	}
	.admin-nav a {
		flex: 0 0 auto;
		color: var(--mb-color-muted);
		text-decoration: none;
		font-size: var(--mb-font-size-sm);
		/* Même graisse inactif/actif : évite le décalage horizontal des onglets. */
		font-weight: 600;
	}
	.admin-nav a:hover {
		color: var(--mb-color-fg);
	}
	.admin-nav a[aria-current='page'] {
		color: var(--mb-color-fg);
		text-decoration: underline;
		text-underline-offset: 0.25em;
	}
</style>
