<script lang="ts">
	import { onMount } from 'svelte';
	import { listTemplates, type TemplateSummary } from '$lib/api/templates';
	import { session } from '$lib/auth/session';
	import { templateNavLabel } from '$lib/i18n/uiLabels';
	import { inputValue } from '$lib/mb';

	const boot = session();
	const canManage = boot.can_edit;
	const templatesLabel = $derived(templateNavLabel(boot.organization?.ui_run_label));
	const newTemplateCTA = $derived(templatesLabel === 'Listes' ? 'Nouvelle liste' : 'Nouveau modèle');

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
	<title>{templatesLabel} — Revues</title>
</svelte:head>

<div class="page">
	<header class="page-header">
		<h1>{templatesLabel}</h1>
		<p class="lede">Catalogue des check-lists versionnées de l’organisation.</p>
		{#if canManage}
			<p class="actions">
				<mb-button variant="primary" href="/modeles/new">{newTemplateCTA}</mb-button>
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
		<mb-table columns="2fr 0.7fr 0.7fr 1.4fr auto" sticky-header>
			<mb-table-row slot="head">
				<mb-table-cell>Nom</mb-table-cell>
				<mb-table-cell>Version</mb-table-cell>
				<mb-table-cell>Points</mb-table-cell>
				<mb-table-cell>Domaines</mb-table-cell>
				<mb-table-cell>Actions</mb-table-cell>
			</mb-table-row>
			{#each templates as t (t.id)}
				<mb-table-row>
					<mb-table-cell label="Nom" primary>
						<a href={`/modeles/${t.id}`}>{t.name}</a>
					</mb-table-cell>
					<mb-table-cell label="Version">v{t.latest_version}</mb-table-cell>
					<mb-table-cell label="Points">{t.item_count}</mb-table-cell>
					<mb-table-cell label="Domaines">
						{#if t.domains?.length}
							{t.domains.join(', ')}
						{:else}
							<span class="muted">—</span>
						{/if}
					</mb-table-cell>
					<mb-table-cell actions>
						<a href={`/modeles/${t.id}`}>Ouvrir</a>
						{#if canManage}
							<a href={`/modeles/${t.id}/edit`}>Éditer</a>
						{/if}
					</mb-table-cell>
				</mb-table-row>
			{/each}
		</mb-table>
	{/if}
</div>
