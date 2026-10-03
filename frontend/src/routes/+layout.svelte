<script lang="ts">
	import '$lib/styles/app.css';
	import type { Snippet } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { logout } from '$lib/api/auth';
	import { resetSession } from '$lib/auth/session';
	import type { LayoutData } from './$types';

	let { data, children }: { data: LayoutData; children: Snippet } = $props();

	const boot = $derived(data.boot);
	let logoutError = $state('');
	let loggingOut = $state(false);

	const links = [
		{ href: '/runs', label: 'Revues' },
		{ href: '/subjects', label: 'Sujets' },
		{ href: '/modeles', label: 'Modèles' },
		{ href: '/mes-taches', label: 'Mes tâches' }
	] as const;

	function current(href: string): 'page' | undefined {
		const path = page.url.pathname;
		return path === href || path.startsWith(`${href}/`) ? 'page' : undefined;
	}

	async function onLogout() {
		logoutError = '';
		loggingOut = true;
		try {
			await logout(boot.csrf_token);
			resetSession();
			await goto('/login', { invalidateAll: true });
		} catch (e) {
			logoutError = e instanceof Error ? e.message : 'Déconnexion impossible.';
		} finally {
			loggingOut = false;
		}
	}
</script>

<a class="skip-link" href="#main">Aller au contenu</a>

<header class="app-header">
	<div class="app-header__inner">
		<a class="brand" href="/">Revues</a>
		{#if boot.authenticated && boot.user}
			<mb-nav-toggle for="main-nav" label-open="Ouvrir le menu" label-close="Fermer le menu"
			></mb-nav-toggle>
			<mb-nav id="main-nav" label="Principale">
				{#each links as link (link.href)}
					<a href={link.href} aria-current={current(link.href)}>{link.label}</a>
				{/each}
				{#if boot.can_admin}
					<a href="/admin" aria-current={current('/admin')}>Admin</a>
				{/if}
			</mb-nav>
			<div class="app-header__account">
				{#if boot.organization}
					<span>
						<strong>{boot.organization.name}</strong>
						{#if boot.organization_count > 1}
							· <a href="/org/select">Changer</a>
						{/if}
					</span>
				{/if}
				<span>{boot.user.display_name}</span>
				<mb-button variant="ghost" size="sm" disabled={loggingOut} onclick={onLogout}>
					Se déconnecter
				</mb-button>
			</div>
		{/if}
	</div>
	{#if logoutError}
		<mb-alert variant="danger">{logoutError}</mb-alert>
	{/if}
</header>

<main id="main" tabindex="-1">
	{#if data.apiError}
		<div class="page page--narrow">
			<header class="page-header">
				<h1>Service indisponible</h1>
				<p class="lede">L'API Revues ne répond pas. Réessayez dans un instant.</p>
			</header>
			<mb-alert variant="danger">{data.apiError}</mb-alert>
			<p class="actions"><a href={page.url.pathname} data-sveltekit-reload>Réessayer</a></p>
		</div>
	{:else}
		{@render children()}
	{/if}
</main>
