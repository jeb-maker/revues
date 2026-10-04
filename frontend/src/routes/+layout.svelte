<script lang="ts">
	import '$lib/styles/app.css';
	import type { Snippet } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { logout } from '$lib/api/auth';
	import { resetSession } from '$lib/auth/session';
	import { runLabels, subjectLabels, templateNavLabel } from '$lib/i18n/uiLabels';
	import { inputValue } from '$lib/mb';
	import type { LayoutData } from './$types';

	let { data, children }: { data: LayoutData; children: Snippet } = $props();

	const boot = $derived(data.boot);
	let logoutError = $state('');
	let loggingOut = $state(false);
	let headerQuery = $state('');

	/** Pages auth : pas de shell (navbar / compte / recherche). */
	const hideAppChrome = $derived(
		page.url.pathname === '/login' || page.url.pathname === '/register'
	);

	$effect(() => {
		if (page.url.pathname === '/search') {
			headerQuery = page.url.searchParams.get('q') ?? '';
		}
	});

	async function onHeaderSearch(e: Event) {
		e.preventDefault();
		const q = headerQuery.trim();
		const href = q ? `/search?q=${encodeURIComponent(q)}` : '/search';
		await goto(href, { keepFocus: true });
	}

	const links = $derived.by(() => {
		const run = runLabels(boot.organization?.ui_run_label);
		const subject = subjectLabels(boot.organization?.ui_subject_label);
		const items: { href: string; label: string }[] = [
			{ href: '/runs', label: run.nav }
		];
		items.push({ href: '/subjects', label: subject.plural });
		if (boot.show_modeles) {
			items.push({ href: '/modeles', label: templateNavLabel(boot.organization?.ui_run_label) });
		}
		if (boot.show_my_tasks) {
			items.push({ href: '/mes-taches', label: 'Mes tâches' });
		}
		return items;
	});

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

{#if !hideAppChrome && boot.authenticated && boot.user}
	<header class="app-header">
		<div class="app-header__inner">
			<div class="app-header__nav">
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
			</div>
			<form class="app-header__search" role="search" onsubmit={onHeaderSearch}>
				<mb-input
					label="Recherche"
					hide-label
					density="compact"
					type="search"
					name="q"
					placeholder="Rechercher…"
					value={headerQuery}
					oninput={(e) => (headerQuery = inputValue(e))}
				></mb-input>
				<mb-button type="submit" variant="secondary" size="sm">OK</mb-button>
			</form>
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
				<mb-button variant="ghost" size="sm" loading={loggingOut} onclick={onLogout}>
					Se déconnecter
				</mb-button>
			</div>
		</div>
		{#if logoutError}
			<mb-alert variant="danger">{logoutError}</mb-alert>
		{/if}
	</header>
{/if}

<main id="main" tabindex="-1">
	{#if data.apiError}
		<div class="page page--narrow">
			<header class="page-header">
				<h1>Service indisponible</h1>
				<p class="lede">Revues ne répond pas pour le moment. Réessayez dans un instant.</p>
			</header>
			<mb-alert variant="danger">{data.apiError}</mb-alert>
			<p class="actions"><a href={page.url.pathname} data-sveltekit-reload>Réessayer</a></p>
		</div>
	{:else}
		{@render children()}
	{/if}
</main>
