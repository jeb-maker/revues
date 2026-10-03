<script lang="ts">
	import { onMount } from 'svelte';
	import { listTemplates, type TemplateSummary } from '$lib/api/templates';
	import { session } from '$lib/auth/session';
	import { inputValue } from '$lib/mb';

	const boot = session();
	const canManage = boot.can_edit;

	let templates = $state<TemplateSummary[]>([]);
	let q = $state('');
	let error = $state('');
	let loading = $state(true);

	async function load(query = q) {
		loading = true;
		error = '';
		try {
			templates = await listTemplates(boot.csrf_token, query);
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

<div class="page">
	<header class="page-header">
		<h1>Modèles</h1>
		<p class="lede">Catalogue des check-lists versionnées de l’organisation.</p>
		{#if canManage}
			<p class="actions">
				<mb-button variant="primary" href="/modeles/new">Nouveau modèle</mb-button>
				<mb-button variant="secondary" href="/modeles/notion-import">Importer depuis Notion</mb-button>
			</p>
		{/if}
	</header>

	{#if error}
		<mb-alert variant="danger">{error}</mb-alert>
	{/if}

	<form class="filters" onsubmit={onSearch}>
		<mb-input
			label="Rechercher"
			hide-label
			type="search"
			placeholder="Rechercher…"
			value={q}
			oninput={(e) => (q = inputValue(e))}
		></mb-input>
		<mb-button type="submit" variant="secondary">Filtrer</mb-button>
	</form>

	{#if loading}
		<p class="loading"><mb-spinner label="Chargement"></mb-spinner> Chargement…</p>
	{:else if templates.length === 0}
		<mb-empty-state heading="Aucun modèle">
			{#if canManage}<a href="/modeles/new">Créer le premier modèle</a>.{:else}Aucun modèle publié.{/if}
		</mb-empty-state>
	{:else}
		<ul class="card-list">
			{#each templates as t (t.id)}
				<li>
					<a href={`/modeles/${t.id}`}>
						<strong>{t.name}</strong>
						<span class="desc">v{t.latest_version} · {t.item_count} points</span>
						{#if t.domains?.length}
							<span class="desc">{t.domains.join(', ')}</span>
						{/if}
					</a>
				</li>
			{/each}
		</ul>
	{/if}
</div>
