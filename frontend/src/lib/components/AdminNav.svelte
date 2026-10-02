<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		section,
		children
	}: {
		section: 'hub' | 'users' | 'members' | 'teams' | 'policies' | 'integrations' | 'smtp';
		children: Snippet;
	} = $props();

	const links = [
		{ href: '/admin', id: 'hub', label: 'Organisation' },
		{ href: '/admin/users', id: 'users', label: 'Emails autorisés' },
		{ href: '/admin/members', id: 'members', label: 'Membres' },
		{ href: '/admin/teams', id: 'teams', label: 'Équipes' },
		{ href: '/admin/settings/policies', id: 'policies', label: 'Politiques' },
		{ href: '/admin/integrations', id: 'integrations', label: 'Intégrations' },
		{ href: '/admin/settings/smtp', id: 'smtp', label: 'SMTP' }
	] as const;
</script>

<nav class="admin-nav" aria-label="Administration">
	{#each links as link}
		<a href={link.href} aria-current={section === link.id ? 'page' : undefined}>{link.label}</a>
	{/each}
</nav>

{@render children()}

<style>
	.admin-nav {
		display: flex;
		flex-wrap: wrap;
		gap: 0.65rem 1rem;
		margin: 0 0 1.5rem;
		padding-bottom: 0.85rem;
		border-bottom: 1px solid rgba(148, 163, 184, 0.35);
	}
	.admin-nav a {
		color: #99f6e4;
		text-decoration: none;
		font-size: 0.92rem;
	}
	.admin-nav a[aria-current='page'] {
		color: #f8fafc;
		font-weight: 600;
		text-decoration: underline;
		text-underline-offset: 0.25em;
	}
</style>
