<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { bootstrap } from '$lib/api/auth';
	import { listTemplates, type TemplateSummary } from '$lib/api/templates';

	let csrf = $state('');
	let templates = $state<TemplateSummary[]>([]);
	let q = $state('');
	let error = $state('');
	let loading = $state(true);
	let canManage = $state(false);

	async function load(query = q) {
		loading = true;
		error = '';
		try {
			const boot = await bootstrap();
			if (!boot.authenticated) {
				await goto('/login');
				return;
			}
			csrf = boot.csrf_token;
			canManage = boot.user?.role === 'admin' || boot.user?.role === 'editor';
			templates = await listTemplates(csrf, query);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Erreur de chargement';
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		void load();
	});

	async function onSearch(e: Event) {
		e.preventDefault();
		await load(q);
	}
</script>

<svelte:head>
	<title>Modèles — Revues</title>
</svelte:head>

<main class="page">
	<header class="top">
		<p class="brand"><a href="/">Revues</a></p>
		<h1>Modèles</h1>
		<p class="lede">Catalogue des check-lists versionnées de l’organisation.</p>
	</header>

	{#if error}
		<p class="err" role="alert">{error}</p>
	{/if}

	<form class="toolbar" onsubmit={onSearch}>
		<input type="search" bind:value={q} placeholder="Rechercher…" aria-label="Recherche" />
		<button type="submit">Filtrer</button>
		{#if canManage}
			<a class="primary" href="/modeles/new">Nouveau modèle</a>
			<a class="secondary" href="/modeles/notion-import">Import Notion</a>
		{/if}
	</form>

	{#if loading}
		<p class="muted">Chargement…</p>
	{:else if templates.length === 0}
		<p class="muted">Aucun modèle. {#if canManage}<a href="/modeles/new">Créer le premier</a>{/if}</p>
	{:else}
		<ul class="list">
			{#each templates as t (t.id)}
				<li>
					<a href={`/modeles/${t.id}`}>
						<strong>{t.name}</strong>
						<span class="meta">v{t.latest_version} · {t.item_count} points</span>
						{#if t.domains?.length}
							<span class="domains">{t.domains.join(', ')}</span>
						{/if}
					</a>
				</li>
			{/each}
		</ul>
	{/if}
</main>

<style>
	h1 {
		margin: 0 0 0.35rem;
		font-size: 1.35rem;
		font-weight: 600;
	}
	.toolbar {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
		margin-bottom: 1.25rem;
	}
	.toolbar input {
		flex: 1;
		min-width: 10rem;
		padding: 0.45rem 0.6rem;
		border-radius: 0.35rem;
		border: 1px solid #334155;
		background: #0b1220;
		color: #f8fafc;
		font: inherit;
	}
	.toolbar button,
	.toolbar a {
		padding: 0.45rem 0.75rem;
		border-radius: 0.35rem;
		border: 1px solid #334155;
		background: transparent;
		color: #5eead4;
		text-decoration: none;
		font: inherit;
		cursor: pointer;
	}
	.toolbar a.primary {
		background: #0f766e;
		border-color: #0f766e;
		color: #ecfdf5;
		font-weight: 600;
	}
	.toolbar a.secondary {
		border-color: #5eead4;
	}
	.list {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}
	.list a {
		display: block;
		padding: 0.75rem 0.9rem;
		border-radius: 0.5rem;
		border: 1px solid #334155;
		color: inherit;
		text-decoration: none;
		background: rgba(15, 23, 42, 0.55);
	}
	.meta,
	.domains {
		display: block;
		font-size: 0.85rem;
		color: #94a3b8;
	}
	.muted a {
		color: #5eead4;
	}
</style>
